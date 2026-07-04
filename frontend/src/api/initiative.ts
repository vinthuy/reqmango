import api from './index'
import type { Initiative, InitiativeCreateRequest, InitiativeUpdateRequest, InitiativeProgress } from '@/types/initiative'

export type { Initiative, InitiativeCreateRequest, InitiativeUpdateRequest, InitiativeProgress }

export async function listInitiatives(workspaceId: number): Promise<Initiative[]> {
  const r = await api.get(`/workspaces/${workspaceId}/initiatives`); return r.data.data || []
}

export async function getInitiative(id: number): Promise<Initiative> {
  const r = await api.get(`/initiatives/${id}`); return r.data.data
}

export async function createInitiative(workspaceId: number, data: Partial<Initiative>): Promise<Initiative> {
  const r = await api.post(`/workspaces/${workspaceId}/initiatives`, data); return r.data.data
}

export async function updateInitiative(id: number, data: Partial<Initiative>): Promise<Initiative> {
  const r = await api.put(`/initiatives/${id}`, data); return r.data.data
}

export async function deleteInitiative(id: number): Promise<void> {
  await api.delete(`/initiatives/${id}`)
}

export async function getInitiativeProgress(id: number): Promise<any> {
  const r = await api.get(`/initiatives/${id}/progress`); return r.data.data
}

export const initiativeApi = { list: listInitiatives, get: getInitiative, create: createInitiative, update: updateInitiative, delete: deleteInitiative, getProgress: getInitiativeProgress }
