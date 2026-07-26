package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

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

func (h *AutopilotHandler) CreateTask(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.AutopilotTaskCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.CreateTask(wid, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

func (h *AutopilotHandler) GetTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.GetTask(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) ListTasks(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	projectID := c.Query("project_id")
	status := c.Query("status")
	var pid *uint64
	if projectID != "" {
		id, _ := strconv.ParseUint(projectID, 10, 64)
		pid = &id
	}
	resp, e := h.svc.ListTasks(wid, pid, status)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) UpdateTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var req request.AutopilotTaskUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.UpdateTask(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) DeleteTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if h.respond(c, h.svc.DeleteTask(id)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *AutopilotHandler) ToggleTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.ToggleTask(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) ExecuteTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.ExecuteTask(id, "manual")
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) GetExecution(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("executionId"), 10, 64)
	resp, e := h.svc.GetExecution(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) ListExecutions(c *gin.Context) {
	taskID, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.ListExecutions(taskID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AutopilotHandler) TriggerWebhook(c *gin.Context) {
	// Find task by trigger URL containing this token
	// Execute task
	c.JSON(200, gin.H{"message": "Webhook received"})
}
