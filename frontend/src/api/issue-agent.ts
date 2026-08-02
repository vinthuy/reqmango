import apiClient from './index'

export interface AssignAgentRequest {
  agent_id: number
  priority?: string
  deadline?: string
  notes?: string
}

export interface AgentStatus {
  issue_id: number
  agent_id: number | null
  agent_name: string
  task_id: number | null
  task_status: string
  task_progress: number
  started_at: string | null
  estimated_end: string | null
  sla_breach: boolean
}

export interface ExecutionPreview {
  agent_id: number
  agent_name: string
  steps: string[]
  estimated_time: number
  estimated_cost: number
  confidence: number
}

export interface BulkAssignRequest {
  issue_ids: number[]
  agent_id: number
}

export interface CompleteWorkRequest {
  summary?: string
  decision_ids?: number[]
}

export interface EscalateRequest {
  reason: string
  escalation_type: 'needs_human' | 'approval' | 'resource_limit' | 'deadline_risk'
}

export interface UpdateStatusRequest {
  status: string
}

export const issueAgentApi = {
  assign(issueId: number, data: AssignAgentRequest) {
    return apiClient.post(`/issues/${issueId}/assign-agent`, data)
  },

  unassign(issueId: number) {
    return apiClient.delete(`/issues/${issueId}/unassign-agent`)
  },

  getStatus(issueId: number) {
    return apiClient.get<AgentStatus>(`/issues/${issueId}/agent-status`)
  },

  previewExecution(issueId: number, agentId: number) {
    return apiClient.get<ExecutionPreview>(`/issues/${issueId}/preview-agent`, {
      params: { agent_id: agentId }
    })
  },

  bulkAssign(data: BulkAssignRequest) {
    return apiClient.post('/issues/bulk/assign-agent', data)
  },

  startWork(issueId: number) {
    return apiClient.post(`/issues/${issueId}/start-work`)
  },

  completeWork(issueId: number, data: CompleteWorkRequest) {
    return apiClient.post(`/issues/${issueId}/complete-work`, data)
  },

  escalate(issueId: number, data: EscalateRequest) {
    return apiClient.post(`/issues/${issueId}/escalate`, data)
  },

  updateStatus(issueId: number, data: UpdateStatusRequest) {
    return apiClient.patch(`/issues/${issueId}/agent-status`, data)
  }
}
