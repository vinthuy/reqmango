/**
 * Workflow API - 工作流和自动化 API 调用模块
 */
import api from './index'
import type {
  StateTransition,
  StateTransitionCreate,
  StateTransitionUpdate,
  AutomationRule,
  AutomationRuleCreate,
  AutomationRuleUpdate,
  AutomationRuleLite,
  AutomationExecutionLog,
  AutomationTemplate
} from '@/types/workflow'

// ==================== State Transition API ====================

/**
 * 创建状态转换规则
 */
export async function createStateTransition(
  projectId: string,
  workspaceId: string,
  data: StateTransitionCreate
): Promise<StateTransition> {
  const response = await api.post(
    `/projects/${projectId}/workflow/transitions?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出状态转换规则
 */
export async function listStateTransitions(
  projectId: string,
  sourceStateId?: string
): Promise<StateTransition[]> {
  const params = new URLSearchParams()
  if (sourceStateId) params.append('source_state_id', sourceStateId)
  
  const response = await api.get(
    `/projects/${projectId}/workflow/transitions?${params.toString()}`
  )
  return response.data
}

/**
 * 获取状态转换规则详情
 */
export async function getStateTransition(
  transitionId: string
): Promise<StateTransition> {
  const response = await api.get(
    `/projects/workflow/transitions/${transitionId}`
  )
  return response.data
}

/**
 * 更新状态转换规则
 */
export async function updateStateTransition(
  transitionId: string,
  data: StateTransitionUpdate
): Promise<StateTransition> {
  const response = await api.put(
    `/projects/workflow/transitions/${transitionId}`,
    data
  )
  return response.data
}

/**
 * 删除状态转换规则
 */
export async function deleteStateTransition(
  transitionId: string
): Promise<void> {
  await api.delete(`/projects/workflow/transitions/${transitionId}`)
}

/**
 * 检查是否允许状态转换
 */
export async function checkCanTransition(
  projectId: string,
  sourceStateId: string,
  targetStateId: string,
  issueTypeId?: string
): Promise<boolean> {
  const params = new URLSearchParams()
  params.append('project_id', projectId)
  params.append('source_state_id', sourceStateId)
  params.append('target_state_id', targetStateId)
  if (issueTypeId) params.append('issue_type_id', issueTypeId)
  
  const response = await api.get(
    `/projects/${projectId}/workflow/transitions/can-transition?${params.toString()}`
  )
  return response.data
}

// ==================== Automation Rule API ====================

/**
 * 创建自动化规则
 */
export async function createAutomationRule(
  projectId: string,
  workspaceId: string,
  data: AutomationRuleCreate
): Promise<AutomationRule> {
  const response = await api.post(
    `/projects/${projectId}/workflow/automations?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出自动化规则
 */
export async function listAutomationRules(
  projectId: string,
  enabledOnly?: boolean
): Promise<AutomationRule[]> {
  const params = new URLSearchParams()
  if (enabledOnly) params.append('enabled_only', 'true')
  
  const response = await api.get(
    `/projects/${projectId}/workflow/automations?${params.toString()}`
  )
  return response.data
}

/**
 * 列出自动化规则（轻量版）
 */
export async function listAutomationRulesLite(
  projectId: string
): Promise<AutomationRuleLite[]> {
  const response = await api.get(
    `/projects/${projectId}/workflow/automations/lite`
  )
  return response.data
}

/**
 * 获取自动化规则详情
 */
export async function getAutomationRule(
  ruleId: string
): Promise<AutomationRule> {
  const response = await api.get(
    `/projects/workflow/automations/${ruleId}`
  )
  return response.data
}

/**
 * 更新自动化规则
 */
export async function updateAutomationRule(
  ruleId: string,
  data: AutomationRuleUpdate
): Promise<AutomationRule> {
  const response = await api.put(
    `/projects/workflow/automations/${ruleId}`,
    data
  )
  return response.data
}

/**
 * 删除自动化规则
 */
export async function deleteAutomationRule(
  ruleId: string
): Promise<void> {
  await api.delete(`/projects/workflow/automations/${ruleId}`)
}

/**
 * 启用/禁用自动化规则
 */
export async function toggleAutomationRule(
  ruleId: string,
  enabled: boolean
): Promise<AutomationRule> {
  const response = await api.post(
    `/projects/workflow/automations/${ruleId}/toggle?enabled=${enabled}`
  )
  return response.data
}

// ==================== Automation Execution Log API ====================

/**
 * 获取自动化规则执行日志
 */
export async function listAutomationLogs(
  ruleId: string,
  limit?: number
): Promise<AutomationExecutionLog[]> {
  const params = new URLSearchParams()
  if (limit) params.append('limit', limit.toString())
  
  const response = await api.get(
    `/projects/workflow/automations/${ruleId}/logs?${params.toString()}`
  )
  return response.data
}

// ==================== Automation Template API ====================

/**
 * 列出自动化模板
 */
export async function listAutomationTemplates(
  category?: string
): Promise<AutomationTemplate[]> {
  const params = new URLSearchParams()
  if (category) params.append('category', category)
  
  const response = await api.get(
    `/projects/workflow/templates?${params.toString()}`
  )
  return response.data
}

/**
 * 从模板创建自动化规则
 */
export async function applyAutomationTemplate(
  projectId: string,
  workspaceId: string,
  templateId: string,
  nameOverride?: string
): Promise<AutomationRule> {
  const params = new URLSearchParams()
  params.append('project_id', projectId)
  params.append('workspace_id', workspaceId)
  if (nameOverride) params.append('name_override', nameOverride)
  
  const response = await api.post(
    `/projects/workflow/templates/${templateId}/apply?${params.toString()}`
  )
  return response.data
}

// ==================== Export ====================

export const workflowApi = {
  // State Transition
  createStateTransition,
  listStateTransitions,
  getStateTransition,
  updateStateTransition,
  deleteStateTransition,
  checkCanTransition,
  
  // Automation Rule
  createAutomationRule,
  listAutomationRules,
  listAutomationRulesLite,
  getAutomationRule,
  updateAutomationRule,
  deleteAutomationRule,
  toggleAutomationRule,
  
  // Execution Log
  listAutomationLogs,
  
  // Template
  listAutomationTemplates,
  applyAutomationTemplate
}

export default workflowApi