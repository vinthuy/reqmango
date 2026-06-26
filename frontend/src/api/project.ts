/**
 * Project API - 项目 API 调用模块
 */
import api from './index'
import type {
  ProjectCreate,
  ProjectUpdate,
  ProjectResponse,
  ProjectMember,
  ProjectMemberCreate,
  ProjectSubscriber,
  ProjectStatistics,
  ProjectIssuesSummary
} from '@/types/project'

// ==================== Project CRUD ====================

/**
 * 创建项目
 */
export async function createProject(
  workspaceId: number,
  data: ProjectCreate
): Promise<ProjectResponse> {
  let url = `/projects/?workspace_id=${workspaceId}`
  if (data.template_id) {
    url += `&template_id=${data.template_id}`
  }
  const response = await api.post(url, data)
  return response.data
}

/**
 * 列出工作空间的项目
 */
export async function listProjects(
  workspaceId: number,
  options?: {
    include_archived?: boolean
    limit?: number
    offset?: number
  }
): Promise<ProjectResponse[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', workspaceId.toString())
  
  if (options?.include_archived) params.append('include_archived', 'true')
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  
  const response = await api.get(`/projects/?${params.toString()}`)
  return response.data
}

/**
 * 获取项目详情
 */
export async function getProject(projectId: number): Promise<ProjectResponse> {
  const response = await api.get(`/projects/${projectId}`)
  return response.data
}

/**
 * 更新项目
 */
export async function updateProject(
  projectId: number,
  data: ProjectUpdate
): Promise<ProjectResponse> {
  const response = await api.patch(`/projects/${projectId}`, data)
  return response.data
}

/**
 * 删除项目
 */
export async function deleteProject(projectId: number): Promise<void> {
  await api.delete(`/projects/${projectId}`)
}

// ==================== Project Archive ====================

/**
 * 归档项目
 */
export async function archiveProject(projectId: number): Promise<ProjectResponse> {
  const response = await api.post(`/projects/${projectId}/archive`)
  return response.data
}

/**
 * 恢复项目
 */
export async function restoreProject(projectId: number): Promise<ProjectResponse> {
  const response = await api.post(`/projects/${projectId}/restore`)
  return response.data
}

// ==================== Project Members ====================

/**
 * 列出项目成员
 */
export async function listProjectMembers(
  projectId: number,
  onlyActive: boolean = true
): Promise<ProjectMember[]> {
  const response = await api.get(
    `/projects/${projectId}/members?only_active=${onlyActive}`
  )
  return response.data
}

/**
 * 添加项目成员
 */
export async function addProjectMember(
  projectId: number,
  data: ProjectMemberCreate
): Promise<{ id: number; project_id: number; user_id: number; role: number }> {
  const response = await api.post(
    `/projects/${projectId}/members?user_id=${data.user_id}&role=${data.role || 15}`
  )
  return response.data
}

/**
 * 更新项目成员角色
 */
export async function updateProjectMember(
  projectId: number,
  userId: number,
  role: number
): Promise<{ id: number; project_id: number; user_id: number; role: number }> {
  const response = await api.patch(
    `/projects/${projectId}/members/${userId}?role=${role}`
  )
  return response.data
}

/**
 * 移除项目成员
 */
export async function removeProjectMember(
  projectId: number,
  userId: number
): Promise<{ project_id: number; user_id: number; action: string }> {
  const response = await api.delete(`/projects/${projectId}/members/${userId}`)
  return response.data
}

// ==================== Project Lead & Default Assignee ====================

/**
 * 更新项目默认处理人
 */
export async function updateDefaultAssignee(
  projectId: number,
  userId: number | null
): Promise<ProjectResponse> {
  const response = await api.patch(
    `/projects/${projectId}?default_assignee_id=${userId ?? ''}`
  )
  return response.data
}

/**
 * 更新项目负责人
 */
export async function updateProjectLead(
  projectId: number,
  userId: number | null
): Promise<ProjectResponse> {
  const response = await api.patch(
    `/projects/${projectId}?project_lead_id=${userId ?? ''}`
  )
  return response.data
}

// ==================== Project Subscribers ====================

/**
 * 列出项目订阅者
 */
export async function listProjectSubscribers(
  projectId: number
): Promise<ProjectSubscriber[]> {
  const response = await api.get(`/projects/${projectId}/subscribers`)
  return response.data
}

/**
 * 添加项目订阅者
 */
export async function addProjectSubscriber(
  projectId: number,
  userId: number
): Promise<any> {
  const response = await api.post(
    `/projects/${projectId}/subscribers?user_id=${userId}`
  )
  return response.data
}

/**
 * 移除项目订阅者
 */
export async function removeProjectSubscriber(
  projectId: number,
  userId: number
): Promise<any> {
  const response = await api.delete(
    `/projects/${projectId}/subscribers/${userId}`
  )
  return response.data
}

// ==================== Project Statistics ====================

/**
 * 获取项目统计
 */
export async function getProjectStatistics(
  projectId: number
): Promise<ProjectStatistics> {
  const response = await api.get(`/projects/${projectId}/statistics`)
  return response.data
}

/**
 * 获取项目工作项摘要
 */
export async function getProjectIssuesSummary(
  projectId: number
): Promise<ProjectIssuesSummary> {
  const response = await api.get(`/projects/${projectId}/issues-summary`)
  return response.data
}

// ==================== Export all ====================

export const projectApi = {
  // CRUD
  createProject,
  listProjects,
  getProject,
  updateProject,
  deleteProject,
  
  // Archive
  archiveProject,
  restoreProject,
  
  // Members
  listProjectMembers,
  addProjectMember,
  updateProjectMember,
  removeProjectMember,
  
  // Lead & Default Assignee
  updateDefaultAssignee,
  updateProjectLead,
  
  // Subscribers
  listProjectSubscribers,
  addProjectSubscriber,
  removeProjectSubscriber,
  
  // Statistics
  getProjectStatistics,
  getProjectIssuesSummary
}

export default projectApi