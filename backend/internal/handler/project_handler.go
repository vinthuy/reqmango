package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

type ProjectHandler struct {
	svc        *service.ProjectService
	templateSvc *service.ProjectTemplateService
}

func NewProjectHandler(svc *service.ProjectService, templateSvc *service.ProjectTemplateService) *ProjectHandler {
	return &ProjectHandler{svc: svc, templateSvc: templateSvc}
}

// Create handles POST /projects/?workspace_id=int&template_id=int
func (h *ProjectHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	// Security Check: Verify user is a member of the workspace
	var member model.WorkspaceMember
	if err := h.svc.DB().Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, user.ID, true).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied: you are not a member of this workspace"})
		return
	}

	var req request.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// template_id from query param takes precedence over body
	if templateIDStr := c.Query("template_id"); templateIDStr != "" {
		if tid, err := strconv.ParseUint(templateIDStr, 10, 64); err == nil {
			req.TemplateID = &tid
		}
	}

	resp, svcErr := h.svc.Create(&req, workspaceID, user.ID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	// Auto-apply template after creation (non-fatal on failure)
	if req.TemplateID != nil {
		if svcErr := h.templateSvc.Apply(*req.TemplateID, resp.ID); svcErr != nil {
			// Log warning but don't fail the request — project is already created
			fmt.Printf("WARNING: project %d created but template %d apply failed: %v\n", resp.ID, *req.TemplateID, svcErr)
		}
	}

	c.JSON(http.StatusCreated, resp)
}

// List handles GET /projects/?workspace_id=int
func (h *ProjectHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	// Security Check: Verify user is a member of the workspace
	var member model.WorkspaceMember
	if err := h.svc.DB().Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, user.ID, true).First(&member).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": "Access denied: you are not a member of this workspace"})
		return
	}

	includeArchived := c.Query("include_archived") == "true"

	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)

	projects, svcErr := h.svc.ListByWorkspace(workspaceID, includeArchived, p.Limit, p.Offset)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// Get handles GET /projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	resp, svcErr := h.svc.GetByID(projectID)
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

// Update handles PATCH /projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var req request.ProjectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.Update(projectID, &req)
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

// Delete handles DELETE /projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	if svcErr := h.svc.Delete(projectID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Archive handles POST /projects/:id/archive
func (h *ProjectHandler) Archive(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	resp, svcErr := h.svc.Archive(projectID)
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

// Restore handles POST /projects/:id/restore
func (h *ProjectHandler) Restore(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	resp, svcErr := h.svc.Restore(projectID)
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

// ListMembers handles GET /projects/:id/members?only_active=true
func (h *ProjectHandler) ListMembers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	onlyActive := c.DefaultQuery("only_active", "true") == "true"

	members, svcErr := h.svc.ListMembers(projectID, onlyActive)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// AddMember handles POST /projects/:id/members?user_id=int&role=int
func (h *ProjectHandler) AddMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user_id"})
		return
	}

	role := 15
	if roleStr := c.Query("role"); roleStr != "" {
		if r, err := strconv.Atoi(roleStr); err == nil && r >= 1 && r <= 20 {
			role = r
		}
	}

	currentUser := middleware.GetCurrentUser(c)

	member, svcErr := h.svc.AddMember(projectID, userID, role, currentUser.ID)
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

// UpdateMember handles PATCH /projects/:id/members/:userId?role=int
func (h *ProjectHandler) UpdateMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
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

	currentUser := middleware.GetCurrentUser(c)
	member, svcErr := h.svc.UpdateMember(projectID, currentUser.ID, userID, role)
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

// RemoveMember handles DELETE /projects/:id/members/:userId
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	currentUser := middleware.GetCurrentUser(c)
	if svcErr := h.svc.RemoveMember(projectID, currentUser.ID, userID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project_id": projectID, "user_id": userID, "action": "removed"})
}

// GetStatistics handles GET /projects/:id/statistics
func (h *ProjectHandler) GetStatistics(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	stats, svcErr := h.svc.GetStatistics(projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateProjectLead handles PATCH /projects/:id/lead?user_id=int
func (h *ProjectHandler) UpdateProjectLead(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var userID *uint64
	if uidStr := c.Query("user_id"); uidStr != "" {
		uid, err := strconv.ParseUint(uidStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user_id"})
			return
		}
		userID = &uid
	}

	resp, svcErr := h.svc.UpdateProjectLead(projectID, userID)
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

// ListSubscribers handles GET /projects/:id/subscribers
func (h *ProjectHandler) ListSubscribers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	subs, svcErr := h.svc.ListSubscribers(projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// AddSubscriber handles POST /projects/:id/subscribers?user_id=int
func (h *ProjectHandler) AddSubscriber(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user_id"})
		return
	}

	sub, svcErr := h.svc.AddSubscriber(projectID, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// RemoveSubscriber handles DELETE /projects/:id/subscribers/:userId
func (h *ProjectHandler) RemoveSubscriber(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	if svcErr := h.svc.RemoveSubscriber(projectID, userID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project_id": projectID, "user_id": userID, "action": "removed"})
}

// GetIssuesSummary handles GET /projects/:id/issues-summary
func (h *ProjectHandler) GetIssuesSummary(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	summary, svcErr := h.svc.GetIssuesSummary(projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
