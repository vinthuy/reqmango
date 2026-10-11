package handler

import (
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

func TestBuildContext_IgnoresIDsOutsideTheAuthorisedScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.Issue{}, &model.WorkspaceMember{}, &model.Page{}))

	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	other := model.Project{Name: "other", Identifier: "OTH", WorkspaceID: 20}
	require.NoError(t, db.Create(&own).Error)
	require.NoError(t, db.Create(&other).Error)
	ownIssue := model.Issue{Name: "own", ProjectID: own.ID, WorkspaceID: 10, StateID: 1}
	otherIssue := model.Issue{Name: "other", ProjectID: other.ID, WorkspaceID: 20, StateID: 1}
	require.NoError(t, db.Create(&ownIssue).Error)
	require.NoError(t, db.Create(&otherIssue).Error)
	require.NoError(t, db.Create(&model.WorkspaceMember{WorkspaceID: 10, UserID: 1, IsActive: true}).Error)

	h := &AIHandler{db: db}
	build := func(projectID uint64, query string) (ws, issue uint64) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/x?"+query, nil)
		c.Set("currentUser", &model.User{BaseModel: model.BaseModel{ID: 1}})
		if projectID != 0 {
			c.Params = gin.Params{{Key: "projectId", Value: strconv.FormatUint(projectID, 10)}}
		}
		actx := h.buildContext(c)
		return actx.WorkspaceID, actx.IssueID
	}
	u := func(n uint64) string { return strconv.FormatUint(n, 10) }

	ws, _ := build(own.ID, "workspace_id=20")
	assert.EqualValues(t, 10, ws, "a project route cannot switch to another workspace's AI key")

	ws, _ = build(0, "workspace_id=20")
	assert.EqualValues(t, 0, ws, "not a member of workspace 20")
	ws, _ = build(0, "workspace_id=10")
	assert.EqualValues(t, 10, ws)

	_, issue := build(own.ID, "issue_id="+u(ownIssue.ID))
	assert.Equal(t, ownIssue.ID, issue)
	_, issue = build(own.ID, "issue_id="+u(otherIssue.ID))
	assert.Zero(t, issue, "an issue outside the authorised project is dropped")
}
