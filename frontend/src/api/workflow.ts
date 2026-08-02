import apiClient from './index'

export interface Workflow {
  id: number
  name: string
  description: string
  project_id: number
  workspace_id: number
  version: number
  is_active: boolean
  trigger_type: string
  node_count: number
  edge_count: number
  created_at: string
  updated_at: string
}

export interface WorkflowNode {
  id: number
  workflow_id: number
  agent_id: number
  agent_name: string
  node_type: string
  name: string
  config: any
  sort_order: number
  timeout: number
  retry_policy: string
  max_retries: number
}

export interface WorkflowEdge {
  id: number
  workflow_id: number
  source_node_id: number
  target_node_id: number
  condition: string
  context_mapping: any
}

export interface WorkflowDetail extends Workflow {
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}

export interface WorkflowRun {
  id: number
  workflow_id: number
  issue_id: number | null
  status: string
  started_at: string | null
  completed_at: string | null
  total_tokens: number
  total_cost: number
  error_info: string
  created_at: string
}

export interface WorkflowNodeRun {
  id: number
  workflow_run_id: number
  node_id: number
  node_name: string
  agent_id: number
  agent_name: string
  status: string
  started_at: string | null
  completed_at: string | null
  tokens_used: number
  cost: number
  error_info: string
  retry_count: number
}

export interface WorkflowRunDetail extends WorkflowRun {
  node_runs: WorkflowNodeRun[]
}

export interface CreateWorkflowRequest {
  name: string
  description?: string
  trigger_type?: string
}

export interface UpdateWorkflowRequest {
  name?: string
  description?: string
  is_active?: boolean
  trigger_type?: string
  issue_type_ids?: number[]
}

export interface CreateNodeRequest {
  agent_id: number
  node_type?: string
  name: string
  config?: any
  sort_order?: number
  timeout?: number
  retry_policy?: string
  max_retries?: number
}

export interface CreateEdgeRequest {
  source_node_id: number
  target_node_id: number
  condition?: string
  context_mapping?: any
}

export const workflowApi = {
  list(projectId: number) {
    return apiClient.get<{ data: Workflow[] }>(`/projects/${projectId}/workflows`)
  },

  create(projectId: number, data: CreateWorkflowRequest) {
    return apiClient.post<Workflow>(`/projects/${projectId}/workflows`, data)
  },

  get(projectId: number, workflowId: number) {
    return apiClient.get<WorkflowDetail>(`/projects/${projectId}/workflows/${workflowId}`)
  },

  update(projectId: number, workflowId: number, data: UpdateWorkflowRequest) {
    return apiClient.put(`/projects/${projectId}/workflows/${workflowId}`, data)
  },

  delete(projectId: number, workflowId: number) {
    return apiClient.delete(`/projects/${projectId}/workflows/${workflowId}`)
  },

  execute(projectId: number, workflowId: number, issueId?: number) {
    return apiClient.post<WorkflowRun>(`/projects/${projectId}/workflows/${workflowId}/execute`, {
      issue_id: issueId
    })
  },

  listRuns(projectId: number, workflowId: number) {
    return apiClient.get<{ data: WorkflowRun[] }>(`/projects/${projectId}/workflows/${workflowId}/runs`)
  },

  getRun(projectId: number, workflowId: number, runId: number) {
    return apiClient.get<WorkflowRunDetail>(`/projects/${projectId}/workflows/${workflowId}/runs/${runId}`)
  },

  cancelRun(projectId: number, workflowId: number, runId: number) {
    return apiClient.post(`/projects/${projectId}/workflows/${workflowId}/runs/${runId}/cancel`)
  },

  addNode(projectId: number, workflowId: number, data: CreateNodeRequest) {
    return apiClient.post<WorkflowNode>(`/projects/${projectId}/workflows/${workflowId}/nodes`, data)
  },

  updateNode(projectId: number, workflowId: number, nodeId: number, data: CreateNodeRequest) {
    return apiClient.put(`/projects/${projectId}/workflows/${workflowId}/nodes/${nodeId}`, data)
  },

  deleteNode(projectId: number, workflowId: number, nodeId: number) {
    return apiClient.delete(`/projects/${projectId}/workflows/${workflowId}/nodes/${nodeId}`)
  },

  addEdge(projectId: number, workflowId: number, data: CreateEdgeRequest) {
    return apiClient.post<WorkflowEdge>(`/projects/${projectId}/workflows/${workflowId}/edges`, data)
  },

  updateEdge(projectId: number, workflowId: number, edgeId: number, data: CreateEdgeRequest) {
    return apiClient.put(`/projects/${projectId}/workflows/${workflowId}/edges/${edgeId}`, data)
  },

  deleteEdge(projectId: number, workflowId: number, edgeId: number) {
    return apiClient.delete(`/projects/${projectId}/workflows/${workflowId}/edges/${edgeId}`)
  }
}

// ==================== Workspace-level Automation Stubs ====================

export function listWorkspaceWorkflows(workspaceId: number) {
  return apiClient.get(`/workspaces/${workspaceId}/workflows`).then(r => r.data)
}

export function listWorkspaceAutomations(workspaceId: number) {
  return apiClient.get(`/workspaces/${workspaceId}/automations`).then(r => r.data)
}

export function createWorkspaceAutomation(workspaceId: number, data: any) {
  return apiClient.post(`/workspaces/${workspaceId}/automations`, data).then(r => r.data)
}

export function updateWorkspaceAutomation(workspaceId: number, automationId: number, data: any) {
  return apiClient.put(`/workspaces/${workspaceId}/automations/${automationId}`, data).then(r => r.data)
}

export function deleteWorkspaceAutomation(workspaceId: number, automationId: number) {
  return apiClient.delete(`/workspaces/${workspaceId}/automations/${automationId}`)
}

export function createWorkspaceWorkflow(workspaceId: number, data: any) {
  return apiClient.post(`/workspaces/${workspaceId}/workflows`, data).then(r => r.data)
}

export function deleteWorkspaceWorkflow(workspaceId: number, workflowId: number) {
  return apiClient.delete(`/workspaces/${workspaceId}/workflows/${workflowId}`)
}

export function addWorkspaceTransition(workspaceId: number, workflowId: number, data: any) {
  return apiClient.post(`/workspaces/${workspaceId}/workflows/${workflowId}/edges`, data).then(r => r.data)
}

export function deleteWorkspaceTransition(workspaceId: number, workflowId: number, edgeId: number) {
  return apiClient.delete(`/workspaces/${workspaceId}/workflows/${workflowId}/edges/${edgeId}`)
}

// ==================== State Transition Stubs ====================

export function listStateTransitions(projectId: number, workflowId: number) {
  return apiClient.get(`/projects/${projectId}/workflows/${workflowId}/transitions`).then(r => r.data)
}

export function createStateTransition(projectId: number, workflowId: number, data: any) {
  return apiClient.post(`/projects/${projectId}/workflows/${workflowId}/transitions`, data).then(r => r.data)
}

export function updateStateTransition(projectId: number, workflowId: number, transitionId: number, data: any) {
  return apiClient.put(`/projects/${projectId}/workflows/${workflowId}/transitions/${transitionId}`, data).then(r => r.data)
}

export function deleteStateTransition(projectId: number, workflowId: number, transitionId: number) {
  return apiClient.delete(`/projects/${projectId}/workflows/${workflowId}/transitions/${transitionId}`)
}

// ==================== Automation Rule Stubs ====================

export function listAutomationRules(projectId: number) {
  return apiClient.get(`/projects/${projectId}/automation-rules`).then(r => r.data)
}

export function toggleAutomationRule(projectId: number, ruleId: number, enabled: boolean) {
  return apiClient.put(`/projects/${projectId}/automation-rules/${ruleId}`, { is_enabled: enabled }).then(r => r.data)
}

export function listAutomationTemplates() {
  return apiClient.get('/automation-templates').then(r => r.data)
}

// ==================== Budget API ====================

export interface ProjectBudget {
  id: number
  project_id: number
  monthly_budget: number
  current_cost: number
  alert_threshold: number // 0-100 (percentage)
  auto_block: boolean
  last_reset_at: string | null
  budget_usage: number // 0-100 (percentage)
}

export interface UpdateBudgetRequest {
  monthly_budget?: number
  alert_threshold?: number
  auto_block?: boolean
}

export const budgetApi = {
  get(projectId: number) {
    return apiClient.get<ProjectBudget>(`/projects/${projectId}/budget`)
  },

  update(projectId: number, data: UpdateBudgetRequest) {
    return apiClient.put(`/projects/${projectId}/budget`, data)
  },
}

// ==================== SLA API ====================

export interface SLAConfig {
  id: number
  project_id: number
  normal_task_max: number  // seconds
  complex_task_max: number // seconds
  auto_escalation: boolean
  enabled: boolean
}

export interface UpdateSLARequest {
  normal_task_max?: number
  complex_task_max?: number
  auto_escalation?: boolean
  enabled?: boolean
}

export const slaApi = {
  get(projectId: number) {
    return apiClient.get<SLAConfig>(`/projects/${projectId}/sla`)
  },

  update(projectId: number, data: UpdateSLARequest) {
    return apiClient.put(`/projects/${projectId}/sla`, data)
  },
}

// ==================== Decision API ====================

export interface DecisionRecord {
  id: number
  agent_id: number
  agent_name: string
  issue_id: number | null
  agent_task_id: number | null
  workflow_run_id: number | null
  node_type: string
  thinking: string
  decision: string
  reasoning: string
  alternatives: string[]
  confidence: number
  created_at: string
}

export const decisionApi = {
  list(projectId: number, limit = 100) {
    return apiClient.get<{ data: DecisionRecord[]; total: number }>(
      `/projects/${projectId}/decisions`,
      { params: { limit } }
    )
  },
}
