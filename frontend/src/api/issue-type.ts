/**
 * Issue Type API - 工作项类型 API
 */
import api from './index'
import type {
  IssueType,
  IssueTypeCreate,
  IssueTypeUpdate,
  IssueTypeField,
  IssueTypeFieldCreate
} from '@/types/issue-type'
import type { CustomField } from '@/types/custom-field'

export interface IssueTypeWithFields extends IssueType {
  fields: (IssueTypeField & { CustomField?: CustomField })[]
}

export interface IssueTypeDetailResponse {
  issue_type: IssueType
  fields: IssueTypeField[]
}

/**
 * 获取工作项类型列表（workspace 范围）
 */
export async function getIssueTypes(workspaceId: number, projectId?: number): Promise<IssueType[]> {
  const params: Record<string, string | number> = { workspace_id: workspaceId }
  if (projectId) params.project_id = projectId
  const response = await api.get('/issue-types', { params })
  return response.data
}

/**
 * 获取工作项类型详情
 */
export async function getIssueType(typeId: number): Promise<IssueTypeDetailResponse> {
  const response = await api.get(`/issue-types/${typeId}`)
  return response.data
}

/**
 * 创建工作项类型
 */
export async function createIssueType(workspaceId: number, data: IssueTypeCreate): Promise<IssueType> {
  const response = await api.post('/issue-types', data, { params: { workspace_id: workspaceId } })
  return response.data
}

/**
 * 更新工作项类型
 */
export async function updateIssueType(typeId: number, data: IssueTypeUpdate): Promise<IssueType> {
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
 * 禁用/启用工作项类型
 */
export async function disableIssueType(typeId: number, isActive: boolean): Promise<void> {
  await api.patch(`/issue-types/${typeId}/disable`, { is_active: isActive })
}

/**
 * 获取工作项类型关联的字段
 */
export async function getIssueTypeFields(typeId: number): Promise<(IssueTypeField & { CustomField?: CustomField })[]> {
  const response = await api.get(`/issue-types/${typeId}/fields`)
  return response.data
}

/**
 * 将字段关联到工作项类型
 */
export async function addFieldToIssueType(typeId: number, data: IssueTypeFieldCreate): Promise<IssueTypeField> {
  const response = await api.post(`/issue-types/${typeId}/fields`, data)
  return response.data
}

/**
 * 更新类型-字段关联（切换必填等）
 */
export async function updateIssueTypeField(typeId: number, fieldId: number, data: { is_required?: boolean; sequence?: number }): Promise<IssueTypeField> {
  const response = await api.put(`/issue-types/${typeId}/fields/${fieldId}`, data)
  return response.data
}

/**
 * 移除工作项类型的字段关联
 */
export async function removeFieldFromIssueType(typeId: number, fieldId: number): Promise<void> {
  await api.delete(`/issue-types/${typeId}/fields/${fieldId}`)
}

const issueTypeApi = {
  getIssueTypes,
  getIssueType,
  createIssueType,
  updateIssueType,
  deleteIssueType,
  disableIssueType,
  getIssueTypeFields,
  addFieldToIssueType,
  updateIssueTypeField,
  removeFieldFromIssueType,
}

export default issueTypeApi
