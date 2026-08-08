package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/service"
)

// MCPSyncResult represents the result of syncing MCP tools.
type MCPSyncResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
}

// MCPHandler handles MCP configuration and tool execution endpoints.
type MCPHandler struct {
	svc *service.MCPService
}

func NewMCPHandler(svc *service.MCPService) *MCPHandler {
	return &MCPHandler{svc: svc}
}

// List handles GET /workspaces/:wsParam/mcp
func (h *MCPHandler) List(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
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

// Create handles POST /workspaces/:wsParam/mcp
func (h *MCPHandler) Create(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	var req service.MCPCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.Create(workspaceID, middleware.GetCurrentUser(c).ID, &req)
	if appError(c, err) {
		return
	}
	c.JSON(201, result)
}

// Update handles PUT /workspaces/:wsParam/mcp/:id
func (h *MCPHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.MCPUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	result, err := h.svc.Update(id, middleware.GetCurrentUser(c).ID, &req)
	if appError(c, err) {
		return
	}
	c.JSON(200, result)
}

// Delete handles DELETE /workspaces/:wsParam/mcp/:id
func (h *MCPHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id, middleware.GetCurrentUser(c).ID); err != nil {
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

// SyncTools handles POST /workspaces/:wsParam/mcp/:id/sync
func (h *MCPHandler) SyncTools(c *gin.Context) {
	workspaceID, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	added, updated, err := h.svc.SyncTools(workspaceID, id, middleware.GetCurrentUser(c).ID)
	if appError(c, err) {
		return
	}
	c.JSON(200, MCPSyncResult{Added: added, Updated: updated})
}
