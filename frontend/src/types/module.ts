/**
 * Module Types - 模块类型定义
 */

// ==================== Module Base ====================

export interface ModuleBase {
  name: string
  description?: string
  target_date?: string
}

// ==================== Module Create ====================

export interface ModuleCreate extends ModuleBase {
  project_id: string
}

// ==================== Module Update ====================

export interface ModuleUpdate {
  name?: string
  description?: string
  target_date?: string
  archived_at?: string
}

// ==================== Module Response ====================

export interface ModuleResponse extends ModuleBase {
  id: string
  sequence: number
  progress: number
  total_issues: number
  completed_issues: number
  project_id: string
  workspace_id: string
  parent_id?: string
  created_at: string
  updated_at: string
}

// ==================== Module Lite ====================

export interface ModuleLite {
  id: string
  name: string
}

// ==================== Module Progress ====================

export interface ModuleProgress {
  module_id: string
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
  id: string
  name: string
  description?: string
  sequence: number
  progress: number
  total_issues: number
  completed_issues: number
  children: ModuleTreeNode[]
}

// ==================== Helper Functions ====================

/**
 * 计算模块进度百分比显示
 */
export function formatModuleProgress(progress: number): string {
  return `${Math.round(progress)}%`
}

/**
 * 创建空模块对象
 */
export function createEmptyModule(projectId: string): ModuleCreate {
  return {
    name: '',
    description: '',
    target_date: undefined,
    project_id: projectId
  }
}

/**
 * 判断模块是否有子模块
 */
export function hasSubModules(module: ModuleTreeNode): boolean {
  return module.children && module.children.length > 0
}

/**
 * 计算模块树的总工作项数
 */
export function calculateTotalIssuesInTree(node: ModuleTreeNode): number {
  let total = node.total_issues
  if (node.children) {
    for (const child of node.children) {
      total += calculateTotalIssuesInTree(child)
    }
  }
  return total
}

/**
 * 计算模块树的总完成工作项数
 */
export function calculateCompletedIssuesInTree(node: ModuleTreeNode): number {
  let completed = node.completed_issues
  if (node.children) {
    for (const child of node.children) {
      completed += calculateCompletedIssuesInTree(child)
    }
  }
  return completed
}

/**
 * 计算模块树的整体进度
 */
export function calculateTreeProgress(node: ModuleTreeNode): number {
  const total = calculateTotalIssuesInTree(node)
  const completed = calculateCompletedIssuesInTree(node)
  return (completed / total * 100) if total > 0 else 0
}

/**
 * 在模块树中查找特定模块
 */
export function findModuleInTree(tree: ModuleTreeNode[], moduleId: string): ModuleTreeNode | null {
  for (const node of tree) {
    if (node.id === moduleId) return node
    if (node.children) {
      const found = findModuleInTree(node.children, moduleId)
      if (found) return found
    }
  }
  return null
}

/**
 * 获取模块树的深度
 */
export function getTreeDepth(node: ModuleTreeNode): number {
  if (!node.children || node.children.length === 0) return 1
  return 1 + Math.max(...node.children.map(child => getTreeDepth(child)))
}

/**
 * 判断模块是否过期
 */
export function isModuleOverdue(module: ModuleResponse): boolean {
  if (!module.target_date) return false
  const targetDate = new Date(module.target_date)
  const today = new Date()
  return targetDate < today && module.progress < 100
}

/**
 * 格式化模块目标日期
 */
export function formatTargetDate(targetDate?: string): string {
  if (!targetDate) return '无目标日期'
  const date = new Date(targetDate)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
}

/**
 * 获取模块状态颜色（基于进度）
 */
export function getModuleStatusColor(progress: number): string {
  if (progress >= 100) return '#10B981' // 绿色 - 完成
  if (progress >= 75) return '#3B82F6'  // 蓝色 - 接近完成
  if (progress >= 50) return '#F59E0B'  // 橙色 - 进行中
  if (progress >= 25) return '#6B7280'  // 灰色 - 刚开始
  return '#EF4444'                       // 红色 - 未开始或落后
}

/**
 * 获取模块状态文本（基于进度）
 */
export function getModuleStatusText(progress: number): string {
  if (progress >= 100) return '已完成'
  if (progress >= 75) return '接近完成'
  if (progress >= 50) return '进行中'
  if (progress >= 25) return '刚开始'
  return '未开始'
}

/**
 * 扁平化模块树（用于列表显示）
 */
export function flattenModuleTree(tree: ModuleTreeNode[], level: number = 0): Array<ModuleTreeNode & { level: number }> {
  const result: Array<ModuleTreeNode & { level: number }> = []
  
  for (const node of tree) {
    result.push({ ...node, level })
    if (node.children && node.children.length > 0) {
      result.push(...flattenModuleTree(node.children, level + 1))
    }
  }
  
  return result
}