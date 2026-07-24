import api from './index'

export interface ApprovalRecordResponse {
  id: number
  approver_id: number
  approver_name: string
  decision: 'approved' | 'rejected'
  note: string
  decided_at: string
}

export interface ApprovalResponse {
  id: number
  issue_id: number
  workflow_id: number
  transition_id: number
  project_id: number
  workspace_id: number
  requester_id: number
  requester_name: string
  request_note: string
  source_state_id: number
  source_state_name: string
  approve_target_state_id: number
  approve_target_state_name: string
  reject_target_state_id: number
  reject_target_state_name: string
  approver_ids: number[]
  approver_names: string[]
  status: 'pending' | 'approved' | 'rejected' | 'cancelled'
  decided_by: number | null
  decided_by_name: string
  decided_at: string | null
  decision_note: string
  created_at: string
  issue_key: string
  issue_title: string
  project_name: string
  records: ApprovalRecordResponse[]
}

export interface ApprovalListQuery {
  status?: string
  project_id?: number
  approver_id?: number
}

export const approvalApi = {
  submit: (issueId: number, data: { transition_id: number; request_note: string }) =>
    api.post(`/issues/${issueId}/approvals`, data).then(r => r.data),

  listByWorkspace: (workspaceId: number, query: ApprovalListQuery = {}) =>
    api.get(`/workspaces/${workspaceId}/approvals`, { params: query }).then(r => r.data),

  listByProject: (projectId: number, workspaceId: number, query: ApprovalListQuery = {}) =>
    api.get(`/projects/${projectId}/approvals`, { params: { workspace_id: workspaceId, ...query } }).then(r => r.data),

  get: (id: number): Promise<ApprovalResponse> =>
    api.get(`/approvals/${id}`).then(r => r.data),

  decide: (id: number, data: { decision: 'approved' | 'rejected'; note?: string }) =>
    api.post(`/approvals/${id}/decide`, data).then(r => r.data),

  cancel: (id: number) =>
    api.post(`/approvals/${id}/cancel`).then(r => r.data),

  countPending: (workspaceId: number) =>
    api.get(`/workspaces/${workspaceId}/approvals/count`).then(r => r.data?.pending_count ?? 0),
}

export default approvalApi
