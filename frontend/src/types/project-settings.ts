/**
 * Project Settings Types - 项目设置类型定义（IssueType、State、Label）
 */

// ==================== Enums ====================

export enum StateGroupEnum {
  BACKLOG = 'backlog',
  TODO = 'todo',
  IN_PROGRESS = 'in_progress',
  DONE = 'done',
  CANCELLED = 'cancelled'
}

export enum IssuePriorityEnum {
  URGENT = 'urgent',
  HIGH = 'high',
  MEDIUM = 'medium',
  LOW = 'low',
  NONE = 'none'
}

// ==================== IssueType ====================

export interface IssueType {
  id: string
  name: string
  color: string
  icon: string
  is_default: boolean
  sequence: number
  is_active: boolean
  project_id: string
  workspace_id: string
  created_at: string
  updated_at: string
}

export interface IssueTypeCreate {
  name: string
  color?: string
  icon?: string
  is_default?: boolean
  sequence?: number
  is_active?: boolean
  project_id: string
  custom_fields?: string[]
}

export interface IssueTypeUpdate {
  name?: string
  color?: string
  icon?: string
  is_default?: boolean
  sequence?: number
  is_active?: boolean
  custom_fields?: string[]
}

export interface IssueTypeLite {
  id: string
  name: string
  color: string
  icon: string
  is_default: boolean
}

// ==================== State ====================

export interface State {
  id: string
  name: string
  color: string
  group: StateGroupEnum
  sequence: number
  is_active: boolean
  description?: string
  project_id: string
  workspace_id: string
  created_at: string
  updated_at: string
}

export interface StateCreate {
  name: string
  color?: string
  group?: StateGroupEnum
  sequence?: number
  is_active?: boolean
  description?: string
  project_id: string
}

export interface StateUpdate {
  name?: string
  color?: string
  group?: StateGroupEnum
  sequence?: number
  is_active?: boolean
  description?: string
}

export interface StateLite {
  id: string
  name: string
  color: string
  group: string
}

// ==================== Label ====================

export interface Label {
  id: string
  name: string
  color: string
  description?: string
  project_id: string
  created_at: string
  updated_at: string
}

export interface LabelCreate {
  name: string
  color?: string
  description?: string
  project_id: string
}

export interface LabelUpdate {
  name?: string
  color?: string
  description?: string
}

export interface LabelLite {
  id: string
  name: string
  color: string
}

// ==================== Helper Functions ====================

/**
 * 获取状态分组的显示名称
 */
export function getStateGroupName(group: StateGroupEnum): string {
  const names: Record<StateGroupEnum, string> = {
    [StateGroupEnum.BACKLOG]: '待办池',
    [StateGroupEnum.TODO]: '待办',
    [StateGroupEnum.IN_PROGRESS]: '进行中',
    [StateGroupEnum.DONE]: '已完成',
    [StateGroupEnum.CANCELLED]: '已取消'
  }
  return names[group] || group
}

/**
 * 获取优先级的显示名称
 */
export function getPriorityName(priority: IssuePriorityEnum): string {
  const names: Record<IssuePriorityEnum, string> = {
    [IssuePriorityEnum.URGENT]: '紧急',
    [IssuePriorityEnum.HIGH]: '高',
    [IssuePriorityEnum.MEDIUM]: '中',
    [IssuePriorityEnum.LOW]: '低',
    [IssuePriorityEnum.NONE]: '无'
  }
  return names[priority] || priority
}

/**
 * 获取优先级的颜色
 */
export function getPriorityColor(priority: IssuePriorityEnum): string {
  const colors: Record<IssuePriorityEnum, string> = {
    [IssuePriorityEnum.URGENT]: '#EF4444',
    [IssuePriorityEnum.HIGH]: '#F59E0B',
    [IssuePriorityEnum.MEDIUM]: '#3B82F6',
    [IssuePriorityEnum.LOW]: '#6B7280',
    [IssuePriorityEnum.NONE]: '#9CA3AF'
  }
  return colors[priority] || '#6B7280'
}

/**
 * 默认工作项类型
 */
export const DEFAULT_ISSUE_TYPES: IssueTypeCreate[] = [
  { name: 'Issue', color: '#3B82F6', icon: 'circle', is_default: true, sequence: 1 },
  { name: 'Task', color: '#10B981', icon: 'check-circle', sequence: 2 },
  { name: 'Bug', color: '#EF4444', icon: 'alert-circle', sequence: 3 },
  { name: 'Story', color: '#F59E0B', icon: 'bookmark', sequence: 4 },
  { name: 'Epic', color: '#8B5CF6', icon: 'layers', sequence: 5 }
]

/**
 * 默认状态
 */
export const DEFAULT_STATES: StateCreate[] = [
  { name: 'Backlog', color: '#6B7280', group: StateGroupEnum.BACKLOG, sequence: 1 },
  { name: 'Todo', color: '#3B82F6', group: StateGroupEnum.TODO, sequence: 2 },
  { name: 'In Progress', color: '#F59E0B', group: StateGroupEnum.IN_PROGRESS, sequence: 3 },
  { name: 'In Review', color: '#8B5CF6', group: StateGroupEnum.IN_PROGRESS, sequence: 4 },
  { name: 'Done', color: '#10B981', group: StateGroupEnum.DONE, sequence: 5 },
  { name: 'Cancelled', color: '#EF4444', group: StateGroupEnum.CANCELLED, sequence: 6 }
]