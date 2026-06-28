import api from './index'

export interface ProjectUpdate {
  id: number; project_id: number; author_id: number
  author?: { id: number; display_name: string; avatar_url?: string }
  status: 'on_track' | 'at_risk' | 'off_track'
  content: string; created_at: string
}

export async function listUpdates(projectId: number, limit = 20): Promise<ProjectUpdate[]> {
  const r = await api.get(`/projects/${projectId}/updates?limit=${limit}`); return r.data.data || []
}

export async function createUpdate(projectId: number, status: string, content: string): Promise<ProjectUpdate> {
  const r = await api.post(`/projects/${projectId}/updates`, { status, content }); return r.data.data
}

export const projectUpdateApi = { list: listUpdates, create: createUpdate }
