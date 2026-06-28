/**
 * GitHub API - GitHub integration and repository sync
 */
import api from './index'

export interface GitHubConnection {
  id: number
  workspace_id: number
  project_id: number
  repo_owner: string
  repo_name: string
  is_enabled: boolean
  sync_issues: boolean
  sync_prs: boolean
  last_sync_at: string | null
  webhook_id: number | null
  created_at: string
  updated_at: string
}

export interface GitHubIssue {
  number: number
  title: string
  body: string
  state: string
  labels: string[]
  html_url: string
  created_at: string
  updated_at: string
}

export const githubApi = {
  list: async (workspaceId: number): Promise<GitHubConnection[]> => {
    const res = await api.get(`/workspaces/${workspaceId}/github`)
    return res.data
  },

  get: async (workspaceId: number, id: number): Promise<GitHubConnection> => {
    const res = await api.get(`/workspaces/${workspaceId}/github/${id}`)
    return res.data
  },

  create: async (workspaceId: number, data: {
    project_id: number
    repo_owner: string
    repo_name: string
    access_token?: string
    webhook_secret?: string
    is_enabled?: boolean
    sync_issues?: boolean
    sync_prs?: boolean
  }): Promise<GitHubConnection> => {
    const res = await api.post(`/workspaces/${workspaceId}/github`, data)
    return res.data
  },

  update: async (workspaceId: number, id: number, data: Partial<{
    repo_owner: string
    repo_name: string
    access_token: string
    webhook_secret: string
    is_enabled: boolean
    sync_issues: boolean
    sync_prs: boolean
  }>): Promise<GitHubConnection> => {
    const res = await api.put(`/workspaces/${workspaceId}/github/${id}`, data)
    return res.data
  },

  delete: async (workspaceId: number, id: number): Promise<void> => {
    await api.delete(`/workspaces/${workspaceId}/github/${id}`)
  },

  syncIssues: async (workspaceId: number, id: number): Promise<{ synced: number, issues: GitHubIssue[] }> => {
    const res = await api.post(`/workspaces/${workspaceId}/github/${id}/sync`)
    return res.data
  },
}
