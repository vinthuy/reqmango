import api from '@/api'

// 与后端 dto/response/agent.go:AutopilotTaskResponse 对齐
export interface AutopilotTask {
  id: number
  workspace_id: number
  project_id: number | null
  name: string
  description: string
  trigger_type: 'cron' | 'webhook' | 'manual'
  cron_expression: string
  trigger_url: string
  task_type: string
  agent_template_id: number | null
  agent_config_id: number | null
  input_data: Record<string, any> | null
  status: string
  last_run_at: string | null
  next_run_at: string | null
  config: Record<string, any> | null
  notification_config: Record<string, any> | null
  timeout_seconds: number
  retry_count: number
  enabled: boolean
  created_at: string
  updated_at: string
}

// 与后端 dto/response/agent.go:AutopilotExecutionResponse 对齐
export interface AutopilotExecution {
  id: number
  task_id: number
  status: 'pending' | 'running' | 'completed' | 'failed'
  trigger_type: string
  input_data: Record<string, any> | null
  output_data: Record<string, any> | null
  error_info: string
  logs: Record<string, any> | null
  started_at: string | null
  completed_at: string | null
  failed_at: string | null
  duration_ms: number
  retry_count: number
  created_at: string
}

// 与后端 dto/request/agent.go:AutopilotTaskCreate 对齐
export interface AutopilotCreateRequest {
  name: string
  description?: string
  trigger_type: 'cron' | 'webhook' | 'manual'
  cron_expression?: string
  task_type: string
  agent_template_id?: number | null
  agent_config_id?: number | null
  project_id?: number | null
  input_data?: Record<string, any>
  config?: Record<string, any>
  notification_config?: Record<string, any>
  timeout_seconds?: number
  retry_count?: number
  enabled?: boolean
}

// 与后端 dto/request/agent.go:AutopilotTaskUpdate 对齐
export interface AutopilotUpdateRequest {
  name?: string
  description?: string
  cron_expression?: string
  input_data?: Record<string, any>
  config?: Record<string, any>
  enabled?: boolean
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
