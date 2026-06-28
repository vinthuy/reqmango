package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend/internal/common"
	"github.com/reqmanpy/backend/internal/dto/request"
	"github.com/reqmanpy/backend/internal/model"
	"github.com/reqmanpy/backend/internal/service"
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

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
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

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
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

// CopyFromWorkspace handles POST /projects/:projectId/issue-types/copy-from-workspace
func (h *ProjectIssueTypeHandler) CopyFromWorkspace(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
		return
	}

	types, svcErr := h.svc.CopyFromWorkspace(workspaceID, projectID, h.getUserID(c))
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, types)
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
