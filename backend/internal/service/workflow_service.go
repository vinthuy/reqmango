package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// WorkflowService manages workflow definitions and execution.
type WorkflowService struct {
	db         *gorm.DB
	contextSvc *ContextPayloadService
	decisionSvc *AgentDecisionService
	budgetSvc  *AgentCostBudgetService
}

// NewWorkflowService creates a new WorkflowService.
func NewWorkflowService(
	db *gorm.DB,
	contextSvc *ContextPayloadService,
	decisionSvc *AgentDecisionService,
	budgetSvc *AgentCostBudgetService,
) *WorkflowService {
	return &WorkflowService{
		db:          db,
		contextSvc:  contextSvc,
		decisionSvc: decisionSvc,
		budgetSvc:   budgetSvc,
	}
}

// WorkflowResponse represents a workflow in API response.
type WorkflowResponse struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ProjectID     uint64  `json:"project_id"`
	WorkspaceID   uint64  `json:"workspace_id"`
	Version       int     `json:"version"`
	IsActive      bool    `json:"is_active"`
	TriggerType   string  `json:"trigger_type"`
	NodeCount     int     `json:"node_count"`
	EdgeCount     int     `json:"edge_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// WorkflowDetail represents a workflow with full details.
type WorkflowDetail struct {
	WorkflowResponse
	Nodes []WorkflowNodeResponse `json:"nodes"`
	Edges []WorkflowEdgeResponse `json:"edges"`
}

// WorkflowNodeResponse represents a workflow node in API response.
type WorkflowNodeResponse struct {
	ID            uint64  `json:"id"`
	WorkflowID    uint64  `json:"workflow_id"`
	AgentID       uint64  `json:"agent_id"`
	AgentName     string  `json:"agent_name"`
	NodeType      string  `json:"node_type"`
	Name          string  `json:"name"`
	Config        json.RawMessage `json:"config"`
	SortOrder     int     `json:"sort_order"`
	Timeout       int     `json:"timeout"`
	RetryPolicy   string  `json:"retry_policy"`
	MaxRetries    int     `json:"max_retries"`
}

// WorkflowEdgeResponse represents a workflow edge in API response.
type WorkflowEdgeResponse struct {
	ID             uint64  `json:"id"`
	WorkflowID     uint64  `json:"workflow_id"`
	SourceNodeID   uint64  `json:"source_node_id"`
	TargetNodeID   uint64  `json:"target_node_id"`
	Condition      string  `json:"condition"`
	ContextMapping json.RawMessage `json:"context_mapping"`
}

// CreateWorkflowRequest represents the request to create a workflow.
type CreateWorkflowRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	TriggerType   string  `json:"trigger_type"`
}

// UpdateWorkflowRequest represents the request to update a workflow.
type UpdateWorkflowRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	IsActive      *bool   `json:"is_active"`
	TriggerType   *string `json:"trigger_type"`
}

// CreateNodeRequest represents the request to create a workflow node.
type CreateNodeRequest struct {
	AgentID       uint64  `json:"agent_id" binding:"required"`
	NodeType      string  `json:"node_type"`
	Name          string  `json:"name" binding:"required"`
	Config        json.RawMessage `json:"config"`
	SortOrder     int     `json:"sort_order"`
	Timeout       int     `json:"timeout"`
	RetryPolicy   string  `json:"retry_policy"`
	MaxRetries    int     `json:"max_retries"`
}

// CreateEdgeRequest represents the request to create a workflow edge.
type CreateEdgeRequest struct {
	SourceNodeID   uint64  `json:"source_node_id" binding:"required"`
	TargetNodeID   uint64  `json:"target_node_id" binding:"required"`
	Condition      string  `json:"condition"`
	ContextMapping json.RawMessage `json:"context_mapping"`
}

// WorkflowRunResponse represents a workflow run in API response.
type WorkflowRunResponse struct {
	ID            uint64  `json:"id"`
	WorkflowID    uint64  `json:"workflow_id"`
	IssueID       *uint64 `json:"issue_id"`
	Status        string  `json:"status"`
	StartedAt     *string `json:"started_at"`
	CompletedAt   *string `json:"completed_at"`
	TotalTokens   int     `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	ErrorInfo     string  `json:"error_info"`
	CreatedAt     string  `json:"created_at"`
}

// WorkflowRunDetail represents a workflow run with node runs.
type WorkflowRunDetail struct {
	WorkflowRunResponse
	NodeRuns []WorkflowNodeRunResponse `json:"node_runs"`
}

// WorkflowNodeRunResponse represents a node run in API response.
type WorkflowNodeRunResponse struct {
	ID            uint64  `json:"id"`
	WorkflowRunID uint64  `json:"workflow_run_id"`
	NodeID        uint64  `json:"node_id"`
	NodeName      string  `json:"node_name"`
	AgentID       uint64  `json:"agent_id"`
	AgentName     string  `json:"agent_name"`
	Status        string  `json:"status"`
	StartedAt     *string `json:"started_at"`
	CompletedAt   *string `json:"completed_at"`
	TokensUsed    int     `json:"tokens_used"`
	Cost          float64 `json:"cost"`
	ErrorInfo     string  `json:"error_info"`
	RetryCount    int     `json:"retry_count"`
}

// ListByProject returns all workflows for a project.
func (s *WorkflowService) ListByProject(projectID uint64) ([]WorkflowResponse, error) {
	var workflows []struct {
		ID          uint64 `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ProjectID   uint64 `json:"project_id"`
		WorkspaceID uint64 `json:"workspace_id"`
		Version     int    `json:"version"`
		IsActive    bool   `json:"is_active"`
		TriggerType string `json:"trigger_type"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	err := s.db.Raw(`
		SELECT id, name, description, project_id, workspace_id, version, is_active, trigger_type, created_at, updated_at
		FROM agent_workflows
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, projectID).Scan(&workflows).Error

	if err != nil {
		return nil, err
	}

	var result []WorkflowResponse
	for _, w := range workflows {
		resp := WorkflowResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			ProjectID:   w.ProjectID,
			WorkspaceID: w.WorkspaceID,
			Version:     w.Version,
			IsActive:    w.IsActive,
			TriggerType: w.TriggerType,
			CreatedAt:   w.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   w.UpdatedAt.Format(time.RFC3339),
		}

		// Get node and edge counts
		s.db.Raw("SELECT COUNT(*) FROM workflow_nodes WHERE workflow_id = ?", w.ID).Scan(&resp.NodeCount)
		s.db.Raw("SELECT COUNT(*) FROM workflow_edges WHERE workflow_id = ?", w.ID).Scan(&resp.EdgeCount)

		result = append(result, resp)
	}

	if result == nil {
		result = []WorkflowResponse{}
	}

	return result, nil
}

// Create creates a new workflow.
func (s *WorkflowService) Create(projectID uint64, req CreateWorkflowRequest) (*WorkflowResponse, error) {
	if req.TriggerType == "" {
		req.TriggerType = "manual"
	}

	// Get workspace ID from project
	var workspaceID uint64
	s.db.Raw("SELECT workspace_id FROM projects WHERE id = ?", projectID).Scan(&workspaceID)

	workflow := &model.AgentWorkflow{
		Name:         req.Name,
		Description:  req.Description,
		ProjectID:    projectID,
		WorkspaceID:  workspaceID,
		Version:      1,
		IsActive:     true,
		TriggerType:  req.TriggerType,
	}

	if err := s.db.Create(workflow).Error; err != nil {
		return nil, err
	}

	return &WorkflowResponse{
		ID:          workflow.ID,
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		Version:     1,
		IsActive:    true,
		TriggerType: req.TriggerType,
		CreatedAt:   workflow.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   workflow.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetByNodeID returns the workflow that contains the given node id.
func (s *WorkflowService) GetByNodeID(nodeID uint64) (*WorkflowDetail, error) {
	var workflowID uint64
	err := s.db.Raw(`SELECT workflow_id FROM workflow_nodes WHERE id = ? AND deleted_at IS NULL`, nodeID).Scan(&workflowID).Error
	if err != nil {
		return nil, err
	}
	if workflowID == 0 {
		return nil, errors.New("workflow not found for node")
	}
	return s.Get(workflowID)
}

// Get returns a workflow with full details.
func (s *WorkflowService) Get(workflowID uint64) (*WorkflowDetail, error) {
	var workflow struct {
		ID          uint64 `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ProjectID   uint64 `json:"project_id"`
		WorkspaceID uint64 `json:"workspace_id"`
		Version     int    `json:"version"`
		IsActive    bool   `json:"is_active"`
		TriggerType string `json:"trigger_type"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	err := s.db.Raw(`
		SELECT id, name, description, project_id, workspace_id, version, is_active, trigger_type, created_at, updated_at
		FROM agent_workflows
		WHERE id = ? AND deleted_at IS NULL
	`, workflowID).Scan(&workflow).Error

	if err != nil {
		return nil, err
	}
	if workflow.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	detail := &WorkflowDetail{
		WorkflowResponse: WorkflowResponse{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Description: workflow.Description,
			ProjectID:   workflow.ProjectID,
			WorkspaceID: workflow.WorkspaceID,
			Version:     workflow.Version,
			IsActive:    workflow.IsActive,
			TriggerType: workflow.TriggerType,
			CreatedAt:   workflow.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   workflow.UpdatedAt.Format(time.RFC3339),
		},
		Nodes: []WorkflowNodeResponse{},
		Edges: []WorkflowEdgeResponse{},
	}

	// Get nodes
	var nodes []struct {
		ID        uint64 `json:"id"`
		AgentID   uint64 `json:"agent_id"`
		NodeType  string `json:"node_type"`
		Name      string `json:"name"`
		Config    []byte `json:"config"`
		SortOrder int    `json:"sort_order"`
		Timeout   int    `json:"timeout"`
		RetryPolicy string `json:"retry_policy"`
		MaxRetries int    `json:"max_retries"`
	}

	s.db.Raw(`
		SELECT id, agent_id, node_type, name, config, sort_order, timeout, retry_policy, max_retries
		FROM workflow_nodes
		WHERE workflow_id = ? AND deleted_at IS NULL
		ORDER BY sort_order
	`, workflowID).Scan(&nodes)

	for _, n := range nodes {
		var agentName string
		s.db.Raw("SELECT name FROM agents WHERE id = ?", n.AgentID).Scan(&agentName)

		detail.Nodes = append(detail.Nodes, WorkflowNodeResponse{
			ID:          n.ID,
			WorkflowID:  workflowID,
			AgentID:     n.AgentID,
			AgentName:   agentName,
			NodeType:    n.NodeType,
			Name:        n.Name,
			Config:      n.Config,
			SortOrder:   n.SortOrder,
			Timeout:     n.Timeout,
			RetryPolicy: n.RetryPolicy,
			MaxRetries:  n.MaxRetries,
		})
	}

	// Get edges
	var edges []struct {
		ID             uint64 `json:"id"`
		SourceNodeID   uint64 `json:"source_node_id"`
		TargetNodeID   uint64 `json:"target_node_id"`
		Condition      string `json:"condition"`
		ContextMapping []byte `json:"context_mapping"`
	}

	s.db.Raw(`
		SELECT id, source_node_id, target_node_id, condition, context_mapping
		FROM workflow_edges
		WHERE workflow_id = ? AND deleted_at IS NULL
	`, workflowID).Scan(&edges)

	for _, e := range edges {
		detail.Edges = append(detail.Edges, WorkflowEdgeResponse{
			ID:             e.ID,
			WorkflowID:     workflowID,
			SourceNodeID:   e.SourceNodeID,
			TargetNodeID:   e.TargetNodeID,
			Condition:      e.Condition,
			ContextMapping: e.ContextMapping,
		})
	}

	return detail, nil
}

// Update updates a workflow.
func (s *WorkflowService) Update(workflowID uint64, req UpdateWorkflowRequest) error {
	updates := map[string]interface{}{}

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TriggerType != nil {
		updates["trigger_type"] = *req.TriggerType
	}

	if len(updates) == 0 {
		return nil
	}

	result := s.db.Table("agent_workflows").
		Where("id = ?", workflowID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("workflow not found")
	}

	return result.Error
}

// Delete soft-deletes a workflow.
func (s *WorkflowService) Delete(workflowID uint64) error {
	result := s.db.Where("id = ?", workflowID).Delete(&struct{}{})
	if result.RowsAffected == 0 {
		return errors.New("workflow not found")
	}
	return nil
}

// AddNode adds a node to a workflow.
func (s *WorkflowService) AddNode(workflowID uint64, req CreateNodeRequest) (*WorkflowNodeResponse, error) {
	if req.NodeType == "" {
		req.NodeType = "agent"
	}
	if req.Timeout == 0 {
		req.Timeout = 1800
	}
	if req.RetryPolicy == "" {
		req.RetryPolicy = "retry"
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}

	node := &model.WorkflowNode{
		WorkflowID:  workflowID,
		AgentID:     req.AgentID,
		NodeType:    req.NodeType,
		Name:        req.Name,
		Config:      req.Config,
		SortOrder:   req.SortOrder,
		Timeout:     req.Timeout,
		RetryPolicy: req.RetryPolicy,
		MaxRetries:  req.MaxRetries,
	}

	if err := s.db.Create(node).Error; err != nil {
		return nil, err
	}

	// Get agent name
	var agentName string
	s.db.Raw("SELECT name FROM agents WHERE id = ?", req.AgentID).Scan(&agentName)

	return &WorkflowNodeResponse{
		ID:          node.ID,
		WorkflowID:  workflowID,
		AgentID:     req.AgentID,
		AgentName:   agentName,
		NodeType:    req.NodeType,
		Name:        req.Name,
		Config:      req.Config,
		SortOrder:   req.SortOrder,
		Timeout:     req.Timeout,
		RetryPolicy: req.RetryPolicy,
		MaxRetries:  req.MaxRetries,
	}, nil
}

// UpdateNode updates a workflow node.
func (s *WorkflowService) UpdateNode(nodeID uint64, req CreateNodeRequest) error {
	updates := map[string]interface{}{
		"agent_id":     req.AgentID,
		"name":         req.Name,
		"config":       req.Config,
		"sort_order":   req.SortOrder,
		"timeout":      req.Timeout,
		"retry_policy": req.RetryPolicy,
		"max_retries":  req.MaxRetries,
	}

	if req.NodeType != "" {
		updates["node_type"] = req.NodeType
	}

	result := s.db.Table("workflow_nodes").
		Where("id = ?", nodeID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("node not found")
	}

	return result.Error
}

// DeleteNode deletes a workflow node.
func (s *WorkflowService) DeleteNode(nodeID uint64) error {
	result := s.db.Where("id = ?", nodeID).Delete(&struct{}{})
	if result.RowsAffected == 0 {
		return errors.New("node not found")
	}
	return nil
}

// AddEdge adds an edge to a workflow.
func (s *WorkflowService) AddEdge(workflowID uint64, req CreateEdgeRequest) (*WorkflowEdgeResponse, error) {
	// Validate nodes exist
	var sourceExists, targetExists bool
	s.db.Raw("SELECT EXISTS(SELECT 1 FROM workflow_nodes WHERE id = ? AND workflow_id = ?)", req.SourceNodeID, workflowID).Scan(&sourceExists)
	s.db.Raw("SELECT EXISTS(SELECT 1 FROM workflow_nodes WHERE id = ? AND workflow_id = ?)", req.TargetNodeID, workflowID).Scan(&targetExists)

	if !sourceExists || !targetExists {
		return nil, errors.New("source or target node not found")
	}

	edge := &model.WorkflowEdge{
		WorkflowID:     workflowID,
		SourceNodeID:   req.SourceNodeID,
		TargetNodeID:   req.TargetNodeID,
		Condition:      req.Condition,
		ContextMapping: req.ContextMapping,
	}

	if err := s.db.Create(edge).Error; err != nil {
		return nil, err
	}

	return &WorkflowEdgeResponse{
		ID:             edge.ID,
		WorkflowID:     workflowID,
		SourceNodeID:   req.SourceNodeID,
		TargetNodeID:   req.TargetNodeID,
		Condition:      req.Condition,
		ContextMapping: req.ContextMapping,
	}, nil
}

// UpdateEdge updates a workflow edge.
func (s *WorkflowService) UpdateEdge(edgeID uint64, req CreateEdgeRequest) error {
	updates := map[string]interface{}{
		"source_node_id": req.SourceNodeID,
		"target_node_id": req.TargetNodeID,
		"condition":      req.Condition,
		"context_mapping": req.ContextMapping,
	}

	result := s.db.Table("workflow_edges").
		Where("id = ?", edgeID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("edge not found")
	}

	return result.Error
}

// DeleteEdge deletes a workflow edge.
func (s *WorkflowService) DeleteEdge(edgeID uint64) error {
	result := s.db.Where("id = ?", edgeID).Delete(&struct{}{})
	if result.RowsAffected == 0 {
		return errors.New("edge not found")
	}
	return nil
}

// GetRuns returns all runs for a workflow.
func (s *WorkflowService) GetRuns(workflowID uint64) ([]WorkflowRunResponse, error) {
	var runs []struct {
		ID          uint64     `json:"id"`
		WorkflowID  uint64     `json:"workflow_id"`
		IssueID     *uint64    `json:"issue_id"`
		Status      string     `json:"status"`
		StartedAt   *time.Time `json:"started_at"`
		CompletedAt *time.Time `json:"completed_at"`
		TotalTokens int        `json:"total_tokens"`
		TotalCost   float64    `json:"total_cost"`
		ErrorInfo   string     `json:"error_info"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	err := s.db.Raw(`
		SELECT id, workflow_id, issue_id, status, started_at, completed_at, total_tokens, total_cost, error_info, created_at
		FROM workflow_runs
		WHERE workflow_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, workflowID).Scan(&runs).Error

	if err != nil {
		return nil, err
	}

	var result []WorkflowRunResponse
	for _, r := range runs {
		resp := WorkflowRunResponse{
			ID:          r.ID,
			WorkflowID:  r.WorkflowID,
			IssueID:     r.IssueID,
			Status:      r.Status,
			TotalTokens: r.TotalTokens,
			TotalCost:   r.TotalCost,
			ErrorInfo:   r.ErrorInfo,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		}

		if r.StartedAt != nil {
			started := r.StartedAt.Format(time.RFC3339)
			resp.StartedAt = &started
		}
		if r.CompletedAt != nil {
			completed := r.CompletedAt.Format(time.RFC3339)
			resp.CompletedAt = &completed
		}

		result = append(result, resp)
	}

	if result == nil {
		result = []WorkflowRunResponse{}
	}

	return result, nil
}

// GetRun returns a workflow run with node runs.
func (s *WorkflowService) GetRun(runID uint64) (*WorkflowRunDetail, error) {
	var run struct {
		ID          uint64     `json:"id"`
		WorkflowID  uint64     `json:"workflow_id"`
		IssueID     *uint64    `json:"issue_id"`
		Status      string     `json:"status"`
		StartedAt   *time.Time `json:"started_at"`
		CompletedAt *time.Time `json:"completed_at"`
		TotalTokens int        `json:"total_tokens"`
		TotalCost   float64    `json:"total_cost"`
		ErrorInfo   string     `json:"error_info"`
		CreatedAt   time.Time  `json:"created_at"`
	}

	err := s.db.Raw(`
		SELECT id, workflow_id, issue_id, status, started_at, completed_at, total_tokens, total_cost, error_info, created_at
		FROM workflow_runs
		WHERE id = ? AND deleted_at IS NULL
	`, runID).Scan(&run).Error

	if err != nil {
		return nil, err
	}

	detail := &WorkflowRunDetail{
		WorkflowRunResponse: WorkflowRunResponse{
			ID:          run.ID,
			WorkflowID:  run.WorkflowID,
			IssueID:     run.IssueID,
			Status:      run.Status,
			TotalTokens: run.TotalTokens,
			TotalCost:   run.TotalCost,
			ErrorInfo:   run.ErrorInfo,
			CreatedAt:   run.CreatedAt.Format(time.RFC3339),
		},
		NodeRuns: []WorkflowNodeRunResponse{},
	}

	if run.StartedAt != nil {
		started := run.StartedAt.Format(time.RFC3339)
		detail.StartedAt = &started
	}
	if run.CompletedAt != nil {
		completed := run.CompletedAt.Format(time.RFC3339)
		detail.CompletedAt = &completed
	}

	// Get node runs
	var nodeRuns []struct {
		ID            uint64     `json:"id"`
		WorkflowRunID uint64     `json:"workflow_run_id"`
		NodeID        uint64     `json:"node_id"`
		AgentID       uint64     `json:"agent_id"`
		Status        string     `json:"status"`
		StartedAt     *time.Time `json:"started_at"`
		CompletedAt   *time.Time `json:"completed_at"`
		TokensUsed    int        `json:"tokens_used"`
		Cost          float64    `json:"cost"`
		ErrorInfo     string     `json:"error_info"`
		RetryCount    int        `json:"retry_count"`
	}

	s.db.Raw(`
		SELECT id, workflow_run_id, node_id, agent_id, status, started_at, completed_at, tokens_used, cost, error_info, retry_count
		FROM workflow_node_runs
		WHERE workflow_run_id = ?
		ORDER BY id
	`, runID).Scan(&nodeRuns)

	for _, nr := range nodeRuns {
		nrResp := WorkflowNodeRunResponse{
			ID:            nr.ID,
			WorkflowRunID: nr.WorkflowRunID,
			NodeID:        nr.NodeID,
			AgentID:       nr.AgentID,
			Status:        nr.Status,
			TokensUsed:    nr.TokensUsed,
			Cost:          nr.Cost,
			ErrorInfo:     nr.ErrorInfo,
			RetryCount:    nr.RetryCount,
		}

		// Get node name
		s.db.Raw("SELECT name FROM workflow_nodes WHERE id = ?", nr.NodeID).Scan(&nrResp.NodeName)

		// Get agent name
		s.db.Raw("SELECT name FROM agents WHERE id = ?", nr.AgentID).Scan(&nrResp.AgentName)

		if nr.StartedAt != nil {
			started := nr.StartedAt.Format(time.RFC3339)
			nrResp.StartedAt = &started
		}
		if nr.CompletedAt != nil {
			completed := nr.CompletedAt.Format(time.RFC3339)
			nrResp.CompletedAt = &completed
		}

		detail.NodeRuns = append(detail.NodeRuns, nrResp)
	}

	return detail, nil
}

// ExecuteWorkflow starts a workflow execution.
func (s *WorkflowService) ExecuteWorkflow(workflowID uint64, issueID *uint64) (*WorkflowRunResponse, error) {
	// Get workflow
	workflow, err := s.Get(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// Check budget
	if issueID != nil {
		ok, msg, err := s.budgetSvc.CheckBudget(workflow.ProjectID, 0.1) // estimated cost
		if err != nil || !ok {
			return nil, fmt.Errorf("budget check failed: %s", msg)
		}
	}

	// Build initial context (safe for nil issueID — BuildInitialContext handles id=0)
	var issueIDVal uint64
	if issueID != nil {
		issueIDVal = *issueID
	}
	ctx, err := s.contextSvc.BuildInitialContext(issueIDVal)
	if err != nil {
		return nil, err
	}

	ctxJSON, _ := s.contextSvc.ContextToJSON(ctx)

	// Create run (use model struct so GORM populates the auto-increment ID)
	now := time.Now()
	run := &model.WorkflowRun{
		WorkflowID: workflowID,
		IssueID:    issueID,
		Status:     "pending",
		Context:    ctxJSON,
		StartedAt:  &now,
	}

	if err := s.db.Create(run).Error; err != nil {
		return nil, err
	}

	// Create node runs for all nodes
	for _, node := range workflow.Nodes {
		nodeRun := &model.WorkflowNodeRun{
			WorkflowRunID: run.ID,
			NodeID:        node.ID,
			AgentID:       node.AgentID,
			Status:        "pending",
		}
		s.db.Create(nodeRun)
	}

	return &WorkflowRunResponse{
		ID:          run.ID,
		WorkflowID:  workflowID,
		IssueID:     issueID,
		Status:      "pending",
		StartedAt:   ptrString(now.Format(time.RFC3339)),
		CreatedAt:   run.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ptrString returns a pointer to a string.
func ptrString(s string) *string {
	return &s
}
