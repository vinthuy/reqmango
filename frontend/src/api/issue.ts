/**
 * Issue API - 工作项 API 调用模块
 */
import api from './index'
import type {
  IssueCreate,
  IssueUpdate,
  IssueResponse,
  IssueSearchResult,
  IssueActivity,
  IssueStatistics,
  IssuePriority
} from '@/types/issue'

// ==================== Issue CRUD ====================

/**
 * 创建工作项
 */
export async function createIssue(
  projectId: number,
  workspaceId: number,
  data: IssueCreate
): Promise<IssueResponse> {
  const response = await api.post(
    `/issues/?project_id=${projectId}&workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出项目的工作项（带分页信息）
 */
export interface IssueListResult {
  items: IssueResponse[]
  total: number
}

export async function listIssues(
  projectId: number,
  workspaceId: number,
  filters?: {
    state_id?: number
    priority?: IssuePriority
    assignee_id?: number
    parent_id?: number
    cycle_id?: number
    module_id?: number
    search?: string
    is_draft?: boolean
    issue_type_id?: number
    cf_field_id?: number
    cf_value?: string
    cf_and?: string
    limit?: number
    offset?: number
  }
): Promise<IssueListResult> {
  const params = new URLSearchParams()
  params.append('project_id', projectId.toString())
  params.append('workspace_id', workspaceId.toString())
  
  if (filters) {
    if (filters.state_id) params.append('state_id', filters.state_id.toString())
    if (filters.priority) params.append('priority', filters.priority)
    if (filters.assignee_id) params.append('assignee_id', filters.assignee_id.toString())
    if (filters.parent_id) params.append('parent_id', filters.parent_id.toString())
    if (filters.cycle_id) params.append('cycle_id', filters.cycle_id.toString())
    if (filters.module_id) params.append('module_id', filters.module_id.toString())
    if (filters.search) params.append('search', filters.search)
    if (filters.is_draft !== undefined) params.append('is_draft', filters.is_draft.toString())
    if (filters.issue_type_id) params.append('issue_type_id', filters.issue_type_id.toString())
    if (filters.cf_field_id) params.append('cf_field_id', filters.cf_field_id.toString())
    if (filters.cf_value) params.append('cf_value', filters.cf_value)
    if (filters.cf_and) params.append('cf_and', filters.cf_and)
    if (filters.limit) params.append('limit', filters.limit.toString())
    if (filters.offset) params.append('offset', filters.offset.toString())
  }
  
  const response = await api.get(`/issues/?${params.toString()}`)
  const total = parseInt(response.headers['x-total-count'] || '0', 10)
  return { items: response.data || [], total }
}

/**
 * 获取工作项详情
 */
export async function getIssue(issueId: number): Promise<IssueResponse> {
  const response = await api.get(`/issues/${issueId}`)
  return response.data
}

/**
 * 更新工作项
 */
export async function updateIssue(
  issueId: number,
  data: IssueUpdate
): Promise<IssueResponse> {
  const response = await api.put(`/issues/${issueId}`, data)
  return response.data
}

/**
 * 删除工作项
 */
export async function deleteIssue(issueId: number): Promise<void> {
  await api.delete(`/issues/${issueId}`)
}

/**
 * 归档工作项
 */
export async function archiveIssue(issueId: number): Promise<void> {
  await api.post(`/issues/${issueId}/archive`)
}

/**
 * 恢复工作项
 */
export async function restoreIssue(issueId: number): Promise<IssueResponse> {
  const response = await api.post(`/issues/${issueId}/restore`)
  return response.data
}

// ==================== Issue Activities ====================

/**
 * 获取工作项活动历史
 */
export async function getIssueActivities(
  issueId: number,
  limit?: number,
  offset?: number
): Promise<IssueActivity[]> {
  const params = new URLSearchParams()
  if (limit) params.append('limit', limit.toString())
  if (offset) params.append('offset', offset.toString())
  
  const response = await api.get(`/issues/${issueId}/activities?${params.toString()}`)
  return response.data
}

// ==================== Issue Statistics ====================

/**
 * 获取项目工作项统计
 */
export async function getIssueStatistics(projectId: number): Promise<IssueStatistics> {
  const response = await api.get(`/issues/statistics?project_id=${projectId}`)
  return response.data
}

// ==================== Issue Search ====================

/**
 * 搜索工作项
 */
export async function searchIssues(
  workspaceId: number,
  query: string,
  projectId?: number,
  limit?: number
): Promise<IssueSearchResult[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', workspaceId.toString())
  params.append('query', query)
  
  if (projectId) params.append('project_id', projectId.toString())
  if (limit) params.append('limit', limit.toString())
  
  const response = await api.get(`/issues/search?${params.toString()}`)
  return response.data
}

// ==================== Bulk Operations ====================

/**
 * 批量更新工作项
 */
export async function bulkUpdateIssues(
  projectId: number,
  issueIds: number[],
  data: IssueUpdate
): Promise<IssueResponse[]> {
  const response = await api.post(
    `/issues/bulk/update?project_id=${projectId}`,
    { issue_ids: issueIds, ...data }
  )
  return response.data
}

/**
 * 批量删除工作项
 */
export async function bulkDeleteIssues(issueIds: number[]): Promise<void> {
  await api.post('/issues/bulk/delete', { issue_ids: issueIds })
}

export interface ImportError {
  row: number
  title: string
  message: string
}

export interface ImportResult {
  success_count: number
  fail_count: number
  errors: ImportError[]
  imported_ids: number[]
}

export interface ImportIssueItem {
  name: string
  description?: string
  priority?: string
  state_name?: string
  type_name?: string
  assignee_emails?: string[]
  label_names?: string[]
  start_date?: string
  target_date?: string
  parent_title?: string
}

/**
 * JSON 批量导入工作项
 */
export async function importIssuesJSON(
  projectId: number,
  workspaceId: number,
  items: ImportIssueItem[]
): Promise<ImportResult> {
  const response = await api.post(
    `/issues/import/json?project_id=${projectId}&workspace_id=${workspaceId}`,
    items
  )
  return response.data
}

/**
 * CSV 批量导入工作项
 */
export async function importIssuesCSV(
  projectId: number,
  workspaceId: number,
  file: File
): Promise<ImportResult> {
  const formData = new FormData()
  formData.append('file', file)
  const response = await api.post(
    `/issues/import/csv?project_id=${projectId}&workspace_id=${workspaceId}`,
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    }
  )
  return response.data
}

// ==================== Assignee Management ====================

/**
 * 添加工作项负责人
 */
export async function addIssueAssignee(
  issueId: number,
  userId: number
): Promise<{ issue_id: number; user_id: number; action: string }> {
  const response = await api.post(`/issues/${issueId}/assignees?user_id=${userId}`)
  return response.data
}

/**
 * 移除工作项负责人
 */
export async function removeIssueAssignee(
  issueId: number,
  userId: number
): Promise<{ issue_id: number; user_id: number; action: string }> {
  const response = await api.delete(`/issues/${issueId}/assignees/${userId}`)
  return response.data
}

// ==================== Label Management ====================

/**
 * 添加工作项标签
 */
export async function addIssueLabel(
  issueId: number,
  labelId: number
): Promise<{ issue_id: number; label_id: number; action: string }> {
  const response = await api.post(`/issues/${issueId}/labels?label_id=${labelId}`)
  return response.data
}

/**
 * 移除工作项标签
 */
export async function removeIssueLabel(
  issueId: number,
  labelId: number
): Promise<{ issue_id: number; label_id: number; action: string }> {
  const response = await api.delete(`/issues/${issueId}/labels/${labelId}`)
  return response.data
}

// ==================== Cycle Management ====================

/**
 * 设置工作项周期
 */
export async function setIssueCycle(
  issueId: number,
  cycleId: number
): Promise<{ issue_id: number; cycle_id: number; action: string }> {
  const response = await api.post(`/issues/${issueId}/cycle?cycle_id=${cycleId}`)
  return response.data
}

/**
 * 移除工作项周期
 */
export async function removeIssueCycle(
  issueId: number
): Promise<{ issue_id: number; cycle_id: null; action: string }> {
  const response = await api.delete(`/issues/${issueId}/cycle`)
  return response.data
}

// ==================== Pages ====================

export async function listIssuePages(issueId: number): Promise<any[]> {
  const response = await api.get(`/issues/${issueId}/pages`)
  return response.data
}

export async function addIssuePage(issueId: number, pageId: number): Promise<any> {
  const response = await api.post(`/issues/${issueId}/pages?page_id=${pageId}`)
  return response.data
}

export async function removeIssuePage(issueId: number, pageId: number): Promise<any> {
  const response = await api.delete(`/issues/${issueId}/pages?page_id=${pageId}`)
  return response.data
}

// ==================== Export all ====================

export const issueApi = {
  // CRUD
  createIssue,
  listIssues,
  getIssue,
  updateIssue,
  deleteIssue,
  archiveIssue,
  restoreIssue,
  
  // Activities
  getIssueActivities,
  
  // Statistics
  getIssueStatistics,
  
  // Search
  searchIssues,
  
  // Bulk
  bulkUpdateIssues,
  bulkDeleteIssues,
  importIssuesJSON,
  importIssuesCSV,
  
  // Assignee
  addIssueAssignee,
  removeIssueAssignee,
  
  // Label
  addIssueLabel,
  removeIssueLabel,
  
  // Cycle
  setIssueCycle,
  removeIssueCycle,
  
  // Pages
  listIssuePages,
  addIssuePage,
  removeIssuePage
}

export default issueApi