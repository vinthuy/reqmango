package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmanpy/backend-go/internal/middleware"
	"github.com/reqmanpy/backend-go/internal/model"
	"github.com/reqmanpy/backend-go/internal/service"
)

type ProjectPageTabHandler struct {
	svc *service.ProjectPageTabService
}

func NewProjectPageTabHandler(svc *service.ProjectPageTabService) *ProjectPageTabHandler {
	return &ProjectPageTabHandler{svc: svc}
}

// List returns all page tabs for a project.
// GET /api/v1/projects/:projectId/page-tabs
func (h *ProjectPageTabHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}
	userID := middleware.GetUserID(c)
	tabs, err := h.svc.List(projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tabs == nil {
		tabs = []model.ProjectPageTab{}
	}
	c.JSON(http.StatusOK, gin.H{"data": tabs})
}

type createTabRequest struct {
	Name       string `json:"name" binding:"required"`
	Icon       string `json:"icon"`
	TabType    string `json:"tab_type"`
	RouteKey   string `json:"route_key"`
	TargetType string `json:"target_type"`
	TargetID   *uint64 `json:"target_id"`
	TargetURL  string `json:"target_url"`
}

// Create adds a new page tab.
// POST /api/v1/projects/:projectId/page-tabs
func (h *ProjectPageTabHandler) Create(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}
	var req createTabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	tab := &model.ProjectPageTab{
		ProjectID: projectID,
		OwnerID:   userID,
		Name:      req.Name,
		Icon:      req.Icon,
		TabType:   req.TabType,
		RouteKey:  req.RouteKey,
		TargetType: req.TargetType,
		TargetID:  req.TargetID,
		TargetURL: req.TargetURL,
		Visible:   true,
	}
	if tab.TabType == "" {
		tab.TabType = "custom"
	}
	if err := h.svc.Create(tab); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": tab})
}

// Update modifies an existing page tab.
// PUT /api/v1/projects/:projectId/page-tabs/:tabId
func (h *ProjectPageTabHandler) Update(c *gin.Context) {
	tabID, err := strconv.ParseUint(c.Param("tabId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tabId"})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.Update(tabID, userID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Delete removes a page tab.
// DELETE /api/v1/projects/:projectId/page-tabs/:tabId
func (h *ProjectPageTabHandler) Delete(c *gin.Context) {
	tabID, err := strconv.ParseUint(c.Param("tabId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tabId"})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.Delete(tabID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type batchSaveRequest struct {
	Tabs []model.ProjectPageTab `json:"tabs" binding:"required"`
}

// BatchSave replaces all tabs atomically.
// PUT /api/v1/projects/:projectId/page-tabs/batch
func (h *ProjectPageTabHandler) BatchSave(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}
	var req batchSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.BatchSave(projectID, userID, req.Tabs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type reorderRequest struct {
	IDs []uint64 `json:"ids" binding:"required"`
}

// Reorder updates sort_order for all tabs.
// PUT /api/v1/projects/:projectId/page-tabs/reorder
func (h *ProjectPageTabHandler) Reorder(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
		return
	}
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.svc.Reorder(projectID, userID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
