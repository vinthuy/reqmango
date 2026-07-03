package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
)

// PageHandler handles page HTTP endpoints.
type PageHandler struct {
	svc *service.PageService
}

// NewPageHandler creates a new PageHandler.
func NewPageHandler(svc *service.PageService) *PageHandler {
	return &PageHandler{svc: svc}
}

func (h *PageHandler) getProjectID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("projectId"), 10, 64)
}

func getUserIDFromContext(c *gin.Context) uint64 {
	user, exists := c.Get("currentUser")
	if !exists {
		return 0
	}
	if u, ok := user.(*model.User); ok {
		return u.ID
	}
	return 0
}

func (h *PageHandler) getUserID(c *gin.Context) uint64 {
	return getUserIDFromContext(c)
}

// List handles GET /projects/:projectId/pages
func (h *PageHandler) List(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	includeArchived := c.Query("include_archived") == "true"

	pages, svcErr := h.svc.List(projectID, includeArchived)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, pages)
}

// GetTree handles GET /projects/:projectId/pages/tree
func (h *PageHandler) GetTree(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	tree, svcErr := h.svc.GetTree(projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// Get handles GET /projects/:projectId/pages/:pageId
func (h *PageHandler) Get(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	page, svcErr := h.svc.Get(pageID, projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, page)
}

// Create handles POST /projects/:projectId/pages
func (h *PageHandler) Create(c *gin.Context) {
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

	var req request.PageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	page, svcErr := h.svc.Create(&req, projectID, workspaceID, h.getUserID(c))
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, page)
}

// Update handles PUT /projects/:projectId/pages/:pageId
func (h *PageHandler) Update(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	var req request.PageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	page, svcErr := h.svc.Update(pageID, projectID, h.getUserID(c), &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, page)
}

// Delete handles DELETE /projects/:projectId/pages/:pageId
func (h *PageHandler) Delete(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	if svcErr := h.svc.Delete(pageID, projectID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Page deleted"})
}

// Archive handles POST /projects/:projectId/pages/:pageId/archive
func (h *PageHandler) Archive(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	if svcErr := h.svc.Archive(pageID, projectID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Page archived"})
}

// Restore handles POST /projects/:projectId/pages/:pageId/restore
func (h *PageHandler) Restore(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	if svcErr := h.svc.Restore(pageID, projectID); svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Page restored"})
}

// Move handles POST /projects/:projectId/pages/:pageId/move
func (h *PageHandler) Move(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	var req request.PageMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	page, svcErr := h.svc.Move(pageID, projectID, &req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, page)
}

// ListChildren handles GET /projects/:projectId/pages/:pageId/children
func (h *PageHandler) ListChildren(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	children, svcErr := h.svc.ListChildren(pageID, projectID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, children)
}

// Lock handles POST /projects/:projectId/pages/:pageId/lock
func (h *PageHandler) Lock(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	page, svcErr := h.svc.Lock(pageID, projectID, h.getUserID(c))
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, page)
}

// Unlock handles POST /projects/:projectId/pages/:pageId/unlock
func (h *PageHandler) Unlock(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	page, svcErr := h.svc.Unlock(pageID, projectID, h.getUserID(c))
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, page)
}

// Export handles GET /projects/:projectId/pages/:pageId/export?format=md|html|txt
func (h *PageHandler) Export(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	format := c.DefaultQuery("format", "md")
	filename, content, svcErr := h.svc.GetForExport(pageID, projectID, format)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	contentTypes := map[string]string{
		"md":   "text/markdown; charset=utf-8",
		"html": "text/html; charset=utf-8",
		"txt":  "text/plain; charset=utf-8",
	}
	ct, ok := contentTypes[format]
	if !ok {
		ct = "text/plain; charset=utf-8"
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, ct, []byte(content))
}

// ConvertToIssue handles POST /projects/:projectId/pages/:pageId/convert-to-issue
func (h *PageHandler) ConvertToIssue(c *gin.Context) {
	projectID, err := h.getProjectID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
		return
	}

	var req request.PageConvertRequest
	c.ShouldBindJSON(&req)

	issue, svcErr := h.svc.ConvertToIssue(pageID, projectID, h.getUserID(c), req.IssueTypeID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*common.AppError); ok {
			c.JSON(appErr.Code, gin.H{"message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Issue created", "issue": issue})
}
