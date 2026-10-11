package service

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

func TestBuildAgentSystemPrompt_ListsDeclinedSuggestionsForTheIssue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Comment{}))

	agentID := uint64(3)
	add := func(issueID uint64, suggestions []model.IssueSuggestion, outcomes map[string]model.SuggestionOutcome) {
		raw, _ := json.Marshal(suggestions)
		oc, _ := json.Marshal(outcomes)
		require.NoError(t, db.Create(&model.Comment{
			IssueID: issueID, AgentID: &agentID, Body: "建议", Suggestions: raw, SuggestionOutcomes: oc,
		}).Error)
	}
	add(42, []model.IssueSuggestion{
		{Field: "priority", Value: "urgent", Label: "urgent"},
		{Field: "title", Value: "新标题", Label: "新标题"},
		{Field: "state", Value: float64(2), Label: "待办 (Todo)"},
	}, map[string]model.SuggestionOutcome{
		"priority": {Status: model.SuggestionDismissed},
		"title":    {Status: model.SuggestionApplied},
		"state":    {Status: model.SuggestionFailed, Reason: "工作流不允许"},
	})
	add(7, []model.IssueSuggestion{{Field: "priority", Value: "low", Label: "low"}},
		map[string]model.SuggestionOutcome{"priority": {Status: model.SuggestionDismissed}})

	s := &AgentService{db: db}
	issueID := uint64(42)
	prompt := s.buildAgentSystemPrompt(&model.Agent{Name: "请求分诊"}, &DispatchContext{WorkspaceID: 1, IssueID: &issueID})

	assert.Contains(t, prompt, "priority → urgent（用户已忽略）")
	assert.Contains(t, prompt, "state → 待办 (Todo)（无法应用：工作流不允许）")
	assert.NotContains(t, prompt, "新标题", "applied suggestions are not feedback")
	assert.NotContains(t, prompt, "→ low", "other issues' declines stay out")

	other := uint64(99)
	assert.NotContains(t, s.buildAgentSystemPrompt(&model.Agent{Name: "x"}, &DispatchContext{IssueID: &other}), "未被采纳")
}

func TestBuildAgentSystemPrompt_IssueDispatchStatesTheIssuesProject(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Issue{}, &model.Comment{}))
	issue := model.Issue{Name: "支付失败", ProjectID: 3015, WorkspaceID: 3658, StateID: 1}
	require.NoError(t, db.Create(&issue).Error)

	s := &AgentService{db: db}
	prompt := s.buildAgentSystemPrompt(&model.Agent{Name: "请求分诊"},
		&DispatchContext{WorkspaceID: 3658, IssueID: &issue.ID})

	assert.Contains(t, prompt, "project_id=3015",
		"without it the model guesses workspace_id as project_id and list_states/list_labels come back empty")
}
