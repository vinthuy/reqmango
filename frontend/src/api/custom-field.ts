/**
 * Custom Field API - 自定义字段 API 调用模块
 */
import api from './index'
import type {
  CustomField,
  CustomFieldCreate,
  CustomFieldUpdate,
  CustomFieldOption,
  CustomFieldOptionCreate,
  IssueCustomFieldValue,
  IssueCustomFieldValueCreate,
  IssueCustomFieldValueUpdate,
  BulkCustomFieldValueUpdate,
  IssueCustomFieldsResponse
} from '@/types/custom-field'

// ==================== Custom Field CRUD ====================

/**
 * 创建自定义字段
 */
export async function createCustomField(
  workspaceId: number,
  data: CustomFieldCreate
): Promise<CustomField> {
  const response = await api.post(`/custom-fields?workspace_id=${workspaceId}`, data)
  return response.data
}

/**
 * 列出自定义字段
 */
export async function listCustomFields(
  workspaceId: number,
  projectId?: number,
  issueTypeId?: number,
  includeInactive?: boolean
): Promise<CustomField[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', workspaceId.toString())
  if (projectId) params.append('project_id', projectId.toString())
  if (issueTypeId) params.append('issue_type_id', issueTypeId.toString())
  if (includeInactive) params.append('include_inactive', 'true')
  
  const response = await api.get(`/custom-fields?${params.toString()}`)
  return response.data
}

/**
 * 获取自定义字段详情
 */
export async function getCustomField(fieldId: number): Promise<CustomField> {
  const response = await api.get(`/custom-fields/${fieldId}`)
  return response.data
}

/**
 * 更新自定义字段
 */
export async function updateCustomField(
  fieldId: number,
  data: CustomFieldUpdate
): Promise<CustomField> {
  const response = await api.put(`/custom-fields/${fieldId}`, data)
  return response.data
}

/**
 * 删除自定义字段
 */
export async function deleteCustomField(fieldId: number): Promise<void> {
  await api.delete(`/custom-fields/${fieldId}`)
}

// ==================== Custom Field Options ====================

/**
 * 创建字段选项
 */
export async function createFieldOption(
  fieldId: number,
  data: CustomFieldOptionCreate
): Promise<CustomFieldOption> {
  const response = await api.post(`/custom-fields/${fieldId}/options`, data)
  return response.data
}

/**
 * 更新字段选项
 */
export async function updateFieldOption(
  fieldId: number,
  optionId: number,
  data: {
    value?: string
    color?: string
    sequence?: number
    is_default?: boolean
    is_active?: boolean
  }
): Promise<CustomFieldOption> {
  const params = new URLSearchParams()
  if (data.value) params.append('value', data.value)
  if (data.color) params.append('color', data.color)
  if (data.sequence !== undefined) params.append('sequence', data.sequence.toString())
  if (data.is_default !== undefined) params.append('is_default', data.is_default.toString())
  if (data.is_active !== undefined) params.append('is_active', data.is_active.toString())
  
  const response = await api.put(`/custom-fields/${fieldId}/options/${optionId}?${params.toString()}`)
  return response.data
}

/**
 * 删除字段选项
 */
export async function deleteFieldOption(
  fieldId: number,
  optionId: number
): Promise<void> {
  await api.delete(`/custom-fields/${fieldId}/options/${optionId}`)
}

// ==================== Issue Custom Field Values ====================

/**
 * 设置工作项的自定义字段值
 */
export async function setIssueCustomFieldValue(
  issueId: number,
  data: IssueCustomFieldValueCreate
): Promise<IssueCustomFieldValue> {
  const response = await api.post(`/custom-fields/issues/${issueId}/values`, data)
  return response.data
}

/**
 * 获取工作项的所有自定义字段值
 */
export async function listIssueCustomFieldValues(
  issueId: number
): Promise<IssueCustomFieldValue[]> {
  const response = await api.get(`/custom-fields/issues/${issueId}/values`)
  return response.data
}

/**
 * 更新工作项的特定自定义字段值
 */
export async function updateIssueCustomFieldValue(
  issueId: number,
  fieldId: number,
  data: IssueCustomFieldValueUpdate
): Promise<IssueCustomFieldValue> {
  const response = await api.put(`/custom-fields/issues/${issueId}/values/${fieldId}`, data)
  return response.data
}

/**
 * 删除工作项的自定义字段值
 */
export async function deleteIssueCustomFieldValue(
  issueId: number,
  fieldId: number
): Promise<void> {
  await api.delete(`/custom-fields/issues/${issueId}/values/${fieldId}`)
}

/**
 * 批量更新工作项的自定义字段值
 */
export async function bulkUpdateIssueCustomFieldValues(
  issueId: number,
  values: IssueCustomFieldValueUpdate[]
): Promise<IssueCustomFieldValue[]> {
  const data: BulkCustomFieldValueUpdate = {
    issue_id: issueId,
    values
  }
  const response = await api.post(`/custom-fields/issues/${issueId}/values/bulk`, data)
  return response.data
}

/**
 * 获取工作项的所有自定义字段定义及其值
 */
export async function getIssueCustomFieldsWithDefinitions(
  issueId: number
): Promise<IssueCustomFieldsResponse> {
  const response = await api.get(`/custom-fields/issues/${issueId}/fields`)
  return response.data
}

// ==================== Export all ====================

export const customFieldApi = {
  // Field CRUD
  createCustomField,
  listCustomFields,
  getCustomField,
  updateCustomField,
  deleteCustomField,
  
  // Options
  createFieldOption,
  updateFieldOption,
  deleteFieldOption,
  
  // Values
  setIssueCustomFieldValue,
  listIssueCustomFieldValues,
  updateIssueCustomFieldValue,
  deleteIssueCustomFieldValue,
  bulkUpdateIssueCustomFieldValues,
  getIssueCustomFieldsWithDefinitions
}

export default customFieldApi