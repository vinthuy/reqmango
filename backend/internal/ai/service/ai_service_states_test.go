package service

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ptrU64(v uint64) *uint64 { return &v }

type statesFixture struct {
	svc        *AIService
	states     map[string]model.State
	workflowID uint64
}

// newStatesFixture builds a tiny project: four states and one typed issue.
// Args mirror the API path, where JSON numbers arrive as float64.
func newStatesFixture(t *testing.T) statesFixture {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Project{}, &model.State{}, &model.Workflow{},
		&model.StateTransition{}, &model.Issue{},
	))
	require.NoError(t, db.Create(&model.Project{
		BaseModel: model.BaseModel{ID: 1}, Name: "P", WorkspaceID: 1,
	}).Error)

	states := map[string]model.State{}
	for id, name := range map[uint64]string{1: "Backlog", 2: "Todo", 3: "In Progress", 4: "Done"} {
		st := model.State{
			BaseModel: model.BaseModel{ID: id}, Name: name, ProjectID: ptrU64(1),
			WorkspaceID: 1, IsActive: true, Sequence: int(id),
		}
		require.NoError(t, db.Create(&st).Error)
		states[name] = st
	}

	issueTypeID := uint64(12369)
	require.NoError(t, db.Create(&model.Issue{
		BaseModel: model.BaseModel{ID: 1}, Name: "崩溃", ProjectID: 1, WorkspaceID: 1,
		StateID: states["Backlog"].ID, IssueTypeID: &issueTypeID,
	}).Error)

	return statesFixture{svc: &AIService{db: db}, states: states}
}

// addWorkflow creates the single workflow all edges in a test hang off.
func (f *statesFixture) addWorkflow(t *testing.T, wf model.Workflow) {
	t.Helper()
	wf.WorkspaceID = 1
	wf.IsActive = true
	if wf.ProjectID == nil {
		wf.ProjectID = ptrU64(1)
	}
	require.NoError(t, f.svc.db.Create(&wf).Error)
	f.workflowID = wf.ID
}

func (f *statesFixture) addEdge(t *testing.T, from, to, ruleType string) {
	t.Helper()
	require.NoError(t, f.svc.db.Create(&model.StateTransition{
		Name: from + "→" + to, WorkflowID: f.workflowID,
		SourceStateID: f.states[from].ID, TargetStateID: f.states[to].ID,
		RuleType: ruleType, WorkspaceID: 1, ProjectID: ptrU64(1),
	}).Error)
}

// allowedNextByName flattens the tool's per-state targets so a test can assert
// on the exact contract the model reads.
func allowedNextByName(t *testing.T, result any) map[string][]string {
	t.Helper()
	entries, ok := result.([]map[string]interface{})
	require.True(t, ok, "toolListStates should return a slice of entries")
	out := map[string][]string{}
	for _, e := range entries {
		targets, ok := e["allowed_next"].([]workflow.Target)
		require.True(t, ok, "allowed_next should be a []workflow.Target")
		names := make([]string, len(targets))
		for i, tg := range targets {
			names[i] = tg.Name
		}
		out[e["name"].(string)] = names
	}
	return out
}

func TestToolListStates_ExposesAllowedTargetsPerState(t *testing.T) {
	f := newStatesFixture(t)
	f.addWorkflow(t, model.Workflow{Name: "默认流程"})
	f.addEdge(t, "Backlog", "In Progress", "allow")
	f.addEdge(t, "In Progress", "Done", "allow")

	res, err := f.svc.toolListStates(map[string]interface{}{"project_id": float64(1)}, &AIContext{ProjectID: 1})
	require.NoError(t, err)

	next := allowedNextByName(t, res)
	assert.Equal(t, []string{"In Progress"}, next["Backlog"])
	assert.Equal(t, []string{"Done"}, next["In Progress"])
	assert.Empty(t, next["Todo"], "a state with no outgoing edge offers no legal move")
}

func TestToolListStates_UnrestrictedProjectOffersEveryOtherState(t *testing.T) {
	f := newStatesFixture(t)

	res, err := f.svc.toolListStates(map[string]interface{}{"project_id": float64(1)}, &AIContext{ProjectID: 1})
	require.NoError(t, err)

	next := allowedNextByName(t, res)
	assert.Len(t, next["Todo"], 3, "no workflow means every state except the current one")
	assert.NotContains(t, next["Todo"], "Todo")
}

func TestToolListStates_MarksTheIssueCurrentState(t *testing.T) {
	f := newStatesFixture(t)

	res, err := f.svc.toolListStates(
		map[string]interface{}{"project_id": float64(1), "issue_id": float64(1)},
		&AIContext{ProjectID: 1},
	)
	require.NoError(t, err)

	var current []string
	for _, e := range res.([]map[string]interface{}) {
		if e["is_current"] == true {
			current = append(current, e["name"].(string))
		}
	}
	assert.Equal(t, []string{"Backlog"}, current)
}

func suggest(t *testing.T, f statesFixture, suggestions ...map[string]interface{}) (any, []model.IssueSuggestion, error) {
	t.Helper()
	recorded := make([]model.IssueSuggestion, 0)
	items := make([]interface{}, len(suggestions))
	for i, s := range suggestions {
		items[i] = s
	}
	res, err := f.svc.toolSuggestIssueChanges(
		map[string]interface{}{"issue_id": float64(1), "suggestions": items},
		&AIContext{ProjectID: 1, IssueID: 1, Suggestions: &recorded},
	)
	return res, recorded, err
}

func TestToolSuggestIssueChanges_DropsAnIllegalStateAndSaysWhatIsLegal(t *testing.T) {
	f := newStatesFixture(t)
	f.addWorkflow(t, model.Workflow{Name: "默认流程"})
	f.addEdge(t, "Backlog", "In Progress", "allow")

	res, recorded, err := suggest(t, f,
		map[string]interface{}{"field": "priority", "value": "medium"},
		map[string]interface{}{"field": "state", "value": float64(2), "label": "Todo"},
	)
	require.NoError(t, err)

	require.Len(t, recorded, 1, "the legal proposal is kept")
	assert.Equal(t, "priority", recorded[0].Field)
	out := res.(map[string]interface{})
	assert.Equal(t, 1, out["recorded"])
	rejected := out["rejected"].([]string)
	require.Len(t, rejected, 1)
	assert.Contains(t, rejected[0], "Todo")
	assert.Contains(t, rejected[0], "In Progress", "the model is told which targets are legal")
}

func TestToolSuggestIssueChanges_ErrorsWhenOnlyAnIllegalStateWasProposed(t *testing.T) {
	f := newStatesFixture(t)
	f.addWorkflow(t, model.Workflow{Name: "默认流程"})
	f.addEdge(t, "Backlog", "In Progress", "allow")

	_, recorded, err := suggest(t, f, map[string]interface{}{"field": "state", "value": float64(2)})
	require.Error(t, err, "the model must retry instead of believing it submitted something")
	assert.Contains(t, err.Error(), "In Progress")
	assert.Empty(t, recorded)
}

func TestToolSuggestIssueChanges_KeepsLegalAndApprovalStates(t *testing.T) {
	f := newStatesFixture(t)
	f.addWorkflow(t, model.Workflow{Name: "默认流程"})
	f.addEdge(t, "Backlog", "Done", "approval")

	_, recorded, err := suggest(t, f, map[string]interface{}{"field": "state", "value": float64(4)})
	require.NoError(t, err)
	require.Len(t, recorded, 1, "an approval edge is still a legal proposal")
}

func TestToolSuggestIssueChanges_RejectsAStateFromAnotherProject(t *testing.T) {
	f := newStatesFixture(t)

	_, recorded, err := suggest(t, f, map[string]interface{}{"field": "state", "value": float64(999)})
	require.Error(t, err)
	assert.Empty(t, recorded)
}

func TestToolListStates_ScopesTargetsToTheIssuesWorkflow(t *testing.T) {
	// A workflow that only governs the issue's type restricts it, while other
	// states in the same project stay free to move anywhere.
	f := newStatesFixture(t)
	f.addWorkflow(t, model.Workflow{Name: "Bug 流程", IssueTypeID: ptrU64(12369)})
	f.addEdge(t, "Backlog", "Done", "approval")

	res, err := f.svc.toolListStates(
		map[string]interface{}{"project_id": float64(1), "issue_id": float64(1)},
		&AIContext{ProjectID: 1},
	)
	require.NoError(t, err)

	next := allowedNextByName(t, res)
	assert.Equal(t, []string{"Done"}, next["Backlog"], "only the typed workflow's edge applies")

	for _, e := range res.([]map[string]interface{}) {
		if e["name"] != "Backlog" {
			continue
		}
		targets := e["allowed_next"].([]workflow.Target)
		require.Len(t, targets, 1)
		assert.True(t, targets[0].RequiresApproval, "the approval edge is surfaced to the model")
	}
}
