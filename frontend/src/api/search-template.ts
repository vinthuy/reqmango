import api from './index'
import type { SearchTemplate, SearchTemplateCreate } from '@/types/search-template'

export async function listSearchTemplates(projectId: number): Promise<SearchTemplate[]> {
  const response = await api.get(`/projects/${projectId}/search-templates`)
  return response.data
}

export async function getSearchTemplate(projectId: number, templateId: number): Promise<SearchTemplate> {
  const response = await api.get(`/projects/${projectId}/search-templates/${templateId}`)
  return response.data
}

export async function createSearchTemplate(projectId: number, data: SearchTemplateCreate): Promise<SearchTemplate> {
  const response = await api.post(`/projects/${projectId}/search-templates`, data)
  return response.data
}

export async function deleteSearchTemplate(projectId: number, templateId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/search-templates/${templateId}`)
}

export async function applySearchTemplate(projectId: number, templateId: number): Promise<SearchTemplate> {
  const response = await api.post(`/projects/${projectId}/search-templates/${templateId}/apply`)
  return response.data
}

export default {
  listSearchTemplates,
  getSearchTemplate,
  createSearchTemplate,
  deleteSearchTemplate,
  applySearchTemplate,
}