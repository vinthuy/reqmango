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

type ModuleHandler struct {
	svc *service.ModuleService
}

func NewModuleHandler(svc *service.ModuleService) *ModuleHandler {
	return &ModuleHandler{svc: svc}
}

func (h *ModuleHandler) parseModuleID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("moduleId"), 10, 64)
}

func (h *ModuleHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var req request.ModuleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	workspaceID := req.WorkspaceID
	if workspaceID == 0 {
		id, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
			return
		}
		workspaceID = id
	}

	module, svcErr := h.svc.Create(workspaceID, user.ID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, module)
}

func (h *ModuleHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var workspaceID uint64
	if wsStr := c.Query("workspace_id"); wsStr != "" {
		workspaceID, err = strconv.ParseUint(wsStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
			return
		}
	} else {
		var project model.Project
		if err := h.svc.DB().First(&project, projectID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Project not found"})
			return
		}
		workspaceID = project.WorkspaceID
	}

	includeArchived := c.DefaultQuery("include_archived", "false") == "true"

	modules, _, svcErr := h.svc.List(projectID, workspaceID, includeArchived)
	if svcErr != nil {
		fmt.Printf("[ModuleHandler.List ERROR] projectID=%d workspaceID=%d includeArchived=%v err=%v\n", projectID, workspaceID, includeArchived, svcErr)
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message, "detail": appErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error", "detail": svcErr.Error()})
		return
	}

	c.JSON(http.StatusOK, modules)
}

func (h *ModuleHandler) Search(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	var workspaceID uint64
	if wsStr := c.Query("workspace_id"); wsStr != "" {
		workspaceID, err = strconv.ParseUint(wsStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
			return
		}
	} else {
		var project model.Project
		if err := h.svc.DB().First(&project, projectID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Project not found"})
			return
		}
		workspaceID = project.WorkspaceID
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "query parameter 'q' is required"})
		return
	}
	modules, svcErr := h.svc.Search(projectID, workspaceID, query)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": modules})
}

func (h *ModuleHandler) Get(c *gin.Context) {
	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
			return
		}
		module, svcErr := h.svc.GetWithProjectContext(moduleID, projectID)
		if svcErr != nil {
			if appErr, ok := svcErr.(*common.AppError); ok {
				c.JSON(appErr.Code, gin.H{"message": appErr.Message})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}
		c.JSON(http.StatusOK, module)
		return
	}

	module, svcErr := h.svc.Get(moduleID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, module)
}

func (h *ModuleHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	var req request.ModuleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	module, svcErr := h.svc.Update(moduleID, user.ID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, module)
}

func (h *ModuleHandler) Delete(c *gin.Context) {
	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	if svcErr := h.svc.Delete(moduleID, middleware.GetCurrentUser(c).ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module deleted successfully"})
}

// ==================== Issue Association ====================

func (h *ModuleHandler) AddIssue(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	issueID, err := strconv.ParseUint(c.Query("issue_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue_id"})
		return
	}
	if svcErr := h.svc.AddIssue(moduleID, issueID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_id": moduleID, "issue_id": issueID, "action": "added"})
}

func (h *ModuleHandler) RemoveIssue(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
		return
	}
	if svcErr := h.svc.RemoveIssue(moduleID, issueID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"module_id": moduleID, "issue_id": issueID, "action": "removed"})
}

func (h *ModuleHandler) ListIssues(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	p := common.ParsePagination(c.Query("limit"), c.Query("offset"), 50, 100)
	var stateID *uint64
	if v := c.Query("state_id"); v != "" {
		if id, e := strconv.ParseUint(v, 10, 64); e == nil {
			stateID = &id
		}
	}
	issues, _, svcErr := h.svc.ListIssues(moduleID, stateID, c.Query("priority"), p.Limit, p.Offset)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, issues)
}

// ==================== Analysis ====================

func (h *ModuleHandler) GetProgress(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	resp, svcErr := h.svc.GetProgress(moduleID)
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

func (h *ModuleHandler) GetStatistics(c *gin.Context) {
	moduleID, err := h.parseModuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}
	resp, svcErr := h.svc.GetStatistics(moduleID)
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

func (h *ModuleHandler) GetTree(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Query("project_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project_id"})
		return
	}

	var workspaceID uint64
	if wsStr := c.Query("workspace_id"); wsStr != "" {
		workspaceID, err = strconv.ParseUint(wsStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace_id"})
			return
		}
	} else {
		var project model.Project
		if err := h.svc.DB().First(&project, projectID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Project not found"})
			return
		}
		workspaceID = project.WorkspaceID
	}

	resp, svcErr := h.svc.BuildTree(projectID, workspaceID)
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

// ==================== Inheritance Override ====================

func (h *ModuleHandler) CreateOrUpdateOverride(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	var req request.ModuleOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	module, svcErr := h.svc.CreateOrUpdateOverride(projectID, moduleID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, module)
}

func (h *ModuleHandler) DeleteOverride(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	if svcErr := h.svc.DeleteOverride(projectID, moduleID, middleware.GetCurrentUser(c).ID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Override deleted successfully"})
}

// ==================== Workspace Modules ====================

func (h *ModuleHandler) ListWorkspaceModules(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	modules, svcErr := h.svc.ListWorkspaceModules(workspaceID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, modules)
}

func (h *ModuleHandler) CreateWorkspaceModule(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	var req request.ModuleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	module, svcErr := h.svc.CreateWorkspaceModule(workspaceID, user.ID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, module)
}

func (h *ModuleHandler) GetWorkspaceModule(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	module, svcErr := h.svc.GetWorkspaceModule(workspaceID, moduleID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, module)
}

func (h *ModuleHandler) UpdateWorkspaceModule(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	var req request.ModuleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	module, svcErr := h.svc.UpdateWorkspaceModule(moduleID, user.ID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, module)
}

func (h *ModuleHandler) DeleteWorkspaceModule(c *gin.Context) {
	workspaceID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
		return
	}

	if svcErr := h.svc.DeleteWorkspaceModule(workspaceID, moduleID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace module deleted successfully"})
}
