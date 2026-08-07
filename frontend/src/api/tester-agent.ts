/**
 * Tester Agent API - PRD P4-002
 *
 * The Tester Agent takes a requirement (with optional acceptance criteria),
 * generates test cases, executes them, and reports any bugs found as work
 * items. Each run is tracked as a TesterJob with a full lifecycle:
 * pending → generating_cases → executing → reporting → completed.
 */
import api from './index'

export type TesterJobStatus =
  | 'pending'
  | 'generating_cases'
  | 'executing'
  | 'reporting'
  | 'completed'
  | 'failed'
  | 'cancelled'

export interface TestCase {
  id: string
  name: string
  description: string
  steps: string[]
  expected: string
}

export interface TestResult {
  case_id: string
  name: string
  status: 'passed' | 'failed' | 'skipped'
  duration_ms: number
  error?: string
}

export interface TesterJob {
  id: number
  workspace_id: number
  project_id?: number
  issue_id?: number
  agent_task_id?: number
  title: string
  requirement_text: string
  acceptance_criteria: string
  test_scope: string // unit | integration | e2e
  input_context: Record<string, unknown>
  generated_cases: TestCase[]
  test_results: TestResult[]
  total_cases: number
  pass_count: number
  fail_count: number
  skip_count: number
  bug_issue_ids: number[]
  status: TesterJobStatus
  progress: number
  current_step?: string
  error_message?: string
  started_at?: string
  completed_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
}

export interface TesterJobCreate {
  title: string
  requirement_text?: string
  acceptance_criteria?: string
  test_scope?: string
  project_id?: number
  issue_id?: number
  cases?: TestCase[]
  input_context?: Record<string, unknown>
}

export interface TesterJobListFilter {
  status?: TesterJobStatus
  limit?: number
}

export const testerAgentApi = {
  list(workspaceId: number, filter: TesterJobListFilter = {}): Promise<TesterJob[]> {
    const params: Record<string, string | number> = {}
    if (filter.status) params.status = filter.status
    if (filter.limit) params.limit = filter.limit
    return api
      .get(`/workspaces/${workspaceId}/tester-agent/jobs`, { params })
      .then(res => res.data)
  },

  get(workspaceId: number, jobId: number): Promise<TesterJob> {
    return api
      .get(`/workspaces/${workspaceId}/tester-agent/jobs/${jobId}`)
      .then(res => res.data)
  },

  create(workspaceId: number, payload: TesterJobCreate): Promise<TesterJob> {
    return api
      .post(`/workspaces/${workspaceId}/tester-agent/jobs`, payload)
      .then(res => res.data)
  },

  cancel(workspaceId: number, jobId: number): Promise<TesterJob> {
    return api
      .post(`/workspaces/${workspaceId}/tester-agent/jobs/${jobId}/cancel`)
      .then(res => res.data)
  },

  delete(workspaceId: number, jobId: number): Promise<void> {
    return api
      .delete(`/workspaces/${workspaceId}/tester-agent/jobs/${jobId}`)
      .then(() => undefined)
  }
}
