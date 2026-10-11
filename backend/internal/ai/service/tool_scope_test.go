package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestExecuteTool_KeepsModelSuppliedIDsInsideTheCallersScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.ProjectMember{}, &model.WorkspaceMember{}, &model.Issue{},
		&model.State{}, &model.IssueType{}, &model.Label{}, &model.Cycle{}, &model.IssueCycle{}))

	const user = uint64(7)
	mk := func(ws uint64, member bool) uint64 {
		p := model.Project{Name: "p", Identifier: "P", WorkspaceID: ws}
		require.NoError(t, db.Create(&p).Error)
		if member {
			require.NoError(t, db.Create(&model.ProjectMember{ProjectID: p.ID, UserID: user, IsActive: true}).Error)
		}
		require.NoError(t, db.Create(&model.Label{Name: "l", ProjectID: p.ID, WorkspaceID: ws}).Error)
		return p.ID
	}
	own := mk(10, true)
	sibling := mk(10, false) // same workspace, caller not a member
	foreign := mk(20, true)  // caller is a member, but of another workspace

	issueIn := func(pid, ws uint64) uint64 {
		iss := model.Issue{Name: "i", ProjectID: pid, WorkspaceID: ws, StateID: 1}
		require.NoError(t, db.Create(&iss).Error)
		return iss.ID
	}
	ownIssue, foreignIssue := issueIn(own, 10), issueIn(foreign, 20)
	cycle := model.Cycle{Name: "c", ProjectID: foreign, WorkspaceID: 20}
	require.NoError(t, db.Create(&cycle).Error)

	s := &AIService{db: db}
	actx := &AIContext{WorkspaceID: 10, ProjectID: own, UserID: user, Mode: "agent"}
	run := func(tool string, args map[string]interface{}) error {
		raw, _ := json.Marshal(args)
		_, err := s.ExecuteTool(tool, raw, actx)
		return err
	}

	assert.NoError(t, run("list_labels", nil), "context defaults are trusted")
	assert.NoError(t, run("list_labels", map[string]interface{}{"project_id": own}))
	assert.NoError(t, run("get_issue", map[string]interface{}{"issue_id": ownIssue}))

	denied := func(tool string, args map[string]interface{}) {
		t.Helper()
		err := run(tool, args)
		if assert.Error(t, err, tool) {
			assert.Contains(t, err.Error(), "超出当前范围", tool)
		}
	}
	err = run("list_labels", map[string]interface{}{"project_id": 10})
	if assert.Error(t, err, "a workspace id guessed as project_id") {
		assert.Contains(t, err.Error(), fmt.Sprintf("project_id=%d", own), "the error names the right id so the model can retry")
	}
	denied("list_members", map[string]interface{}{"project_id": foreign})
	denied("list_states", map[string]interface{}{"project_id": sibling})
	denied("get_issue", map[string]interface{}{"issue_id": foreignIssue})
	denied("get_issue_activities", map[string]interface{}{"issue_id": foreignIssue})
	denied("get_cycle_progress", map[string]interface{}{"cycle_id": cycle.ID})
	denied("list_issue_types", map[string]interface{}{"workspace_id": 20})

	// Workspace roles reach every project in the workspace over HTTP too.
	require.NoError(t, db.Create(&model.WorkspaceMember{WorkspaceID: 10, UserID: user, IsActive: true}).Error)
	assert.NoError(t, run("list_labels", map[string]interface{}{"project_id": sibling}))
	denied("list_members", map[string]interface{}{"project_id": foreign})
}
