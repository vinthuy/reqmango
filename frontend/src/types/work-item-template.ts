/**
 * Work Item Template Types - 工作项模板类型定义
 */
import type { IssuePriority } from './issue'

export interface WorkItemTemplateDefaults {
  name_prefix?: string
  priority?: IssuePriority
  state_id?: number
  assignee_ids?: number[]
  label_ids?: number[]
  description_html?: string
}

export interface WorkItemTemplate {
  id: number
  name: string
  description?: string
  issue_type_id?: number
  issue_type?: {
    id: number
    name: string
    color: string
    icon: string
  }
  defaults: WorkItemTemplateDefaults
  is_default: boolean
  project_id: number
  workspace_id: number
  created_at: string
  updated_at: string
}

export interface WorkItemTemplateCreate {
  name: string
  description?: string
  issue_type_id?: number
  is_default?: boolean
  defaults: WorkItemTemplateDefaults
}

export interface WorkItemTemplateUpdate {
  name?: string
  description?: string
  issue_type_id?: number
  is_default?: boolean
  defaults?: WorkItemTemplateDefaults
}
