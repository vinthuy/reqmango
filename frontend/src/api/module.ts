/**
 * Module API - 模块 API 调用模块
 */
import api from './index'
import type {
  ModuleCreate,
  ModuleUpdate,
  ModuleResponse,
  ModuleProgress,
  ModuleStatistics,
  ModuleTreeNode
} from '@/types/module'

// ==================== Module CRUD ====================

/**
 * 创建模块
 */
export async function createModule(
  workspaceId: string,
  data: ModuleCreate
): Promise<ModuleResponse> {
  const response = await api.post(
    `/modules?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出项目的模块
 */
export async function listModules(
  projectId: string,
  workspaceId: string,
  options?: {
    parent_id?: string
    include_archived?: boolean
    limit?: number
    offset?: number
  }
): Promise<ModuleResponse[]> {
  const params = new URLSearchParams()
  params.append('project_id', projectId)
  params.append('workspace_id', workspaceId)
  
  if (options?.parent_id) params.append('parent_id', options.parent_id)
  if (options?.include_archived) params.append('include_archived', 'true')
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  
  const response = await api.get(`/modules?${params.toString()}`)
  return response.data
}

/**
 * 获取模块详情
 */
export async function getModule(moduleId: string): Promise<ModuleResponse> {
  const response = await api.get(`/modules/${moduleId}`)
  return response.data
}

/**
 * 更新模块
 */
export async function updateModule(
  moduleId: string,
  data: ModuleUpdate
): Promise<ModuleResponse> {
  const response = await api.put(`/modules/${moduleId}`, data)
  return response.data
}

/**
 * 删除模块
 */
export async function deleteModule(moduleId: string): Promise<void> {
  await api.delete(`/modules/${moduleId}`)
}

// ==================== Module Issue Management ====================

/**
 * 将工作项添加到模块
 */
export async function addIssueToModule(
  moduleId: string,
  issueId: string
): Promise<{ module_id: string; issue_id: string; action: string }> {
  const response = await api.post(`/modules/${moduleId}/issues?issue_id=${issueId}`)
  return response.data
}

/**
 * 从模块移除工作项
 */
export async function removeIssueFromModule(
  moduleId: string,
  issueId: string
): Promise<{ module_id: string; issue_id: string; action: string }> {
  const response = await api.delete(`/modules/${moduleId}/issues/${issueId}`)
  return response.data
}

/**
 * 获取模块内的工作项
 */
export async function getModuleIssues(
  moduleId: string,
  options?: {
    state_id?: string
    priority?: string
    limit?: number
    offset?: number
  }
): Promise<any[]> {
  const params = new URLSearchParams()
  
  if (options?.state_id) params.append('state_id', options.state_id)
  if (options?.priority) params.append('priority', options.priority)
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  
  const response = await api.get(`/modules/${moduleId}/issues?${params.toString()}`)
  return response.data
}

// ==================== Module Progress & Statistics ====================

/**
 * 获取模块进度
 */
export async function getModuleProgress(moduleId: string): Promise<ModuleProgress> {
  const response = await api.get(`/modules/${moduleId}/progress`)
  return response.data
}

/**
 * 获取模块详细统计
 */
export async function getModuleStatistics(moduleId: string): Promise<ModuleStatistics> {
  const response = await api.get(`/modules/${moduleId}/statistics`)
  return response.data
}

// ==================== Module Tree ====================

/**
 * 获取模块树形结构
 */
export async function getModuleTree(projectId: string): Promise<ModuleTreeNode[]> {
  const response = await api.get(`/modules/tree?project_id=${projectId}`)
  return response.data
}

// ==================== Export all ====================

export const moduleApi = {
  // CRUD
  createModule,
  listModules,
  getModule,
  updateModule,
  deleteModule,
  
  // Issues
  addIssueToModule,
  removeIssueFromModule,
  getModuleIssues,
  
  // Statistics
  getModuleProgress,
  getModuleStatistics,
  
  // Tree
  getModuleTree
}

export default moduleApi