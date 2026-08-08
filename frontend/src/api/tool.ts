/**
 * Tool Calling API - Tool CRUD, call logs, permissions, MCP sync
 */
import api from './index'

export interface Tool {
  id: number
  name: string
  description: string
  category: string
  is_builtin: boolean
  status: string
  tool_type: 'api' | 'function' | 'mcp' | 'workflow'
  mcp_config_id?: number
  endpoint?: string
  method?: string
  auth_type?: string
  rate_limit: number
  timeout: number
  workspace_id?: number
  created_at: string
  updated_at: string
}

export interface ToolCallLog {
  id: number
  workspace_id: number
  agent_task_id?: number
  tool_id: number
  tool_name: string
  agent_id?: number
  caller_user_id?: number
  input_params: any
  output_result: any
  status: 'success' | 'failed' | 'timeout'
  error_message?: string
  duration_ms: number
  rate_limited: boolean
  created_at: string
}

export interface ToolPermissionView {
  tool_id: number
  agent_template_id?: number
  allowed: boolean
}

export interface MCPSyncResult {
  added: number
  updated: number
}

export const toolApi = {
  list(workspaceId: number): Promise<Tool[]> {
    return api.get(`/workspaces/${workspaceId}/tools`).then(res => res.data)
  },

  create(workspaceId: number, data: Partial<Tool>): Promise<Tool> {
    return api.post(`/workspaces/${workspaceId}/tools`, data).then(res => res.data)
  },

  update(workspaceId: number, id: number, data: Partial<Tool>): Promise<Tool> {
    return api.put(`/workspaces/${workspaceId}/tools/${id}`, data).then(res => res.data)
  },

  delete(workspaceId: number, id: number): Promise<void> {
    return api.delete(`/workspaces/${workspaceId}/tools/${id}`).then(res => res.data)
  },

  call(workspaceId: number, data: { tool_id: number; input_params: any; agent_task_id?: number }): Promise<ToolCallLog> {
    return api.post(`/workspaces/${workspaceId}/tools/call`, data).then(res => res.data)
  },

  listCallLogs(workspaceId: number, params?: {
    status?: string
    tool_id?: number
    agent_id?: number
    from_time?: string
    to_time?: string
    page?: number
    per_page?: number
  }): Promise<{ items: ToolCallLog[]; total: number }> {
    return api.get(`/workspaces/${workspaceId}/tools/call-logs`, { params }).then(res => res.data)
  },

  listPermissions(workspaceId: number, toolId: number): Promise<ToolPermissionView[]> {
    return api.get(`/workspaces/${workspaceId}/tools/${toolId}/permissions`).then(res => res.data)
  },

  setPermission(workspaceId: number, toolId: number, data: ToolPermissionView): Promise<void> {
    return api.put(`/workspaces/${workspaceId}/tools/${toolId}/permissions`, data).then(res => res.data)
  },

  syncMCP(workspaceId: number, mcpConfigId: number): Promise<MCPSyncResult> {
    return api.post(`/workspaces/${workspaceId}/mcp/${mcpConfigId}/sync`).then(res => res.data)
  },
}
