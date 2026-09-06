package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListIssueTypes_PathAndQuery verifies path + query assembly, including
// the optional project_id (set only when non-zero).
func TestListIssueTypes_PathAndQuery(t *testing.T) {
	cases := []struct {
		name        string
		workspaceID uint64
		projectID   uint64
		wantQuery   string
	}{
		{"no project filter", 2, 0, "workspace_id=2"},
		{"project filter", 2, 7, "project_id=7&workspace_id=2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/issue-types" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				if r.URL.RawQuery != tc.wantQuery {
					t.Errorf("unexpected query %s", r.URL.RawQuery)
				}
				json.NewEncoder(w).Encode([]map[string]any{{
					"id": 1, "name": "Bug", "workspace_id": 2, "project_id": 7,
				}})
			}))
			defer srv.Close()

			c := New(srv.URL+"/api/v1", "t")
			types, err := c.ListIssueTypes(context.Background(), tc.workspaceID, tc.projectID)
			if err != nil {
				t.Fatalf("ListIssueTypes: %v", err)
			}
			if len(types) != 1 || types[0].Name != "Bug" {
				t.Fatalf("unexpected issue types %+v", types)
			}
		})
	}
}

// TestListStates_Decodes verifies JSON decoding of the workflow states list.
func TestListStates_Decodes(t *testing.T) {
	cases := []struct {
		name      string
		projectID uint64
		body      string
		wantLen   int
		wantName  string
		wantID    uint64
	}{
		{
			name:      "single state",
			projectID: 5,
			body:      `[{"id":1,"name":"Todo","color":"#ff0000","group":"started","sequence":1,"is_default":true,"is_active":true,"project_id":5,"workspace_id":2}]`,
			wantLen:   1,
			wantName:  "Todo",
			wantID:    1,
		},
		{
			name:      "two states",
			projectID: 5,
			body:      `[{"id":1,"name":"Todo","color":"#ff0000","group":"started","sequence":1,"is_default":true,"is_active":true,"project_id":5,"workspace_id":2},{"id":2,"name":"Done","color":"#00ff00","group":"completed","sequence":2,"is_default":false,"is_active":true,"project_id":5,"workspace_id":2}]`,
			wantLen:   2,
			wantName:  "Done",
			wantID:    2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/projects/5/settings/states" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c := New(srv.URL+"/api/v1", "t")
			states, err := c.ListStates(context.Background(), tc.projectID)
			if err != nil {
				t.Fatalf("ListStates: %v", err)
			}
			if len(states) != tc.wantLen {
				t.Fatalf("expected %d states, got %d", tc.wantLen, len(states))
			}
			last := states[len(states)-1]
			if last.Name != tc.wantName || last.ID != tc.wantID {
				t.Fatalf("unexpected state %+v", last)
			}
			if !states[0].IsDefault {
				t.Fatalf("expected first state to be default, got %+v", states[0])
			}
		})
	}
}
