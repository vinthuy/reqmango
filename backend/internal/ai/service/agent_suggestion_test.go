package service

import (
	"encoding/json"
	"testing"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseSuggestions(t *testing.T, raw string) []string {
	t.Helper()
	got, err := ParseIssueSuggestions(json.RawMessage(raw))
	require.NoError(t, err)
	out := make([]string, 0, len(got))
	for _, sg := range got {
		out = append(out, sg.Field)
	}
	return out
}

func TestParseIssueSuggestions_CanonicalFields(t *testing.T) {
	fields := parseSuggestions(t, `[
		{"field":"title","value":"????? 500","label":"????? 500"},
		{"field":"priority","value":"high"},
		{"field":"type","value":3,"label":"Bug (type_id=3)"},
		{"field":"state","value":7},
		{"field":"assignee","value":3},
		{"field":"description","value":"??????"}
	]`)
	assert.Equal(t, []string{"title", "priority", "type", "state", "assignee", "description"}, fields)
}

func TestParseIssueSuggestions_MapsAliases(t *testing.T) {
	// Agents drift between names; aliases must collapse onto canonical fields.
	fields := parseSuggestions(t, `[
		{"field":"type_id","value":3},
		{"field":"status_id","value":7},
		{"field":"owner","value":3},
		{"field":"name","value":"???"}
	]`)
	assert.Equal(t, []string{"type", "state", "assignee", "title"}, fields)
}

func TestParseIssueSuggestions_AcceptsStringifiedIDs(t *testing.T) {
	// Models frequently emit ids as strings; dropping those would silently
	// remove the most actionable suggestions.
	got, err := ParseIssueSuggestions(json.RawMessage(`[
		{"field":"type","value":"3"},
		{"field":"assignee","value":" 14 "}
	]`))
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, uint64(3), got[0].Value)
	assert.Equal(t, uint64(14), got[1].Value)
	assert.Equal(t, "#3", got[0].Label, "label defaults to the id when omitted")
}

func TestParseIssueSuggestions_DropsInvalidEntries(t *testing.T) {
	// A bad entry must be skipped, never fail the whole call: the agent would
	// otherwise lose its entire reply over one malformed suggestion.
	fields := parseSuggestions(t, `[
		{"field":"priority","value":"very-high"},
		{"field":"nonsense","value":"x"},
		{"field":"type","value":0},
		{"field":"type","value":"abc"},
		{"field":"title","value":"   "},
		{"field":"priority","value":"low"}
	]`)
	assert.Equal(t, []string{"priority"}, fields)
}

func TestParseIssueSuggestions_DedupesKeepingFirst(t *testing.T) {
	got, err := ParseIssueSuggestions(json.RawMessage(`[
		{"field":"priority","value":"high"},
		{"field":"priority","value":"low"}
	]`))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "high", got[0].Value)
}

func TestParseIssueSuggestions_RejectsNonArray(t *testing.T) {
	_, err := ParseIssueSuggestions(json.RawMessage(`{"field":"title","value":"x"}`))
	assert.Error(t, err)
}

func TestParseIssueSuggestions_EmptyInputIsNotAnError(t *testing.T) {
	got, err := ParseIssueSuggestions(nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestParseIssueSuggestions_TrimsAndKeepsReason(t *testing.T) {
	got, err := ParseIssueSuggestions(json.RawMessage(`[
		{"field":" Priority ","value":" HIGH ","current":" medium ","reason":" ??? 5xx ?? Bug "}
	]`))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "priority", got[0].Field)
	assert.Equal(t, "high", got[0].Value, "priority is lowercased to match storage values")
	assert.Equal(t, "medium", got[0].Current)
	assert.Equal(t, "??? 5xx ?? Bug", got[0].Reason)
}

func TestParseIssueSuggestions_FlagsKeepAsIsProposals(t *testing.T) {
	// Recommending the current value is a valid stance: it carries reasoning
	// ("staying at high until impact is confirmed"). It must survive parsing but
	// be marked so no "apply" action is offered for it.
	got, err := ParseIssueSuggestions(json.RawMessage(`[
		{"field":"priority","value":"high","label":"high","current":"high","reason":"impact unconfirmed"},
		{"field":"title","value":"new title","label":"new title","current":"old title"}
	]`))
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.True(t, got[0].NoOp, "value equal to the stated current value is a keep-as-is")
	assert.False(t, got[1].NoOp)
}

func TestParseIssueSuggestions_NoCurrentValueIsNeverNoOp(t *testing.T) {
	// Without a stated current value there is nothing to compare against, so the
	// proposal must stay actionable rather than being silently neutralised.
	got, err := ParseIssueSuggestions(json.RawMessage(`[
		{"field":"priority","value":"high","label":"high"}
	]`))
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.False(t, got[0].NoOp)
}

func TestSuggestionIsNoOp(t *testing.T) {
	cases := []struct {
		name    string
		current string
		value   interface{}
		label   string
		want    bool
	}{
		{"label matches current", "high", "high", "high", true},
		{"case and whitespace insensitive", " High ", "high", "HIGH", true},
		{"value matches when label is prose", "urgent", "urgent", "urgent (P0)", true},
		{"different value", "low", "high", "high", false},
		{"no current value", "", "high", "high", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, suggestionIsNoOp(tc.current, tc.value, tc.label))
		})
	}
}

func TestEncodeSuggestions(t *testing.T) {
	assert.Nil(t, encodeSuggestions(nil), "empty list must stay NULL, not the string 'null'")

	raw := encodeSuggestions([]model.IssueSuggestion{{Field: "priority", Value: "high"}})
	require.NotNil(t, raw)
	assert.True(t, json.Valid(raw))

	var back []model.IssueSuggestion
	require.NoError(t, json.Unmarshal(raw, &back))
	assert.Equal(t, "high", back[0].Value)
}
