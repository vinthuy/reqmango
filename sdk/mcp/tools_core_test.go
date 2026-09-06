package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/reqmango/tools/client"
)

// backend returns an httptest server with the routes the core tools exercise.
func backend(t *testing.T, routes map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := routes[r.URL.Path]
		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCoreTools_ToolCountAndNames(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	names := listToolNames(t, httpSrv.URL, sessionID)
	want := []string{
		"list_workspaces", "list_projects", "get_project",
		"create_issue", "list_issues", "get_issue", "update_issue", "search_issues",
		"add_comment", "list_cycles", "get_cycle_progress", "add_issue_to_cycle",
		"list_members", "get_states", "get_labels", "list_issue_types",
		"list_notifications", "list_pages", "get_page",
	}
	// >=（不是 ==）：Task 12 还会注册 5 个 AI 工具，总数变 24。
	if len(names) < len(want) {
		t.Fatalf("expected at least %d core tools, got %d: %v", len(want), len(names), names)
	}
	for _, w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing tool %s", w)
		}
	}
}

func TestListWorkspaces(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/workspaces": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 2, "name": "Acme", "slug": "acme"}})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "list_workspaces", map[string]any{})
	if isError || len(texts) != 1 || !strings.Contains(texts[0], `"slug": "acme"`) {
		t.Fatalf("list_workspaces: texts=%v isError=%v", texts, isError)
	}
}

func TestCreateIssue(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/issues": func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("project_id") != "5" || q.Get("workspace_id") != "2" {
				t.Errorf("unexpected query %s", r.URL.RawQuery)
			}
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] != "Login broken" || body["priority"] != "high" {
				t.Errorf("unexpected body %v", body)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high"})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "create_issue", map[string]any{
		"project_id": 5, "workspace_id": 2, "name": "Login broken", "priority": "high",
	})
	if isError || len(texts) != 1 || !strings.Contains(texts[0], `"sequence_id": 42`) {
		t.Fatalf("create_issue: texts=%v isError=%v", texts, isError)
	}
}

func TestGetIssue_WithCode(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/projects": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "identifier": "DEMO", "workspace_id": 2}})
		},
		"/api/v1/issues": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Total-Count", "1")
			json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "sequence_id": 42, "name": "x"}})
		},
		"/api/v1/issues/11": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42})
		},
		"/api/v1/comments/issue/11": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]any{"comments": []map[string]any{{"id": 1, "body": "repro"}}, "total": 1})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "get_issue", map[string]any{
		"issue": "DEMO-42", "workspace_id": 2,
	})
	if isError || len(texts) != 1 {
		t.Fatalf("get_issue: texts=%v isError=%v", texts, isError)
	}
	if !strings.Contains(texts[0], `"body": "repro"`) || !strings.Contains(texts[0], `"id": 11`) {
		t.Fatalf("get_issue should merge comments, got %s", texts[0])
	}
}

func TestToolError_401(t *testing.T) {
	srv := backend(t, map[string]http.HandlerFunc{
		"/api/v1/workspaces": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "token expired"})
		},
	})
	httpSrv, sessionID := newTestMCPServer(t, client.New(srv.URL+"/api/v1", "reqmango_pat_test"))

	texts, isError := callTool(t, httpSrv.URL, sessionID, "list_workspaces", map[string]any{})
	if !isError || len(texts) != 1 || !strings.Contains(texts[0], "reqmango auth login") {
		t.Fatalf("expected 401 hint, got texts=%v isError=%v", texts, isError)
	}
}
