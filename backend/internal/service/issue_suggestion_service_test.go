package service

import (
	"testing"

	"github.com/reqmango/backend/internal/dto/request"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplySuggestionToRequest_ScalarFields(t *testing.T) {
	req := &request.IssueUpdateRequest{}
	cases := []struct {
		sg    model.IssueSuggestion
		check func(t *testing.T)
	}{
		{model.IssueSuggestion{Field: "title", Value: "登录页返回 500"},
			func(t *testing.T) { require.NotNil(t, req.Name); assert.Equal(t, "登录页返回 500", *req.Name) }},
		{model.IssueSuggestion{Field: "priority", Value: "HIGH"},
			func(t *testing.T) { require.NotNil(t, req.Priority); assert.Equal(t, "high", *req.Priority) }},
		{model.IssueSuggestion{Field: "type", Value: float64(3)},
			func(t *testing.T) { require.NotNil(t, req.TypeID); assert.Equal(t, uint64(3), *req.TypeID) }},
		{model.IssueSuggestion{Field: "state", Value: float64(7)},
			func(t *testing.T) { require.NotNil(t, req.StateID); assert.Equal(t, uint64(7), *req.StateID) }},
		{model.IssueSuggestion{Field: "assignee", Value: float64(14)},
			func(t *testing.T) { assert.Equal(t, []uint64{14}, req.AssigneeIDs) }},
		{model.IssueSuggestion{Field: "description", Value: "补齐复现步骤"},
			func(t *testing.T) { require.NotNil(t, req.DescriptionHTML) }},
	}
	for _, tc := range cases {
		assert.True(t, applySuggestionToRequest(req, tc.sg), tc.sg.Field)
		tc.check(t)
	}
}

func TestApplySuggestionToRequest_RejectsUnusableValues(t *testing.T) {
	// Each entry must be reported as not applied so the caller drops it from
	// the "applied" list instead of claiming a change that never happened.
	cases := []model.IssueSuggestion{
		{Field: "unknown", Value: "x"},
		{Field: "title", Value: "   "},
		{Field: "title", Value: 42}, // wrong type
		{Field: "priority", Value: ""},
		{Field: "type", Value: float64(0)},
		{Field: "type", Value: "abc"},
		{Field: "state", Value: nil},
		{Field: "assignee", Value: float64(0)},
	}
	for _, sg := range cases {
		req := &request.IssueUpdateRequest{}
		assert.False(t, applySuggestionToRequest(req, sg), "%s=%v must not apply", sg.Field, sg.Value)
		assert.Nil(t, req.Name)
		assert.Nil(t, req.StateID)
		assert.Empty(t, req.AssigneeIDs)
	}
}

func TestSuggestionUint_AcceptedShapes(t *testing.T) {
	// The stored value round-trips through JSONB, so float64 is the common case;
	// strings show up when a model stringifies an id.
	assert.Equal(t, uint64(3), suggestionUint(float64(3)))
	assert.Equal(t, uint64(3), suggestionUint("3"))
	assert.Equal(t, uint64(14), suggestionUint("14"))
	assert.Equal(t, uint64(9), suggestionUint(9))
	assert.Equal(t, uint64(9), suggestionUint(uint64(9)))
}

func TestSuggestionUint_RejectsGarbage(t *testing.T) {
	assert.Equal(t, uint64(0), suggestionUint("abc"))
	assert.Equal(t, uint64(0), suggestionUint("3a"))
	assert.Equal(t, uint64(0), suggestionUint(""))
	assert.Equal(t, uint64(0), suggestionUint(float64(0)))
	assert.Equal(t, uint64(0), suggestionUint(nil))
	assert.Equal(t, uint64(0), suggestionUint([]string{"3"}))
}

func TestSelectSuggestions_BulkSkipsNoOps(t *testing.T) {
	// A no-op only exists to show the agent's reasoning; applying everything
	// must not write back the value it explicitly recommended keeping.
	all := []model.IssueSuggestion{
		{Field: "title", Value: "登录页返回 500", Label: "登录页返回 500"},
		{Field: "priority", Value: "high", Label: "high", Current: "high", NoOp: true},
	}

	got := selectSuggestions(all, nil)

	assert.Len(t, got, 1)
	assert.Equal(t, "title", got[0].Field)
}

func TestSelectSuggestions_ExplicitSelectionIsHonoured(t *testing.T) {
	all := []model.IssueSuggestion{
		{Field: "title", Value: "登录页返回 500", Label: "登录页返回 500"},
		{Field: "priority", Value: "high", Label: "high", Current: "high", NoOp: true},
	}

	got := selectSuggestions(all, map[string]bool{"priority": true})

	assert.Len(t, got, 1)
	assert.Equal(t, "priority", got[0].Field)

	// An unknown field name selects nothing rather than silently applying all.
	assert.Empty(t, selectSuggestions(all, map[string]bool{"nope": true}))
}
