/**
 * Agent API — AI 自动化代理模块
 */
import api from './index'
import type { Agent, AgentActivity, AgentCreateRequest, AgentUpdateRequest, AgentDispatchRequest } from '@/types/agent'

export type { Agent, AgentActivity, AgentCreateRequest, AgentUpdateRequest, AgentDispatchRequest }

// Module-level cache: deduplicate concurrent agent list calls and serve from cache within TTL
const _listCache = new Map<number, { promise: Promise<Agent[]>; ts: number }>()
const CACHE_TTL = 30_000

function invalidateListCache(workspaceId: number) {
  _listCache.delete(workspaceId)
}

export function invalidateAgentCache(workspaceId?: number) {
  if (workspaceId != null) {
    _listCache.delete(workspaceId)
  } else {
    _listCache.clear()
  }
}

export const agentApi = {
  list(workspaceId: number): Promise<Agent[]> {
    const cached = _listCache.get(workspaceId)
    if (cached && Date.now() - cached.ts < CACHE_TTL) {
      return cached.promise
    }
    const promise = api.get(`/workspaces/${workspaceId}/agents`).then((r) => r.data)
    _listCache.set(workspaceId, { promise, ts: Date.now() })
    return promise
  },

  get(workspaceId: number, agentId: number): Promise<Agent> {
    return api.get(`/workspaces/${workspaceId}/agents/${agentId}`).then((r) => r.data)
  },

  create(workspaceId: number, data: AgentCreateRequest): Promise<Agent> {
    invalidateListCache(workspaceId)
    return api.post(`/workspaces/${workspaceId}/agents`, data).then((r) => r.data)
  },

  update(workspaceId: number, agentId: number, data: AgentUpdateRequest): Promise<Agent> {
    invalidateListCache(workspaceId)
    return api.put(`/workspaces/${workspaceId}/agents/${agentId}`, data).then((r) => r.data)
  },

  delete(workspaceId: number, agentId: number): Promise<void> {
    invalidateListCache(workspaceId)
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

  // ==================== Workspace-level activity ====================

  listWorkspaceActivity(workspaceId: number, params?: {
    agent_id?: number
    action?: string
    limit?: number
  }): Promise<AgentActivity[]> {
    return api.get(`/workspaces/${workspaceId}/agents/activity`, { params }).then((r) => r.data)
  },

  rateActivity(workspaceId: number, activityId: number, rating: 1 | -1): Promise<void> {
    return api.patch(`/workspaces/${workspaceId}/agents/activity/${activityId}/feedback`, { rating })
  },
}
