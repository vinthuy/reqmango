package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentActivitySummary_UsesResponseContent(t *testing.T) {
	got := agentActivitySummary("  已将优先级调整为高。  ", []string{"update_issue({})"})
	assert.Equal(t, "已将优先级调整为高。", got)
}

func TestAgentActivitySummary_FallsBackToToolCount(t *testing.T) {
	got := agentActivitySummary("", []string{"get_issue({})", "update_issue({})"})
	assert.Equal(t, "执行了 2 个工具调用", got)
}

func TestAgentActivitySummary_EmptyEverything(t *testing.T) {
	assert.Equal(t, "未产生输出", agentActivitySummary("   ", nil))
}

func TestAgentActivitySummary_TruncatesByRune(t *testing.T) {
	got := agentActivitySummary(strings.Repeat("中", 600), nil)
	assert.True(t, strings.HasSuffix(got, "..."))
	assert.Equal(t, agentSummaryMaxRunes+3, len([]rune(got)))
	assert.True(t, json.Valid([]byte(`"`+got+`"`)), "must stay valid UTF-8")
}

func TestToolErrorJSON_EscapesQuotes(t *testing.T) {
	out := toolErrorJSON(`field "issue_id" is required`)
	var parsed map[string]string
	require.NoError(t, json.Unmarshal([]byte(out), &parsed))
	assert.Equal(t, `field "issue_id" is required`, parsed["error"])
}

func TestWithSuggestionNotice_FlagsClaimedButMissingSuggestions(t *testing.T) {
	for _, body := range []string{
		"## 已提交的一键采纳建议\n- priority → medium",
		"已通过 suggest_issue_changes 提交建议。",
		"Submitted via SUGGEST_ISSUE_CHANGES.",
	} {
		got := withSuggestionNotice(body, nil)
		assert.True(t, strings.HasPrefix(got, body), body)
		assert.True(t, strings.HasSuffix(got, agentNoSuggestionsNotice), body)
	}
}

func TestWithSuggestionNotice_LeavesHonestRepliesAlone(t *testing.T) {
	plain := "分析完毕：优先级建议保持 high。"
	assert.Equal(t, plain, withSuggestionNotice(plain, nil))

	claimed := "已提交的一键采纳建议见下方。"
	stored := []model.IssueSuggestion{{Field: "priority", Value: "medium"}}
	assert.Equal(t, claimed, withSuggestionNotice(claimed, stored))
}

func TestFormatToolCall_ShowsReadableJSON(t *testing.T) {
	got := formatToolCall("list_states", json.RawMessage(`{"issue_id": 81}`))
	assert.Equal(t, `list_states({"issue_id": 81})`, got)
}

func TestAgentIssueContextBlock_IncludesFacts(t *testing.T) {
	ic := IssueAnalysisContext{
		Identifier:  "CORE-704",
		Name:        "登录页白屏",
		Priority:    "high",
		State:       "In Progress",
		Description: "打开 /login 后白屏",
	}
	block := agentIssueContextBlock(6444, ic)
	assert.Contains(t, block, "issue_id=6444")
	assert.Contains(t, block, "CORE-704")
	assert.Contains(t, block, "登录页白屏")
	assert.Contains(t, block, "打开 /login 后白屏")
}
