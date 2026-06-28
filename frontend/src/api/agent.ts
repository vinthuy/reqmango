/**
 * Agent API — AI 自动化代理模块
 */
import api from './index'
import type { Agent, AgentActivity, AgentCreateRequest, AgentUpdateRequest, AgentDispatchRequest } from '@/types/agent'

export type { Agent, AgentActivity, AgentCreateRequest, AgentUpdateRequest, AgentDispatchRequest }

export const agentApi = {
  list(workspaceId: number): Promise<Agent[]> {
    return api.get(`/workspaces/${workspaceId}/agents`).then((r) => r.data)
  },

  get(workspaceId: number, agentId: number): Promise<Agent> {
    return api.get(`/workspaces/${workspaceId}/agents/${agentId}`).then((r) => r.data)
  },

  create(workspaceId: number, data: AgentCreateRequest): Promise<Agent> {
    return api.post(`/workspaces/${workspaceId}/agents`, data).then((r) => r.data)
  },

  update(workspaceId: number, agentId: number, data: AgentUpdateRequest): Promise<Agent> {
    return api.put(`/workspaces/${workspaceId}/agents/${agentId}`, data).then((r) => r.data)
  },

  delete(workspaceId: number, agentId: number): Promise<void> {
    return api.delete(`/workspaces/${workspaceId}/agents/${agentId}`)
  },

  dispatch(workspaceId: number, agentId: number, req: AgentDispatchRequest): Promise<AgentActivity> {
    return api.post(`/workspaces/${workspaceId}/agents/${agentId}/dispatch`, req).then((r) => r.data)
  },

  getActivity(workspaceId: number, agentId: number): Promise<AgentActivity[]> {
    return api.get(`/workspaces/${workspaceId}/agents/${agentId}/activity`).then((r) => r.data)
  },

  // ==================== Project-level convenience ====================

  autoTriage(projectId: number): Promise<AgentActivity> {
    return api.post(`/projects/${projectId}/agent/auto-triage`).then((r) => r.data)
  },

  autoAssign(projectId: number): Promise<AgentActivity> {
    return api.post(`/projects/${projectId}/agent/auto-assign`).then((r) => r.data)
  },
}
