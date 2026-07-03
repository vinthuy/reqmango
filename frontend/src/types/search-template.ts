export interface SearchTemplate {
  id: number
  name: string
  description?: string
  icon: string
  rql_template: string
  view_type: 'list' | 'kanban' | 'tree' | 'gantt' | 'calendar'
  sort_config: SortConfigEntry[]
  group_by?: string
  columns: string[]
  is_built_in: boolean
  is_public: boolean
  owner_id: number
  project_id: number
  created_at: string
  updated_at: string
}

export interface SortConfigEntry {
  field: string
  dir: 'asc' | 'desc'
}

export interface SearchTemplateCreate {
  name: string
  description?: string
  icon?: string
  rql_template?: string
  view_type?: 'list' | 'kanban' | 'tree' | 'gantt' | 'calendar'
  sort_config?: SortConfigEntry[]
  group_by?: string
  columns?: string[]
}