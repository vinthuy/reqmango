package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/config"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Settings live in two places — the workspace state list and the project's own
// states and labels — and every mutation of them has to ask for
// settings:manage. This asserts the wiring, not just the middleware: a member
// whose role only grants settings:view may read them and must be refused when
// it writes, on both surfaces, while an admin still gets through.
func TestSettingsMutationsRequireSettingsManage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Workspace{}, &model.WorkspaceMember{},
		&model.Role{}, &model.Permission{}, &model.RolePermission{},
	))
	// Projects carry associations the settings routes never touch, so create
	// just the columns the tenancy guard reads.
	require.NoError(t, db.Exec(`CREATE TABLE projects (id integer primary key, name text, identifier text,
		workspace_id integer, deleted_at datetime)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE project_members (id integer primary key, project_id integer,
		user_id integer, role integer, is_active boolean, deleted_at datetime)`).Error)

	const member, admin, wsID, projectID = 1, 2, 10, 20
	mk := func(v interface{}) { require.NoError(t, db.Create(v).Error) }
	mk(&model.User{BaseModel: model.BaseModel{ID: member}, Email: "member@example.com", Username: "member", DisplayName: "member"})
	mk(&model.User{BaseModel: model.BaseModel{ID: admin}, Email: "admin@example.com", Username: "admin", DisplayName: "admin"})
	mk(&model.Workspace{BaseModel: model.BaseModel{ID: wsID}, Name: "acme", Slug: "acme"})
	mk(&model.WorkspaceMember{WorkspaceID: wsID, UserID: member, Role: 15, IsActive: true})
	mk(&model.WorkspaceMember{WorkspaceID: wsID, UserID: admin, Role: 20, IsActive: true})
	require.NoError(t, db.Exec(`INSERT INTO projects (id, name, identifier, workspace_id) VALUES (?, 'p', 'P', ?)`, projectID, wsID).Error)
	require.NoError(t, db.Exec(`INSERT INTO project_members (project_id, user_id, role, is_active) VALUES (?, ?, 15, 1)`, projectID, member).Error)
	require.NoError(t, db.Exec(`INSERT INTO project_members (project_id, user_id, role, is_active) VALUES (?, ?, 20, 1)`, projectID, admin).Error)

	view := model.Permission{Code: "settings:view", Name: "v", Resource: "settings", Action: "view", Scope: "project"}
	manage := model.Permission{Code: "settings:manage", Name: "m", Resource: "settings", Action: "manage", Scope: "project"}
	mk(&view)
	mk(&manage)
	memberRole := model.Role{Name: "Member", Scope: "workspace", Level: 15, IsSystem: true}
	adminRole := model.Role{Name: "Admin", Scope: "workspace", Level: 20, IsSystem: true}
	mk(&memberRole)
	mk(&adminRole)
	mk(&model.RolePermission{RoleID: memberRole.ID, PermissionID: view.ID})
	mk(&model.RolePermission{RoleID: adminRole.ID, PermissionID: view.ID})
	mk(&model.RolePermission{RoleID: adminRole.ID, PermissionID: manage.ID})

	cfg := config.Load()
	auth := service.NewAuthService(db, cfg)
	memberToken, _, err := auth.GenerateTokenForTest(member)
	require.NoError(t, err)
	adminToken, _, err := auth.GenerateTokenForTest(admin)
	require.NoError(t, err)

	r := gin.New()
	SetupRoutes(r, db, cfg)

	call := func(token, method, path string) (int, string) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(`{"name":"Blocked"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}

	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/workspaces/10/settings/states"},
		{http.MethodPut, "/api/v1/workspaces/10/settings/states/3"},
		{http.MethodDelete, "/api/v1/workspaces/10/settings/states/3"},
		{http.MethodPost, "/api/v1/projects/20/settings/states"},
		{http.MethodPost, "/api/v1/projects/20/settings/states/default"},
		{http.MethodPut, "/api/v1/projects/20/settings/states/3"},
		{http.MethodDelete, "/api/v1/projects/20/settings/states/3"},
		{http.MethodPost, "/api/v1/projects/20/settings/labels"},
		{http.MethodPut, "/api/v1/projects/20/settings/labels/3"},
		{http.MethodDelete, "/api/v1/projects/20/settings/labels/3"},
	} {
		code, body := call(memberToken, route.method, route.path)
		// The tenancy guard answers 403 too, so insist on the permission
		// middleware's payload: the refusal has to be settings:manage.
		assert.Equal(t, http.StatusForbidden, code, "%s %s", route.method, route.path)
		assert.Contains(t, body, "settings:manage", "%s %s", route.method, route.path)
	}

	for _, path := range []string{"/api/v1/workspaces/10/settings/states", "/api/v1/projects/20/settings/states"} {
		code, body := call(memberToken, http.MethodGet, path)
		assert.NotEqual(t, http.StatusForbidden, code, "member must still read %s: %s", path, body)
	}

	// An admin is admitted by the gate; what the handler then answers depends
	// on fixtures this test does not set up.
	code, _ := call(adminToken, http.MethodPost, "/api/v1/workspaces/10/settings/states")
	assert.NotEqual(t, http.StatusForbidden, code, "admin must not be refused")
}
