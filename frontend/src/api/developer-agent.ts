/**
 * Developer Agent API - PRD P4-001
 *
 * The Developer Agent takes a requirement (US work item + optional design doc),
 * generates code, commits it to a Git provider (GitHub), and opens a Pull
 * Request. Each run is tracked as a DeveloperJob with a full lifecycle:
 * pending → analyzing → generating → committing → opening_pr → completed.
 */
import api from './index'

export type DeveloperJobStatus =
  | 'pending'
  | 'analyzing'
  | 'generating'
  | 'committing'
  | 'opening_pr'
  | 'completed'
  | 'failed'
  | 'cancelled'

export interface GeneratedFile {
  path: string
  content: string
  mode?: string
}

export interface DeveloperJob {
  id: number
  workspace_id: number
  project_id?: number
  issue_id?: number
  agent_task_id?: number
  git_provider: string
  git_connection_id?: number
  title: string
  requirement_text: string
  design_doc_url?: string
  input_context: Record<string, unknown>
  branch_name: string
  base_branch: string
  commit_message: string
  generated_files: GeneratedFile[]
  pr_number?: number
  pr_url?: string
  pr_title?: string
  commit_sha?: string
  status: DeveloperJobStatus
  progress: number
  current_step?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
}

export interface DeveloperJobCreate {
  title: string
  requirement_text?: string
  design_doc_url?: string
  project_id?: number
  issue_id?: number
  git_connection_id: number
  git_provider?: string
  branch_name?: string
  base_branch?: string
  commit_message?: string
  pr_title?: string
  pr_body?: string
  language?: string
  files?: GeneratedFile[]
  input_context?: Record<string, unknown>
}

export interface DeveloperJobListFilter {
  status?: DeveloperJobStatus
  limit?: number
}

export const developerAgentApi = {
  list(workspaceId: number, filter: DeveloperJobListFilter = {}): Promise<DeveloperJob[]> {
    const params: Record<string, string | number> = {}
    if (filter.status) params.status = filter.status
    if (filter.limit) params.limit = filter.limit
    return api
      .get(`/workspaces/${workspaceId}/developer-agent/jobs`, { params })
      .then(res => res.data)
  },

  get(workspaceId: number, jobId: number): Promise<DeveloperJob> {
    return api
      .get(`/workspaces/${workspaceId}/developer-agent/jobs/${jobId}`)
      .then(res => res.data)
  },

  create(workspaceId: number, payload: DeveloperJobCreate): Promise<DeveloperJob> {
    return api
      .post(`/workspaces/${workspaceId}/developer-agent/jobs`, payload)
      .then(res => res.data)
  },

  cancel(workspaceId: number, jobId: number): Promise<DeveloperJob> {
    return api
      .post(`/workspaces/${workspaceId}/developer-agent/jobs/${jobId}/cancel`)
      .then(res => res.data)
  },

  delete(workspaceId: number, jobId: number): Promise<void> {
    return api
      .delete(`/workspaces/${workspaceId}/developer-agent/jobs/${jobId}`)
      .then(() => undefined)
  }
}
