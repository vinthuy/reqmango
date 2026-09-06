package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIssues_ReadsTotalHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("rql") != `priority = "high"` {
			t.Errorf("rql not forwarded: %s", r.URL.RawQuery)
		}
		w.Header().Set("X-Total-Count", "7")
		json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "bug"}})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	res, err := c.ListIssues(context.Background(), IssueListOptions{ProjectID: 1, RQL: `priority = "high"`})
	if err != nil || res.Total != 7 || len(res.Items) != 1 {
		t.Fatalf("ListIssues: %v %+v", err, res)
	}
}

func TestCreateIssue_PathAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("project_id") != "5" || q.Get("workspace_id") != "2" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Login broken" || body["type_id"] != float64(9) {
			t.Errorf("unexpected body %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Login broken", "sequence_id": 42})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	iss, err := c.CreateIssue(context.Background(), 5, 2, &CreateIssueRequest{
		Name: "Login broken", TypeID: uintPtr(9),
	})
	if err != nil || iss.ID != 11 || iss.SequenceID != 42 {
		t.Fatalf("CreateIssue: %v %+v", err, iss)
	}
}

func TestResolveIssueCode(t *testing.T) {
	var step int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/projects":
			json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "identifier": "DEMO", "workspace_id": 2}})
		case r.URL.Path == "/api/v1/issues":
			if r.URL.Query().Get("project_id") != "5" || r.URL.Query().Get("search") != "42" {
				t.Errorf("unexpected issue query %s", r.URL.RawQuery)
			}
			step++
			w.Header().Set("X-Total-Count", "1")
			json.NewEncoder(w).Encode([]map[string]any{{"id": 11, "sequence_id": 42}})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	id, err := c.ResolveIssueCode(context.Background(), 2, "DEMO-42")
	if err != nil || id != 11 {
		t.Fatalf("ResolveIssueCode: %v id=%d", err, id)
	}
}

func TestGetIssue_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues/11" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": 11, "name": "Login broken", "sequence_id": 42, "state_name": "In Progress",
			"assignees": []map[string]any{{"id": 7, "display_name": "Ada", "email": "ada@x.io"}},
			"labels":    []uint64{1, 2},
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	iss, err := c.GetIssue(context.Background(), 11)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if iss.Name != "Login broken" || iss.StateName != "In Progress" {
		t.Fatalf("unexpected issue %+v", iss)
	}
	if len(iss.Assignees) != 1 || iss.Assignees[0].DisplayName != "Ada" {
		t.Fatalf("unexpected assignees %+v", iss.Assignees)
	}
	if len(iss.Labels) != 2 || iss.Labels[1] != 2 {
		t.Fatalf("unexpected labels %+v", iss.Labels)
	}
}

func TestUpdateIssue_PutPathAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/api/v1/issues/11" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Renamed" || body["state_id"] != float64(3) {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 11, "name": "Renamed", "state_id": 3})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	name := "Renamed"
	iss, err := c.UpdateIssue(context.Background(), 11, &UpdateIssueRequest{Name: &name, StateID: uintPtr(3)})
	if err != nil || iss.Name != "Renamed" {
		t.Fatalf("UpdateIssue: %v %+v", err, iss)
	}
}

func TestSearchIssues_Query(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issues/search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("workspace_id") != "2" || q.Get("query") != "login" || q.Get("project_id") != "5" || q.Get("limit") != "10" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode([]map[string]any{{
			"id": 11, "name": "Login broken", "sequence_id": 42,
			"project_identifier": "DEMO", "project_id": 5, "workspace_slug": "acme",
		}})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	p := uintPtr(5)
	res, err := c.SearchIssues(context.Background(), 2, "login", p, 10)
	if err != nil || len(res) != 1 || res[0].ProjectIdentifier != "DEMO" {
		t.Fatalf("SearchIssues: %v %+v", err, res)
	}
}

func TestAddComment_PostBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/api/v1/comments" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["issue_id"] != float64(11) || body["body"] != "looks good" || body["parent_id"] != float64(4) {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 8, "issue_id": 11, "body": "looks good", "parent_id": 4})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	cm, err := c.AddComment(context.Background(), 11, "looks good", uintPtr(4))
	if err != nil || cm.ID != 8 || cm.Body != "looks good" {
		t.Fatalf("AddComment: %v %+v", err, cm)
	}
}

func TestListComments_WrappedShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/comments/issue/11" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" || r.URL.Query().Get("page_size") != "5" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"comments": []map[string]any{{"id": 8, "issue_id": 11, "body": "looks good"}},
			"total":    12,
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	comments, total, err := c.ListComments(context.Background(), 11, 2, 5)
	if err != nil || total != 12 || len(comments) != 1 {
		t.Fatalf("ListComments: %v total=%d %+v", err, total, comments)
	}
}

func uintPtr(v uint64) *uint64 { return &v }
