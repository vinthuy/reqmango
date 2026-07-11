/**
 * Cycle Types - 周期类型定义
 */

// ==================== Type ====================

export type CycleStatus = 'upcoming' | 'active' | 'completed' | 'cancelled'

// ==================== Cycle Base ====================

export interface CycleBase {
  name: string
  description?: string
  start_date?: string
  end_date?: string
  timezone?: string
}

// ==================== Cycle Create ====================

export interface CycleCreate extends CycleBase {
  project_id: number
}

// ==================== Cycle Update ====================

export interface CycleUpdate {
  name?: string
  description?: string
  start_date?: string
  end_date?: string
  archived_at?: string
  auto_add_enabled?: boolean
  auto_add_rql?: string
  auto_close_enabled?: boolean
  auto_progress_enabled?: boolean
}

// ==================== Project Lite ====================

export interface ProjectLite {
  id: number
  name: string
  identifier: string
}

// ==================== Cycle Response ====================

export interface CycleResponse extends CycleBase {
  id: number
  status: CycleStatus
  progress: number
  total_issues: number
  completed_issues: number
  project_id: number
  workspace_id: number
  owned_by?: UserLite
  project?: ProjectLite
  created_by_id?: number
  updated_by_id?: number
  total_issues_count?: number
  completed_issues_count?: number
  progress_snapshot?: Record<string, any>
  version?: number
  created_at: string
  updated_at: string
  auto_add_enabled?: boolean
  auto_add_rql?: string
  auto_close_enabled?: boolean
  auto_progress_enabled?: boolean
}

// ==================== Cycle Lite ====================

export interface CycleLite {
  id: number
  name: string
  start_date?: string
  end_date?: string
}

// ==================== Cycle Progress ====================

export interface CycleProgress {
  cycle_id: number
  cycle_name: string
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

// ==================== Cycle Statistics ====================

export interface CycleStatistics extends CycleProgress {
  priority_breakdown: Record<string, number>
  issue_stats: {
    total: number
    with_start_date: number
    with_target_date: number
  }
  date_range: {
    start_date: string | null
    end_date: string | null
  }
}

// ==================== Burndown Chart ====================

export interface BurndownDayPoint {
  day_index: number
  date: string
  ideal_remaining: number
  actual_completed: number
  actual_remaining: number
}

export interface BurndownData {
  cycle_id: number
  cycle_name: string
  start_date: string
  end_date: string
  total_issues: number
  total_days: number
  days_elapsed: number
  ideal_daily_burn: number
  ideal_remaining: number
  actual_completed: number
  actual_remaining: number
  is_on_track: boolean
  daily_points?: BurndownDayPoint[]
}

// ==================== User Lite ====================

export interface UserLite {
  id: number
  display_name: string
  username?: string
  email: string
  avatar_url?: string
}

// ==================== Helper Functions ====================

/**
 * 获取周期状态的显示名称
 */
export function getCycleStatusName(status: CycleStatus): string {
  const names: Record<CycleStatus, string> = {
    upcoming: '未开始',
    active: '进行中',
    completed: '已完成',
    cancelled: '已取消'
  }
  return names[status] || status
}

/**
 * 获取周期状态的颜色
 */
export function getCycleStatusColor(status: CycleStatus): string {
  const colors: Record<CycleStatus, string> = {
    upcoming: '#6B7280', // 灰色
    active: '#3B82F6',   // 蓝色
    completed: '#10B981', // 绿色
    cancelled: '#EF4444'  // 红色
  }
  return colors[status] || '#6B7280'
}

/**
 * 获取周期状态图标
 */
export function getCycleStatusIcon(status: CycleStatus): string {
  const icons: Record<CycleStatus, string> = {
    upcoming: 'clock',
    active: 'play',
    completed: 'check',
    cancelled: 'x'
  }
  return icons[status] || 'circle'
}

/**
 * 判断周期是否活跃
 */
export function isCycleActive(cycle: CycleResponse): boolean {
  return cycle.status === 'active'
}

/**
 * 判断周期是否已完成
 */
export function isCycleCompleted(cycle: CycleResponse): boolean {
  return cycle.status === 'completed'
}

/**
 * 计算周期剩余天数
 */
export function getDaysRemaining(cycle: CycleResponse): number {
  if (!cycle.end_date) return -1
  const endDate = new Date(cycle.end_date)
  const today = new Date()
  const diffTime = endDate.getTime() - today.getTime()
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  return diffDays
}

/**
 * 计算周期进度百分比显示
 */
export function formatProgress(progress: number): string {
  return `${Math.round(progress)}%`
}

/**
 * 创建空周期对象
 */
export function createEmptyCycle(projectId: number): CycleCreate {
  return {
    name: '',
    description: '',
    start_date: undefined,
    end_date: undefined,
    project_id: projectId
  }
}

/**
 * 判断周期是否过期
 */
export function isCycleOverdue(cycle: CycleResponse): boolean {
  if (!cycle.end_date) return false
  if (cycle.status === 'completed') return false
  const endDate = new Date(cycle.end_date)
  const today = new Date()
  return endDate < today
}

/**
 * 获取燃尽图状态
 */
export function getBurndownStatus(burndown: BurndownData): 'ahead' | 'behind' | 'on_track' {
  if (burndown.is_on_track) return 'on_track'
  if (burndown.actual_remaining > burndown.ideal_remaining) return 'behind'
  return 'ahead'
}

/**
 * 格式化周期日期范围
 */
export function formatDateRange(startDate?: string, endDate?: string): string {
  if (!startDate && !endDate) return '无日期'
  if (startDate && !endDate) return `${formatDate(startDate)} 起`
  if (!startDate && endDate) return `至 ${formatDate(endDate)}`
  return `${formatDate(startDate!)} - ${formatDate(endDate!)}`
}

/**
 * 格式化日期
 */
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}
