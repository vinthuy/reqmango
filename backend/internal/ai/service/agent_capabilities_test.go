package service

import (
	"testing"

	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/stretchr/testify/assert"
)

func toolNames(tools []llm.Tool) []string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = t.Name
	}
	return names
}

func allTestTools() []llm.Tool {
	return []llm.Tool{
		{Name: "search_issues"}, {Name: "get_issue"}, {Name: "get_issue_activities"},
		{Name: "create_issue"}, {Name: "update_issue"}, {Name: "suggest_issue_changes"},
		{Name: "add_comment"}, {Name: "list_issue_types"}, {Name: "list_states"},
		{Name: "list_labels"}, {Name: "get_project_stats"},
	}
}

// An agent that can propose field changes must also be able to look up the
// values it is allowed to propose. Without these it guesses ids from memory and
// proposes issue types and transitions the project rejects.
func TestToolsForCapabilities_ProposingAgentsCanLookUpLegalValues(t *testing.T) {
	for _, cap := range []string{"update", "analyze", "comment"} {
		names := toolNames(toolsForCapabilities([]string{cap}, allTestTools()))

		assert.Contains(t, names, "list_states", cap)
		assert.Contains(t, names, "list_issue_types", cap)
		if cap != "comment" {
			assert.Contains(t, names, "suggest_issue_changes", cap)
		}
	}
}

func TestToolsForCapabilities_ReadOnlyAgentGetsNoMutatingTools(t *testing.T) {
	names := toolNames(toolsForCapabilities([]string{"search"}, allTestTools()))

	assert.ElementsMatch(t, []string{"search_issues", "get_issue", "get_issue_activities"}, names)
	assert.NotContains(t, names, "update_issue")
	assert.NotContains(t, names, "suggest_issue_changes")
}

func TestToolsForCapabilities_EmptyOrAllMeansNoRestriction(t *testing.T) {
	all := allTestTools()

	assert.Len(t, toolsForCapabilities(nil, all), len(all))
	assert.Len(t, toolsForCapabilities([]string{"all"}, all), len(all))
	assert.Len(t, toolsForCapabilities([]string{"search", "all"}, all), len(all))
}

func TestToolsForCapabilities_AcceptsToolNames(t *testing.T) {
	names := toolNames(toolsForCapabilities([]string{"get_issue", "list_labels"}, allTestTools()))

	assert.ElementsMatch(t, []string{"get_issue", "list_labels"}, names,
		"a tool name grants exactly that tool, not the read-only fallback")
}

func TestToolsForCapabilities_MixesCategoriesAndToolNames(t *testing.T) {
	names := toolNames(toolsForCapabilities([]string{"search", "list_labels"}, allTestTools()))

	assert.ElementsMatch(t, []string{"search_issues", "get_issue", "get_issue_activities", "list_labels"}, names)
}

func TestToolsForCapabilities_ProposalToolBringsLookupsRegardlessOfSpelling(t *testing.T) {
	names := toolNames(toolsForCapabilities([]string{"get_issue", "suggest_issue_changes"}, allTestTools()))

	assert.Contains(t, names, "list_states")
	assert.Contains(t, names, "list_issue_types")
	assert.NotContains(t, names, "update_issue")
}

func TestToolsForCapabilities_IgnoresUnknownEntriesAlongsideKnownOnes(t *testing.T) {
	names := toolNames(toolsForCapabilities([]string{"triage", "get_issue"}, allTestTools()))

	assert.Equal(t, []string{"get_issue"}, names)
}

// The built-in agents declare tool names. If one of them stops matching a real
// tool, the agent silently loses it; if none match, the agent falls back to the
// broad read-only set and its configuration means nothing.
func TestPMAgentDefs_EveryCapabilityIsARealTool(t *testing.T) {
	real := map[string]bool{}
	for _, tl := range (&AIService{}).GetTools() {
		real[tl.Name] = true
	}
	for _, def := range PMAgentDefs() {
		for _, c := range def.Capabilities {
			assert.True(t, real[c], "%s declares unknown tool %q", def.Name, c)
		}
		assert.NotContains(t, def.Capabilities, "add_comment",
			"%s: the reply is already posted as a comment", def.Name)
	}
}

func TestPMAgentDefs_TriageCanProposeLegalChanges(t *testing.T) {
	var triage PMAgentDef
	for _, def := range PMAgentDefs() {
		if def.Name == "请求分诊" {
			triage = def
		}
	}
	names := toolNames(toolsForCapabilities(triage.Capabilities, (&AIService{}).GetTools()))

	for _, want := range []string{"get_issue", "suggest_issue_changes", "list_states", "list_issue_types"} {
		assert.Contains(t, names, want)
	}
	for _, forbidden := range []string{"update_issue", "create_issue", "add_comment"} {
		assert.NotContains(t, names, forbidden, "triage only proposes")
	}
}

func TestToolsForCapabilities_UnknownCapabilityStillYieldsSomethingUsable(t *testing.T) {
	// An agent configured with a typo'd capability should not end up with an
	// empty tool set that silently does nothing.
	names := toolNames(toolsForCapabilities([]string{"nonsense"}, allTestTools()))

	assert.NotEmpty(t, names)
	assert.NotContains(t, names, "update_issue")
	assert.NotContains(t, names, "create_issue")
	assert.NotContains(t, names, "add_comment")
}
