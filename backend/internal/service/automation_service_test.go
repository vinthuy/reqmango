package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateConditions(t *testing.T) {
	evaluator := NewDefaultConditionEvaluator(nil)

	tests := []struct {
		name       string
		conditions []Condition
		context    map[string]interface{}
		want       bool
	}{
		{"empty conditions", nil, map[string]interface{}{}, true},
		{"empty slice", []Condition{}, map[string]interface{}{}, true},
		{"equals match", []Condition{{Field: "priority", Operator: "equals", Value: "urgent"}}, map[string]interface{}{"priority": "urgent"}, true},
		{"equals no match", []Condition{{Field: "priority", Operator: "equals", Value: "high"}}, map[string]interface{}{"priority": "low"}, false},
		{"not_equals match", []Condition{{Field: "priority", Operator: "not_equals", Value: "low"}}, map[string]interface{}{"priority": "high"}, true},
		{"not_equals no match", []Condition{{Field: "priority", Operator: "not_equals", Value: "urgent"}}, map[string]interface{}{"priority": "urgent"}, false},
		{"contains match", []Condition{{Field: "title", Operator: "contains", Value: "bug"}}, map[string]interface{}{"title": "fix login bug"}, true},
		{"contains no match", []Condition{{Field: "title", Operator: "contains", Value: "feature"}}, map[string]interface{}{"title": "fix login bug"}, false},
		{"in operator match", []Condition{{Field: "state", Operator: "in", Value: []interface{}{"todo", "in_progress"}}}, map[string]interface{}{"state": "todo"}, true},
		{"in operator no match", []Condition{{Field: "state", Operator: "in", Value: []interface{}{"done", "cancelled"}}}, map[string]interface{}{"state": "todo"}, false},
		{"gt operator match", []Condition{{Field: "priority_score", Operator: "gt", Value: float64(5)}}, map[string]interface{}{"priority_score": float64(10)}, true},
		{"gt operator no match", []Condition{{Field: "priority_score", Operator: "gt", Value: float64(10)}}, map[string]interface{}{"priority_score": float64(5)}, false},
		{"lt operator match", []Condition{{Field: "priority_score", Operator: "lt", Value: float64(10)}}, map[string]interface{}{"priority_score": float64(5)}, true},
		{"is_empty match", []Condition{{Field: "description", Operator: "is_empty"}}, map[string]interface{}{"description": ""}, true},
		{"is_not_empty match", []Condition{{Field: "description", Operator: "is_not_empty"}}, map[string]interface{}{"description": "hello"}, true},
		{"is_not_empty no match", []Condition{{Field: "description", Operator: "is_not_empty"}}, map[string]interface{}{"description": ""}, false},
		{"matches_regex match", []Condition{{Field: "title", Operator: "matches_regex", Value: "^BUG-\\d+"}}, map[string]interface{}{"title": "BUG-123"}, true},
		{"matches_regex no match", []Condition{{Field: "title", Operator: "matches_regex", Value: "^BUG-\\d+"}}, map[string]interface{}{"title": "FEAT-456"}, false},
		{"missing field", []Condition{{Field: "nonexistent", Operator: "equals", Value: "x"}}, map[string]interface{}{"priority": "high"}, false},
		{"multiple conditions all match", []Condition{
			{Field: "priority", Operator: "equals", Value: "high"},
			{Field: "state_group", Operator: "equals", Value: "started"},
		}, map[string]interface{}{"priority": "high", "state_group": "started"}, true},
		{"multiple conditions one mismatch", []Condition{
			{Field: "priority", Operator: "equals", Value: "high"},
			{Field: "state_group", Operator: "equals", Value: "completed"},
		}, map[string]interface{}{"priority": "high", "state_group": "started"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluator.Evaluate(tt.conditions, tt.context)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name, jsonStr string
		wantErr       bool
	}{
		{"empty string", "", false},
		{"empty array", "[]", false},
		{"valid conditions", `[{"field":"p","operator":"equals","value":"v"}]`, false},
		{"valid actions", `[{"type":"add_comment","value":"hello"}]`, false},
		{"valid object", `{"key":"value"}`, false},
		{"invalid json", `{bad}`, true},
		{"garbage", `not json at all`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJSON(tt.jsonStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToUint64(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  uint64
		ok    bool
	}{
		{"float64", float64(42), 42, true},
		{"int", int(100), 100, true},
		{"int64", int64(200), 200, true},
		{"uint64", uint64(300), 300, true},
		{"string number", "456", 456, true},
		{"string not number", "abc", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toUint64(tt.input)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAutomationCreateRequest_JSON(t *testing.T) {
	req := &AutomationCreateRequest{
		Name:        "Test Rule",
		TriggerType: "issue.created",
		Conditions:  `[{"field":"priority","operator":"equals","value":"high"}]`,
		Actions:     `[{"type":"add_comment","value":"test"}]`,
	}
	body, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "issue.created")
	assert.Contains(t, string(body), "add_comment")
}
