package workflow

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func u64(v uint64) *uint64 { return &v }

func newRulesDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}, &model.State{}, &model.Workflow{}, &model.StateTransition{}))
	return db
}

// fixture builds project 1 in workspace 1 with three states and returns the
// states by name. Backlog → In Progress is the only configured edge unless the
// test adds more.
type fixture struct {
	db     *gorm.DB
	states map[string]model.State
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := newRulesDB(t)
	require.NoError(t, db.Create(&model.Project{
		BaseModel: model.BaseModel{ID: 1}, Name: "P", WorkspaceID: 1,
	}).Error)

	f := &fixture{db: db, states: map[string]model.State{}}
	for id, name := range map[uint64]string{1: "Backlog", 2: "Todo", 3: "In Progress", 4: "Done"} {
		st := model.State{
			BaseModel: model.BaseModel{ID: id}, Name: name,
			WorkspaceID: 1, ProjectID: u64(1), IsActive: true, Sequence: int(id),
		}
		require.NoError(t, db.Create(&st).Error)
		f.states[name] = st
	}
	return f
}

func (f *fixture) workflow(t *testing.T, wf model.Workflow) model.Workflow {
	t.Helper()
	if wf.Name == "" {
		wf.Name = "wf"
	}
	wf.WorkspaceID = 1
	wf.IsActive = true
	if wf.ProjectID == nil {
		wf.ProjectID = u64(1)
	}
	require.NoError(t, f.db.Create(&wf).Error)
	return wf
}

func (f *fixture) edge(t *testing.T, workflowID uint64, from, to string, ruleType string) {
	t.Helper()
	require.NoError(t, f.db.Create(&model.StateTransition{
		Name:          from + "→" + to,
		WorkflowID:    workflowID,
		SourceStateID: f.states[from].ID,
		TargetStateID: f.states[to].ID,
		RuleType:      ruleType,
		WorkspaceID:   1,
		ProjectID:     u64(1),
		BaseModel:     model.BaseModel{},
	}).Error)
}

func (f *fixture) load(t *testing.T, issueTypeID *uint64) *Rules {
	t.Helper()
	rules, err := Load(f.db, 1, issueTypeID)
	require.NoError(t, err)
	return rules
}

func TestLoad_WithoutWorkflowsEveryMoveIsAllowed(t *testing.T) {
	f := newFixture(t)

	rules := f.load(t, nil)

	assert.False(t, rules.Restricts)
	verdict := rules.Check(f.states["Backlog"], f.states["Todo"])
	assert.Equal(t, Allowed, verdict.Outcome)
}

func TestLoad_DeniesAMoveWithNoConfiguredEdge(t *testing.T) {
	// The live failure this rule exists to prevent: an agent proposed
	// Backlog → Todo, which the project's workflow has no edge for.
	f := newFixture(t)
	wf := f.workflow(t, model.Workflow{Name: "默认流程"})
	f.edge(t, wf.ID, "Backlog", "In Progress", "allow")

	rules := f.load(t, nil)

	assert.True(t, rules.Restricts)
	assert.Equal(t, Allowed, rules.Check(f.states["Backlog"], f.states["In Progress"]).Outcome)
	assert.Equal(t, Denied, rules.Check(f.states["Backlog"], f.states["Todo"]).Outcome)
}

func TestLoad_AllowsEverythingWhenWorkflowDefinesNoEdges(t *testing.T) {
	// A workflow that exists but has no transitions yet must not lock the
	// project out of every state change.
	f := newFixture(t)
	f.workflow(t, model.Workflow{Name: "空流程"})

	rules := f.load(t, nil)

	assert.False(t, rules.Restricts)
	assert.Equal(t, Allowed, rules.Check(f.states["Backlog"], f.states["Todo"]).Outcome)
}

func TestLoad_RequiresApprovalAndReportsTheEdge(t *testing.T) {
	f := newFixture(t)
	wf := f.workflow(t, model.Workflow{Name: "带审批"})
	f.edge(t, wf.ID, "Backlog", "In Progress", "allow")
	f.edge(t, wf.ID, "Backlog", "Done", "approval")

	rules := f.load(t, nil)

	verdict := rules.Check(f.states["Backlog"], f.states["Done"])
	require.Equal(t, ApprovalRequired, verdict.Outcome)
	assert.Equal(t, wf.ID, verdict.WorkflowID)
	assert.Equal(t, "带审批", verdict.WorkflowName)
	assert.NotZero(t, verdict.TransitionID)
	assert.Equal(t, Denied, rules.Check(f.states["Todo"], f.states["Done"]).Outcome)
}

func TestLoad_ApprovalBeatsAllowForTheSamePair(t *testing.T) {
	// The update path returns "approval required" first, so the rule set has to
	// agree: the move still needs approval even if it is also allowed outright.
	f := newFixture(t)
	wf := f.workflow(t, model.Workflow{Name: "双重定义"})
	f.edge(t, wf.ID, "Backlog", "In Progress", "allow")
	f.edge(t, wf.ID, "Backlog", "In Progress", "approval")

	assert.Equal(t, ApprovalRequired, f.load(t, nil).Check(f.states["Backlog"], f.states["In Progress"]).Outcome)
}

func TestLoad_ScopesWorkflowsByIssueType(t *testing.T) {
	f := newFixture(t)
	// Only governs type 12369.
	f.workflow(t, model.Workflow{Name: "Bug 流程", IssueTypeID: u64(12369)})
	// Governs types 7 and 8 through the list column.
	f.workflow(t, model.Workflow{
		Name: "多类型流程", IssueTypeIDs: json.RawMessage(`[7,8]`),
	})

	rules := f.load(t, u64(12369))
	assert.False(t, rules.Restricts, "a workflow with no edges restricts nothing")
	assert.Len(t, rules.edges, 0)

	// A type covered by neither workflow sees no rules at all.
	other := f.load(t, u64(999))
	assert.False(t, other.Restricts)

	// An untyped issue only sees workflows that are not type-specific.
	untyped := f.load(t, nil)
	assert.False(t, untyped.Restricts)
}

func TestLoad_UntypedWorkflowGovernsEveryType(t *testing.T) {
	f := newFixture(t)
	wf := f.workflow(t, model.Workflow{Name: "项目默认"})
	f.edge(t, wf.ID, "Backlog", "In Progress", "allow")

	assert.True(t, f.load(t, u64(12369)).Restricts)
	assert.True(t, f.load(t, nil).Restricts)
}

func TestLoad_AppliesEdgesByStateNameAcrossProjects(t *testing.T) {
	// Workflows are shared per workspace, so an edge written against one
	// project's "Backlog" also governs this project's state of the same name.
	f := newFixture(t)
	other := model.State{
		BaseModel: model.BaseModel{ID: 900}, Name: "backlog",
		WorkspaceID: 1, ProjectID: u64(2), IsActive: true,
	}
	require.NoError(t, f.db.Create(&other).Error)
	otherDone := model.State{
		BaseModel: model.BaseModel{ID: 901}, Name: "done",
		WorkspaceID: 1, ProjectID: u64(2), IsActive: true,
	}
	require.NoError(t, f.db.Create(&otherDone).Error)

	wf := f.workflow(t, model.Workflow{Name: "共享流程"})
	require.NoError(t, f.db.Create(&model.StateTransition{
		Name: "shared", WorkflowID: wf.ID,
		SourceStateID: other.ID, TargetStateID: otherDone.ID,
		RuleType: "allow", WorkspaceID: 1,
	}).Error)

	rules := f.load(t, nil)

	assert.Equal(t, Allowed, rules.Check(f.states["Backlog"], f.states["Done"]).Outcome,
		"the edge is written for project 2's states but names match this project's")
}

func TestTargets_ExcludesSelfAndMarksApproval(t *testing.T) {
	f := newFixture(t)
	wf := f.workflow(t, model.Workflow{Name: "流程"})
	f.edge(t, wf.ID, "Backlog", "In Progress", "allow")
	f.edge(t, wf.ID, "Backlog", "Done", "approval")

	states := []model.State{f.states["Backlog"], f.states["Todo"], f.states["In Progress"], f.states["Done"]}
	targets := f.load(t, nil).Targets(f.states["Backlog"], states)

	names := make([]string, len(targets))
	for i, tg := range targets {
		names[i] = tg.Name
	}
	assert.Equal(t, []string{"In Progress", "Done"}, names)
	assert.True(t, targets[1].RequiresApproval, "Done is behind an approval edge")
	assert.False(t, targets[0].RequiresApproval)
}

func TestTargets_UnrestrictedProjectOffersEveryOtherState(t *testing.T) {
	f := newFixture(t)

	states := []model.State{f.states["Backlog"], f.states["Todo"], f.states["In Progress"], f.states["Done"]}
	targets := f.load(t, nil).Targets(f.states["Todo"], states)

	assert.Len(t, targets, 3, "everything except the state itself")
	for _, tg := range targets {
		assert.NotEqual(t, f.states["Todo"].ID, tg.ID)
	}
}

func TestContainsTypeID(t *testing.T) {
	assert.True(t, containsTypeID(json.RawMessage(`[7,8]`), 8))
	assert.True(t, containsTypeID(json.RawMessage(`[7]`), 7))
	assert.False(t, containsTypeID(json.RawMessage(`[7]`), 8))
	assert.False(t, containsTypeID(nil, 7))
	assert.False(t, containsTypeID(json.RawMessage(`null`), 7))
	assert.False(t, containsTypeID(json.RawMessage(`not json`), 7))
}
