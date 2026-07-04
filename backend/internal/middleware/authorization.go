package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequirePermission checks if the current user has the required permission.
// It looks up the permission set from the database using user ID + workspace/project context.
func RequirePermission(db *gorm.DB, requiredPerm string, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}
		user := currentUser.(interface{ GetID() uint64 })
		userID := user.GetID()

		workspaceID := parseContextID(c, "workspaceID")
		projectID := parseContextID(c, "projectID")

		// Superusers bypass all permission checks
		// Cast to model.User to check IsSuperuser
		type superuserChecker interface {
			IsSuper() bool
		}
		if su, ok := currentUser.(superuserChecker); ok && su.IsSuper() {
			c.Next()
			return
		}

		// If no workspace context, just check if user is authenticated
		if workspaceID == 0 {
			c.Next()
			return
		}

		if !hasPermissionDB(db, userID, workspaceID, projectID, requiredPerm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":     "Permission denied",
				"required":  requiredPerm,
			})
			return
		}
		c.Next()
	}
}

// RequireRoleLevel checks if the current user has at least the given role level.
func RequireRoleLevel(db *gorm.DB, minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, exists := c.Get("currentUser")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		type superuserChecker interface {
			IsSuper() bool
		}
		if su, ok := currentUser.(superuserChecker); ok && su.IsSuper() {
			c.Next()
			return
		}

		user := currentUser.(interface{ GetID() uint64 })
		workspaceID := parseContextID(c, "workspaceID")
		projectID := parseContextID(c, "projectID")

		level := getMaxRoleLevelDB(db, user.GetID(), workspaceID, projectID)
		if level < minLevel {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		c.Next()
	}
}

func parseContextID(c *gin.Context, key string) uint64 {
	val := c.GetString(key)
	if val == "" {
		// Try query param
		val = c.Query(key)
	}
	if val == "" {
		// Try param
		val = c.Param(key)
	}
	if val == "" {
		return 0
	}
	n, _ := strconv.ParseUint(val, 10, 64)
	return n
}

func hasPermissionDB(db *gorm.DB, userID, workspaceID, projectID uint64, requiredPerm string) bool {
	// Get user's role level from workspace member
	var member struct {
		Role     int
		CustomRoleID *uint64
	}
	if err := db.Raw("SELECT role, custom_role_id FROM workspace_members WHERE user_id = ? AND workspace_id = ? LIMIT 1",
		userID, workspaceID).Scan(&member).Error; err != nil || member.Role == 0 {
		return false
	}

	// Check custom role first (if assigned)
	if member.CustomRoleID != nil {
		var count int64
		db.Raw(`
			SELECT COUNT(*) FROM role_permissions rp
			JOIN permissions p ON rp.permission_id = p.id
			WHERE rp.role_id = ? AND p.code = ?
		`, *member.CustomRoleID, requiredPerm).Count(&count)
		if count > 0 {
			return true
		}
	}

	// Check project-level custom role
	if projectID > 0 {
		var prjMember struct {
			CustomRoleID *uint64
		}
		if err := db.Raw("SELECT custom_role_id FROM project_members WHERE user_id = ? AND project_id = ? LIMIT 1",
			userID, projectID).Scan(&prjMember).Error; err == nil && prjMember.CustomRoleID != nil {
			var count int64
			db.Raw(`
				SELECT COUNT(*) FROM role_permissions rp
				JOIN permissions p ON rp.permission_id = p.id
				WHERE rp.role_id = ? AND p.code = ?
			`, *prjMember.CustomRoleID, requiredPerm).Count(&count)
			if count > 0 {
				return true
			}
		}
	}

	// Fallback to system role level matching
	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		JOIN roles r ON rp.role_id = r.id
		WHERE r.level = ? AND r.workspace_id IS NULL AND p.code = ?
	`, member.Role, requiredPerm).Count(&count)
	return count > 0
}

func getMaxRoleLevelDB(db *gorm.DB, userID, workspaceID, projectID uint64) int {
	var level int
	db.Raw("SELECT role FROM workspace_members WHERE user_id = ? AND workspace_id = ? LIMIT 1",
		userID, workspaceID).Scan(&level)
	if projectID > 0 {
		var prjLevel int
		db.Raw("SELECT role FROM project_members WHERE user_id = ? AND project_id = ? LIMIT 1",
			userID, projectID).Scan(&prjLevel)
		if prjLevel > level {
			level = prjLevel
		}
	}
	return level
}
