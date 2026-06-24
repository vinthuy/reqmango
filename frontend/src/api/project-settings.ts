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
  _projectId: number,
  workspaceId: number,
  data: IssueTypeCreate
): Promise<IssueType> {
  const response = await api.post(
    `/issue-types?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出工作项类型
 */
export async function listIssueTypes(
  projectId: number,
  includeInactive?: boolean
): Promise<IssueType[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', String(projectId)) // NOTE: projectId used as workspace_id for backward compat
  if (includeInactive) params.append('include_inactive', 'true')

  const response = await api.get(
    `/issue-types?${params.toString()}`
  )
  return response.data
}

/**
 * 获取工作项类型详情
 */
export async function getIssueType(typeId: number): Promise<IssueType> {
  const response = await api.get(`/issue-types/${typeId}`)
  return response.data
}

/**
 * 更新工作项类型
 */
export async function updateIssueType(
  typeId: number,
  data: IssueTypeUpdate
): Promise<IssueType> {
  const response = await api.put(`/issue-types/${typeId}`, data)
  return response.data
}

/**
 * 删除工作项类型
 */
export async function deleteIssueType(typeId: number): Promise<void> {
  await api.delete(`/issue-types/${typeId}`)
}

/**
 * 创建默认工作项类型
 */
export async function createDefaultIssueTypes(
  _projectId: number,
  workspaceId: number
): Promise<IssueType[]> {
  const response = await api.post(
    `/issue-types?workspace_id=${workspaceId}`
  )
  return response.data
}

// ==================== State API ====================

/**
 * 创建状态
 */
export async function createState(
  projectId: number,
  workspaceId: number,
  data: StateCreate
): Promise<State> {
  const response = await api.post(
    `/projects/${projectId}/settings/states?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出状态
 */
export async function listStates(
  projectId: number,
  includeInactive?: boolean
): Promise<State[]> {
  const params = new URLSearchParams()
  if (includeInactive) params.append('include_inactive', 'true')

  const response = await api.get(
    `/projects/${projectId}/settings/states?${params.toString()}`
  )
  return response.data
}

/**
 * 获取状态详情
 */
export async function getState(projectId: number, stateId: number): Promise<State> {
  const response = await api.get(`/projects/${projectId}/settings/states/${stateId}`)
  return response.data
}

/**
 * 更新状态
 */
export async function updateState(
  projectId: number,
  stateId: number,
  data: StateUpdate
): Promise<State> {
  const response = await api.put(`/projects/${projectId}/settings/states/${stateId}`, data)
  return response.data
}

/**
 * 删除状态
 */
export async function deleteState(projectId: number, stateId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/settings/states/${stateId}`)
}

/**
 * 创建默认状态
 */
export async function createDefaultStates(
  projectId: number,
  workspaceId: number
): Promise<State[]> {
  const response = await api.post(
    `/projects/${projectId}/settings/states/default?workspace_id=${workspaceId}`
  )
  return response.data
}

// ==================== Label API ====================

/**
 * 创建标签
 */
export async function createLabel(
  projectId: number,
  workspaceId: number,
  data: LabelCreate
): Promise<Label> {
  const response = await api.post(
    `/projects/${projectId}/settings/labels?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出标签
 */
export async function listLabels(projectId: number): Promise<Label[]> {
  const response = await api.get(`/projects/${projectId}/settings/labels`)
  return response.data
}

/**
 * 获取标签详情
 */
export async function getLabel(projectId: number, labelId: number): Promise<Label> {
  const response = await api.get(`/projects/${projectId}/settings/labels/${labelId}`)
  return response.data
}

/**
 * 更新标签
 */
export async function updateLabel(
  projectId: number,
  labelId: number,
  data: LabelUpdate
): Promise<Label> {
  const response = await api.put(`/projects/${projectId}/settings/labels/${labelId}`, data)
  return response.data
}

/**
 * 删除标签
 */
export async function deleteLabel(projectId: number, labelId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/settings/labels/${labelId}`)
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