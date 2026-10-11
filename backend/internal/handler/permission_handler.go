package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/middleware"
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
