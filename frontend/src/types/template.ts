export interface ProjectTemplate {
  id: number
  name: string
  description: string
  workspace_id: number
  is_default: boolean
  types: ProjectTemplateType[]
  created_at: string
  updated_at: string
}

export interface ProjectTemplateType {
  template_id: number
  issue_type_id: number
  is_required: boolean
  default_state_id?: number
  sequence: number
  type_name?: string
  type_color?: string
  type_icon?: string
}

export interface ProjectTemplateCreate {
  name: string
  description?: string
  is_default?: boolean
}

export interface ProjectTemplateUpdate {
  name?: string
  description?: string
  is_default?: boolean
}

export interface TemplateTypeAdd {
  issue_type_id: number
  is_required?: boolean
  default_state_id?: number
  sequence?: number
}
