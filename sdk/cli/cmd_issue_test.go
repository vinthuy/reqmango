package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// newCLI prepares a root command against a backend and a temp config.
func newCLI(t *testing.T, srvURL string, cfgPatch *Config) (*bytes.Buffer, func(args ...string) error) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{APIURL: srvURL, PAT: "reqmango_pat_x", WorkspaceID: 2, ProjectID: 5}
	if cfgPatch != nil {
		*cfg = *cfgPatch
	}
	if err := SaveConfig(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	run := func(args ...string) error {
		out.Reset()
		root.SetArgs(append(args, "--config", cfgPath))
		return root.Execute()
	}
	return &out, run
}

func TestIssueList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("X-Total-Count", "1")
		json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high", "state_name": "Todo"}})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("issue", "list", "--project", "5"); err != nil {
		t.Fatalf("issue list: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Login broken")) {
		t.Fatalf("expected issue in table output, got %q", out.String())
	}
}

func TestIssueShow_ByCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/projects":
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "identifier": "DEMO", "workspace_id": 2}})
		case "/api/v1/issues":
			w.Header().Set("X-Total-Count", "1")
			json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "sequence_id": 42}})
		case "/api/v1/issues/11":
			json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42, "priority": "high", "state_name": "Todo"})
		case "/api/v1/comments/issue/11":
			json.NewEncoder(w).Encode(map[string]any{"comments": []map[string]any{}, "total": 0})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("issue", "show", "DEMO-42", "--comments"); err != nil {
		t.Fatalf("issue show: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("DEMO-42")) {
		t.Fatalf("expected code in output, got %q", out.String())
	}
}

func TestIssueCreate_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/projects/5":
			json.NewEncoder(w).Encode(map[string]any{"id": 5, "identifier": "DEMO", "workspace_id": 2})
		case "/api/v1/issues":
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["name"] != "New bug" || body["priority"] != "high" {
				t.Errorf("unexpected body %v", body)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": 12, "name": "New bug", "sequence_id": 43, "priority": "high"})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("issue", "create", "--project", "5", "--title", "New bug", "--priority", "high", "--output", "json"); err != nil {
		t.Fatalf("issue create: %v", err)
	}
	var created map[string]any
	if err := json.Unmarshal(out.Bytes(), &created); err != nil {
		t.Fatalf("expected JSON output, got %q: %v", out.String(), err)
	}
	if created["id"] != float64(12) {
		t.Fatalf("unexpected output %v", created)
	}
}
