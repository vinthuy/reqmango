import { describe, it, expect } from 'vitest'
import {
  buildRQL,
  parseRQL,
  FILTER_FIELDS,
  SORT_OPTIONS,
  GROUP_OPTIONS,
  type FilterCondition,
} from './filters'

describe('buildRQL', () => {
  it('should return empty string for no filters', () => {
    expect(buildRQL([])).toBe('')
  })

  it('should build simple equality clause', () => {
    const filters: FilterCondition[] = [{ field: 'priority', operator: 'is', value: 'high', displayValue: 'high' }]
    expect(buildRQL(filters)).toBe('priority = "high"')
  })

  it('should build "is not" clause', () => {
    const filters: FilterCondition[] = [{ field: 'priority', operator: 'is not', value: 'low', displayValue: 'low' }]
    expect(buildRQL(filters)).toBe('priority != "low"')
  })

  it('should build "contains" clause', () => {
    const filters: FilterCondition[] = [{ field: 'title', operator: 'contains', value: 'bug', displayValue: 'bug' }]
    expect(buildRQL(filters)).toBe('name LIKE "%bug%"')
  })

  it('should build "does not contain" clause', () => {
    const filters: FilterCondition[] = [{ field: 'title', operator: 'does not contain', value: 'spam', displayValue: 'spam' }]
    expect(buildRQL(filters)).toBe('name NOT LIKE "%spam%"')
  })

  it('should build "is empty" clause', () => {
    const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is empty', value: '', displayValue: '' }]
    expect(buildRQL(filters)).toBe('assignee_id IS NULL')
  })

  it('should build "is not empty" clause', () => {
    const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is not empty', value: '', displayValue: '' }]
    expect(buildRQL(filters)).toBe('assignee_id IS NOT NULL')
  })

  it('should build "is any of" clause', () => {
    const filters: FilterCondition[] = [
      { field: 'state_id', operator: 'is any of', value: [1, 2, 3], displayValue: 'TODO, Doing, Done' }
    ]
    expect(buildRQL(filters)).toBe('state_id IN (1, 2, 3)')
  })

  it('should build "is not any of" clause', () => {
    const filters: FilterCondition[] = [
      { field: 'state_id', operator: 'is not any of', value: [4, 5], displayValue: 'Cancelled, Archived' }
    ]
    expect(buildRQL(filters)).toBe('state_id NOT IN (4, 5)')
  })

  it('should build date "before" clause', () => {
    const filters: FilterCondition[] = [
      { field: 'target_date', operator: 'before', value: '2025-12-31', displayValue: '2025-12-31' }
    ]
    expect(buildRQL(filters)).toBe('target_date < "2025-12-31"')
  })

  it('should build date "after" clause', () => {
    const filters: FilterCondition[] = [
      { field: 'start_date', operator: 'after', value: '2025-01-01', displayValue: '2025-01-01' }
    ]
    expect(buildRQL(filters)).toBe('start_date > "2025-01-01"')
  })

  it('should build date "between" clause', () => {
    const filters: FilterCondition[] = [
      { field: 'created_at', operator: 'between', value: ['2025-01-01', '2025-06-30'], displayValue: '2025-01-01 - 2025-06-30' }
    ]
    expect(buildRQL(filters)).toBe('created_at >= "2025-01-01" AND created_at <= "2025-06-30"')
  })

  it('should build date "not between" clause', () => {
    const filters: FilterCondition[] = [
      { field: 'target_date', operator: 'not between', value: ['2025-01-01', '2025-01-31'], displayValue: 'Jan 1-31' }
    ]
    expect(buildRQL(filters)).toBe('target_date < "2025-01-01" OR target_date > "2025-01-31"')
  })

  it('should combine multiple filters with AND', () => {
    const filters: FilterCondition[] = [
      { field: 'priority', operator: 'is', value: 'urgent', displayValue: 'urgent' },
      { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
    ]
    expect(buildRQL(filters)).toBe('priority = "urgent" AND state_id = 1')
  })

  it('should handle custom field (cf_ prefix) by using field key as dbKey', () => {
    const filters: FilterCondition[] = [
      { field: 'cf_1', operator: 'contains', value: 'hello', displayValue: 'hello' }
    ]
    expect(buildRQL(filters)).toBe('cf_1 LIKE "%hello%"')
  })

  it('should handle custom field with "is" operator', () => {
    const filters: FilterCondition[] = [
      { field: 'cf_7', operator: 'is', value: 'active', displayValue: 'active' }
    ]
    expect(buildRQL(filters)).toBe('cf_7 = "active"')
  })

  it('should handle custom field with IS NULL', () => {
    const filters: FilterCondition[] = [
      { field: 'cf_3', operator: 'is empty', value: '', displayValue: '' }
    ]
    expect(buildRQL(filters)).toBe('cf_3 IS NULL')
  })

  it('should include quick search as LIKE on name and description', () => {
    const filters: FilterCondition[] = []
    expect(buildRQL(filters, 'login')).toBe('(name LIKE "%login%" OR description LIKE "%login%")')
  })

  it('should include quick search as issue key lookup for UPPER-123 format', () => {
    const filters: FilterCondition[] = []
    expect(buildRQL(filters, 'PROJ-42')).toBe('sequence_id = 42')
  })

  it('should combine quick search with filters', () => {
    const filters: FilterCondition[] = [
      { field: 'priority', operator: 'is', value: 'high', displayValue: 'high' }
    ]
    const rql = buildRQL(filters, 'bug')
    expect(rql).toContain('(name LIKE "%bug%" OR description LIKE "%bug%")')
    expect(rql).toContain('priority = "high"')
    expect(rql).toContain(' AND ')
  })

  it('should handle empty quick search by ignoring it', () => {
    const filters: FilterCondition[] = [
      { field: 'priority', operator: 'is', value: 'high', displayValue: 'high' }
    ]
    expect(buildRQL(filters, '')).toBe('priority = "high"')
    expect(buildRQL(filters, '  ')).toBe('priority = "high"')
  })

  it('should handle SYMBOL fields (like type_id) with correct db_key mapping', () => {
    const filters: FilterCondition[] = [
      { field: 'type_id', operator: 'is', value: '1', displayValue: 'Bug' }
    ]
    expect(buildRQL(filters)).toBe('issue_type_id = 1')
  })

  it('should return empty string when only quick search is empty string', () => {
    expect(buildRQL([], '')).toBe('')
  })
})

describe('parseRQL', () => {
  it('should return empty filters for empty string', () => {
    const result = parseRQL('')
    expect(result.filters).toEqual([])
  })

  it('should parse simple equality', () => {
    const result = parseRQL('priority = "high"')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'priority', operator: 'is', value: 'high' })
  })

  it('should parse simple inequality', () => {
    const result = parseRQL('priority != "low"')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'priority', operator: 'is not', value: 'low' })
  })

  it('should parse LIKE clause', () => {
    const result = parseRQL('name LIKE "%bug%"')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'name', operator: 'contains', value: 'bug' })
  })

  it('should parse NOT LIKE clause', () => {
    const result = parseRQL('name NOT LIKE "%spam%"')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'name', operator: 'does not contain', value: 'spam' })
  })

  it('should parse IN clause', () => {
    const result = parseRQL('state_id IN (1, 2, 3)')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'state_id', operator: 'is any of' })
    expect(result.filters[0].value).toEqual(['1', '2', '3'])
  })

  it('should parse NOT IN clause', () => {
    const result = parseRQL('state_id NOT IN (4, 5)')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'state_id', operator: 'is not any of' })
    expect(result.filters[0].value).toEqual(['4', '5'])
  })

  it('should parse IS NULL clause', () => {
    const result = parseRQL('assignee_id IS NULL')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'assignee_id', operator: 'is empty' })
  })

  it('should parse IS NOT NULL clause', () => {
    const result = parseRQL('assignee_id IS NOT NULL')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'assignee_id', operator: 'is not empty' })
  })

  it('should parse date comparison operators', () => {
    expect(parseRQL('target_date > "2025-01-01"').filters[0]).toMatchObject({ operator: 'after' })
    expect(parseRQL('target_date < "2025-01-01"').filters[0]).toMatchObject({ operator: 'before' })
    expect(parseRQL('target_date >= "2025-01-01"').filters[0]).toMatchObject({ operator: 'after or on' })
    expect(parseRQL('target_date <= "2025-01-01"').filters[0]).toMatchObject({ operator: 'before or on' })
  })

  it('should parse multiple AND-separated clauses', () => {
    const result = parseRQL('priority = "urgent" AND state_id = 1')
    expect(result.filters).toHaveLength(2)
  })

  it('should parse orderby clause', () => {
    const result = parseRQL('orderby created_at desc')
    expect(result.sortBy).toMatchObject({ key: 'created_at', direction: 'desc' })
  })

  it('should parse custom field keys (cf_ prefix)', () => {
    const result = parseRQL('cf_1 LIKE "%hello%"')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'cf_1', operator: 'contains', value: 'hello' })
  })

  it('should parse custom field with IS NULL', () => {
    const result = parseRQL('cf_3 IS NULL')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'cf_3', operator: 'is empty' })
  })

  it('should parse complex query with system and custom fields', () => {
    const result = parseRQL('priority = "urgent" AND cf_1 = "active"')
    expect(result.filters).toHaveLength(2)
    expect(result.filters[0]).toMatchObject({ field: 'priority', operator: 'is' })
    expect(result.filters[1]).toMatchObject({ field: 'cf_1', operator: 'is' })
  })

  it('should handle whitespace-only string', () => {
    expect(parseRQL('   ').filters).toEqual([])
  })

  it('should handle strings with quoted IN values', () => {
    const result = parseRQL('priority IN ("high", "urgent")')
    expect(result.filters).toHaveLength(1)
    expect(result.filters[0]).toMatchObject({ field: 'priority', operator: 'is any of' })
    expect(result.filters[0].value).toEqual(['high', 'urgent'])
  })
})

describe('Round-trip: buildRQL -> parseRQL', () => {
  it('should preserve filters through round trip', () => {
    const filters: FilterCondition[] = [
      { field: 'priority', operator: 'is', value: 'urgent', displayValue: 'urgent' },
      { field: 'title', operator: 'contains', value: 'login', displayValue: 'login' },
    ]
    const rql = buildRQL(filters)
    const parsed = parseRQL(rql)
    expect(parsed.filters).toHaveLength(2)
    // priority is a string field, preserved as-is
    expect(parsed.filters[0]).toMatchObject({ field: 'priority', operator: 'is', value: 'urgent' })
    // title's dbKey is 'name', parseRQL returns the dbKey
    expect(parsed.filters[1]).toMatchObject({ field: 'name', operator: 'contains', value: 'login' })
  })

  it('should preserve custom field filters through round trip', () => {
    const filters: FilterCondition[] = [
      { field: 'cf_5', operator: 'contains', value: 'alpha', displayValue: 'alpha' },
      { field: 'cf_7', operator: 'is', value: 'active', displayValue: 'active' },
    ]
    const rql = buildRQL(filters)
    const parsed = parseRQL(rql)
    expect(parsed.filters).toHaveLength(2)
    expect(parsed.filters[0]).toMatchObject({ field: 'cf_5', operator: 'contains', value: 'alpha' })
    expect(parsed.filters[1]).toMatchObject({ field: 'cf_7', operator: 'is', value: 'active' })
  })
})

describe('FILTER_FIELDS', () => {
  it('should contain all system fields', () => {
    const keys = FILTER_FIELDS.map(f => f.key)
    expect(keys).toContain('title')
    expect(keys).toContain('state_id')
    expect(keys).toContain('state_group')
    expect(keys).toContain('priority')
    expect(keys).toContain('assignee_id')
    expect(keys).toContain('label')
    expect(keys).toContain('cycle_id')
    expect(keys).toContain('module_id')
    expect(keys).toContain('type_id')
    expect(keys).toContain('start_date')
    expect(keys).toContain('target_date')
    expect(keys).toContain('created_at')
  })

  it('should have valid operators for each field', () => {
    for (const field of FILTER_FIELDS) {
      expect(field.operators.length).toBeGreaterThan(0)
      expect(field.type).toMatch(/^(select|multi|date|text)$/)
      expect(field.valueType).toMatch(/^(string|number|date)$/)
    }
  })

  it('should have label keys for i18n', () => {
    for (const field of FILTER_FIELDS) {
      expect(field.labelKey).toBeTruthy()
      expect(field.labelKey.startsWith('filter.')).toBe(true)
    }
  })
})

describe('SORT_OPTIONS', () => {
  it('should have sort options with valid fields', () => {
    expect(SORT_OPTIONS.length).toBeGreaterThan(0)
    for (const option of SORT_OPTIONS) {
      expect(option.key).toBeTruthy()
      expect(option.labelKey).toBeTruthy()
      expect(['asc', 'desc']).toContain(option.direction)
    }
  })
})

describe('GROUP_OPTIONS', () => {
  it('should have "none" as first option', () => {
    expect(GROUP_OPTIONS[0].key).toBe('none')
  })

  it('should have group options with valid fields', () => {
    const keys = GROUP_OPTIONS.map(g => g.key)
    expect(keys).toContain('state_id')
    expect(keys).toContain('priority')
    expect(keys).toContain('assignee_id')
    expect(keys).toContain('type_id')
  })
})

describe('buildRQL edge cases', () => {
  it('should handle number type values without quotes', () => {
    const filters: FilterCondition[] = [
      { field: 'state_id', operator: 'is', value: '3', displayValue: 'Done' }
    ]
    expect(buildRQL(filters)).toBe('state_id = 3')
  })

  it('should handle array values for "is any of"', () => {
    const filters: FilterCondition[] = [
      { field: 'label', operator: 'is any of', value: [10, 20], displayValue: 'bug, feature' }
    ]
    expect(buildRQL(filters)).toBe('label IN (10, 20)')
  })

  it('should handle single value for "is any of" gracefully', () => {
    const filters: FilterCondition[] = [
      { field: 'state_id', operator: 'is any of', value: '1', displayValue: 'Open' }
    ]
    // Should wrap in array
    const result = buildRQL(filters)
    expect(result).toContain('state_id IN')
  })
})
