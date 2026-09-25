package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeAutomationPreview_AllowsSupportedActions(t *testing.T) {
	raw := map[string]interface{}{
		"name":         "高优自动评论",
		"description":  "创建高优先级时提醒",
		"trigger_type": "issue.created",
		"conditions": []interface{}{
			map[string]interface{}{"field": "priority", "operator": "in", "value": []interface{}{"urgent", "high"}},
		},
		"actions": []interface{}{
			map[string]interface{}{"type": "add_comment", "value": "请尽快处理"},
			map[string]interface{}{"type": "dispatch_agent", "value": nil, "agent_name": "请求分诊", "field": "请分诊"},
			map[string]interface{}{"type": "call_webhook", "value": "https://evil.example"},
			map[string]interface{}{"type": "set_priority", "value": "high"},
		},
	}
	agents := map[string]uint64{"请求分诊": 7, "交付风险": 8}

	out, err := SanitizeAutomationPreview(raw, agents)
	require.NoError(t, err)
	assert.Equal(t, "高优自动评论", out.Name)
	assert.Equal(t, "issue.created", out.TriggerType)
	require.Len(t, out.Actions, 3)
	assert.Equal(t, "add_comment", out.Actions[0].Type)
	assert.Equal(t, "dispatch_agent", out.Actions[1].Type)
	assert.EqualValues(t, 7, out.Actions[1].Value)
	assert.Equal(t, "set_priority", out.Actions[2].Type)
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "call_webhook")
}

func TestSanitizeAutomationPreview_RejectsUnsupportedTrigger(t *testing.T) {
	raw := map[string]interface{}{
		"name":         "定时规则",
		"trigger_type": "scheduled",
		"actions": []interface{}{
			map[string]interface{}{"type": "add_comment", "value": "ping"},
		},
	}
	out, err := SanitizeAutomationPreview(raw, nil)
	require.NoError(t, err)
	assert.Equal(t, "issue.created", out.TriggerType) // safe default
	require.NotEmpty(t, out.Warnings)
	assert.Contains(t, out.Warnings[0], "scheduled")
}

func TestSanitizeAutomationPreview_RequiresNameAndAction(t *testing.T) {
	_, err := SanitizeAutomationPreview(map[string]interface{}{
		"trigger_type": "issue.created",
		"actions":      []interface{}{},
	}, nil)
	assert.Error(t, err)
}
