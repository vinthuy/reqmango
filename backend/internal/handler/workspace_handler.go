package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/dto/response"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
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

	// Security Check: Only return workspaces the user is a member of
	var memberships []model.WorkspaceMember
	if err := h.svc.DB().Where("user_id = ? AND is_active = ?", user.ID, true).Preload("Workspace").Find(&memberships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	workspaces := make([]response.WorkspaceResponse, 0, len(memberships))
	for _, m := range memberships {
		resp := response.WorkspaceResponse{
			ID:        m.Workspace.ID,
			Name:      m.Workspace.Name,
			Slug:      m.Workspace.Slug,
			CreatedAt: m.Workspace.CreatedAt,
			UpdatedAt: m.Workspace.UpdatedAt,
		}
		workspaces = append(workspaces, resp)
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

// Get handles GET /workspaces/:id
func (h *WorkspaceHandler) Get(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, ok := h.resolveWsParam(c)
	if !ok {
		return
	}

	// Security Check: Verify user is a member of the workspace
	var member model.WorkspaceMember
	if err := h.svc.DB().Where("workspace_id = ? AND user_id = ? AND is_active = ?", id, user.ID, true).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied: you are not a member of this workspace"})
		return
	}

	resp, svcErr := h.svc.Get(id)
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

// ListMembers handles GET /workspaces/:wsParam/members
func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	workspaceID, ok := h.resolveWsParam(c)
	if !ok {
		return
	}

	members, err := h.svc.ListMembers(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// AddMember handles POST /workspaces/:wsParam/members
func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	workspaceID, ok := h.resolveWsParam(c)
	if !ok {
		return
	}

	var req request.WorkspaceAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	currentUser := middleware.GetCurrentUser(c)

	member, err := h.svc.AddMember(workspaceID, &req, currentUser.ID)
	if err != nil {
		if appErr, ok := err.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// UpdateMember handles PATCH /workspaces/:wsParam/members/:userId?role=N
func (h *WorkspaceHandler) UpdateMember(c *gin.Context) {
	workspaceID, ok := h.resolveWsParam(c)
	if !ok {
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	role, err := strconv.Atoi(c.Query("role"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid role"})
		return
	}

	member, svcErr := h.svc.UpdateMember(workspaceID, userID, role)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// RemoveMember handles DELETE /workspaces/:wsParam/members/:userId
func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	workspaceID, ok := h.resolveWsParam(c)
	if !ok {
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	if svcErr := h.svc.RemoveMember(workspaceID, userID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workspace_id": workspaceID, "user_id": userID, "action": "removed"})
}

// resolveWsParam resolves wsParam (slug or numeric ID) to a workspace ID.
func (h *WorkspaceHandler) resolveWsParam(c *gin.Context) (uint64, bool) {
	param := c.Param("wsParam")
	if id, err := strconv.ParseUint(param, 10, 64); err == nil {
		return id, true
	}

	ws, svcErr := h.svc.GetBySlug(param)
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Workspace not found"})
		return 0, false
	}
	return ws.ID, true
}

// Update handles PATCH /workspaces/:id
func (h *WorkspaceHandler) Update(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
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
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
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
