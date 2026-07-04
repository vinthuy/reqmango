/**
 * useRQL Composable 单元测试
 * 覆盖：search 执行、localStorage 历史记录 CRUD、状态管理
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useRQL, HISTORY_KEY } from './useRQL'

// Mock rqlApi
vi.mock('@/api/rql', () => ({
  rqlApi: {
    search: vi.fn(),
  },
}))

import { rqlApi } from '@/api/rql'

// Mock localStorage
const storage = new Map<string, string>()
vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => storage.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => { storage.set(key, value) }),
  removeItem: vi.fn((key: string) => { storage.delete(key) }),
  clear: vi.fn(() => { storage.clear() }),
})

describe('useRQL', () => {
  let instance: ReturnType<typeof useRQL>

  beforeEach(() => {
    vi.clearAllMocks()
    storage.clear()
    instance = useRQL()
    instance.rql.value = ''
    instance.results.value = []
    instance.total.value = 0
    instance.error.value = null
  })

  describe('initial state', () => {
    it('should have empty rql', () => {
      expect(instance.rql.value).toBe('')
    })
    it('should have loading false', () => {
      expect(instance.loading.value).toBe(false)
    })
    it('should have null error', () => {
      expect(instance.error.value).toBeNull()
    })
    it('should have empty results', () => {
      expect(instance.results.value).toEqual([])
    })
    it('should have total 0', () => {
      expect(instance.total.value).toBe(0)
    })
  })

  describe('getHistory', () => {
    it('should return empty array when no history stored', () => {
      expect(instance.getHistory()).toEqual([])
    })

    it('should return parsed history items', () => {
      const items = [
        { id: '1', rql: 'priority = "high"', timestamp: Date.now(), entityType: 'issue' },
      ]
      storage.set(HISTORY_KEY, JSON.stringify(items))
      expect(instance.getHistory()).toHaveLength(1)
    })

    it('should handle corrupt localStorage data', () => {
      storage.set(HISTORY_KEY, 'not-json')
      expect(instance.getHistory()).toEqual([])
    })
  })

  describe('addToHistory', () => {
    it('should add item to history', () => {
      instance.addToHistory('priority = "high"')
      const history = instance.getHistory()
      expect(history).toHaveLength(1)
      expect(history[0].rql).toBe('priority = "high"')
    })

    it('should not add empty RQL', () => {
      instance.addToHistory('')
      expect(instance.getHistory()).toHaveLength(0)
    })

    it('should deduplicate by RQL text', () => {
      instance.addToHistory('priority = "high"')
      instance.addToHistory('priority = "high"')
      expect(instance.getHistory()).toHaveLength(1)
    })

    it('should move duplicate to front', () => {
      instance.addToHistory('rql-a')
      instance.addToHistory('rql-b')
      instance.addToHistory('rql-a')
      const history = instance.getHistory()
      expect(history[0].rql).toBe('rql-a')
    })

    it('should respect MAX_HISTORY limit of 50', () => {
      for (let i = 0; i < 60; i++) {
        instance.addToHistory(`rql-${i}`)
      }
      expect(instance.getHistory().length).toBeLessThanOrEqual(50)
    })
  })

  describe('clearHistory', () => {
    it('should remove all history from localStorage', () => {
      instance.addToHistory('rql-a')
      instance.addToHistory('rql-b')
      instance.clearHistory()
      expect(instance.getHistory()).toEqual([])
    })
  })

  describe('search', () => {
    it('should not call API when RQL is empty', async () => {
      instance.rql.value = ''
      await instance.search(1)
      expect(rqlApi.search).not.toHaveBeenCalled()
      expect(instance.results.value).toEqual([])
      expect(instance.total.value).toBe(0)
    })

    it('should call API with correct params', async () => {
      const mockResponse = {
        data: {
          success: true,
          data: { items: [{ id: 1, name: 'Issue 1' }], total: 1, page: 1, page_size: 20 },
        },
      }
      const mockSearch = rqlApi.search as ReturnType<typeof vi.fn>
      mockSearch.mockResolvedValue(mockResponse)

      instance.rql.value = 'priority = "high"'
      await instance.search(1, 'issue', 2, 10)

      expect(mockSearch).toHaveBeenCalledTimes(1)
      const callArg = mockSearch.mock.calls[0][0]
      expect(callArg.entity).toBe('issue')
      expect(callArg.project_id).toBe(1)
      expect(callArg.rql).toBe('priority = "high"')
      expect(callArg.page).toBe(2)
      expect(callArg.page_size).toBe(10)
    })

    it('should update results on success', async () => {
      const mockResponse = {
        data: {
          success: true,
          data: { items: [{ id: 1 }, { id: 2 }], total: 2, page: 1, page_size: 20 },
        },
      }
      const mockSearch = rqlApi.search as ReturnType<typeof vi.fn>
      mockSearch.mockResolvedValue(mockResponse)

      instance.rql.value = 'state_id = 1'
      await instance.search(1)

      expect(instance.results.value).toEqual([{ id: 1 }, { id: 2 }])
      expect(instance.total.value).toBe(2)
      expect(instance.loading.value).toBe(false)
    })

    it('should set error on failure response', async () => {
      const mockResponse = {
        data: {
          success: false,
          error: { code: 'PARSE_ERROR', message: 'Invalid RQL syntax' },
        },
      }
      const mockSearch = rqlApi.search as ReturnType<typeof vi.fn>
      mockSearch.mockResolvedValue(mockResponse)

      instance.rql.value = 'invalid'
      await instance.search(1)

      expect(instance.error.value).toBe('Invalid RQL syntax')
    })

    it('should handle network error', async () => {
      const mockSearch = rqlApi.search as ReturnType<typeof vi.fn>
      mockSearch.mockRejectedValue({ message: 'Network error' })

      instance.rql.value = 'priority = "high"'
      await instance.search(1)

      expect(instance.error.value).toBe('Network error')
    })

    it('should handle axios error with response', async () => {
      const mockSearch = rqlApi.search as ReturnType<typeof vi.fn>
      mockSearch.mockRejectedValue({
        response: { data: { error: { message: 'Server error' } } },
      })

      instance.rql.value = 'test'
      await instance.search(1)

      expect(instance.error.value).toBe('Server error')
    })

    it('should add to history after successful search', async () => {
      const mockSearch = rqlApi.search as ReturnType<typeof vi.fn>
      mockSearch.mockResolvedValue({
        data: { success: true, data: { items: [], total: 0, page: 1, page_size: 20 } },
      })

      instance.rql.value = 'cycle_id = 5'
      await instance.search(1, 'cycle')

      const history = instance.getHistory()
      expect(history).toHaveLength(1)
      expect(history[0].rql).toBe('cycle_id = 5')
      expect(history[0].entityType).toBe('cycle')
    })
  })
})
