package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// WorkflowExecutor handles the execution of workflows.
type WorkflowExecutor struct {
	db          *gorm.DB
	workflowSvc *WorkflowService
	contextSvc  *ContextPayloadService
	decisionSvc *AgentDecisionService
	budgetSvc   *AgentCostBudgetService
	slaSvc      *AgentSLAService
	agentSvc    AgentExecutorInterface
	semaphore   chan struct{}
	maxParallel int
}

// NewWorkflowExecutor creates a new WorkflowExecutor.
func NewWorkflowExecutor(
	db *gorm.DB,
	workflowSvc *WorkflowService,
	contextSvc *ContextPayloadService,
	decisionSvc *AgentDecisionService,
	budgetSvc *AgentCostBudgetService,
	slaSvc *AgentSLAService,
) *WorkflowExecutor {
	return &WorkflowExecutor{
		db:          db,
		workflowSvc: workflowSvc,
		contextSvc:  contextSvc,
		decisionSvc: decisionSvc,
		budgetSvc:   budgetSvc,
		slaSvc:      slaSvc,
		semaphore:   make(chan struct{}, 10),
		maxParallel: 5,
	}
}

// SetAgentExecutor injects the agent executor used to dispatch node tasks.
func (e *WorkflowExecutor) SetAgentExecutor(agentSvc AgentExecutorInterface) {
	e.agentSvc = agentSvc
}

// Execute runs a workflow.
func (e *WorkflowExecutor) Execute(runID, userID uint64) error {
	// Get run details
	run, err := e.workflowSvc.GetRun(runID)
	if err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	// Get workflow with nodes and edges
	workflow, err := e.workflowSvc.Get(run.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Update run status to running
	now := time.Now()
	e.db.Table("workflow_runs").
		Where("id = ?", runID).
		Updates(map[string]interface{}{
			"status":     "running",
			"started_at": now,
		})

	// Guard against panics so a run is never left stuck in "running".
	defer func() {
		if r := recover(); r != nil {
			e.failRun(runID, fmt.Sprintf("panic: %v", r))
		}
	}()

	// Build topology
	topology := e.getTopology(workflow.Nodes, workflow.Edges)

	// Build initial context
	var issueID uint64
	if run.IssueID != nil {
		issueID = *run.IssueID
	}
	ctx, err := e.contextSvc.BuildInitialContext(issueID)
	if err != nil {
		e.failRun(runID, err.Error())
		return err
	}

	// Track nodes skipped by condition branching (nodeID -> true).
	// A skipped node is marked as "skipped" and not dispatched.
	skipped := make(map[uint64]bool)

	// Execute nodes in topology order
	for _, layer := range topology {
		// Filter out skipped nodes within this layer.
		var active []WorkflowNodeResponse
		for _, n := range layer {
			if skipped[n.ID] {
				e.markNodeSkipped(runID, n.ID)
				continue
			}
			active = append(active, n)
		}
		if len(active) == 0 {
			continue
		}

		var layerErr error
		if len(active) == 1 {
			layerErr = e.executeNode(runID, &active[0], ctx, userID, skipped)
		} else {
			layerErr = e.executeParallel(runID, active, ctx, userID, skipped)
		}

		if layerErr != nil {
			e.failRun(runID, layerErr.Error())
			return layerErr
		}

		// Aggregate outputs from ALL active nodes in this layer (not just the
		// first one) so downstream layers receive the full parallel output.
		var aggregated *ContextPayload
		for _, n := range active {
			nodeCtx, bErr := e.contextSvc.BuildNodeInput(runID, n.ID, nil)
			if bErr != nil || nodeCtx == nil {
				continue
			}
			if aggregated == nil {
				aggregated = nodeCtx
			} else {
				aggregated.AgentOutputs = append(aggregated.AgentOutputs, nodeCtx.AgentOutputs...)
			}
		}
		if aggregated != nil {
			ctx = aggregated
		}
	}

	// Update run status to completed
	completedAt := time.Now()
	e.db.Table("workflow_runs").
		Where("id = ?", runID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": completedAt,
		})

	return nil
}

// executeNode executes a single workflow node.
func (e *WorkflowExecutor) executeNode(runID uint64, node *WorkflowNodeResponse, ctx *ContextPayload, userID uint64, skipped map[uint64]bool) error {
	// Condition / gate nodes evaluate a branch rather than dispatch an agent.
	if node.NodeType == "condition" || node.NodeType == "gate" {
		return e.executeConditionNode(runID, node, ctx, userID, skipped)
	}

	// Get node run
	nodeRunID, err := e.getNodeRunID(runID, node.ID)
	if err != nil {
		return fmt.Errorf("failed to get node run: %w", err)
	}

	// Update node run status to running
	now := time.Now()
	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRunID).
		Updates(map[string]interface{}{
			"status":     "running",
			"started_at": now,
		})

	// Build context for this node
	nodeCtxJSON, _ := e.contextSvc.ContextToJSON(ctx)
	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRunID).
		Update("input_context", nodeCtxJSON)

	// Build the task description from node config + context
	task := e.buildNodeTask(node, ctx)

	// Resolve workspace / project / issue for the agent dispatch context
	workspaceID, projectID := e.resolveRunScope(runID)
	var issueIDPtr *uint64
	if issueID := e.resolveRunIssueID(runID); issueID > 0 {
		issueIDPtr = &issueID
	}

	// Execute with timeout + retry based on node policy
	resultSummary, tokensUsed, cost, execErr := e.dispatchWithRetry(node, task, userID, workspaceID, projectID, issueIDPtr)

	completedAt := time.Now()

	if execErr != nil {
		// Honour retry policy: "skip" marks node failed but continues; "abort" returns error.
		e.db.Table("workflow_node_runs").
			Where("id = ?", nodeRunID).
			Updates(map[string]interface{}{
				"status":       "failed",
				"completed_at": completedAt,
				"error_info":   execErr.Error(),
				"tokens_used":  tokensUsed,
				"cost":         cost,
			})

		e.decisionSvc.Record(&AgentDecisionRecord{
			AgentID:       node.AgentID,
			WorkflowRunID: &runID,
			NodeType:      node.NodeType,
			Thinking:      fmt.Sprintf("Node %s failed after retries", node.Name),
			Decision:      "failed",
			Reasoning:     execErr.Error(),
			Confidence:    0.0,
		})

		if node.RetryPolicy == "abort" {
			return execErr
		}
		// "skip" policy: record failure but continue the workflow
		return nil
	}

	// Persist output context
	outputCtx := &AgentOutput{
		AgentID:    node.AgentID,
		AgentName:  node.AgentName,
		NodeType:   node.NodeType,
		Content:    resultSummary,
		TokenCount: tokensUsed,
	}
	outputCtxJSON, _ := e.contextSvc.ContextToJSON(&ContextPayload{
		AgentOutputs: []AgentOutput{*outputCtx},
	})

	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRunID).
		Updates(map[string]interface{}{
			"status":         "completed",
			"completed_at":   completedAt,
			"output_context": outputCtxJSON,
			"tokens_used":    tokensUsed,
			"cost":           cost,
		})

	// Roll up totals on the run
	e.db.Table("workflow_runs").
		Where("id = ?", runID).
		Updates(map[string]interface{}{
			"total_tokens": gorm.Expr("total_tokens + ?", tokensUsed),
			"total_cost":   gorm.Expr("total_cost + ?", cost),
		})

	// Record cost against the project budget
	if cost > 0 && projectID > 0 {
		e.budgetSvc.RecordCost(projectID, cost)
	}

	// Record decision audit
	e.decisionSvc.Record(&AgentDecisionRecord{
		AgentID:       node.AgentID,
		WorkflowRunID: &runID,
		NodeType:      node.NodeType,
		Thinking:      fmt.Sprintf("Executing node: %s", node.Name),
		Decision:      fmt.Sprintf("Agent %s completed task", node.AgentName),
		Reasoning:     truncate(resultSummary, 500),
		Confidence:    0.9,
	})

	return nil
}

// executeConditionNode evaluates a condition/gate node and selects a single
// outgoing branch. Nodes reachable from non-selected branches are marked as
// skipped so they are not dispatched.
func (e *WorkflowExecutor) executeConditionNode(runID uint64, node *WorkflowNodeResponse, ctx *ContextPayload, userID uint64, skipped map[uint64]bool) error {
	nodeRunID, err := e.getNodeRunID(runID, node.ID)
	if err != nil {
		return fmt.Errorf("failed to get node run: %w", err)
	}

	now := time.Now()
	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRunID).
		Updates(map[string]interface{}{
			"status":     "running",
			"started_at": now,
		})

	workflow, err := e.workflowSvc.GetByNodeID(node.ID)
	if err != nil {
		return err
	}

	// Outgoing edges of this condition node.
	var outgoing []WorkflowEdgeResponse
	for _, edge := range workflow.Edges {
		if edge.SourceNodeID == node.ID {
			outgoing = append(outgoing, edge)
		}
	}

	chosenCondition, reasoning := e.evaluateCondition(node, ctx, userID, outgoing)

	// Determine which outgoing edge matches the chosen condition.
	chosenTarget, found := pickBranch(outgoing, chosenCondition)

	// Build the set of nodes to keep (reachable from chosen target) vs skip.
	keepSet := make(map[uint64]bool)
	if found && chosenTarget > 0 {
		reachable := e.forwardReachable(chosenTarget, workflow.Nodes, workflow.Edges)
		for id := range reachable {
			keepSet[id] = true
		}
	}

	// Mark every other outgoing branch's subtree as skipped (unless also kept).
	for _, edge := range outgoing {
		if edge.TargetNodeID == chosenTarget {
			continue
		}
		branchReachable := e.forwardReachable(edge.TargetNodeID, workflow.Nodes, workflow.Edges)
		for id := range branchReachable {
			if !keepSet[id] {
				skipped[id] = true
			}
		}
	}

	completedAt := time.Now()
	outputCtx := &AgentOutput{
		AgentID:   node.AgentID,
		AgentName: node.AgentName,
		NodeType:  node.NodeType,
		Content:   fmt.Sprintf("Condition resolved to: %s", chosenCondition),
	}
	outputCtxJSON, _ := e.contextSvc.ContextToJSON(&ContextPayload{
		AgentOutputs: []AgentOutput{*outputCtx},
	})

	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRunID).
		Updates(map[string]interface{}{
			"status":         "completed",
			"completed_at":   completedAt,
			"output_context": outputCtxJSON,
		})

	e.decisionSvc.Record(&AgentDecisionRecord{
		AgentID:       node.AgentID,
		WorkflowRunID: &runID,
		NodeType:      node.NodeType,
		Thinking:      fmt.Sprintf("Evaluating condition node: %s", node.Name),
		Decision:      chosenCondition,
		Reasoning:     reasoning,
		Confidence:    0.8,
	})

	return nil
}

// evaluateCondition resolves a condition node to a branch label.
// Strategy:
//  1. If node.Config contains an "expression" field, evaluate it against the
//     context (simple key==value / key!=value over shared_data + issue fields).
//  2. Otherwise, if an agent is configured and available, ask the agent to
//     pick one of the branch conditions.
//  3. Fallback: pick the first outgoing edge condition.
func (e *WorkflowExecutor) evaluateCondition(node *WorkflowNodeResponse, ctx *ContextPayload, userID uint64, outgoing []WorkflowEdgeResponse) (string, string) {
	branches := make([]string, 0, len(outgoing))
	for _, edge := range outgoing {
		branches = append(branches, edge.Condition)
	}

	// 1. Expression-based evaluation
	if expr := extractStringField(node.Config, "expression"); expr != "" {
		if chosen, ok := evalSimpleExpression(expr, ctx, branches); ok {
			return chosen, fmt.Sprintf("Expression %q matched branch %q", expr, chosen)
		}
	}

	// 2. Agent-based evaluation
	if e.agentSvc != nil && node.AgentID > 0 && len(branches) > 0 {
		task := fmt.Sprintf(
			"你是一个流程判断节点。根据以下上下文，从给定的分支中选择一个最合适的。\n\n"+
				"可选分支：%s\n"+
				"上下文：%s\n\n"+
				"请只回复所选分支的名称（与上面列出的完全一致），不要其它内容。",
			strings.Join(branches, ", "), summarizeContext(ctx))

		workspaceID, projectID := e.resolveRunScopeByWorkflow(node.WorkflowID)
		result, err := e.agentSvc.DispatchAgent(node.AgentID, userID, task, &AgentDispatchContext{
			WorkspaceID: workspaceID,
			ProjectID:   &projectID,
			TriggeredBy: "workflow_condition",
		})
		if err == nil && result != nil {
			if picked := matchBranch(result.ResultSummary, branches); picked != "" {
				return picked, fmt.Sprintf("Agent %s selected branch %q", node.AgentName, picked)
			}
		}
	}

	// 3. Fallback: first branch
	if len(branches) > 0 {
		branch := branches[0]
		if branch == "" {
			branch = "default"
		}
		return branch, "No expression/agent result; selected first branch as default"
	}
	return "default", "No outgoing branches; using default"
}

// dispatchWithRetry dispatches the node task to the configured agent, honouring
// the node timeout and retry policy. Returns the result summary, estimated
// token usage, estimated cost, and the final error (nil on success).
func (e *WorkflowExecutor) dispatchWithRetry(node *WorkflowNodeResponse, task string, userID, workspaceID, projectID uint64, issueID *uint64) (string, int, float64, error) {
	maxRetries := node.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	// NOTE: node.Timeout is intentionally not applied here. The agent
	// executor interface does not accept a context, so per-call timeout is
	// enforced by the underlying LLM/SLA layer rather than at this layer.

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if e.agentSvc == nil || node.AgentID == 0 {
			// No agent configured: produce a no-op result so the workflow can
			// still progress (mirrors the previous fallback behaviour but
			// without pretending to be an agent).
			summary := fmt.Sprintf("Node %s completed (no agent configured)", node.Name)
			return summary, estimateTokens(summary), 0, nil
		}

		result, err := e.agentSvc.DispatchAgent(node.AgentID, userID, task, &AgentDispatchContext{
			IssueID:     issueID,
			ProjectID:   &projectID,
			WorkspaceID: workspaceID,
			TriggeredBy: "workflow",
		})

		if err == nil && result != nil {
			summary := result.ResultSummary
			tokens := estimateTokens(summary)
			cost := estimateCost(tokens)
			return summary, tokens, cost, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second) // linear backoff
		}
	}
	return "", 0, 0, lastErr
}

// executeParallel executes multiple nodes in parallel.
func (e *WorkflowExecutor) executeParallel(runID uint64, nodes []WorkflowNodeResponse, ctx *ContextPayload, userID uint64, skipped map[uint64]bool) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for i := range nodes {
		wg.Add(1)
		go func(n WorkflowNodeResponse) {
			defer wg.Done()

			// Acquire semaphore
			e.semaphore <- struct{}{}
			defer func() { <-e.semaphore }()

			if skipped[n.ID] {
				e.markNodeSkipped(runID, n.ID)
				return
			}

			if err := e.executeNode(runID, &n, ctx, userID, skipped); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(nodes[i])
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("parallel execution failed: %v", errs)
	}

	return nil
}

// getTopology returns nodes in topological order.
func (e *WorkflowExecutor) getTopology(nodes []WorkflowNodeResponse, edges []WorkflowEdgeResponse) [][]WorkflowNodeResponse {
	// Build adjacency list and in-degree map
	inDegree := make(map[uint64]int)
	adj := make(map[uint64][]uint64)

	for _, edge := range edges {
		adj[edge.SourceNodeID] = append(adj[edge.SourceNodeID], edge.TargetNodeID)
		inDegree[edge.TargetNodeID]++
	}

	// Kahn's algorithm
	var result [][]WorkflowNodeResponse
	var queue []uint64

	for _, node := range nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node.ID)
		}
	}

	for len(queue) > 0 {
		var layer []WorkflowNodeResponse
		var nextQueue []uint64

		for _, id := range queue {
			for _, node := range nodes {
				if node.ID == id {
					layer = append(layer, node)
					break
				}
			}

			for _, next := range adj[id] {
				inDegree[next]--
				if inDegree[next] == 0 {
					nextQueue = append(nextQueue, next)
				}
			}
		}

		result = append(result, layer)
		queue = nextQueue
	}

	return result
}

// failRun marks a run as failed.
func (e *WorkflowExecutor) failRun(runID uint64, errorInfo string) {
	now := time.Now()
	e.db.Table("workflow_runs").
		Where("id = ?", runID).
		Updates(map[string]interface{}{
			"status":       "failed",
			"completed_at": now,
			"error_info":   errorInfo,
		})
}

// CancelRun cancels a running workflow.
func (e *WorkflowExecutor) CancelRun(runID uint64) error {
	result := e.db.Table("workflow_runs").
		Where("id = ? AND status IN ('pending', 'running')", runID).
		Updates(map[string]interface{}{
			"status": "cancelled",
		})

	if result.RowsAffected == 0 {
		return fmt.Errorf("run not found or already completed")
	}

	return nil
}

// ======== Helpers ========

// getNodeRunID resolves the workflow_node_runs row id for a given (run, node).
func (e *WorkflowExecutor) getNodeRunID(runID, nodeID uint64) (uint64, error) {
	var nodeRun struct {
		ID uint64 `json:"id"`
	}
	err := e.db.Raw(`
		SELECT id FROM workflow_node_runs
		WHERE workflow_run_id = ? AND node_id = ?
		LIMIT 1
	`, runID, nodeID).Scan(&nodeRun).Error
	if err != nil {
		return 0, err
	}
	return nodeRun.ID, nil
}

// markNodeSkipped marks a node run as skipped (used by condition branching).
func (e *WorkflowExecutor) markNodeSkipped(runID, nodeID uint64) {
	e.db.Table("workflow_node_runs").
		Where("workflow_run_id = ? AND node_id = ?", runID, nodeID).
		Updates(map[string]interface{}{
			"status":       "skipped",
			"completed_at": time.Now(),
		})
}

// buildNodeTask constructs the task description sent to the agent for a node.
func (e *WorkflowExecutor) buildNodeTask(node *WorkflowNodeResponse, ctx *ContextPayload) string {
	var sb strings.Builder

	if prompt := extractStringField(node.Config, "prompt"); prompt != "" {
		sb.WriteString(prompt)
	} else if task := extractStringField(node.Config, "task"); task != "" {
		sb.WriteString(task)
	} else {
		sb.WriteString(fmt.Sprintf("执行工作流节点「%s」的任务。", node.Name))
	}

	if ctx != nil && ctx.IssueContext != nil {
		ic := ctx.IssueContext
		sb.WriteString(fmt.Sprintf("\n\n关联 Issue #%d：%s（优先级：%s）", ic.IssueID, ic.Name, ic.Priority))
		if ic.Description != "" {
			sb.WriteString("\n描述：")
			sb.WriteString(truncate(ic.Description, 500))
		}
	}

	if ctx != nil && len(ctx.AgentOutputs) > 0 {
		sb.WriteString("\n\n前置节点输出：")
		for _, out := range ctx.AgentOutputs {
			sb.WriteString(fmt.Sprintf("\n- [%s] %s", out.AgentName, truncate(out.Content, 300)))
		}
	}

	return sb.String()
}

// resolveRunScope returns the workspaceID and projectID for a workflow run.
func (e *WorkflowExecutor) resolveRunScope(runID uint64) (uint64, uint64) {
	var scope struct {
		WorkspaceID uint64 `json:"workspace_id"`
		ProjectID   uint64 `json:"project_id"`
	}
	e.db.Raw(`
		SELECT aw.workspace_id, aw.project_id
		FROM workflow_runs wr
		JOIN agent_workflows aw ON aw.id = wr.workflow_id
		WHERE wr.id = ?
	`, runID).Scan(&scope)
	return scope.WorkspaceID, scope.ProjectID
}

// resolveRunScopeByWorkflow returns workspace/project for a workflow definition.
func (e *WorkflowExecutor) resolveRunScopeByWorkflow(workflowID uint64) (uint64, uint64) {
	var scope struct {
		WorkspaceID uint64 `json:"workspace_id"`
		ProjectID   uint64 `json:"project_id"`
	}
	e.db.Raw(`SELECT workspace_id, project_id FROM agent_workflows WHERE id = ?`, workflowID).Scan(&scope)
	return scope.WorkspaceID, scope.ProjectID
}

// resolveRunIssueID returns the issue id associated with a run, if any.
func (e *WorkflowExecutor) resolveRunIssueID(runID uint64) uint64 {
	var issueID *uint64
	e.db.Raw(`SELECT issue_id FROM workflow_runs WHERE id = ?`, runID).Scan(&issueID)
	if issueID == nil {
		return 0
	}
	return *issueID
}

// forwardReachable returns the set of node ids reachable from startNodeID
// (exclusive of startNodeID) by following edges forward.
func (e *WorkflowExecutor) forwardReachable(startNodeID uint64, nodes []WorkflowNodeResponse, edges []WorkflowEdgeResponse) map[uint64]bool {
	adj := make(map[uint64][]uint64)
	for _, edge := range edges {
		adj[edge.SourceNodeID] = append(adj[edge.SourceNodeID], edge.TargetNodeID)
	}

	visited := make(map[uint64]bool)
	var stack []uint64
	stack = append(stack, adj[startNodeID]...)
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[n] {
			continue
		}
		visited[n] = true
		stack = append(stack, adj[n]...)
	}
	return visited
}

// estimateTokens returns a rough token estimate for a string (≈ chars/4).
func estimateTokens(s string) int {
	if len(s) == 0 {
		return 0
	}
	tokens := len(s) / 4
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}

// estimateCost returns a rough USD cost estimate from a token count.
// Uses a conservative $0.002 / 1K tokens (input+output blended).
func estimateCost(tokens int) float64 {
	if tokens <= 0 {
		return 0
	}
	return float64(tokens) / 1000.0 * 0.002
}

// truncate caps a string to n characters, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// summarizeContext produces a compact text representation of a ContextPayload.
func summarizeContext(ctx *ContextPayload) string {
	if ctx == nil {
		return "(无上下文)"
	}
	var sb strings.Builder
	if ctx.IssueContext != nil {
		sb.WriteString(fmt.Sprintf("Issue #%d %s (优先级 %s)", ctx.IssueContext.IssueID, ctx.IssueContext.Name, ctx.IssueContext.Priority))
	}
	for _, out := range ctx.AgentOutputs {
		sb.WriteString(fmt.Sprintf("\n- [%s] %s", out.AgentName, truncate(out.Content, 200)))
	}
	return sb.String()
}

// extractStringField reads a string field from a JSON raw message.
func extractStringField(raw []byte, field string) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}

// evalSimpleExpression evaluates a very small expression language against the
// context. Supported forms (case-insensitive, whitespace-tolerant):
//
//	key == value
//	key != value
//
// Keys are looked up in shared_data first, then issue context fields.
// "value" may also be one of the provided branch labels; if it matches a
// branch label the expression is considered satisfied for that branch.
func evalSimpleExpression(expr string, ctx *ContextPayload, branches []string) (string, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", false
	}

	// Parse "key op value"
	var key, op, val string
	if idx := strings.Index(expr, "=="); idx >= 0 {
		key = strings.TrimSpace(expr[:idx])
		op = "=="
		val = strings.TrimSpace(expr[idx+2:])
	} else if idx := strings.Index(expr, "!="); idx >= 0 {
		key = strings.TrimSpace(expr[:idx])
		op = "!="
		val = strings.TrimSpace(expr[idx+2:])
	} else {
		return "", false
	}

	actual := lookupContextValue(ctx, key)
	matched := false
	switch op {
	case "==":
		matched = strings.EqualFold(actual, val)
	case "!=":
		matched = !strings.EqualFold(actual, val)
	}

	if !matched {
		return "", false
	}
	// If the value matches one of the branch labels, return that branch.
	for _, b := range branches {
		if strings.EqualFold(b, val) || strings.EqualFold(b, actual) {
			return b, true
		}
	}
	// Otherwise return the value itself as the chosen branch.
	if val != "" {
		return val, true
	}
	return actual, true
}

// lookupContextValue resolves a key against the context payload.
func lookupContextValue(ctx *ContextPayload, key string) string {
	if ctx == nil {
		return ""
	}
	lk := strings.ToLower(key)
	if ctx.SharedData != nil {
		if v, ok := ctx.SharedData[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		if v, ok := ctx.SharedData[lk]; ok {
			return fmt.Sprintf("%v", v)
		}
	}
	if ctx.IssueContext != nil {
		switch lk {
		case "priority":
			return ctx.IssueContext.Priority
		case "type":
			return ctx.IssueContext.Type
		case "state":
			return ctx.IssueContext.State
		case "name":
			return ctx.IssueContext.Name
		}
	}
	return ""
}

// matchBranch finds a branch label mentioned in an agent response.
func matchBranch(response string, branches []string) string {
	if response == "" || len(branches) == 0 {
		return ""
	}
	lower := strings.ToLower(response)
	// Prefer exact branch mention; fall back to a branch that appears as a word.
	for _, b := range branches {
		if b == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(b)) {
			return b
		}
	}
	return ""
}

// pickBranch selects the target node id whose edge condition matches the
// chosen condition. Falls back to the first outgoing edge.
func pickBranch(outgoing []WorkflowEdgeResponse, chosen string) (uint64, bool) {
	if len(outgoing) == 0 {
		return 0, false
	}
	for _, edge := range outgoing {
		if edge.Condition != "" && strings.EqualFold(edge.Condition, chosen) {
			return edge.TargetNodeID, true
		}
	}
	// Empty-condition edge acts as default branch.
	for _, edge := range outgoing {
		if edge.Condition == "" {
			return edge.TargetNodeID, true
		}
	}
	return outgoing[0].TargetNodeID, true
}
