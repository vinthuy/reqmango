/**
 * Cycle API - 周期 API 调用模块
 */
import api from './index'
import type { CycleCreate, CycleUpdate, CycleResponse, CycleProgress, CycleStatistics, BurndownData } from '@/types/cycle'

// ==================== CRUD ====================

export async function createCycle(projectId: number, workspaceId: number, data: CycleCreate): Promise<CycleResponse> {
  const response = await api.post(`/projects/${projectId}/cycles?workspace_id=${workspaceId}`, data)
  return response.data
}

export async function listCycles(projectId: number, options?: { status?: string; limit?: number; offset?: number }): Promise<{ items: CycleResponse[]; total: number; limit: number; offset: number }> {
  const params = new URLSearchParams()
  if (options?.status) params.append('status', options.status)
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  const response = await api.get(`/projects/${projectId}/cycles?${params.toString()}`)
  return response.data
}

export async function getCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.get(`/cycles/${cycleId}`)
  return response.data
}

export async function updateCycle(cycleId: number, data: CycleUpdate): Promise<CycleResponse> {
  const response = await api.put(`/cycles/${cycleId}`, data)
  return response.data
}

export async function deleteCycle(cycleId: number): Promise<void> {
  await api.delete(`/cycles/${cycleId}`)
}

// ==================== Status Transitions ====================

export async function startCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/start`)
  return response.data
}

export async function endCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/end`)
  return response.data
}

export async function cancelCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/cancel`)
  return response.data
}

// ==================== Issue Association ====================

export async function addIssueToCycle(cycleId: number, issueId: number): Promise<{ cycle_id: number; issue_id: number; action: string }> {
  const response = await api.post(`/cycles/${cycleId}/issues?issue_id=${issueId}`)
  return response.data
}

export async function removeIssueFromCycle(cycleId: number, issueId: number): Promise<{ cycle_id: number; issue_id: number; action: string }> {
  const response = await api.delete(`/cycles/${cycleId}/issues/${issueId}`)
  return response.data
}

export async function getCycleIssues(cycleId: number, options?: { state_id?: number; priority?: string; limit?: number; offset?: number }): Promise<any[]> {
  const params = new URLSearchParams()
  if (options?.state_id) params.append('state_id', options.state_id.toString())
  if (options?.priority) params.append('priority', options.priority)
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  const response = await api.get(`/cycles/${cycleId}/issues?${params.toString()}`)
  return response.data
}

// ==================== Analysis ====================

export async function getCycleProgress(cycleId: number): Promise<CycleProgress> {
  const response = await api.get(`/cycles/${cycleId}/progress`)
  return response.data
}

export async function getCycleStatistics(cycleId: number): Promise<CycleStatistics> {
  const response = await api.get(`/cycles/${cycleId}/statistics`)
  return response.data
}

export async function getBurndownData(cycleId: number): Promise<BurndownData> {
  const response = await api.get(`/cycles/${cycleId}/burndown`)
  return response.data
}

// ==================== Automation ====================

export async function applyAutoAddRules(cycleId: number): Promise<void> {
  await api.post(`/cycles/${cycleId}/apply-auto-add`)
}

export async function applyAutoCloseRules(cycleId: number): Promise<void> {
  await api.post(`/cycles/${cycleId}/apply-auto-close`)
}

// ==================== Export all ====================

export const cycleApi = {
  createCycle, listCycles, getCycle, updateCycle, deleteCycle,
  startCycle, endCycle, cancelCycle,
  addIssueToCycle, removeIssueFromCycle, getCycleIssues,
  getCycleProgress, getCycleStatistics, getBurndownData,
  applyAutoAddRules, applyAutoCloseRules
}

export default cycleApi
