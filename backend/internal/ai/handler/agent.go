package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/ai/common"
	"github.com/reqmango/backend/internal/ai/service"
	"github.com/reqmango/backend/internal/middleware"
)

// AgentHandler handles HTTP requests for AI Agent management.
type AgentHandler struct {
	svc *service.AgentService
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(svc *service.AgentService) *AgentHandler {
	return &AgentHandler{svc: svc}
}

// appError checks if err is non-nil, writes an appropriate JSON error response,
// and returns true if the error was handled (caller should return early).
func appError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	return true
}

// ==================== CRUD ====================

// List handles GET /workspaces/:wsParam/agents
func (h *AgentHandler) List(c *gin.Context) {
	wsID := h.resolveWorkspaceID(c)
	if wsID == 0 {
		return
	}

	agents, err := h.svc.ListByWorkspace(wsID)
	if appError(c, err) {
		return
	}
	c.JSON(http.StatusOK, agents)
}

// Create handles POST /workspaces/:wsParam/agents
func (h *AgentHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	wsID := h.resolveWorkspaceID(c)
	if wsID == 0 {
		return
	}

	var req service.AgentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	agent, svcErr := h.svc.Create(wsID, user.ID, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusCreated, agent)
}

// Update handles PUT /workspaces/:wsParam/agents/:id
func (h *AgentHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	var req service.AgentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	agent, svcErr := h.svc.Update(agentID, user.ID, &req)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, agent)
}

// Delete handles DELETE /workspaces/:wsParam/agents/:id
func (h *AgentHandler) Delete(c *gin.Context) {
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	if svcErr := h.svc.Delete(agentID); appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted"})
}

// ==================== Agent Actions ====================

// Dispatch handles POST /workspaces/:wsParam/agents/:id/dispatch
func (h *AgentHandler) Dispatch(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	var req struct {
		Task      string  `json:"task" binding:"required"`
		IssueID   *uint64 `json:"issue_id"`
		ProjectID *uint64 `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	wsID := h.resolveWorkspaceID(c)
	ctx := &service.DispatchContext{
		IssueID:     req.IssueID,
		ProjectID:   req.ProjectID,
		WorkspaceID: wsID,
		TriggeredBy: "manual",
	}

	activity, svcErr := h.svc.DispatchAgent(agentID, user.ID, req.Task, ctx)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, activity)
}

// GetActivity handles GET /workspaces/:wsParam/agents/:id/activity
func (h *AgentHandler) GetActivity(c *gin.Context) {
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	activities, svcErr := h.svc.GetActivity(agentID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, activities)
}

// ListWorkspaceActivity handles GET /workspaces/:wsParam/agents/activity
func (h *AgentHandler) ListWorkspaceActivity(c *gin.Context) {
	wsID := h.resolveWorkspaceID(c)
	if wsID == 0 {
		return
	}
	agentIDStr := c.Query("agent_id")
	action := c.Query("action")
	limitStr := c.Query("limit")

	var agentID *uint64
	if agentIDStr != "" {
		id, err := strconv.ParseUint(agentIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent_id"})
			return
		}
		agentID = &id
	}
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	activities, svcErr := h.svc.ListWorkspaceActivity(wsID, agentID, action, limit)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, activities)
}

// AutoTriage handles POST /workspaces/:wsParam/agents/:id/auto-triage
func (h *AgentHandler) AutoTriage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	var req struct {
		IssueID uint64 `json:"issue_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	activity, svcErr := h.svc.AutoTriage(req.IssueID, user.ID)
	if appError(c, svcErr) {
		return
	}
	// Use the provided agent ID for the activity response
	_ = agentID
	c.JSON(http.StatusOK, activity)
}

// AutoAssign handles POST /workspaces/:wsParam/agents/:id/auto-assign
func (h *AgentHandler) AutoAssign(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	var req struct {
		IssueID uint64 `json:"issue_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	activity, svcErr := h.svc.AutoAssign(req.IssueID, user.ID)
	if appError(c, svcErr) {
		return
	}
	_ = agentID
	c.JSON(http.StatusOK, activity)
}

// HandleMention handles POST /issues/:issueId/agents/:agentId/mention
func (h *AgentHandler) HandleMention(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}

	agentID, err := strconv.ParseUint(c.Param("agentId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	var req struct {
		CommentBody string `json:"comment_body" binding:"required"`
		IssueName   string `json:"issue_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	activity, svcErr := h.svc.HandleMention(agentID, 0, user.ID, req.CommentBody, req.IssueName, &issueID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, activity)
}

// GetByID handles GET /workspaces/:wsParam/agents/:id (single agent)
func (h *AgentHandler) GetByID(c *gin.Context) {
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid agent ID"})
		return
	}

	agent, svcErr := h.svc.GetByID(agentID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, agent)
}

// AutoTriageProject handles POST /projects/:projectId/agent/auto-triage (project-level, no agent ID needed).
func (h *AgentHandler) AutoTriageProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var req struct {
		IssueID uint64 `json:"issue_id"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.IssueID > 0 {
		activity, svcErr := h.svc.AutoTriage(req.IssueID, user.ID)
		if appError(c, svcErr) {
			return
		}
		c.JSON(http.StatusOK, activity)
		return
	}

	cfg := struct {
		Action  string `json:"action"`
		IssueID uint64 `json:"issue_id"`
	}{}
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if cfg.IssueID > 0 {
		activity, svcErr := h.svc.AutoTriage(cfg.IssueID, user.ID)
		if appError(c, svcErr) {
			return
		}
		c.JSON(http.StatusOK, activity)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "issue_id is required"})
}

// AutoAssignProject handles POST /projects/:projectId/agent/auto-assign (project-level, no agent ID needed).
func (h *AgentHandler) AutoAssignProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var req struct {
		IssueID uint64 `json:"issue_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	activity, svcErr := h.svc.AutoAssign(req.IssueID, user.ID)
	if appError(c, svcErr) {
		return
	}
	c.JSON(http.StatusOK, activity)
}

// ==================== Helpers ====================

// resolveWorkspaceID extracts the workspace ID from the :wsParam route parameter,
// trying numeric ID first, then looking up by slug via DB (simplified: numeric only for now).
func (h *AgentHandler) resolveWorkspaceID(c *gin.Context) uint64 {
	wsParam := c.Param("wsParam")
	// Try numeric ID
	if id, err := strconv.ParseUint(wsParam, 10, 64); err == nil {
		return id
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace identifier"})
	return 0
}

// UpdateActivityFeedback handles PATCH /workspaces/:wsParam/agents/activity/:id/feedback
func (h *AgentHandler) UpdateActivityFeedback(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid activity ID"})
		return
	}
	var req struct {
		Rating int `json:"rating" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if svcErr := h.svc.UpdateActivityFeedback(activityID, req.Rating); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Feedback recorded"})
}
