import api from './index'

export interface LoopDef {
  goal?: string
  max_iterations?: number
  max_tokens?: number
  max_cost?: number
  max_duration_sec?: number
  trigger?: { type: string; schedule?: string; event?: string }
  actions?: string[]
  [key: string]: any
}

export interface Loop {
  id: number
  workspace_id: number
  name: string
  description?: string
  loop_def: LoopDef
  version: string
  status: 'active' | 'draft' | 'archived'
  created_by_id?: number
  created_at: string
  updated_at: string
}

export interface LoopRun {
  id: number
  loop_id: number
  status: 'running' | 'completed' | 'failed' | 'escalated' | 'stopped'
  current_iteration: number
  max_iterations: number
  goal: string
  goal_metrics?: Record<string, number>
  working_memory?: Record<string, any>
  tokens_used: number
  cost_usd: number
  stopped_reason?: string
  started_at: string
  completed_at?: string
}

export interface LoopIteration {
  id: number
  loop_run_id: number
  iteration_num: number
  action_taken: { task: string; result: string }
  result_observed: Record<string, number>
  reasoning?: string
  decision: 'continue' | 'stop' | 'escalate' | 'wait'
  tokens_used: number
  duration_ms?: number
  created_at: string
}

export interface LoopRunDetail {
  run: LoopRun
  iterations: LoopIteration[]
}

export const loopApi = {
  list(workspaceId: number): Promise<Loop[]> {
    return api.get(`/workspaces/${workspaceId}/loops`).then(r => r.data)
  },

  get(workspaceId: number, loopId: number): Promise<Loop> {
    return api.get(`/workspaces/${workspaceId}/loops/${loopId}`).then(r => r.data)
  },

  create(workspaceId: number, data: { name: string; description?: string; loop_def: LoopDef }): Promise<Loop> {
    return api.post(`/workspaces/${workspaceId}/loops`, data).then(r => r.data)
  },

  update(workspaceId: number, loopId: number, data: Partial<Loop>): Promise<Loop> {
    return api.put(`/workspaces/${workspaceId}/loops/${loopId}`, data).then(r => r.data)
  },

  delete(workspaceId: number, loopId: number): Promise<void> {
    return api.delete(`/workspaces/${workspaceId}/loops/${loopId}`)
  },

  start(workspaceId: number, loopId: number): Promise<LoopRun> {
    return api.post(`/workspaces/${workspaceId}/loops/${loopId}/start`).then(r => r.data)
  },

  stop(workspaceId: number, runId: number): Promise<void> {
    return api.post(`/workspaces/${workspaceId}/loops/runs/${runId}/stop`)
  },

  getRuns(workspaceId: number, loopId: number, limit = 20): Promise<LoopRun[]> {
    return api.get(`/workspaces/${workspaceId}/loops/${loopId}/runs`, { params: { limit } }).then(r => r.data)
  },

  getRun(workspaceId: number, runId: number): Promise<LoopRunDetail> {
    return api.get(`/workspaces/${workspaceId}/loops/runs/${runId}`).then(r => r.data)
  },
}
