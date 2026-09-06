package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCycles_WrappedShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/5/cycles" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"items":  []map[string]any{{"id": 1, "name": "Sprint 1", "status": "active", "progress": 40.5}},
			"total":  1, "limit": 50, "offset": 0,
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	res, err := c.ListCycles(context.Background(), 5, "active", 50, 0)
	if err != nil || res.Total != 1 || len(res.Items) != 1 || res.Items[0].Progress != 40.5 {
		t.Fatalf("ListCycles: %v %+v", err, res)
	}
}

func TestGetCycleBurndown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/burndown" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"cycle_id": 3, "cycle_name": "S1", "total_issues": 10, "is_on_track": true,
			"daily_points": []map[string]any{{"day_index": 0, "actual_remaining": 9.0}},
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	b, err := c.GetCycleBurndown(context.Background(), 3)
	if err != nil || b.IsOnTrack != true || len(b.DailyPoints) != 1 {
		t.Fatalf("GetCycleBurndown: %v %+v", err, b)
	}
}

func TestAddIssueToCycle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/issues" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("issue_id") != "11" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		json.NewEncoder(w).Encode(map[string]any{"cycle_id": 3, "issue_id": 11, "action": "added"})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	if err := c.AddIssueToCycle(context.Background(), 3, 11); err != nil {
		t.Fatalf("AddIssueToCycle: %v", err)
	}
}

func TestGetCycle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": 3, "name": "Sprint 1", "status": "active", "progress": 40.5,
			"total_issues": 10, "completed_issues": 4, "start_date": "2026-09-01",
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	cy, err := c.GetCycle(context.Background(), 3)
	if err != nil || cy.Name != "Sprint 1" || cy.TotalIssues != 10 {
		t.Fatalf("GetCycle: %v %+v", err, cy)
	}
}

func TestGetCycleProgress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/progress" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"cycle_id": 3, "cycle_name": "Sprint 1", "total_issues": 10, "completed_issues": 4, "progress": 40.5,
			"state_breakdown": []map[string]any{{"state": "Done", "group": "completed", "count": 4}},
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	p, err := c.GetCycleProgress(context.Background(), 3)
	if err != nil || p.Progress != 40.5 || len(p.StateBreakdown) != 1 || p.StateBreakdown[0].State != "Done" {
		t.Fatalf("GetCycleProgress: %v %+v", err, p)
	}
}
