package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentDispatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspaces/2/agents/7/dispatch" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["task"] != "triage bugs" {
			t.Errorf("unexpected body %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"id": 1, "agent_id": 7, "action": "dispatch", "agent_name": "Triage"})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("agent", "dispatch", "7", "triage bugs"); err != nil {
		t.Fatalf("agent dispatch: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Triage")) {
		t.Fatalf("unexpected output %q", out.String())
	}
}
