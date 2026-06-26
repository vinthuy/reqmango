/**
 * Project Types - 项目类型定义
 */

// ==================== Project Base ====================

export interface ProjectBase {
  name: string
  identifier: string
  description?: string
  is_public?: boolean
  timezone?: string
}

// ==================== Project Create ====================

export interface ProjectCreate extends ProjectBase {
  default_assignee_id?: number
  template_id?: number
}

// ==================== Project Update ====================

export interface ProjectUpdate {
  name?: string
  identifier?: string
  description?: string
  is_public?: boolean
  archived_at?: string
  default_assignee_id?: number
  color?: string
}

// ==================== Project Response ====================

export interface ProjectResponse extends ProjectBase {
  id: number
  archived_at?: string
  workspace_id: number
  default_assignee_id?: number
  project_lead_id?: number
  project_lead?: UserLite
  created_at: string
  updated_at: string
  color?: string
  is_archived?: boolean
}

// ==================== Project Member ====================

export interface ProjectMember {
  id: number
  user_id: number
  role: number
  is_active: boolean
  user?: UserLite
  display_name?: string
  username?: string
  email?: string
}

export interface ProjectMemberCreate {
  user_id?: number
  role?: number
  email?: string
}

export interface ProjectMemberUpdate {
  role: number
}

// ==================== Project Subscriber ====================

export interface ProjectSubscriber {
  id: number
  project_id: number
  user_id: number
  user?: UserLite
  created_at: string
  updated_at: string
}

// ==================== Project Statistics ====================

export interface StateBreakdown {
  state: string
  group: string
  count: number
}

export interface ProjectStatistics {
  project_id: number
  project_name: string
  total_issues: number
  completed_issues: number
  progress: number
  state_breakdown: StateBreakdown[]
  active_members: number
  is_archived: boolean
  in_progress_issues?: number
  member_count?: number
}

// ==================== Project Issues Summary ====================

export interface ProjectIssuesSummary {
  project_id: number
  project_name: string
  issues: {
    todo: number
    in_progress: number
    done: number
    cancelled: number
  }
}

// ==================== User Lite ====================

export interface UserLite {
  id: number
  display_name: string
  username: string
  avatar?: string
}

// ==================== Project Role Constants ====================

export enum ProjectRole {
  VIEWER = 5,
  MEMBER = 10,
  ADMIN = 15,
  OWNER = 20
}

export const ProjectRoleLabels: Record<number, string> = {
  [ProjectRole.VIEWER]: '观察者',
  [ProjectRole.MEMBER]: '成员',
  [ProjectRole.ADMIN]: '管理员',
  [ProjectRole.OWNER]: '所有者'
}

// ==================== Helper Functions ====================

/**
 * 获取项目角色显示名称
 */
export function getProjectRoleName(role: number): string {
  return ProjectRoleLabels[role] || '未知角色'
}

/**
 * 判断是否为项目所有者
 */
export function isProjectOwner(role: number): boolean {
  return role === ProjectRole.OWNER
}

/**
 * 判断是否为项目管理员
 */
export function isProjectAdmin(role: number): boolean {
  return role >= ProjectRole.ADMIN
}

/**
 * 计算项目进度百分比显示
 */
export function formatProjectProgress(progress: number): string {
  return `${Math.round(progress)}%`
}

/**
 * 判断项目是否已归档
 */
export function isProjectArchived(project: ProjectResponse): boolean {
  return project.archived_at !== undefined && project.archived_at !== null
}

/**
 * 判断项目是否公开
 */
export function isProjectPublic(project: ProjectResponse): boolean {
  return project.is_public === true
}

/**
 * 创建空项目对象
 */
export function createEmptyProject(): ProjectCreate {
  return {
    name: '',
    identifier: '',
    description: '',
    is_public: false,
    timezone: 'UTC'
  }
}