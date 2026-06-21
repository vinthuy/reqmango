/**
 * Cycle API - 周期 API 调用模块
 */
import api from './index'
import type {
  CycleCreate,
  CycleUpdate,
  CycleResponse,
  CycleProgress,
  CycleStatistics,
  BurndownData
} from '@/types/cycle'

// ==================== Cycle CRUD ====================

/**
 * 创建周期
 */
export async function createCycle(
  workspaceId: number,
  data: CycleCreate
): Promise<CycleResponse> {
  const response = await api.post(
    `/cycles?workspace_id=${workspaceId}`,
    data
  )
  return response.data
}

/**
 * 列出项目的周期
 */
export async function listCycles(
  projectId: number,
  workspaceId: number,
  options?: {
    status?: string
    include_completed?: boolean
    limit?: number
    offset?: number
  }
): Promise<CycleResponse[]> {
  const params = new URLSearchParams()
  params.append('workspace_id', workspaceId.toString())
  
  if (options?.status) params.append('status', options.status)
  if (options?.include_completed) params.append('include_completed', 'true')
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  
  const response = await api.get(`/projects/${projectId}/cycles?${params.toString()}`)
  return response.data
}

/**
 * 获取周期详情
 */
export async function getCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.get(`/cycles/${cycleId}`)
  return response.data
}

/**
 * 更新周期
 */
export async function updateCycle(
  cycleId: number,
  data: CycleUpdate
): Promise<CycleResponse> {
  const response = await api.put(`/cycles/${cycleId}`, data)
  return response.data
}

/**
 * 删除周期
 */
export async function deleteCycle(cycleId: number): Promise<void> {
  await api.delete(`/cycles/${cycleId}`)
}

// ==================== Cycle Status Management ====================

/**
 * 开始周期
 */
export async function startCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/start`)
  return response.data
}

/**
 * 结束周期
 */
export async function endCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/end`)
  return response.data
}

/**
 * 取消周期
 */
export async function cancelCycle(cycleId: number): Promise<CycleResponse> {
  const response = await api.post(`/cycles/${cycleId}/cancel`)
  return response.data
}

// ==================== Cycle Issue Management ====================

/**
 * 将工作项添加到周期
 */
export async function addIssueToCycle(
  cycleId: number,
  issueId: number
): Promise<{ cycle_id: number; issue_id: number; action: string }> {
  const response = await api.post(`/cycles/${cycleId}/issues?issue_id=${issueId}`)
  return response.data
}

/**
 * 从周期移除工作项
 */
export async function removeIssueFromCycle(
  cycleId: number,
  issueId: number
): Promise<{ cycle_id: number; issue_id: number; action: string }> {
  const response = await api.delete(`/cycles/${cycleId}/issues/${issueId}`)
  return response.data
}

/**
 * 获取周期内的工作项
 */
export async function getCycleIssues(
  cycleId: number,
  options?: {
    state_id?: number
    priority?: string
    limit?: number
    offset?: number
  }
): Promise<any[]> {
  const params = new URLSearchParams()
  
  if (options?.state_id) params.append('state_id', options.state_id.toString())
  if (options?.priority) params.append('priority', options.priority)
  if (options?.limit) params.append('limit', options.limit.toString())
  if (options?.offset) params.append('offset', options.offset.toString())
  
  const response = await api.get(`/cycles/${cycleId}/issues?${params.toString()}`)
  return response.data
}

// ==================== Cycle Progress & Statistics ====================

/**
 * 获取周期进度
 */
export async function getCycleProgress(cycleId: number): Promise<CycleProgress> {
  const response = await api.get(`/cycles/${cycleId}/progress`)
  return response.data
}

/**
 * 获取周期详细统计
 */
export async function getCycleStatistics(cycleId: number): Promise<CycleStatistics> {
  const response = await api.get(`/cycles/${cycleId}/statistics`)
  return response.data
}

/**
 * 获取燃尽图数据
 */
export async function getBurndownData(cycleId: number): Promise<BurndownData> {
  const response = await api.get(`/cycles/${cycleId}/burndown`)
  return response.data
}

// ==================== Export all ====================

export const cycleApi = {
  // CRUD
  createCycle,
  listCycles,
  getCycle,
  updateCycle,
  deleteCycle,
  
  // Status
  startCycle,
  endCycle,
  cancelCycle,
  
  // Issues
  addIssueToCycle,
  removeIssueFromCycle,
  getCycleIssues,
  
  // Statistics
  getCycleProgress,
  getCycleStatistics,
  getBurndownData
}

export default cycleApi