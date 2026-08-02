import apiClient from './index'

export interface AgentMember {
  id: number
  project_id: number
  agent_id: number
  agent_name: string
  agent_type: string
  avatar: string
  role: string
  is_active: boolean
}

export interface CreateAgentMemberRequest {
  agent_id: number
  role?: string
}

export interface UpdateAgentMemberRequest {
  role: string
}

export const agentMemberApi = {
  list(projectId: number) {
    return apiClient.get<{ data: AgentMember[] }>(`/projects/${projectId}/agent-members`)
  },

  add(projectId: number, data: CreateAgentMemberRequest) {
    return apiClient.post<AgentMember>(`/projects/${projectId}/agent-members`, data)
  },

  updateRole(projectId: number, agentId: number, data: UpdateAgentMemberRequest) {
    return apiClient.patch(`/projects/${projectId}/agent-members/${agentId}`, data)
  },

  remove(projectId: number, agentId: number) {
    return apiClient.delete(`/projects/${projectId}/agent-members/${agentId}`)
  }
}
