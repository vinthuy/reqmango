package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/reqmango/backend/internal/issuetype"
	"github.com/reqmango/backend/internal/model"
	"github.com/reqmango/backend/internal/rql"
)

// aiSearchRQLFields are the fields both the RQL executor and the FilterBar
// understand. The executor treats unknown fields as `1 = 1`, so a field outside
// this set would silently match every issue.
var aiSearchRQLFields = map[string]bool{
	"sequence_id":   true,
	"name":          true,
	"state_id":      true,
	"state_group":   true,
	"priority":      true,
	"assignee_id":   true,
	"label":         true,
	"cycle_id":      true,
	"module_id":     true,
	"issue_type_id": true,
	"start_date":    true,
	"target_date":   true,
	"created_at":    true,
	"updated_at":    true,
}

const aiSearchMaxResults = 20

// validateAISearchRQL accepts only queries the FilterBar can round-trip: they
// must parse completely, use allowlisted fields, and avoid NOT (which the
// FilterBar cannot display).
func validateAISearchRQL(q string) error {
	q = strings.TrimSpace(q)
	if q == "" {
		return fmt.Errorf("empty query")
	}
	// Template variables are resolved by the FilterBar; substitute a
	// placeholder so the lexer accepts them.
	node, err := rql.ParseStrict(rql.ResolveTemplateVars(q, 1))
	if err != nil {
		return err
	}
	if node == nil {
		return fmt.Errorf("empty query")
	}
	return checkAISearchNode(node)
}

func checkAISearchNode(node rql.Node) error {
	var field string
	switch n := node.(type) {
	case *rql.BinaryExpr:
		if err := checkAISearchNode(n.Left); err != nil {
			return err
		}
		return checkAISearchNode(n.Right)
	case *rql.NotExpr:
		return fmt.Errorf("NOT is not supported")
	case *rql.Comparison:
		field = n.Field
	case *rql.LikeExpr:
		field = n.Field
	case *rql.InExpr:
		field = n.Field
	case *rql.NullCheck:
		field = n.Field
	default:
		return fmt.Errorf("unsupported expression %T", node)
	}
	if !aiSearchRQLFields[field] {
		return fmt.Errorf("unsupported field %q", field)
	}
	return nil
}

// translateSearchRQL asks the model for a filter over this project's real
// states, types, labels and members. It returns ok=false when the model could
// not express the request or produced something that fails validation.
func (s *AIService) translateSearchRQL(ctx context.Context, query string, actx *AIContext) (q, explanation string, ok bool) {
	prompt := fmt.Sprintf(`把用户的自然语言需求翻译成 reqmango 的 RQL 筛选条件。

可用字段（只能用这些）：
- name：标题，用 LIKE，如 name LIKE "%%登录%%"
- priority：urgent | high | medium | low | none
- state_id：状态 ID；state_group：backlog | unstarted | started | completed | cancelled
- issue_type_id：类型 ID；label：标签 ID；assignee_id：负责人用户 ID（"我的"用 $CURRENT_USER）
- cycle_id、module_id、sequence_id
- start_date、target_date、created_at、updated_at：日期 "YYYY-MM-DD"，可用 $TODAY、$ONE_WEEK_AGO、$END_OF_WEEK

语法：field = 值、!=、>、<、>=、<=、field IN (a, b)、field LIKE "%%x%%"、field IS NULL、field IS NOT NULL，
用 AND / OR 组合，可加括号。字符串值一律加双引号，如 state_group IN ("backlog", "started")。不要用 NOT，不要排序。

本项目的状态（ID:名称(分组)）：%s
类型（ID:名称）：%s
标签（ID:名称）：%s
成员（用户ID:名称）：%s

用户需求：%s

只输出纯 JSON：{"rql": "筛选条件", "explanation": "一句中文说明筛了什么"}。
如果需求无法用上述字段表达，rql 输出空字符串。`,
		s.searchStatesInfo(actx), s.searchTypesInfo(actx), s.searchLabelsInfo(actx), s.searchMembersInfo(actx), query)

	content, err := s.llm.Complete(ctx, "你是查询翻译器，只输出纯 JSON。", prompt)
	if err != nil {
		return "", "", false
	}
	out := extractJSON(content)
	if out == nil {
		return "", "", false
	}
	q = strings.TrimSpace(getStrArg(out, "rql", ""))
	if q == "" {
		return "", "", false
	}
	if err := validateAISearchRQL(q); err != nil {
		log.Printf("ai search: rejected generated RQL %q: %v", q, err)
		return "", "", false
	}
	return q, strings.TrimSpace(getStrArg(out, "explanation", "")), true
}

// runSearchRQL executes a validated query in the project, returning the same
// issue shape as the search_issues tool.
func (s *AIService) runSearchRQL(q string, actx *AIContext) ([]map[string]interface{}, error) {
	issues, _, err := rql.NewRQLService().SearchIssuesWithUser(s.db, actx.ProjectID, q, 1, aiSearchMaxResults, actx.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(issues))
	for i, iss := range issues {
		out[i] = map[string]interface{}{
			"id":          iss.ID,
			"name":        iss.Name,
			"sequence_id": iss.SequenceID,
			"priority":    iss.Priority,
			"state_id":    iss.StateID,
		}
	}
	return out, nil
}

func (s *AIService) searchStatesInfo(actx *AIContext) string {
	var states []model.State
	s.db.Where("project_id = ? AND is_active = ?", actx.ProjectID, true).Order("sequence").Find(&states)
	parts := make([]string, len(states))
	for i, st := range states {
		parts[i] = fmt.Sprintf("%d:%s(%s)", st.ID, st.Name, st.Group)
	}
	return strings.Join(parts, ", ")
}

func (s *AIService) searchTypesInfo(actx *AIContext) string {
	projectID := actx.ProjectID
	types, err := issuetype.Visible(s.db, actx.WorkspaceID, &projectID)
	if err != nil {
		return ""
	}
	parts := make([]string, len(types))
	for i, t := range types {
		parts[i] = fmt.Sprintf("%d:%s", t.ID, t.Name)
	}
	return strings.Join(parts, ", ")
}

func (s *AIService) searchLabelsInfo(actx *AIContext) string {
	var labels []model.Label
	s.db.Where("project_id = ?", actx.ProjectID).Find(&labels)
	parts := make([]string, len(labels))
	for i, l := range labels {
		parts[i] = fmt.Sprintf("%d:%s", l.ID, l.Name)
	}
	return strings.Join(parts, ", ")
}

func (s *AIService) searchMembersInfo(actx *AIContext) string {
	var members []model.ProjectMember
	s.db.Preload("User").Where("project_id = ? AND is_active = ?", actx.ProjectID, true).Find(&members)
	parts := make([]string, len(members))
	for i, m := range members {
		name := m.User.DisplayName
		if name == "" {
			name = m.User.Username
		}
		parts[i] = fmt.Sprintf("%d:%s", m.UserID, name)
	}
	return strings.Join(parts, ", ")
}
