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
  parent_id?: number
  state_id?: number
  assignee_ids?: number[]
  label_ids?: number[]
  cycle_id?: number
  module_ids?: number[]
  estimate_point_id?: number
  type_id?: number
  external_id?: string
  external_source?: string
  project_id?: number
  workspace_id?: number
}

// ==================== Issue Update ====================

export interface IssueUpdate {
  name?: string
  description_html?: string
  priority?: IssuePriority
  state_id?: number
  sort_order?: number
  assignee_ids?: number[]
  label_ids?: number[]
  start_date?: string
  target_date?: string
  estimate_point_id?: number
  cycle_id?: number
  module_ids?: number[]
  type_id?: number
  parent_id?: number | null
}

// ==================== Issue Response ====================

export interface IssueResponse extends IssueBase {
  id: number
  sequence_id: number
  sort_order: number
  completed_at?: string
  is_draft: boolean
  archived_at?: string
  
  project_id: number
  workspace_id: number
  parent_id?: number
  state_id: number
  state_name?: string
  state_group?: string

  project?: { id: number; name: string; identifier: string }

  // 关联信息
  assignees?: UserLite[]
  labels?: number[]
  label_details?: { id: number; name: string; color: string }[]
  cycle_id?: number
  sub_issues_count?: number
  sub_issues?: IssueResponse[]
  parent?: IssueResponse | null
  link_count?: number
  attachment_count?: number
  
  estimate_point_id?: number
  module_ids?: number[]
  type_id?: number
  issue_type?: {
    id: number
    name: string
    color: string
    icon: string
  }

  // 时间戳
  created_at: string
  updated_at: string
  created_by_id?: number
  updated_by_id?: number
  deleted_at?: string
  is_deleted: boolean
}

// ==================== Issue Lite ====================

export interface IssueLite {
  id: number
  name: string
  sequence_id: number
  priority: IssuePriority
  state_id: number
  project_id: number
  project_identifier: string
}

// ==================== Issue Search Result ====================

export interface IssueSearchResult {
  id: number
  name: string
  sequence_id: number
  project_identifier: string
  project_id: number
  workspace_slug: string
}

// ==================== Issue Activity ====================

export interface IssueActivity {
  id: number
  issue_id: number
  verb: string
  field?: string
  old_value?: string
  new_value?: string
  comment?: string
  actor_id: number
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
  id: number
  display_name: string
  email: string
  avatar_url?: string
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
export function createEmptyIssue(projectId: number, workspaceId: number): IssueCreate {
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
  const identifier = (issue as any).project_identifier || (issue as IssueResponse).project?.identifier
  return identifier ? `${identifier}-${issue.sequence_id}` : `#${issue.sequence_id}`
}

// ==================== Tree View ====================

export interface TreeIssueResponse {
  id: number
  name: string
  sequence_id: number
  priority: string
  state_id: number
  state_name: string
  state_group: string
  parent_id?: number
  depth: number
  sub_issues_count: number
  has_children: boolean
  issue_type_id?: number
  issue_type?: {
    id: number
    name: string
    color: string
    icon: string
  }
  start_date?: string
  target_date?: string
  is_search_match: boolean
}

export interface AncestorInfo {
  id: number
  name: string
  sequence_id: number
}

export interface TreeSearchResult {
  root_issue: TreeIssueResponse
  matched_issue: TreeIssueResponse
  ancestor_chain: AncestorInfo[]
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
