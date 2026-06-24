/**
 * Saved View API - 保存视图API模块
 */
import api from './index'
import type { SavedView, SavedViewCreate, SavedViewUpdate } from '@/types/saved-view'

/**
 * List saved views for a project.
 */
export async function listSavedViews(projectId: number): Promise<SavedView[]> {
  const response = await api.get(`/projects/${projectId}/views`)
  return response.data
}

/**
 * Get a single saved view.
 */
export async function getSavedView(projectId: number, viewId: number): Promise<SavedView> {
  const response = await api.get(`/projects/${projectId}/views/${viewId}`)
  return response.data
}

/**
 * Create a new saved view.
 */
export async function createSavedView(projectId: number, data: SavedViewCreate): Promise<SavedView> {
  const response = await api.post(`/projects/${projectId}/views`, data)
  return response.data
}

/**
 * Update a saved view.
 */
export async function updateSavedView(projectId: number, viewId: number, data: SavedViewUpdate): Promise<SavedView> {
  const response = await api.put(`/projects/${projectId}/views/${viewId}`, data)
  return response.data
}

/**
 * Delete a saved view.
 */
export async function deleteSavedView(projectId: number, viewId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/views/${viewId}`)
}

/**
 * Set a saved view as default.
 */
export async function setDefaultView(projectId: number, viewId: number): Promise<SavedView> {
  const response = await api.post(`/projects/${projectId}/views/${viewId}/set-default`)
  return response.data
}

/**
 * Duplicate a saved view.
 */
export async function duplicateSavedView(projectId: number, viewId: number): Promise<SavedView> {
  const response = await api.post(`/projects/${projectId}/views/${viewId}/duplicate`)
  return response.data
}

export default {
  listSavedViews,
  getSavedView,
  createSavedView,
  updateSavedView,
  deleteSavedView,
  setDefaultView,
  duplicateSavedView,
}
