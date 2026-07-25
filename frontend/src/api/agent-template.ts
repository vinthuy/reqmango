import api from './index'

export interface AgentTemplateCreate {
  name: string
  description?: string
  icon?: string
  system_prompt: string
  available_skills?: string[]
  available_tools?: string[]
  default_config?: any
  version?: string
}

export interface AgentTemplateUpdate {
  name?: string
  description?: string
  icon?: string
  system_prompt?: string
  available_skills?: string[]
  available_tools?: string[]
  default_config?: any
  version?: string
  status?: string
}

export interface AgentTemplateResponse {
  id: number
  name: string
  description?: string
  is_preset: boolean
  icon: string
  system_prompt: string
  available_skills: string[]
  available_tools: string[]
  default_config: any
  version: string
  status: string
  workspace_id?: number
  created_at: string
  updated_at: string
}

export const agentTemplateApi = {
  list(workspaceId: number): Promise<AgentTemplateResponse[]> {
    return api.get(`/workspaces/${workspaceId}/agent-templates`).then(res => res.data)
  },

  create(workspaceId: number, data: AgentTemplateCreate): Promise<AgentTemplateResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-templates`, data).then(res => res.data)
  },

  get(workspaceId: number, templateId: number): Promise<AgentTemplateResponse> {
    return api.get(`/workspaces/${workspaceId}/agent-templates/${templateId}`).then(res => res.data)
  },

  update(workspaceId: number, templateId: number, data: AgentTemplateUpdate): Promise<AgentTemplateResponse> {
    return api.put(`/workspaces/${workspaceId}/agent-templates/${templateId}`, data).then(res => res.data)
  },

  delete(workspaceId: number, templateId: number): Promise<{ message: string }> {
    return api.delete(`/workspaces/${workspaceId}/agent-templates/${templateId}`).then(res => res.data)
  }
}
