/**
 * Report API — 报表生成 + 保存
 */
import api from './index'

export interface ReportRequest {
  rql?: string
  report_type: string  // distribution | created_vs_resolved | avg_age | current_age | created_trend
  group_by?: string     // state | priority | assignee | type | label | cycle | module
  chart?: string        // bar | pie | doughnut | table | line
  date_from?: string
  date_to?: string
  interval?: string     // day | week | month
}

export interface ReportResponse {
  type: string
  labels: string[]
  values: number[]
  values2?: number[]
  total: number
  group_by?: string
  colors?: Record<string, string>
  summary?: { avg_days?: number }
}

export interface SavedReport {
  id?: number
  name: string
  report_type: string
  group_by: string
  chart_type: string
  rql: string
  interval: string
  date_from: string
  date_to: string
  project_id?: number
  created_at?: string
  updated_at?: string
}

export const reportApi = {
  generate: async (projectId: number, data: ReportRequest): Promise<ReportResponse> => {
    const res = await api.post(`/projects/${projectId}/reports`, data)
    return res.data
  },
}

export const savedReportApi = {
  list: async (projectId: number): Promise<SavedReport[]> => {
    const res = await api.get(`/projects/${projectId}/saved-reports`)
    return res.data
  },
  create: async (projectId: number, data: SavedReport): Promise<SavedReport> => {
    const res = await api.post(`/projects/${projectId}/saved-reports`, data)
    return res.data
  },
  update: async (projectId: number, id: number, data: Partial<SavedReport>): Promise<any> => {
    const res = await api.patch(`/projects/${projectId}/saved-reports/${id}`, data)
    return res.data
  },
  delete: async (projectId: number, id: number): Promise<any> => {
    const res = await api.delete(`/projects/${projectId}/saved-reports/${id}`)
    return res.data
  },
}
