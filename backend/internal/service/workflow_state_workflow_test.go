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

func newStateWorkflowDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Workspace{}, &model.Project{}, &model.State{}, &model.StateTransition{},
		&model.Workflow{}, &model.AgentWorkflow{}))
	require.NoError(t, db.Create(&model.Workspace{BaseModel: model.BaseModel{ID: 10}, Name: "ws", Slug: "ws"}).Error)
	return db
}

// TestAddTransitionScopesToTheStateWorkflow covers transition ownership now
// that workflow_id has a single parent: scope comes from `workflows`, which is
// the only table that can own a transition.
func TestAddTransitionScopesToTheStateWorkflow(t *testing.T) {
	db := newStateWorkflowDB(t)
	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	require.NoError(t, db.Create(&own).Error)
	state := model.State{Name: "Todo", ProjectID: &own.ID, WorkspaceID: 10}
	require.NoError(t, db.Create(&state).Error)

	wf := model.Workflow{Name: "state", ProjectID: &own.ID, WorkspaceID: 10}
	require.NoError(t, db.Create(&wf).Error)

	svc := NewStateWorkflowService(db)
	created, err := svc.AddTransition(wf.ID, AddTransitionRequest{
		Name:          "edge",
		SourceStateID: state.ID,
		TargetStateID: state.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, created.ProjectID)
	assert.Equal(t, own.ID, *created.ProjectID)
	assert.Equal(t, uint64(10), created.WorkspaceID)

	listed, err := svc.ListTransitions(wf.ID)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)
}

// TestStateWorkflowReadPathsCoverOnlyTheirOwnTransitions guards the workspace
// detail and list responses: they must return the transitions of the requested
// workflow and nothing else.
func TestStateWorkflowReadPathsCoverOnlyTheirOwnTransitions(t *testing.T) {
	db := newStateWorkflowDB(t)
	own := model.Project{Name: "own", Identifier: "OWN", WorkspaceID: 10}
	require.NoError(t, db.Create(&own).Error)
	to := model.State{Name: "Done", ProjectID: &own.ID, WorkspaceID: 10}
	require.NoError(t, db.Create(&to).Error)

	wf := model.Workflow{Name: "state", WorkspaceID: 10}
	other := model.Workflow{Name: "other", WorkspaceID: 10}
	require.NoError(t, db.Create(&wf).Error)
	require.NoError(t, db.Create(&other).Error)

	mine := model.StateTransition{
		Name: "mine", WorkflowID: wf.ID,
		SourceStateID: to.ID, TargetStateID: to.ID, WorkspaceID: 10,
	}
	require.NoError(t, db.Create(&mine).Error)
	require.NoError(t, db.Create(&model.StateTransition{
		Name: "other's", WorkflowID: other.ID,
		SourceStateID: to.ID, TargetStateID: to.ID, WorkspaceID: 10,
	}).Error)

	svc := NewStateWorkflowService(db)

	detail, err := svc.GetStateWorkflow(10, wf.ID)
	require.NoError(t, err)
	require.Len(t, detail.Transitions, 1, "another workflow's transition must not leak in")
	assert.Equal(t, mine.ID, detail.Transitions[0].ID)

	list, err := svc.ListStateWorkflowsByWorkspace(10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	for _, item := range list {
		if item.ID == wf.ID {
			require.Len(t, item.Transitions, 1)
			assert.Equal(t, mine.ID, item.Transitions[0].ID)
		}
	}

	// Deleting one workflow leaves the other's transitions intact.
	require.NoError(t, svc.DeleteStateWorkflow(10, wf.ID))
	var remaining int64
	require.NoError(t, db.Model(&model.StateTransition{}).Where("workflow_id = ?", other.ID).Count(&remaining).Error)
	assert.Equal(t, int64(1), remaining)
	var deleted int64
	require.NoError(t, db.Unscoped().Model(&model.StateTransition{}).Where("deleted_at IS NOT NULL").Count(&deleted).Error)
	assert.Equal(t, int64(1), deleted, "only the deleted workflow's transition is soft-deleted")
}
