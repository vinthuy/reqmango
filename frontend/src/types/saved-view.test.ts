/**
 * SavedView 类型单元测试
 */
import { describe, it, expect } from 'vitest'
import type { SavedView, SavedViewCreate, SavedViewUpdate, SortConfigEntry } from './saved-view'

describe('SortConfigEntry', () => {
  it('should accept asc direction', () => {
    const s: SortConfigEntry = { field: 'created_at', dir: 'asc' }
    expect(s.dir).toBe('asc')
  })
  it('should accept desc direction', () => {
    const s: SortConfigEntry = { field: 'priority', dir: 'desc' }
    expect(s.dir).toBe('desc')
  })
})

describe('SavedView', () => {
  const fullView: SavedView = {
    id: 1, name: 'My View', view_type: 'list',
    filters: { priority: 'high' },
    rql: 'priority = "high"',
    sort_config: [{ field: 'created_at', dir: 'desc' }],
    columns: ['name', 'state', 'priority'],
    group_by: 'state_id', sub_group_by: 'priority',
    is_default: false, is_shared: false,
    owner_id: 1, project_id: 1,
    created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z',
  }

  it('should accept all fields', () => {
    expect(fullView.id).toBe(1)
    expect(fullView.sort_config).toHaveLength(1)
    expect(fullView.columns).toContain('name')
  })

  it('should support all 5 view types', () => {
    const types: SavedView['view_type'][] = ['list', 'kanban', 'tree', 'gantt', 'calendar']
    types.forEach(t => {
      const v: SavedView = { ...fullView, view_type: t }
      expect(v.view_type).toBe(t)
    })
  })

  it('should allow optional rql', () => {
    const v: SavedView = { ...fullView, rql: undefined }
    expect(v.rql).toBeUndefined()
  })
})

describe('SavedViewCreate', () => {
  it('should require only name', () => {
    const c: SavedViewCreate = { name: 'New View' }
    expect(c.name).toBe('New View')
    expect(c.view_type).toBeUndefined()
  })

  it('should accept full create payload', () => {
    const c: SavedViewCreate = {
      name: 'Full View',
      description: 'A complete view',
      view_type: 'kanban',
      filters: { state_id: [1, 2] },
      rql: 'state_id IN [1, 2]',
      sort_config: [{ field: 'target_date', dir: 'asc' }],
      columns: ['name', 'assignee'],
      group_by: 'assignee_id',
      sub_group_by: 'priority',
      is_shared: true,
    }
    expect(c.view_type).toBe('kanban')
    expect(c.is_shared).toBe(true)
  })
})

describe('SavedViewUpdate', () => {
  it('should allow empty update', () => {
    const u: SavedViewUpdate = {}
    expect(u).toBeDefined()
  })

  it('should allow partial updates', () => {
    const u: SavedViewUpdate = { sort_config: [], columns: ['name'] }
    expect(u.columns).toHaveLength(1)
    expect(u.sort_config).toEqual([])
  })
})
