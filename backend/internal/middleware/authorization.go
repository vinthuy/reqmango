package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// RequireWorkspacePermission admits users whose workspace role grants
// requiredPerm. The workspace is resolved from :wsParam, ?workspace_id=,
// ?project_id=, or resource IDs such as :typeId / :fieldId / :templateId.
func RequireWorkspacePermission(db *gorm.DB, requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := accessUser(c)
		if !ok {
			return
		}
		if user.IsSuperuser {
			c.Next()
			return
		}
		wsID := permissionWorkspace(c, db)
		if wsID == 0 {
			c.Next()
			return
		}
		level := workspaceRoleLevel(db, user.ID, wsID)
		if level == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Not a member of this workspace"})
			return
		}
		var count int64
		systemRolePermissions(db, level).Where("p.code = ?", requiredPerm).Count(&count)
		if count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message":  "Permission denied",
				"required": requiredPerm,
			})
			return
		}
		c.Next()
	}
}

// WorkspacePermissions lists the permission codes RequireWorkspacePermission
// would grant the user in the workspace, with their workspace role level.
func WorkspacePermissions(db *gorm.DB, user *model.User, workspaceID uint64) ([]string, int) {
	codes := []string{}
	if user.IsSuperuser {
		db.Model(&model.Permission{}).Order("code").Pluck("code", &codes)
		return codes, common.RoleAdmin
	}
	level := workspaceRoleLevel(db, user.ID, workspaceID)
	if level > 0 {
		systemRolePermissions(db, level).Order("p.code").Pluck("p.code", &codes)
	}
	return codes, level
}

func workspaceRoleLevel(db *gorm.DB, userID, workspaceID uint64) int {
	var roles []int
	db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, userID, true).
		Pluck("role", &roles)
	level := 0
	for _, r := range roles {
		if r > level {
			level = r
		}
	}
	return level
}

func permissionWorkspace(c *gin.Context, db *gorm.DB) uint64 {
	if id := queryID(c, "workspace_id"); id != 0 {
		return id
	}
	if c.Param("wsParam") != "" {
		if id, ok := workspaceFromParam(c, db); ok {
			return id
		}
	}
	if pid := queryID(c, "project_id"); pid != 0 {
		var ws uint64
		db.Raw("SELECT workspace_id FROM projects WHERE id = ?", pid).Scan(&ws)
		if ws != 0 {
			return ws
		}
	}
	lookups := []struct{ param, table string }{
		{"typeId", "issue_types"},
		{"fieldId", "custom_fields"},
		{"templateId", "project_templates"},
		{"cycleId", "cycles"},
		{"initiativeId", "initiatives"},
		{"id", ""}, // table depends on path
	}
	for _, l := range lookups {
		id := paramID(c, l.param)
		if id == 0 {
			continue
		}
		table := l.table
		if table == "" {
			switch {
			case strings.Contains(c.FullPath(), "/conditional-fields"):
				table = "conditional_fields"
			case strings.Contains(c.FullPath(), "/type-templates"):
				table = "issue_type_templates"
			case strings.Contains(c.FullPath(), "/templates"):
				table = "project_templates"
			default:
				continue
			}
		}
		var ws uint64
		db.Raw(fmt.Sprintf("SELECT workspace_id FROM %s WHERE id = ?", table), id).Scan(&ws)
		if ws != 0 {
			return ws
		}
	}
	return 0
}

// RequirePermission admits users whose role grants requiredPerm in every
// project the request touches. The project is resolved from the route
// (:issueId, :projectId), ?project_id=, and for bulk endpoints
// the JSON body's issue_ids / target_project_id. A request that resolves to no
// project (e.g. an unknown ID) is passed through so the handler can 404.
func RequirePermission(db *gorm.DB, requiredPerm string, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := accessUser(c)
		if !ok {
			return
		}
		if user.IsSuperuser {
			c.Next()
			return
		}
		for _, projectID := range permissionProjects(c, db) {
			if !hasProjectPermission(db, user.ID, projectID, requiredPerm) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"message":  "Permission denied",
					"required": requiredPerm,
				})
				return
			}
		}
		c.Next()
	}
}

// permissionProjects lists the distinct projects a request acts on.
func permissionProjects(c *gin.Context, db *gorm.DB) []uint64 {
	seen := map[uint64]bool{}
	var out []uint64
	add := func(id uint64) {
		if id != 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}

	add(paramID(c, "projectId"))
	if id := paramID(c, "issueId"); id != 0 {
		// Unscoped so restoring a deleted work item is still checked.
		var issue model.Issue
		if db.Unscoped().Select("id", "project_id").First(&issue, id).Error == nil {
			add(issue.ProjectID)
		}
	}

	if id, err := strconv.ParseUint(c.Query("project_id"), 10, 64); err == nil {
		add(id)
	}

	body := peekJSONBody(c)
	if len(body.IssueIDs) > 0 {
		var projectIDs []uint64
		db.Unscoped().Model(&model.Issue{}).Where("id IN ?", body.IssueIDs).
			Distinct().Pluck("project_id", &projectIDs)
		for _, id := range projectIDs {
			add(id)
		}
	}
	add(body.TargetProjectID)
	return out
}

type permissionBody struct {
	IssueIDs        []uint64 `json:"issue_ids"`
	TargetProjectID uint64   `json:"target_project_id"`
}

// peekJSONBody reads the JSON body and puts it back for the handler.
func peekJSONBody(c *gin.Context) permissionBody {
	var b permissionBody
	if c.Request.Body == nil || !strings.HasPrefix(c.ContentType(), "application/json") {
		return b
	}
	raw, err := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	if err == nil {
		_ = json.Unmarshal(raw, &b)
	}
	return b
}

func paramID(c *gin.Context, key string) uint64 {
	id, _ := strconv.ParseUint(c.Param(key), 10, 64)
	return id
}

// hasProjectPermission checks the system role matching the user's effective
// level in the project: the higher of their workspace and project roles.
func hasProjectPermission(db *gorm.DB, userID, projectID uint64, requiredPerm string) bool {
	level := projectRoleLevel(db, userID, projectID)
	if level == 0 {
		return false
	}
	var count int64
	systemRolePermissions(db, level).Where("p.code = ?", requiredPerm).Count(&count)
	return count > 0
}

// ProjectPermissions lists the permission codes RequirePermission would grant
// the user in the project, with their effective role level.
func ProjectPermissions(db *gorm.DB, user *model.User, projectID uint64) ([]string, int) {
	codes := []string{}
	if user.IsSuperuser {
		db.Model(&model.Permission{}).Order("code").Pluck("code", &codes)
		return codes, common.RoleAdmin
	}
	level := projectRoleLevel(db, user.ID, projectID)
	if level > 0 {
		systemRolePermissions(db, level).Order("p.code").Pluck("p.code", &codes)
	}
	return codes, level
}

func systemRolePermissions(db *gorm.DB, level int) *gorm.DB {
	return db.Table("role_permissions rp").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Joins("JOIN roles r ON rp.role_id = r.id").
		Where("r.level = ? AND r.workspace_id IS NULL AND r.is_system = ? AND r.deleted_at IS NULL", level, true)
}

// projectRoleLevel is the higher of the user's active workspace and project
// roles for the project, or 0 when they have neither.
func projectRoleLevel(db *gorm.DB, userID, projectID uint64) int {
	var project model.Project
	if db.Select("id", "workspace_id").First(&project, projectID).Error != nil {
		return 0
	}
	var wsRoles, prjRoles []int
	db.Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND is_active = ?", project.WorkspaceID, userID, true).
		Pluck("role", &wsRoles)
	db.Model(&model.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, userID, true).
		Pluck("role", &prjRoles)
	level := 0
	for _, r := range append(wsRoles, prjRoles...) {
		if r > level {
			level = r
		}
	}
	return level
}
