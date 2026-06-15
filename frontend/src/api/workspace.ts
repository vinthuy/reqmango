import api from './index'
import type { Workspace, WorkspaceCreate } from '@/types'

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
  }
}