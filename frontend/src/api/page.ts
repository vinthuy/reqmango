/**
 * Page API - 页面API模块
 */
import api from './index'
import type { Page, PageCreate, PageUpdate, PageMove, PageVersion, PageTemplate, PageTemplateCreate, PageTemplateUpdate } from '@/types/page'

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

// ============ Locking ============

/**
 * Lock a page for editing.
 */
export async function lockPage(projectId: number, pageId: number): Promise<Page> {
  const response = await api.post(`/projects/${projectId}/pages/${pageId}/lock`)
  return response.data
}

/**
 * Unlock a page.
 */
export async function unlockPage(projectId: number, pageId: number): Promise<Page> {
  const response = await api.post(`/projects/${projectId}/pages/${pageId}/unlock`)
  return response.data
}

// ============ Export ============

/**
 * Export a page in the specified format (via blob download).
 */
export async function exportPage(projectId: number, pageId: number, format: 'md' | 'html' | 'txt' = 'md'): Promise<void> {
  const response = await api.get(`/projects/${projectId}/pages/${pageId}/export`, {
    params: { format },
    responseType: 'blob',
  })
  const contentDisposition = response.headers['content-disposition']
  let filename = `page.${format}`
  if (contentDisposition) {
    const match = contentDisposition.match(/filename="?([^";]+)"?/)
    if (match) filename = match[1]
  }
  const url = window.URL.createObjectURL(new Blob([response.data]))
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  window.URL.revokeObjectURL(url)
}

// ============ Convert to Issue ============

/**
 * Convert a page to an issue.
 */
export async function convertPageToIssue(projectId: number, pageId: number, issueTypeId?: number): Promise<{ message: string; issue: any }> {
  const response = await api.post(`/projects/${projectId}/pages/${pageId}/convert-to-issue`, {
    issue_type_id: issueTypeId
  })
  return response.data
}

// ============ Version History ============

/**
 * Get version history for a page.
 */
export async function getPageVersions(projectId: number, pageId: number): Promise<PageVersion[]> {
  const response = await api.get(`/projects/${projectId}/pages/${pageId}/versions`)
  return response.data
}

/**
 * Get a specific version of a page.
 */
export async function getPageVersion(projectId: number, pageId: number, versionNumber: number): Promise<PageVersion> {
  const response = await api.get(`/projects/${projectId}/pages/${pageId}/versions/${versionNumber}`)
  return response.data
}

/**
 * Restore a page to a previous version.
 */
export async function restorePageVersion(projectId: number, pageId: number, versionNumber: number): Promise<{ message: string }> {
  const response = await api.post(`/projects/${projectId}/pages/${pageId}/versions/${versionNumber}/restore`)
  return response.data
}

// ============ Templates ============

/**
 * List page templates.
 */
export async function listPageTemplates(projectId: number, workspaceId: number): Promise<PageTemplate[]> {
  const response = await api.get(`/projects/${projectId}/page-templates`, {
    params: { workspace_id: workspaceId, project_id: projectId }
  })
  return response.data
}

/**
 * Get a page template.
 */
export async function getPageTemplate(projectId: number, templateId: number): Promise<PageTemplate> {
  const response = await api.get(`/projects/${projectId}/page-templates/${templateId}`)
  return response.data
}

/**
 * Create a page template.
 */
export async function createPageTemplate(projectId: number, workspaceId: number, data: PageTemplateCreate): Promise<PageTemplate> {
  const response = await api.post(`/projects/${projectId}/page-templates?workspace_id=${workspaceId}`, data)
  return response.data
}

/**
 * Update a page template.
 */
export async function updatePageTemplate(projectId: number, templateId: number, data: PageTemplateUpdate): Promise<PageTemplate> {
  const response = await api.put(`/projects/${projectId}/page-templates/${templateId}`, data)
  return response.data
}

/**
 * Delete a page template.
 */
export async function deletePageTemplate(projectId: number, templateId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/page-templates/${templateId}`)
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
  lockPage,
  unlockPage,
  exportPage,
  convertPageToIssue,
  getPageVersions,
  getPageVersion,
  restorePageVersion,
  listPageTemplates,
  getPageTemplate,
  createPageTemplate,
  updatePageTemplate,
  deletePageTemplate,
}
