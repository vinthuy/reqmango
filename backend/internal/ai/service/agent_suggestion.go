package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reqmango/backend/internal/ai/llm"
	"github.com/reqmango/backend/internal/model"
)

// maxIssueSuggestions bounds a single tool call so one runaway agent reply
// cannot flood a work item with apply buttons.
const maxIssueSuggestions = 12

// suggestionFieldAliases maps the many names an LLM may use onto the canonical
// field names the apply path understands.
var suggestionFieldAliases = map[string]string{
	"title":       "title",
	"name":        "title",
	"标题":          "title",
	"priority":    "priority",
	"优先级":        "priority",
	"type":        "type",
	"type_id":     "type",
	"issue_type":  "type",
	"issue_type_id": "type",
	"类型":          "type",
	"state":       "state",
	"state_id":    "state",
	"status":      "state",
	"status_id":   "state",
	"状态":          "state",
	"assignee":    "assignee",
	"assignee_id": "assignee",
	"assignees":   "assignee",
	"owner":       "assignee",
	"负责人":         "assignee",
	"description": "description",
	"desc":        "description",
	"描述":          "description",
}

var validPriorities = map[string]bool{
	"urgent": true, "high": true, "medium": true, "low": true, "none": true,
}

// suggestIssueChangesTool advertises the structured-suggestion tool. The nested
// array shape is described in prose because llm.SchemaProp has no Items support;
// ParseIssueSuggestions is the actual contract enforcement.
func suggestIssueChangesTool() llm.Tool {
	return llm.Tool{
		Name: "suggest_issue_changes",
		Description: "把「建议修改工作项字段」结构化地提出来交给用户一键采纳。" +
			"分诊、评审、复盘这类场景应当用它，而不要直接调用 update_issue 改数据。\n" +
			"建议列表每项为对象：{\"field\": \"title|priority|type|state|assignee|description\", " +
			"\"value\": 目标值（type/state/assignee 用数字 ID，priority 用 urgent|high|medium|low|none，title/description 用文本）, " +
			"\"label\": \"给人看的目标值(如 Bug (type_id=3))\", \"current\": \"当前值\", \"reason\": \"建议理由\"}。\n" +
			"只提有把握的字段，最多 12 条。",
		InputSchema: &llm.ToolSchema{
			Type: "object",
			Properties: map[string]llm.SchemaProp{
				"issue_id":    {Type: "integer", Description: "工作项 ID"},
				"suggestions": {Type: "array", Description: "建议修改的字段列表（见工具描述中的对象结构）"},
			},
			Required: []string{"issue_id", "suggestions"},
		},
	}
}

// ParseIssueSuggestions validates the raw tool input into storable suggestions.
// Unknown fields, empty values and malformed priorities are dropped rather than
// failing the whole call: a partial suggestion list is still useful to the user,
// and an agent should never lose its reply over one bad entry.
func ParseIssueSuggestions(raw json.RawMessage) ([]model.IssueSuggestion, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("suggestions must be an array of objects")
	}

	out := make([]model.IssueSuggestion, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		field := suggestionFieldAliases[strings.ToLower(strings.TrimSpace(getStrArg(item, "field", "")))]
		if field == "" || seen[field] {
			continue
		}
		value, label, ok := normalizeSuggestionValue(field, item)
		if !ok {
			continue
		}
		seen[field] = true
		// A proposal that keeps the agent's own view of the current value is not
		// dropped: the reasoning ("保持 high，暂不建议升到 urgent") is signal. It is
		// flagged so the UI can offer no apply action for it.
		current := strings.TrimSpace(getStrArg(item, "current", ""))
		out = append(out, model.IssueSuggestion{
			Field:   field,
			Value:   value,
			Label:   label,
			Current: current,
			Reason:  strings.TrimSpace(getStrArg(item, "reason", "")),
			NoOp:    suggestionIsNoOp(current, value, label),
		})
		if len(out) == maxIssueSuggestions {
			break
		}
	}
	return out, nil
}

// normalizeSuggestionValue coerces and range-checks the target value per field.
func normalizeSuggestionValue(field string, item map[string]interface{}) (interface{}, string, bool) {
	label := strings.TrimSpace(getStrArg(item, "label", ""))

	switch field {
	case "title", "description":
		text := strings.TrimSpace(getStrArg(item, "value", ""))
		if text == "" {
			return nil, "", false
		}
		if label == "" {
			label = text
		}
		return text, label, true

	case "priority":
		priority := strings.ToLower(strings.TrimSpace(getStrArg(item, "value", "")))
		if !validPriorities[priority] {
			return nil, "", false
		}
		if label == "" {
			label = priority
		}
		return priority, label, true

	case "type", "state", "assignee":
		// Ids may arrive as JSON numbers, json.Number or numeric strings.
		id := coerceSuggestionID(item["value"])
		if id == 0 {
			return nil, "", false
		}
		if label == "" {
			label = fmt.Sprintf("#%d", id)
		}
		return id, label, true
	}

	return nil, "", false
}

// maxSuggestionIDDigits bounds the digit loop so a pathological string cannot
// overflow the uint64 accumulator and wrap into a bogus id.
const maxSuggestionIDDigits = 18

// coerceSuggestionID accepts the shapes a model may use for an id: a JSON
// number, a json.Number, or a numeric string. Anything else yields 0.
func coerceSuggestionID(v interface{}) uint64 {
	switch n := v.(type) {
	case float64:
		if n > 0 {
			return uint64(n)
		}
	case int:
		if n > 0 {
			return uint64(n)
		}
	case uint64:
		return n
	case json.Number:
		if i, err := n.Int64(); err == nil && i > 0 {
			return uint64(i)
		}
	case string:
		s := strings.TrimSpace(n)
		if s == "" || len(s) > maxSuggestionIDDigits {
			return 0
		}
		var id uint64
		for _, r := range s {
			if r < '0' || r > '9' {
				return 0
			}
			id = id*10 + uint64(r-'0')
		}
		return id
	}
	return 0
}

// suggestionIsNoOp reports whether a proposal matches the agent's own description
// of the current value, i.e. it recommends keeping things as they are.
func suggestionIsNoOp(current string, value interface{}, label string) bool {
	current = strings.ToLower(strings.TrimSpace(current))
	if current == "" {
		return false
	}
	if current == strings.ToLower(strings.TrimSpace(label)) {
		return true
	}
	if s, ok := value.(string); ok && current == strings.ToLower(strings.TrimSpace(s)) {
		return true
	}
	return false
}

// encodeSuggestions serialises suggestions for the JSONB comment column.
// Returns nil (not "null") for an empty list so the column stays NULL.
func encodeSuggestions(suggestions []model.IssueSuggestion) json.RawMessage {
	if len(suggestions) == 0 {
		return nil
	}
	b, err := json.Marshal(suggestions)
	if err != nil {
		return nil
	}
	return b
}
