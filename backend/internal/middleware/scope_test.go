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

func TestScopeChecker(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.WorkspaceMember{}, &model.Project{}, &model.ProjectMember{},
		&model.Issue{}, &model.Comment{}, &model.Cycle{}))
	mk := func(v interface{}) { require.NoError(t, db.Create(v).Error) }

	mk(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "acme", Slug: "acme"})
	mk(&model.Workspace{BaseModel: model.BaseModel{ID: 20}, Name: "other", Slug: "other"})
	const alice, bob, carol, dave, root = 1, 2, 3, 4, 5 // member, ws admin, outsider, project-only guest, superuser
	mk(&model.WorkspaceMember{WorkspaceID: 10, UserID: alice, Role: 15, IsActive: true})
	mk(&model.WorkspaceMember{WorkspaceID: 10, UserID: bob, Role: 20, IsActive: true})
	mk(&model.WorkspaceMember{WorkspaceID: 20, UserID: carol, Role: 20, IsActive: true})
	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	own2 := model.Project{Name: "own2", Identifier: "OW2", WorkspaceID: 10}
	foreign := model.Project{Name: "foreign", Identifier: "FOR", WorkspaceID: 20}
	mk(&own)
	mk(&own2)
	mk(&foreign)
	mk(&model.ProjectMember{ProjectID: own.ID, UserID: dave, Role: 5, IsActive: true})

	issue := model.Issue{Name: "i", ProjectID: own.ID, WorkspaceID: 10}
	issue2 := model.Issue{Name: "j", ProjectID: own2.ID, WorkspaceID: 10}
	foreignIssue := model.Issue{Name: "k", ProjectID: foreign.ID, WorkspaceID: 20}
	mk(&issue)
	mk(&issue2)
	mk(&foreignIssue)
	comment := model.Comment{IssueID: issue.ID, Body: "c"}
	foreignComment := model.Comment{IssueID: foreignIssue.ID, Body: "c"}
	mk(&comment)
	mk(&foreignComment)
	cycle := model.Cycle{Name: "c", ProjectID: own.ID, WorkspaceID: 10}
	foreignCycle := model.Cycle{Name: "c", ProjectID: foreign.ID, WorkspaceID: 20}
	mk(&cycle)
	mk(&foreignCycle)

	sc := newScopeChecker(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ids := map[string]uint64{"alice": alice, "bob": bob, "carol": carol, "dave": dave, "root": root}
		user := &model.User{BaseModel: model.BaseModel{ID: ids[c.GetHeader("X-User")]}, IsSuperuser: c.GetHeader("X-User") == "root"}
		if sc.check(c, user) {
			c.Status(http.StatusOK)
		}
	})
	ok := func(c *gin.Context) {}
	r.GET("/api/v1/workspaces/:wsParam", ok)
	r.DELETE("/api/v1/workspaces/:wsParam", ok)
	r.DELETE("/api/v1/projects/:projectId", ok)
	r.GET("/api/v1/projects/:projectId/cycles/:cycleId", ok)
	r.GET("/api/v1/issues/:issueId/comments/:commentId", ok)
	r.PATCH("/api/v1/cycles/:id", ok)
	r.GET("/api/v1/issues", ok)

	status := func(user, method, path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-User", user)
		r.ServeHTTP(w, req)
		return w.Code
	}
	id := uitoa

	assert.Equal(t, 200, status("alice", "GET", "/api/v1/workspaces/acme"))
	assert.Equal(t, 403, status("carol", "GET", "/api/v1/workspaces/10"))
	assert.Equal(t, 404, status("carol", "GET", "/api/v1/workspaces/nope"))
	assert.Equal(t, 403, status("alice", "DELETE", "/api/v1/workspaces/10"), "deleting a workspace needs its admin")
	assert.Equal(t, 200, status("bob", "DELETE", "/api/v1/workspaces/10"))
	assert.Equal(t, 403, status("carol", "DELETE", "/api/v1/workspaces/10"))
	assert.Equal(t, 200, status("root", "DELETE", "/api/v1/workspaces/10"))

	assert.Equal(t, 403, status("carol", "DELETE", "/api/v1/projects/"+id(own.ID)))
	assert.Equal(t, 403, status("alice", "DELETE", "/api/v1/projects/"+id(own.ID)))
	assert.Equal(t, 200, status("bob", "DELETE", "/api/v1/projects/"+id(own.ID)))
	assert.Equal(t, 404, status("carol", "DELETE", "/api/v1/projects/9999"))

	ownCycle := "/api/v1/projects/" + id(own.ID) + "/cycles/"
	assert.Equal(t, 200, status("alice", "GET", ownCycle+id(cycle.ID)))
	assert.Equal(t, 200, status("dave", "GET", ownCycle+id(cycle.ID)), "project-only members reach their project")
	assert.Equal(t, 404, status("alice", "GET", ownCycle+id(foreignCycle.ID)), "a foreign row through our project")
	assert.Equal(t, 404, status("alice", "GET", "/api/v1/projects/"+id(own2.ID)+"/cycles/"+id(cycle.ID)), "a row of a sibling project")
	assert.Equal(t, 403, status("dave", "GET", "/api/v1/projects/"+id(own2.ID)+"/cycles/"+id(cycle.ID)))
	assert.Equal(t, 200, status("alice", "GET", ownCycle+"9999"), "missing rows reach the handler")

	assert.Equal(t, 200, status("alice", "PATCH", "/api/v1/cycles/"+id(cycle.ID)))
	assert.Equal(t, 403, status("alice", "PATCH", "/api/v1/cycles/"+id(foreignCycle.ID)))
	assert.Equal(t, 403, status("carol", "PATCH", "/api/v1/cycles/"+id(cycle.ID)))

	assert.Equal(t, 200, status("alice", "GET", "/api/v1/issues/"+id(issue.ID)+"/comments/"+id(comment.ID)))
	assert.Equal(t, 404, status("alice", "GET", "/api/v1/issues/"+id(issue2.ID)+"/comments/"+id(comment.ID)), "comment of another issue")
	assert.Equal(t, 403, status("alice", "GET", "/api/v1/issues/"+id(foreignIssue.ID)+"/comments/"+id(foreignComment.ID)))
	assert.Equal(t, 403, status("carol", "GET", "/api/v1/issues/"+id(issue.ID)+"/comments/"+id(comment.ID)))

	assert.Equal(t, 200, status("alice", "GET", "/api/v1/issues?workspace_id=10&project_id="+id(own.ID)))
	assert.Equal(t, 403, status("alice", "GET", "/api/v1/issues?workspace_id=20"))
	assert.Equal(t, 403, status("alice", "GET", "/api/v1/issues?project_id="+id(foreign.ID)))
	assert.Equal(t, 403, status("alice", "GET", "/api/v1/issues?issue_id="+id(foreignIssue.ID)))
	assert.Equal(t, 200, status("dave", "GET", "/api/v1/issues?project_id="+id(own.ID)))
}
