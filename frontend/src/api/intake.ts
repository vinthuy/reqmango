/**
 * Intake API — 需求池/Intake 管理模块
 */
import api from './index'

export interface IntakeIssue {
  id: number
  name: string
  description_html?: string
  priority: string
  submitter: string
  intake_status: string
  intake_source: string
  created_at: string
}

export interface IntakeSubmit {
  name: string
  description?: string
  priority?: string
  type_id?: number
  submitter: string
}

export interface IntakeTriage {
  action: 'accept' | 'reject'
  assignee_id?: number
  state_id?: number
}

export const intakeApi = {
  /** 公开提交需求（无需认证） */
  submit: async (projectId: number, data: IntakeSubmit): Promise<any> => {
    const res = await api.post(`/intake/${projectId}`, data)
    return res.data
  },

  /** 列出待处理 intake 项 */
  listPending: async (projectId: number): Promise<IntakeIssue[]> => {
    const res = await api.get(`/projects/${projectId}/intake`)
    return res.data
  },

  /** 接受或拒绝 intake 项 */
  triage: async (projectId: number, issueId: number, data: IntakeTriage): Promise<any> => {
    const res = await api.post(`/projects/${projectId}/intake/${issueId}/triage`, data)
    return res.data
  },

  /** AI 分析 intake 项 */
  aiAnalyze: async (projectId: number, issueId: number): Promise<any> => {
    const res = await api.post(`/projects/${projectId}/intake/${issueId}/ai-analyze`)
    return res.data
  },
}
