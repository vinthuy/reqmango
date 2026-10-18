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

// TestTransitionKindSeparation covers the bug where `workflows` (state machine)
// and `agent_workflows` share id values: transitions of one kind must never be
// listed or mutated through the other.
func TestTransitionKindSeparation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.Project{}, &model.State{}, &model.StateTransition{},
		&model.Workflow{}, &model.AgentWorkflow{}, &model.WorkflowNode{}, &model.WorkflowEdge{}, &model.WorkflowRun{}))

	require.NoError(t, db.Create(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "ws", Slug: "ws"}).Error)
	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	other := model.Project{Name: "other", Identifier: "OTH", WorkspaceID: 10}
	require.NoError(t, db.Create(&own).Error)
	require.NoError(t, db.Create(&other).Error)

	state := model.State{Name: "Todo", ProjectID: &own.ID, WorkspaceID: 10}
	require.NoError(t, db.Create(&state).Error)

	// The same numeric id in both workflow tables: the exact collision that
	// made a project-scoped route resolve the wrong workflow.
	const sharedID uint64 = 7
	stateWF := model.Workflow{BaseModel: model.BaseModel{ID: sharedID}, Name: "state", ProjectID: &other.ID, WorkspaceID: 10}
	agentWF := model.AgentWorkflow{BaseModel: model.BaseModel{ID: sharedID}, Name: "agent", ProjectID: own.ID, WorkspaceID: 10}
	require.NoError(t, db.Create(&stateWF).Error)
	require.NoError(t, db.Create(&agentWF).Error)

	svc := NewWorkflowService(db, nil, nil, nil)

	// A transition created through the agent-workflow API belongs to that
	// workflow's project, not to the colliding state workflow's project.
	created, err := svc.AddTransition(model.TransitionKindAgent, sharedID, AddTransitionRequest{
		Name:          "agent edge",
		SourceStateID: state.ID,
		TargetStateID: state.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, model.TransitionKindAgent, created.WorkflowKind)
	require.NotNil(t, created.ProjectID)
	assert.Equal(t, own.ID, *created.ProjectID, "scope comes from agent_workflows, not the colliding workflows row")

	// ListTransitions is kind-scoped in both directions.
	agentList, err := svc.ListTransitions(model.TransitionKindAgent, sharedID)
	require.NoError(t, err)
	require.Len(t, agentList, 1)

	stateList, err := svc.ListTransitions(model.TransitionKindState, sharedID)
	require.NoError(t, err)
	assert.Empty(t, stateList, "an agent transition must not appear under the state workflow")
}
