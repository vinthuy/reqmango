import api from './index'

export interface RuntimeCreate {
  name: string
  runtime_type: string
  runtime_mode?: string
  endpoint?: string
  capacity?: number
  metadata?: any
}

export interface RuntimeUpdate {
  name?: string
  runtime_type?: string
  runtime_mode?: string
  endpoint?: string
  capacity?: number
  metadata?: any
}

export interface RuntimeHeartbeat {
  version: string
  host_info?: any
  current_load: number
}

export interface RuntimeResponse {
  id: number
  name: string
  runtime_type: string
  runtime_mode: string
  endpoint?: string
  status: string
  capacity: number
  current_load: number
  version?: string
  host_info?: any
  last_heartbeat?: string
  metadata?: any
  workspace_id: number
  created_at: string
  updated_at: string
}

export const runtimeApi = {
  list(workspaceId: number): Promise<RuntimeResponse[]> {
    return api.get(`/workspaces/${workspaceId}/runtimes`).then(res => res.data)
  },

  create(workspaceId: number, data: RuntimeCreate): Promise<RuntimeResponse> {
    return api.post(`/workspaces/${workspaceId}/runtimes`, data).then(res => res.data)
  },

  get(workspaceId: number, runtimeId: number): Promise<RuntimeResponse> {
    return api.get(`/workspaces/${workspaceId}/runtimes/${runtimeId}`).then(res => res.data)
  },

  update(workspaceId: number, runtimeId: number, data: RuntimeUpdate): Promise<RuntimeResponse> {
    return api.put(`/workspaces/${workspaceId}/runtimes/${runtimeId}`, data).then(res => res.data)
  },

  delete(workspaceId: number, runtimeId: number): Promise<{ message: string }> {
    return api.delete(`/workspaces/${workspaceId}/runtimes/${runtimeId}`).then(res => res.data)
  },

  register(workspaceId: number, data: RuntimeCreate): Promise<RuntimeResponse> {
    return api.post(`/workspaces/${workspaceId}/runtimes/register`, data).then(res => res.data)
  },

  heartbeat(workspaceId: number, runtimeId: number, data: RuntimeHeartbeat): Promise<RuntimeResponse> {
    return api.post(`/workspaces/${workspaceId}/runtimes/${runtimeId}/heartbeat`, data).then(res => res.data)
  },

  findAvailable(workspaceId: number): Promise<RuntimeResponse> {
    return api.get(`/workspaces/${workspaceId}/runtimes/available`).then(res => res.data)
  }
}
