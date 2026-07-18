package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/reqmango/backend/internal/common"
	"github.com/reqmango/backend/internal/model"
	"gorm.io/gorm"
)

// ======== 新架构：事件总线 ========

// Event 表示自动化系统中的事件
type Event struct {
	Type        string                 `json:"type"`         // issue_created, state_changed, etc.
	ProjectID   uint64                 `json:"project_id"`
	IssueID     uint64                 `json:"issue_id"`
	Context     map[string]interface{} `json:"context"`      // old_state, new_state, state_group, priority, etc.
	Timestamp   time.Time              `json:"timestamp"`
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

	log.Printf("[EventBus] Publishing event: type=%s, project=%d, issue=%d", event.Type, event.ProjectID, event.IssueID)
	
	handlers, exists := bus.handlers[event.Type]
	if !exists {
		log.Printf("[EventBus] No handlers for event type: %s", event.Type)
		return nil // 没有订阅者，静默跳过
	}
	
	log.Printf("[EventBus] Found %d handlers for event type: %s", len(handlers), event.Type)

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
	Field       string      `json:"field"`
	Operator    string      `json:"operator"` // equals, not_equals, contains, in, gt, lt, is_empty, is_not_empty, matches_regex
	Value       interface{} `json:"value"`
	Logic       string      `json:"logic,omitempty"`       // AND, OR, NOT (用于组合多个条件)
	Conditions  []Condition `json:"conditions,omitempty"` // 嵌套条件组
	IsNegated   bool        `json:"is_negated,omitempty"` // 是否取反（NOT）
}

// DefaultConditionEvaluator 默认条件评估器实现（支持复杂逻辑组合）
type DefaultConditionEvaluator struct{}

func NewDefaultConditionEvaluator() *DefaultConditionEvaluator {
	return &DefaultConditionEvaluator{}
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
		return false // 字段不存在，不匹配
	}

	switch cond.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", cond.Value)
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
	handlers map[string]ActionHandler
	db       *gorm.DB
	agentSvc *AgentService // 用于 dispatch_agent 动作
}

func NewDefaultActionExecutor(db *gorm.DB, agentSvc *AgentService) *DefaultActionExecutor {
	executor := &DefaultActionExecutor{
		handlers: make(map[string]ActionHandler),
		db:       db,
		agentSvc: agentSvc,
	}

	// 注册内置动作处理器
	executor.registerBuiltinActions()

	return executor
}

func (e *DefaultActionExecutor) registerBuiltinActions() {
	e.RegisterAction("set_field", e.handleSetField)
	e.RegisterAction("add_label", e.handleAddLabel)
	e.RegisterAction("remove_label", e.handleRemoveLabel)
	e.RegisterAction("add_comment", e.handleAddComment)
	e.RegisterAction("assign_to", e.handleAssignTo)
	e.RegisterAction("unassign", e.handleUnassign)
	e.RegisterAction("change_state", e.handleChangeState)
	e.RegisterAction("set_priority", e.handleSetPriority)
	e.RegisterAction("archive", e.handleArchive)
	e.RegisterAction("close", handleClose)
	e.RegisterAction("dispatch_agent", e.handleDispatchAgent)
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
	
	// 验证字段是否允许修改
	allowedFields := map[string]bool{
		"priority": true,
		"state_id": true,
	}
	
	if !allowedFields[field] {
		return fmt.Errorf("field %s is not allowed to be set via automation", field)
	}
	
	return db.Model(&model.Issue{}).Where("id = ?", issueID).Update(field, value).Error
}

func (e *DefaultActionExecutor) handleAddLabel(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}
	
	labelID, ok := toUint64(action.Value)
	if !ok {
		return fmt.Errorf("invalid label_id: %v", action.Value)
	}
	
	// 检查是否已存在
	var count int64
	db.Model(&model.IssueLabel{}).Where("issue_id = ? AND label_id = ?", issueID, labelID).Count(&count)
	if count > 0 {
		return nil // 已存在，跳过
	}
	
	return db.Create(&model.IssueLabel{IssueID: issueID, LabelID: labelID}).Error
}

func (e *DefaultActionExecutor) handleRemoveLabel(action Action, context map[string]interface{}, db *gorm.DB) error {
	issueID, ok := context["issue_id"].(uint64)
	if !ok {
		return fmt.Errorf("missing issue_id in context")
	}
	
	labelID, ok := toUint64(action.Value)
	if !ok {
		return fmt.Errorf("invalid label_id: %v", action.Value)
	}
	
	return db.Where("issue_id = ? AND label_id = ?", issueID, labelID).Delete(&model.IssueLabel{}).Error
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
	
	stateID, ok := toUint64(action.Value)
	if !ok {
		return fmt.Errorf("invalid state_id: %v", action.Value)
	}
	
	// 验证状态转换是否合法（可选：检查工作流规则）
	var newState model.State
	if err := db.First(&newState, stateID).Error; err != nil {
		return fmt.Errorf("state %d not found", stateID)
	}
	
	updateData := map[string]interface{}{"state_id": stateID}
	
	// 如果新状态是完成状态，设置 completed_at
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
		"urgent":   true,
		"high":     true,
		"medium":   true,
		"low":      true,
		"none":     true,
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
	if e.agentSvc == nil {
		return fmt.Errorf("agent service not available")
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

	// Build dispatch context
	var issueIDPtr *uint64
	if issueID > 0 {
		issueIDPtr = &issueID
	}
	var projectIDPtr *uint64
	if projectID > 0 {
		projectIDPtr = &projectID
	}

	dispatchCtx := &DispatchContext{
		IssueID:     issueIDPtr,
		ProjectID:   projectIDPtr,
		TriggeredBy: "automation",
	}

	// Use system user (user ID 1) as the actor for automation-triggered dispatches
	_, err := e.agentSvc.DispatchAgent(agentID, 1, task, dispatchCtx)
	if err != nil {
		log.Printf("[Automation] Agent dispatch failed: agent=%d issue=%d err=%v", agentID, issueID, err)
		return err
	}

	log.Printf("[Automation] Agent dispatched: agent=%d issue=%d task=%s", agentID, issueID, task)
	return nil
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

// ======== 重构后的 AutomationService ========

// AutomationService 重构后的自动化服务（参考主流架构）
type AutomationService struct {
	db              *gorm.DB
	eventBus        EventBus
	ruleEngine      ConditionEvaluator
	actionExecutor  ActionExecutor
	executionHistory sync.Map // 保留用于循环检测
	maxExecutions    int
	execWindow       time.Duration
}

func NewAutomationService(db *gorm.DB) *AutomationService {
	eventBus := NewInMemoryEventBus()
	ruleEngine := NewDefaultConditionEvaluator()
	actionExecutor := NewDefaultActionExecutor(db, nil) // agentSvc set via SetAgentService after construction
	
	service := &AutomationService{
		db:              db,
		eventBus:        eventBus,
		ruleEngine:      ruleEngine,
		actionExecutor:  actionExecutor,
		maxExecutions:   10,
		execWindow:      5 * time.Minute,
	}
	
	// 注册事件处理器
	service.registerEventHandlers()
	
	return service
}

func (s *AutomationService) registerEventHandlers() {
	// 订阅所有触发类型的事件
	triggerTypes := []string{
		"issue_created",
		"issue_updated",
		"state_changed",
		"assignee_changed",
		"comment_added",
	}
	
	for _, triggerType := range triggerTypes {
		s.eventBus.Subscribe(triggerType, s.handleAutomationEvent)
	}
}

func (s *AutomationService) handleAutomationEvent(ctx context.Context, event Event) error {
	startTime := time.Now()
	
	log.Printf("[Automation] Received event: type=%s, project=%d, issue=%d", event.Type, event.ProjectID, event.IssueID)
	
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
					s.recordExecutionHistory(event, nil, "skipped", "Cycle detected", startTime)
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
	
	if err := s.db.Where("project_id = ? AND trigger_type = ? AND is_enabled = ?", 
		event.ProjectID, event.Type, true).Order("sequence ASC").Find(&rules).Error; err != nil {
		log.Printf("[Automation] Failed to query project rules: %v", err)
		return err
	}

	var workspaceRules []model.AutomationRule
	var project model.Project
	if err := s.db.Select("workspace_id").First(&project, event.ProjectID).Error; err == nil {
		if err := s.db.Where("workspace_id = ? AND project_id = 0 AND trigger_type = ? AND is_enabled = ?",
			project.WorkspaceID, event.Type, true).Order("sequence ASC").Find(&workspaceRules).Error; err != nil {
			log.Printf("[Automation] Failed to query workspace rules: %v", err)
		}
	}
	
	rules = append(rules, workspaceRules...)
	
	var allResults []string
	var hasError bool
	
	for _, rule := range rules {
		// 解析条件
		var conditions []Condition
		if rule.Conditions != "" && rule.Conditions != "[]" {
			if err := json.Unmarshal([]byte(rule.Conditions), &conditions); err != nil {
				log.Printf("[Automation] Failed to parse conditions for rule %d: %v", rule.ID, err)
				continue
			}
		}
		
		// 评估条件
		if !s.ruleEngine.Evaluate(conditions, event.Context) {
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
			hasError = true
		}
		
		allResults = append(allResults, results...)
		
		// 更新规则执行计数
		s.db.Model(&rule).Update("execution_count", gorm.Expr("execution_count + 1"))
	}
	
	// 记录执行历史
	status := "success"
	errorMsg := ""
	if hasError {
		status = "failed"
		errorMsg = "Some actions failed"
	}
	
	s.recordExecutionHistory(event, allResults, status, errorMsg, startTime)
	
	return nil
}

func (s *AutomationService) recordExecutionHistory(event Event, results []string, status string, errorMsg string, startTime time.Time) {
	contextJSON, _ := json.Marshal(event.Context)
	actionsJSON, _ := json.Marshal(results)
	duration := time.Since(startTime).Milliseconds()
	
	execution := model.AutomationExecution{
		RuleID:       0, // TODO: 跟踪具体规则 ID
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
// SetAgentService sets the agent service on the action executor (breaks circular dependency).
func (s *AutomationService) SetAgentService(agentSvc *AgentService) {
	if exec, ok := s.actionExecutor.(*DefaultActionExecutor); ok {
		exec.agentSvc = agentSvc
	}
}

func (s *AutomationService) PublishEvent(event Event) error {
	return s.eventBus.Publish(event)
}

// GetExecutionHistory 获取自动化执行历史（新增 API）
func (s *AutomationService) GetExecutionHistory(issueID uint64, limit int) ([]model.AutomationExecution, error) {
	var executions []model.AutomationExecution
	if err := s.db.Where("issue_id = ?", issueID).Order("executed_at DESC").Limit(limit).Find(&executions).Error; err != nil {
		return nil, common.Internal("Failed to get execution history")
	}
	return executions, nil
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
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	TriggerType string  `json:"trigger_type" binding:"required"`
	Conditions  string  `json:"conditions"`
	Actions     string  `json:"actions" binding:"required"`
	IsEnabled   *bool   `json:"is_enabled"`
	Sequence    int     `json:"sequence"`
}

type AutomationUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TriggerType *string `json:"trigger_type"`
	Conditions  *string `json:"conditions"`
	Actions     *string `json:"actions"`
	IsEnabled   *bool   `json:"is_enabled"`
	Sequence    *int    `json:"sequence"`
}

type AutomationResponse struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ProjectID      uint64 `json:"project_id"`
	WorkspaceID    uint64 `json:"workspace_id"`
	TriggerType    string `json:"trigger_type"`
	Conditions     string `json:"conditions"`
	Actions        string `json:"actions"`
	IsEnabled      bool   `json:"is_enabled"`
	IsInherited    bool   `json:"is_inherited"`
	Sequence       int    `json:"sequence"`
	ExecutionCount int    `json:"execution_count"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ======== CRUD 方法（保留原有 API 兼容性）========

func (s *AutomationService) List(projectID uint64) ([]AutomationResponse, error) {
	var project model.Project
	if err := s.db.Select("workspace_id").First(&project, projectID).Error; err != nil {
		return nil, common.NotFound("Project not found")
	}

	var projectRules []model.AutomationRule
	if err := s.db.Where("project_id = ?", projectID).Order("sequence ASC").Find(&projectRules).Error; err != nil {
		return nil, common.Internal("Failed to list project automation rules")
	}

	var workspaceRules []model.AutomationRule
	if err := s.db.Where("workspace_id = ? AND project_id = 0", project.WorkspaceID).Order("sequence ASC").Find(&workspaceRules).Error; err != nil {
		return nil, common.Internal("Failed to list workspace automation rules")
	}

	mergedRules := append(projectRules, workspaceRules...)

	res := make([]AutomationResponse, len(mergedRules))
	for i, r := range mergedRules {
		isInherited := r.ProjectID == 0
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
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectID,
		TriggerType: req.TriggerType,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
		IsEnabled:   enabled,
		Sequence:    req.Sequence,
	}

	if err := s.db.Create(&rule).Error; err != nil {
		return nil, common.Internal("Failed to create automation rule")
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Update(id uint64, req *AutomationUpdateRequest) (*AutomationResponse, error) {
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

	if err := s.db.Save(&rule).Error; err != nil {
		return nil, common.Internal("Failed to update automation rule")
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) Delete(id uint64) error {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Automation rule not found")
		}
		return common.Internal("Failed to get automation rule")
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

	rule := model.AutomationRule{
		Name:        req.Name,
		Description: req.Description,
		WorkspaceID: workspaceID,
		TriggerType: req.TriggerType,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
		IsEnabled:   enabled,
		Sequence:    req.Sequence,
	}

	if err := s.db.Create(&rule).Error; err != nil {
		return nil, common.Internal("Failed to create workspace automation rule")
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

	if err := s.db.Save(&rule).Error; err != nil {
		return nil, common.Internal("Failed to update automation rule")
	}

	r := s.toResponse(&rule)
	return &r, nil
}

func (s *AutomationService) DeleteWorkspace(id uint64) error {
	var rule model.AutomationRule
	if err := s.db.First(&rule, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.NotFound("Automation rule not found")
		}
		return common.Internal("Failed to get automation rule")
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
	return AutomationResponse{
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
		CreatedAt:      rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      rule.UpdatedAt.Format(time.RFC3339),
	}
}

func validateJSON(s string) error {
	if s == "" || s == "[]" {
		return nil
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js)
}
