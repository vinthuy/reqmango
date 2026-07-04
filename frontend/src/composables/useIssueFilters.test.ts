/**
 * useIssueFilters Composable 单元测试
 * 覆盖：setFilter, removeFilter, clearAllFilters, buildFilterChips, getQueryParams
 */
import { describe, it, expect, beforeEach } from 'vitest'
import { useIssueFilters } from './useIssueFilters'

describe('useIssueFilters', () => {
  let instance: ReturnType<typeof useIssueFilters>

  beforeEach(() => {
    instance = useIssueFilters()
    instance.clearAllFilters()
  })

  describe('initial state', () => {
    it('should have empty search', () => {
      expect(instance.search.value).toBe('')
    })
    it('should have empty filters', () => {
      expect(Object.keys(instance.filters).length).toBe(0)
    })
    it('should have list view mode by default', () => {
      expect(instance.viewMode.value).toBe('list')
    })
    it('should have activeFilterCount of 0', () => {
      expect(instance.activeFilterCount.value).toBe(0)
    })
  })

  describe('setFilter', () => {
    it('should set a numeric filter', () => {
      instance.setFilter('state_id', 3)
      expect(instance.filters.state_id).toBe(3)
    })

    it('should set a string filter', () => {
      instance.setFilter('priority', 'urgent')
      expect(instance.filters.priority).toBe('urgent')
    })

    it('should remove filter when set to empty string', () => {
      instance.setFilter('priority', 'high')
      instance.setFilter('priority', '')
      expect(instance.filters.priority).toBeUndefined()
    })

    it('should remove filter when set to 0', () => {
      instance.setFilter('state_id', 5)
      instance.setFilter('state_id', 0)
      expect(instance.filters.state_id).toBeUndefined()
    })

    it('should remove filter when set to null', () => {
      instance.setFilter('state_id', 3)
      instance.setFilter('state_id', null)
      expect(instance.filters.state_id).toBeUndefined()
    })

    it('should remove filter when set to undefined', () => {
      instance.setFilter('state_id', 3)
      instance.setFilter('state_id', undefined)
      expect(instance.filters.state_id).toBeUndefined()
    })

    it('should remove filter when set to empty array', () => {
      instance.setFilter('label_ids', [1, 2])
      instance.setFilter('label_ids', [])
      expect(instance.filters.label_ids).toBeUndefined()
    })
  })

  describe('removeFilter', () => {
    it('should delete a filter by key', () => {
      instance.setFilter('state_id', 3)
      instance.setFilter('priority', 'high')
      instance.removeFilter('state_id')
      expect(instance.filters.state_id).toBeUndefined()
      expect(instance.filters.priority).toBe('high')
    })
  })

  describe('clearAllFilters', () => {
    it('should clear all filters and search', () => {
      instance.setFilter('state_id', 3)
      instance.setFilter('priority', 'high')
      instance.search.value = 'test query'
      instance.clearAllFilters()
      expect(Object.keys(instance.filters).length).toBe(0)
      expect(instance.search.value).toBe('')
    })
  })

  describe('buildFilterChips', () => {
    const states = [
      { id: 1, name: 'Backlog' },
      { id: 2, name: 'In Progress' },
    ]
    const cycles = [{ id: 10, name: 'Sprint 1' }]
    const labels = [
      { id: 1, name: 'bug' },
      { id: 2, name: 'feature' },
    ]
    const members = [
      { id: 5, display_name: 'Alice', user_id: 5 },
      { id: 6, display_name: 'Bob', user_id: 6 },
    ]

    it('should return state chip when state_id is set', () => {
      instance.setFilter('state_id', 1)
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      expect(chips.length).toBe(1)
      expect(chips[0]).toMatchObject({ key: 'state_id', label: '状态', displayValue: 'Backlog' })
    })

    it('should return priority chip', () => {
      instance.setFilter('priority', 'urgent')
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      expect(chips.length).toBe(1)
      expect(chips[0].displayValue).toBe('紧急')
    })

    it('should return cycle chip', () => {
      instance.setFilter('cycle_id', 10)
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      expect(chips.length).toBe(1)
      expect(chips[0].displayValue).toBe('Sprint 1')
    })

    it('should return assignee chip', () => {
      instance.setFilter('assignee_id', 5)
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      expect(chips.length).toBe(1)
      expect(chips[0].displayValue).toBe('Alice')
    })

    it('should return label chip', () => {
      instance.setFilter('label_ids', [1, 2])
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      expect(chips.length).toBe(1)
      expect(chips[0].displayValue).toBe('bug, feature')
    })

    it('should return "my issues" chip when assignee is "me"', () => {
      instance.filters.assignee_id = 'me' as any
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      const meChip = chips.find(c => c.value === 'me')
      expect(meChip).toBeDefined()
      expect(meChip!.displayValue).toBe('我的')
    })

    it('should return empty chips when no filters set', () => {
      const chips = instance.buildFilterChips(states, cycles, labels, members)
      expect(chips.length).toBe(0)
    })
  })

  describe('getQueryParams', () => {
    it('should return empty when no filters', () => {
      expect(instance.getQueryParams()).toEqual({})
    })

    it('should return state_id param', () => {
      instance.setFilter('state_id', 2)
      const params = instance.getQueryParams()
      expect(params.state_id).toBe(2)
    })

    it('should return priority param', () => {
      instance.setFilter('priority', 'urgent')
      expect(instance.getQueryParams().priority).toBe('urgent')
    })

    it('should return cycle_id param', () => {
      instance.setFilter('cycle_id', 10)
      expect(instance.getQueryParams().cycle_id).toBe(10)
    })

    it('should return assignee_id param', () => {
      instance.setFilter('assignee_id', 5)
      expect(instance.getQueryParams().assignee_id).toBe(5)
    })

    it('should return date params', () => {
      instance.setFilter('start_date', '2024-01-01')
      instance.setFilter('end_date', '2024-12-31')
      const params = instance.getQueryParams()
      expect(params.start_date).toBe('2024-01-01')
      expect(params.target_date).toBe('2024-12-31')
    })

    it('should return search param', () => {
      instance.search.value = 'login bug'
      expect(instance.getQueryParams().search).toBe('login bug')
    })

    it('should join label_ids with comma', () => {
      instance.setFilter('label_ids', [1, 2, 3])
      expect(instance.getQueryParams().label_ids).toBe('1,2,3')
    })
  })

  describe('activeFilterCount', () => {
    it('should count filters', () => {
      instance.setFilter('state_id', 1)
      instance.setFilter('priority', 'high')
      expect(instance.activeFilterCount.value).toBe(2)
    })

    it('should count search as a filter', () => {
      instance.search.value = 'test'
      expect(instance.activeFilterCount.value).toBe(1)
    })

    it('should count filters + search', () => {
      instance.setFilter('state_id', 1)
      instance.search.value = 'test'
      expect(instance.activeFilterCount.value).toBe(2)
    })
  })

  describe('viewMode', () => {
    it('should support kanban mode', () => {
      instance.viewMode.value = 'kanban'
      expect(instance.viewMode.value).toBe('kanban')
    })

    it('should support tree mode', () => {
      instance.viewMode.value = 'tree'
      expect(instance.viewMode.value).toBe('tree')
    })

    it('should support calendar mode', () => {
      instance.viewMode.value = 'calendar'
      expect(instance.viewMode.value).toBe('calendar')
    })

    it('should support gantt mode', () => {
      instance.viewMode.value = 'gantt'
      expect(instance.viewMode.value).toBe('gantt')
    })
  })
})
