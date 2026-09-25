package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reqmango/backend/internal/model"
)

// Allowed automation triggers for B2 NL preview (no scheduled/webhook-only).
var allowedAutomationTriggers = map[string]bool{
	"issue.created":         true,
	"issue.updated":         true,
	"issue.state_changed":    true,
	"issue.assigned":        true,
	"issue.due_soon":        true,
	"issue.due_date_passed": true,
	"cycle.started":         true,
	"cycle.ended":           true,
	"comment.added":         true,
	// scheduled intentionally excluded from B2 NL generation
}

// Allowed action types for B2 NL preview (design: option B).
var allowedAutomationActions = map[string]bool{
	"add_comment":    true,
	"set_priority":   true,
	"change_state":   true,
	"assign_to":      true,
	"dispatch_agent": true,
}

// AIAutomationPreviewRequest is the NL → rule draft request.
type AIAutomationPreviewRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

// AIAutomationAction is one sanitized action in a preview.
type AIAutomationAction struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value,omitempty"`
	Field string      `json:"field,omitempty"`
}

// AIAutomationCondition is one condition in a preview.
type AIAutomationCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// AIAutomationPreviewResponse is returned to the UI for RuleBuilder prefill.
type AIAutomationPreviewResponse struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	TriggerType string                   `json:"trigger_type"`
	Conditions  []AIAutomationCondition  `json:"conditions"`
	Actions     []AIAutomationAction     `json:"actions"`
	Warnings    []string                 `json:"warnings,omitempty"`
	// RuleShape mirrors persisted automation fields so the frontend can pass
	// it straight into AutomationRuleBuilder as :rule.
	TriggerTypeJSON string `json:"trigger_type_json"`
	ConditionsJSON  string `json:"conditions_json"`
	ActionsJSON     string `json:"actions_json"`
}

// SanitizeAutomationPreview validates and normalizes LLM JSON into a safe draft.
// agentByName maps PM agent display names → IDs (may be empty).
func SanitizeAutomationPreview(raw map[string]interface{}, agentByName map[string]uint64) (*AIAutomationPreviewResponse, error) {
	if raw == nil {
		return nil, fmt.Errorf("empty preview")
	}

	out := &AIAutomationPreviewResponse{
		Conditions: []AIAutomationCondition{},
		Actions:    []AIAutomationAction{},
		Warnings:   []string{},
	}

	if name, ok := raw["name"].(string); ok {
		out.Name = strings.TrimSpace(name)
	}
	if desc, ok := raw["description"].(string); ok {
		out.Description = strings.TrimSpace(desc)
	}

	trigger := ""
	switch v := raw["trigger_type"].(type) {
	case string:
		trigger = strings.TrimSpace(v)
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok {
			trigger = strings.TrimSpace(t)
		}
	}
	if trigger == "" || !allowedAutomationTriggers[trigger] {
		if trigger != "" {
			out.Warnings = append(out.Warnings, fmt.Sprintf("unsupported trigger %q; defaulted to issue.created", trigger))
		}
		trigger = "issue.created"
	}
	out.TriggerType = trigger

	if conds, ok := raw["conditions"].([]interface{}); ok {
		for _, c := range conds {
			cm, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			field, _ := cm["field"].(string)
			op, _ := cm["operator"].(string)
			if field == "" {
				continue
			}
			if op == "" {
				op = "equals"
			}
			out.Conditions = append(out.Conditions, AIAutomationCondition{
				Field:    field,
				Operator: op,
				Value:    cm["value"],
			})
		}
	}

	if acts, ok := raw["actions"].([]interface{}); ok {
		for _, a := range acts {
			am, ok := a.(map[string]interface{})
			if !ok {
				continue
			}
			typ, _ := am["type"].(string)
			typ = strings.TrimSpace(typ)
			if typ == "" {
				continue
			}
			if !allowedAutomationActions[typ] {
				out.Warnings = append(out.Warnings, fmt.Sprintf("dropped unsupported action %q (use RuleBuilder manually)", typ))
				continue
			}
			act := AIAutomationAction{
				Type:  typ,
				Value: am["value"],
				Field: stringField(am, "field"),
			}
			if typ == "dispatch_agent" {
				act = resolveDispatchAgent(act, am, agentByName, &out.Warnings)
			}
			out.Actions = append(out.Actions, act)
		}
	}

	if out.Name == "" {
		return nil, fmt.Errorf("preview missing name")
	}
	if len(out.Actions) == 0 {
		return nil, fmt.Errorf("preview has no supported actions")
	}

	trigObj := map[string]interface{}{"type": out.TriggerType}
	tb, _ := json.Marshal(trigObj)
	out.TriggerTypeJSON = string(tb)
	cb, _ := json.Marshal(out.Conditions)
	out.ConditionsJSON = string(cb)
	ab, _ := json.Marshal(out.Actions)
	out.ActionsJSON = string(ab)

	return out, nil
}

func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func resolveDispatchAgent(act AIAutomationAction, raw map[string]interface{}, agentByName map[string]uint64, warnings *[]string) AIAutomationAction {
	// Prefer explicit numeric value
	switch v := act.Value.(type) {
	case float64:
		if v > 0 {
			act.Value = uint64(v)
			return act
		}
	case int:
		if v > 0 {
			act.Value = uint64(v)
			return act
		}
	case uint64:
		if v > 0 {
			return act
		}
	case string:
		if id, ok := agentByName[v]; ok && id > 0 {
			act.Value = id
			return act
		}
		if v != "" {
			*warnings = append(*warnings, fmt.Sprintf("unknown agent %q; select in RuleBuilder", v))
			act.Value = nil
		}
	}

	name := stringField(raw, "agent_name")
	if name == "" {
		name = stringField(raw, "agent")
	}
	if name != "" {
		if id, ok := agentByName[name]; ok && id > 0 {
			act.Value = id
			return act
		}
		// fuzzy: substring match on known names
		for k, id := range agentByName {
			if strings.Contains(name, k) || strings.Contains(k, name) {
				act.Value = id
				return act
			}
		}
		*warnings = append(*warnings, fmt.Sprintf("could not resolve agent %q; select in RuleBuilder", name))
		act.Value = nil
	}
	return act
}

// AutomationPreview turns a natural-language intent into a sanitized automation draft.
func (s *AIService) AutomationPreview(ctx context.Context, req *AIAutomationPreviewRequest, actx *AIContext) (*AIAutomationPreviewResponse, error) {
	if req == nil || strings.TrimSpace(req.Prompt) == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	var states []model.State
	s.db.Where("project_id = ? AND is_active = ?", actx.ProjectID, true).Order("sequence").Find(&states)
	stateInfo := make([]string, 0, len(states))
	for _, st := range states {
		stateInfo = append(stateInfo, fmt.Sprintf("%d:%s", st.ID, st.Name))
	}

	var members []model.ProjectMember
	s.db.Preload("User").Where("project_id = ?", actx.ProjectID).Find(&members)
	memberInfo := make([]string, 0, len(members))
	for _, m := range members {
		name := ""
		if m.User.DisplayName != "" {
			name = m.User.DisplayName
		} else {
			name = m.User.Email
		}
		memberInfo = append(memberInfo, fmt.Sprintf("%d:%s", m.UserID, name))
	}

	var agents []model.Agent
	s.db.Where("workspace_id = ? AND status = ?", actx.WorkspaceID, "active").Find(&agents)
	agentByName := map[string]uint64{}
	agentInfo := make([]string, 0, len(agents))
	for _, a := range agents {
		agentByName[a.Name] = a.ID
		agentInfo = append(agentInfo, fmt.Sprintf("%d:%s", a.ID, a.Name))
	}

	allowedTriggers := []string{
		"issue.created", "issue.updated", "issue.state_changed", "issue.assigned",
		"issue.due_soon", "issue.due_date_passed", "cycle.started", "cycle.ended", "comment.added",
	}
	allowedActions := []string{"add_comment", "set_priority", "change_state", "assign_to", "dispatch_agent"}

	system := `你是自动化规则配置助手。只输出纯JSON，不要markdown标记或额外文字。`
	userPrompt := fmt.Sprintf(`将用户的自然语言意图解析为一条项目管理自动化规则草稿。

项目：%s（%s）
可用状态（ID:名称）：%s
可用成员（ID:显示名）：%s
可用 Agent（ID:名称）：%s

允许的 trigger_type（只能选其一）：%s
允许的 action.type（只能用这些）：%s

规则：
1. name 简短中文标题；description 一句话说明。
2. trigger_type 必须是允许列表中的字符串（不要用 scheduled / call_webhook）。
3. conditions 可选；常用 field: priority/state_id/assignee_id；operator: equals|not_equals|in|contains。
4. actions 至少一条。dispatch_agent 时填 agent_name（用上面 Agent 名称）和 field（给 Agent 的提示词）；value 可先 null。
5. set_priority 的 value 只能是 urgent|high|medium|low|none。
6. change_state / assign_to 的 value 用上面列出的状态ID / 成员ID。
7. 若用户要 Webhook、定时、邮件等不支持能力，不要捏造；可在 JSON 顶层加 "warnings": ["..."]。

用户意图：
%s

只输出JSON：
{
  "name": "",
  "description": "",
  "trigger_type": "issue.created",
  "conditions": [],
  "actions": [{"type":"add_comment","value":"..."}],
  "warnings": []
}`,
		actx.ProjectName, actx.ProjectIdentifier,
		strings.Join(stateInfo, ", "),
		strings.Join(memberInfo, ", "),
		strings.Join(agentInfo, ", "),
		strings.Join(allowedTriggers, ", "),
		strings.Join(allowedActions, ", "),
		strings.TrimSpace(req.Prompt),
	)

	content, err := s.llm.Complete(ctx, system, userPrompt)
	if err != nil {
		return nil, err
	}

	raw := extractJSON(content)
	if raw == nil {
		return nil, fmt.Errorf("failed to parse automation preview JSON")
	}
	if w, ok := raw["warnings"].([]interface{}); ok {
		// keep LLM warnings; Sanitize will append more
		_ = w
	}

	out, err := SanitizeAutomationPreview(raw, agentByName)
	if err != nil {
		return nil, err
	}
	// Merge LLM warnings if any
	if w, ok := raw["warnings"].([]interface{}); ok {
		for _, item := range w {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out.Warnings = append(out.Warnings, s)
			}
		}
	}
	return out, nil
}
