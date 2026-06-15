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
  default_assignee_id?: string
}

// ==================== Project Update ====================

export interface ProjectUpdate {
  name?: string
  description?: string
  is_public?: boolean
  archived_at?: string
  default_assignee_id?: string
}

// ==================== Project Response ====================

export interface ProjectResponse extends ProjectBase {
  id: string
  archived_at?: string
  workspace_id: string
  default_assignee_id?: string
  created_at: string
  updated_at: string
}

// ==================== Project Member ====================

export interface ProjectMember {
  id: string
  user_id: string
  role: number
  is_active: boolean
  user?: UserLite
}

export interface ProjectMemberCreate {
  user_id: string
  role?: number
}

export interface ProjectMemberUpdate {
  role: number
}

// ==================== Project Statistics ====================

export interface StateBreakdown {
  state: string
  group: string
  count: number
}

export interface ProjectStatistics {
  project_id: string
  project_name: string
  total_issues: number
  completed_issues: number
  progress: number
  state_breakdown: StateBreakdown[]
  active_members: number
  is_archived: boolean
}

// ==================== Project Issues Summary ====================

export interface ProjectIssuesSummary {
  project_id: string
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
  id: string
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
export function createEmptyProject(workspaceId: string): ProjectCreate {
  return {
    name: '',
    identifier: '',
    description: '',
    is_public: false,
    timezone: 'UTC'
  }
}