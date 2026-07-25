package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

type ToolHandler struct {
	service *service.ToolService
}

func NewToolHandler(db *gorm.DB) *ToolHandler {
	return &ToolHandler{
		service: service.NewToolService(db),
	}
}

func (h *ToolHandler) respond(c *gin.Context, err error) bool {
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

func (h *ToolHandler) parseWorkspaceID(c *gin.Context) uint64 {
	wid, _ := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	return wid
}

// CreateTool creates a new tool
func (h *ToolHandler) CreateTool(c *gin.Context) {
	wid := h.parseWorkspaceID(c)

	var req request.CreateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}

	result, err := h.service.Create(wid, req)
	if h.respond(c, err) {
		return
	}

	c.JSON(201, result)
}

// GetTool retrieves a tool by ID
func (h *ToolHandler) GetTool(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("toolId"), 10, 64)

	result, err := h.service.Get(id)
	if h.respond(c, err) {
		return
	}

	c.JSON(200, result)
}

// ListTools retrieves all tools for a workspace
func (h *ToolHandler) ListTools(c *gin.Context) {
	wid := h.parseWorkspaceID(c)

	result, err := h.service.List(wid)
	if h.respond(c, err) {
		return
	}

	c.JSON(200, result)
}

// UpdateTool updates a tool
func (h *ToolHandler) UpdateTool(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("toolId"), 10, 64)

	var req request.UpdateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}

	result, err := h.service.Update(id, req)
	if h.respond(c, err) {
		return
	}

	c.JSON(200, result)
}

// DeleteTool deletes a tool
func (h *ToolHandler) DeleteTool(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("toolId"), 10, 64)

	if err := h.service.Delete(id); err != nil {
		h.respond(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "Tool deleted successfully"})
}

// CallTool executes a tool
func (h *ToolHandler) CallTool(c *gin.Context) {
	wid := h.parseWorkspaceID(c)

	var req request.CallToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid body"})
		return
	}

	result, err := h.service.Call(wid, req)
	if h.respond(c, err) {
		return
	}

	c.JSON(200, result)
}

// GetToolCallLogs retrieves tool call logs for a workspace
func (h *ToolHandler) GetToolCallLogs(c *gin.Context) {
	wid := h.parseWorkspaceID(c)

	result, err := h.service.GetCallLogs(wid)
	if h.respond(c, err) {
		return
	}

	c.JSON(200, result)
}
