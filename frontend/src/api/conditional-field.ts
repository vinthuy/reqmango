/**
 * Conditional Field API — 条件字段管理
 */
import api from './index'

export interface ConditionalField {
  id: number
  name: string
  field_type: string
  condition_field_id: number
  condition_value: string
  project_id: number
}

export interface ConditionalFieldCreate {
  name: string
  field_type: string
  condition_field_id: number
  condition_value: string
}

export const conditionalFieldApi = {
  list: async (projectId: number): Promise<ConditionalField[]> => {
    const res = await api.get(`/projects/${projectId}/conditional-fields`)
    return res.data
  },

  create: async (projectId: number, data: ConditionalFieldCreate): Promise<ConditionalField> => {
    const res = await api.post(`/projects/${projectId}/conditional-fields`, data)
    return res.data
  },

  update: async (projectId: number, fieldId: number, data: Partial<ConditionalFieldCreate>): Promise<ConditionalField> => {
    const res = await api.put(`/projects/${projectId}/conditional-fields/${fieldId}`, data)
    return res.data
  },

  remove: async (projectId: number, fieldId: number): Promise<void> => {
    await api.delete(`/projects/${projectId}/conditional-fields/${fieldId}`)
  },
}
