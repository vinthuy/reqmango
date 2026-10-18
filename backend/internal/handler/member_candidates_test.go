package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestListMemberCandidates checks the endpoint that replaced the never-existing
// GET /users: it must return exactly the users who can still be added, so the
// add-member form offers nothing the API would reject.
func TestListMemberCandidates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.WorkspaceMember{}, &model.Project{},
		&model.ProjectMember{}, &model.User{}))

	ws := model.Workspace{Name: "acme", Slug: "acme"}
	require.NoError(t, db.Create(&ws).Error)
	project := model.Project{Name: "core", Identifier: "CORE", WorkspaceID: ws.ID}
	require.NoError(t, db.Create(&project).Error)

	mkUser := func(email, name string) model.User {
		u := model.User{Email: email, Username: email, DisplayName: name, IsActive: true}
		require.NoError(t, db.Create(&u).Error)
		return u
	}
	alice := mkUser("alice@example.com", "Alice")
	bob := mkUser("bob@example.com", "Bob")
	carol := mkUser("carol@example.com", "Carol")
	inactive := mkUser("gone@example.com", "Gone")
	require.NoError(t, db.Model(&inactive).Update("is_active", false).Error)

	// Alice is already a member, Bob is not, Carol is a member but not on the project.
	require.NoError(t, db.Create(&model.WorkspaceMember{WorkspaceID: ws.ID, UserID: alice.ID, Role: 15, IsActive: true}).Error)
	require.NoError(t, db.Create(&model.WorkspaceMember{WorkspaceID: ws.ID, UserID: carol.ID, Role: 15, IsActive: true}).Error)
	require.NoError(t, db.Create(&model.ProjectMember{ProjectID: project.ID, UserID: alice.ID, Role: 15, IsActive: true}).Error)

	wsSvc := service.NewWorkspaceService(db)
	projSvc := service.NewProjectService(db)
	wh := NewWorkspaceHandler(wsSvc)
	ph := NewProjectHandler(projSvc, service.NewProjectTemplateService(db))

	r := gin.New()
	r.GET("/workspaces/:wsParam/member-candidates", wh.ListMemberCandidates)
	r.GET("/projects/:projectId/member-candidates", ph.ListMemberCandidates)

	get := func(path string) (int, []map[string]any) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		var out []map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	code, list := get("/workspaces/acme/member-candidates")
	require.Equal(t, http.StatusOK, code)
	ids := map[float64]bool{}
	for _, u := range list {
		ids[u["id"].(float64)] = true
	}
	assert.True(t, ids[float64(bob.ID)], "a non-member is a candidate")
	assert.False(t, ids[float64(alice.ID)], "an existing member is not")
	assert.False(t, ids[float64(inactive.ID)], "an inactive account is not")

	code, list = get("/workspaces/acme/member-candidates?q=bob")
	require.Equal(t, http.StatusOK, code)
	require.Len(t, list, 1)
	assert.Equal(t, "Bob", list[0]["display_name"])

	// Project candidates are workspace members not yet on the project.
	code, list = get("/projects/" + itoa(project.ID) + "/member-candidates")
	require.Equal(t, http.StatusOK, code)
	ids = map[float64]bool{}
	for _, u := range list {
		ids[u["id"].(float64)] = true
	}
	assert.True(t, ids[float64(carol.ID)], "a workspace member not on the project is a candidate")
	assert.False(t, ids[float64(alice.ID)], "an existing project member is not")
	assert.False(t, ids[float64(bob.ID)], "a non-member of the workspace is not")

	code, _ = get("/projects/9999/member-candidates")
	assert.Equal(t, http.StatusNotFound, code)
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
