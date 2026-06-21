package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/common"
	"github.com/reqmanpy/backend-go/internal/dto/request"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/service"
)

type WorkspaceHandler struct {
	svc *service.WorkspaceService
}

func NewWorkspaceHandler(svc *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{svc: svc}
}

// List handles GET /workspaces/
func (h *WorkspaceHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaces, err := h.svc.ListByUser(user.ID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, workspaces)
}

// Create handles POST /workspaces/
func (h *WorkspaceHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var req request.WorkspaceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := h.svc.Create(&req, user.ID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Get handles GET /workspaces/:wsParam (could be slug or numeric ID)
func (h *WorkspaceHandler) Get(c *gin.Context) {
	param := c.Param("wsParam")

	// Try to parse as numeric ID first
	if id, err := strconv.ParseUint(param, 10, 64); err == nil {
		resp, svcErr := h.svc.GetByID(id)
		if svcErr == nil {
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	// Fall back to slug lookup
	resp, err := h.svc.GetBySlug(param)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListMembers handles GET /workspaces/:wsParam/members
func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	param := c.Param("wsParam")
	workspaceID, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		// try slug lookup
		ws, svcErr := h.svc.GetBySlug(param)
		if svcErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Workspace not found"})
			return
		}
		workspaceID = ws.ID
	}

	members, err := h.svc.ListMembers(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// Update handles PATCH /workspaces/:id
func (h *WorkspaceHandler) Update(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	var req request.WorkspaceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Update(workspaceID, &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /workspaces/:id
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	if svcErr := h.svc.Delete(workspaceID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
