package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/service"
)

// MCPHandler handles MCP configuration and tool execution endpoints.
type MCPHandler struct {
	svc *service.MCPService
}

func NewMCPHandler(svc *service.MCPService) *MCPHandler {
	return &MCPHandler{svc: svc}
}

// List handles GET /workspaces/:workspaceId/mcp
func (h *MCPHandler) List(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	result, err := h.svc.List(workspaceID)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Get handles GET /workspaces/:workspaceId/mcp/:id
func (h *MCPHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := h.svc.Get(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Create handles POST /workspaces/:workspaceId/mcp
func (h *MCPHandler) Create(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("workspaceId"), 10, 64)
	var req service.MCPCreateRequest
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

// Update handles PUT /workspaces/:workspaceId/mcp/:id
func (h *MCPHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.MCPUpdateRequest
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

// Delete handles DELETE /workspaces/:workspaceId/mcp/:id
func (h *MCPHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		appError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "deleted"})
}

// DiscoverTools handles POST /workspaces/:workspaceId/mcp/:id/discover
func (h *MCPHandler) DiscoverTools(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	tools, err := h.svc.DiscoverTools(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, tools)
}

// GetTools handles GET /workspaces/:workspaceId/mcp/:id/tools
func (h *MCPHandler) GetTools(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	tools, err := h.svc.GetTools(id)
	if appError(c, err) {
		return
	}
	c.JSON(200, tools)
}

// ExecuteTool handles POST /workspaces/:workspaceId/mcp/:id/execute
func (h *MCPHandler) ExecuteTool(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.MCPExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.ExecuteTool(id, &req)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}
