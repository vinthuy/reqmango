import api from '@/api'

export interface AutopilotTask {
  id: number
  workspace_id: number
  name: string
  description: string
  trigger_type: 'cron' | 'webhook' | 'manual'
  cron_expression: string
  trigger_url: string
  task_type: string
  enabled: boolean
  status: string
  last_run_at: string | null
  next_run_at: string | null
  created_at: string
  updated_at: string
}

export interface AutopilotExecution {
  id: number
  task_id: number
  status: 'pending' | 'running' | 'completed' | 'failed'
  started_at: string | null
  completed_at: string | null
  error_message: string | null
  result: Record<string, any> | null
  created_at: string
}

export interface AutopilotCreateRequest {
  name: string
  description?: string
  trigger_type: 'cron' | 'webhook' | 'manual'
  cron_expression?: string
  is_enabled?: boolean
  payload?: Record<string, any>
  agent_id?: number | null
  skill_ids?: number[]
  task_type?: string
}

export interface AutopilotUpdateRequest {
  name?: string
  description?: string
  trigger_type?: 'cron' | 'webhook' | 'manual'
  cron_expression?: string
  is_enabled?: boolean
  payload?: Record<string, any>
  agent_id?: number | null
  skill_ids?: number[]
}

export const autopilotApi = {
  async list(workspaceId: number): Promise<AutopilotTask[]> {
    const response = await api.get(`/workspaces/${workspaceId}/autopilot-tasks`)
    return response.data
  },
  
  async get(workspaceId: number, taskId: number): Promise<AutopilotTask> {
    const response = await api.get(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}`)
    return response.data
  },
  
  async create(workspaceId: number, data: AutopilotCreateRequest): Promise<AutopilotTask> {
    const response = await api.post(`/workspaces/${workspaceId}/autopilot-tasks`, data)
    return response.data
  },
  
  async update(workspaceId: number, taskId: number, data: AutopilotUpdateRequest): Promise<AutopilotTask> {
    const response = await api.put(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}`, data)
    return response.data
  },
  
  async delete(workspaceId: number, taskId: number): Promise<void> {
    await api.delete(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}`)
  },
  
  async toggle(workspaceId: number, taskId: number): Promise<AutopilotTask> {
    const response = await api.post(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}/toggle`)
    return response.data
  },
  
  async execute(workspaceId: number, taskId: number): Promise<AutopilotExecution> {
    const response = await api.post(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}/execute`)
    return response.data
  },
  
  async listExecutions(workspaceId: number, taskId: number): Promise<AutopilotExecution[]> {
    const response = await api.get(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}/executions`)
    return response.data
  },
  
  async getExecution(workspaceId: number, taskId: number, executionId: number): Promise<AutopilotExecution> {
    const response = await api.get(`/workspaces/${workspaceId}/autopilot-tasks/${taskId}/executions/${executionId}`)
    return response.data
  }
}
