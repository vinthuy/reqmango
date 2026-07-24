package loop

import (
	"encoding/json"
	"testing"
)

func TestSprintGuardianPreset(t *testing.T) {
	raw := SprintGuardianPreset()
	if len(raw) == 0 {
		t.Fatal("preset should not be empty")
	}

	var def map[string]interface{}
	if err := json.Unmarshal(raw, &def); err != nil {
		t.Fatalf("failed to unmarshal preset: %v", err)
	}

	// Verify required fields
	assertField(t, def, "name", "sprint-guardian")
	assertField(t, def, "version", "1.0")

	// Verify goal
	goal, ok := def["goal"].(string)
	if !ok || goal == "" {
		t.Error("goal should be a non-empty string")
	}

	// Verify iteration limits
	maxIter, ok := def["max_iterations"].(float64)
	if !ok || maxIter <= 0 {
		t.Error("max_iterations should be a positive number")
	}

	// Verify actions
	actions, ok := def["actions"].([]interface{})
	if !ok || len(actions) == 0 {
		t.Error("actions should be a non-empty array")
	}

	// Verify expected actions
	expectedActions := map[string]bool{
		"analyze_progress":       true,
		"detect_blockers":        true,
		"check_workload":         true,
		"suggest_rebalance":      true,
		"auto_rebalance":         true,
		"notify_stakeholders":    true,
		"generate_daily_digest":  true,
		"generate_sprint_review": true,
	}
	foundActions := make(map[string]bool)
	for _, a := range actions {
		if s, ok := a.(string); ok {
			foundActions[s] = true
		}
	}
	for expected := range expectedActions {
		if !foundActions[expected] {
			t.Errorf("missing expected action: %s", expected)
		}
	}

	// Verify trigger
	trigger, ok := def["trigger"].(map[string]interface{})
	if !ok {
		t.Error("trigger should be a map")
	} else {
		if trigger["type"] != "cron" {
			t.Errorf("trigger type = %v, want 'cron'", trigger["type"])
		}
	}

	// Verify notifications
	notifications, ok := def["notifications"].([]interface{})
	if !ok || len(notifications) == 0 {
		t.Error("notifications should be a non-empty array")
	}

	// Verify budget
	budget, ok := def["budget"].(map[string]interface{})
	if !ok {
		t.Error("budget should be a map")
	} else {
		if budget["max_tokens_per_day"] == nil {
			t.Error("budget should have max_tokens_per_day")
		}
		if budget["max_cost_per_sprint"] == nil {
			t.Error("budget should have max_cost_per_sprint")
		}
	}
}

func assertField(t *testing.T, m map[string]interface{}, key string, expected interface{}) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Errorf("missing field: %s", key)
		return
	}
	if got != expected {
		t.Errorf("%s = %v, want %v", key, got, expected)
	}
}
