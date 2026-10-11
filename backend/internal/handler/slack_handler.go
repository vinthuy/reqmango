package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

type SlackHandler struct {
	svc *service.SlackService
}

func NewSlackHandler(svc *service.SlackService) *SlackHandler {
	return &SlackHandler{svc: svc}
}

// List handles GET /workspaces/:wsParam/slack
func (h *SlackHandler) List(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	result, err := h.svc.List(workspaceID)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Get handles GET /workspaces/:wsParam/slack/:id
func (h *SlackHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := h.svc.Get(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Create handles POST /workspaces/:wsParam/slack
func (h *SlackHandler) Create(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid workspace"})
		return
	}
	var req service.SlackCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.Create(workspaceID, &req)
	if appError(c, err) {
		return
	}
	c.JSON(201, result)
}

// Update handles PUT /workspaces/:wsParam/slack/:id
func (h *SlackHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.SlackUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.Update(id, &req)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Delete handles DELETE /workspaces/:wsParam/slack/:id
func (h *SlackHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		appError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "deleted"})
}

// SendNotification handles POST /workspaces/:wsParam/slack/:id/notify
func (h *SlackHandler) SendNotification(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.SlackNotification
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if err := h.svc.SendNotification(id, &req); err != nil {
		appError(c, err)
		return
	}
	c.JSON(200, gin.H{"status": "sent"})
}

// TestNotification handles POST /workspaces/:wsParam/slack/:id/test
func (h *SlackHandler) TestNotification(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := h.svc.TestNotification(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}
