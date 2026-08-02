package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/reqmango/backend/internal/client"
	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// ======== 新架构：事件总线 ========

// Event 表示自动化系统中的事件
type Event struct {
	Type      string                 `json:"type"` // issue_created, state_changed, etc.
	ProjectID uint64                 `json:"project_id"`
	IssueID   uint64                 `json:"issue_id"`
	Context   map[string]interface{} `json:"context"` // old_state, new_state, state_group, priority, etc.
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler 事件处理器接口
type EventHandler func(ctx context.Context, event Event) error

// EventBus 事件总线接口（参考主流事件驱动架构）
type EventBus interface {
	Publish(event Event) error
	Subscribe(triggerType string, handler EventHandler)
}

// InMemoryEventBus 基于内存的事件总线实现（生产环境可替换为 Redis Pub/Sub）
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (bus *InMemoryEventBus) Publish(event Event) error {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	handlers, exists := bus.handlers[event.Type]
	if !exists {
		return nil // 没有订阅者，静默跳过
	}

	// 异步执行所有处理器（参考主流后台任务模式）
	for _, handler := range handlers {
		go func(h EventHandler) {
			ctx := context.Background()
			if err := h(ctx, event); err != nil {
				log.Printf("[EventBus] Handler failed for event %s: %v", event.Type, err)
			}
		}(handler)
	}

	return nil
}

func (bus *InMemoryEventBus) Subscribe(triggerType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[triggerType] = append(bus.handlers[triggerType], handler)
}

// ======== 新架构：规则引擎 ========

// ConditionEvaluator 条件评估器（支持复杂逻辑）
type ConditionEvaluator interface {
	Evaluate(conditions []Condition, context map[string]interface{}) bool
}

// Condition 增强的条件结构（支持 AND/OR/NOT 组合）
type Condition struct {
	Field      string      `json:"field"`
	Operator   string      `json:"operator"` // equals, not_equals, contains, in, gt, lt, is_empty, is_not_empty, matches_regex
	Value      interface{} `json:"value"`
	Logic      string      `json:"logic,omitempty"`      // AND, OR, NOT (用于组合多个条件)
	Conditions []Condition `json:"conditions,omitempty"` // 嵌套条件组
	IsNegated  bool        `json:"is_negated,omitempty"` // 是否取反（NOT）
}

// DefaultConditionEvaluator 默认条件评估器实现（支持复杂逻辑组合和自定义字段）
type DefaultConditionEvaluator struct {
	db *gorm.DB
}

func NewDefaultConditionEvaluator(db *gorm.DB) *DefaultConditionEvaluator {
	return &DefaultConditionEvaluator{db: db}
}

func (e *DefaultConditionEvaluator) Evaluate(conditions []Condition, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	result := e.evaluateConditions(conditions, context)
	return result
}

// evaluateConditions 递归评估条件列表，支持 AND/OR/NOT 逻辑
func (e *DefaultConditionEvaluator) evaluateConditions(conditions []Condition, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	// 默认逻辑为 AND
	logic := "AND"
	if len(conditions) > 0 && conditions[0].Logic != "" {
		logic = conditions[0].Logic
	}

	switch strings.ToUpper(logic) {
	case "OR":
		for _, cond := range conditions {
			if e.evaluateCondition(cond, context) {
				return true
			}
		}
		return false

	case "NOT":
		// NOT 逻辑：取反第一个条件的结果
		if len(conditions) > 0 {
			return !e.evaluateCondition(conditions[0], context)
		}
		return true

	default: // AND
		for _, cond := range conditions {
			if !e.evaluateCondition(cond, context) {
				return false
			}
		}
		return true
	}
}

// evaluateCondition 评估单个条件（可能是叶子条件或条件组）
func (e *DefaultConditionEvaluator) evaluateCondition(cond Condition, context map[string]interface{}) bool {
	result := false

	// 如果有嵌套条件，递归评估
	if len(cond.Conditions) > 0 {
		result = e.evaluateConditions(cond.Conditions, context)
	} else {
		// 叶子条件，评估单个条件
		result = e.evaluateSingleCondition(cond, context)
	}

	// 如果条件被取反，返回相反结果
	if cond.IsNegated {
		return !result
	}

	return result
}

func (e *DefaultConditionEvaluator) evaluateSingleCondition(cond Condition, context map[string]interface{}) bool {
	fieldValue, exists := context[cond.Field]
	if !exists {
		// 尝试从自定义字段获取值
		if strings.HasPrefix(cond.Field, "custom_") {
			fieldValue = e.getCustomFieldValue(cond.Field, context)
			if fieldValue == nil {
				log.Printf("[Automation] Custom field '%s' not found or has no value", cond.Field)
				return false
			}
		} else {
			log.Printf("[Automation] Condition field '%s' not found in context", cond.Field)
			return false
		}
	}

	switch cond.Operator {
	case "equals":
		result := fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", cond.Value)
		log.Printf("[Automation] Condition: field=%s, operator=%s, context_value=%v, condition_value=%v, result=%v",
			cond.Field, cond.Operator, fieldValue, cond.Value, result)
		return result
	case "not_equals":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", cond.Value)
	case "contains":
		strVal, ok := fieldValue.(string)
		if !ok {
			return false
		}
		strSearch, ok := cond.Value.(string)
		if !ok {
			return false
		}
		return strings.Contains(strVal, strSearch)
	case "in":
		// 检查 fieldValue 是否在 cond.Value 列表中
		valueList, ok := cond.Value.([]interface{})
		if !ok {
			return false
		}
		for _, v := range valueList {
			if fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", v) {
				return true
			}
		}
		return false
	case "gt":
		return compareNumeric(fieldValue, cond.Value) > 0
	case "lt":
		return compareNumeric(fieldValue, cond.Value) < 0
	case "is_empty":
		return fieldValue == nil || fieldValue == ""
	case "is_not_empty":
		return fieldValue != nil && fieldValue != ""
	case "matches_regex":
		strVal, ok := fieldValue.(string)
		if !ok {
			return false
		}
		pattern, ok := cond.Value.(string)
		if !ok {
			return false
		}
		matched, _ := regexp.MatchString(pattern, strVal)
		return matched
	default:
		log.Printf("[ConditionEvaluator] Unknown operator: %s", cond.Operator)
		return false
	}
}

// getCustomFieldValue 从数据库获取自定义字段的值
// field格式: custom_{field_id}
func (e *DefaultConditionEvaluator) getCustomFieldValue(fieldName string, context map[string]interface{}) interface{} {
	// 提取字段ID: custom_5 -> 5
	fieldIDStr := strings.TrimPrefix(fieldName, "custom_")
	var fieldID uint64
	_, err := fmt.Sscanf(fieldIDStr, "%d", &fieldID)
	if err != nil {
		log.Printf("[Automation] Invalid custom field format: %s", fieldName)
		return nil
	}

	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		log.Printf("[Automation] issue_id not found in context for custom field lookup")
		return nil
	}

	var fieldValue model.IssueCustomFieldValue
	if err := e.db.Where("issue_id = ? AND field_id = ?", issueID, fieldID).First(&fieldValue).Error; err != nil {
		log.Printf("[Automation] Custom field value not found for issue %d, field %d", issueID, fieldID)
		return nil
	}

	return fieldValue.Value
}

func compareNumeric(a, b interface{}) int {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return 0
	}
	if aFloat > bFloat {
		return 1
	} else if aFloat < bFloat {
		return -1
	}
	return 0
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}

// ======== 新架构：动作执行器 ========

// ActionExecutor 动作执行器接口
type ActionExecutor interface {
	RegisterAction(actionType string, handler ActionHandler)
	Execute(actions []Action, context map[string]interface{}) ([]string, error)
}

// ActionHandler 动作处理器函数类型
type ActionHandler func(action Action, context map[string]interface{}, db *gorm.DB) error

// Action 增强的动作结构
type Action struct {
	Type  string      `json:"type"`
	Field string      `json:"field,omitempty"`
	Value interface{} `json:"value"`
}

// DefaultActionExecutor 默认动作执行器实现
type DefaultActionExecutor struct {
	handlers    map[string]ActionHandler
	db          *gorm.DB
	agentClient *client.AgentClient // 用于 dispatch_agent 动作
}

func NewDefaultActionExecutor(db *gorm.DB, agentClient *client.AgentClient) *DefaultActionExecutor {
	executor := &DefaultActionExecutor{
		handlers:    make(map[string]ActionHandler),
		db:          db,
		agentClient: agentClient,
	}

	// 注册内置动作处理器
	executor.registerBuiltinActions()

	return executor
}

func (e *DefaultActionExecutor) registerBuiltinActions() {
	e.RegisterAction("set_field", e.handleSetField)
	e.RegisterAction("add_comment", e.handleAddComment)
	e.RegisterAction("assign_to", e.handleAssignTo)
	e.RegisterAction("unassign", e.handleUnassign)
	e.RegisterAction("change_state", e.handleChangeState)
	e.RegisterAction("set_priority", e.handleSetPriority)
	e.RegisterAction("archive", e.handleArchive)
	e.RegisterAction("close", handleClose)
	e.RegisterAction("dispatch_agent", e.handleDispatchAgent)
	e.RegisterAction("call_webhook", e.handleCallWebhook)
	e.RegisterAction("rollup_to_parent", e.handleRollupToParent)
}

func (e *DefaultActionExecutor) RegisterAction(actionType string, handler ActionHandler) {
	e.handlers[actionType] = handler
}

func (e *DefaultActionExecutor) Execute(actions []Action, context map[string]interface{}) ([]string, error) {
	var results []string

	for _, action := range actions {
		handler, exists := e.handlers[action.Type]
		if !exists {
			log.Printf("[ActionExecutor] Unknown action type: %s", action.Type)
			continue
		}

		if err := handler(action, context, e.db); err != nil {
			log.Printf("[ActionExecutor] Failed to execute action %s: %v", action.Type, err)
			return results, err
		}

		results = append(results, fmt.Sprintf("Executed %s", action.Type))
	}

	return results, nil
}

// 内置动作处理器实现
func (e *DefaultActionExecutor) handleSetField(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	field := action.Field
	value := action.Value

	// 自定义字段：custom_{field_id}
	if strings.HasPrefix(field, "custom_") {
		fieldIDStr := strings.TrimPrefix(field, "custom_")
		fieldID, err := strconv.ParseUint(fieldIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid custom field id: %s", fieldIDStr)
		}

		// 验证自定义字段存在
		var cf model.CustomField
		if err := db.First(&cf, fieldID).Error; err != nil {
			return fmt.Errorf("custom field %d not found: %w", fieldID, err)
		}

		strValue := fmt.Sprintf("%v", value)

		// Upsert: 如果已存在则更新，否则创建
		var existing model.IssueCustomFieldValue
		err = db.Where("issue_id = ? AND field_id = ?", issueID, fieldID).First(&existing).Error
		if err == nil {
			return db.Model(&existing).Update("value", strValue).Error
		}
		return db.Create(&model.IssueCustomFieldValue{
			IssueID: issueID,
			FieldID: fieldID,
			Value:   strValue,
		}).Error
	}

	// 系统字段
	allowedFields := map[string]bool{
		"priority":    true,
		"state_id":    true,
		"target_date": true,
		"start_date":  true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("field %s is not allowed to be set via automation", field)
	}

	return db.Model(&model.Issue{}).Where("id = ?", issueID).Update(field, value).Error
}

func (e *DefaultActionExecutor) handleAddComment(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	comment, ok := action.Value.(string)
	if !ok {
		return fmt.Errorf("invalid comment value: %v", action.Value)
	}

	return db.Create(&model.Comment{IssueID: issueID, Body: comment}).Error
}

func (e *DefaultActionExecutor) handleAssignTo(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	userID, ok := toUint64(action.Value)
	if !ok {
		return fmt.Errorf("invalid user_id: %v", action.Value)
	}

	// 检查是否已分配
	var count int64
	db.Model(&model.IssueAssignee{}).Where("issue_id = ? AND user_id = ?", issueID, userID).Count(&count)
	if count > 0 {
		return nil // 已分配，跳过
	}

	return db.Create(&model.IssueAssignee{IssueID: issueID, UserID: userID}).Error
}

func (e *DefaultActionExecutor) handleUnassign(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	userID, ok := toUint64(action.Value)
	if !ok {
		return fmt.Errorf("invalid user_id: %v", action.Value)
	}

	return db.Where("issue_id = ? AND user_id = ?", issueID, userID).Delete(&model.IssueAssignee{}).Error
}

func (e *DefaultActionExecutor) handleChangeState(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	projectID, _ := context["project_id"].(uint64)

	var stateID uint64

	if strValue, ok := action.Value.(string); ok {
		var state model.State
		if err := db.Where("project_id = ? AND name = ?", projectID, strValue).First(&state).Error; err != nil {
			return fmt.Errorf("state '%s' not found for project %d", strValue, projectID)
		}
		stateID = state.ID
	} else {
		var ok bool
		stateID, ok = toUint64(action.Value)
		if !ok {
			return fmt.Errorf("invalid state_id: %v", action.Value)
		}
	}

	var newState model.State
	if err := db.First(&newState, stateID).Error; err != nil {
		return fmt.Errorf("state %d not found", stateID)
	}

	updateData := map[string]interface{}{"state_id": stateID}

	if newState.Group == common.StateGroupCompleted {
		now := time.Now()
		updateData["completed_at"] = now
	} else {
		updateData["completed_at"] = nil
	}

	return db.Model(&model.Issue{}).Where("id = ?", issueID).Updates(updateData).Error
}

func (e *DefaultActionExecutor) handleSetPriority(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	priority, ok := action.Value.(string)
	if !ok {
		return fmt.Errorf("invalid priority value: %v", action.Value)
	}

	// 验证优先级值
	validPriorities := map[string]bool{
		"urgent": true,
		"high":   true,
		"medium": true,
		"low":    true,
		"none":   true,
	}

	if !validPriorities[priority] {
		return fmt.Errorf("invalid priority: %s", priority)
	}

	return db.Model(&model.Issue{}).Where("id = ?", issueID).Update("priority", priority).Error
}

func (e *DefaultActionExecutor) handleArchive(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	return db.Model(&model.Issue{}).Where("id = ?", issueID).Update("archived", true).Error
}

func (e *DefaultActionExecutor) handleDispatchAgent(action Action, context map[string]interface{}, db *gorm.DB) error {
	if e.agentClient == nil {
		return fmt.Errorf("agent client not available")
	}

	issueID, _ := context["issue_id"].(uint64)
	projectID, _ := context["project_id"].(uint64)

	// Get agent_id from action value
	agentID, ok := toUint64(action.Value)
	if !ok {
		return fmt.Errorf("invalid agent_id: %v", action.Value)
	}

	// Build task description from action field or default
	task := action.Field
	if task == "" {
		task = fmt.Sprintf("处理工作项 #%d 的自动化触发", issueID)
	}

	var issueIDPtr *uint64
	if issueID > 0 {
		issueIDPtr = &issueID
	}
	var projectIDPtr *uint64
	if projectID > 0 {
		projectIDPtr = &projectID
	}

	// Look up agent to get workspaceID
	var agent model.Agent
	if err := db.First(&agent, agentID).Error; err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// Use system user (user ID 1) as the actor for automation-triggered dispatches
	err := e.agentClient.DispatchAgent(agent.WorkspaceID, agentID, 1, task, issueIDPtr, projectIDPtr, "automation")
	if err != nil {
		log.Printf("[Automation] Agent dispatch failed: agent=%d issue=%d err=%v", agentID, issueID, err)
		return err
	}

	log.Printf("[Automation] Agent dispatched: agent=%d issue=%d task=%s", agentID, issueID, task)
	return nil
}

// handleCallWebhook 调用外部 Webhook
// action.Field = URL (必填)
// action.Value = { method: "POST", headers: {"Content-Type": "application/json"}, body: "..." } 或纯字符串作为 body
func (e *DefaultActionExecutor) handleCallWebhook(action Action, ctxData map[string]interface{}, db *gorm.DB) error {
	url := action.Field
	if url == "" {
		// 尝试从 value map 中获取
		if vMap, ok := action.Value.(map[string]interface{}); ok {
			if u, ok := vMap["url"].(string); ok {
				url = u
			}
		}
	}
	if url == "" {
		return fmt.Errorf("webhook URL is required")
	}

	method := "POST"
	var headers map[string]string
	var bodyStr string

	if vMap, ok := action.Value.(map[string]interface{}); ok {
		if m, ok := vMap["method"].(string); ok && m != "" {
			method = strings.ToUpper(m)
		}
		if h, ok := vMap["headers"].(map[string]interface{}); ok {
			headers = make(map[string]string)
			for k, v := range h {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}
		if b, ok := vMap["body"].(string); ok {
			bodyStr = b
		}
	} else if vStr, ok := action.Value.(string); ok {
		bodyStr = vStr
	}

	// 模板变量替换
	bodyStr = renderWebhookTemplate(bodyStr, ctxData)

	// 默认 headers
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}

	// 构建请求
	var reqBody io.Reader
	if bodyStr != "" {
		reqBody = bytes.NewBufferString(bodyStr)
	}

	httpCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("webhook request build failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Webhook] Call failed: url=%s method=%s err=%v", url, method, err)
		return fmt.Errorf("webhook call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	log.Printf("[Webhook] Called: url=%s method=%s status=%d response=%s", url, method, resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// renderWebhookTemplate 替换 webhook body 中的模板变量
func renderWebhookTemplate(body string, ctxData map[string]interface{}) string {
	if body == "" {
		return body
	}
	result := body
	replacements := map[string]string{
		"{{issue_id}}":     fmt.Sprintf("%v", ctxData["issue_id"]),
		"{{project_id}}":   fmt.Sprintf("%v", ctxData["project_id"]),
		"{{workspace_id}}": fmt.Sprintf("%v", ctxData["workspace_id"]),
		"{{event_type}}":   fmt.Sprintf("%v", ctxData["event_type"]),
		"{{trigger_type}}": fmt.Sprintf("%v", ctxData["trigger_type"]),
	}
	for k, v := range replacements {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}

func handleClose(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	// 查找"已关闭"状态
	var closedState model.State
	if err := db.Where("project_id = (SELECT project_id FROM issues WHERE id = ?) AND group = ?", issueID, common.StateGroupCompleted).First(&closedState).Error; err != nil {
		return fmt.Errorf("closed state not found")
	}

	now := time.Now()
	return db.Model(&model.Issue{}).Where("id = ?", issueID).Updates(map[string]interface{}{
		"state_id":     closedState.ID,
		"completed_at": now,
	}).Error
}

// handleRollupToParent 状态卷积：将子工作项状态聚合到父工作项
func (e *DefaultActionExecutor) handleRollupToParent(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}

	var issue model.Issue
	if err := db.First(&issue, issueID).Error; err != nil {
		return fmt.Errorf("issue %d not found: %w", issueID, err)
	}

	if issue.ParentID == nil {
		return nil // root issue, nothing to rollup
	}

	// Parse rules from action Value
	configMap, ok := action.Value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("rollup_to_parent requires config map as value")
	}
	rulesRaw, ok := configMap["rules"]
	if !ok {
		return fmt.Errorf("rollup_to_parent config missing 'rules' array")
	}
	rules, ok := rulesRaw.([]interface{})
	if !ok || len(rules) == 0 {
		return fmt.Errorf("rollup_to_parent rules must be a non-empty array")
	}

	// Get all children of the parent
	var allChildren []model.Issue
	if err := db.Where("parent_id = ?", *issue.ParentID).Find(&allChildren).Error; err != nil {
		return fmt.Errorf("failed to query children: %w", err)
	}
	totalCount := len(allChildren)

	// Evaluate rules top-to-bottom, first match wins
	for _, ruleRaw := range rules {
		rule, ok := ruleRaw.(map[string]interface{})
		if !ok {
			continue
		}
		condition, _ := rule["condition"].(string)
		if condition == "" || rule["child_state"] == nil || rule["parent_state"] == nil {
			continue
		}

		childStateID, err := resolveStateID(db, rule["child_state"], issue.ProjectID)
		if err != nil {
			log.Printf("[RollupToParent] Failed to resolve child_state: %v", err)
			continue
		}
		parentStateID, err := resolveStateID(db, rule["parent_state"], issue.ProjectID)
		if err != nil {
			log.Printf("[RollupToParent] Failed to resolve parent_state: %v", err)
			continue
		}

		// Count children in target state
		matchCount := 0
		for _, child := range allChildren {
			if child.StateID == childStateID {
				matchCount++
			}
		}

		var matched bool
		switch condition {
		case "all":
			matched = matchCount == totalCount && totalCount > 0
		case "any":
			matched = matchCount > 0
		default:
			log.Printf("[RollupToParent] Unknown condition: %s", condition)
			continue
		}

		if matched {
			var parentState model.State
			if err := db.First(&parentState, parentStateID).Error; err != nil {
				return fmt.Errorf("parent state %d not found: %w", parentStateID, err)
			}
			updateData := map[string]interface{}{"state_id": parentStateID}
			if parentState.Group == common.StateGroupCompleted {
				updateData["completed_at"] = time.Now()
			}
			if err := db.Model(&model.Issue{}).Where("id = ?", *issue.ParentID).Updates(updateData).Error; err != nil {
				return fmt.Errorf("failed to update parent state: %w", err)
			}
			log.Printf("[RollupToParent] Updated parent %d state to %d (%s) [%s: %d/%d]",
				*issue.ParentID, parentStateID, parentState.Name, condition, matchCount, totalCount)
			return nil
		}
	}

	return nil
}

// resolveStateID 将状态引用（名称或ID）解析为 state ID
func resolveStateID(db *gorm.DB, value interface{}, projectID uint64) (uint64, error) {
	switch v := value.(type) {
	case string:
		var state model.State
		// 1) Exact match (project-level)
		if db.Where("project_id = ? AND name = ?", projectID, v).First(&state).Error == nil {
			return state.ID, nil
		}
		// 2) Exact match (workspace-level)
		if db.Where("project_id IS NULL AND name = ?", v).First(&state).Error == nil {
			return state.ID, nil
		}
		// 3) Fuzzy match: name LIKE '%value%' (project-level)
		if db.Where("project_id = ? AND name LIKE ?", projectID, "%"+v+"%").First(&state).Error == nil {
			return state.ID, nil
		}
		// 4) Fuzzy match: name LIKE '%value%' (workspace-level)
		if db.Where("project_id IS NULL AND name LIKE ?", "%"+v+"%").First(&state).Error == nil {
			return state.ID, nil
		}
		return 0, fmt.Errorf("state '%s' not found", v)
	case float64:
		return uint64(v), nil
	default:
		id, ok := toUint64(v)
		if !ok {
			return 0, fmt.Errorf("cannot resolve state from %v", v)
		}
		return id, nil
	}
}

// ======== 重构后的 AutomationService ========

// AutomationService 重构后的自动化服务（参考主流架构）
type AutomationService struct {
	db               *gorm.DB
	eventBus         EventBus
	ruleEngine       ConditionEvaluator
	actionExecutor   ActionExecutor
	executionHistory sync.Map // 保留用于循环检测
	maxExecutions    int
	execWindow       time.Duration
}

func NewAutomationService(db *gorm.DB) *AutomationService {
	eventBus := NewInMemoryEventBus()
	ruleEngine := NewDefaultConditionEvaluator(db)
	actionExecutor := NewDefaultActionExecutor(db, nil) // agentClient set via SetAgentService after construction

	service := &AutomationService{
		db:             db,
		eventBus:       eventBus,
		ruleEngine:     ruleEngine,
		actionExecutor: actionExecutor,
		maxExecutions:  10,
		execWindow:     5 * time.Minute,
	}

	// 注册事件处理器
	service.registerEventHandlers()

	return service
}

// checkWorkspaceAdmin verifies that the caller is an active admin-level member
// of the workspace. Guards automation mutations against privilege escalation.
func (s *AutomationService) checkWorkspaceAdmin(workspaceID, callerID uint64) error {
	var member model.WorkspaceMember
	if err := s.db.Where("workspace_id = ? AND user_id = ? AND is_active = ?", workspaceID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a workspace admin to manage automation rules")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a workspace admin to manage automation rules")
	}
	return nil
}

// checkProjectAdmin verifies that the caller is an active admin-level member
// of the project. Guards automation mutations against privilege escalation.
func (s *AutomationService) checkProjectAdmin(projectID, callerID uint64) error {
	var member model.ProjectMember
	if err := s.db.Where("project_id = ? AND user_id = ? AND is_active = ?", projectID, callerID, true).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.Forbidden("You must be a project admin to manage automation rules")
		}
		return common.Internal("Database error")
	}
	if member.Role < common.RoleAdmin {
		return common.Forbidden("You must be a project admin to manage automation rules")
	}
	return nil
}

func (s *AutomationService) registerEventHandlers() {
	// 订阅所有触发类型的事件
	triggerTypes := []string{
		"issue.created",
		"issue.updated",
		"issue.state_changed",
		"issue.assigned",
		"comment.added",
		"scheduled",
	}

	for _, triggerType := range triggerTypes {
		s.eventBus.Subscribe(triggerType, s.handleAutomationEvent)
	}
}

func (s *AutomationService) handleAutomationEvent(ctx context.Context, event Event) error {
	startTime := time.Now()

	// 循环检测
	key := fmt.Sprintf("%d:%s", event.IssueID, event.Type)
	now := time.Now()

	if lastTime, ok := s.executionHistory.Load(key); ok {
		if lastExecTime, isTime := lastTime.(time.Time); isTime {
			if now.Sub(lastExecTime) < s.execWindow {
				execCount := 0
				s.executionHistory.Range(func(k, v interface{}) bool {
					if kStr, ok := k.(string); ok {
						if strings.HasPrefix(kStr, fmt.Sprintf("%d:", event.IssueID)) {
							if vTime, ok := v.(time.Time); ok && now.Sub(vTime) < s.execWindow {
								execCount++
							}
						}
					}
					return true
				})

				if execCount >= s.maxExecutions {
					log.Printf("[Automation] Cycle detected: issue %d trigger %s executed %d times in %v, skipping",
						event.IssueID, event.Type, execCount, s.execWindow)

					// 记录跳过的执行历史
					s.recordExecutionHistory(event, nil, 0, "skipped", "Cycle detected", startTime)
					return nil
				}
			}
		}
	}

	// 记录本次执行
	s.executionHistory.Store(key, now)

	// Ensure issue_id is always in context for action handlers
	if event.Context == nil {
		event.Context = make(map[string]interface{})
	}
	event.Context["issue_id"] = event.IssueID

	// 查询匹配的自动化规则
	var rules []model.AutomationRule

	if err := s.db.Where("project_id = ? AND is_enabled = ?",
		event.ProjectID, true).Order("sequence ASC").Find(&rules).Error; err != nil {
		log.Printf("[Automation] Failed to query project rules: %v", err)
		s.recordExecutionHistory(event, nil, 0, "failed", "Failed to query project rules", startTime)
		return err
	}

	var workspaceRules []model.AutomationRule
	var project model.Project
	if err := s.db.Select("workspace_id").First(&project, event.ProjectID).Error; err == nil {
		var candidateRules []model.AutomationRule
		if err := s.db.Where("workspace_id = ? AND project_id = 0 AND is_enabled = ?",
			project.WorkspaceID, true).Order("sequence ASC").Find(&candidateRules).Error; err != nil {
			log.Printf("[Automation] Failed to query workspace rules: %v", err)
		}

		// Load project overrides for workspace rules
		var ruleIDs []uint64
		for _, wr := range candidateRules {
			ruleIDs = append(ruleIDs, wr.ID)
		}
		var overrides []model.AutomationRuleOverride
		s.db.Where("rule_id IN ? AND project_id = ?", ruleIDs, event.ProjectID).Find(&overrides)
		overrideMap := make(map[uint64]bool)
		for _, ov := range overrides {
			if ov.IsEnabled != nil {
				overrideMap[ov.RuleID] = *ov.IsEnabled
			}
		}

		// Filter by project scope and apply overrides
		for _, wr := range candidateRules {
			if !s.isRuleInProjectScope(wr.Scope, event.ProjectID) {
				continue
			}
			// Check if project has overridden is_enabled
			if ovEnabled, exists := overrideMap[wr.ID]; exists {
				wr.IsEnabled = ovEnabled
			}
			workspaceRules = append(workspaceRules, wr)
		}
	}

	rules = append(rules, workspaceRules...)

	// 过滤匹配的触发器类型（支持 JSON 格式和纯字符串格式）
	var matchedRules []model.AutomationRule
	for _, rule := range rules {
		triggerType := rule.TriggerType
		parsedType := triggerType
		if strings.HasPrefix(triggerType, "{") {
			var triggerObj map[string]interface{}
			if err := json.Unmarshal([]byte(triggerType), &triggerObj); err == nil {
				if t, ok := triggerObj["type"].(string); ok {
					parsedType = t
				}
			}
		}
		log.Printf("[Automation] Rule %d trigger_type: '%s', parsed type: '%s', event type: '%s', enabled: %v",
			rule.ID, rule.TriggerType, parsedType, event.Type, rule.IsEnabled)
		if parsedType == event.Type {
			matchedRules = append(matchedRules, rule)
		}
	}
	rules = matchedRules
	log.Printf("[Automation] Found %d matching rules for event %s", len(rules), event.Type)

	var allResults []string

	for _, rule := range rules {
		// 解析条件
		var conditions []Condition
		if rule.Conditions != "" && rule.Conditions != "[]" {
			if err := json.Unmarshal([]byte(rule.Conditions), &conditions); err != nil {
				log.Printf("[Automation] Failed to parse conditions for rule %d: %v", rule.ID, err)
				s.recordExecutionHistory(event, nil, rule.ID, "failed", "Failed to parse conditions", startTime)
				continue
			}
		}

		// 评估条件
		if !s.ruleEngine.Evaluate(conditions, event.Context) {
			log.Printf("[Automation] Rule %d conditions not met for event %s, skipping", rule.ID, event.Type)
			s.recordExecutionHistory(event, nil, rule.ID, "skipped", "Conditions not met", startTime)
			continue // 条件不匹配，跳过
		}

		// 解析动作
		var actions []Action
		if err := json.Unmarshal([]byte(rule.Actions), &actions); err != nil {
			log.Printf("[Automation] Failed to parse actions for rule %d: %v", rule.ID, err)
			continue
		}

		// 执行动作
		results, err := s.actionExecutor.Execute(actions, event.Context)
		if err != nil {
			log.Printf("[Automation] Failed to execute actions for rule %d: %v", rule.ID, err)
		}

		allResults = append(allResults, results...)

		// 更新规则执行计数
		s.db.Model(&rule).Update("execution_count", gorm.Expr("execution_count + 1"))

		// 记录每条规则的执行历史
		ruleStatus := "success"
		ruleError := ""
		if err != nil {
			ruleStatus = "failed"
			ruleError = err.Error()
		}
		s.recordExecutionHistory(event, results, rule.ID, ruleStatus, ruleError, startTime)
	}

	return nil
}

func (s *AutomationService) recordExecutionHistory(event Event, results []string, ruleID uint64, status string, errorMsg string, startTime time.Time) {
	contextJSON, _ := json.Marshal(event.Context)
	actionsJSON, _ := json.Marshal(results)
	duration := time.Since(startTime).Milliseconds()

	execution := model.AutomationExecution{
		RuleID:       ruleID,
		IssueID:      event.IssueID,
		TriggerType:  event.Type,
		ContextJSON:  string(contextJSON),
		ActionsTaken: string(actionsJSON),
		Status:       status,
		Error:        errorMsg,
		Duration:     duration,
		ExecutedAt:   time.Now(),
	}

	if err := s.db.Create(&execution).Error; err != nil {
		log.Printf("[Automation] Failed to record execution history: %v", err)
	}
}

// PublishEvent 发布事件到事件总线（供 IssueService 调用）
// SetAgentService sets the agent client on the action executor (breaks circular dependency).
func (s *AutomationService) SetAgentService(agentClient *client.AgentClient) {
	if exec, ok := s.actionExecutor.(*DefaultActionExecutor); ok {
		exec.agentClient = agentClient
	}
}

func (s *AutomationService) PublishEvent(event Event) error {
	return s.eventBus.Publish(event)
}

// GetExecutionHistory 获取自动化执行历史（新增 API）
func (s *AutomationService) GetExecutionHistory(issueID uint64, limit int, offset int) ([]model.AutomationExecution, int64, error) {
	var executions []model.AutomationExecution
	var total int64

	query := s.db.Model(&model.AutomationExecution{}).Where("issue_id = ?", issueID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Failed to count execution history")
	}

	if err := query.Order("executed_at DESC").Offset(offset).Limit(limit).Find(&executions).Error; err != nil {
		return nil, 0, common.Internal("Failed to get execution history")
	}
	return executions, total, nil
}

// GetRuleExecutionHistory 获取指定规则的执行历史
func (s *AutomationService) GetRuleExecutionHistory(ruleID uint64, limit int, offset int, startTime *time.Time, endTime *time.Time) ([]model.AutomationExecution, int64, error) {
	var executions []model.AutomationExecution
	var total int64

	query := s.db.Model(&model.AutomationExecution{}).Where("rule_id = ?", ruleID)

	if startTime != nil {
		query = query.Where("executed_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("executed_at <= ?", *endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Failed to count execution history")
	}

	if err := query.Order("executed_at DESC").Offset(offset).Limit(limit).Find(&executions).Error; err != nil {
		return nil, 0, common.Internal("Failed to get rule execution history")
	}
	return executions, total, nil
}

// GetProjectExecutionHistory 获取项目的自动化执行历史
func (s *AutomationService) GetProjectExecutionHistory(projectID uint64, limit int, offset int, startTime *time.Time, endTime *time.Time) ([]model.AutomationExecution, int64, error) {
	var executions []model.AutomationExecution
	var total int64

	query := s.db.Model(&model.AutomationExecution{}).
		Joins("JOIN automation_rules ON automation_rules.id = automation_executions.rule_id").
		Where("automation_rules.project_id = ?", projectID)

	if startTime != nil {
		query = query.Where("automation_executions.executed_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("automation_executions.executed_at <= ?", *endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.Internal("Failed to count project execution history")
	}

	if err := query.Order("automation_executions.executed_at DESC").Offset(offset).Limit(limit).Find(&executions).Error; err != nil {
		return nil, 0, common.Internal("Failed to get project execution history")
	}
	return executions, total, nil
}

// Helper functions

func toUint64(v interface{}) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		return uint64(val), true
	case int:
		return uint64(val), true
	case int64:
		return uint64(val), true
	case uint64:
		return val, true
	case string:
		var n uint64
		if _, err := fmt.Sscanf(val, "%d", &n); err == nil {
			return n, true
		}
	}
	return 0, false
}

// ======== Request/Response types ========

type AutomationCreateRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	TriggerType    string `json:"trigger_type" binding:"required"`
	Conditions     string `json:"conditions"`
	Actions        string `json:"actions" binding:"required"`
	IsEnabled      *bool  `json:"is_enabled"`
	Sequence       int    `json:"sequence"`
	Scope          string `json:"scope"`           // workspace rule project scope: "all" or "[1,2,3]"
	ScheduleConfig string `json:"schedule_config"` // scheduled trigger config JSON
}

type AutomationUpdateRequest struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	TriggerType    *string `json:"trigger_type"`
	Conditions     *string `json:"conditions"`
	Actions        *string `json:"actions"`
	IsEnabled      *bool   `json:"is_enabled"`
	Sequence       *int    `json:"sequence"`
	Scope          *string `json:"scope"`
	ScheduleConfig *string `json:"schedule_config"`
}

type AutomationResponse struct {
	ID             uint64  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	ProjectID      uint64  `json:"project_id"`
	WorkspaceID    uint64  `json:"workspace_id"`
	TriggerType    string  `json:"trigger_type"`
	Conditions     string  `json:"conditions"`
	Actions        string  `json:"actions"`
	IsEnabled      bool    `json:"is_enabled"`
	IsInherited    bool    `json:"is_inherited"`
	Sequence       int     `json:"sequence"`
	ExecutionCount int     `json:"execution_count"`
	Scope          string  `json:"scope,omitempty"`
	ScheduleConfig string  `json:"schedule_config,omitempty"`
	LastTriggeredAt *string `json:"last_triggered_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ======== CRUD 方法（保留原有 API 兼容性）========

func (s *AutomationService) List(projectID uint64) ([]AutomationResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").First(&project, projectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	var projectRules []model.AutomationRule
	if err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&projectRules).Error; err != nil {
		return nil, common.Internal("Failed to list project automation rules")
	}

	var candidateWorkspaceRules []model.AutomationRule
	if err := s.db.Where("workspace_id = ? AND project_id = 0", project.WorkspaceID).Order("created_at DESC").Find(&candidateWorkspaceRules).Error; err != nil {
		return nil, common.Internal("Failed to list workspace automation rules")
	}

	// Filter workspace rules by project scope
	var workspaceRules []model.AutomationRule
	for _, wr := range candidateWorkspaceRules {
		if s.isRuleInProjectScope(wr.Scope, projectID) {
			workspaceRules = append(workspaceRules, wr)
		}
	}

	mergedRules := append(projectRules, workspaceRules...)

	// Load project-level overrides for inherited workspace rules
	overrideMap := make(map[uint64]*model.AutomationRuleOverride)
	if len(workspaceRules) > 0 {
		var ruleIDs []uint64
		for _, wr := range workspaceRules {
			ruleIDs = append(ruleIDs, wr.ID)
		}
		var overrides []model.AutomationRuleOverride
		s.db.Where("rule_id IN ? AND project_id = ?", ruleIDs, projectID).Find(&overrides)
		for i := range overrides {
			overrideMap[overrides[i].RuleID] = &overrides[i]
		}
	}

	res := make([]AutomationResponse, len(mergedRules))
	for i, r := range mergedRules {
		isInherited := r.ProjectID == 0
		// Apply override: if project has overridden is_enabled for this inherited rule
		if isInherited {
			if ov, ok := overrideMap[r.ID]; ok && ov.IsEnabled != nil {
				r.IsEnabled = *ov.IsEnabled
			}
		}
		res[i] = s.toResponseWithInherited(&r, isInherited)
	}
	if res == nil {
		res = []AutomationResponse{}
	}
	return res, nil
}

func (s *AutomationService) Get(id uint64) (*AutomationResponse, error) {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Automation rule not found")
		}
		return nil, common.Internal("Failed to get automation rule")
	}
	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Create(projectID uint64, req *AutomationCreateRequest) (*AutomationResponse, error) {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}

	if err := validateJSON(req.Conditions); err != nil {
		return nil, common.BadRequest("invalid conditions JSON: " + err.Error())
	}
	if err := validateJSON(req.Actions); err != nil {
		return nil, common.BadRequest("invalid actions JSON: " + err.Error())
	}

	rule := model.AutomationRule{
		Name:           req.Name,
		Description:    req.Description,
		ProjectID:      projectID,
		TriggerType:    req.TriggerType,
		Conditions:     req.Conditions,
		Actions:        req.Actions,
		IsEnabled:      enabled,
		Sequence:       req.Sequence,
		ScheduleConfig: req.ScheduleConfig,
	}

	if err := s.db.Create(&rule).Error; err != nil {
		return nil, common.Internal("Failed to create automation rule")
	}

	// If scheduled, set initial last_triggered_at
	if req.TriggerType == "scheduled" {
		now := time.Now()
		s.db.Model(&rule).Update("last_triggered_at", now)
		rule.LastTriggeredAt = &now
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Update(id uint64, projectID uint64, req *AutomationUpdateRequest) (*AutomationResponse, error) {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Automation rule not found")
		}
		return nil, common.Internal("Failed to get automation rule")
	}

	// If this is an inherited workspace rule and request is from project context,
	// create/update a per-project override instead of modifying the rule directly.
	if rule.ProjectID == 0 && projectID > 0 {
		// Only is_enabled override is supported from project context
		if req.IsEnabled != nil {
			var ov model.AutomationRuleOverride
			err := s.db.Where("rule_id = ? AND project_id = ?", rule.ID, projectID).First(&ov).Error
			if err == gorm.ErrRecordNotFound {
				ov = model.AutomationRuleOverride{RuleID: rule.ID, ProjectID: projectID}
			} else if err != nil {
				return nil, common.Internal("Failed to check override")
			}
			ov.IsEnabled = req.IsEnabled
			if err := s.db.Save(&ov).Error; err != nil {
				return nil, common.Internal("Failed to save override")
			}
			// Apply the override to the response
			rule.IsEnabled = *req.IsEnabled
			r := s.toResponseWithInherited(&rule, true)
			return &r, nil
		}
		// Other fields cannot be overridden from project context
		return nil, common.BadRequest("Cannot modify inherited workspace rule from project context")
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.TriggerType != nil {
		rule.TriggerType = *req.TriggerType
	}
	if req.Conditions != nil {
		if err := validateJSON(*req.Conditions); err != nil {
			return nil, common.BadRequest("invalid conditions JSON: " + err.Error())
		}
		rule.Conditions = *req.Conditions
	}
	if req.Actions != nil {
		if err := validateJSON(*req.Actions); err != nil {
			return nil, common.BadRequest("invalid actions JSON: " + err.Error())
		}
		rule.Actions = *req.Actions
	}
	if req.IsEnabled != nil {
		rule.IsEnabled = *req.IsEnabled
	}
	if req.Sequence != nil {
		rule.Sequence = *req.Sequence
	}
	if req.Scope != nil {
		rule.Scope = *req.Scope
	}
	if req.ScheduleConfig != nil {
		rule.ScheduleConfig = *req.ScheduleConfig
	}

	if err := s.db.Save(&rule).Error; err != nil {
		return nil, common.Internal("Failed to update automation rule")
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Delete(id, callerID uint64) error {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Automation rule not found")
		}
		return common.Internal("Failed to get automation rule")
	}
	if rule.ProjectID > 0 {
		if err := s.checkProjectAdmin(rule.ProjectID, callerID); err != nil {
			return err
		}
	} else if rule.WorkspaceID > 0 {
		if err := s.checkWorkspaceAdmin(rule.WorkspaceID, callerID); err != nil {
			return err
		}
	}
	if err := s.db.Delete(&rule).Error; err != nil {
		return common.Internal("Failed to delete automation rule")
	}
	return nil
}

// ExecuteTrigger 手动触发规则执行（用于测试）
func (s *AutomationService) ExecuteTrigger(projectID uint64, triggerType string, issueID uint64, eventCtx map[string]interface{}) []string {
	if eventCtx == nil {
		eventCtx = make(map[string]interface{})
	}
	// Ensure issue_id is always in context for action handlers
	eventCtx["issue_id"] = issueID

	event := Event{
		Type:      triggerType,
		ProjectID: projectID,
		IssueID:   issueID,
		Context:   eventCtx,
		Timestamp: time.Now(),
	}

	// 同步执行用于测试
	ctx := context.Background()
	_ = s.handleAutomationEvent(ctx, event)

	return []string{"Executed"}
}

// ======== 工作区级自动化规则 CRUD ========

func (s *AutomationService) ListWorkspace(workspaceID uint64) ([]AutomationResponse, error) {
	var rules []model.AutomationRule
	if err := s.db.Where("workspace_id = ?", workspaceID).Order("sequence ASC").Find(&rules).Error; err != nil {
		return nil, common.Internal("Failed to list workspace automation rules")
	}
	res := make([]AutomationResponse, len(rules))
	for i, r := range rules {
		res[i] = s.toResponse(&r)
	}
	if res == nil {
		res = []AutomationResponse{}
	}
	return res, nil
}

func (s *AutomationService) CreateWorkspace(workspaceID uint64, req *AutomationCreateRequest) (*AutomationResponse, error) {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}

	if err := validateJSON(req.Conditions); err != nil {
		return nil, common.BadRequest("invalid conditions JSON: " + err.Error())
	}
	if err := validateJSON(req.Actions); err != nil {
		return nil, common.BadRequest("invalid actions JSON: " + err.Error())
	}

	scope := req.Scope
	if scope == "" {
		scope = "all"
	}

	rule := model.AutomationRule{
		Name:           req.Name,
		Description:    req.Description,
		WorkspaceID:    workspaceID,
		TriggerType:    req.TriggerType,
		Conditions:     req.Conditions,
		Actions:        req.Actions,
		IsEnabled:      enabled,
		Sequence:       req.Sequence,
		Scope:          scope,
		ScheduleConfig: req.ScheduleConfig,
	}

	if err := s.db.Create(&rule).Error; err != nil {
		return nil, common.Internal("Failed to create workspace automation rule")
	}

	// If it's a scheduled rule, start tracking from now
	if req.TriggerType == "scheduled" {
		now := time.Now()
		s.db.Model(&rule).Update("last_triggered_at", now)
		rule.LastTriggeredAt = &now
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) UpdateWorkspace(id uint64, req *AutomationUpdateRequest) (*AutomationResponse, error) {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.NotFound("Automation rule not found")
		}
		return nil, common.Internal("Failed to get automation rule")
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.TriggerType != nil {
		rule.TriggerType = *req.TriggerType
	}
	if req.Conditions != nil {
		if err := validateJSON(*req.Conditions); err != nil {
			return nil, common.BadRequest("invalid conditions JSON: " + err.Error())
		}
		rule.Conditions = *req.Conditions
	}
	if req.Actions != nil {
		if err := validateJSON(*req.Actions); err != nil {
			return nil, common.BadRequest("invalid actions JSON: " + err.Error())
		}
		rule.Actions = *req.Actions
	}
	if req.IsEnabled != nil {
		rule.IsEnabled = *req.IsEnabled
	}
	if req.Sequence != nil {
		rule.Sequence = *req.Sequence
	}
	if req.Scope != nil {
		rule.Scope = *req.Scope
	}
	if req.ScheduleConfig != nil {
		rule.ScheduleConfig = *req.ScheduleConfig
	}

	if err := s.db.Save(&rule).Error; err != nil {
		return nil, common.Internal("Failed to update automation rule")
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) DeleteWorkspace(id, callerID uint64) error {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Automation rule not found")
		}
		return common.Internal("Failed to get automation rule")
	}
	if err := s.checkWorkspaceAdmin(rule.WorkspaceID, callerID); err != nil {
		return err
	}
	if err := s.db.Delete(&rule).Error; err != nil {
		return common.Internal("Failed to delete automation rule")
	}
	return nil
}

// Helpers

func (s *AutomationService) toResponse(rule *model.AutomationRule) AutomationResponse {
	return s.toResponseWithInherited(rule, false)
}

func (s *AutomationService) toResponseWithInherited(rule *model.AutomationRule, isInherited bool) AutomationResponse {
	resp := AutomationResponse{
		ID:             rule.ID,
		Name:           rule.Name,
		Description:    rule.Description,
		ProjectID:      rule.ProjectID,
		WorkspaceID:    rule.WorkspaceID,
		TriggerType:    rule.TriggerType,
		Conditions:     rule.Conditions,
		Actions:        rule.Actions,
		IsEnabled:      rule.IsEnabled,
		IsInherited:    isInherited,
		Sequence:       rule.Sequence,
		ExecutionCount: rule.ExecutionCount,
		Scope:          rule.Scope,
		ScheduleConfig: rule.ScheduleConfig,
		CreatedAt:      rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      rule.UpdatedAt.Format(time.RFC3339),
	}
	if rule.LastTriggeredAt != nil {
		s := rule.LastTriggeredAt.Format(time.RFC3339)
		resp.LastTriggeredAt = &s
	}
	return resp
}

func validateJSON(s string) error {
	if s == "" || s == "[]" {
		return nil
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js)
}

// isRuleInProjectScope checks whether a workspace-level rule's scope includes a given project.
// scope values:
//   - "all" or "" → applies to all projects
//   - JSON array like "[1,2,3]" → only those project IDs
func (s *AutomationService) isRuleInProjectScope(scope string, projectID uint64) bool {
	if scope == "" || scope == "all" {
		return true
	}
	var projectIDs []uint64
	if err := json.Unmarshal([]byte(scope), &projectIDs); err != nil {
		// If not valid JSON array, fall back to "all"
		return true
	}
	for _, pid := range projectIDs {
		if pid == projectID {
			return true
		}
	}
	return false
}

// ======== Scheduled Trigger Scheduler ========

// StartScheduler starts a background goroutine that checks for scheduled automation
// rules every minute and triggers them when due.
func (s *AutomationService) StartScheduler(ctx context.Context) {
	log.Println("[Automation] Starting scheduled trigger scheduler (interval: 1 min)")
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Automation] Scheduler stopped")
				return
			case <-ticker.C:
				s.processScheduledTriggers()
			}
		}
	}()
}

// processScheduledTriggers finds enabled scheduled rules and triggers those
// whose schedule matches the current time.
func (s *AutomationService) processScheduledTriggers() {
	var rules []model.AutomationRule
	if err := s.db.Where("trigger_type = ? AND is_enabled = ?", "scheduled", true).Find(&rules).Error; err != nil {
		log.Printf("[Automation] Failed to query scheduled rules: %v", err)
		return
	}

	now := time.Now()
	for _, rule := range rules {
		if !s.isScheduleDue(&rule, now) {
			continue
		}

		// Determine target projects
		var targetProjects []uint64
		if rule.ProjectID > 0 {
			targetProjects = append(targetProjects, rule.ProjectID)
		} else {
			// Workspace-level rule: resolve projects based on scope
			targetProjects = s.resolveWorkspaceRuleProjects(rule.WorkspaceID, rule.Scope)
		}

		for _, projectID := range targetProjects {
			event := Event{
				Type:      "scheduled",
				ProjectID: projectID,
				IssueID:   0, // scheduled events may not have a specific issue
				Context:   map[string]interface{}{"rule_id": rule.ID, "workspace_id": rule.WorkspaceID},
				Timestamp: now,
			}
			_ = s.PublishEvent(event)
		}

		// Update last_triggered_at
		s.db.Model(&rule).Update("last_triggered_at", now)
	}
}

// isScheduleDue checks if a scheduled rule should fire now.
// schedule_config JSON formats supported:
//   - {"frequency":"hourly", "minute": 0} — at minute 0 of every hour
//   - {"frequency":"daily", "time":"09:00"} — every day at 9:00
//   - {"frequency":"weekly", "time":"09:00", "days":["mon","wed","fri"]} — Mon/Wed/Fri at 9:00
//   - {"frequency":"monthly", "day":1, "time":"09:00"} — 1st of each month at 9:00
//   - {"frequency":"cron", "cron":"0 9 * * 1"} — standard cron expression
func (s *AutomationService) isScheduleDue(rule *model.AutomationRule, now time.Time) bool {
	if rule.ScheduleConfig == "" {
		return false
	}

	var config struct {
		Frequency string   `json:"frequency"`
		Time      string   `json:"time"`
		Days      []string `json:"days"`
		Day       int      `json:"day"`
		Cron      string   `json:"cron"`
		Minute    int      `json:"minute"`
	}

	if err := json.Unmarshal([]byte(rule.ScheduleConfig), &config); err != nil {
		log.Printf("[Automation] Invalid schedule_config for rule %d: %v", rule.ID, err)
		return false
	}

	// Guard: don't fire more than once per window
	if rule.LastTriggeredAt != nil {
		switch config.Frequency {
		case "hourly":
			if now.Sub(*rule.LastTriggeredAt) < 55*time.Minute {
				return false
			}
		case "daily", "weekly":
			if now.Sub(*rule.LastTriggeredAt) < 23*time.Hour {
				return false
			}
		case "monthly":
			if now.Sub(*rule.LastTriggeredAt) < 27*24*time.Hour {
				return false
			}
		default:
			if now.Sub(*rule.LastTriggeredAt) < 5*time.Minute {
				return false
			}
		}
	}

	switch config.Frequency {
	case "hourly":
		return now.Minute() == config.Minute

	case "daily":
		if config.Time == "" {
			return false
		}
		return now.Format("15:04") == config.Time

	case "weekly":
		if config.Time == "" || len(config.Days) == 0 {
			return false
		}
		if now.Format("15:04") != config.Time {
			return false
		}
		currentDay := strings.ToLower(now.Format("Mon"))
		for _, d := range config.Days {
			if strings.ToLower(d) == currentDay {
				return true
			}
		}
		return false

	case "monthly":
		if config.Time == "" {
			return false
		}
		if config.Day > 0 && now.Day() != config.Day {
			return false
		}
		return now.Format("15:04") == config.Time

	case "cron":
		if config.Cron == "" {
			return false
		}
		return matchCronExpression(config.Cron, now)

	default:
		return false
	}
}

// matchCronExpression does a basic 5-field cron match.
// Format: minute hour day-of-month month day-of-week
func matchCronExpression(cron string, now time.Time) bool {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return false
	}

	current := []int{now.Minute(), now.Hour(), now.Day(), int(now.Month()), int(now.Weekday())}
	for i, part := range parts {
		if part == "*" {
			continue
		}
		if val, err := fmt.Sscanf(part, "%d", new(int)); err == nil && val == 1 {
			// simple single value match handled below
		}
		// For simplicity, only support single-value or wildcard matches for now
		if part != "*" {
			var val int
			if _, err := fmt.Sscanf(part, "%d", &val); err != nil {
				return false
			}
			if current[i] != val {
				return false
			}
		}
	}
	return true
}

// resolveWorkspaceRuleProjects returns the project IDs that a workspace rule
// should apply to, based on its scope.
func (s *AutomationService) resolveWorkspaceRuleProjects(workspaceID uint64, scope string) []uint64 {
	if scope == "" || scope == "all" {
		var projects []model.Project
		s.db.Where("workspace_id = ?", workspaceID).Select("id").Find(&projects)
		var ids []uint64
		for _, p := range projects {
			ids = append(ids, p.ID)
		}
		return ids
	}

	var projectIDs []uint64
	if err := json.Unmarshal([]byte(scope), &projectIDs); err != nil {
		return nil
	}
	return projectIDs
}
