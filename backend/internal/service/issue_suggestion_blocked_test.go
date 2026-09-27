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

// These cover the pre-flight checks that let one stale proposal be skipped
// instead of failing the whole "apply all" batch. The workflow-forbidden
// transition path was originally found on a live project (Backlog → Todo with
// no matching edge), so the cases below pin the same conditions.
func setupSuggestionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Issue{},
		&model.State{},
		&model.Workflow{},
		&model.StateTransition{},
		&model.IssueType{},
		&model.IssueTypeImport{},
		&model.Project{},
	))
	return db
}

func ptrU64(v uint64) *uint64 { return &v }

// newSuggestionFixture builds a project whose issue is in Backlog (state 1) and
// whose only legal transition is Backlog → In Progress (state 3).
func newSuggestionFixture(t *testing.T) (*IssueService, *gorm.DB, *model.Issue) {
	t.Helper()
	db := setupSuggestionDB(t)
	require.NoError(t, db.Create(&model.Project{BaseModel: model.BaseModel{ID: 1}, Name: "P", WorkspaceID: 1}).Error)

	for id, name := range map[uint64]string{1: "待处理 (Backlog)", 3: "进行中 (In Progress)", 2: "待办 (Todo)"} {
		require.NoError(t, db.Create(&model.State{
			BaseModel: model.BaseModel{ID: id}, Name: name, WorkspaceID: 1, ProjectID: ptrU64(1), IsActive: true,
		}).Error)
	}

	wf := model.Workflow{BaseModel: model.BaseModel{ID: 5}, Name: "默认流程", WorkspaceID: 1, ProjectID: ptrU64(1), IsActive: true}
	require.NoError(t, db.Create(&wf).Error)
	require.NoError(t, db.Create(&model.StateTransition{
		BaseModel: model.BaseModel{ID: 9}, Name: "开工", WorkflowID: 5,
		SourceStateID: 1, TargetStateID: 3, RuleType: "allow", WorkspaceID: 1, ProjectID: ptrU64(1),
	}).Error)

	// No issue_type_id: keeps the workflow lookup on the branch that works on
	// SQLite (the typed lookup uses a Postgres-only jsonb operator).
	issue := model.Issue{
		BaseModel: model.BaseModel{ID: 42}, Name: "登录页 500",
		ProjectID: 1, WorkspaceID: 1, StateID: 1,
	}
	require.NoError(t, db.Create(&issue).Error)

	return NewIssueService(db, nil, nil, nil, nil), db, &issue
}

func TestSuggestionBlockedReason_AllowsLegalStateTransition(t *testing.T) {
	svc, _, issue := newSuggestionFixture(t)

	reason := svc.suggestionBlockedReason(model.IssueSuggestion{Field: "state", Value: float64(3)}, issue, 1)

	assert.Empty(t, reason, "Backlog → In Progress has an allow edge")
}

func TestSuggestionBlockedReason_BlocksTransitionWithoutEdge(t *testing.T) {
	// The live failure: the agent proposed Backlog → Todo, which the project's
	// workflow has no edge for. It must be reported, not thrown at Update.
	svc, _, issue := newSuggestionFixture(t)

	reason := svc.suggestionBlockedReason(model.IssueSuggestion{Field: "state", Value: float64(2)}, issue, 1)

	assert.Contains(t, reason, "不被当前工作流允许")
}

func TestSuggestionBlockedReason_IgnoresStateTheIssueIsAlreadyIn(t *testing.T) {
	svc, _, issue := newSuggestionFixture(t)

	assert.Empty(t, svc.suggestionBlockedReason(model.IssueSuggestion{Field: "state", Value: float64(1)}, issue, 1))
	assert.Empty(t, svc.suggestionBlockedReason(model.IssueSuggestion{Field: "state", Value: 0}, issue, 1))
}

func TestSuggestionBlockedReason_BlocksTypeOutsideProject(t *testing.T) {
	// The other live failure: the agent read the workspace-level type list and
	// proposed id 3, which this project never imported.
	svc, db, issue := newSuggestionFixture(t)
	require.NoError(t, db.Create(&model.IssueType{
		BaseModel: model.BaseModel{ID: 3}, Name: "Bug", WorkspaceID: 1, IsActive: true,
	}).Error)
	require.NoError(t, db.Create(&model.IssueType{
		BaseModel: model.BaseModel{ID: 12369}, Name: "Bug", WorkspaceID: 1, ProjectID: ptrU64(1), IsActive: true,
	}).Error)

	blocked := svc.suggestionBlockedReason(model.IssueSuggestion{Field: "type", Value: float64(3)}, issue, 1)
	assert.Contains(t, blocked, "不在本项目可用类型中")

	allowed := svc.suggestionBlockedReason(model.IssueSuggestion{Field: "type", Value: float64(12369)}, issue, 1)
	assert.Empty(t, allowed)
}

func TestSuggestionBlockedReason_LeavesOtherFieldsAlone(t *testing.T) {
	// Only type and state carry project-scoped rules; everything else is the
	// update path's business and must not be pre-empted here.
	svc, _, issue := newSuggestionFixture(t)

	for _, field := range []string{"title", "description", "priority", "assignee"} {
		assert.Empty(t, svc.suggestionBlockedReason(model.IssueSuggestion{Field: field, Value: "x"}, issue, 1), field)
	}
}
