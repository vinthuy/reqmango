package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/service"
)

type ProjectSettingsHandler struct {
	svc *service.ProjectSettingsService
}

func NewProjectSettingsHandler(svc *service.ProjectSettingsService) *ProjectSettingsHandler {
	return &ProjectSettingsHandler{svc: svc}
}

func (h *ProjectSettingsHandler) getProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

func (h *ProjectSettingsHandler) getWorkspaceID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("wsParam"), 10, 64)
}

// ==================== States ====================

// ListWorkspaceStates handles GET /workspaces/:wsParam/settings/states
func (h *ProjectSettingsHandler) ListWorkspaceStates(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	states, svcErr := h.svc.ListWorkspaceStates(workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, states)
}

// CreateWorkspaceState handles POST /workspaces/:wsParam/settings/states
func (h *ProjectSettingsHandler) CreateWorkspaceState(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	var req request.StateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.CreateWorkspaceState(&req, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetWorkspaceState handles GET /workspaces/:wsParam/settings/states/:stateId
func (h *ProjectSettingsHandler) GetWorkspaceState(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	stateID, err := strconv.ParseUint(c.Param("stateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid state ID"})
		return
	}

	resp, svcErr := h.svc.GetWorkspaceState(workspaceID, stateID)
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

// UpdateWorkspaceState handles PUT /workspaces/:wsParam/settings/states/:stateId
func (h *ProjectSettingsHandler) UpdateWorkspaceState(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	stateID, err := strconv.ParseUint(c.Param("stateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid state ID"})
		return
	}

	var req request.StateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.UpdateWorkspaceState(workspaceID, stateID, &req)
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

// DeleteWorkspaceState handles DELETE /workspaces/:wsParam/settings/states/:stateId
func (h *ProjectSettingsHandler) DeleteWorkspaceState(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	stateID, err := strconv.ParseUint(c.Param("stateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid state ID"})
		return
	}

	if svcErr := h.svc.DeleteWorkspaceState(workspaceID, stateID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CreateState handles POST /projects/:id/settings/states?workspace_id=int
func (h *ProjectSettingsHandler) CreateState(c *gin.Context) {
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

	var req request.StateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.CreateState(&req, projectID, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListStates handles GET /projects/:id/settings/states?include_inactive=false&workspace_id=int
func (h *ProjectSettingsHandler) ListStates(c *gin.Context) {
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

	includeInactive := c.Query("include_inactive") == "true"

	states, svcErr := h.svc.ListStates(projectID, workspaceID, includeInactive)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, states)
}

// GetState handles GET /projects/:id/settings/states/:stateId
func (h *ProjectSettingsHandler) GetState(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	stateID, err := strconv.ParseUint(c.Param("stateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid state ID"})
		return
	}

	resp, svcErr := h.svc.GetState(projectID, stateID)
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

// UpdateState handles PUT /projects/:id/settings/states/:stateId
func (h *ProjectSettingsHandler) UpdateState(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	stateID, err := strconv.ParseUint(c.Param("stateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid state ID"})
		return
	}

	var req request.StateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.UpdateState(projectID, stateID, &req)
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

// DeleteState handles DELETE /projects/:id/settings/states/:stateId
func (h *ProjectSettingsHandler) DeleteState(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	stateID, err := strconv.ParseUint(c.Param("stateId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid state ID"})
		return
	}

	if svcErr := h.svc.DeleteState(projectID, stateID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			fmt.Println("DeleteState error:", appErr.Message, "detail:", appErr.Detail)
			c.JSON(appErr.Code, gin.H{"message": appErr.Message, "detail": appErr.Detail, "error_code": appErr.ErrorCode})
			return
		}
		fmt.Println("DeleteState unknown error:", svcErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CreateDefaultStates handles POST /projects/:id/settings/states/default?workspace_id=int
func (h *ProjectSettingsHandler) CreateDefaultStates(c *gin.Context) {
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

	states, svcErr := h.svc.CreateDefaultStates(projectID, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, states)
}

// ==================== Labels ====================

// CreateLabel handles POST /projects/:id/settings/labels?workspace_id=int
func (h *ProjectSettingsHandler) CreateLabel(c *gin.Context) {
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

	var req request.LabelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.CreateLabel(&req, projectID, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListLabels handles GET /projects/:id/settings/labels?workspace_id=int
func (h *ProjectSettingsHandler) ListLabels(c *gin.Context) {
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

	labels, svcErr := h.svc.ListLabels(projectID, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, labels)
}

// SearchLabels handles GET /projects/:id/settings/labels/search?q=xxx&workspace_id=int
func (h *ProjectSettingsHandler) SearchLabels(c *gin.Context) {
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

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "query parameter 'q' is required"})
		return
	}
	labels, svcErr := h.svc.SearchLabels(projectID, workspaceID, query)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": labels})
}

// GetLabel handles GET /projects/:id/settings/labels/:labelId
func (h *ProjectSettingsHandler) GetLabel(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	resp, svcErr := h.svc.GetLabel(projectID, labelID)
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

// UpdateLabel handles PUT /projects/:id/settings/labels/:labelId
func (h *ProjectSettingsHandler) UpdateLabel(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	var req request.LabelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.UpdateLabel(projectID, labelID, &req)
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

// DeleteLabel handles DELETE /projects/:id/settings/labels/:labelId
func (h *ProjectSettingsHandler) DeleteLabel(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	if svcErr := h.svc.DeleteLabel(projectID, labelID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ==================== Workspace Labels ====================

// ListWorkspaceLabels handles GET /workspaces/:wsParam/settings/labels
func (h *ProjectSettingsHandler) ListWorkspaceLabels(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	labels, svcErr := h.svc.ListWorkspaceLabels(workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, labels)
}

// CreateWorkspaceLabel handles POST /workspaces/:wsParam/settings/labels
func (h *ProjectSettingsHandler) CreateWorkspaceLabel(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	var req request.LabelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.CreateWorkspaceLabel(&req, workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetWorkspaceLabel handles GET /workspaces/:wsParam/settings/labels/:labelId
func (h *ProjectSettingsHandler) GetWorkspaceLabel(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	resp, svcErr := h.svc.GetWorkspaceLabel(workspaceID, labelID)
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

// UpdateWorkspaceLabel handles PUT /workspaces/:wsParam/settings/labels/:labelId
func (h *ProjectSettingsHandler) UpdateWorkspaceLabel(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	var req request.LabelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, svcErr := h.svc.UpdateWorkspaceLabel(workspaceID, labelID, &req)
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

// DeleteWorkspaceLabel handles DELETE /workspaces/:wsParam/settings/labels/:labelId
func (h *ProjectSettingsHandler) DeleteWorkspaceLabel(c *gin.Context) {
	workspaceID, err := h.getWorkspaceID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	labelID, err := strconv.ParseUint(c.Param("labelId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid label ID"})
		return
	}

	if svcErr := h.svc.DeleteWorkspaceLabel(workspaceID, labelID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
