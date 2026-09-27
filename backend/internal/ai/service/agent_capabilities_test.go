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

func TestToolsForCapabilities_UnknownCapabilityStillYieldsSomethingUsable(t *testing.T) {
	// An agent configured with a typo'd capability should not end up with an
	// empty tool set that silently does nothing.
	names := toolNames(toolsForCapabilities([]string{"nonsense"}, allTestTools()))

	assert.NotEmpty(t, names)
	assert.NotContains(t, names, "update_issue")
	assert.NotContains(t, names, "create_issue")
	assert.NotContains(t, names, "add_comment")
}
