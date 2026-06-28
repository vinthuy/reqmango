export interface Permission {
  id: number
  code: string
  name: string
  description: string
  resource: string
  action: string
  scope: 'workspace' | 'project'
}

export interface Role {
  id: number
  name: string
  description: string
  scope: 'workspace' | 'project'
  workspace_id: number | null
  project_id: number | null
  is_system: boolean
  sort_order: number
  level: number
  permissions: Permission[]
  created_at: string
}

export interface CreateRoleRequest {
  name: string
  description?: string
  scope: 'workspace' | 'project'
  workspace_id?: number
  project_id?: number
  level?: number
  permissions: number[]
}

export interface UpdateRoleRequest {
  name?: string
  description?: string
  level?: number
  permissions?: number[]
}
