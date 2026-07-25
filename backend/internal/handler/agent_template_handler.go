package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

type AgentTemplateHandler struct{ svc *service.AgentTemplateService }

func NewAgentTemplateHandler(svc *service.AgentTemplateService) *AgentTemplateHandler {
	return &AgentTemplateHandler{svc: svc}
}

func (h *AgentTemplateHandler) respond(c *gin.Context, err error) bool {
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

func (h *AgentTemplateHandler) parseWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

func (h *AgentTemplateHandler) CreateAgentTemplate(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	var req request.AgentTemplateCreate
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

func (h *AgentTemplateHandler) GetAgentTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("templateId"), 10, 64)
	resp, e := h.svc.Get(id)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTemplateHandler) ListAgentTemplates(c *gin.Context) {
	wid, _ := h.parseWorkspaceID(c)
	resp, e := h.svc.List(wid)
	if h.respond(c, e) {
		return
	}
	c.JSON(200, resp)
}

func (h *AgentTemplateHandler) UpdateAgentTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("templateId"), 10, 64)
	var req request.AgentTemplateUpdate
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

func (h *AgentTemplateHandler) DeleteAgentTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("templateId"), 10, 64)
	if h.respond(c, h.svc.Delete(id)) {
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}
