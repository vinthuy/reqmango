package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// Every authenticated route passes through checkScope before its handler.
// It enforces tenancy from the request itself, so a handler that forgets to
// check membership cannot leak or change another workspace's data:
//
//   - ?workspace_id=, ?project_id=, ?issue_id= must be reachable by the caller;
//   - :wsParam, :projectId and :issueId likewise;
//   - any other :param is looked up in its table (see resourceTable) and the
//     row's project / workspace / issue must be reachable and must match the
//     route's own :wsParam, :projectId or :issueId;
//   - a few destructive or membership-changing routes need an admin.
//
// Rows that cannot be found are passed through so handlers can 404.

// segmentTables maps the path segment before a parameter to its table.
var segmentTables = map[string]string{
	"agent-configs":       "agent_configs",
	"agent-sessions":      "agent_sessions",
	"agent-tasks":         "agent_tasks",
	"agent-templates":     "agent_templates",
	"agents":              "agents",
	"approvals":           "approvals",
	"attachments":         "attachments",
	"automations":         "automation_rules",
	"autopilot-tasks":     "autopilot_tasks",
	"builds":              "build_records",
	"charts":              "metric_charts",
	"comments":            "comments",
	"conditional-fields":  "conditional_fields",
	"custom-fields":       "custom_fields",
	"cycles":              "cycles",
	"dashboards":          "saved_dashboards",
	"estimate-categories": "estimate_categories",
	"estimate-points":     "estimate_points",
	"estimate-time":       "estimate_times",
	"field-permissions":   "field_permissions",
	"fields":              "custom_fields",
	"github":              "github_connections",
	"initiatives":         "initiatives",
	"issue-types":         "issue_types",
	"labels":              "labels",
	"loops":               "agent_loops",
	"mcp":                 "mcp_configs",
	"memories":            "memory_entries",
	"memory-sessions":     "memory_sessions",
	"modules":             "modules",
	"page-tabs":           "project_page_tabs",
	"page-templates":      "page_templates",
	"pages":               "pages",
	"pipelines":           "agent_pipelines",
	"plugins":             "plugins",
	"releases":            "releases",
	"relations":           "issue_relations",
	"reports":             "saved_reports",
	"roles":               "roles",
	"runtimes":            "runtimes",
	"search-templates":    "search_templates",
	"skills":              "skills",
	"slack":               "slack_connections",
	"squads":              "squads",
	"stages":              "sdlc_stages",
	"states":              "states",
	"templates":           "project_templates",
	"time-tracks":         "time_tracks",
	"tools":               "tools",
	"type-templates":      "issue_type_templates",
	"versions":            "page_versions",
	"views":               "saved_views",
	"webhooks":            "webhooks",
	"work-item-templates": "work_item_templates",
}

// paramTables resolves parameters whose name alone identifies the table, for
// routes where the preceding segment does not (e.g. /custom-fields/:fieldId,
// /templates/:templateId/types/:typeId).
var paramTables = map[string]string{
	"attachmentId": "attachments",
	"chartId":      "metric_charts",
	"commentId":    "comments",
	"cycleId":      "cycles",
	"fieldId":      "custom_fields",
	"initiativeId": "initiatives",
	"labelId":      "labels",
	"moduleId":     "modules",
	"pageId":       "pages",
	"relationId":   "issue_relations",
	"releaseId":    "releases",
	"ruleId":       "automation_rules",
	"stateId":      "states",
	"typeId":       "issue_types",
	"versionId":    "page_versions",
	"viewId":       "saved_views",
}

// pathTables handles segments whose table depends on the wider route.
func pathTable(fullPath, segment string) string {
	switch {
	case segment == "types" && strings.Contains(fullPath, "/relations/types"):
		return "relation_types"
	case segment == "configs" && strings.Contains(fullPath, "/cicd/"):
		return "cicd_configs"
	case segment == "jobs" && strings.Contains(fullPath, "/developer-agent/"):
		return "developer_jobs"
	case segment == "jobs" && strings.Contains(fullPath, "/tester-agent/"):
		return "tester_jobs"
	case segment == "workflows" && strings.Contains(fullPath, "/sdlc/"):
		return "sdlc_workflows"
	}
	return ""
}

// Parameters handled elsewhere or not naming a tenant-owned row.
var unscopedParams = map[string]bool{
	"wsParam": true, "projectId": true, "issueId": true, // checked directly
	"userId": true, "token": true, "contextKey": true, "memberId": true, "executionId": true,
	// Workflow routes bind these in RequireProjectWorkflow / RequireWorkspaceWorkflow;
	// agent and MCP handlers bind :id to the workspace themselves.
	"workflowId": true, "nodeId": true, "edgeId": true, "runId": true, "transitionId": true,
}

// adminRoutes need an admin of the route's workspace ("ws") or project.
var adminRoutes = map[string]string{
	"PATCH /api/v1/workspaces/:wsParam":                  "ws",
	"DELETE /api/v1/workspaces/:wsParam":                 "ws",
	"POST /api/v1/workspaces/:wsParam/members":           "ws",
	"PATCH /api/v1/workspaces/:wsParam/members/:userId":  "ws",
	"DELETE /api/v1/workspaces/:wsParam/members/:userId": "ws",
	"PATCH /api/v1/projects/:projectId":                  "project",
	"DELETE /api/v1/projects/:projectId":                 "project",
	"POST /api/v1/projects/:projectId/archive":           "project",
	"POST /api/v1/projects/:projectId/restore":           "project",
	"PATCH /api/v1/projects/:projectId/lead":             "project",
	"POST /api/v1/projects/:projectId/members":           "project",
	"PATCH /api/v1/projects/:projectId/members/:userId":  "project",
	"DELETE /api/v1/projects/:projectId/members/:userId": "project",
}

type scopeChecker struct {
	db   *gorm.DB
	mu   sync.RWMutex
	cols map[string]map[string]bool // table -> owner columns present; nil entry = no such table
}

func newScopeChecker(db *gorm.DB) *scopeChecker {
	return &scopeChecker{db: db, cols: map[string]map[string]bool{}}
}

var ownerColumns = []string{"project_id", "workspace_id", "issue_id", "page_id"}

func (s *scopeChecker) columns(table string) map[string]bool {
	s.mu.RLock()
	cols, ok := s.cols[table]
	s.mu.RUnlock()
	if ok {
		return cols
	}
	m := s.db.Migrator()
	if m.HasTable(table) {
		cols = map[string]bool{}
		for _, col := range ownerColumns {
			if m.HasColumn(table, col) {
				cols[col] = true
			}
		}
	}
	s.mu.Lock()
	s.cols[table] = cols
	s.mu.Unlock()
	return cols
}

// deny aborts with status and returns false.
func deny(c *gin.Context, status int, message string) bool {
	c.AbortWithStatusJSON(status, gin.H{"message": message})
	return false
}

func (s *scopeChecker) check(c *gin.Context, user *model.User) bool {
	if user.IsSuperuser {
		return true
	}
	db := s.db

	// Query parameters name the tenant for list and create endpoints.
	if id := queryID(c, "workspace_id"); id != 0 && !IsWorkspaceMember(db, id, user.ID) {
		return deny(c, http.StatusForbidden, "Not a member of this workspace")
	}
	if id := queryID(c, "project_id"); id != 0 && !s.projectReachable(id, user.ID) {
		return deny(c, http.StatusForbidden, "No access to this project")
	}
	if id := queryID(c, "issue_id"); id != 0 {
		if pid := issueProject(db, id); pid != 0 && !s.projectReachable(pid, user.ID) {
			return deny(c, http.StatusForbidden, "No access to this issue")
		}
	}

	// Route parents.
	var wsID, projectID, issueID uint64
	if c.Param("wsParam") != "" {
		id, ok := s.workspaceParam(c)
		if !ok {
			return false
		}
		wsID = id
		if !IsWorkspaceMember(db, wsID, user.ID) {
			return deny(c, http.StatusForbidden, "Not a member of this workspace")
		}
	}
	if projectID = paramID(c, "projectId"); projectID != 0 {
		var project model.Project
		if db.Unscoped().Select("id", "workspace_id").First(&project, projectID).Error != nil {
			return deny(c, http.StatusNotFound, "Project not found")
		}
		if !CanAccessProject(db, project.WorkspaceID, projectID, user.ID) {
			return deny(c, http.StatusForbidden, "No access to this project")
		}
	}
	if issueID = paramID(c, "issueId"); issueID != 0 {
		if pid := issueProject(db, issueID); pid != 0 && !s.projectReachable(pid, user.ID) {
			return deny(c, http.StatusForbidden, "No access to this issue")
		}
	}

	// Child resources.
	fullPath := c.FullPath()
	segs := strings.Split(fullPath, "/")
	for i, seg := range segs {
		if !strings.HasPrefix(seg, ":") || i == 0 {
			continue
		}
		table := pathTable(fullPath, segs[i-1])
		if table == "" && unscopedParams[seg[1:]] {
			continue
		}
		id := paramID(c, seg[1:])
		if id == 0 {
			continue
		}
		if table == "" {
			table = segmentTables[segs[i-1]]
		}
		if table == "" {
			table = paramTables[seg[1:]]
		}
		if table == "" {
			continue
		}
		if ok := s.checkRow(c, user.ID, table, id, wsID, projectID, issueID); !ok {
			return false
		}
	}

	switch adminRoutes[c.Request.Method+" "+fullPath] {
	case "ws":
		var n int64
		db.Model(&model.WorkspaceMember{}).
			Where("workspace_id = ? AND user_id = ? AND is_active = ? AND role >= ?", wsID, user.ID, true, common.RoleAdmin).
			Count(&n)
		if n == 0 {
			return deny(c, http.StatusForbidden, "Workspace admin required")
		}
	case "project":
		if projectRoleLevel(db, user.ID, projectID) < common.RoleAdmin {
			return deny(c, http.StatusForbidden, "Project admin required")
		}
	}
	return true
}

// checkRow verifies the caller may reach the row's owner and that the row
// belongs to the route's own workspace / project / issue.
func (s *scopeChecker) checkRow(c *gin.Context, userID uint64, table string, id, wsID, projectID, issueID uint64) bool {
	cols := s.columns(table)
	if cols == nil {
		return true
	}
	var selects []string
	for _, col := range ownerColumns {
		if cols[col] {
			selects = append(selects, fmt.Sprintf("COALESCE(%s, 0) AS %s", col, col))
		}
	}
	if len(selects) == 0 {
		return true
	}
	var row struct{ ProjectID, WorkspaceID, IssueID, PageID uint64 }
	q := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(selects, ", "), table)
	if s.db.Raw(q, id).Scan(&row).Error != nil {
		return true
	}
	if row.ProjectID == 0 && row.IssueID != 0 {
		row.ProjectID = issueProject(s.db, row.IssueID)
	}
	if row.ProjectID == 0 && row.PageID != 0 {
		s.db.Raw("SELECT project_id FROM pages WHERE id = ?", row.PageID).Scan(&row.ProjectID)
	}
	if row.ProjectID != 0 && row.WorkspaceID == 0 {
		s.db.Raw("SELECT workspace_id FROM projects WHERE id = ?", row.ProjectID).Scan(&row.WorkspaceID)
	}

	notFound := func() bool { return deny(c, http.StatusNotFound, "Resource not found") }
	if wsID != 0 && row.WorkspaceID != 0 && row.WorkspaceID != wsID {
		return notFound()
	}
	if projectID != 0 && row.ProjectID != 0 && row.ProjectID != projectID {
		return notFound()
	}
	if issueID != 0 && row.IssueID != 0 && row.IssueID != issueID {
		return notFound()
	}
	switch {
	case row.ProjectID != 0:
		if !s.projectReachable(row.ProjectID, userID) {
			return deny(c, http.StatusForbidden, "No access to this project")
		}
	case row.WorkspaceID != 0:
		if !IsWorkspaceMember(s.db, row.WorkspaceID, userID) {
			return deny(c, http.StatusForbidden, "Not a member of this workspace")
		}
	}
	return true
}

func (s *scopeChecker) projectReachable(projectID, userID uint64) bool {
	var project model.Project
	if s.db.Unscoped().Select("id", "workspace_id").First(&project, projectID).Error != nil {
		return true // unknown project: let the handler report it
	}
	return CanAccessProject(s.db, project.WorkspaceID, projectID, userID)
}

func (s *scopeChecker) workspaceParam(c *gin.Context) (uint64, bool) {
	param := c.Param("wsParam")
	if id, err := strconv.ParseUint(param, 10, 64); err == nil {
		return id, true
	}
	var ws model.Workspace
	if s.db.Select("id").Where("slug = ?", param).First(&ws).Error != nil {
		return 0, deny(c, http.StatusNotFound, "Workspace not found")
	}
	return ws.ID, true
}

func issueProject(db *gorm.DB, issueID uint64) uint64 {
	var pid uint64
	db.Raw("SELECT project_id FROM issues WHERE id = ?", issueID).Scan(&pid)
	return pid
}

func queryID(c *gin.Context, key string) uint64 {
	id, _ := strconv.ParseUint(c.Query(key), 10, 64)
	return id
}
