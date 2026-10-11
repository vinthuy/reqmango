package service

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSuggestionComment(t *testing.T, suggestions []model.IssueSuggestion) (*IssueService, *model.Comment) {
	t.Helper()
	svc, db, issue := newSuggestionFixture(t)
	require.NoError(t, db.AutoMigrate(&model.Comment{}))
	raw, _ := json.Marshal(suggestions)
	agentID := uint64(9)
	c := model.Comment{IssueID: issue.ID, AgentID: &agentID, Body: "建议", Suggestions: raw}
	require.NoError(t, db.Create(&c).Error)
	return svc, &c
}

func reloadComment(t *testing.T, svc *IssueService, id uint64) model.Comment {
	t.Helper()
	var c model.Comment
	require.NoError(t, svc.db.First(&c, id).Error)
	return c
}

func TestDismissIssueSuggestions_KeepsRestActionableUntilAllDecided(t *testing.T) {
	svc, c := newSuggestionComment(t, []model.IssueSuggestion{
		{Field: "priority", Value: "high"},
		{Field: "title", Value: "更清楚的标题"},
		{Field: "state", Value: float64(1), NoOp: true},
	})

	res, err := svc.DismissIssueSuggestions(42, c.ID, 7, []string{"priority"})
	require.NoError(t, err)
	assert.Equal(t, []string{"priority"}, res.Dismissed)

	got := reloadComment(t, svc, c.ID)
	assert.Nil(t, got.SuggestionsAppliedAt, "title is still pending")
	outcomes := DecodeSuggestionOutcomes(got.SuggestionOutcomes)
	assert.Equal(t, model.SuggestionDismissed, outcomes["priority"].Status)
	assert.EqualValues(t, 7, outcomes["priority"].UserID)

	_, err = svc.DismissIssueSuggestions(42, c.ID, 7, []string{"priority"})
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code, "a decided field cannot be decided again")

	res, err = svc.DismissIssueSuggestions(42, c.ID, 7, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"title"}, res.Dismissed, "the no-op is never an actionable proposal")

	got = reloadComment(t, svc, c.ID)
	assert.NotNil(t, got.SuggestionsAppliedAt, "every actionable proposal now has an outcome")
}

func TestApplyIssueSuggestions_RejectsAlreadyDecidedFields(t *testing.T) {
	svc, c := newSuggestionComment(t, []model.IssueSuggestion{
		{Field: "priority", Value: "high"},
		{Field: "title", Value: "x"},
	})
	_, err := svc.DismissIssueSuggestions(42, c.ID, 7, []string{"priority"})
	require.NoError(t, err)

	_, err = svc.ApplyIssueSuggestions(42, c.ID, 7, []string{"priority"})
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
}
