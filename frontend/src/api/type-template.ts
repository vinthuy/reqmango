/**
 * Type Template API — 类型模板管理
 */
import api from './index'

export interface TypeTemplateCreate {
  name: string
  description?: string
  issue_type_id: number
  template_data?: Record<string, any>
}

export interface TypeTemplate {
  id: number
  name: string
  description: string
  issue_type_id: number
  template_data: Record<string, any>
  project_id: number
  created_at: string
}

export const typeTemplateApi = {
  list: async (projectId: number): Promise<TypeTemplate[]> => {
    const res = await api.get(`/projects/${projectId}/type-templates`)
    return res.data
  },

  create: async (projectId: number, data: TypeTemplateCreate): Promise<TypeTemplate> => {
    const res = await api.post(`/projects/${projectId}/type-templates`, data)
    return res.data
  },

  update: async (projectId: number, templateId: number, data: Partial<TypeTemplateCreate>): Promise<TypeTemplate> => {
    const res = await api.put(`/projects/${projectId}/type-templates/${templateId}`, data)
    return res.data
  },

  remove: async (projectId: number, templateId: number): Promise<void> => {
    await api.delete(`/projects/${projectId}/type-templates/${templateId}`)
  },
}
