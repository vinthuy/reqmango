/**
 * Squad API — 多Agent协作团队 API 模块
 */
import api from './index'

export interface Squad {
  id: number
  workspace_id: number
  project_id?: number
  name: string
  description: string
  leader_agent_id?: number
  status: string
  goal: string
  members: SquadMember[]
  execution_count?: number
  created_at: string
  updated_at: string
}

export interface SquadMember {
  id: number
  squad_id: number
  agent_id: number
  role: string
  agent_config_id: number
  status: string
  assigned_at: string
  removed_at?: string
}

export interface SquadExecution {
  id: number
  squad_id: number
  status: string
  goal: string
  input_data?: any
  output_data?: any
  logs?: any[]
  started_at?: string
  completed_at?: string
  failed_at?: string
  error_info?: string
  created_at: string
}

export interface SquadCreateRequest {
  name: string
  description?: string
  leader_agent_id?: number
  project_id?: number
  goal?: string
  config?: Record<string, any>
  members?: { agent_id: number; role: string; agent_config_id: number }[]
}

export interface SquadUpdateRequest {
  name?: string
  description?: string
  leader_agent_id?: number
  goal?: string
  config?: Record<string, any>
}

export interface SquadMemberAddRequest {
  agent_id: number
  role: string
  agent_config_id: number
}

export interface SquadExecutionStartRequest {
  goal: string
  input_data?: Record<string, any>
}

export async function getSquads(workspaceId: number, projectId?: number): Promise<Squad[]> {
  const params: Record<string, any> = {}
  if (projectId) params.project_id = projectId
  const response = await api.get(`/workspaces/${workspaceId}/squads`, { params })
  return response.data
}

export async function listSquads(workspaceId: number): Promise<Squad[]> {
  const response = await api.get(`/workspaces/${workspaceId}/squads`)
  return response.data
}

export async function createSquad(workspaceId: number, data: SquadCreateRequest): Promise<Squad> {
  const response = await api.post(`/workspaces/${workspaceId}/squads`, data)
  return response.data
}

export async function getSquad(workspaceId: number, squadId: number): Promise<Squad> {
  const response = await api.get(`/workspaces/${workspaceId}/squads/${squadId}`)
  return response.data
}

export async function updateSquad(workspaceId: number, squadId: number, data: SquadUpdateRequest): Promise<Squad> {
  const response = await api.put(`/workspaces/${workspaceId}/squads/${squadId}`, data)
  return response.data
}

export async function deleteSquad(workspaceId: number, squadId: number): Promise<void> {
  await api.delete(`/workspaces/${workspaceId}/squads/${squadId}`)
}

export async function addMember(workspaceId: number, squadId: number, data: SquadMemberAddRequest): Promise<SquadMember> {
  const response = await api.post(`/workspaces/${workspaceId}/squads/${squadId}/members`, data)
  return response.data
}

export async function removeMember(workspaceId: number, squadId: number, memberId: number): Promise<void> {
  await api.delete(`/workspaces/${workspaceId}/squads/${squadId}/members/${memberId}`)
}

export async function startExecution(workspaceId: number, squadId: number, data: SquadExecutionStartRequest): Promise<SquadExecution> {
  const response = await api.post(`/workspaces/${workspaceId}/squads/${squadId}/executions`, data)
  return response.data
}

export async function getExecutions(workspaceId: number, squadId: number): Promise<SquadExecution[]> {
  const response = await api.get(`/workspaces/${workspaceId}/squads/${squadId}/executions`)
  return response.data
}

export async function getExecution(workspaceId: number, squadId: number, executionId: number): Promise<SquadExecution> {
  const response = await api.get(`/workspaces/${workspaceId}/squads/${squadId}/executions/${executionId}`)
  return response.data
}

export async function cancelExecution(workspaceId: number, squadId: number, executionId: number): Promise<void> {
  await api.delete(`/workspaces/${workspaceId}/squads/${squadId}/executions/${executionId}`)
}

export default {
  getSquads,
  listSquads,
  createSquad,
  getSquad,
  updateSquad,
  deleteSquad,
  addMember,
  removeMember,
  startExecution,
  getExecutions,
  getExecution,
  cancelExecution,
}
