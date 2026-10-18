package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

// WorkflowHandler handles workflow related requests. It fronts two independent
// models: stateSvc owns `workflows` + `state_transitions` (the state machine),
// agentSvc owns `agent_workflows` + nodes/edges/runs (orchestration).
type WorkflowHandler struct {
	stateSvc *service.StateWorkflowService
	agentSvc *service.WorkflowService
	executor *service.WorkflowExecutor
}

// NewWorkflowHandler creates a new WorkflowHandler.
func NewWorkflowHandler(stateSvc *service.StateWorkflowService, agentSvc *service.WorkflowService, executor *service.WorkflowExecutor) *WorkflowHandler {
	return &WorkflowHandler{
		stateSvc: stateSvc,
		agentSvc: agentSvc,
		executor: executor,
	}
}

// ListWorkflows returns all workflows for a project.
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	workflows, err := h.agentSvc.ListByProject(projectID)
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

	workflow, err := h.agentSvc.Create(projectID, req)
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

	workflow, err := h.agentSvc.Get(workflowID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// GetStateWorkflow returns a workspace state-machine workflow. The workspace
// routes address `workflows`; GetWorkflow reads `agent_workflows`, whose ids
// overlap.
func (h *WorkflowHandler) GetStateWorkflow(c *gin.Context) {
	workspaceID, err := h.stateSvc.ResolveWorkspaceID(c.Param("wsParam"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace"})
		return
	}
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	workflow, err := h.stateSvc.GetStateWorkflow(workspaceID, workflowID)
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

	if err := h.agentSvc.Update(workflowID, req); err != nil {
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

	if err := h.agentSvc.Delete(workflowID); err != nil {
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

	// The body is optional: an empty or malformed body simply means no issue_id.
	_ = c.ShouldBindJSON(&req)

	run, err := h.agentSvc.ExecuteWorkflow(workflowID, req.IssueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Execute in background with the invoking user's identity (for agent permission checks)
	userID := middleware.GetUserID(c)
	go func() { _ = h.executor.Execute(run.ID, userID) }()

	c.JSON(http.StatusAccepted, run)
}

// ListRuns returns all runs for a workflow.
func (h *WorkflowHandler) ListRuns(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	runs, err := h.agentSvc.GetRuns(workflowID)
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

	run, err := h.agentSvc.GetRun(runID)
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

	node, err := h.agentSvc.AddNode(workflowID, req)
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

	if err := h.agentSvc.UpdateNode(nodeID, req); err != nil {
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

	if err := h.agentSvc.DeleteNode(nodeID); err != nil {
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

	edge, err := h.agentSvc.AddEdge(workflowID, req)
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

	if err := h.agentSvc.UpdateEdge(edgeID, req); err != nil {
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

	if err := h.agentSvc.DeleteEdge(edgeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "edge deleted"})
}

// --- Workspace-level state-machine workflow methods ---

// ListWorkspaceWorkflows returns all state-machine workflows for a workspace.
func (h *WorkflowHandler) ListWorkspaceWorkflows(c *gin.Context) {
	workspaceID, err := h.stateSvc.ResolveWorkspaceID(c.Param("wsParam"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace"})
		return
	}

	workflows, err := h.stateSvc.ListStateWorkflowsByWorkspace(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflows)
}

// CreateWorkspaceWorkflow creates a state-machine workflow in a workspace.
func (h *WorkflowHandler) CreateWorkspaceWorkflow(c *gin.Context) {
	workspaceID, err := h.stateSvc.ResolveWorkspaceID(c.Param("wsParam"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace"})
		return
	}

	var req service.CreateStateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.stateSvc.CreateStateWorkflow(workspaceID, nil, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

// UpdateWorkspaceWorkflow updates a workspace state-machine workflow.
func (h *WorkflowHandler) UpdateWorkspaceWorkflow(c *gin.Context) {
	workspaceID, err := h.stateSvc.ResolveWorkspaceID(c.Param("wsParam"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace"})
		return
	}
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req service.UpdateStateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.stateSvc.UpdateStateWorkflow(workspaceID, workflowID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// DeleteWorkspaceWorkflow deletes a workspace state-machine workflow.
func (h *WorkflowHandler) DeleteWorkspaceWorkflow(c *gin.Context) {
	workspaceID, err := h.stateSvc.ResolveWorkspaceID(c.Param("wsParam"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace"})
		return
	}
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	if err := h.stateSvc.DeleteStateWorkflow(workspaceID, workflowID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workflow deleted"})
}

// --- State transition methods ---
//
// Transitions belong to state-machine workflows only. The project route group
// no longer exposes transition CRUD: it addresses agent orchestration, whose
// structure is nodes and edges.

// ListWorkspaceTransitions returns the transitions of a workspace workflow.
func (h *WorkflowHandler) ListWorkspaceTransitions(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	transitions, err := h.stateSvc.ListTransitions(workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transitions)
}

// AddWorkspaceTransition creates a transition for a workspace workflow.
func (h *WorkflowHandler) AddWorkspaceTransition(c *gin.Context) {
	workflowID, err := strconv.ParseUint(c.Param("workflowId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req service.AddTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transition, err := h.stateSvc.AddTransition(workflowID, req)
	if err != nil {
		respondTransitionError(c, err)
		return
	}

	c.JSON(http.StatusCreated, transition)
}

// UpdateWorkspaceTransition updates a workspace workflow transition.
func (h *WorkflowHandler) UpdateWorkspaceTransition(c *gin.Context) {
	transitionID, err := strconv.ParseUint(c.Param("transitionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transition ID"})
		return
	}

	var req service.UpdateTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.stateSvc.UpdateTransition(transitionID, req); err != nil {
		respondTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transition updated"})
}

// DeleteWorkspaceTransition deletes a workspace workflow transition.
func (h *WorkflowHandler) DeleteWorkspaceTransition(c *gin.Context) {
	transitionID, err := strconv.ParseUint(c.Param("transitionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transition ID"})
		return
	}

	if err := h.stateSvc.DeleteTransition(transitionID); err != nil {
		respondTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transition deleted"})
}

// respondTransitionError maps a bad workflow or transition id to 404; anything
// else stays a server error.
func respondTransitionError(c *gin.Context, err error) {
	msg := err.Error()
	if msg == "workflow not found" || msg == "transition not found" {
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
