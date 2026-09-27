/**
 * Page Types - 页面/文档类型定义
 */

export interface Page {
  id: number
  title: string
  content: string
  content_json?: string | object
  published: boolean
  archived_at?: string
  sequence: number
  parent_id?: number
  depth: number
  locked_by_id?: number
  locked_at?: string
  locked_by_name?: string
  project_id: number
  workspace_id: number
  /** Provenance of a generated page, e.g. { source_type: 'cycle', source_id: 12 }. */
  source_type?: string
  source_id?: number
  created_by_id?: number
  updated_by_id?: number
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
  /** Set when the page is generated from something, so it can link back. */
  source_type?: PageSourceType
  source_id?: number
}

export type PageSourceType = 'cycle' | 'issue'

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

export interface PageVersion {
  id: number
  page_id: number
  title: string
  content: string
  content_json?: string
  version_number: number
  change_summary?: string
  created_at: string
  created_by_id?: number
  created_by_name?: string
}

export interface PageTemplate {
  id: number
  name: string
  description: string
  content: string
  content_json?: string
  is_default: boolean
  workspace_id: number
  project_id?: number
  created_at: string
  updated_at: string
}

export interface PageTemplateCreate {
  name: string
  description?: string
  content?: string
  content_json?: string
  is_default?: boolean
  project_id?: number
}

export interface PageTemplateUpdate {
  name?: string
  description?: string
  content?: string
  content_json?: string
  is_default?: boolean
}
