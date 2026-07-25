import api from './index'

export interface AgentConfigCreate {
  name: string
  description?: string
  provider: string
  model: string
  api_key: string
  api_endpoint?: string
  inference_level?: string
  service_level?: string
  max_tokens?: number
  temperature?: number
  top_p?: number
  is_default?: boolean
}

export interface AgentConfigUpdate {
  name?: string
  description?: string
  provider?: string
  model?: string
  api_key?: string
  api_endpoint?: string
  inference_level?: string
  service_level?: string
  max_tokens?: number
  temperature?: number
  top_p?: number
  is_default?: boolean
  is_active?: boolean
}

export interface AgentConfigResponse {
  id: number
  name: string
  description?: string
  provider: string
  model: string
  api_endpoint?: string
  inference_level: string
  service_level: string
  max_tokens: number
  temperature: number
  top_p: number
  is_default: boolean
  is_active: boolean
  workspace_id: number
  created_at: string
  updated_at: string
}

export const agentConfigApi = {
  list(workspaceId: number): Promise<AgentConfigResponse[]> {
    return api.get(`/workspaces/${workspaceId}/agent-configs`).then(res => res.data)
  },

  create(workspaceId: number, data: AgentConfigCreate): Promise<AgentConfigResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-configs`, data).then(res => res.data)
  },

  get(workspaceId: number, configId: number): Promise<AgentConfigResponse> {
    return api.get(`/workspaces/${workspaceId}/agent-configs/${configId}`).then(res => res.data)
  },

  update(workspaceId: number, configId: number, data: AgentConfigUpdate): Promise<AgentConfigResponse> {
    return api.put(`/workspaces/${workspaceId}/agent-configs/${configId}`, data).then(res => res.data)
  },

  delete(workspaceId: number, configId: number): Promise<{ message: string }> {
    return api.delete(`/workspaces/${workspaceId}/agent-configs/${configId}`).then(res => res.data)
  },

  getDefault(workspaceId: number): Promise<AgentConfigResponse> {
    return api.get(`/workspaces/${workspaceId}/agent-configs/default`).then(res => res.data)
  }
}
