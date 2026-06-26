/**
 * Work Item Template API - 工作项模板 API 调用模块
 */
import api from './index'
import type {
  WorkItemTemplate,
  WorkItemTemplateCreate,
  WorkItemTemplateUpdate
} from '@/types/work-item-template'

/**
 * 列出项目的工作项模板
 */
export async function listTemplates(projectId: number): Promise<WorkItemTemplate[]> {
  const response = await api.get(`/projects/${projectId}/work-item-templates`)
  return response.data
}

/**
 * 获取单个工作项模板
 */
export async function getTemplate(projectId: number, templateId: number): Promise<WorkItemTemplate> {
  const response = await api.get(`/projects/${projectId}/work-item-templates/${templateId}`)
  return response.data
}

/**
 * 创建工作项模板
 */
export async function createTemplate(
  projectId: number,
  data: WorkItemTemplateCreate
): Promise<WorkItemTemplate> {
  const response = await api.post(`/projects/${projectId}/work-item-templates`, data)
  return response.data
}

/**
 * 更新工作项模板
 */
export async function updateTemplate(
  projectId: number,
  templateId: number,
  data: WorkItemTemplateUpdate
): Promise<WorkItemTemplate> {
  const response = await api.put(`/projects/${projectId}/work-item-templates/${templateId}`, data)
  return response.data
}

/**
 * 删除工作项模板
 */
export async function deleteTemplate(projectId: number, templateId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/work-item-templates/${templateId}`)
}

export const workItemTemplateApi = {
  listTemplates,
  getTemplate,
  createTemplate,
  updateTemplate,
  deleteTemplate
}

export default workItemTemplateApi
