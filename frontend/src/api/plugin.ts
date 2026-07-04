/**
 * Plugin API - Plugin management and catalog
 */
import api from './index'
import type { Plugin, PluginInfo, PluginInstallRequest, PluginUpdateRequest, PluginEventLog } from '@/types/plugin'

export const pluginApi = {
  /** List available plugins in the built-in catalog */
  catalog: async (workspaceId: number): Promise<PluginInfo[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/plugins/catalog`)
    return res.data
  },

  /** List installed plugins in a workspace */
  list: async (workspaceId: number): Promise<Plugin[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/plugins`)
    return res.data
  },

  /** Install a plugin from the catalog */
  install: async (workspaceId: number, data: PluginInstallRequest): Promise<Plugin> => {
    const res = await api.post(`/workspaces/${workspaceId}/plugins`, data)
    return res.data
  },

  /** Get a single installed plugin */
  get: async (workspaceId: number, id: number): Promise<Plugin> => {
    const res = await api.get(`/workspaces/${workspaceId}/plugins/${id}`)
    return res.data
  },

  /** Update plugin configuration */
  update: async (workspaceId: number, id: number, data: PluginUpdateRequest): Promise<Plugin> => {
    const res = await api.put(`/workspaces/${workspaceId}/plugins/${id}`, data)
    return res.data
  },

  /** Uninstall a plugin */
  uninstall: async (workspaceId: number, id: number): Promise<void> => {
    await api.delete(`/workspaces/${workspaceId}/plugins/${id}`)
  },

  /** Enable a plugin */
  enable: async (workspaceId: number, id: number): Promise<Plugin> => {
    const res = await api.post(`/workspaces/${workspaceId}/plugins/${id}/enable`)
    return res.data
  },

  /** Disable a plugin */
  disable: async (workspaceId: number, id: number): Promise<Plugin> => {
    const res = await api.post(`/workspaces/${workspaceId}/plugins/${id}/disable`)
    return res.data
  },

  /** Get plugin execution logs */
  logs: async (workspaceId: number, id: number, limit?: number): Promise<PluginEventLog[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/plugins/${id}/logs`, {
      params: limit ? { limit } : undefined,
    })
    return res.data
  },

  /** Test plugin execution */
  test: async (workspaceId: number, id: number): Promise<{ message: string; payload: string }> => {
    const res = await api.post(`/workspaces/${workspaceId}/plugins/${id}/test`)
    return res.data
  },
}
