/**
 * Webhook API — Webhook 管理模块
 */
import api from './index'

export interface WebhookCreate {
  name: string
  url: string
  secret?: string
  events: string
}

export interface WebhookUpdate {
  name?: string
  url?: string
  secret?: string
  events?: string
  is_active?: boolean
}

export interface Webhook {
  id: number
  name: string
  url: string
  secret: string
  events: string
  is_active: boolean
  project_id: number
  workspace_id: number
  created_at: string
  updated_at: string
}

export const webhookApi = {
  list: async (projectId: number): Promise<Webhook[]> => {
    const res = await api.get(`/projects/${projectId}/webhooks`)
    return res.data
  },

  create: async (projectId: number, data: WebhookCreate): Promise<Webhook> => {
    const res = await api.post(`/projects/${projectId}/webhooks`, data)
    return res.data
  },

  update: async (projectId: number, id: number, data: WebhookUpdate): Promise<Webhook> => {
    const res = await api.put(`/projects/${projectId}/webhooks/${id}`, data)
    return res.data
  },

  remove: async (projectId: number, id: number): Promise<void> => {
    await api.delete(`/projects/${projectId}/webhooks/${id}`)
  },
}
