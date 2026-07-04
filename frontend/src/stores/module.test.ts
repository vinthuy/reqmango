/**
 * Module Store 单元测试
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('@/api/module', () => ({
  default: {
    listModules: vi.fn(),
    getModuleTree: vi.fn(),
    createModule: vi.fn(),
    updateModule: vi.fn(),
    deleteModule: vi.fn(),
    addIssueToModule: vi.fn(),
    removeIssueFromModule: vi.fn(),
    getModuleIssues: vi.fn(),
    getModuleProgress: vi.fn(),
    getModuleStatistics: vi.fn(),
  },
}))

import { useModuleStore } from './module'
import moduleApi from '@/api/module'

const mockApi = moduleApi as any

function makeModule(id: number, overrides = {}) {
  return {
    id, name: `Module ${id}`, description: '',
    project_id: 1, workspace_id: 1, parent_id: null as number | null,
    order: 0, is_archived: false, archived_at: null,
    created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function makeTreeNode(id: number, children: any[] = []) {
  return {
    id, name: `Node ${id}`, description: '',
    project_id: 1, workspace_id: 1, parent_id: null as number | null,
    order: 0, is_archived: false, archived_at: null,
    created_at: '', updated_at: '',
    children, total_issues: 0, completed_issues: 0, progress: 0,
  }
}

describe('useModuleStore', () => {
  let store: ReturnType<typeof useModuleStore>

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    store = useModuleStore()
  })

  describe('Initial State', () => {
    it('should have empty modules array', () => {
      expect(store.modules).toEqual([])
    })
    it('should have empty moduleTree', () => {
      expect(store.moduleTree).toEqual([])
    })
    it('should have null currentModule', () => {
      expect(store.currentModule).toBeNull()
    })
    it('should not be loading', () => {
      expect(store.isLoading).toBe(false)
    })
    it('should have null error', () => {
      expect(store.error).toBeNull()
    })
  })

  // ==================== fetchModules ====================
  describe('fetchModules', () => {
    it('should fetch and set modules', async () => {
      const items = [makeModule(1), makeModule(2)]
      mockApi.listModules.mockResolvedValue(items)
      await store.fetchModules(1, 1)
      expect(store.modules).toHaveLength(2)
      expect(mockApi.listModules).toHaveBeenCalledWith(1, 1)
    })

    it('should set error on failure', async () => {
      mockApi.listModules.mockRejectedValue(new Error('Failed'))
      await store.fetchModules(1, 1)
      expect(store.error).toBe('Failed')
    })

    it('should set isLoading during fetch', async () => {
      mockApi.listModules.mockResolvedValue([])
      const promise = store.fetchModules(1, 1)
      expect(store.isLoading).toBe(true)
      await promise
      expect(store.isLoading).toBe(false)
    })
  })

  // ==================== fetchModuleTree ====================
  describe('fetchModuleTree', () => {
    it('should fetch tree structure', async () => {
      const tree = [makeTreeNode(1, [makeTreeNode(2)])]
      mockApi.getModuleTree.mockResolvedValue(tree)
      await store.fetchModuleTree(1)
      expect(store.moduleTree).toEqual(tree)
      expect(store.moduleTree[0].children).toHaveLength(1)
    })
  })

  // ==================== createModule ====================
  describe('createModule', () => {
    it('should create module and refresh tree', async () => {
      const created = makeModule(3)
      mockApi.createModule.mockResolvedValue(created)
      mockApi.getModuleTree.mockResolvedValue([])
      const result = await store.createModule(1, { name: 'New', project_id: 1, workspace_id: 1 })
      expect(store.modules).toHaveLength(1)
      expect(mockApi.getModuleTree).toHaveBeenCalledWith(1)
      expect(result).toEqual(created)
    })

    it('should return null on failure', async () => {
      mockApi.createModule.mockRejectedValue(new Error('Invalid'))
      const result = await store.createModule(1, { name: 'Fail', project_id: 1, workspace_id: 1 })
      expect(result).toBeNull()
    })
  })

  // ==================== updateModuleAction ====================
  describe('updateModuleAction', () => {
    it('should update in list and current, refresh tree', async () => {
      const orig = makeModule(1, { name: 'Old' })
      store.modules = [orig]
      store.currentModule = orig
      const updated = { ...orig, name: 'New' }
      mockApi.updateModule.mockResolvedValue(updated)
      mockApi.getModuleTree.mockResolvedValue([])
      await store.updateModuleAction(1, { name: 'New' })
      expect(store.modules[0].name).toBe('New')
      expect(store.currentModule?.name).toBe('New')
      expect(mockApi.getModuleTree).toHaveBeenCalledWith(1)
    })
  })

  // ==================== deleteModuleAction ====================
  describe('deleteModuleAction', () => {
    it('should remove from list and refresh tree', async () => {
      const m = makeModule(1)
      store.modules = [m]
      store.currentModule = m
      mockApi.deleteModule.mockResolvedValue(undefined)
      mockApi.getModuleTree.mockResolvedValue([])
      await store.deleteModuleAction(1)
      expect(store.modules).toHaveLength(0)
      expect(store.currentModule).toBeNull()
      expect(mockApi.getModuleTree).toHaveBeenCalledWith(1)
    })
  })

  // ==================== Issue Management ====================
  describe('addIssueToModule', () => {
    it('should add issue and refresh issues/progress', async () => {
      mockApi.addIssueToModule.mockResolvedValue({ module_id: 1, issue_id: 42, action: 'add' })
      mockApi.getModuleIssues.mockResolvedValue([])
      mockApi.getModuleProgress.mockResolvedValue({ module_id: 1, module_name: 'Core', total_issues: 0, completed: 0, progress: 0 })
      const result = await store.addIssueToModule(1, 42)
      expect(mockApi.getModuleIssues).toHaveBeenCalledWith(1, undefined)
      expect(mockApi.getModuleProgress).toHaveBeenCalledWith(1)
      expect(result).toEqual({ module_id: 1, issue_id: 42, action: 'add' })
    })
  })

  describe('removeIssueFromModule', () => {
    it('should remove issue and refresh progress', async () => {
      store.moduleIssues = [{ id: 1 }, { id: 2 }]
      mockApi.removeIssueFromModule.mockResolvedValue(undefined)
      mockApi.getModuleProgress.mockResolvedValue(null)
      await store.removeIssueFromModule(1, 1)
      expect(store.moduleIssues).toHaveLength(1)
    })
  })

  // ==================== fetchProgress ====================
  describe('fetchProgress', () => {
    it('should fetch module progress', async () => {
      const progress = { module_id: 1, module_name: 'Core', total_issues: 20, completed: 12, progress: 60 }
      mockApi.getModuleProgress.mockResolvedValue(progress)
      await store.fetchProgress(1)
      expect(store.progress).toEqual(progress)
    })
  })
})
