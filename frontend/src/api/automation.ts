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

export interface AutomationExecution {
  id: number
  rule_id: number
  issue_id: number
  trigger_type: string
  context_json: string
  actions_taken: string
  status: string
  error: string
  duration: number
  executed_at: string
}

export interface AutomationExecutionResponse {
  data: AutomationExecution[]
  total: number
  limit: number
  offset: number
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

  getExecutionHistory: async (issueId: number, limit?: number): Promise<AutomationExecution[]> => {
    const params = new URLSearchParams()
    if (limit) params.append('limit', limit.toString())
    const res = await api.get(`/issues/${issueId}/automation-history?${params.toString()}`)
    return res.data
  },

  getRuleExecutionHistory: async (ruleId: number, params?: { limit?: number; offset?: number; startTime?: string; endTime?: string }): Promise<AutomationExecutionResponse> => {
    const query = new URLSearchParams()
    if (params?.limit) query.append('limit', params.limit.toString())
    if (params?.offset) query.append('offset', params.offset.toString())
    if (params?.startTime) query.append('start_time', params.startTime)
    if (params?.endTime) query.append('end_time', params.endTime)
    const res = await api.get(`/automations/${ruleId}/execution-history?${query.toString()}`)
    return res.data
  },

  getProjectExecutionHistory: async (projectId: number, params?: { limit?: number; offset?: number; startTime?: string; endTime?: string }): Promise<AutomationExecutionResponse> => {
    const query = new URLSearchParams()
    if (params?.limit) query.append('limit', params.limit.toString())
    if (params?.offset) query.append('offset', params.offset.toString())
    if (params?.startTime) query.append('start_time', params.startTime)
    if (params?.endTime) query.append('end_time', params.endTime)
    const res = await api.get(`/projects/${projectId}/automation-executions?${query.toString()}`)
    return res.data
  },
}
