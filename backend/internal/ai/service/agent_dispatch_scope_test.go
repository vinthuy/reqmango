package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCheckDispatchTenant_KeepsAgentIssueAndProjectInOneWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.Issue{}))

	ownP := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	otherP := model.Project{Name: "other", Identifier: "OTH", WorkspaceID: 10}
	foreignP := model.Project{Name: "foreign", Identifier: "FOR", WorkspaceID: 20}
	for _, p := range []*model.Project{&ownP, &otherP, &foreignP} {
		require.NoError(t, db.Create(p).Error)
	}
	ownIssue := model.Issue{Name: "i", ProjectID: ownP.ID, WorkspaceID: 10, StateID: 1}
	foreignIssue := model.Issue{Name: "i", ProjectID: foreignP.ID, WorkspaceID: 20, StateID: 1}
	require.NoError(t, db.Create(&ownIssue).Error)
	require.NoError(t, db.Create(&foreignIssue).Error)

	s := &AgentService{db: db}
	agent := &model.Agent{WorkspaceID: 10}
	ptr := func(v uint64) *uint64 { return &v }

	assert.NoError(t, s.checkDispatchTenant(agent, &DispatchContext{WorkspaceID: 10}))
	assert.NoError(t, s.checkDispatchTenant(agent, &DispatchContext{IssueID: ptr(ownIssue.ID), ProjectID: ptr(ownP.ID)}))

	assert.Error(t, s.checkDispatchTenant(agent, &DispatchContext{WorkspaceID: 20}), "agent from another workspace")
	assert.Error(t, s.checkDispatchTenant(agent, &DispatchContext{WorkspaceID: 10, IssueID: ptr(foreignIssue.ID)}))
	assert.Error(t, s.checkDispatchTenant(agent, &DispatchContext{WorkspaceID: 10, ProjectID: ptr(foreignP.ID)}))
	assert.Error(t, s.checkDispatchTenant(agent, &DispatchContext{IssueID: ptr(ownIssue.ID), ProjectID: ptr(otherP.ID)}),
		"issue and project must agree, or tools run against a project the issue is not in")
	assert.Error(t, s.checkDispatchTenant(agent, &DispatchContext{IssueID: ptr(999)}))
}

func TestActivityAccess_StaysInTheWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Agent{}, &model.AgentActivity{}))
	for _, a := range []model.Agent{
		{BaseModel: model.BaseModel{ID: 1}, WorkspaceID: 10, Name: "own"},
		{BaseModel: model.BaseModel{ID: 2}, WorkspaceID: 20, Name: "foreign"},
	} {
		a.Capabilities, a.InvocationTargets = []byte(`[]`), []byte(`[]`)
		require.NoError(t, db.Create(&a).Error)
	}
	own := model.AgentActivity{AgentID: 1, Action: "dispatch", ResultSummary: "own"}
	foreign := model.AgentActivity{AgentID: 2, Action: "dispatch", ResultSummary: "foreign"}
	require.NoError(t, db.Create(&own).Error)
	require.NoError(t, db.Create(&foreign).Error)

	s := &AgentService{db: db}
	foreignAgent := uint64(2)
	list, err := s.ListWorkspaceActivity(10, &foreignAgent, nil, "", 50)
	require.NoError(t, err)
	assert.Empty(t, list, "agent_id filter cannot reach another workspace's agent")

	assert.NoError(t, s.UpdateActivityFeedback(10, own.ID, 1))
	assert.Error(t, s.UpdateActivityFeedback(10, foreign.ID, 1))

	_, err = s.AgentInWorkspace(10, 2)
	assert.Error(t, err)
	got, err := s.AgentInWorkspace(10, 1)
	require.NoError(t, err)
	assert.Equal(t, "own", got.Name)
}
