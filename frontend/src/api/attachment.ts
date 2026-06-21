/**
 * Attachment API - 附件API模块
 */
import api from './index'
import type { Attachment, AttachmentCreate, AttachmentUpdate } from '@/types/attachment'

const BASE_URL = '/api/v1/attachments'

/**
 * 上传附件
 */
export async function uploadAttachment(
  file: File,
  data: AttachmentCreate & { name?: string }
): Promise<Attachment> {
  const formData = new FormData()
  formData.append('file', file)
  if (data.name) formData.append('name', data.name)
  if (data.issue_id) formData.append('issue_id', data.issue_id.toString())
  if (data.project_id) formData.append('project_id', data.project_id.toString())
  if (data.is_protected !== undefined) formData.append('is_protected', String(data.is_protected))

  const response = await api.post(BASE_URL, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
  return response.data
}

/**
 * 获取工作项的附件列表
 */
export async function listIssueAttachments(
  issueId: number
): Promise<Attachment[]> {
  const response = await api.get(`${BASE_URL}/issue/${issueId}`)
  return response.data
}

/**
 * 获取项目的附件列表
 */
export async function listProjectAttachments(
  projectId: number
): Promise<Attachment[]> {
  const response = await api.get(`${BASE_URL}/project/${projectId}`)
  return response.data
}

/**
 * 获取附件详情
 */
export async function getAttachment(
  attachmentId: number
): Promise<Attachment> {
  const response = await api.get(`${BASE_URL}/${attachmentId}`)
  return response.data
}

/**
 * 下载附件
 */
export async function downloadAttachment(
  attachmentId: number
): Promise<Blob> {
  const response = await api.get(`${BASE_URL}/${attachmentId}/download`, {
    responseType: 'blob'
  })
  return response.data
}

/**
 * 更新附件
 */
export async function updateAttachment(
  attachmentId: number,
  data: AttachmentUpdate
): Promise<Attachment> {
  const response = await api.patch(`${BASE_URL}/${attachmentId}`, data)
  return response.data
}

/**
 * 删除附件
 */
export async function deleteAttachment(
  attachmentId: number
): Promise<void> {
  await api.delete(`${BASE_URL}/${attachmentId}`)
}

/**
 * 获取附件下载URL
 */
export function getAttachmentDownloadUrl(attachmentId: number): string {
  return `/api/v1/attachments/${attachmentId}/download`
}

export default {
  uploadAttachment,
  listIssueAttachments,
  listProjectAttachments,
  getAttachment,
  downloadAttachment,
  updateAttachment,
  deleteAttachment,
  getAttachmentDownloadUrl
}
