/**
 * Issue Types - 工作项类型定义
 */

// ==================== Enums ====================

export enum IssuePriority {
  URGENT = 'urgent',
  HIGH = 'high',
  MEDIUM = 'medium',
  LOW = 'low',
  NONE = 'none'
}

export enum IssueType {
  ISSUE = 'issue',
  TASK = 'task',
  BUG = 'bug',
  STORY = 'story',
  EPIC = 'epic'
}

export enum IssueStateGroup {
  BACKLOG = 'backlog',
  TODO = 'todo',
  IN_PROGRESS = 'in_progress',
  DONE = 'done',
  CANCELLED = 'cancelled'
}

// ==================== Issue Base ====================

export interface IssueBase {
  name: string
  description_html?: string
  description_json?: Record<string, any>
  priority: IssuePriority
  start_date?: string
  target_date?: string
}

// ==================== Issue Create ====================

export interface IssueCreate extends IssueBase {
  parent_id?: string
  state_id?: string
  assignee_ids?: string[]
  label_ids?: string[]
  cycle_id?: string
  module_ids?: string[]
  estimate_point_id?: string
  type_id?: string
  external_id?: string
  external_source?: string
}

// ==================== Issue Update ====================

export interface IssueUpdate {
  name?: string
  description_html?: string
  priority?: IssuePriority
  state_id?: string
  assignee_ids?: string[]
  label_ids?: string[]
  start_date?: string
  target_date?: string
  estimate_point_id?: string
  cycle_id?: string
  module_ids?: string[]
}

// ==================== Issue Response ====================

export interface IssueResponse extends IssueBase {
  id: string
  sequence_id: number
  sort_order: number
  completed_at?: string
  is_draft: boolean
  archived_at?: string
  
  project_id: string
  workspace_id: string
  parent_id?: string
  state_id: string
  state_name?: string
  state_group?: string
  
  project_identifier?: string
  
  // 关联信息
  assignees?: UserLite[]
  labels?: LabelLite[]
  cycle?: CycleLite
  sub_issues_count?: number
  link_count?: number
  attachment_count?: number
  
  estimate_point_id?: string
  module_ids?: string[]
  
  // 时间戳
  created_at: string
  updated_at: string
  created_by: string
  updated_by?: string
}

// ==================== Issue Lite ====================

export interface IssueLite {
  id: string
  name: string
  sequence_id: number
  priority: IssuePriority
  state_id: string
  project_id: string
  project_identifier: string
}

// ==================== Issue Search Result ====================

export interface IssueSearchResult {
  id: string
  name: string
  sequence_id: number
  project_identifier: string
  project_id: string
  workspace_slug: string
}

// ==================== Issue Activity ====================

export interface IssueActivity {
  id: string
  issue_id: string
  verb: string
  field?: string
  old_value?: string
  new_value?: string
  comment?: string
  actor_id: string
  created_at: string
}

// ==================== Issue Statistics ====================

export interface IssueStatistics {
  total: number
  by_state: Record<string, number>
  by_priority: Record<string, number>
}

// ==================== User Lite ====================

export interface UserLite {
  id: string
  display_name: string
  username: string
  avatar?: string
}

// ==================== Label Lite ====================

export interface LabelLite {
  id: string
  name: string
  color: string
}

// ==================== Cycle Lite ====================

export interface CycleLite {
  id: string
  name: string
}

// ==================== Helper Functions ====================

/**
 * 获取优先级的显示名称
 */
export function getPriorityName(priority: IssuePriority): string {
  const names: Record<IssuePriority, string> = {
    [IssuePriority.URGENT]: '紧急',
    [IssuePriority.HIGH]: '高',
    [IssuePriority.MEDIUM]: '中',
    [IssuePriority.LOW]: '低',
    [IssuePriority.NONE]: '无'
  }
  return names[priority] || priority
}

/**
 * 获取优先级的颜色
 */
export function getPriorityColor(priority: IssuePriority): string {
  const colors: Record<IssuePriority, string> = {
    [IssuePriority.URGENT]: '#EF4444', // 红色
    [IssuePriority.HIGH]: '#F59E0B',   // 橙色
    [IssuePriority.MEDIUM]: '#3B82F6', // 蓝色
    [IssuePriority.LOW]: '#10B981',    // 绿色
    [IssuePriority.NONE]: '#6B7280'    // 灰色
  }
  return colors[priority] || '#6B7280'
}

/**
 * 获取状态分组的显示名称
 */
export function getStateGroupName(group: IssueStateGroup): string {
  const names: Record<IssueStateGroup, string> = {
    [IssueStateGroup.BACKLOG]: '待办',
    [IssueStateGroup.TODO]: '计划中',
    [IssueStateGroup.IN_PROGRESS]: '进行中',
    [IssueStateGroup.DONE]: '已完成',
    [IssueStateGroup.CANCELLED]: '已取消'
  }
  return names[group] || group
}

/**
 * 获取状态分组的颜色
 */
export function getStateGroupColor(group: IssueStateGroup): string {
  const colors: Record<IssueStateGroup, string> = {
    [IssueStateGroup.BACKLOG]: '#6B7280', // 灰色
    [IssueStateGroup.TODO]: '#3B82F6',    // 蓝色
    [IssueStateGroup.IN_PROGRESS]: '#F59E0B', // 橙色
    [IssueStateGroup.DONE]: '#10B981',    // 绿色
    [IssueStateGroup.CANCELLED]: '#EF4444' // 红色
  }
  return colors[group] || '#6B7280'
}

/**
 * 格式化工作项标识符
 */
export function formatIssueIdentifier(projectIdentifier: string, sequenceId: number): string {
  return `${projectIdentifier}-${sequenceId}`
}

/**
 * 创建空工作项对象
 */
export function createEmptyIssue(projectId: string, workspaceId: string): IssueCreate {
  return {
    name: '',
    description_html: '<p></p>',
    description_json: {},
    priority: IssuePriority.NONE,
    project_id: projectId,
    workspace_id: workspaceId
  }
}

/**
 * 判断工作项是否已完成
 */
export function isIssueCompleted(issue: IssueResponse): boolean {
  return issue.state_group === IssueStateGroup.DONE || issue.completed_at !== undefined
}

/**
 * 判断工作项是否过期
 */
export function isIssueOverdue(issue: IssueResponse): boolean {
  if (!issue.target_date) return false
  const targetDate = new Date(issue.target_date)
  const now = new Date()
  return targetDate < now && !isIssueCompleted(issue)
}

/**
 * 判断工作项是否为草稿
 */
export function isIssueDraft(issue: IssueResponse): boolean {
  return issue.is_draft
}

/**
 * 判断工作项是否已归档
 */
export function isIssueArchived(issue: IssueResponse): boolean {
  return issue.archived_at !== undefined
}

/**
 * 获取工作项的完整标识符
 */
export function getIssueFullId(issue: IssueResponse | IssueLite): string {
  return issue.project_identifier ? `${issue.project_identifier}-${issue.sequence_id}` : `#${issue.sequence_id}`
}

/**
 * 计算工作项的进度百分比
 */
export function calculateIssueProgress(issue: IssueResponse): number {
  if (isIssueCompleted(issue)) return 100
  if (issue.state_group === IssueStateGroup.BACKLOG) return 0
  if (issue.state_group === IssueStateGroup.TODO) return 10
  if (issue.state_group === IssueStateGroup.IN_PROGRESS) return 50
  return 0
}