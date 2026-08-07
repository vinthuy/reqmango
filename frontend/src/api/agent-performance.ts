import api from './index'

// Workspace-wide aggregated metrics.
export interface AgentPerformanceOverview {
  workspace_id: number
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  cancelled_tasks: number
  running_tasks: number
  pending_tasks: number
  success_rate: number          // 0-100
  avg_duration_seconds: number
  total_duration_seconds: number
  period_start?: string
  period_end?: string
}

// Per-template performance breakdown.
export interface TemplatePerformance {
  agent_template_id?: number
  template_name: string
  task_type?: string
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  cancelled_tasks: number
  success_rate: number
  avg_duration_seconds: number
  last_run_at?: string
}

// Time-bucketed metric point for trend charts.
export interface TimelinePoint {
  bucket_start: string
  bucket_end: string
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  success_rate: number
  avg_duration_seconds: number
}

// Failure reason tally.
export interface FailureBreakdown {
  failure_reason: string
  count: number
  percentage: number
}

export type BucketGranularity = 'day' | 'week' | 'month'

export interface PeriodFilter {
  from?: string  // RFC3339
  to?: string    // RFC3339
}

function toParams(period: PeriodFilter, extra: Record<string, string> = {}): Record<string, string> {
  const params: Record<string, string> = { ...extra }
  if (period.from) params.from = period.from
  if (period.to) params.to = period.to
  return params
}

export const agentPerformanceApi = {
  overview(workspaceId: number, period: PeriodFilter = {}): Promise<AgentPerformanceOverview> {
    return api
      .get(`/workspaces/${workspaceId}/agent-performance/overview`, { params: toParams(period) })
      .then(res => res.data)
  },

  byTemplate(workspaceId: number, period: PeriodFilter = {}): Promise<TemplatePerformance[]> {
    return api
      .get(`/workspaces/${workspaceId}/agent-performance/by-template`, { params: toParams(period) })
      .then(res => res.data)
  },

  timeline(
    workspaceId: number,
    bucket: BucketGranularity = 'day',
    period: PeriodFilter = {}
  ): Promise<TimelinePoint[]> {
    return api
      .get(`/workspaces/${workspaceId}/agent-performance/timeline`, {
        params: toParams(period, { bucket })
      })
      .then(res => res.data)
  },

  failureBreakdown(workspaceId: number, period: PeriodFilter = {}): Promise<FailureBreakdown[]> {
    return api
      .get(`/workspaces/${workspaceId}/agent-performance/failures`, { params: toParams(period) })
      .then(res => res.data)
  }
}
