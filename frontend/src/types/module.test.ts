/**
 * Module 类型单元测试
 */
import { describe, it, expect } from 'vitest'
import type {
  ModuleBase,
  ModuleCreate,
  ModuleUpdate,
  ModuleResponse,
  ModuleLite,
  ModuleProgress,
  ModuleStatistics,
  ModuleTreeNode,
} from './module'

// ==================== ModuleBase ====================
describe('ModuleBase', () => {
  it('should require name', () => {
    const m: ModuleBase = { name: 'Core' }
    expect(m.name).toBe('Core')
  })
  it('should accept optional description', () => {
    const m: ModuleBase = { name: 'Core', description: '核心模块' }
    expect(m.description).toBe('核心模块')
  })
})

// ==================== ModuleCreate ====================
describe('ModuleCreate', () => {
  it('should require name, project_id, workspace_id', () => {
    const m: ModuleCreate = { name: 'Auth', project_id: 1, workspace_id: 1 }
    expect(m.project_id).toBe(1)
  })
  it('should accept optional parent_id', () => {
    const m: ModuleCreate = { name: 'Sub Auth', project_id: 1, workspace_id: 1, parent_id: 5 }
    expect(m.parent_id).toBe(5)
  })
})

// ==================== ModuleUpdate ====================
describe('ModuleUpdate', () => {
  it('should allow all fields optional', () => {
    const m: ModuleUpdate = {}
    expect(m).toBeDefined()
  })
  it('should accept partial update', () => {
    const m: ModuleUpdate = { name: 'New Name' }
    expect(m.name).toBe('New Name')
  })
})

// ==================== ModuleResponse ====================
describe('ModuleResponse', () => {
  it('should have all required fields', () => {
    const m: ModuleResponse = {
      id: 1, name: 'API', description: 'API 模块',
      project_id: 1, workspace_id: 1, parent_id: null,
      order: 0, is_archived: false, archived_at: null,
      created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
    }
    expect(m.id).toBe(1)
    expect(m.is_archived).toBe(false)
    expect(m.parent_id).toBeNull()
  })
  it('should allow parent_id as number', () => {
    const m: ModuleResponse = {
      id: 2, name: 'Sub', description: '',
      project_id: 1, workspace_id: 1, parent_id: 1,
      order: 1, is_archived: false, archived_at: null,
      created_at: '', updated_at: '',
    }
    expect(m.parent_id).toBe(1)
  })
})

// ==================== ModuleLite ====================
describe('ModuleLite', () => {
  it('should have id and name', () => {
    const m: ModuleLite = { id: 1, name: 'Core' }
    expect(m.id).toBe(1)
    expect(m.name).toBe('Core')
  })
})

// ==================== ModuleProgress ====================
describe('ModuleProgress', () => {
  it('should have progress fields', () => {
    const p: ModuleProgress = {
      module_id: 1, module_name: 'Core',
      total_issues: 20, completed: 12, progress: 60,
    }
    expect(p.progress).toBe(60)
    expect(p.total_issues).toBe(20)
  })
})

// ==================== ModuleStatistics ====================
describe('ModuleStatistics', () => {
  it('should include priority and state breakdowns', () => {
    const s: ModuleStatistics = {
      module_id: 1, module_name: 'Core',
      total_issues: 50, active_issues: 20, completed: 25, cancelled: 5,
      by_priority: { high: 10, medium: 30, low: 10 },
      by_state: { Todo: 20, 'In Progress': 20, Done: 10 },
    }
    expect(s.by_priority.high).toBe(10)
    expect(s.by_state['Todo']).toBe(20)
  })
})

// ==================== ModuleTreeNode ====================
describe('ModuleTreeNode', () => {
  const leafNode: ModuleTreeNode = {
    id: 1, name: 'Leaf', description: '', project_id: 1, workspace_id: 1,
    parent_id: null, order: 0, is_archived: false, archived_at: null,
    created_at: '', updated_at: '',
    children: [], total_issues: 5, completed_issues: 3, progress: 60,
  }

  it('should require children array', () => {
    expect(leafNode.children).toEqual([])
  })
  it('should have progress and issue counts', () => {
    expect(leafNode.total_issues).toBe(5)
    expect(leafNode.completed_issues).toBe(3)
    expect(leafNode.progress).toBe(60)
  })
  it('should support nested children', () => {
    const parentNode: ModuleTreeNode = {
      ...leafNode,
      id: 0,
      children: [{ ...leafNode, id: 1, children: [] }],
    }
    expect(parentNode.children).toHaveLength(1)
    expect(parentNode.children[0].id).toBe(1)
  })
})
