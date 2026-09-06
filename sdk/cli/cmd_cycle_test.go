package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCycleBurndown_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/cycles/3/burndown" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"cycle_id": 3, "cycle_name": "S1", "total_issues": 10, "is_on_track": true, "daily_points": []map[string]any{}})
	}))
	defer srv.Close()

	out, run := newCLI(t, srv.URL+"/api/v1", nil)
	if err := run("cycle", "burndown", "3", "--output", "json"); err != nil {
		t.Fatalf("cycle burndown: %v", err)
	}
	var b map[string]any
	if err := json.Unmarshal(out.Bytes(), &b); err != nil {
		t.Fatalf("expected JSON, got %q: %v", out.String(), err)
	}
	if b["cycle_id"] != float64(3) {
		t.Fatalf("unexpected output %v", b)
	}
}
