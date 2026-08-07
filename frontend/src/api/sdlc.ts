/**
 * SDLC API - PRD P4-006
 *
 * Orchestrate the end-to-end delivery pipeline (11 canonical stages:
 * requirement_analysis → requirement_design → dispatch_feature →
 * feature_design → breakdown_us → sprint_planning → development →
 * code_review → us_testing → fe_testing → deploy).
 *
 * Workflows run asynchronously; progress streams in via SSE events
 * `sdlc_workflow.*` and `sdlc_stage.*` (see useSSE.ts).
 */
import api from './index'

export type SDLCWorkflowStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'partial_failed'
  | 'cancelled'

export type SDLCStageStatus =
  | 'pending'
  | 'running'
  | 'success'
  | 'failed'
  | 'skipped'

export const SDLC_CANONICAL_STAGES: string[] = [
  'requirement_analysis',
  'requirement_design',
  'dispatch_feature',
  'feature_design',
  'breakdown_us',
  'sprint_planning',
  'development',
  'code_review',
  'us_testing',
  'fe_testing',
  'deploy'
]

export interface SDLCStage {
  id: number
  workflow_id: number
  workspace_id: number
  order: number
  key: string
  name: string
  agent_role: string
  status: SDLCStageStatus
  progress: number
  input: Record<string, unknown>
  output: Record<string, unknown>
  logs: string[]
  error_message?: string
  started_at?: string
  completed_at?: string
  duration_ms: number
  created_at: string
  updated_at: string
}

export interface SDLCWorkflow {
  id: number
  workspace_id: number
  project_id?: number
  squad_id?: number
  title: string
  requirement: string
  status: SDLCWorkflowStatus
  progress: number
  current_stage?: string
  config: Record<string, unknown>
  artifacts: Record<string, unknown>
  error_message?: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
  stages?: SDLCStage[]
}

export interface SDLCWorkflowCreate {
  title: string
  requirement: string
  project_id?: number
  squad_id?: number
  stages?: string[]
  fail_fast?: boolean
  config?: Record<string, unknown>
}

export interface SDLCRetryRequest {
  stage_id: number
}

export interface SDLCWorkflowListFilter {
  status?: SDLCWorkflowStatus
  limit?: number
}

export const sdlcWorkflowApi = {
  list(workspaceId: number, filter: SDLCWorkflowListFilter = {}): Promise<SDLCWorkflow[]> {
    const params: Record<string, string | number> = {}
    if (filter.status) params.status = filter.status
    if (filter.limit) params.limit = filter.limit
    return api
      .get(`/workspaces/${workspaceId}/sdlc/workflows`, { params })
      .then(res => res.data)
  },

  get(workspaceId: number, workflowId: number): Promise<SDLCWorkflow> {
    return api
      .get(`/workspaces/${workspaceId}/sdlc/workflows/${workflowId}`)
      .then(res => res.data)
  },

  create(workspaceId: number, payload: SDLCWorkflowCreate): Promise<SDLCWorkflow> {
    return api
      .post(`/workspaces/${workspaceId}/sdlc/workflows`, payload)
      .then(res => res.data)
  },

  cancel(workspaceId: number, workflowId: number): Promise<SDLCWorkflow> {
    return api
      .post(`/workspaces/${workspaceId}/sdlc/workflows/${workflowId}/cancel`)
      .then(res => res.data)
  },

  delete(workspaceId: number, workflowId: number): Promise<void> {
    return api
      .delete(`/workspaces/${workspaceId}/sdlc/workflows/${workflowId}`)
      .then(() => undefined)
  },

  retry(workspaceId: number, workflowId: number, payload: SDLCRetryRequest): Promise<SDLCWorkflow> {
    return api
      .post(`/workspaces/${workspaceId}/sdlc/workflows/${workflowId}/retry`, payload)
      .then(res => res.data)
  }
}

export const sdlcStageApi = {
  list(workspaceId: number, workflowId: number): Promise<SDLCStage[]> {
    return api
      .get(`/workspaces/${workspaceId}/sdlc/workflows/${workflowId}/stages`)
      .then(res => res.data)
  },

  get(workspaceId: number, workflowId: number, stageId: number): Promise<SDLCStage> {
    return api
      .get(`/workspaces/${workspaceId}/sdlc/workflows/${workflowId}/stages/${stageId}`)
      .then(res => res.data)
  }
}
