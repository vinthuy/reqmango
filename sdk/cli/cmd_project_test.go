package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects" || r.URL.Query().Get("workspace_id") != "2" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "NewProj" || body["identifier"] != "NP" {
			t.Errorf("unexpected body %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 9, "name": "NewProj", "identifier": "NP"})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("project", "create", "--name", "NewProj", "--identifier", "NP", "--workspace", "2"); err != nil {
		t.Fatalf("project create: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("NP")) {
		t.Fatalf("unexpected output %q", out.String())
	}
}

func TestProjectShow_ByIdentifier(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/projects" && r.Method == "GET":
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "name": "Demo", "identifier": "DEMO", "workspace_id": 2}})
		case r.URL.Path == "/api/v1/projects/5":
			json.NewEncoder(w).Encode(map[string]any{"id": 5, "name": "Demo", "identifier": "DEMO", "workspace_id": 2, "total_issues": 42})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("project", "show", "DEMO"); err != nil {
		t.Fatalf("project show: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Demo")) {
		t.Fatalf("unexpected output %q", out.String())
	}
}
