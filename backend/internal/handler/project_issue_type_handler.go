package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

// ProjectIssueTypeHandler handles project-level issue type endpoints.
type ProjectIssueTypeHandler struct {
	svc *service.IssueTypeService
}

// NewProjectIssueTypeHandler creates a new handler.
func NewProjectIssueTypeHandler(svc *service.IssueTypeService) *ProjectIssueTypeHandler {
	return &ProjectIssueTypeHandler{svc: svc}
}

func (h *ProjectIssueTypeHandler) getProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

func (h *ProjectIssueTypeHandler) getUserID(c *gin.Context) uint64 {
	user, exists := c.Get("currentUser")
	if !exists {
		return 0
	}
	if u, ok := user.(*model.User); ok {
		return u.ID
	}
	return 0
}

// ListProjectTypes handles GET /projects/:projectId/issue-types
func (h *ProjectIssueTypeHandler) ListProjectTypes(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if workspaceID == 0 {
		// Resolve workspace_id from the project when not explicitly provided.
		resolved, svcErr := h.svc.ResolveWorkspaceID(projectID)
		if svcErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
			return
		}
		workspaceID = resolved
	}

	// List types scoped to this project OR workspace-shared
	types, svcErr := h.svc.List(workspaceID, &projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, types)
}

// CreateProjectType handles POST /projects/:projectId/issue-types
func (h *ProjectIssueTypeHandler) CreateProjectType(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if workspaceID == 0 {
		resolved, svcErr := h.svc.ResolveWorkspaceID(projectID)
		if svcErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
			return
		}
		workspaceID = resolved
	}

	var req request.IssueTypeCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	req.ProjectID = &projectID

	result, svcErr := h.svc.Create(workspaceID, h.getUserID(c), req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Reorder handles PATCH /projects/:projectId/issue-types/reorder
func (h *ProjectIssueTypeHandler) Reorder(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var body struct {
		TypeIDs []uint64 `json:"type_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if svcErr := h.svc.Reorder(projectID, body.TypeIDs); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Issue types reordered"})
}

// ==================== Plane v3-style Import Model ====================

// ListImportable handles GET /projects/:projectId/issue-types/importable
// Returns workspace-level types the project has NOT yet imported.
func (h *ProjectIssueTypeHandler) ListImportable(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	workspaceID, _ := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if workspaceID == 0 {
		resolved, svcErr := h.svc.ResolveWorkspaceID(projectID)
		if svcErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
			return
		}
		workspaceID = resolved
	}

	types, svcErr := h.svc.ListImportable(workspaceID, projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, types)
}

// ImportType handles POST /projects/:projectId/issue-types/:typeId/import
// Records a project's reference to a workspace-level type (Plane v3 Import).
func (h *ProjectIssueTypeHandler) ImportType(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	typeID, err := strconv.ParseUint(c.Param("typeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	if svcErr := h.svc.ImportType(projectID, typeID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Issue type imported"})
}

// UnimportType handles DELETE /projects/:projectId/issue-types/:typeId/import
// Removes a project's reference to a workspace-level type.
func (h *ProjectIssueTypeHandler) UnimportType(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	typeID, err := strconv.ParseUint(c.Param("typeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue type ID"})
		return
	}

	if svcErr := h.svc.UnimportType(projectID, typeID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Issue type unimported"})
}
