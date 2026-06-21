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

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
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

	workspaceID, err := strconv.ParseUint(c.Query("workspace_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace ID"})
		return
	}

	includeArchived := c.DefaultQuery("include_archived", "false") == "true"

	modules, total, svcErr := h.svc.List(projectID, workspaceID, includeArchived)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  modules,
		"total": total,
	})
}

func (h *ModuleHandler) Get(c *gin.Context) {
	moduleID, err := strconv.ParseUint(c.Param("moduleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid module ID"})
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

	if svcErr := h.svc.Delete(moduleID); svcErr != nil {
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
			c.JSON(appErr.Code, gin.H{"message": appErr.Message}); return
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
			c.JSON(appErr.Code, gin.H{"message": appErr.Message}); return
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
			c.JSON(appErr.Code, gin.H{"message": appErr.Message}); return
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
			c.JSON(appErr.Code, gin.H{"message": appErr.Message}); return
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
			c.JSON(appErr.Code, gin.H{"message": appErr.Message}); return
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
	resp, svcErr := h.svc.BuildTree(projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message}); return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
