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
  ProjectMemberUpdate,
  ProjectStatistics,
  ProjectIssuesSummary
} from '@/types/project'

// ==================== Project CRUD ====================

/**
 * 创建项目
 */
export async function createProject(
  workspaceId: string,
  data: ProjectCreate
): Promise<ProjectResponse> {
  const response = await api.post(
    `/projects?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出工作空间的项目
 */
export async function listProjects(
  workspaceId: string,
  options?: {
    include_archived?: boolean
    limit?: number
    offset?: number
  }
): Promise<ProjectResponse[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', workspaceId)
  
  if (options?.include_archived) params.append('include_archived', 'true')
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  
  const response = await api.get(`/projects?${params.toString()}`)
  return response.data
}

/**
 * 获取项目详情
 */
export async function getProject(projectId: string): Promise<ProjectResponse> {
  const response = await api.get(`/projects/${projectId}`)
  return response.data
}

/**
 * 更新项目
 */
export async function updateProject(
  projectId: string,
  data: ProjectUpdate
): Promise<ProjectResponse> {
  const response = await api.patch(`/projects/${projectId}`, data)
  return response.data
}

/**
 * 删除项目
 */
export async function deleteProject(projectId: string): Promise<void> {
  await api.delete(`/projects/${projectId}`)
}

// ==================== Project Archive ====================

/**
 * 归档项目
 */
export async function archiveProject(projectId: string): Promise<ProjectResponse> {
  const response = await api.post(`/projects/${projectId}/archive`)
  return response.data
}

/**
 * 恢复项目
 */
export async function restoreProject(projectId: string): Promise<ProjectResponse> {
  const response = await api.post(`/projects/${projectId}/restore`)
  return response.data
}

// ==================== Project Members ====================

/**
 * 列出项目成员
 */
export async function listProjectMembers(
  projectId: string,
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
  projectId: string,
  data: ProjectMemberCreate
): Promise<{ id: string; project_id: string; user_id: string; role: number }> {
  const response = await api.post(
    `/projects/${projectId}/members?user_id=${data.user_id}&role=${data.role || 15}`
  )
  return response.data
}

/**
 * 更新项目成员角色
 */
export async function updateProjectMember(
  projectId: string,
  userId: string,
  role: number
): Promise<{ id: string; project_id: string; user_id: string; role: number }> {
  const response = await api.patch(
    `/projects/${projectId}/members/${userId}?role=${role}`
  )
  return response.data
}

/**
 * 移除项目成员
 */
export async function removeProjectMember(
  projectId: string,
  userId: string
): Promise<{ project_id: string; user_id: string; action: string }> {
  const response = await api.delete(`/projects/${projectId}/members/${userId}`)
  return response.data
}

// ==================== Project Statistics ====================

/**
 * 获取项目统计
 */
export async function getProjectStatistics(
  projectId: string
): Promise<ProjectStatistics> {
  const response = await api.get(`/projects/${projectId}/statistics`)
  return response.data
}

/**
 * 获取项目工作项摘要
 */
export async function getProjectIssuesSummary(
  projectId: string
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
  
  // Statistics
  getProjectStatistics,
  getProjectIssuesSummary
}

export default projectApi