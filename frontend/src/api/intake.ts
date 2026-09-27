/**
 * Intake API — 需求入口：多渠道收集、分诊、合并重复、时效与指标
 */
import api from './index'

export type IntakeStatus = 'pending' | 'snoozed' | 'accepted' | 'rejected' | 'duplicate'
export type IntakeSource = 'form' | 'webhook' | 'email'

export interface IntakeItem {
  id: number
  sequence_id: number
  name: string
  description_html: string
  priority: string
  source: IntakeSource
  status: IntakeStatus
  submitter: string | null
  email: string | null
  created_at: string
  triaged_at: string | null
  triaged_by_name: string | null
  snoozed_until: string | null
  duplicate_of: number | null
  duplicate_sequence_id: number | null
  duplicate_name: string | null
  note: string | null
  state_id: number
  age_hours: number
  sla_overdue: boolean
}

export interface IntakeListResult {
  items: IntakeItem[]
  total: number
  counts: Record<IntakeStatus, number>
  sla_hours: number
}

export interface IntakeSettings {
  project_id: number
  form_enabled: boolean
  webhook_enabled: boolean
  email_enabled: boolean
  token: string
  sla_hours: number
}

export interface IntakeMetrics {
  days: number
  sla_hours: number
  received: number
  accepted: number
  rejected: number
  duplicate: number
  pending_now: number
  snoozed_now: number
  overdue_now: number
  acceptance_rate: number | null
  avg_triage_hours: number | null
  median_triage_hours: number | null
  sla_met_rate: number | null
  by_source: { source: IntakeSource; received: number; accepted: number }[]
  trend: { date: string; received: number; triaged: number }[]
}

export interface IntakeSubmit {
  name: string
  description?: string
  priority?: string
  type_id?: number
  submitter?: string
  email?: string
}

export interface IntakeTriage {
  action: 'accept' | 'reject' | 'snooze' | 'duplicate' | 'reopen'
  assignee_id?: number
  state_id?: number
  reason?: string
  snooze_hours?: number
  duplicate_of?: number
}

export interface IntakeProjectInfo {
  id: number
  name: string
  identifier: string
  form_enabled: boolean
}

export const intakeApi = {
  submit: async (projectId: number, data: IntakeSubmit) => (await api.post(`/intake/${projectId}`, data)).data,

  info: async (projectId: number): Promise<IntakeProjectInfo> => (await api.get(`/intake/${projectId}/info`)).data,

  list: async (
    projectId: number,
    params: { status?: IntakeStatus; source?: string; q?: string; limit?: number; offset?: number } = {},
  ): Promise<IntakeListResult> => (await api.get(`/projects/${projectId}/intake`, { params })).data,

  triage: async (projectId: number, issueId: number, data: IntakeTriage) =>
    (await api.post(`/projects/${projectId}/intake/${issueId}/triage`, data)).data,

  aiAnalyze: async (projectId: number, issueId: number) =>
    (await api.post(`/projects/${projectId}/intake/${issueId}/ai-analyze`)).data,

  metrics: async (projectId: number, days: number): Promise<IntakeMetrics> =>
    (await api.get(`/projects/${projectId}/intake/metrics`, { params: { days } })).data,

  getSettings: async (projectId: number): Promise<IntakeSettings> =>
    (await api.get(`/projects/${projectId}/intake/settings`)).data,

  updateSettings: async (
    projectId: number,
    data: Partial<Pick<IntakeSettings, 'form_enabled' | 'webhook_enabled' | 'email_enabled' | 'sla_hours'>> & { rotate_token?: boolean },
  ): Promise<IntakeSettings> => (await api.put(`/projects/${projectId}/intake/settings`, data)).data,
}
