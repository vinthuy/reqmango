package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

type AgentTaskHandler struct{ svc *service.AgentTaskService }

func NewAgentTaskHandler(svc *service.AgentTaskService) *AgentTaskHandler {
	return &AgentTaskHandler{svc: svc}
}

func (h *AgentTaskHandler) respond(c *gin.Context, err error) bool {
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

func (h *AgentTaskHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *AgentTaskHandler) CreateAgentTask(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.AgentTaskCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Create(wid, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

func (h *AgentTaskHandler) GetAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) ListAgentTasks(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	status := c.Query("status")
	resp, e := h.svc.List(wid, status)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) UpdateAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var req request.AgentTaskUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Update(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) DeleteAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	if h.respond(c, h.svc.Delete(id)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *AgentTaskHandler) ClaimAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var req request.AgentTaskClaim
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body for manual claim from UI
		req = request.AgentTaskClaim{}
	}
	resp, e := h.svc.Claim(id, req.RuntimeID)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) StartAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.Start(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) CompleteAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var req request.AgentTaskComplete
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Complete(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) FailAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	var req request.AgentTaskFail
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Fail(id, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) CancelAgentTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.Cancel(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTaskHandler) GetTaskLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("taskId"), 10, 64)
	resp, e := h.svc.GetLogs(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}
