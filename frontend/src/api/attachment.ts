/**
 * Attachment API - 附件API模块
 */
import api from './index'
import type { Attachment } from '@/types/attachment'

/**
 * 获取工作项的附件列表
 */
export async function listIssueAttachments(
  issueId: number
): Promise<Attachment[]> {
  const response = await api.get(`/issues/${issueId}/attachments`)
  return response.data
}

/**
 * 获取附件详情
 */
export async function getAttachment(
  issueId: number,
  attachmentId: number
): Promise<Attachment> {
  const response = await api.get(`/issues/${issueId}/attachments/${attachmentId}`)
  return response.data
}

/**
 * 上传附件
 */
export async function uploadAttachment(
  issueId: number,
  file: File
): Promise<Attachment> {
  const formData = new FormData()
  formData.append('file', file)

  const response = await api.post(`/issues/${issueId}/attachments`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
  return response.data
}

/**
 * 删除附件
 */
export async function deleteAttachment(
  issueId: number,
  attachmentId: number
): Promise<void> {
  await api.delete(`/issues/${issueId}/attachments/${attachmentId}`)
}

/**
 * 获取附件下载URL
 */
export function getAttachmentDownloadUrl(issueId: number, attachmentId: number): string {
  return `/api/v1/issues/${issueId}/attachments/${attachmentId}`
}

export default {
  listIssueAttachments,
  getAttachment,
  uploadAttachment,
  deleteAttachment,
  getAttachmentDownloadUrl
}