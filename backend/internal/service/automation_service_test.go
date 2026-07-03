package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateConditions(t *testing.T) {
	svc := &AutomationService{db: nil}

	tests := []struct {
		name       string
		conditions string
		context    map[string]interface{}
		want       bool
	}{
		{"empty conditions", "", map[string]interface{}{}, true},
		{"empty array", "[]", map[string]interface{}{}, true},
		{"equals match", `[{"field":"priority","operator":"equals","value":"urgent"}]`, map[string]interface{}{"priority": "urgent"}, true},
		{"equals no match", `[{"field":"priority","operator":"equals","value":"high"}]`, map[string]interface{}{"priority": "low"}, false},
		{"not_equals match", `[{"field":"priority","operator":"not_equals","value":"low"}]`, map[string]interface{}{"priority": "high"}, true},
		{"not_equals no match", `[{"field":"priority","operator":"not_equals","value":"urgent"}]`, map[string]interface{}{"priority": "urgent"}, false},
		{"contains match", `[{"field":"title","operator":"contains","value":"bug"}]`, map[string]interface{}{"title": "fix login bug"}, true},
		{"contains no match", `[{"field":"title","operator":"contains","value":"feature"}]`, map[string]interface{}{"title": "fix login bug"}, false},
		{"in operator match", `[{"field":"state","operator":"in","value":["todo","in_progress"]}]`, map[string]interface{}{"state": "todo"}, true},
		{"in operator no match", `[{"field":"state","operator":"in","value":["done","cancelled"]}]`, map[string]interface{}{"state": "todo"}, false},
		{"missing field", `[{"field":"nonexistent","operator":"equals","value":"x"}]`, map[string]interface{}{"priority": "high"}, false},
		{"invalid JSON", `not valid json`, map[string]interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.evaluateConditions(tt.conditions, tt.context)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateJSON(t *testing.T) {
	svc := &AutomationService{db: nil}

	tests := []struct {
		name, conditions, actions string
		wantErr                   bool
	}{
		{"both empty", "", "", false},
		{"empty arrays", "[]", "[]", false},
		{"valid conditions", `[{"field":"p","operator":"equals","value":"v"}]`, "", false},
		{"valid actions", "", `[{"type":"add_comment","value":"hello"}]`, false},
		{"invalid conditions", `{bad}`, "", true},
		{"invalid actions", "", `{bad}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateJSON(tt.conditions, tt.actions)
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
		TriggerType: "issue_created",
		Conditions:  `[{"field":"priority","operator":"equals","value":"high"}]`,
		Actions:     `[{"type":"add_comment","value":"test"}]`,
	}
	body, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "issue_created")
	assert.Contains(t, string(body), "add_comment")
}
