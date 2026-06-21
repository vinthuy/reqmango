/**
 * Module Types - 模块类型定义
 */

// ==================== Module Base ====================

export interface ModuleBase {
  name: string
  description?: string
}

// ==================== Module Create ====================

export interface ModuleCreate extends ModuleBase {
  project_id: number
  workspace_id: number
  parent_id?: number
}

// ==================== Module Update ====================

export interface ModuleUpdate {
  name?: string
  description?: string
  parent_id?: number
}

// ==================== Module Response ====================

export interface ModuleResponse {
  id: number
  name: string
  description: string
  project_id: number
  workspace_id: number
  parent_id: number | null
  order: number
  is_archived: boolean
  archived_at: string | null
  created_at: string
  updated_at: string
}

// ==================== Module Lite ====================

export interface ModuleLite {
  id: number
  name: string
}

// ==================== Module Progress ====================

export interface ModuleProgress {
  module_id: number
  module_name: string
  total_issues: number
  completed_issues: number
  progress: number
  state_breakdown: StateBreakdown[]
}

export interface StateBreakdown {
  state: string
  group: string
  count: number
}

// ==================== Module Statistics ====================

export interface ModuleStatistics extends ModuleProgress {
  priority_breakdown: Record<string, number>
  issue_stats: {
    total: number
    with_start_date: number
    with_target_date: number
  }
  target_date: string | null
  has_sub_modules: boolean
}

// ==================== Module Tree Node ====================

export interface ModuleTreeNode {
  id: number
  name: string
  description?: string
  order: number
  children: ModuleTreeNode[]
}
