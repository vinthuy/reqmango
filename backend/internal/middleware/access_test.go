package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAccessMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.WorkspaceMember{}, &model.Project{}, &model.ProjectMember{}, &model.Issue{}))

	// Workspace 10: alice is a member. Project 1 is in it. Workspace 20 is foreign.
	require.NoError(t, db.Create(&model.WorkspaceMember{WorkspaceID: 10, UserID: 1, Role: 15, IsActive: true}).Error)
	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	foreign := model.Project{Name: "foreign", Identifier: "FOR", WorkspaceID: 20}
	require.NoError(t, db.Create(&own).Error)
	require.NoError(t, db.Create(&foreign).Error)
	// Bob is only a member of the foreign project, not of its workspace.
	require.NoError(t, db.Create(&model.ProjectMember{ProjectID: foreign.ID, UserID: 2, IsActive: true}).Error)
	ownIssue := model.Issue{Name: "i", ProjectID: own.ID, WorkspaceID: 10, StateID: 1}
	foreignIssue := model.Issue{Name: "i", ProjectID: foreign.ID, WorkspaceID: 20, StateID: 1}
	require.NoError(t, db.Create(&ownIssue).Error)
	require.NoError(t, db.Create(&foreignIssue).Error)

	users := map[string]*model.User{
		"alice": {BaseModel: model.BaseModel{ID: 1}},
		"bob":   {BaseModel: model.BaseModel{ID: 2}},
		"root":  {BaseModel: model.BaseModel{ID: 3}, IsSuperuser: true},
	}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("currentUser", users[c.GetHeader("X-User")]) })
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	r.GET("/workspaces/:wsParam/x", RequireWorkspaceMember(db), ok)
	r.GET("/workspaces/:wsParam/admin", RequireWorkspaceAdmin(db), ok)
	r.GET("/projects/:projectId/x", RequireProjectAccess(db), ok)
	r.GET("/issues/:issueId/x", RequireIssueAccess(db), ok)

	status := func(user, path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-User", user)
		r.ServeHTTP(w, req)
		return w.Code
	}
	assert.Equal(t, 200, status("alice", "/workspaces/10/x"))
	assert.Equal(t, 403, status("alice", "/workspaces/20/x"))
	assert.Equal(t, 403, status("bob", "/workspaces/20/x"))
	assert.Equal(t, 200, status("root", "/workspaces/20/x"))
	assert.Equal(t, 400, status("alice", "/workspaces/abc/x"))

	assert.Equal(t, 403, status("alice", "/workspaces/10/admin"), "members are not admins")
	require.NoError(t, db.Model(&model.WorkspaceMember{}).Where("user_id = 1").Update("role", 20).Error)
	assert.Equal(t, 200, status("alice", "/workspaces/10/admin"))
	assert.Equal(t, 403, status("alice", "/workspaces/20/admin"))

	assert.Equal(t, 200, status("alice", "/projects/"+uitoa(own.ID)+"/x"), "workspace members reach its projects")
	assert.Equal(t, 403, status("alice", "/projects/"+uitoa(foreign.ID)+"/x"))
	assert.Equal(t, 200, status("bob", "/projects/"+uitoa(foreign.ID)+"/x"), "project members reach their project")
	assert.Equal(t, 404, status("alice", "/projects/999/x"))

	assert.Equal(t, 200, status("alice", "/issues/"+uitoa(ownIssue.ID)+"/x"))
	assert.Equal(t, 403, status("alice", "/issues/"+uitoa(foreignIssue.ID)+"/x"))
	assert.Equal(t, 200, status("bob", "/issues/"+uitoa(foreignIssue.ID)+"/x"))
}

func uitoa(n uint64) string { return strconv.FormatUint(n, 10) }
