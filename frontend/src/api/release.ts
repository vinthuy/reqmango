import api from './index'
import type { Release, ReleaseCreateRequest, ReleaseUpdateRequest, ReleaseProgress, ReleaseIssueRequest } from '../types/release'

export const releaseApi = {
  list(projectId: number): Promise<Release[]> {
    return api.get(`/projects/${projectId}/releases`).then(res => res.data.data || [])
  },

  create(projectId: number, data: ReleaseCreateRequest): Promise<Release> {
    return api.post(`/projects/${projectId}/releases`, data)
  },

  get(projectId: number, releaseId: number): Promise<Release> {
    return api.get(`/projects/${projectId}/releases/${releaseId}`)
  },

  update(projectId: number, releaseId: number, data: ReleaseUpdateRequest): Promise<Release> {
    return api.put(`/projects/${projectId}/releases/${releaseId}`, data)
  },

  delete(projectId: number, releaseId: number): Promise<void> {
    return api.delete(`/projects/${projectId}/releases/${releaseId}`)
  },

  addIssues(projectId: number, releaseId: number, data: ReleaseIssueRequest): Promise<void> {
    return api.post(`/projects/${projectId}/releases/${releaseId}/issues`, data)
  },

  removeIssues(projectId: number, releaseId: number, data: ReleaseIssueRequest): Promise<void> {
    return api.delete(`/projects/${projectId}/releases/${releaseId}/issues`, { data })
  },

  getProgress(projectId: number, releaseId: number): Promise<ReleaseProgress> {
    return api.get(`/projects/${projectId}/releases/${releaseId}/progress`)
  }
}

export default releaseApi