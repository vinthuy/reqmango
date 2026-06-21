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
 * 获取项目的工作项类型列表
 */
export async function getIssueTypes(projectId: number): Promise<IssueType[]> {
  const response = await api.get('/issue-types', { params: { project_id: projectId } })
  return response.data
}

/**
 * 获取工作项类型详情
 */
export async function getIssueType(
  projectId: number,
  typeId: number
): Promise<IssueTypeDetailResponse> {
  const response = await api.get(`/issue-types/${typeId}`, { params: { project_id: projectId } })
  return response.data
}

/**
 * 创建工作项类型
 */
export async function createIssueType(
  projectId: number,
  data: IssueTypeCreate
): Promise<IssueType> {
  const response = await api.post('/issue-types', data, { params: { project_id: projectId } })
  return response.data
}

/**
 * 更新工作项类型
 */
export async function updateIssueType(
  projectId: number,
  typeId: number,
  data: IssueTypeUpdate
): Promise<IssueType> {
  const response = await api.put(`/issue-types/${typeId}`, data, { params: { project_id: projectId } })
  return response.data
}

/**
 * 删除工作项类型
 */
export async function deleteIssueType(
  projectId: number,
  typeId: number
): Promise<void> {
  await api.delete(`/issue-types/${typeId}`, { params: { project_id: projectId } })
}

/**
 * 禁用/启用工作项类型
 */
export async function disableIssueType(
  projectId: number,
  typeId: number,
  isActive: boolean
): Promise<void> {
  await api.patch(`/issue-types/${typeId}/disable`, { is_active: isActive }, { params: { project_id: projectId } })
}

/**
 * 获取工作项类型关联的字段
 */
export async function getIssueTypeFields(
  projectId: number,
  typeId: number
): Promise<(IssueTypeField & { CustomField?: CustomField })[]> {
  const response = await api.get(`/issue-types/${typeId}/fields`, { params: { project_id: projectId } })
  return response.data
}

/**
 * 将字段关联到工作项类型
 */
export async function addFieldToIssueType(
  projectId: number,
  typeId: number,
  data: IssueTypeFieldCreate
): Promise<IssueTypeField> {
  const response = await api.post(`/issue-types/${typeId}/fields`, data, { params: { project_id: projectId } })
  return response.data
}

/**
 * 移除工作项类型的字段关联
 */
export async function removeFieldFromIssueType(
  projectId: number,
  typeId: number,
  fieldId: number
): Promise<void> {
  await api.delete(`/issue-types/${typeId}/fields/${fieldId}`, { params: { project_id: projectId } })
}
