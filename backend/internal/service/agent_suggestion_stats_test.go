package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestionAcceptance_CountsOutcomesPerAgentAndField(t *testing.T) {
	_, db, issue := newSuggestionFixture(t)
	require.NoError(t, db.AutoMigrate(&model.Comment{}, &model.Agent{}))
	require.NoError(t, db.Create(&model.Agent{
		BaseModel: model.BaseModel{ID: 9}, Name: "请求分诊", WorkspaceID: 1,
		Capabilities: json.RawMessage("[]"), InvocationTargets: json.RawMessage("[]"),
	}).Error)

	agentID := uint64(9)
	addComment := func(suggestions []model.IssueSuggestion, outcomes map[string]model.SuggestionOutcome, settled bool) {
		raw, _ := json.Marshal(suggestions)
		c := model.Comment{IssueID: issue.ID, AgentID: &agentID, Body: "建议", Suggestions: raw}
		if outcomes != nil {
			c.SuggestionOutcomes, _ = json.Marshal(outcomes)
		}
		if settled {
			now := time.Now()
			c.SuggestionsAppliedAt = &now
		}
		require.NoError(t, db.Create(&c).Error)
	}

	addComment([]model.IssueSuggestion{
		{Field: "priority", Value: "high"},
		{Field: "type", Value: float64(3)},
		{Field: "state", Value: float64(1), NoOp: true},
	}, map[string]model.SuggestionOutcome{
		"priority": {Status: model.SuggestionApplied},
		"type":     {Status: model.SuggestionDismissed},
	}, true)
	addComment([]model.IssueSuggestion{
		{Field: "priority", Value: "low"},
		{Field: "state", Value: float64(2)},
	}, map[string]model.SuggestionOutcome{
		"state": {Status: model.SuggestionFailed, Reason: "工作流不允许"},
	}, false)
	// Settled before outcomes existed: ignored rather than guessed.
	addComment([]model.IssueSuggestion{{Field: "title", Value: "x"}}, nil, true)

	svc := NewAgentPerformanceService(db)
	stats, err := svc.SuggestionAcceptance(1, "", "")
	require.NoError(t, err)
	require.Len(t, stats, 1)

	st := stats[0]
	assert.Equal(t, "请求分诊", st.AgentName)
	assert.EqualValues(t, 4, st.Proposed)
	assert.EqualValues(t, 1, st.Applied)
	assert.EqualValues(t, 1, st.Dismissed)
	assert.EqualValues(t, 1, st.Failed)
	assert.EqualValues(t, 1, st.Pending)
	assert.Equal(t, 50.0, st.AcceptanceRate)

	assert.EqualValues(t, 2, st.ByField["priority"].Proposed)
	assert.Equal(t, 100.0, st.ByField["priority"].AcceptanceRate)
	assert.Equal(t, 0.0, st.ByField["type"].AcceptanceRate)
	_, hasTitle := st.ByField["title"]
	assert.False(t, hasTitle)

	other, err := svc.SuggestionAcceptance(2, "", "")
	require.NoError(t, err)
	assert.Empty(t, other, "scoped to the workspace")
}
