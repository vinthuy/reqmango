package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/middleware"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// MyProjectPermissions handles GET /projects/:projectId/my-permissions: the
// permission codes the caller holds in the project, as RequirePermission
// evaluates them, so the UI can hide actions the API would refuse.
func MyProjectPermissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
			return
		}
		user := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
			return
		}
		codes, level := middleware.ProjectPermissions(db, user, projectID)
		c.JSON(http.StatusOK, gin.H{"permissions": codes, "role_level": level})
	}
}

// MyWorkspacePermissions handles GET /workspaces/:wsParam/my-permissions, the
// workspace-level counterpart used to hide workspace settings actions.
func MyWorkspacePermissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
			return
		}
		var ws model.Workspace
		param := c.Param("wsParam")
		q := db.Select("id")
		if id, err := strconv.ParseUint(param, 10, 64); err == nil {
			q = q.Where("id = ?", id)
		} else {
			q = q.Where("slug = ?", param)
		}
		if q.First(&ws).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Workspace not found"})
			return
		}
		codes, level := middleware.WorkspacePermissions(db, user, ws.ID)
		c.JSON(http.StatusOK, gin.H{"permissions": codes, "role_level": level})
	}
}
