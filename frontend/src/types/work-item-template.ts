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
  defaults: Record<string, any>
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
  defaults?: Record<string, any>
  is_default?: boolean
}

export interface WorkItemTemplateUpdate {
  name?: string
  description?: string
  issue_type_id?: number
  defaults?: Record<string, any>
  is_default?: boolean
}
