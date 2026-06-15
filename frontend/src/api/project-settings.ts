/**
 * Project Settings API - 项目设置 API 调用模块（IssueType、State、Label）
 */
import api from './index'
import type {
  IssueType,
  IssueTypeCreate,
  IssueTypeUpdate,
  State,
  StateCreate,
  StateUpdate,
  Label,
  LabelCreate,
  LabelUpdate
} from '@/types/project-settings'

// ==================== IssueType API ====================

/**
 * 创建工作项类型
 */
export async function createIssueType(
  projectId: string,
  workspaceId: string,
  data: IssueTypeCreate
): Promise<IssueType> {
  const response = await api.post(
    `/projects/${projectId}/issue-types?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出工作项类型
 */
export async function listIssueTypes(
  projectId: string,
  includeInactive?: boolean
): Promise<IssueType[]> {
  const params = new URLSearchParams()
  if (includeInactive) params.append('include_inactive', 'true')
  
  const response = await api.get(
    `/projects/${projectId}/issue-types?${params.toString()}`
  )
  return response.data
}

/**
 * 获取工作项类型详情
 */
export async function getIssueType(typeId: string): Promise<IssueType> {
  const response = await api.get(`/projects/issue-types/${typeId}`)
  return response.data
}

/**
 * 更新工作项类型
 */
export async function updateIssueType(
  typeId: string,
  data: IssueTypeUpdate
): Promise<IssueType> {
  const response = await api.put(`/projects/issue-types/${typeId}`, data)
  return response.data
}

/**
 * 删除工作项类型
 */
export async function deleteIssueType(typeId: string): Promise<void> {
  await api.delete(`/projects/issue-types/${typeId}`)
}

/**
 * 创建默认工作项类型
 */
export async function createDefaultIssueTypes(
  projectId: string,
  workspaceId: string
): Promise<IssueType[]> {
  const response = await api.post(
    `/projects/${projectId}/issue-types/default?workspace_id=${workspaceId}`
  )
  return response.data
}

// ==================== State API ====================

/**
 * 创建状态
 */
export async function createState(
  projectId: string,
  workspaceId: string,
  data: StateCreate
): Promise<State> {
  const response = await api.post(
    `/projects/${projectId}/states?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出状态
 */
export async function listStates(
  projectId: string,
  includeInactive?: boolean
): Promise<State[]> {
  const params = new URLSearchParams()
  if (includeInactive) params.append('include_inactive', 'true')
  
  const response = await api.get(
    `/projects/${projectId}/states?${params.toString()}`
  )
  return response.data
}

/**
 * 获取状态详情
 */
export async function getState(stateId: string): Promise<State> {
  const response = await api.get(`/projects/states/${stateId}`)
  return response.data
}

/**
 * 更新状态
 */
export async function updateState(
  stateId: string,
  data: StateUpdate
): Promise<State> {
  const response = await api.put(`/projects/states/${stateId}`, data)
  return response.data
}

/**
 * 删除状态
 */
export async function deleteState(stateId: string): Promise<void> {
  await api.delete(`/projects/states/${stateId}`)
}

/**
 * 创建默认状态
 */
export async function createDefaultStates(
  projectId: string,
  workspaceId: string
): Promise<State[]> {
  const response = await api.post(
    `/projects/${projectId}/states/default?workspace_id=${workspaceId}`
  )
  return response.data
}

// ==================== Label API ====================

/**
 * 创建标签
 */
export async function createLabel(
  projectId: string,
  workspaceId: string,
  data: LabelCreate
): Promise<Label> {
  const response = await api.post(
    `/projects/${projectId}/labels?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出标签
 */
export async function listLabels(projectId: string): Promise<Label[]> {
  const response = await api.get(`/projects/${projectId}/labels`)
  return response.data
}

/**
 * 获取标签详情
 */
export async function getLabel(labelId: string): Promise<Label> {
  const response = await api.get(`/projects/labels/${labelId}`)
  return response.data
}

/**
 * 更新标签
 */
export async function updateLabel(
  labelId: string,
  data: LabelUpdate
): Promise<Label> {
  const response = await api.put(`/projects/labels/${labelId}`, data)
  return response.data
}

/**
 * 删除标签
 */
export async function deleteLabel(labelId: string): Promise<void> {
  await api.delete(`/projects/labels/${labelId}`)
}

// ==================== Export ====================

export const projectSettingsApi = {
  // IssueType
  createIssueType,
  listIssueTypes,
  getIssueType,
  updateIssueType,
  deleteIssueType,
  createDefaultIssueTypes,
  
  // State
  createState,
  listStates,
  getState,
  updateState,
  deleteState,
  createDefaultStates,
  
  // Label
  createLabel,
  listLabels,
  getLabel,
  updateLabel,
  deleteLabel
}

export default projectSettingsApi