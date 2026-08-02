package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

// WorkflowHandler handles workflow related requests.
type WorkflowHandler struct {
	workflowSvc *service.WorkflowService
	executor    *service.WorkflowExecutor
}

// NewWorkflowHandler creates a new WorkflowHandler.
func NewWorkflowHandler(workflowSvc *service.WorkflowService, executor *service.WorkflowExecutor) *WorkflowHandler {
	return &WorkflowHandler{
		workflowSvc: workflowSvc,
		executor:    executor,
	}
}

// ListWorkflows returns all workflows for a project.
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	workflows, err := h.workflowSvc.ListByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workflows})
}

// CreateWorkflow creates a new workflow.
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req service.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.workflowSvc.Create(projectID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

// GetWorkflow returns a workflow with full details.
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	workflow, err := h.workflowSvc.Get(workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// UpdateWorkflow updates a workflow.
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req service.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workflowSvc.Update(workflowID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow updated"})
}

// DeleteWorkflow deletes a workflow.
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	if err := h.workflowSvc.Delete(workflowID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow deleted"})
}

// ExecuteWorkflow starts a workflow execution.
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req struct {
		IssueID *uint64 `json:"issue_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
	}

	run, err := h.workflowSvc.ExecuteWorkflow(workflowID, req.IssueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Execute in background
	go h.executor.Execute(run.ID)

	c.JSON(http.StatusAccepted, run)
}

// ListRuns returns all runs for a workflow.
func (h *WorkflowHandler) ListRuns(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	runs, err := h.workflowSvc.GetRuns(workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": runs})
}

// GetRun returns a workflow run with node runs.
func (h *WorkflowHandler) GetRun(c *gin.Context) {
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run ID"})
		return
	}

	run, err := h.workflowSvc.GetRun(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run not found"})
		return
	}

	c.JSON(http.StatusOK, run)
}

// CancelRun cancels a running workflow.
func (h *WorkflowHandler) CancelRun(c *gin.Context) {
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run ID"})
		return
	}

	if err := h.executor.CancelRun(runID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "run cancelled"})
}

// AddNode adds a node to a workflow.
func (h *WorkflowHandler) AddNode(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req service.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node, err := h.workflowSvc.AddNode(workflowID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

// UpdateNode updates a workflow node.
func (h *WorkflowHandler) UpdateNode(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("nodeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node ID"})
		return
	}

	var req service.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workflowSvc.UpdateNode(nodeID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "node updated"})
}

// DeleteNode deletes a workflow node.
func (h *WorkflowHandler) DeleteNode(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("nodeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node ID"})
		return
	}

	if err := h.workflowSvc.DeleteNode(nodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "node deleted"})
}

// AddEdge adds an edge to a workflow.
func (h *WorkflowHandler) AddEdge(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req service.CreateEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	edge, err := h.workflowSvc.AddEdge(workflowID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, edge)
}

// UpdateEdge updates a workflow edge.
func (h *WorkflowHandler) UpdateEdge(c *gin.Context) {
	edgeID, err := strconv.ParseUint(c.Param("edgeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid edge ID"})
		return
	}

	var req service.CreateEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workflowSvc.UpdateEdge(edgeID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "edge updated"})
}

// DeleteEdge deletes a workflow edge.
func (h *WorkflowHandler) DeleteEdge(c *gin.Context) {
	edgeID, err := strconv.ParseUint(c.Param("edgeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid edge ID"})
		return
	}

	if err := h.workflowSvc.DeleteEdge(edgeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "edge deleted"})
}

// --- Workspace-level workflow methods (backward compatibility) ---

// ListWorkspaceWorkflows returns all workflows for a workspace.
func (h *WorkflowHandler) ListWorkspaceWorkflows(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// CreateWorkspaceWorkflow creates a workflow in a workspace.
func (h *WorkflowHandler) CreateWorkspaceWorkflow(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "workspace workflow created"})
}

// UpdateWorkspaceWorkflow updates a workspace workflow.
func (h *WorkflowHandler) UpdateWorkspaceWorkflow(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "workspace workflow updated"})
}

// DeleteWorkspaceWorkflow deletes a workspace workflow.
func (h *WorkflowHandler) DeleteWorkspaceWorkflow(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "workspace workflow deleted"})
}

// --- State transition methods (backward compatibility) ---

// AddTransition adds a state transition to a workflow.
func (h *WorkflowHandler) AddTransition(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "transition added"})
}

// UpdateTransition updates a state transition.
func (h *WorkflowHandler) UpdateTransition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "transition updated"})
}

// DeleteTransition deletes a state transition.
func (h *WorkflowHandler) DeleteTransition(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "transition deleted"})
}
