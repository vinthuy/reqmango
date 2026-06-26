/**
 * Workflow Types - 工作流和自动化类型定义
 */

// ==================== Enums ====================

export enum TriggerTypeEnum {
  // 工作项相关
  ISSUE_CREATED = 'issue.created',
  ISSUE_UPDATED = 'issue.updated',
  ISSUE_DELETED = 'issue.deleted',
  ISSUE_ASSIGNED = 'issue.assigned',
  STATE_CHANGED = 'issue.state_changed',
  PRIORITY_CHANGED = 'issue.priority_changed',
  DUE_SOON = 'issue.due_soon',
  DUE_DATE_PASSED = 'issue.due_date_passed',
  
  // 周期相关
  CYCLE_STARTED = 'cycle.started',
  CYCLE_ENDED = 'cycle.ended',
  
  // 评论相关
  COMMENT_ADDED = 'comment.added'
}

export const TriggerTypeOptions = [
  { value: 'issue.created', label: '工作项创建时', icon: '➕' },
  { value: 'issue.updated', label: '工作项更新时', icon: '✏️' },
  { value: 'issue.state_changed', label: '状态变更时', icon: '🔄' },
  { value: 'issue.assigned', label: '分配人变更时', icon: '👤' },
  { value: 'issue.priority_changed', label: '优先级变更时', icon: '⚡' },
  { value: 'issue.due_soon', label: '截止日期临近时', icon: '⏰' },
  { value: 'issue.due_date_passed', label: '截止日期过期时', icon: '⚠️' },
  { value: 'cycle.started', label: '周期开始时', icon: '▶️' },
  { value: 'cycle.ended', label: '周期结束时', icon: '⏹️' },
  { value: 'comment.added', label: '添加评论时', icon: '💬' },
]

export const StateGroupOptions = [
  { value: 'backlog', label: '待处理 (Backlog)', color: '#6B7280' },
  { value: 'unstarted', label: '未开始 (Todo)', color: '#3B82F6' },
  { value: 'started', label: '进行中 (In Progress)', color: '#8B5CF6' },
  { value: 'completed', label: '已完成 (Done)', color: '#10B981' },
  { value: 'cancelled', label: '已取消 (Cancelled)', color: '#EF4444' },
]

export const PriorityOptions = [
  { value: 'urgent', label: '紧急 (Urgent)', color: '#EF4444' },
  { value: 'high', label: '高 (High)', color: '#F59E0B' },
  { value: 'medium', label: '中 (Medium)', color: '#3B82F6' },
  { value: 'low', label: '低 (Low)', color: '#6B7280' },
]

export enum ConditionOperatorEnum {
  EQUALS = 'equals',
  NOT_EQUALS = 'not_equals',
  CONTAINS = 'contains',
  NOT_CONTAINS = 'not_contains',
  IN = 'in',
  NOT_IN = 'not_in',
  GREATER_THAN = 'greater_than',
  LESS_THAN = 'less_than',
  IS_EMPTY = 'is_empty',
  IS_NOT_EMPTY = 'is_not_empty'
}

export const ConditionOperatorOptions = [
  { value: 'equals', label: '等于' },
  { value: 'not_equals', label: '不等于' },
  { value: 'contains', label: '包含' },
  { value: 'not_contains', label: '不包含' },
  { value: 'in', label: '在列表中' },
  { value: 'not_in', label: '不在列表中' },
  { value: 'is_empty', label: '为空' },
  { value: 'is_not_empty', label: '不为空' },
]

export enum ActionTypeEnum {
  // 工作项动作
  ISSUE_UPDATE = 'issue.update',
  ISSUE_ASSIGN = 'issue.assign',
  ISSUE_ADD_LABEL = 'issue.add_label',
  ISSUE_REMOVE_LABEL = 'issue.remove_label',
  ISSUE_CHANGE_STATE = 'issue.change_state',
  ISSUE_SET_PRIORITY = 'issue.set_priority',
  ISSUE_ADD_COMMENT = 'issue.add_comment',
  ISSUE_SET_DUE_DATE = 'issue.set_due_date',
  
  // 通知动作
  NOTIFICATION_CREATE = 'notification.create',
  EMAIL_SEND = 'email.send',
}

export const ActionTypeOptions = [
  { value: 'issue.change_state', label: '更新状态', icon: '🔄' },
  { value: 'issue.set_priority', label: '设置优先级', icon: '⚡' },
  { value: 'issue.assign', label: '分配给', icon: '👤' },
  { value: 'issue.add_label', label: '添加标签', icon: '🏷️' },
  { value: 'issue.remove_label', label: '移除标签', icon: '❌' },
  { value: 'issue.add_comment', label: '添加评论', icon: '💬' },
  { value: 'notification.create', label: '发送通知', icon: '🔔' },
]

// ==================== State Transition ====================

export interface StateTransition {
  id: number
  name: string
  description?: string
  source_state_id: number
  target_state_id: number
  issue_type_id?: number
  is_auto: boolean
  project_id: number
  workspace_id: number
  is_deleted: boolean
  created_at: string
  updated_at: string
}

export interface StateTransitionCreate {
  name: string
  description?: string
  source_state_id: number
  target_state_id: number
  issue_type_id?: number
  is_auto?: boolean
}

export interface StateTransitionUpdate {
  name?: string
  description?: string
  is_auto?: boolean
  issue_type_id?: number
}

// ==================== Automation Rule ====================

export interface Trigger {
  type: TriggerTypeEnum
  state_group?: string
  priority?: string
  days_before?: number
  field?: string
}

export interface Condition {
  field: string
  operator: ConditionOperatorEnum
  value?: any
}

export interface Action {
  type: ActionTypeEnum
  field?: string
  value?: any
  label?: string
  state_id?: number
  message?: string
  subject?: string
  recipients?: number[]
}

export interface AutomationRule {
  id: number
  name: string
  description?: string
  is_enabled: boolean
  trigger: Trigger
  conditions: Condition[]
  actions: Action[]
  execution_count: number
  last_executed_at?: string
  project_id: number
  workspace_id: number
  is_deleted: boolean
  created_at: string
  updated_at: string
}

export interface AutomationRuleCreate {
  name: string
  description?: string
  is_enabled?: boolean
  trigger: Trigger
  conditions?: Condition[]
  actions?: Action[]
  project_id: number
}

export interface AutomationRuleUpdate {
  name?: string
  description?: string
  is_enabled?: boolean
  trigger?: Trigger
  conditions?: Condition[]
  actions?: Action[]
}

export interface AutomationRuleLite {
  id: number
  name: string
  description?: string
  is_enabled: boolean
  trigger: Trigger
  execution_count: number
}

// ==================== Automation Execution Log ====================

export interface AutomationExecutionLog {
  id: number
  rule_id: number
  status: 'success' | 'failed' | 'skipped'
  trigger_event: string
  triggered_issue_id?: number
  execution_details: Record<string, any>
  error_message?: string
  execution_time_ms?: number
  created_at: string
}

// ==================== Automation Template ====================

export interface AutomationTemplate {
  id: number
  name: string
  description?: string
  category: string
  trigger: Trigger
  conditions: Condition[]
  actions: Action[]
  is_system: boolean
  usage_count: number
}

// ==================== Helper Functions ====================

/**
 * 获取触发器类型的显示名称
 */
export function getTriggerDisplayName(triggerType: TriggerTypeEnum): string {
  const names: Record<TriggerTypeEnum, string> = {
    [TriggerTypeEnum.ISSUE_CREATED]: '工作项创建时',
    [TriggerTypeEnum.ISSUE_UPDATED]: '工作项更新时',
    [TriggerTypeEnum.ISSUE_DELETED]: '工作项删除时',
    [TriggerTypeEnum.ISSUE_ASSIGNED]: '工作项分配时',
    [TriggerTypeEnum.STATE_CHANGED]: '状态变更时',
    [TriggerTypeEnum.PRIORITY_CHANGED]: '优先级变更时',
    [TriggerTypeEnum.DUE_SOON]: '截止日期临近时',
    [TriggerTypeEnum.DUE_DATE_PASSED]: '截止日期过期时',
    [TriggerTypeEnum.CYCLE_STARTED]: '周期开始时',
    [TriggerTypeEnum.CYCLE_ENDED]: '周期结束时',
    [TriggerTypeEnum.COMMENT_ADDED]: '添加评论时'
  }
  return names[triggerType] || triggerType
}

/**
 * 获取动作类型的显示名称
 */
export function getActionDisplayName(actionType: ActionTypeEnum): string {
  const names: Record<ActionTypeEnum, string> = {
    [ActionTypeEnum.ISSUE_UPDATE]: '更新工作项',
    [ActionTypeEnum.ISSUE_ASSIGN]: '分配工作项',
    [ActionTypeEnum.ISSUE_ADD_LABEL]: '添加标签',
    [ActionTypeEnum.ISSUE_REMOVE_LABEL]: '移除标签',
    [ActionTypeEnum.ISSUE_CHANGE_STATE]: '改变状态',
    [ActionTypeEnum.ISSUE_SET_PRIORITY]: '设置优先级',
    [ActionTypeEnum.ISSUE_ADD_COMMENT]: '添加评论',
    [ActionTypeEnum.ISSUE_SET_DUE_DATE]: '设置截止日期',
    [ActionTypeEnum.NOTIFICATION_CREATE]: '创建通知',
    [ActionTypeEnum.EMAIL_SEND]: '发送邮件',
  }
  return names[actionType] || actionType
}

/**
 * 获取条件操作符的显示名称
 */
export function getOperatorDisplayName(operator: ConditionOperatorEnum): string {
  const names: Record<ConditionOperatorEnum, string> = {
    [ConditionOperatorEnum.EQUALS]: '等于',
    [ConditionOperatorEnum.NOT_EQUALS]: '不等于',
    [ConditionOperatorEnum.CONTAINS]: '包含',
    [ConditionOperatorEnum.NOT_CONTAINS]: '不包含',
    [ConditionOperatorEnum.IN]: '在列表中',
    [ConditionOperatorEnum.NOT_IN]: '不在列表中',
    [ConditionOperatorEnum.GREATER_THAN]: '大于',
    [ConditionOperatorEnum.LESS_THAN]: '小于',
    [ConditionOperatorEnum.IS_EMPTY]: '为空',
    [ConditionOperatorEnum.IS_NOT_EMPTY]: '不为空'
  }
  return names[operator] || operator
}

/**
 * 获取触发器类型的图标
 */
export function getTriggerIcon(triggerType: TriggerTypeEnum): string {
  const icons: Record<TriggerTypeEnum, string> = {
    [TriggerTypeEnum.ISSUE_CREATED]: 'plus-circle',
    [TriggerTypeEnum.ISSUE_UPDATED]: 'edit',
    [TriggerTypeEnum.ISSUE_DELETED]: 'trash',
    [TriggerTypeEnum.ISSUE_ASSIGNED]: 'user',
    [TriggerTypeEnum.STATE_CHANGED]: 'git-branch',
    [TriggerTypeEnum.PRIORITY_CHANGED]: 'flag',
    [TriggerTypeEnum.DUE_SOON]: 'clock',
    [TriggerTypeEnum.DUE_DATE_PASSED]: 'alert-circle',
    [TriggerTypeEnum.CYCLE_STARTED]: 'play',
    [TriggerTypeEnum.CYCLE_ENDED]: 'check-circle',
    [TriggerTypeEnum.COMMENT_ADDED]: 'message-circle'
  }
  return icons[triggerType] || 'zap'
}

/**
 * 获取动作类型的图标
 */
export function getActionIcon(actionType: ActionTypeEnum): string {
  const icons: Record<ActionTypeEnum, string> = {
    [ActionTypeEnum.ISSUE_UPDATE]: 'edit-2',
    [ActionTypeEnum.ISSUE_ASSIGN]: 'user-plus',
    [ActionTypeEnum.ISSUE_ADD_LABEL]: 'tag',
    [ActionTypeEnum.ISSUE_REMOVE_LABEL]: 'x-circle',
    [ActionTypeEnum.ISSUE_CHANGE_STATE]: 'git-branch',
    [ActionTypeEnum.ISSUE_SET_PRIORITY]: 'flag',
    [ActionTypeEnum.ISSUE_ADD_COMMENT]: 'message-circle',
    [ActionTypeEnum.ISSUE_SET_DUE_DATE]: 'calendar',
    [ActionTypeEnum.NOTIFICATION_CREATE]: 'bell',
    [ActionTypeEnum.EMAIL_SEND]: 'mail',
  }
  return icons[actionType] || 'zap'
}