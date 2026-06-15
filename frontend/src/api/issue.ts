/**
 * Issue API - 工作项 API 调用模块
 */
import api from './index'
import type {
  IssueCreate,
  IssueUpdate,
  IssueResponse,
  IssueLite,
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
  projectId: string,
  workspaceId: string,
  data: IssueCreate
): Promise<IssueResponse> {
  const response = await api.post(
    `/issues?project_id=${projectId}&workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出项目的工作项
 */
export async function listIssues(
  projectId: string,
  workspaceId: string,
  filters?: {
    state_id?: string
    priority?: IssuePriority
    assignee_id?: string
    parent_id?: string
    cycle_id?: string
    module_id?: string
    search?: string
    is_draft?: boolean
    limit?: number
    offset?: number
  }
): Promise<IssueResponse[]> {
  const params = new URLSearchParams()
  params.append('project_id', projectId)
  params.append('workspace_id', workspaceId)
  
  if (filters) {
    if (filters.state_id) params.append('state_id', filters.state_id)
    if (filters.priority) params.append('priority', filters.priority)
    if (filters.assignee_id) params.append('assignee_id', filters.assignee_id)
    if (filters.parent_id) params.append('parent_id', filters.parent_id)
    if (filters.cycle_id) params.append('cycle_id', filters.cycle_id)
    if (filters.module_id) params.append('module_id', filters.module_id)
    if (filters.search) params.append('search', filters.search)
    if (filters.is_draft !== undefined) params.append('is_draft', filters.is_draft.toString())
    if (filters.limit) params.append('limit', filters.limit.toString())
    if (filters.offset) params.append('offset', filters.offset.toString())
  }
  
  const response = await api.get(`/issues?${params.toString()}`)
  return response.data
}

/**
 * 获取工作项详情
 */
export async function getIssue(issueId: string): Promise<IssueResponse> {
  const response = await api.get(`/issues/${issueId}`)
  return response.data
}

/**
 * 更新工作项
 */
export async function updateIssue(
  issueId: string,
  data: IssueUpdate
): Promise<IssueResponse> {
  const response = await api.put(`/issues/${issueId}`, data)
  return response.data
}

/**
 * 删除工作项
 */
export async function deleteIssue(issueId: string): Promise<void> {
  await api.delete(`/issues/${issueId}`)
}

/**
 * 归档工作项
 */
export async function archiveIssue(issueId: string): Promise<void> {
  await api.post(`/issues/${issueId}/archive`)
}

/**
 * 恢复工作项
 */
export async function restoreIssue(issueId: string): Promise<IssueResponse> {
  const response = await api.post(`/issues/${issueId}/restore`)
  return response.data
}

// ==================== Issue Activities ====================

/**
 * 获取工作项活动历史
 */
export async function getIssueActivities(
  issueId: string,
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
export async function getIssueStatistics(projectId: string): Promise<IssueStatistics> {
  const response = await api.get(`/issues/statistics?project_id=${projectId}`)
  return response.data
}

// ==================== Issue Search ====================

/**
 * 搜索工作项
 */
export async function searchIssues(
  workspaceId: string,
  query: string,
  projectId?: string,
  limit?: number
): Promise<IssueSearchResult[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', workspaceId)
  params.append('query', query)
  
  if (projectId) params.append('project_id', projectId)
  if (limit) params.append('limit', limit.toString())
  
  const response = await api.get(`/issues/search?${params.toString()}`)
  return response.data
}

// ==================== Bulk Operations ====================

/**
 * 批量更新工作项
 */
export async function bulkUpdateIssues(
  projectId: string,
  issueIds: string[],
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
export async function bulkDeleteIssues(issueIds: string[]): Promise<void> {
  await api.post('/issues/bulk/delete', { issue_ids: issueIds })
}

// ==================== Assignee Management ====================

/**
 * 添加工作项负责人
 */
export async function addIssueAssignee(
  issueId: string,
  userId: string
): Promise<{ issue_id: string; user_id: string; action: string }> {
  const response = await api.post(`/issues/${issueId}/assignees?user_id=${userId}`)
  return response.data
}

/**
 * 移除工作项负责人
 */
export async function removeIssueAssignee(
  issueId: string,
  userId: string
): Promise<{ issue_id: string; user_id: string; action: string }> {
  const response = await api.delete(`/issues/${issueId}/assignees/${userId}`)
  return response.data
}

// ==================== Label Management ====================

/**
 * 添加工作项标签
 */
export async function addIssueLabel(
  issueId: string,
  labelId: string
): Promise<{ issue_id: string; label_id: string; action: string }> {
  const response = await api.post(`/issues/${issueId}/labels?label_id=${labelId}`)
  return response.data
}

/**
 * 移除工作项标签
 */
export async function removeIssueLabel(
  issueId: string,
  labelId: string
): Promise<{ issue_id: string; label_id: string; action: string }> {
  const response = await api.delete(`/issues/${issueId}/labels/${labelId}`)
  return response.data
}

// ==================== Cycle Management ====================

/**
 * 设置工作项周期
 */
export async function setIssueCycle(
  issueId: string,
  cycleId: string
): Promise<{ issue_id: string; cycle_id: string; action: string }> {
  const response = await api.post(`/issues/${issueId}/cycle?cycle_id=${cycleId}`)
  return response.data
}

/**
 * 移除工作项周期
 */
export async function removeIssueCycle(
  issueId: string
): Promise<{ issue_id: string; cycle_id: null; action: string }> {
  const response = await api.delete(`/issues/${issueId}/cycle`)
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
  
  // Assignee
  addIssueAssignee,
  removeIssueAssignee,
  
  // Label
  addIssueLabel,
  removeIssueLabel,
  
  // Cycle
  setIssueCycle,
  removeIssueCycle
}

export default issueApi