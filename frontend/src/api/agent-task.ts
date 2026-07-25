import api from './index'

// New API model (matches backend)
export interface AgentTaskCreate {
  title?: string
  description?: string
  priority?: string
  task_type?: string
  input_data?: any
  agent_template_id?: number
  agent_config_id?: number
  project_id?: number
  issue_id?: number
  estimated_time?: number

  // Backward-compatible aliases for older views
  name?: string
  template_id?: number
  config_id?: number
  runtime_id?: number
  input_params?: any
  timeout?: number
}

export interface AgentTaskUpdate {
  title?: string
  description?: string
  status?: string
  priority?: string
  progress?: number
  output_data?: any
  error_info?: string
}

export interface AgentTaskResponse {
  id: number
  title?: string
  description?: string
  status: string
  priority: string
  progress?: number
  task_type?: string
  input_data?: any
  output_data?: any
  error_info?: string
  logs?: any
  agent_template_id?: number
  agent_config_id?: number
  runtime_id?: number
  workspace_id: number
  project_id?: number
  issue_id?: number
  enqueued_at: string
  claimed_at?: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
  estimated_time?: number
  actual_time?: number
  created_at: string
  updated_at: string

  // Backward-compatible aliases for older views
  name?: string
  template_name?: string
  config_name?: string
  input_params?: any
  output_result?: any
  error_message?: string
  duration_ms?: number
  timeout?: number
  execution_log?: string
}

export const agentTaskApi = {
  list(workspaceId: number, status?: string): Promise<AgentTaskResponse[]> {
    const params = status ? { status } : {}
    return api.get(`/workspaces/${workspaceId}/agent-tasks`, { params }).then(res => res.data)
  },

  create(workspaceId: number, data: AgentTaskCreate): Promise<AgentTaskResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-tasks`, data).then(res => res.data)
  },

  get(workspaceId: number, taskId: number): Promise<AgentTaskResponse> {
    return api.get(`/workspaces/${workspaceId}/agent-tasks/${taskId}`).then(res => res.data)
  },

  update(workspaceId: number, taskId: number, data: AgentTaskUpdate): Promise<AgentTaskResponse> {
    return api.put(`/workspaces/${workspaceId}/agent-tasks/${taskId}`, data).then(res => res.data)
  },

  delete(workspaceId: number, taskId: number): Promise<{ message: string }> {
    return api.delete(`/workspaces/${workspaceId}/agent-tasks/${taskId}`).then(res => res.data)
  },

  claim(workspaceId: number, taskId: number): Promise<AgentTaskResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-tasks/${taskId}/claim`).then(res => res.data)
  },

  start(workspaceId: number, taskId: number): Promise<AgentTaskResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-tasks/${taskId}/start`).then(res => res.data)
  },

  // Backward-compatible alias
  execute(workspaceId: number, taskId: number): Promise<AgentTaskResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-tasks/${taskId}/start`).then(res => res.data)
  },

  cancel(workspaceId: number, taskId: number): Promise<AgentTaskResponse> {
    return api.post(`/workspaces/${workspaceId}/agent-tasks/${taskId}/cancel`).then(res => res.data)
  },

  getLogs(workspaceId: number, taskId: number): Promise<any[]> {
    return api.get(`/workspaces/${workspaceId}/agent-tasks/${taskId}/logs`).then(res => res.data)
  }
}
