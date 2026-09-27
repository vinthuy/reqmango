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

// Get handles GET /workspaces/:wsParam/analytics?days=30&project_id=
func (h *WorkspaceAnalyticsHandler) Get(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
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
		return
	}

	var members int64
	h.db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND is_active = ?", ws.ID, user.ID, true).
		Count(&members)
	if members == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "Not a member of this workspace"})
		return
	}

	days := 30
	if v := c.Query("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || !allowedAnalyticsDays[n] {
			c.JSON(http.StatusBadRequest, gin.H{"message": "days must be one of 7, 14, 30, 90, 180"})
			return
		}
		days = n
	}

	var projectID uint64
	if v := c.Query("project_id"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project_id"})
			return
		}
		var cnt int64
		h.db.Model(&model.Project{}).Where("id = ? AND workspace_id = ?", n, ws.ID).Count(&cnt)
		if cnt == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
			return
		}
		projectID = n
	}

	data, err := h.svc.Get(ws.ID, projectID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to compute analytics"})
		return
	}
	c.JSON(http.StatusOK, data)
}
