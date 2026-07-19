package loop

import "encoding/json"

// SprintGuardianPreset returns the pre-configured Loop definition for the Sprint Guardian.
func SprintGuardianPreset() json.RawMessage {
	def := map[string]interface{}{
		"name":        "sprint-guardian",
		"description": "Sprint 自主守护Agent — 每日检查、自动调整、风险预警",
		"version":     "1.0",
		"goal":        "Sprint进度偏差 < 10% AND 无人过载",
		"max_iterations": 3,
		"max_tokens":     30000,
		"max_duration_sec": 3600,
		"trigger": map[string]interface{}{
			"type":     "cron",
			"schedule": "0 9 * * 1-5",
		},
		"actions": []string{
			"analyze_progress",
			"detect_blockers",
			"check_workload",
			"suggest_rebalance",
			"auto_rebalance",
			"notify_stakeholders",
			"generate_daily_digest",
			"generate_sprint_review",
		},
		"notifications": []map[string]interface{}{
			{
				"channel": "in_app",
				"on":      []string{"blocker_detected", "risk_escalated", "daily_digest"},
			},
		},
		"budget": map[string]interface{}{
			"max_tokens_per_day":  50000,
			"max_cost_per_sprint": 2.00,
			"on_budget_critical":  "notify_admin",
		},
	}
	raw, _ := json.Marshal(def)
	return raw
}
