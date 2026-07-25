import api from './index'

export interface SkillCreate {
  name: string
  description?: string
  category?: string
  parameters?: any[]
  steps?: any[]
  output_format?: string
  version?: string
  is_shared?: boolean
}

export interface SkillUpdate {
  name?: string
  description?: string
  category?: string
  parameters?: any[]
  steps?: any[]
  output_format?: string
  version?: string
  is_shared?: boolean
  status?: string
}

export interface SkillResponse {
  id: number
  name: string
  description?: string
  category: string
  parameters: any[]
  steps: any[]
  output_format: string
  version: string
  is_shared: boolean
  status: string
  usage_count: number
  workspace_id?: number
  created_at: string
  updated_at: string
}

export const skillApi = {
  list(workspaceId: number): Promise<SkillResponse[]> {
    return api.get(`/workspaces/${workspaceId}/skills`).then(res => res.data)
  },

  create(workspaceId: number, data: SkillCreate): Promise<SkillResponse> {
    return api.post(`/workspaces/${workspaceId}/skills`, data).then(res => res.data)
  },

  get(workspaceId: number, skillId: number): Promise<SkillResponse> {
    return api.get(`/workspaces/${workspaceId}/skills/${skillId}`).then(res => res.data)
  },

  update(workspaceId: number, skillId: number, data: SkillUpdate): Promise<SkillResponse> {
    return api.put(`/workspaces/${workspaceId}/skills/${skillId}`, data).then(res => res.data)
  },

  delete(workspaceId: number, skillId: number): Promise<{ message: string }> {
    return api.delete(`/workspaces/${workspaceId}/skills/${skillId}`).then(res => res.data)
  }
}
