/**
 * Automation API — 自动化规则管理
 */
import api from './index'

export interface AutomationRule {
  id: number
  name: string
  description: string
  project_id: number
  trigger_type: string
  conditions: string
  actions: string
  is_enabled: boolean
  sequence: number
  execution_count: number
  created_at: string
  updated_at: string
}

export interface AutomationCreate {
  name: string
  description?: string
  trigger_type: string
  conditions?: string
  actions: string
  is_enabled?: boolean
  sequence?: number
}

export const automationApi = {
  list: async (projectId: number): Promise<AutomationRule[]> => {
    const res = await api.get(`/projects/${projectId}/automations`)
    return res.data
  },

  get: async (projectId: number, id: number): Promise<AutomationRule> => {
    const res = await api.get(`/projects/${projectId}/automations/${id}`)
    return res.data
  },

  create: async (projectId: number, data: AutomationCreate): Promise<AutomationRule> => {
    const res = await api.post(`/projects/${projectId}/automations`, data)
    return res.data
  },

  update: async (projectId: number, id: number, data: Partial<AutomationCreate>): Promise<AutomationRule> => {
    const res = await api.put(`/projects/${projectId}/automations/${id}`, data)
    return res.data
  },

  delete: async (projectId: number, id: number): Promise<void> => {
    await api.delete(`/projects/${projectId}/automations/${id}`)
  },

  execute: async (projectId: number, id: number, issueId: number, context?: Record<string, any>): Promise<any> => {
    const res = await api.post(`/projects/${projectId}/automations/${id}/execute`, {
      issue_id: issueId,
      context: context || {},
    })
    return res.data
  },
}
