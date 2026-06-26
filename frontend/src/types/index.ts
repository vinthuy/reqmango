export interface User {
  id: number
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
  created_by?: number
  updated_by?: number
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
  id: number
  name: string
  slug: string
  logo_url?: string
  organization_size?: string
  timezone: string
  owner_id: number
  created_by_id?: number
  created_at: string
  updated_at: string
}

export interface WorkspaceLite {
  id: number
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
  id: number
  name: string
  identifier: string
  description?: string
  is_public: boolean
  timezone: string
  archived_at?: string
  workspace_id: number
  default_assignee_id?: number
  created_by_id: number
  created_at: string
  updated_at: string
}

export interface ProjectCreate {
  name: string
  identifier: string
  description?: string
  is_public?: boolean
  timezone?: string
  default_assignee_id?: number
}

// Work Item Type
export interface WorkItemType {
  id: number
  name: string
  description?: string
  icon: string
  color: string
  is_active: boolean
}

// Custom Field
export interface CustomField {
  id: number
  name: string
  type: 'text' | 'number' | 'dropdown' | 'boolean' | 'date' | 'member' | 'url'
  description?: string
  options?: string[]
  is_required: boolean
}

// State
export interface State {
  id: number
  name: string
  category: 'unstarted' | 'started' | 'completed'
  color: string
  is_default: boolean
  allowNew?: boolean
}

// Workflow Transition
export interface WorkflowTransition {
  id: number
  fromStateId: number
  toStateId: number
  type: 'transition' | 'approval'
  allowedRoles: string[]
  approvers?: string[]
}

// Workflow
export interface Workflow {
  id: number
  name: string
  description?: string
  is_active: boolean
}

// Label
export interface Label {
  id: number
  name: string
  color: string
}

// Automation Condition
export interface AutomationCondition {
  field: string
  operator: string
  value: string
}

// Automation Action
export interface AutomationAction {
  type: string
  field: string
  value: string
}

// Automation
export interface Automation {
  id: number
  name: string
  description?: string
  trigger: string
  conditions: AutomationCondition[]
  actions: AutomationAction[]
  is_enabled: boolean
}

export interface WorkspaceMember {
  id: number
  workspace_id: number
  user_id: number
  role: number
  is_active: boolean
  user?: {
    id: number
    display_name: string
    email: string
  }
  created_at: string
  updated_at: string
}

// Re-export custom field types
export * from './custom-field'

// Re-export project settings types
export * from './project-settings'
export * from './project'

// Re-export workflow types
export * from './workflow'