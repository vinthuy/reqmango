package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRequireWorkspacePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.WorkspaceMember{}, &model.Role{}, &model.Permission{}, &model.RolePermission{}, &model.IssueType{}))

	mk := func(v interface{}) { require.NoError(t, db.Create(v).Error) }
	mk(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "acme", Slug: "acme"})
	const alice, bob = 1, 2 // member 15, admin 20
	mk(&model.WorkspaceMember{WorkspaceID: 10, UserID: alice, Role: 15, IsActive: true})
	mk(&model.WorkspaceMember{WorkspaceID: 10, UserID: bob, Role: 20, IsActive: true})
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
	it := model.IssueType{Name: "Bug", WorkspaceID: 10}
	mk(&it)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		ids := map[string]uint64{"alice": alice, "bob": bob}
		c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: ids[c.GetHeader("X-User")]}})
	})
	r.DELETE("/api/v1/issue-types/:typeId", RequireWorkspacePermission(db, "settings:manage"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.POST("/api/v1/issue-types", RequireWorkspacePermission(db, "settings:manage"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	status := func(user, method, path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-User", user)
		r.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, 403, status("alice", "DELETE", "/api/v1/issue-types/"+uitoa(it.ID)))
	assert.Equal(t, 200, status("bob", "DELETE", "/api/v1/issue-types/"+uitoa(it.ID)))
	assert.Equal(t, 403, status("alice", "POST", "/api/v1/issue-types?workspace_id=10"))
	assert.Equal(t, 200, status("bob", "POST", "/api/v1/issue-types?workspace_id=10"))
}
