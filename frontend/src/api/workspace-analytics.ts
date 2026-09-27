import api from './index'

export interface AnalyticsBucket {
  key: string
  count: number
}

export interface WorkspaceAnalyticsProject {
  id: number
  name: string
  identifier: string
  color: string
  total: number
  open: number
  completed: number
  overdue: number
  created_in_range: number
  completed_in_range: number
  completion_rate: number
}

export interface WorkspaceAnalyticsAssignee {
  user_id: number
  name: string
  open: number
  overdue: number
  completed_in_range: number
}

export interface WorkspaceAnalytics {
  days: number
  granularity: 'day' | 'week'
  summary: {
    total: number
    open: number
    overdue: number
    created_in_range: number
    completed_in_range: number
    completion_rate: number
    avg_cycle_days: number | null
    project_count: number
  }
  trend: { date: string; created: number; completed: number }[]
  state_groups: AnalyticsBucket[]
  priorities: AnalyticsBucket[]
  issue_types: AnalyticsBucket[]
  projects: WorkspaceAnalyticsProject[]
  assignees: WorkspaceAnalyticsAssignee[]
  generated_at: string
}

export type LoopStageKey = 'triage' | 'queue' | 'develop' | 'release'

export interface LoopDuration {
  key: string
  samples: number
  median_hours: number | null
  p85_hours: number | null
  avg_hours: number | null
}

export interface LoopStage extends LoopDuration {
  key: LoopStageKey
  wip: number
  wip_median_hours: number | null
}

export interface DeliveryLoop {
  days: number
  summary: {
    received: number
    from_intake: number
    accepted: number
    rejected: number
    done: number
    shipped: number
    ship_rate: number
    spec_coverage: number | null
    delivered_human: number
    delivered_ai: number
  }
  funnel: { key: 'received' | 'accepted' | 'started' | 'pr_linked' | 'done' | 'shipped'; count: number; rate: number; step_rate: number }[]
  stages: LoopStage[]
  bottleneck: LoopStageKey | ''
  pr_review: LoopDuration
  lead_to_done: LoopDuration
  lead_to_ship: LoopDuration
  stalled: { issue_id: number; project_id: number; key: string; name: string; stage: LoopStageKey; since: string; age_hours: number }[]
  generated_at: string
}

export async function getDeliveryLoop(
  wsParam: string,
  params: { days: number; project_id?: number },
): Promise<DeliveryLoop> {
  const response = await api.get(`/workspaces/${wsParam}/analytics/loop`, { params })
  return response.data
}

export async function getWorkspaceAnalytics(
  wsParam: string,
  params: { days: number; project_id?: number },
): Promise<WorkspaceAnalytics> {
  const response = await api.get(`/workspaces/${wsParam}/analytics`, { params })
  return response.data
}
