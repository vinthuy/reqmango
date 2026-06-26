export interface Release {
  id: number
  name: string
  version: string
  description: string
  status: 'planned' | 'in_progress' | 'released' | 'cancelled'
  release_date: string | null
  project_id: number
  created_at: string
  updated_at: string
  progress?: number
}

export interface ReleaseProgress {
  id: number
  name: string
  total_issues: number
  done_issues: number
  progress: number
}

export interface ReleaseCreateRequest {
  name: string
  version: string
  description?: string
  status?: string
  release_date?: string
}

export interface ReleaseUpdateRequest {
  name?: string
  version?: string
  description?: string
  status?: string
  release_date?: string
}

export interface ReleaseIssueRequest {
  issue_ids: number[]
}