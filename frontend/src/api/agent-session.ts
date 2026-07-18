import api from './index'

export interface AgentSession {
  id: string
  workspace_id: number
  agent_type: 'loop_iteration' | 'pipeline_stage' | 'standalone_dispatch'
  agent_ref?: string
  status: 'running' | 'completed' | 'failed'
  model_used?: string
  input_summary?: string
  output_summary?: string
  tokens_input: number
  tokens_output: number
  cost_usd: number
  tools_called?: { tool_name: string; count: number }[]
  error_message?: string
  started_at: string
  completed_at?: string
}

export const sessionApi = {
  list(workspaceId: number, params?: { agent_type?: string; status?: string; limit?: number }): Promise<AgentSession[]> {
    return api.get(`/workspaces/${workspaceId}/agent-sessions`, { params }).then(r => r.data)
  },

  get(workspaceId: number, sessionId: string): Promise<AgentSession> {
    return api.get(`/workspaces/${workspaceId}/agent-sessions/${sessionId}`).then(r => r.data)
  },
}
