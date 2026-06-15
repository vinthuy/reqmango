/**
 * Attachment Types - 附件类型定义
 */

export interface Attachment {
  id: string
  name: string
  file_name: string
  file_size: number
  mime_type: string
  file_url?: string
  is_protected: boolean
  issue_id?: string
  project_id?: string
  uploaded_by_id: string
  access_url?: string
  thumbnail_url?: string
  created_at: string
  updated_at: string
}

export interface AttachmentCreate {
  name?: string
  issue_id?: string
  project_id?: string
  is_protected?: boolean
}

export interface AttachmentUpdate {
  name?: string
  is_protected?: boolean
}

export interface AttachmentUploadResponse {
  id: string
  name: string
  file_url: string
  file_size: number
  mime_type: string
}

export interface AttachmentListResponse {
  items: Attachment[]
  total: number
}
