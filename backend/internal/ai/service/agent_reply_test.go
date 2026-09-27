package service

import (
	"encoding/json"
	"strings"
	"testing"

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
