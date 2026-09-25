package service

import (
	"testing"

	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/stretchr/testify/assert"
)

func TestIsWriteTool(t *testing.T) {
	assert.True(t, isWriteTool("create_issue"))
	assert.True(t, isWriteTool("update_issue"))
	assert.True(t, isWriteTool("add_comment"))
	assert.False(t, isWriteTool("search_issues"))
	assert.False(t, isWriteTool("get_issue"))
}

func TestFilterReadOnlyTools_ExcludesWrites(t *testing.T) {
	tools := []llm.Tool{
		{Name: "search_issues"},
		{Name: "create_issue"},
		{Name: "update_issue"},
		{Name: "get_issue"},
		{Name: "add_comment"},
		{Name: "list_states"},
	}
	filtered := filterReadOnlyTools(tools)
	names := make([]string, len(filtered))
	for i, t := range filtered {
		names[i] = t.Name
	}
	assert.Equal(t, []string{"search_issues", "get_issue", "list_states"}, names)
}
