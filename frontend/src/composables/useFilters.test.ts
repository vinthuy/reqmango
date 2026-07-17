/**
 * useFilters Composable 单元测试
 * 覆盖：CRUD 筛选条件、排序、分组、快速搜索、搜索历史、RQL 双向同步
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useFilters, type FiltersContext } from './useFilters'

// Mock localStorage
const storage = new Map<string, string>()
vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => storage.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => { storage.set(key, value) }),
  removeItem: vi.fn((key: string) => { storage.delete(key) }),
  clear: vi.fn(() => { storage.clear() }),
})

describe('useFilters', () => {
  let ctx: FiltersContext

  beforeEach(() => {
    vi.clearAllMocks()
    storage.clear()
    const pinia = createPinia()
    setActivePinia(pinia)
    ctx = useFilters()
    ctx.clearAll()
  })

  describe('initial state', () => {
    it('should have empty filters array', () => {
      expect(ctx.state.filters).toEqual([])
    })
    it('should have empty sortBy', () => {
      expect(ctx.state.sortBy).toEqual([])
    })
    it('should have null groupBy', () => {
      expect(ctx.state.groupBy).toBeNull()
    })
    it('should have null subGroupBy', () => {
      expect(ctx.state.subGroupBy).toBeNull()
    })
    it('should have empty quickSearch', () => {
      expect(ctx.state.quickSearch).toBe('')
    })
    it('should have empty searchHistory', () => {
      expect(ctx.state.searchHistory).toEqual([])
    })
    it('should have empty rql computed', () => {
      expect(ctx.rql.value).toBe('')
    })
  })

  describe('computed properties', () => {
    it('rql should reflect added filters', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      expect(ctx.rql.value).toContain('priority')
      expect(ctx.rql.value).toContain('high')
    })

    it('activeFilterCount should count filters + quickSearch', () => {
      expect(ctx.activeFilterCount.value).toBe(0)
      ctx.addFilter('state_id', 'is', 1, 'Backlog')
      expect(ctx.activeFilterCount.value).toBe(1)
      ctx.setQuickSearch('bug')
      expect(ctx.activeFilterCount.value).toBe(2)
    })

    it('isEmpty should be true when no filters or search', () => {
      expect(ctx.isEmpty.value).toBe(true)
      ctx.addFilter('priority', 'is', 'urgent', 'Urgent')
      expect(ctx.isEmpty.value).toBe(false)
    })
  })

  describe('addFilter', () => {
    it('should add a filter condition', () => {
      ctx.addFilter('state_id', 'is', 2, 'In Progress')
      expect(ctx.state.filters.length).toBe(1)
      expect(ctx.state.filters[0]).toMatchObject({
        field: 'state_id',
        operator: 'is',
        value: 2,
        displayValue: 'In Progress',
      })
    })

    it('should handle multiple filters', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      ctx.addFilter('assignee_id', 'is', 5, 'Alice')
      expect(ctx.state.filters.length).toBe(2)
      expect(ctx.state.filters[0].field).toBe('priority')
      expect(ctx.state.filters[1].field).toBe('assignee_id')
    })
  })

  describe('removeFilter', () => {
    it('should remove filter by index', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      ctx.addFilter('state_id', 'is', 1, 'Backlog')
      ctx.removeFilter(0)
      expect(ctx.state.filters.length).toBe(1)
      expect(ctx.state.filters[0].field).toBe('state_id')
    })
  })

  describe('updateFilter', () => {
    it('should update an existing filter', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      ctx.updateFilter(0, { value: 'urgent', displayValue: 'Urgent' })
      expect(ctx.state.filters[0].value).toBe('urgent')
      expect(ctx.state.filters[0].displayValue).toBe('Urgent')
    })

    it('should update field and operator', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      ctx.updateFilter(0, { field: 'type_id', operator: 'is not' })
      expect(ctx.state.filters[0].field).toBe('type_id')
      expect(ctx.state.filters[0].operator).toBe('is not')
    })
  })

  describe('clearAll', () => {
    it('should reset all state', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      ctx.setSortBy([{ key: 'created_at', labelKey: 'Created', direction: 'desc' }])
      ctx.setGroupBy({ key: 'state_id', labelKey: 'State' })
      ctx.setSubGroupBy({ key: 'priority', labelKey: 'Priority' })
      ctx.setQuickSearch('bug')
      ctx.clearAll()
      expect(ctx.state.filters).toEqual([])
      expect(ctx.state.sortBy).toEqual([])
      expect(ctx.state.groupBy).toBeNull()
      expect(ctx.state.subGroupBy).toBeNull()
      expect(ctx.state.quickSearch).toBe('')
    })
  })

  describe('sortBy', () => {
    it('setSortBy should replace sorts', () => {
      ctx.setSortBy([{ key: 'created_at', labelKey: 'Created', direction: 'desc' }, { key: 'priority', labelKey: 'Priority', direction: 'asc' }])
      expect(ctx.state.sortBy.length).toBe(2)
    })

    it('addSortBy should add new sort', () => {
      ctx.addSortBy({ key: 'created_at', labelKey: 'Created', direction: 'desc' })
      expect(ctx.state.sortBy.length).toBe(1)
      ctx.addSortBy({ key: 'priority', labelKey: 'Priority', direction: 'asc' })
      expect(ctx.state.sortBy.length).toBe(2)
    })

    it('addSortBy should replace duplicate key', () => {
      ctx.addSortBy({ key: 'created_at', labelKey: 'Created', direction: 'desc' })
      ctx.addSortBy({ key: 'created_at', labelKey: 'Created', direction: 'asc' })
      expect(ctx.state.sortBy.length).toBe(1)
      expect(ctx.state.sortBy[0].direction).toBe('asc')
    })

    it('removeSortBy should remove by index', () => {
      ctx.setSortBy([{ key: 'created_at', labelKey: 'Created', direction: 'desc' }, { key: 'priority', labelKey: 'Priority', direction: 'asc' }])
      ctx.removeSortBy(0)
      expect(ctx.state.sortBy.length).toBe(1)
      expect(ctx.state.sortBy[0].key).toBe('priority')
    })
  })

  describe('groupBy', () => {
    it('should set groupBy', () => {
      ctx.setGroupBy({ key: 'state_id', labelKey: 'State' })
      expect(ctx.state.groupBy?.key).toBe('state_id')
    })

    it('should clear groupBy with null', () => {
      ctx.setGroupBy({ key: 'priority', labelKey: 'Priority' })
      ctx.setGroupBy(null)
      expect(ctx.state.groupBy).toBeNull()
    })
  })

  describe('subGroupBy', () => {
    it('should set subGroupBy', () => {
      ctx.setSubGroupBy({ key: 'priority', labelKey: 'Priority' })
      expect(ctx.state.subGroupBy?.key).toBe('priority')
    })
  })

  describe('quickSearch', () => {
    it('should set quickSearch', () => {
      ctx.setQuickSearch('login bug')
      expect(ctx.state.quickSearch).toBe('login bug')
    })

    it('should clear quickSearch', () => {
      ctx.setQuickSearch('test')
      ctx.setQuickSearch('')
      expect(ctx.state.quickSearch).toBe('')
    })
  })

  describe('searchHistory', () => {
    it('addToHistory should add entry', () => {
      ctx.addToHistory('API crash fix')
      expect(ctx.state.searchHistory).toContain('API crash fix')
    })

    it('addToHistory should deduplicate', () => {
      ctx.addToHistory('bug')
      ctx.addToHistory('feature')
      ctx.addToHistory('bug')
      expect(ctx.state.searchHistory).toEqual(['bug', 'feature'])
    })

    it('addToHistory should enforce max 10 items', () => {
      for (let i = 0; i < 15; i++) {
        ctx.addToHistory(`query-${i}`)
      }
      expect(ctx.state.searchHistory.length).toBe(10)
      expect(ctx.state.searchHistory[0]).toBe('query-14')
    })

    it('addToHistory should persist to localStorage', () => {
      ctx.addToHistory('persist test')
      expect(localStorage.setItem).toHaveBeenCalled()
      const key = (localStorage.setItem as any).mock.calls[0][0]
      expect(key).toContain('reqmango')
    })

    it('removeFromHistory should remove entry', () => {
      ctx.addToHistory('a')
      ctx.addToHistory('b')
      ctx.removeFromHistory(0)
      expect(ctx.state.searchHistory).toEqual(['a'])
    })

    it('clearHistory should clear all', () => {
      ctx.addToHistory('a')
      ctx.addToHistory('b')
      ctx.clearHistory()
      expect(ctx.state.searchHistory).toEqual([])
    })
  })

  describe('restoreFromRQL', () => {
    it('should clear filters for empty RQL', () => {
      ctx.addFilter('priority', 'is', 'high', 'High')
      ctx.restoreFromRQL('')
      expect(ctx.state.filters).toEqual([])
    })

    it('should not crash on invalid RQL', () => {
      ctx.restoreFromRQL('garbage_input')
      // Should not throw
      expect(ctx.state.filters).toBeDefined()
    })
  })
})
