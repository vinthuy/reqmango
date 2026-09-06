package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListProjects_PathAndQuery verifies path + required workspace_id query
// assembly (GET /projects?workspace_id=...).
func TestListProjects_PathAndQuery(t *testing.T) {
	cases := []struct {
		name        string
		workspaceID uint64
		wantQuery   string
	}{
		{"workspace 2", 2, "workspace_id=2"},
		{"workspace 42", 42, "workspace_id=42"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/projects" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				if r.URL.Query().Get("workspace_id") == "" || r.URL.RawQuery != tc.wantQuery {
					t.Errorf("unexpected query %q", r.URL.RawQuery)
				}
				json.NewEncoder(w).Encode([]map[string]any{{
					"id": 5, "name": "Demo", "identifier": "DEMO", "workspace_id": 2,
				}})
			}))
			defer srv.Close()

			c := New(srv.URL+"/api/v1", "t")
			projects, err := c.ListProjects(context.Background(), tc.workspaceID)
			if err != nil {
				t.Fatalf("ListProjects: %v", err)
			}
			if len(projects) != 1 || projects[0].ID != 5 || projects[0].Identifier != "DEMO" {
				t.Fatalf("unexpected projects %+v", projects)
			}
		})
	}
}

// TestCreateProject_PathAndBody verifies POST /projects with workspace_id query.
func TestCreateProject_PathAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("workspace_id") != "2" {
			t.Errorf("expected workspace_id=2, got %q", r.URL.RawQuery)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Demo" || body["identifier"] != "DEMO" {
			t.Errorf("unexpected body %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id": 10, "name": "Demo", "identifier": "DEMO", "workspace_id": 2,
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	proj, err := c.CreateProject(context.Background(), 2, &ProjectCreateRequest{
		Name: "Demo", Identifier: "DEMO",
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if proj.ID != 10 || proj.Identifier != "DEMO" {
		t.Fatalf("unexpected project %+v", proj)
	}
}

// TestListWorkspaces_Decodes verifies JSON decoding of the workspace list shape.
func TestListWorkspaces_Decodes(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantLen  int
		wantName string
		wantID   uint64
	}{
		{
			name:     "single workspace",
			body:     `[{"id":1,"name":"Acme","slug":"acme","created_at":"2026-09-06T00:00:00Z","updated_at":"2026-09-07T00:00:00Z"}]`,
			wantLen:  1,
			wantName: "Acme",
			wantID:   1,
		},
		{
			name:     "two workspaces",
			body:     `[{"id":1,"name":"Acme","slug":"acme","created_at":"2026-09-06T00:00:00Z","updated_at":"2026-09-07T00:00:00Z"},{"id":2,"name":"Beta","slug":"beta","created_at":"2026-09-06T00:00:00Z","updated_at":"2026-09-07T00:00:00Z"}]`,
			wantLen:  2,
			wantName: "Beta",
			wantID:   2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/workspaces" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c := New(srv.URL+"/api/v1", "t")
			ws, err := c.ListWorkspaces(context.Background())
			if err != nil {
				t.Fatalf("ListWorkspaces: %v", err)
			}
			if len(ws) != tc.wantLen {
				t.Fatalf("expected %d workspaces, got %d", tc.wantLen, len(ws))
			}
			last := ws[len(ws)-1]
			if last.Name != tc.wantName || last.ID != tc.wantID {
				t.Fatalf("unexpected workspace %+v", last)
			}
		})
	}
}
