package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

// AutopilotHandler exposes the Autopilot HTTP endpoints (PRD P4-008).
//
// Task management routes are grouped under /workspaces/:wsParam/autopilot-tasks
// and require workspace admin. The webhook trigger route is public
// (/api/v1/autopilot/webhook/:token) so external systems can invoke it
// without authentication.
type AutopilotHandler struct{ svc *service.AutopilotService }

func NewAutopilotHandler(svc *service.AutopilotService) *AutopilotHandler {
	return &AutopilotHandler{svc: svc}
}

func (h *AutopilotHandler) respond(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*common.AppError); ok {
		c.JSON(ae.Code, gin.H{"message": ae.Message})
		return true
	}
	c.JSON(500, gin.H{"message": "Internal server error"})
	return true
}

func (h *AutopilotHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *AutopilotHandler) parseTaskID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("taskId"), 10, 64)
}

func (h *AutopilotHandler) parseExecutionID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("executionId"), 10, 64)
}

// callerID extracts the authenticated user id from the context.
func (h *AutopilotHandler) callerID(c *gin.Context) uint64 {
	return middleware.GetCurrentUser(c).ID
}

// CreateTask handles POST /workspaces/:wsParam/autopilot-tasks
func (h *AutopilotHandler) CreateTask(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var req request.AutopilotTaskCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.CreateTask(wid, h.callerID(c), req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

// GetTask handles GET /workspaces/:wsParam/autopilot-tasks/:taskId
func (h *AutopilotHandler) GetTask(c *gin.Context) {
	id, err := h.parseTaskID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid task id"})
		return
	}
	resp, e := h.svc.GetTask(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// ListTasks handles GET /workspaces/:wsParam/autopilot-tasks
// Query params: project_id (optional), status (optional)
func (h *AutopilotHandler) ListTasks(c *gin.Context) {
	wid, err := h.parseWorkspaceID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid workspace id"})
		return
	}
	var pid *uint64
	if projectID := c.Query("project_id"); projectID != "" {
		if v, perr := strconv.ParseUint(projectID, 10, 64); perr == nil {
			pid = &v
		}
	}
	resp, e := h.svc.ListTasks(wid, pid, c.Query("status"))
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// UpdateTask handles PUT /workspaces/:wsParam/autopilot-tasks/:taskId
func (h *AutopilotHandler) UpdateTask(c *gin.Context) {
	id, err := h.parseTaskID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid task id"})
		return
	}
	var req request.AutopilotTaskUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.UpdateTask(id, h.callerID(c), req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// DeleteTask handles DELETE /workspaces/:wsParam/autopilot-tasks/:taskId
func (h *AutopilotHandler) DeleteTask(c *gin.Context) {
	id, err := h.parseTaskID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid task id"})
		return
	}
	if h.respond(c, h.svc.DeleteTask(id, h.callerID(c))) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

// ToggleTask handles POST /workspaces/:wsParam/autopilot-tasks/:taskId/toggle
func (h *AutopilotHandler) ToggleTask(c *gin.Context) {
	id, err := h.parseTaskID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid task id"})
		return
	}
	resp, e := h.svc.ToggleTask(id, h.callerID(c))
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// ExecuteTask handles POST /workspaces/:wsParam/autopilot-tasks/:taskId/execute
func (h *AutopilotHandler) ExecuteTask(c *gin.Context) {
	id, err := h.parseTaskID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid task id"})
		return
	}
	resp, e := h.svc.ExecuteTask(id, h.callerID(c), "manual")
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// GetExecution handles GET /workspaces/:wsParam/autopilot-tasks/:taskId/executions/:executionId
func (h *AutopilotHandler) GetExecution(c *gin.Context) {
	id, err := h.parseExecutionID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid execution id"})
		return
	}
	resp, e := h.svc.GetExecution(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// ListExecutions handles GET /workspaces/:wsParam/autopilot-tasks/:taskId/executions
func (h *AutopilotHandler) ListExecutions(c *gin.Context) {
	taskID, err := h.parseTaskID(c)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid task id"})
		return
	}
	resp, e := h.svc.ListExecutions(taskID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

// TriggerWebhook handles POST /api/v1/autopilot/webhook/:token (public).
// External systems invoke this URL to fire a webhook-triggered autopilot task.
func (h *AutopilotHandler) TriggerWebhook(c *gin.Context) {
	token := c.Param("token")
	resp, e := h.svc.TriggerWebhook(token)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}
