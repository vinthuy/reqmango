/**
 * Page Types - 页面/文档类型定义
 */

export interface Page {
  id: number
  title: string
  content: string
  content_json?: string
  published: boolean
  archived_at?: string
  sequence: number
  parent_id?: number
  depth: number
  project_id: number
  workspace_id: number
  created_at: string
  updated_at: string
  children?: Page[]
}

export interface PageCreate {
  title: string
  content?: string
  content_json?: string
  parent_id?: number
  sequence?: number
}

export interface PageUpdate {
  title?: string
  content?: string
  content_json?: string
  published?: boolean
  sequence?: number
}

export interface PageMove {
  parent_id?: number | null
  sequence: number
}
