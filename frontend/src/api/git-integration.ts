import api from './index'

export interface GitIntegration {
  id: number
  project_id: number
  provider: string
  repo_url: string
  repo_name: string
  access_token?: string
  webhook_secret?: string
  active: boolean
  sync_prs: boolean
  sync_commits: boolean
  sync_branches: boolean
  created_at: string
  updated_at: string
}

export interface GitIssueLink {
  id: number
  issue_id: number
  git_type: string
  git_id: string
  git_url: string
  git_title: string
  git_state: string
  git_author: string
  git_branch: string
  created_at: string
  updated_at: string
}

export async function createGitIntegration(
  workspaceId: number,
  projectId: number,
  data: {
    provider: string
    repo_url: string
    repo_name: string
    access_token?: string
    webhook_secret?: string
  }
): Promise<GitIntegration> {
  const response = await api.post(`/workspaces/${workspaceId}/git-integration?project_id=${projectId}`, data)
  return response.data
}

export async function getGitIntegration(
  workspaceId: number,
  projectId: number
): Promise<GitIntegration> {
  const response = await api.get(`/workspaces/${workspaceId}/git-integration?project_id=${projectId}`)
  return response.data
}

export async function updateGitIntegration(
  workspaceId: number,
  projectId: number,
  data: Partial<GitIntegration>
): Promise<GitIntegration> {
  const response = await api.put(`/workspaces/${workspaceId}/git-integration?project_id=${projectId}`, data)
  return response.data
}

export async function deleteGitIntegration(
  workspaceId: number,
  projectId: number
): Promise<void> {
  await api.delete(`/workspaces/${workspaceId}/git-integration?project_id=${projectId}`)
}

export async function getIssueGitLinks(
  workspaceId: number,
  issueId: number
): Promise<GitIssueLink[]> {
  const response = await api.get(`/workspaces/${workspaceId}/issues/${issueId}/git-links`)
  return response.data
}