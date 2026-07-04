/**
 * Cycle 类型和工具函数单元测试
 */
import { describe, it, expect } from 'vitest'
import {
  getCycleStatusName,
  getCycleStatusColor,
  getCycleStatusIcon,
  isCycleActive,
  isCycleCompleted,
  getDaysRemaining,
  formatProgress,
  createEmptyCycle,
  isCycleOverdue,
  getBurndownStatus,
  formatDateRange,
} from './cycle'
import type {
  CycleCreate,
  CycleUpdate,
  CycleResponse,
  CycleProgress,
  CycleStatistics,
  BurndownData,
  StateBreakdown,
} from './cycle'

// ==================== CycleStatus Helpers ====================
describe('getCycleStatusName', () => {
  it('should return Chinese name for upcoming', () => {
    expect(getCycleStatusName('upcoming')).toBe('未开始')
  })
  it('should return Chinese name for active', () => {
    expect(getCycleStatusName('active')).toBe('进行中')
  })
  it('should return Chinese name for completed', () => {
    expect(getCycleStatusName('completed')).toBe('已完成')
  })
  it('should return Chinese name for cancelled', () => {
    expect(getCycleStatusName('cancelled')).toBe('已取消')
  })
})

describe('getCycleStatusColor', () => {
  it('should return gray for upcoming', () => {
    expect(getCycleStatusColor('upcoming')).toBe('#6B7280')
  })
  it('should return blue for active', () => {
    expect(getCycleStatusColor('active')).toBe('#3B82F6')
  })
  it('should return green for completed', () => {
    expect(getCycleStatusColor('completed')).toBe('#10B981')
  })
  it('should return red for cancelled', () => {
    expect(getCycleStatusColor('cancelled')).toBe('#EF4444')
  })
})

describe('getCycleStatusIcon', () => {
  it('should return clock for upcoming', () => {
    expect(getCycleStatusIcon('upcoming')).toBe('clock')
  })
  it('should return play for active', () => {
    expect(getCycleStatusIcon('active')).toBe('play')
  })
  it('should return check for completed', () => {
    expect(getCycleStatusIcon('completed')).toBe('check')
  })
  it('should return x for cancelled', () => {
    expect(getCycleStatusIcon('cancelled')).toBe('x')
  })
})

// ==================== Cycle Status Checkers ====================
describe('isCycleActive', () => {
  it('should return true for active cycle', () => {
    expect(isCycleActive({ status: 'active' } as CycleResponse)).toBe(true)
  })
  it('should return false for upcoming cycle', () => {
    expect(isCycleActive({ status: 'upcoming' } as CycleResponse)).toBe(false)
  })
  it('should return false for completed cycle', () => {
    expect(isCycleActive({ status: 'completed' } as CycleResponse)).toBe(false)
  })
})

describe('isCycleCompleted', () => {
  it('should return true for completed cycle', () => {
    expect(isCycleCompleted({ status: 'completed' } as CycleResponse)).toBe(true)
  })
  it('should return false for active cycle', () => {
    expect(isCycleCompleted({ status: 'active' } as CycleResponse)).toBe(false)
  })
})

// ==================== getDaysRemaining ====================
describe('getDaysRemaining', () => {
  it('should return -1 if no end_date', () => {
    expect(getDaysRemaining({ end_date: undefined } as CycleResponse)).toBe(-1)
  })
  it('should return positive days for future date', () => {
    const future = new Date()
    future.setDate(future.getDate() + 5)
    const days = getDaysRemaining({ end_date: future.toISOString() } as CycleResponse)
    expect(days).toBeGreaterThanOrEqual(4)
  })
  it('should return negative or zero for past date', () => {
    const past = new Date()
    past.setDate(past.getDate() - 3)
    const days = getDaysRemaining({ end_date: past.toISOString() } as CycleResponse)
    expect(days).toBeLessThanOrEqual(0)
  })
})

// ==================== formatProgress ====================
describe('formatProgress', () => {
  it('should format integer progress', () => {
    expect(formatProgress(75)).toBe('75%')
  })
  it('should round decimal progress', () => {
    expect(formatProgress(33.333)).toBe('33%')
  })
  it('should format zero', () => {
    expect(formatProgress(0)).toBe('0%')
  })
  it('should format 100', () => {
    expect(formatProgress(100)).toBe('100%')
  })
})

// ==================== createEmptyCycle ====================
describe('createEmptyCycle', () => {
  it('should create with correct project_id', () => {
    const cycle = createEmptyCycle(42)
    expect(cycle.project_id).toBe(42)
    expect(cycle.name).toBe('')
    expect(cycle.description).toBe('')
    expect(cycle.start_date).toBeUndefined()
    expect(cycle.end_date).toBeUndefined()
  })
})

// ==================== isCycleOverdue ====================
describe('isCycleOverdue', () => {
  it('should return false if no end_date', () => {
    expect(isCycleOverdue({ end_date: undefined } as CycleResponse)).toBe(false)
  })
  it('should return false if already completed', () => {
    const past = new Date(); past.setDate(past.getDate() - 10)
    expect(isCycleOverdue({ end_date: past.toISOString(), status: 'completed' } as CycleResponse)).toBe(false)
  })
  it('should return true if past end_date and not completed', () => {
    const past = new Date(); past.setDate(past.getDate() - 2)
    expect(isCycleOverdue({ end_date: past.toISOString(), status: 'active' } as CycleResponse)).toBe(true)
  })
  it('should return false if future end_date', () => {
    const future = new Date(); future.setDate(future.getDate() + 10)
    expect(isCycleOverdue({ end_date: future.toISOString(), status: 'active' } as CycleResponse)).toBe(false)
  })
})

// ==================== getBurndownStatus ====================
describe('getBurndownStatus', () => {
  it('should return on_track when is_on_track is true', () => {
    expect(getBurndownStatus({ is_on_track: true } as BurndownData)).toBe('on_track')
  })
  it('should return behind when actual > ideal remaining', () => {
    expect(getBurndownStatus({
      is_on_track: false,
      actual_remaining: 20,
      ideal_remaining: 15,
    } as BurndownData)).toBe('behind')
  })
  it('should return ahead when actual < ideal remaining', () => {
    expect(getBurndownStatus({
      is_on_track: false,
      actual_remaining: 5,
      ideal_remaining: 15,
    } as BurndownData)).toBe('ahead')
  })
})

// ==================== formatDateRange ====================
describe('formatDateRange', () => {
  it('should return No Date when both empty', () => {
    expect(formatDateRange(undefined, undefined)).toBe('无日期')
  })
  it('should handle start-only range', () => {
    const result = formatDateRange('2026-01-15', undefined)
    expect(result).toContain('起')
  })
  it('should handle end-only range', () => {
    const result = formatDateRange(undefined, '2026-03-20')
    expect(result).toContain('至')
  })
  it('should format full range', () => {
    const result = formatDateRange('2026-01-15', '2026-03-20')
    expect(result).toContain('-')
  })
})

// ==================== Type Validation ====================
describe('CycleCreate', () => {
  it('should require name and project_id', () => {
    const c: CycleCreate = { name: 'Sprint 1', project_id: 1 }
    expect(c.name).toBe('Sprint 1')
    expect(c.project_id).toBe(1)
  })
})

describe('CycleUpdate', () => {
  it('should allow all fields optional', () => {
    const u: CycleUpdate = {}
    expect(u).toBeDefined()
  })
})

describe('CycleResponse', () => {
  it('should require all key fields', () => {
    const r: CycleResponse = {
      id: 1, name: 'Sprint 1', status: 'active',
      progress: 50, total_issues: 10, completed_issues: 5,
      project_id: 1, workspace_id: 1,
      created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
    }
    expect(r.id).toBe(1)
    expect(r.progress).toBe(50)
    expect(r.status).toBe('active')
  })
})

describe('CycleProgress', () => {
  it('should accept state_breakdown', () => {
    const breakdown: StateBreakdown = { state: 'Todo', group: 'unstarted', count: 5 }
    const p: CycleProgress = {
      cycle_id: 1, cycle_name: 'Sprint 1',
      total_issues: 10, completed_issues: 5, progress: 50,
      state_breakdown: [breakdown],
    }
    expect(p.state_breakdown).toHaveLength(1)
    expect(p.state_breakdown[0].state).toBe('Todo')
  })
})

describe('CycleStatistics', () => {
  it('should extend CycleProgress with extra fields', () => {
    const s: CycleStatistics = {
      cycle_id: 1, cycle_name: 'Sprint 1',
      total_issues: 10, completed_issues: 5, progress: 50,
      state_breakdown: [],
      priority_breakdown: { high: 3, low: 7 },
      issue_stats: { total: 10, with_start_date: 8, with_target_date: 9 },
      date_range: { start_date: '2026-01-01', end_date: '2026-01-14' },
    }
    expect(s.priority_breakdown.high).toBe(3)
  })
})

describe('BurndownData', () => {
  it('should have all burndown fields', () => {
    const b: BurndownData = {
      cycle_id: 1, cycle_name: 'Sprint 1',
      start_date: '2026-01-01', end_date: '2026-01-14',
      total_issues: 20, total_days: 14, days_elapsed: 7,
      ideal_daily_burn: 1.43, ideal_remaining: 10, actual_completed: 8,
      actual_remaining: 12, is_on_track: false,
    }
    expect(b.total_days).toBe(14)
    expect(b.ideal_daily_burn).toBeCloseTo(1.43, 1)
  })
})
