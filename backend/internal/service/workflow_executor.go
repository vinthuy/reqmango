package service

import (
	"fmt"
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

// Execute runs a workflow.
func (e *WorkflowExecutor) Execute(runID uint64) error {
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
			"status":    "running",
			"started_at": now,
		})

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

	// Execute nodes in topology order
	for _, layer := range topology {
		if len(layer) == 1 {
			// Serial execution
			err = e.executeNode(runID, &layer[0], ctx)
			if err != nil {
				e.failRun(runID, err.Error())
				return err
			}
		} else {
			// Parallel execution
			err = e.executeParallel(runID, layer, ctx)
			if err != nil {
				e.failRun(runID, err.Error())
				return err
			}
		}

		// Update context for next layer
		if len(layer) > 0 {
			newCtx, err := e.contextSvc.BuildNodeInput(runID, layer[0].ID, nil)
			if err == nil {
				ctx = newCtx
			}
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
func (e *WorkflowExecutor) executeNode(runID uint64, node *WorkflowNodeResponse, ctx *ContextPayload) error {
	// Get node run
	var nodeRun struct {
		ID uint64 `json:"id"`
	}

	err := e.db.Raw(`
		SELECT id FROM workflow_node_runs 
		WHERE workflow_run_id = ? AND node_id = ? 
		LIMIT 1
	`, runID, node.ID).Scan(&nodeRun).Error

	if err != nil {
		return fmt.Errorf("failed to get node run: %w", err)
	}

	// Update node run status to running
	now := time.Now()
	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRun.ID).
		Updates(map[string]interface{}{
			"status":     "running",
			"started_at": now,
		})

	// Build context for this node
	nodeCtxJSON, _ := e.contextSvc.ContextToJSON(ctx)
	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRun.ID).
		Update("input_context", nodeCtxJSON)

	// Record decision
	e.decisionSvc.Record(&AgentDecisionRecord{
		AgentID:     node.AgentID,
		WorkflowRunID: &runID,
		NodeType:    node.NodeType,
		Thinking:    fmt.Sprintf("Starting node: %s", node.Name),
		Decision:    fmt.Sprintf("Agent %s will execute task", node.AgentName),
		Confidence:  0.9,
	})

	// Simulate execution (in real implementation, this would call the AI service)
	// For now, we'll just mark it as completed
	time.Sleep(100 * time.Millisecond)

	// Update node run status to completed
	completedAt := time.Now()
	outputCtx := &AgentOutput{
		AgentID:    node.AgentID,
		AgentName:  node.AgentName,
		NodeType:   node.NodeType,
		Content:    fmt.Sprintf("Completed node: %s", node.Name),
		TokenCount: 100,
	}
	outputCtxJSON, _ := e.contextSvc.ContextToJSON(&ContextPayload{
		AgentOutputs: []AgentOutput{*outputCtx},
	})

	e.db.Table("workflow_node_runs").
		Where("id = ?", nodeRun.ID).
		Updates(map[string]interface{}{
			"status":         "completed",
			"completed_at":   completedAt,
			"output_context": outputCtxJSON,
			"tokens_used":    100,
			"cost":           0.001,
		})

	return nil
}

// executeParallel executes multiple nodes in parallel.
func (e *WorkflowExecutor) executeParallel(runID uint64, nodes []WorkflowNodeResponse, ctx *ContextPayload) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(nodes))
	var mu sync.Mutex
	var errs []error

	for _, node := range nodes {
		wg.Add(1)
		go func(n WorkflowNodeResponse) {
			defer wg.Done()

			// Acquire semaphore
			e.semaphore <- struct{}{}
			defer func() { <-e.semaphore }()

			if err := e.executeNode(runID, &n, ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				errCh <- err
			}
		}(node)
	}

	wg.Wait()
	close(errCh)

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
