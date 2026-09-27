package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"gorm.io/gorm"
)

type WorkspaceAnalyticsHandler struct {
	db  *gorm.DB
	svc *service.WorkspaceAnalyticsService
}

func NewWorkspaceAnalyticsHandler(db *gorm.DB, svc *service.WorkspaceAnalyticsService) *WorkspaceAnalyticsHandler {
	return &WorkspaceAnalyticsHandler{db: db, svc: svc}
}

var allowedAnalyticsDays = map[int]bool{7: true, 14: true, 30: true, 90: true, 180: true}

type analyticsScope struct {
	workspaceID uint64
	projectID   uint64
	days        int
}

// scope resolves the workspace (id or slug), checks membership and validates
// the days / project_id query parameters. It writes the error response itself.
func (h *WorkspaceAnalyticsHandler) scope(c *gin.Context) (analyticsScope, bool) {
	var sc analyticsScope
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return sc, false
	}

	var ws model.Workspace
	param := c.Param("wsParam")
	q := h.db.Select("id")
	if id, err := strconv.ParseUint(param, 10, 64); err == nil {
		q = q.Where("id = ?", id)
	} else {
		q = q.Where("slug = ?", param)
	}
	if err := q.First(&ws).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Workspace not found"})
		return sc, false
	}
	sc.workspaceID = ws.ID

	var members int64
	h.db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND is_active = ?", ws.ID, user.ID, true).
		Count(&members)
	if members == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "Not a member of this workspace"})
		return sc, false
	}

	sc.days = 30
	if v := c.Query("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || !allowedAnalyticsDays[n] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "days must be one of 7, 14, 30, 90, 180"})
			return sc, false
		}
		sc.days = n
	}

	if v := c.Query("project_id"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project_id"})
			return sc, false
		}
		var cnt int64
		h.db.Model(&model.Project{}).Where("id = ? AND workspace_id = ?", n, ws.ID).Count(&cnt)
		if cnt == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
			return sc, false
		}
		sc.projectID = n
	}
	return sc, true
}

// Get handles GET /workspaces/:wsParam/analytics?days=30&project_id=
func (h *WorkspaceAnalyticsHandler) Get(c *gin.Context) {
	sc, ok := h.scope(c)
	if !ok {
		return
	}
	data, err := h.svc.Get(sc.workspaceID, sc.projectID, sc.days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to compute analytics"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// Loop handles GET /workspaces/:wsParam/analytics/loop?days=30&project_id=
func (h *WorkspaceAnalyticsHandler) Loop(c *gin.Context) {
	sc, ok := h.scope(c)
	if !ok {
		return
	}
	data, err := h.svc.Loop(sc.workspaceID, sc.projectID, sc.days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to compute delivery loop"})
		return
	}
	c.JSON(http.StatusOK, data)
}
