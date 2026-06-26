/**
 * Page API - 页面API模块
 */
import api from './index'
import type { Page, PageCreate, PageUpdate, PageMove } from '@/types/page'

/**
 * List pages for a project.
 */
export async function listPages(projectId: number, includeArchived = false): Promise<Page[]> {
  const response = await api.get(`/projects/${projectId}/pages`, {
    params: { include_archived: includeArchived }
  })
  return response.data
}

/**
 * Get page tree for a project.
 */
export async function getPageTree(projectId: number): Promise<Page[]> {
  const response = await api.get(`/projects/${projectId}/pages/tree`)
  return response.data
}

/**
 * Get a single page.
 */
export async function getPage(projectId: number, pageId: number): Promise<Page> {
  const response = await api.get(`/projects/${projectId}/pages/${pageId}`)
  return response.data
}

/**
 * Create a new page.
 */
export async function createPage(projectId: number, workspaceId: number, data: PageCreate): Promise<Page> {
  const response = await api.post(`/projects/${projectId}/pages?workspace_id=${workspaceId}`, data)
  return response.data
}

/**
 * Update a page.
 */
export async function updatePage(projectId: number, pageId: number, data: PageUpdate): Promise<Page> {
  const response = await api.put(`/projects/${projectId}/pages/${pageId}`, data)
  return response.data
}

/**
 * Delete a page.
 */
export async function deletePage(projectId: number, pageId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/pages/${pageId}`)
}

/**
 * Archive a page.
 */
export async function archivePage(projectId: number, pageId: number): Promise<void> {
  await api.post(`/projects/${projectId}/pages/${pageId}/archive`)
}

/**
 * Restore a page.
 */
export async function restorePage(projectId: number, pageId: number): Promise<void> {
  await api.post(`/projects/${projectId}/pages/${pageId}/restore`)
}

/**
 * Move a page.
 */
export async function movePage(projectId: number, pageId: number, data: PageMove): Promise<Page> {
  const response = await api.post(`/projects/${projectId}/pages/${pageId}/move`, data)
  return response.data
}

/**
 * Get children of a page.
 */
export async function listPageChildren(projectId: number, pageId: number): Promise<Page[]> {
  const response = await api.get(`/projects/${projectId}/pages/${pageId}/children`)
  return response.data
}

export async function searchPages(projectId: number, query: string): Promise<Page[]> {
  const response = await api.get(`/projects/${projectId}/pages`, {
    params: { search: query }
  })
  return response.data
}

export default {
  listPages,
  getPageTree,
  getPage,
  createPage,
  updatePage,
  deletePage,
  archivePage,
  restorePage,
  movePage,
  listPageChildren,
}
