import api from './index'
import type { WorkItemTemplate, WorkItemTemplateCreate, WorkItemTemplateUpdate } from '../types/work-item-template'

export async function listWorkItemTemplates(projectId: number): Promise<WorkItemTemplate[]> {
  const response = await api.get(`/projects/${projectId}/work-item-templates`)
  return response.data
}

export async function getWorkItemTemplate(projectId: number, templateId: number): Promise<WorkItemTemplate> {
  const response = await api.get(`/projects/${projectId}/work-item-templates/${templateId}`)
  return response.data
}

export async function createWorkItemTemplate(
  projectId: number,
  data: WorkItemTemplateCreate
): Promise<WorkItemTemplate> {
  const response = await api.post(`/projects/${projectId}/work-item-templates`, data)
  return response.data
}

export async function updateWorkItemTemplate(
  projectId: number,
  templateId: number,
  data: WorkItemTemplateUpdate
): Promise<WorkItemTemplate> {
  const response = await api.put(`/projects/${projectId}/work-item-templates/${templateId}`, data)
  return response.data
}

export async function deleteWorkItemTemplate(projectId: number, templateId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/work-item-templates/${templateId}`)
}
