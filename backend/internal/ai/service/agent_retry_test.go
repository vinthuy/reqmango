package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSuggestingLLM calls suggest_issue_changes on its first request and then
// answers; lastBody captures the most recent request for assertions.
func fakeSuggestingLLM(t *testing.T) (*llm.LLMClient, *int32, *string) {
	t.Helper()
	var calls int32
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		b, _ := io.ReadAll(r.Body)
		lastBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			fmt.Fprint(w, `{"choices":[{"finish_reason":"tool_calls","message":{"content":"",
				"tool_calls":[{"id":"c1","type":"function","function":{"name":"suggest_issue_changes","arguments":"{}"}}]}}]}`)
			return
		}
		fmt.Fprint(w, `{"choices":[{"finish_reason":"stop","message":{"content":"已提交"}}]}`)
	}))
	t.Cleanup(srv.Close)
	return llm.NewLLMClient("k-live", "deepseek-chat", srv.URL, "deepseek"), &calls, &lastBody
}

func recordingExecutor(suggestions *[]model.IssueSuggestion) llm.ToolExecutor {
	return func(name string, _ json.RawMessage) (string, error) {
		if name == "suggest_issue_changes" {
			*suggestions = append(*suggestions, model.IssueSuggestion{Field: "priority", Value: "medium"})
		}
		return `{"recorded":1}`, nil
	}
}

var proposingTools = []llm.Tool{{Name: "get_issue"}, {Name: "suggest_issue_changes"}}

func TestRetryClaimedSuggestions_AsksTheAgentToSubmitWhatItClaimed(t *testing.T) {
	client, calls, lastBody := fakeSuggestingLLM(t)
	s := &AgentService{llm: client}
	suggestions := make([]model.IssueSuggestion, 0)

	s.retryClaimedSuggestions(context.Background(), "sys", "分诊", "## 已提交建议（待你一键采纳）\n- 优先级 → medium",
		proposingTools, recordingExecutor(&suggestions), &suggestions)

	require.Len(t, suggestions, 1, "the retry recorded the suggestion the reply promised")
	assert.EqualValues(t, 2, *calls)
	assert.True(t, strings.Contains(*lastBody, "suggest_issue_changes"), "the nudge names the tool")
	assert.True(t, strings.Contains(*lastBody, "待你一键采纳"), "the agent sees its own reply")
}

func TestRetryClaimedSuggestions_SkipsWhenNothingWasClaimed(t *testing.T) {
	client, calls, _ := fakeSuggestingLLM(t)
	s := &AgentService{llm: client}
	suggestions := make([]model.IssueSuggestion, 0)

	s.retryClaimedSuggestions(context.Background(), "sys", "分诊", "分析完毕，优先级保持 high。",
		proposingTools, recordingExecutor(&suggestions), &suggestions)

	assert.Empty(t, suggestions)
	assert.EqualValues(t, 0, *calls)
}

func TestRetryClaimedSuggestions_SkipsWhenSuggestionsExist(t *testing.T) {
	client, calls, _ := fakeSuggestingLLM(t)
	s := &AgentService{llm: client}
	suggestions := []model.IssueSuggestion{{Field: "title", Value: "x"}}

	s.retryClaimedSuggestions(context.Background(), "sys", "分诊", "已提交一键采纳建议",
		proposingTools, recordingExecutor(&suggestions), &suggestions)

	assert.Len(t, suggestions, 1)
	assert.EqualValues(t, 0, *calls)
}

func TestRetryClaimedSuggestions_SkipsAgentsThatCannotPropose(t *testing.T) {
	client, calls, _ := fakeSuggestingLLM(t)
	s := &AgentService{llm: client}
	suggestions := make([]model.IssueSuggestion, 0)

	s.retryClaimedSuggestions(context.Background(), "sys", "分诊", "已提交一键采纳建议",
		[]llm.Tool{{Name: "get_issue"}}, recordingExecutor(&suggestions), &suggestions)

	assert.Empty(t, suggestions)
	assert.EqualValues(t, 0, *calls)
}
