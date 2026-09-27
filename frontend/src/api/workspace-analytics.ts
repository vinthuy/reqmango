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

export async function getWorkspaceAnalytics(
  wsParam: string,
  params: { days: number; project_id?: number },
): Promise<WorkspaceAnalytics> {
  const response = await api.get(`/workspaces/${wsParam}/analytics`, { params })
  return response.data
}
