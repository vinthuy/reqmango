/**
 * Cycle Store 单元测试
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import type { CycleStatus } from '@/types/cycle'

// Mock cycleApi before store import
vi.mock('@/api/cycle', () => ({
  cycleApi: {
    listCycles: vi.fn(),
    getCycle: vi.fn(),
    createCycle: vi.fn(),
    updateCycle: vi.fn(),
    deleteCycle: vi.fn(),
    startCycle: vi.fn(),
    endCycle: vi.fn(),
    cancelCycle: vi.fn(),
    addIssueToCycle: vi.fn(),
    removeIssueFromCycle: vi.fn(),
    getCycleIssues: vi.fn(),
    getCycleProgress: vi.fn(),
    getCycleStatistics: vi.fn(),
    getBurndownData: vi.fn(),
  },
}))

import { useCycleStore } from './cycle'
import { cycleApi } from '@/api/cycle'

const mockApi = cycleApi as any

const activeStatus: CycleStatus = 'active'
function makeCycle(id: number, overrides = {}) {
  return {
    id, name: `Sprint ${id}`, status: activeStatus as CycleStatus, progress: 0,
    total_issues: 0, completed_issues: 0,
    project_id: 1, workspace_id: 1,
    created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

describe('useCycleStore', () => {
  let store: ReturnType<typeof useCycleStore>

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    store = useCycleStore()
  })

  // ==================== Initial State ====================
  describe('Initial State', () => {
    it('should have empty cycles array', () => {
      expect(store.cycles).toEqual([])
    })
    it('should have null currentCycle', () => {
      expect(store.currentCycle).toBeNull()
    })
    it('should have null progress/statistics/burndown', () => {
      expect(store.progress).toBeNull()
      expect(store.statistics).toBeNull()
      expect(store.burndown).toBeNull()
    })
    it('should not be loading', () => {
      expect(store.isLoading).toBe(false)
    })
    it('should have null error', () => {
      expect(store.error).toBeNull()
    })
  })

  // ==================== fetchCycles ====================
  describe('fetchCycles', () => {
    it('should fetch and set cycles', async () => {
      const items = [makeCycle(1), makeCycle(2)]
      mockApi.listCycles.mockResolvedValue({ items, total: 2, limit: 50, offset: 0 })
      await store.fetchCycles(1)
      expect(store.cycles).toHaveLength(2)
      expect(mockApi.listCycles).toHaveBeenCalledWith(1, { status: undefined })
    })

    it('should fetch with status filter', async () => {
      mockApi.listCycles.mockResolvedValue({ items: [], total: 0, limit: 50, offset: 0 })
      await store.fetchCycles(1, 'active')
      expect(mockApi.listCycles).toHaveBeenCalledWith(1, { status: 'active' })
    })

    it('should set error on failure', async () => {
      mockApi.listCycles.mockRejectedValue(new Error('Network error'))
      await store.fetchCycles(1)
      expect(store.error).toBe('Network error')
    })

    it('should set isLoading during fetch', async () => {
      mockApi.listCycles.mockResolvedValue({ items: [], total: 0, limit: 50, offset: 0 })
      const promise = store.fetchCycles(1)
      expect(store.isLoading).toBe(true)
      await promise
      expect(store.isLoading).toBe(false)
    })
  })

  // ==================== fetchCycle ====================
  describe('fetchCycle', () => {
    it('should fetch single cycle', async () => {
      const c = makeCycle(42)
      mockApi.getCycle.mockResolvedValue(c)
      await store.fetchCycle(42)
      expect(store.currentCycle).toEqual(c)
      expect(mockApi.getCycle).toHaveBeenCalledWith(42)
    })

    it('should set error on failure', async () => {
      mockApi.getCycle.mockRejectedValue(new Error('Not found'))
      await store.fetchCycle(999)
      expect(store.error).toBe('Not found')
    })
  })

  // ==================== createCycleAction ====================
  describe('createCycleAction', () => {
    it('should create and prepend to list', async () => {
      store.cycles = [makeCycle(1)]
      const created = makeCycle(2, { name: 'New' })
      mockApi.createCycle.mockResolvedValue(created)
      const result = await store.createCycleAction(1, 1, { name: 'New', project_id: 1 })
      expect(result).toEqual(created)
      expect(store.cycles).toHaveLength(2)
      expect(store.cycles[0].name).toBe('New')
    })

    it('should return null on failure', async () => {
      mockApi.createCycle.mockRejectedValue(new Error('Invalid'))
      const result = await store.createCycleAction(1, 1, { name: 'Fail', project_id: 1 })
      expect(result).toBeNull()
    })
  })

  // ==================== updateCycleAction ====================
  describe('updateCycleAction', () => {
    it('should update cycle in list and current', async () => {
      const original = makeCycle(1, { name: 'Old' })
      store.cycles = [original]
      store.currentCycle = original
      const updated = { ...original, name: 'Updated' }
      mockApi.updateCycle.mockResolvedValue(updated)
      const result = await store.updateCycleAction(1, { name: 'Updated' })
      expect(store.cycles[0].name).toBe('Updated')
      expect(store.currentCycle?.name).toBe('Updated')
      expect(result).toEqual(updated)
    })
  })

  // ==================== deleteCycleAction ====================
  describe('deleteCycleAction', () => {
    it('should remove from list and clear current if matches', async () => {
      const c1 = makeCycle(1), c2 = makeCycle(2)
      store.cycles = [c1, c2]
      store.currentCycle = c1
      mockApi.deleteCycle.mockResolvedValue(undefined)
      await store.deleteCycleAction(1)
      expect(store.cycles).toHaveLength(1)
      expect(store.currentCycle).toBeNull()
    })
  })

  // ==================== Status Transitions ====================
  describe('startCycle', () => {
    it('should start a cycle and update in list', async () => {
      const c = makeCycle(1, { status: 'upcoming' })
      store.cycles = [c]
      const started = { ...c, status: 'active' }
      mockApi.startCycle.mockResolvedValue(started)
      const result = await store.startCycle(1)
      expect(store.cycles[0].status).toBe('active')
      expect(result).toEqual(started)
    })
  })

  describe('endCycle', () => {
    it('should end a cycle', async () => {
      const c = makeCycle(1)
      store.cycles = [c]
      const ended = { ...c, status: 'completed' }
      mockApi.endCycle.mockResolvedValue(ended)
      await store.endCycle(1)
      expect(store.cycles[0].status).toBe('completed')
    })
  })

  describe('cancelCycle', () => {
    it('should cancel a cycle', async () => {
      const c = makeCycle(1)
      store.cycles = [c]
      const cancelled = { ...c, status: 'cancelled' }
      mockApi.cancelCycle.mockResolvedValue(cancelled)
      await store.cancelCycle(1)
      expect(store.cycles[0].status).toBe('cancelled')
    })
  })

  // ==================== Issue Association ====================
  describe('addIssueToCycle', () => {
    it('should add issue and refresh issues/progress', async () => {
      mockApi.addIssueToCycle.mockResolvedValue({ cycle_id: 1, issue_id: 42, action: 'add' })
      mockApi.getCycleIssues.mockResolvedValue([])
      mockApi.getCycleProgress.mockResolvedValue({ cycle_id: 1, cycle_name: 'Sprint 1', total_issues: 0, completed_issues: 0, progress: 0, state_breakdown: [] })
      const result = await store.addIssueToCycle(1, 42)
      expect(mockApi.getCycleIssues).toHaveBeenCalledWith(1, undefined)
      expect(mockApi.getCycleProgress).toHaveBeenCalledWith(1)
      expect(result).toEqual({ cycle_id: 1, issue_id: 42, action: 'add' })
    })
  })

  describe('removeIssueFromCycle', () => {
    it('should remove issue and filter from list', async () => {
      store.cycleIssues = [{ id: 1 }, { id: 2 }]
      mockApi.removeIssueFromCycle.mockResolvedValue({ cycle_id: 1, issue_id: 1, action: 'remove' })
      mockApi.getCycleProgress.mockResolvedValue(null)
      await store.removeIssueFromCycle(1, 1)
      expect(store.cycleIssues).toHaveLength(1)
    })
  })

  // ==================== Analysis ====================
  describe('fetchProgress', () => {
    it('should fetch and set progress', async () => {
      const progress = { cycle_id: 1, cycle_name: 'Sprint 1', total_issues: 10, completed_issues: 5, progress: 50, state_breakdown: [] }
      mockApi.getCycleProgress.mockResolvedValue(progress)
      await store.fetchProgress(1)
      expect(store.progress).toEqual(progress)
    })
  })

  describe('fetchStatistics', () => {
    it('should fetch statistics', async () => {
      const stats = { cycle_id: 1, cycle_name: 'Sprint 1', total_issues: 10, completed_issues: 5, progress: 50, state_breakdown: [], priority_breakdown: {}, issue_stats: { total: 10, with_start_date: 8, with_target_date: 9 }, date_range: { start_date: null, end_date: null } }
      mockApi.getCycleStatistics.mockResolvedValue(stats)
      await store.fetchStatistics(1)
      expect(store.statistics).toEqual(stats)
    })
  })

  describe('fetchBurndown', () => {
    it('should fetch burndown data', async () => {
      const bd = { cycle_id: 1, cycle_name: 'Sprint 1', start_date: '', end_date: '', total_issues: 20, total_days: 14, days_elapsed: 7, ideal_daily_burn: 1.43, ideal_remaining: 10, actual_completed: 8, actual_remaining: 12, is_on_track: false }
      mockApi.getBurndownData.mockResolvedValue(bd)
      await store.fetchBurndown(1)
      expect(store.burndown).toEqual(bd)
    })
  })
})
