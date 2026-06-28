/**
 * MCP API - MCP Server configuration and tool management
 */
import api from './index'

export interface MCPConfig {
  id: number
  name: string
  description: string
  workspace_id: number
  server_url: string
  transport_type: string
  is_enabled: boolean
  tools_count: number
  last_sync_at: string | null
  created_at: string
  updated_at: string
}

export interface MCPTool {
  name: string
  description: string
  input_schema: string
}

export interface MCPExecuteResult {
  [key: string]: any
}

export const mcpApi = {
  list: async (workspaceId: number): Promise<MCPConfig[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/mcp`)
    return res.data
  },

  get: async (workspaceId: number, id: number): Promise<MCPConfig> => {
    const res = await api.get(`/workspaces/${workspaceId}/mcp/${id}`)
    return res.data
  },

  create: async (workspaceId: number, data: {
    name: string
    description?: string
    server_url: string
    transport_type?: string
    api_key?: string
    is_enabled?: boolean
  }): Promise<MCPConfig> => {
    const res = await api.post(`/workspaces/${workspaceId}/mcp`, data)
    return res.data
  },

  update: async (workspaceId: number, id: number, data: Partial<{
    name: string
    description: string
    server_url: string
    transport_type: string
    api_key: string
    is_enabled: boolean
  }>): Promise<MCPConfig> => {
    const res = await api.put(`/workspaces/${workspaceId}/mcp/${id}`, data)
    return res.data
  },

  delete: async (workspaceId: number, id: number): Promise<void> => {
    await api.delete(`/workspaces/${workspaceId}/mcp/${id}`)
  },

  discoverTools: async (workspaceId: number, id: number): Promise<MCPTool[]> => {
    const res = await api.post(`/workspaces/${workspaceId}/mcp/${id}/discover`)
    return res.data
  },

  getTools: async (workspaceId: number, id: number): Promise<MCPTool[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/mcp/${id}/tools`)
    return res.data
  },

  executeTool: async (workspaceId: number, id: number, toolName: string, args: Record<string, any>): Promise<MCPExecuteResult> => {
    const res = await api.post(`/workspaces/${workspaceId}/mcp/${id}/execute`, {
      tool_name: toolName,
      arguments: args,
    })
    return res.data
  },
}
