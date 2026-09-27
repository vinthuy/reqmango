package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeToolServer answers with a tool call for the first toolRounds requests and
// with finalText afterwards.
func fakeToolServer(t *testing.T, toolRounds int32, finalText string) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if n <= toolRounds {
			fmt.Fprintf(w, `{"choices":[{"finish_reason":"tool_calls","message":{"content":"先查一下",
				"tool_calls":[{"id":"c%d","type":"function","function":{"name":"get_issue","arguments":"{}"}}]}}]}`, n)
			return
		}
		b, _ := json.Marshal(finalText)
		fmt.Fprintf(w, `{"choices":[{"finish_reason":"stop","message":{"content":%s}}]}`, b)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func runTools(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	client := NewLLMClient("k-live", "deepseek-chat", srv.URL, "deepseek")
	resp, err := client.ChatSyncWithTools(context.Background(), "sys", []Message{{Role: "user", Content: "分诊"}},
		[]Tool{{Name: "get_issue"}}, func(string, json.RawMessage) (string, error) { return `{}`, nil })
	if err != nil {
		t.Fatalf("ChatSyncWithTools: %v", err)
	}
	return resp.Content
}

func TestChatSyncWithTools_AllowsMultiStepInvestigations(t *testing.T) {
	srv, _ := fakeToolServer(t, 5, "分诊完成")
	if got := runTools(t, srv); got != "分诊完成" {
		t.Errorf("content = %q, want the final answer after 5 tool rounds", got)
	}
}

func TestChatSyncWithTools_ExhaustedRoundsSayWhy(t *testing.T) {
	srv, calls := fakeToolServer(t, 1000, "")
	got := runTools(t, srv)
	if !strings.Contains(got, "已达到最大工具调用轮数") {
		t.Errorf("content = %q, want the round-limit notice", got)
	}
	if int(*calls) != maxToolRounds {
		t.Errorf("requests = %d, want %d", *calls, maxToolRounds)
	}
}
