/**
 * Saved View Types - 保存视图类型定义
 */

export interface SavedView {
  id: number
  name: string
  description?: string
  view_type: 'list' | 'kanban'
  filters: Record<string, any>
  sort_config: SortConfigEntry[]
  columns: string[]
  group_by?: string
  is_default: boolean
  is_shared: boolean
  owner_id: number
  project_id: number
  created_at: string
  updated_at: string
}

export interface SortConfigEntry {
  field: string
  dir: 'asc' | 'desc'
}

export interface SavedViewCreate {
  name: string
  description?: string
  view_type?: 'list' | 'kanban'
  filters?: Record<string, any>
  sort_config?: SortConfigEntry[]
  columns?: string[]
  group_by?: string
  is_shared?: boolean
}

export interface SavedViewUpdate {
  name?: string
  description?: string
  view_type?: 'list' | 'kanban'
  filters?: Record<string, any>
  sort_config?: SortConfigEntry[]
  columns?: string[]
  group_by?: string
  is_shared?: boolean
}
