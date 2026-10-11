package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

// RequirePageAccess admits whoever may access the :pageId page's project.
func RequirePageAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
		if err != nil || pageID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid page ID"})
			return
		}
		var page model.Page
		if err := db.Select("id", "project_id").First(&page, pageID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Page not found"})
			return
		}
		requireProject(c, db, page.ProjectID)
	}
}

func requireProject(c *gin.Context, db *gorm.DB, projectID uint64) {
	if _, ok := allowProject(c, db, projectID); ok {
		c.Next()
	}
}

// allowProject aborts the request unless the caller may access the project.
func allowProject(c *gin.Context, db *gorm.DB, projectID uint64) (*model.Project, bool) {
	user, ok := accessUser(c)
	if !ok {
		return nil, false
	}
	var project model.Project
	if err := db.Select("id", "workspace_id").First(&project, projectID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Project not found"})
		return nil, false
	}
	if !user.IsSuperuser && !CanAccessProject(db, project.WorkspaceID, projectID, user.ID) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "No access to this project"})
		return nil, false
	}
	return &project, true
}

// RequireProjectWorkflow guards /projects/:projectId/workflows routes: the
// caller must reach the project, :workflowId must belong to it, and any
// :nodeId, :edgeId, :runId or :transitionId must belong to that workflow.
// Transition scope follows the service: state workflows first, then agent workflows.
func RequireProjectWorkflow(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 64)
		if err != nil || projectID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
			return
		}
		project, ok := allowProject(c, db, projectID)
		if !ok {
			return
		}
		workflowID := paramID(c, "workflowId")
		if workflowID == 0 {
			c.Next()
			return
		}

		found := false
		if strings.Contains(c.FullPath(), "/transitions") {
			var wf model.Workflow
			if err := db.Select("id", "project_id", "workspace_id").First(&wf, workflowID).Error; err == nil {
				if wf.ProjectID != nil {
					found = *wf.ProjectID == projectID
				} else {
					// Workspace-wide workflows are readable from their projects
					// but only editable from workspace settings.
					found = wf.WorkspaceID == project.WorkspaceID && c.Request.Method == http.MethodGet
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				var n int64
				db.Model(&model.AgentWorkflow{}).Where("id = ? AND project_id = ?", workflowID, projectID).Count(&n)
				found = n > 0
			}
			found = found && childOf(db, &model.StateTransition{}, paramID(c, "transitionId"), workflowID)
		} else {
			var n int64
			db.Model(&model.AgentWorkflow{}).Where("id = ? AND project_id = ?", workflowID, projectID).Count(&n)
			found = n > 0 &&
				childOf(db, &model.WorkflowNode{}, paramID(c, "nodeId"), workflowID) &&
				childOf(db, &model.WorkflowEdge{}, paramID(c, "edgeId"), workflowID) &&
				childOf(db, &model.WorkflowRun{}, paramID(c, "runId"), workflowID)
		}
		if !found {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Workflow not found"})
			return
		}
		c.Next()
	}
}

// RequireWorkspaceWorkflow guards /workspaces/:wsParam/workflows routes:
// members may read, admins may change, and :workflowId / :transitionId must
// belong to the workspace.
func RequireWorkspaceWorkflow(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsID, ok := workspaceFromParam(c, db)
		if !ok {
			return
		}
		user, ok := accessUser(c)
		if !ok {
			return
		}
		if !user.IsSuperuser {
			minRole := common.RoleGuest
			if c.Request.Method != http.MethodGet {
				minRole = common.RoleAdmin
			}
			var n int64
			db.Model(&model.WorkspaceMember{}).
				Where("workspace_id = ? AND user_id = ? AND is_active = ? AND role >= ?", wsID, user.ID, true, minRole).
				Count(&n)
			if n == 0 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Workspace admin required"})
				return
			}
		}
		if workflowID := paramID(c, "workflowId"); workflowID != 0 {
			// GET .../:workflowId is served from agent workflows; every other
			// route here addresses a state workflow.
			var table interface{} = &model.Workflow{}
			if c.Request.Method == http.MethodGet && strings.HasSuffix(c.FullPath(), "/:workflowId") {
				table = &model.AgentWorkflow{}
			}
			var n int64
			db.Model(table).Where("id = ? AND workspace_id = ?", workflowID, wsID).Count(&n)
			if n == 0 || !childOf(db, &model.StateTransition{}, paramID(c, "transitionId"), workflowID) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Workflow not found"})
				return
			}
		}
		c.Next()
	}
}

// childOf reports whether row id of the given model has workflow_id =
// workflowID; an absent id (0) imposes no constraint.
func childOf(db *gorm.DB, m interface{}, id, workflowID uint64) bool {
	if id == 0 {
		return true
	}
	var n int64
	db.Model(m).Where("id = ? AND workflow_id = ?", id, workflowID).Count(&n)
	return n > 0
}

// workspaceFromParam resolves :wsParam as a numeric ID or a slug.
func workspaceFromParam(c *gin.Context, db *gorm.DB) (uint64, bool) {
	param := c.Param("wsParam")
	if id, err := strconv.ParseUint(param, 10, 64); err == nil && id > 0 {
		return id, true
	}
	var ws model.Workspace
	if param == "" || db.Select("id").Where("slug = ?", param).First(&ws).Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Workspace not found"})
		return 0, false
	}
	return ws.ID, true
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
