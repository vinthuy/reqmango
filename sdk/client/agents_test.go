package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDispatchAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agents/7/dispatch" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["task"] != "triage new bugs" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "agent_id": 7, "action": "dispatch", "agent_name": "Triage"})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	act, err := c.DispatchAgent(context.Background(), 2, 7, "triage new bugs", nil, nil)
	if err != nil || act.Action != "dispatch" || act.AgentName != "Triage" {
		t.Fatalf("DispatchAgent: %v %+v", err, act)
	}
}

func TestGetAgentTask(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agent-tasks/15" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 15, "title": "T1", "status": "running", "progress": 50})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	task, err := c.GetAgentTask(context.Background(), 2, 15)
	if err != nil || task.Status != "running" || task.Progress != 50 {
		t.Fatalf("GetAgentTask: %v %+v", err, task)
	}
}

func TestListAgents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agents" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]any{{
			"id": 7, "workspace_id": 2, "name": "Triage", "agent_type": "builtin",
			"capabilities": []string{"triaging"}, "status": "active",
		}})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	agents, err := c.ListAgents(context.Background(), 2)
	if err != nil || len(agents) != 1 || agents[0].Name != "Triage" || agents[0].AgentType != "builtin" {
		t.Fatalf("ListAgents: %v %+v", err, agents)
	}
}

func TestGetAgentTaskLogs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agent-tasks/15/logs" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "task_id": 15, "level": "info", "message": "started"},
			{"id": 2, "task_id": 15, "level": "error", "message": "boom"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL+"/api/v1", "t")
	logs, err := c.GetAgentTaskLogs(context.Background(), 2, 15)
	if err != nil || len(logs) != 2 || logs[1].Level != "error" {
		t.Fatalf("GetAgentTaskLogs: %v %+v", err, logs)
	}
}
