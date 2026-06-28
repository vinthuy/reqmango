/**
 * Slack API - Slack integration and notification management
 */
import api from './index'

export interface SlackConnection {
  id: number
  workspace_id: number
  project_id: number
  channel_name: string
  is_enabled: boolean
  notify_on_create: boolean
  notify_on_update: boolean
  notify_on_comment: boolean
  notify_on_complete: boolean
  created_at: string
  updated_at: string
}

export interface SlackNotification {
  issue_id: number
  issue_name: string
  event: string
  user: string
  url: string
}

export const slackApi = {
  list: async (workspaceId: number): Promise<SlackConnection[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/slack`)
    return res.data
  },

  get: async (workspaceId: number, id: number): Promise<SlackConnection> => {
    const res = await api.get(`/workspaces/${workspaceId}/slack/${id}`)
    return res.data
  },

  create: async (workspaceId: number, data: {
    project_id: number
    channel_name: string
    webhook_url: string
    bot_token?: string
    is_enabled?: boolean
    notify_on_create?: boolean
    notify_on_update?: boolean
    notify_on_comment?: boolean
    notify_on_complete?: boolean
  }): Promise<SlackConnection> => {
    const res = await api.post(`/workspaces/${workspaceId}/slack`, data)
    return res.data
  },

  update: async (workspaceId: number, id: number, data: Partial<{
    channel_name: string
    webhook_url: string
    bot_token: string
    is_enabled: boolean
    notify_on_create: boolean
    notify_on_update: boolean
    notify_on_comment: boolean
    notify_on_complete: boolean
  }>): Promise<SlackConnection> => {
    const res = await api.put(`/workspaces/${workspaceId}/slack/${id}`, data)
    return res.data
  },

  delete: async (workspaceId: number, id: number): Promise<void> => {
    await api.delete(`/workspaces/${workspaceId}/slack/${id}`)
  },

  sendNotification: async (workspaceId: number, id: number, notif: SlackNotification): Promise<void> => {
    await api.post(`/workspaces/${workspaceId}/slack/${id}/notify`, notif)
  },

  testNotification: async (workspaceId: number, id: number): Promise<{ status: string; channel: string }> => {
    const res = await api.post(`/workspaces/${workspaceId}/slack/${id}/test`)
    return res.data
  },
}
