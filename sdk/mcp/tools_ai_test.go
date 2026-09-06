package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/reqmango/tools/client"
)

func TestAITools_ToolCountIs27(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	names := listToolNames(t, httpSrv.URL, sessionID)
	if len(names) != 27 {
		t.Fatalf("expected 27 tools total, got %d: %v", len(names), names)
	}
	for _, w := range []string{"ai_search", "ai_chat", "list_agents", "dispatch_agent", "get_agent_task", "get_cycle", "get_cycle_burndown", "list_comments"} {
		if !contains(names, w) {
			t.Errorf("missing tool %s", w)
		}
	}
}

func contains(hay []string, needle string) bool {
	for _, s := range hay {
		if s == needle {
			return true
		}
	}
	return false
}

func TestAIChat_AggregatesStream(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/projects/5/ai/chat": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			for _, ch := range []map[string]any{
				{"type": "text", "content": "Hello"},
				{"type": "done", "thread_id": 99},
			} {
				b, _ := json.Marshal(ch)
				fmt.Fprintf(w, "data: %s\n\n", b)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "ai_chat", map[string]any{
		"project_id": 5, "message": "hello",
	})
	if isError || len(texts) != 1 {
		t.Fatalf("ai_chat: texts=%v isError=%v", texts, isError)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(texts[0]), &out); err != nil {
		t.Fatalf("ai_chat output is not JSON: %v", err)
	}
	if out["text"] != "Hello" || out["thread_id"] != float64(99) {
		t.Fatalf("unexpected ai_chat output %v", out)
	}
}

func TestGetAgentTask_MergesLogs(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/workspaces/2/agent-tasks/15": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"id": 15, "title": "T1", "status": "running", "progress": 50})
		},
		"/api/v1/workspaces/2/agent-tasks/15/logs": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "message": "working"}})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "get_agent_task", map[string]any{
		"workspace_id": 2, "task_id": 15,
	})
	if isError || len(texts) != 1 {
		t.Fatalf("get_agent_task: texts=%v isError=%v", texts, isError)
	}
	if !strings.Contains(texts[0], `"status": "running"`) || !strings.Contains(texts[0], `"message": "working"`) {
		t.Fatalf("get_agent_task should merge logs, got %s", texts[0])
	}
}
