export interface Initiative {
  id: number
  workspace_id: number
  name: string
  description?: string
  color?: string
  status: string
  target_date?: string
  start_date?: string
  sort_order: number
  projects?: any[]
  created_by?: any
  created_at: string
}

export interface InitiativeCreateRequest {
  name: string
  description?: string
  color?: string
  status?: string
  target_date?: string
  start_date?: string
  project_ids?: number[]
}

export interface InitiativeUpdateRequest {
  name?: string
  description?: string
  color?: string
  status?: string
  target_date?: string
  start_date?: string
  project_ids?: number[]
}

export interface InitiativeProgress {
  total_issues: number
  completed_issues: number
  progress: number
  project_count: number
}
