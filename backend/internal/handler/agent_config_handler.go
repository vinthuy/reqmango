package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

type AgentConfigHandler struct{ svc *service.AgentConfigService }

func NewAgentConfigHandler(svc *service.AgentConfigService) *AgentConfigHandler {
	return &AgentConfigHandler{svc: svc}
}

func (h *AgentConfigHandler) respond(c *gin.Context, err error) bool {
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

func (h *AgentConfigHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *AgentConfigHandler) CreateAgentConfig(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.AgentConfigCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Create(wid, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(201, resp)
}

func (h *AgentConfigHandler) GetAgentConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentConfigHandler) ListAgentConfigs(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.List(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentConfigHandler) UpdateAgentConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	var req request.AgentConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}
	resp, e := h.svc.Update(id, middleware.GetCurrentUser(c).ID, req)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentConfigHandler) DeleteAgentConfig(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("configId"), 10, 64)
	if h.respond(c, h.svc.Delete(id, middleware.GetCurrentUser(c).ID)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *AgentConfigHandler) GetDefaultAgentConfig(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.GetDefault(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}
