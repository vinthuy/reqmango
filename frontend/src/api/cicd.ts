/**
 * CI/CD API - PRD P4-005
 *
 * Manage CI/CD configs (workspace- or project-scoped) and trigger/monitor
 * builds. Each build runs an async workflow whose progress streams in via
 * SSE events `cicd_build.*` (see useSSE.ts).
 */
import api from './index'

export type CICDProvider =
  | 'github_actions'
  | 'gitlab_ci'
  | 'jenkins'
  | 'generic'

export type BuildStatus =
  | 'pending'
  | 'queued'
  | 'running'
  | 'success'
  | 'failed'
  | 'cancelled'
  | 'unknown'

export type BuildTrigger =
  | 'manual'
  | 'push'
  | 'pull_request'
  | 'schedule'
  | 'agent'
  | 'webhook'

export interface BuildStage {
  name: string
  status: 'pending' | 'running' | 'success' | 'failed' | 'skipped'
  duration_ms: number
  started_at?: string
  completed_at?: string
  log_url?: string
}

export interface CICDConfig {
  id: number
  workspace_id: number
  project_id?: number
  name: string
  provider: CICDProvider
  api_endpoint: string
  project_slug: string
  default_branch: string
  auth_token_ref: string
  trigger_events: string[]
  extra_config: Record<string, unknown>
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface CICDConfigCreate {
  name: string
  provider?: CICDProvider
  api_endpoint?: string
  project_slug?: string
  default_branch?: string
  auth_token_ref?: string
  trigger_events?: string[]
  extra_config?: Record<string, unknown>
  project_id?: number
  enabled?: boolean
}

export interface CICDConfigUpdate {
  name?: string
  provider?: CICDProvider
  api_endpoint?: string
  project_slug?: string
  default_branch?: string
  auth_token_ref?: string
  trigger_events?: string[]
  extra_config?: Record<string, unknown>
  enabled?: boolean
}

export interface CICDConfigListFilter {
  project_id?: number
}

export interface BuildRecord {
  id: number
  workspace_id: number
  project_id?: number
  cicd_config_id: number
  cicd_config_name?: string
  trigger: BuildTrigger
  branch: string
  commit_sha: string
  issue_id?: number
  agent_task_id?: number
  triggered_by_id: number
  external_build_id: string
  build_url: string
  stages: BuildStage[]
  status: BuildStatus
  progress: number
  current_stage?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
  duration_ms: number
  created_at: string
  updated_at: string
}

export interface BuildTriggerRequest {
  cicd_config_id: number
  branch?: string
  commit_sha?: string
  trigger?: BuildTrigger
  project_id?: number
  issue_id?: number
  agent_task_id?: number
  extra?: Record<string, unknown>
}

export interface BuildListFilter {
  config_id?: number
  status?: BuildStatus
  project_id?: number
  limit?: number
}

export const cicdConfigApi = {
  list(workspaceId: number, filter: CICDConfigListFilter = {}): Promise<CICDConfig[]> {
    const params: Record<string, number> = {}
    if (filter.project_id) params.project_id = filter.project_id
    return api
      .get(`/workspaces/${workspaceId}/cicd/configs`, { params })
      .then(res => res.data)
  },

  get(workspaceId: number, configId: number): Promise<CICDConfig> {
    return api
      .get(`/workspaces/${workspaceId}/cicd/configs/${configId}`)
      .then(res => res.data)
  },

  create(workspaceId: number, payload: CICDConfigCreate): Promise<CICDConfig> {
    return api
      .post(`/workspaces/${workspaceId}/cicd/configs`, payload)
      .then(res => res.data)
  },

  update(workspaceId: number, configId: number, payload: CICDConfigUpdate): Promise<CICDConfig> {
    return api
      .patch(`/workspaces/${workspaceId}/cicd/configs/${configId}`, payload)
      .then(res => res.data)
  },

  delete(workspaceId: number, configId: number): Promise<void> {
    return api
      .delete(`/workspaces/${workspaceId}/cicd/configs/${configId}`)
      .then(() => undefined)
  }
}

export const cicdBuildApi = {
  list(workspaceId: number, filter: BuildListFilter = {}): Promise<BuildRecord[]> {
    const params: Record<string, string | number> = {}
    if (filter.config_id) params.config_id = filter.config_id
    if (filter.status) params.status = filter.status
    if (filter.project_id) params.project_id = filter.project_id
    if (filter.limit) params.limit = filter.limit
    return api
      .get(`/workspaces/${workspaceId}/cicd/builds`, { params })
      .then(res => res.data)
  },

  get(workspaceId: number, buildId: number): Promise<BuildRecord> {
    return api
      .get(`/workspaces/${workspaceId}/cicd/builds/${buildId}`)
      .then(res => res.data)
  },

  trigger(workspaceId: number, payload: BuildTriggerRequest): Promise<BuildRecord> {
    return api
      .post(`/workspaces/${workspaceId}/cicd/builds`, payload)
      .then(res => res.data)
  },

  cancel(workspaceId: number, buildId: number): Promise<BuildRecord> {
    return api
      .post(`/workspaces/${workspaceId}/cicd/builds/${buildId}/cancel`)
      .then(res => res.data)
  },

  delete(workspaceId: number, buildId: number): Promise<void> {
    return api
      .delete(`/workspaces/${workspaceId}/cicd/builds/${buildId}`)
      .then(() => undefined)
  }
}
