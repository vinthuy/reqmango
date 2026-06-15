export interface User {
  id: string
  email: string
  username: string
  display_name: string
  first_name?: string
  last_name?: string
  avatar_url?: string
  cover_image_url?: string
  is_active: boolean
  is_email_verified: boolean
  user_timezone: string
  last_active?: string
  created_at: string
  updated_at: string
  created_by?: string
  updated_by?: string
  deleted_at?: string
  is_deleted: boolean
}

export interface LoginRequest {
  email: string
  password: string
}

export interface UserCreate {
  email: string
  username: string
  password: string
  display_name?: string
  first_name?: string
  last_name?: string
  user_timezone?: string
}

export interface Token {
  access_token: string
  token_type: string
  expires_at: string
}

export interface Workspace {
  id: string
  name: string
  slug: string
  logo_url?: string
  organization_size?: string
  timezone: string
  owner_id: string
  created_by_id?: string
  created_at: string
  updated_at: string
}

export interface WorkspaceLite {
  id: string
  name: string
  slug: string
}

export interface WorkspaceCreate {
  name: string
  slug: string
  organization_size?: string
  timezone?: string
}

export interface Project {
  id: string
  name: string
  identifier: string
  description?: string
  is_public: boolean
  timezone: string
  archived_at?: string
  workspace_id: string
  default_assignee_id?: string
  created_by_id: string
  created_at: string
  updated_at: string
}

export interface ProjectCreate {
  name: string
  identifier: string
  description?: string
  is_public?: boolean
  timezone?: string
  default_assignee_id?: string
}

// Re-export custom field types
export * from './custom-field'

// Re-export project settings types
export * from './project-settings'

// Re-export workflow types
export * from './workflow'