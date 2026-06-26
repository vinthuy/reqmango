import api from './index'
import type { Workspace, WorkspaceCreate, WorkspaceMember } from '@/types'

export const workspaceApi = {
  create: async (data: WorkspaceCreate): Promise<Workspace> => {
    const response = await api.post('/workspaces', data)
    return response.data
  },
  
  list: async (): Promise<Workspace[]> => {
    const response = await api.get('/workspaces')
    return response.data
  },
  
  getBySlug: async (slug: string): Promise<Workspace> => {
    const response = await api.get(`/workspaces/${slug}`)
    return response.data
  },

  listMembers: async (wsParam: string): Promise<WorkspaceMember[]> => {
    const response = await api.get(`/workspaces/${wsParam}/members`)
    return response.data
  },

  addMember: async (wsParam: string, data: { user_id: number; role?: number }): Promise<WorkspaceMember> => {
    const response = await api.post(`/workspaces/${wsParam}/members`, data)
    return response.data
  },

  updateMember: async (wsParam: string, userId: number, role: number): Promise<WorkspaceMember> => {
    const response = await api.patch(`/workspaces/${wsParam}/members/${userId}?role=${role}`)
    return response.data
  },

  removeMember: async (wsParam: string, userId: number): Promise<void> => {
    await api.delete(`/workspaces/${wsParam}/members/${userId}`)
  }
}