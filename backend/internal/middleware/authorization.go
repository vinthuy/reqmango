package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// RequirePermission admits users whose role grants requiredPerm in every
// project the request touches. The project is resolved from the route
// (:workflowId, :issueId, :projectId), ?project_id=, and for bulk endpoints
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

	if id := paramID(c, "workflowId"); id != 0 {
		var wf model.AgentWorkflow
		if db.Select("id", "project_id").First(&wf, id).Error == nil {
			add(wf.ProjectID)
		}
	} else {
		add(paramID(c, "projectId"))
	}
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
	var project model.Project
	if db.Select("id", "workspace_id").First(&project, projectID).Error != nil {
		return false
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
	if level == 0 {
		return false
	}

	var count int64
	db.Table("role_permissions rp").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Joins("JOIN roles r ON rp.role_id = r.id").
		Where("r.level = ? AND r.workspace_id IS NULL AND r.is_system = ? AND r.deleted_at IS NULL AND p.code = ?",
			level, true, requiredPerm).
		Count(&count)
	return count > 0
}
