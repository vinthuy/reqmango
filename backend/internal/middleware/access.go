package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// RequireWorkspaceMember admits active members of the :wsParam workspace.
// Routes behind it may still need to check that a :id resource belongs to
// that workspace.
func RequireWorkspaceMember(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
		if err != nil || wsID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace identifier"})
			return
		}
		user, ok := accessUser(c)
		if !ok {
			return
		}
		if !user.IsSuperuser && !IsWorkspaceMember(db, wsID, user.ID) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Not a member of this workspace"})
			return
		}
		c.Next()
	}
}

// RequireWorkspaceAdmin admits admins of the :wsParam workspace.
func RequireWorkspaceAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID, err := strconv.ParseUint(c.Param("wsParam"), 10, 64)
		if err != nil || wsID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid workspace identifier"})
			return
		}
		user, ok := accessUser(c)
		if !ok {
			return
		}
		if !user.IsSuperuser {
			var n int64
			db.Model(&model.WorkspaceMember{}).
				Where("workspace_id = ? AND user_id = ? AND is_active = ? AND role >= ?", wsID, user.ID, true, common.RoleAdmin).
				Count(&n)
			if n == 0 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Workspace admin required"})
				return
			}
		}
		c.Next()
	}
}

// RequireProjectAccess admits project members and members of the project's
// workspace, matching how role levels are resolved for project routes.
func RequireProjectAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
		if err != nil || projectID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
			return
		}
		requireProject(c, db, projectID)
	}
}

// RequireIssueAccess admits whoever may access the :issueId work item's project.
func RequireIssueAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		issueID, err := strconv.ParseUint(c.Param("issueId"), 10, 64)
		if err != nil || issueID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid issue ID"})
			return
		}
		var issue model.Issue
		if err := db.Select("id", "project_id").First(&issue, issueID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Issue not found"})
			return
		}
		requireProject(c, db, issue.ProjectID)
	}
}

func requireProject(c *gin.Context, db *gorm.DB, projectID uint64) {
	user, ok := accessUser(c)
	if !ok {
		return
	}
	var project model.Project
	if err := db.Select("id", "workspace_id").First(&project, projectID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Project not found"})
		return
	}
	if !user.IsSuperuser && !CanAccessProject(db, project.WorkspaceID, projectID, user.ID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "No access to this project"})
		return
	}
	c.Next()
}

// IsWorkspaceMember reports an active workspace membership.
func IsWorkspaceMember(db *gorm.DB, workspaceID, userID uint64) bool {
	var n int64
	db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, userID, true).Count(&n)
	return n > 0
}

// CanAccessProject: an active project membership, or membership of the
// project's workspace.
func CanAccessProject(db *gorm.DB, workspaceID, projectID, userID uint64) bool {
	var n int64
	db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).Count(&n)
	return n > 0 || IsWorkspaceMember(db, workspaceID, userID)
}

func accessUser(c *gin.Context) (*model.User, bool) {
	v, exists := c.Get("currentUser")
	user, ok := v.(*model.User)
	if !exists || !ok || user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
		return nil, false
	}
	return user, true
}
