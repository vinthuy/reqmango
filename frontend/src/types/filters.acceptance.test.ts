import { describe, it, expect } from 'vitest'
import {
  buildRQL,
  parseRQL,
  FILTER_FIELDS,
  BUILT_IN_FUNCTIONS,
  flattenFilterGroups,
  type FilterCondition,
  type FilterGroup,
} from './filters'

describe('功能验收测试：所有系统字段的所有算子', () => {
  describe('sequence_id (number类型)', () => {
    const operators = ['is', 'is not']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'sequence_id', operator: 'is', value: '42', displayValue: '42' }]
      expect(buildRQL(filters)).toBe('sequence_id = 42')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'sequence_id', operator: 'is not', value: '100', displayValue: '100' }]
      expect(buildRQL(filters)).toBe('sequence_id != 100')
    })
  })

  describe('title (text类型)', () => {
    const operators = ['is', 'is not', 'contains', 'does not contain']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'title', operator: 'is', value: 'Fix bug', displayValue: 'Fix bug' }]
      expect(buildRQL(filters)).toBe('name = "Fix bug"')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'title', operator: 'is not', value: 'WIP', displayValue: 'WIP' }]
      expect(buildRQL(filters)).toBe('name != "WIP"')
    })
    
    it('should handle "contains" operator with wildcards', () => {
      const filters: FilterCondition[] = [{ field: 'title', operator: 'contains', value: 'bug', displayValue: 'bug' }]
      expect(buildRQL(filters)).toBe('name LIKE "%bug%"')
    })
    
    it('should handle "does not contain" operator with wildcards', () => {
      const filters: FilterCondition[] = [{ field: 'title', operator: 'does not contain', value: 'spam', displayValue: 'spam' }]
      expect(buildRQL(filters)).toBe('name NOT LIKE "%spam%"')
    })
  })

  describe('state_id (select类型，number值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' }]
      expect(buildRQL(filters)).toBe('state_id = 1')
    })
    
    it('should handle "is any of" operator with multiple values', () => {
      const filters: FilterCondition[] = [{ field: 'state_id', operator: 'is any of', value: [1, 2, 3], displayValue: 'Open, In Progress, Done' }]
      expect(buildRQL(filters)).toBe('state_id IN (1, 2, 3)')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_id', operator: 'is not', value: '4', displayValue: 'Cancelled' }]
      expect(buildRQL(filters)).toBe('state_id != 4')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_id', operator: 'is not any of', value: [4, 5], displayValue: 'Cancelled, Archived' }]
      expect(buildRQL(filters)).toBe('state_id NOT IN (4, 5)')
    })
  })

  describe('state_group (select类型，string值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_group', operator: 'is', value: 'started', displayValue: 'Started' }]
      expect(buildRQL(filters)).toBe('state_group = "started"')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_group', operator: 'is any of', value: ['backlog', 'started'], displayValue: 'Backlog, Started' }]
      expect(buildRQL(filters)).toBe('state_group IN ("backlog", "started")')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_group', operator: 'is not', value: 'cancelled', displayValue: 'Cancelled' }]
      expect(buildRQL(filters)).toBe('state_group != "cancelled"')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_group', operator: 'is not any of', value: ['cancelled', 'completed'], displayValue: 'Cancelled, Completed' }]
      expect(buildRQL(filters)).toBe('state_group NOT IN ("cancelled", "completed")')
    })
  })

  describe('priority (select类型，string值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'priority', operator: 'is', value: 'high', displayValue: 'High' }]
      expect(buildRQL(filters)).toBe('priority = "high"')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'priority', operator: 'is any of', value: ['high', 'urgent'], displayValue: 'High, Urgent' }]
      expect(buildRQL(filters)).toBe('priority IN ("high", "urgent")')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'priority', operator: 'is not', value: 'low', displayValue: 'Low' }]
      expect(buildRQL(filters)).toBe('priority != "low"')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'priority', operator: 'is not any of', value: ['none', 'low'], displayValue: 'None, Low' }]
      expect(buildRQL(filters)).toBe('priority NOT IN ("none", "low")')
    })
  })

  describe('assignee_id (select类型，number值，支持空值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is', value: '1', displayValue: 'Admin' }]
      expect(buildRQL(filters)).toBe('assignee_id = 1')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is any of', value: [1, 2], displayValue: 'Admin, User' }]
      expect(buildRQL(filters)).toBe('assignee_id IN (1, 2)')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is not', value: '3', displayValue: 'Guest' }]
      expect(buildRQL(filters)).toBe('assignee_id != 3')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is not any of', value: [4, 5], displayValue: 'User A, User B' }]
      expect(buildRQL(filters)).toBe('assignee_id NOT IN (4, 5)')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('assignee_id IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('assignee_id IS NOT NULL')
    })
  })

  describe('label (multi类型，number值，支持空值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'label', operator: 'is', value: '1', displayValue: 'Bug' }]
      expect(buildRQL(filters)).toBe('label = 1')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'label', operator: 'is any of', value: [1, 2, 3], displayValue: 'Bug, Feature, Enhancement' }]
      expect(buildRQL(filters)).toBe('label IN (1, 2, 3)')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'label', operator: 'is not', value: '4', displayValue: 'Duplicate' }]
      expect(buildRQL(filters)).toBe('label != 4')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'label', operator: 'is not any of', value: [5, 6], displayValue: 'Invalid, Wontfix' }]
      expect(buildRQL(filters)).toBe('label NOT IN (5, 6)')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'label', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('label IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'label', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('label IS NOT NULL')
    })
  })

  describe('cycle_id (select类型，number值，支持空值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cycle_id', operator: 'is', value: '1', displayValue: 'Sprint 1' }]
      expect(buildRQL(filters)).toBe('cycle_id = 1')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cycle_id', operator: 'is any of', value: [1, 2], displayValue: 'Sprint 1, Sprint 2' }]
      expect(buildRQL(filters)).toBe('cycle_id IN (1, 2)')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cycle_id', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cycle_id IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cycle_id', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cycle_id IS NOT NULL')
    })
  })

  describe('module_id (select类型，number值，支持空值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'module_id', operator: 'is', value: '1', displayValue: 'Backend' }]
      expect(buildRQL(filters)).toBe('module_id = 1')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'module_id', operator: 'is any of', value: [1, 2], displayValue: 'Backend, Frontend' }]
      expect(buildRQL(filters)).toBe('module_id IN (1, 2)')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'module_id', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('module_id IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'module_id', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('module_id IS NOT NULL')
    })
  })

  describe('type_id (select类型，number值)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of']
    
    it('should handle "is" operator with dbKey mapping', () => {
      const filters: FilterCondition[] = [{ field: 'type_id', operator: 'is', value: '1', displayValue: 'Bug' }]
      expect(buildRQL(filters)).toBe('issue_type_id = 1')
    })
    
    it('should handle "is any of" operator with dbKey mapping', () => {
      const filters: FilterCondition[] = [{ field: 'type_id', operator: 'is any of', value: [1, 2], displayValue: 'Bug, Feature' }]
      expect(buildRQL(filters)).toBe('issue_type_id IN (1, 2)')
    })
  })

  describe('日期字段 (date类型) - start_date', () => {
    const operators = ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'is', value: '2025-01-15', displayValue: '2025-01-15' }]
      expect(buildRQL(filters)).toBe('start_date = "2025-01-15"')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'is not', value: '2025-01-15', displayValue: '2025-01-15' }]
      expect(buildRQL(filters)).toBe('start_date != "2025-01-15"')
    })
    
    it('should handle "before" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'before', value: '2025-06-30', displayValue: '2025-06-30' }]
      expect(buildRQL(filters)).toBe('start_date < "2025-06-30"')
    })
    
    it('should handle "after" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'after', value: '2025-01-01', displayValue: '2025-01-01' }]
      expect(buildRQL(filters)).toBe('start_date > "2025-01-01"')
    })
    
    it('should handle "before or on" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'before or on', value: '2025-12-31', displayValue: '2025-12-31' }]
      expect(buildRQL(filters)).toBe('start_date <= "2025-12-31"')
    })
    
    it('should handle "after or on" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'after or on', value: '2025-01-01', displayValue: '2025-01-01' }]
      expect(buildRQL(filters)).toBe('start_date >= "2025-01-01"')
    })
    
    it('should handle "between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'between', value: ['2025-01-01', '2025-06-30'], displayValue: '2025-01-01 - 2025-06-30' }]
      expect(buildRQL(filters)).toBe('start_date >= "2025-01-01" AND start_date <= "2025-06-30"')
    })
    
    it('should handle "not between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'not between', value: ['2025-01-01', '2025-01-31'], displayValue: 'Jan 1-31' }]
      expect(buildRQL(filters)).toBe('start_date < "2025-01-01" OR start_date > "2025-01-31"')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('start_date IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'start_date', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('start_date IS NOT NULL')
    })
  })

  describe('日期字段 (date类型) - target_date', () => {
    it('should handle "between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'target_date', operator: 'between', value: ['2025-07-01', '2025-07-31'], displayValue: 'July' }]
      expect(buildRQL(filters)).toBe('target_date >= "2025-07-01" AND target_date <= "2025-07-31"')
    })
    
    it('should handle "not between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'target_date', operator: 'not between', value: ['2025-01-01', '2025-12-31'], displayValue: '2025' }]
      expect(buildRQL(filters)).toBe('target_date < "2025-01-01" OR target_date > "2025-12-31"')
    })
  })

  describe('日期字段 (date类型) - created_at', () => {
    it('should handle "after" operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_at', operator: 'after', value: '2025-06-01', displayValue: 'June 1, 2025' }]
      expect(buildRQL(filters)).toBe('created_at > "2025-06-01"')
    })
    
    it('should handle "before" operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_at', operator: 'before', value: '2025-07-01', displayValue: 'July 1, 2025' }]
      expect(buildRQL(filters)).toBe('created_at < "2025-07-01"')
    })
  })
})

describe('功能验收测试：所有自定义字段类型的所有算子', () => {
  describe('自定义文本字段 (cf_1, text类型)', () => {
    const operators = ['contains', 'does not contain', 'is empty', 'is not empty']
    
    it('should handle "contains" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_1', operator: 'contains', value: 'hello', displayValue: 'hello' }]
      expect(buildRQL(filters)).toBe('cf_1 LIKE "%hello%"')
    })
    
    it('should handle "does not contain" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_1', operator: 'does not contain', value: 'spam', displayValue: 'spam' }]
      expect(buildRQL(filters)).toBe('cf_1 NOT LIKE "%spam%"')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_1', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_1 IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_1', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_1 IS NOT NULL')
    })
  })

  describe('自定义数字字段 (cf_2, number类型)', () => {
    const operators = ['is', 'is not', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_2', operator: 'is', value: '42', displayValue: '42' }]
      expect(buildRQL(filters)).toBe('cf_2 = 42')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_2', operator: 'is not', value: '0', displayValue: '0' }]
      expect(buildRQL(filters)).toBe('cf_2 != 0')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_2', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_2 IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_2', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_2 IS NOT NULL')
    })
  })

  describe('自定义下拉字段 (cf_3, dropdown单选项)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is', value: 'option_1', displayValue: 'Option 1' }]
      expect(buildRQL(filters)).toBe('cf_3 = "option_1"')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is any of', value: ['opt_a', 'opt_b'], displayValue: 'A, B' }]
      expect(buildRQL(filters)).toBe('cf_3 IN ("opt_a", "opt_b")')
    })
    
    it('should handle "is not" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is not', value: 'opt_c', displayValue: 'C' }]
      expect(buildRQL(filters)).toBe('cf_3 != "opt_c"')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is not any of', value: ['opt_d', 'opt_e'], displayValue: 'D, E' }]
      expect(buildRQL(filters)).toBe('cf_3 NOT IN ("opt_d", "opt_e")')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_3 IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_3 IS NOT NULL')
    })
  })

  describe('自定义多选下拉字段 (cf_4, dropdown多选)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is any of" operator with multiple values', () => {
      const filters: FilterCondition[] = [{ field: 'cf_4', operator: 'is any of', value: ['tag1', 'tag2', 'tag3'], displayValue: 'Tag1, Tag2, Tag3' }]
      expect(buildRQL(filters)).toBe('cf_4 IN ("tag1", "tag2", "tag3")')
    })
    
    it('should handle "is not any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_4', operator: 'is not any of', value: ['exclude1', 'exclude2'], displayValue: 'Exclude1, Exclude2' }]
      expect(buildRQL(filters)).toBe('cf_4 NOT IN ("exclude1", "exclude2")')
    })
  })

  describe('自定义布尔字段 (cf_5, boolean类型)', () => {
    const operators = ['is']
    
    it('should handle "is" operator with true', () => {
      const filters: FilterCondition[] = [{ field: 'cf_5', operator: 'is', value: 'true', displayValue: 'Yes' }]
      expect(buildRQL(filters)).toBe('cf_5 = true')
    })
    
    it('should handle "is" operator with false', () => {
      const filters: FilterCondition[] = [{ field: 'cf_5', operator: 'is', value: 'false', displayValue: 'No' }]
      expect(buildRQL(filters)).toBe('cf_5 = false')
    })
  })

  describe('自定义日期字段 (cf_6, date类型)', () => {
    const operators = ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'is', value: '2025-03-15', displayValue: '2025-03-15' }]
      expect(buildRQL(filters)).toBe('cf_6 = "2025-03-15"')
    })
    
    it('should handle "before" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'before', value: '2025-04-01', displayValue: '2025-04-01' }]
      expect(buildRQL(filters)).toBe('cf_6 < "2025-04-01"')
    })
    
    it('should handle "after" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'after', value: '2025-02-01', displayValue: '2025-02-01' }]
      expect(buildRQL(filters)).toBe('cf_6 > "2025-02-01"')
    })
    
    it('should handle "between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'between', value: ['2025-01-01', '2025-06-30'], displayValue: 'H1 2025' }]
      expect(buildRQL(filters)).toBe('cf_6 >= "2025-01-01" AND cf_6 <= "2025-06-30"')
    })
    
    it('should handle "not between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'not between', value: ['2025-07-01', '2025-12-31'], displayValue: 'H2 2025' }]
      expect(buildRQL(filters)).toBe('cf_6 < "2025-07-01" OR cf_6 > "2025-12-31"')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_6 IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_6 IS NOT NULL')
    })
  })

  describe('自定义成员字段 (cf_7, member类型)', () => {
    const operators = ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty']
    
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_7', operator: 'is', value: '1', displayValue: 'Admin' }]
      expect(buildRQL(filters)).toBe('cf_7 = 1')
    })
    
    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_7', operator: 'is any of', value: [1, 2, 3], displayValue: 'Admin, User1, User2' }]
      expect(buildRQL(filters)).toBe('cf_7 IN (1, 2, 3)')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_7', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_7 IS NULL')
    })
    
    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_7', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_7 IS NOT NULL')
    })
  })

  describe('自定义URL字段 (cf_8, url类型)', () => {
    const operators = ['contains', 'does not contain', 'is empty', 'is not empty']
    
    it('should handle "contains" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_8', operator: 'contains', value: 'github', displayValue: 'github' }]
      expect(buildRQL(filters)).toBe('cf_8 LIKE "%github%"')
    })
    
    it('should handle "does not contain" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_8', operator: 'does not contain', value: 'bitbucket', displayValue: 'bitbucket' }]
      expect(buildRQL(filters)).toBe('cf_8 NOT LIKE "%bitbucket%"')
    })
    
    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_8', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('cf_8 IS NULL')
    })
  })
})

describe('功能验收测试：双向转换 (buildRQL -> parseRQL)', () => {
  describe('系统字段双向转换', () => {
    it('should round-trip title field with contains operator', () => {
      const filters: FilterCondition[] = [{ field: 'title', operator: 'contains', value: 'bug', displayValue: 'bug' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('name')
      expect(parsed.filters[0].operator).toBe('contains')
      expect(parsed.filters[0].value).toBe('bug')
    })

    it('should round-trip state_id field with is any of operator', () => {
      const filters: FilterCondition[] = [{ field: 'state_id', operator: 'is any of', value: [1, 2, 3], displayValue: 'Open, In Progress, Done' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('state_id')
      expect(parsed.filters[0].operator).toBe('is any of')
      expect(parsed.filters[0].value).toEqual(['1', '2', '3'])
    })

    it('should round-trip priority field with is operator', () => {
      const filters: FilterCondition[] = [{ field: 'priority', operator: 'is', value: 'high', displayValue: 'High' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('priority')
      expect(parsed.filters[0].operator).toBe('is')
      expect(parsed.filters[0].value).toBe('high')
    })

    it('should round-trip assignee_id field with is empty operator', () => {
      const filters: FilterCondition[] = [{ field: 'assignee_id', operator: 'is empty', value: '', displayValue: '' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('assignee_id')
      expect(parsed.filters[0].operator).toBe('is empty')
    })

    it('should round-trip type_id field with dbKey mapping', () => {
      const filters: FilterCondition[] = [{ field: 'type_id', operator: 'is', value: '1', displayValue: 'Bug' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('issue_type_id')
      expect(parsed.filters[0].operator).toBe('is')
      expect(parsed.filters[0].value).toBe('1')
    })

    it('should round-trip date field with between operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_at', operator: 'between', value: ['2025-01-01', '2025-06-30'], displayValue: '2025-01-01 - 2025-06-30' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('created_at')
      expect(parsed.filters[0].operator).toBe('between')
      expect(parsed.filters[0].value).toEqual(['2025-01-01', '2025-06-30'])
    })

    it('should round-trip date field with not between operator', () => {
      const filters: FilterCondition[] = [{ field: 'target_date', operator: 'not between', value: ['2025-01-01', '2025-01-31'], displayValue: 'Jan 1-31' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('target_date')
      expect(parsed.filters[0].operator).toBe('not between')
      expect(parsed.filters[0].value).toEqual(['2025-01-01', '2025-01-31'])
    })
  })

  describe('自定义字段双向转换', () => {
    it('should round-trip custom text field with contains operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_1', operator: 'contains', value: 'hello', displayValue: 'hello' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('cf_1')
      expect(parsed.filters[0].operator).toBe('contains')
      expect(parsed.filters[0].value).toBe('hello')
    })

    it('should round-trip custom number field with is operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_2', operator: 'is', value: '42', displayValue: '42' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('cf_2')
      expect(parsed.filters[0].operator).toBe('is')
      expect(parsed.filters[0].value).toBe('42')
    })

    it('should round-trip custom dropdown field with is any of operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_3', operator: 'is any of', value: ['opt_a', 'opt_b'], displayValue: 'A, B' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('cf_3')
      expect(parsed.filters[0].operator).toBe('is any of')
      expect(parsed.filters[0].value).toEqual(['opt_a', 'opt_b'])
    })

    it('should round-trip custom boolean field with is operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_5', operator: 'is', value: 'true', displayValue: 'Yes' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('cf_5')
      expect(parsed.filters[0].operator).toBe('is')
      expect(parsed.filters[0].value).toBe('true')
    })

    it('should round-trip custom date field with between operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_6', operator: 'between', value: ['2025-01-01', '2025-06-30'], displayValue: 'H1 2025' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('cf_6')
      expect(parsed.filters[0].operator).toBe('between')
      expect(parsed.filters[0].value).toEqual(['2025-01-01', '2025-06-30'])
    })

    it('should round-trip custom field with is empty operator', () => {
      const filters: FilterCondition[] = [{ field: 'cf_7', operator: 'is empty', value: '', displayValue: '' }]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0].field).toBe('cf_7')
      expect(parsed.filters[0].operator).toBe('is empty')
    })
  })

  describe('复杂组合条件双向转换', () => {
    it('should round-trip multiple filters with different operators', () => {
      const filters: FilterCondition[] = [
        { field: 'priority', operator: 'is', value: 'high', displayValue: 'High' },
        { field: 'title', operator: 'contains', value: 'bug', displayValue: 'bug' },
        { field: 'assignee_id', operator: 'is empty', value: '', displayValue: '' },
      ]
      const rql = buildRQL(filters)
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(3)
      expect(parsed.filters[0]).toMatchObject({ field: 'priority', operator: 'is', value: 'high' })
      expect(parsed.filters[1]).toMatchObject({ field: 'name', operator: 'contains', value: 'bug' })
      expect(parsed.filters[2]).toMatchObject({ field: 'assignee_id', operator: 'is empty' })
    })

    it('should round-trip date range with sort', () => {
      const filters: FilterCondition[] = [
        { field: 'created_at', operator: 'between', value: ['2025-01-01', '2025-12-31'], displayValue: '2025' },
      ]
      const rql = buildRQL(filters, undefined, undefined, [{ key: 'created_at', labelKey: '', direction: 'desc' }])
      const parsed = parseRQL(rql)
      expect(parsed.filters).toHaveLength(1)
      expect(parsed.filters[0]).toMatchObject({ field: 'created_at', operator: 'between' })
      expect(parsed.sortBy).toBeDefined()
      expect(parsed.sortBy).toHaveLength(1)
      expect(parsed.sortBy![0]).toMatchObject({ key: 'created_at', direction: 'desc' })
    })
  })
})

describe('功能验收测试：特殊字符和边界情况', () => {
  it('should handle special characters in LIKE patterns', () => {
    const filters: FilterCondition[] = [{ field: 'title', operator: 'contains', value: 'test%pattern_', displayValue: 'test%pattern_' }]
    const rql = buildRQL(filters)
    expect(rql).toBe('name LIKE "%test\\%pattern\\_%"')
  })

  it('should handle escaped quotes in text values', () => {
    const filters: FilterCondition[] = [{ field: 'title', operator: 'contains', value: 'hello "world"', displayValue: 'hello "world"' }]
    const rql = buildRQL(filters)
    expect(rql).toBe('name LIKE "%hello \\"world\\"%"')
  })

  it('should handle empty filters array', () => {
    expect(buildRQL([])).toBe('')
  })

  it('should handle null values gracefully', () => {
    const filters: FilterCondition[] = [{ field: 'title', operator: 'is', value: null, displayValue: '' }]
    expect(buildRQL(filters)).toBe('')
  })

  it('should handle undefined values gracefully', () => {
    const filters: FilterCondition[] = [{ field: 'title', operator: 'is', value: undefined, displayValue: '' }]
    expect(buildRQL(filters)).toBe('')
  })

  it('should handle empty array values gracefully', () => {
    const filters: FilterCondition[] = [{ field: 'state_id', operator: 'is any of', value: [], displayValue: '' }]
    expect(buildRQL(filters)).toBe('')
  })

  it('should parse empty RQL string', () => {
    const result = parseRQL('')
    expect(result.filters).toEqual([])
  })

  it('should parse whitespace-only RQL string', () => {
    const result = parseRQL('   ')
    expect(result.filters).toEqual([])
  })

  it('should handle quick search with special characters', () => {
    const rql = buildRQL([], 'test%pattern')
    expect(rql).toBe('(name LIKE "%test\\%pattern%" OR description LIKE "%test\\%pattern%")')
  })

  it('should handle issue key format in quick search', () => {
    const rql = buildRQL([], 'PROJ-123')
    expect(rql).toBe('sequence_id = 123')
  })
})

describe('功能验收测试：FILTER_FIELDS 完整性', () => {
  it('should have all expected system fields defined', () => {
    const expectedFields = ['sequence_id', 'title', 'state_id', 'state_group', 'priority', 'assignee_id', 'label', 'cycle_id', 'module_id', 'type_id', 'start_date', 'target_date', 'created_at', 'updated_at', 'created_by', 'milestone']
    const actualFields = FILTER_FIELDS.map(f => f.key)
    expect(actualFields).toEqual(expect.arrayContaining(expectedFields))
    expect(actualFields.length).toBe(expectedFields.length)
  })

  it('should have valid operator sets for each field type', () => {
    for (const field of FILTER_FIELDS) {
      expect(field.operators.length).toBeGreaterThan(0)
      
      const validOperators = new Set([
        'is', 'is not', 'contains', 'does not contain',
        'is any of', 'is not any of', 'is empty', 'is not empty',
        'before', 'after', 'before or on', 'after or on',
        'between', 'not between'
      ])
      
      for (const op of field.operators) {
        expect(validOperators.has(op)).toBe(true)
      }
    }
  })

  it('should have consistent dbKey mapping for fields that need it', () => {
    const mappedFields = FILTER_FIELDS.filter(f => f.key !== f.dbKey)
    expect(mappedFields).toHaveLength(2)
    expect(mappedFields.find(f => f.key === 'title')?.dbKey).toBe('name')
    expect(mappedFields.find(f => f.key === 'type_id')?.dbKey).toBe('issue_type_id')
  })
})

describe('功能验收测试：新增字段 (updated_at, created_by, milestone)', () => {
  describe('updated_at (date类型)', () => {
    it('should handle "after" operator', () => {
      const filters: FilterCondition[] = [{ field: 'updated_at', operator: 'after', value: '2025-06-01', displayValue: 'June 1, 2025' }]
      expect(buildRQL(filters)).toBe('updated_at > "2025-06-01"')
    })

    it('should handle "before" operator', () => {
      const filters: FilterCondition[] = [{ field: 'updated_at', operator: 'before', value: '2025-07-01', displayValue: 'July 1, 2025' }]
      expect(buildRQL(filters)).toBe('updated_at < "2025-07-01"')
    })

    it('should handle "between" operator', () => {
      const filters: FilterCondition[] = [{ field: 'updated_at', operator: 'between', value: ['2025-01-01', '2025-06-30'], displayValue: 'H1 2025' }]
      expect(buildRQL(filters)).toBe('updated_at >= "2025-01-01" AND updated_at <= "2025-06-30"')
    })
  })

  describe('created_by (select类型，number值，支持空值)', () => {
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_by', operator: 'is', value: '1', displayValue: 'Admin' }]
      expect(buildRQL(filters)).toBe('created_by = 1')
    })

    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_by', operator: 'is any of', value: [1, 2], displayValue: 'Admin, User' }]
      expect(buildRQL(filters)).toBe('created_by IN (1, 2)')
    })

    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_by', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('created_by IS NULL')
    })

    it('should handle "is not empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'created_by', operator: 'is not empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('created_by IS NOT NULL')
    })
  })

  describe('milestone (select类型，number值，支持空值)', () => {
    it('should handle "is" operator', () => {
      const filters: FilterCondition[] = [{ field: 'milestone', operator: 'is', value: '1', displayValue: 'V1.0' }]
      expect(buildRQL(filters)).toBe('milestone = 1')
    })

    it('should handle "is any of" operator', () => {
      const filters: FilterCondition[] = [{ field: 'milestone', operator: 'is any of', value: [1, 2, 3], displayValue: 'V1.0, V1.1, V2.0' }]
      expect(buildRQL(filters)).toBe('milestone IN (1, 2, 3)')
    })

    it('should handle "is empty" operator', () => {
      const filters: FilterCondition[] = [{ field: 'milestone', operator: 'is empty', value: '', displayValue: '' }]
      expect(buildRQL(filters)).toBe('milestone IS NULL')
    })
  })
})

describe('功能验收测试：OR逻辑支持', () => {
  describe('buildRQL with OR groups', () => {
    it('should build RQL with simple OR group', () => {
      const groups: FilterGroup[] = [{
        operator: 'OR',
        children: [
          { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
          { field: 'state_id', operator: 'is', value: '2', displayValue: 'In Progress' },
        ]
      }]
      const rql = buildRQL(groups)
      expect(rql).toBe('(state_id = 1 OR state_id = 2)')
    })

    it('should build RQL with nested AND inside OR', () => {
      const groups: FilterGroup[] = [{
        operator: 'OR',
        children: [
          { field: 'priority', operator: 'is', value: 'high', displayValue: 'High' },
          {
            operator: 'AND',
            children: [
              { field: 'type_id', operator: 'is', value: '1', displayValue: 'Bug' },
              { field: 'assignee_id', operator: 'is empty', value: '', displayValue: '' },
            ]
          }
        ]
      }]
      const rql = buildRQL(groups)
      expect(rql).toBe('(priority = "high" OR (issue_type_id = 1 AND assignee_id IS NULL))')
    })

    it('should build RQL with nested OR inside AND', () => {
      const groups: FilterGroup[] = [{
        operator: 'AND',
        children: [
          {
            operator: 'OR',
            children: [
              { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
              { field: 'state_id', operator: 'is', value: '2', displayValue: 'In Progress' },
            ]
          },
          { field: 'priority', operator: 'is', value: 'high', displayValue: 'High' },
        ]
      }]
      const rql = buildRQL(groups)
      expect(rql).toBe('((state_id = 1 OR state_id = 2) AND priority = "high")')
    })

    it('should build RQL with complex nested groups', () => {
      const groups: FilterGroup[] = [{
        operator: 'OR',
        children: [
          {
            operator: 'AND',
            children: [
              { field: 'priority', operator: 'is', value: 'high', displayValue: 'High' },
              { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
            ]
          },
          {
            operator: 'AND',
            children: [
              { field: 'type_id', operator: 'is', value: '1', displayValue: 'Bug' },
              { field: 'assignee_id', operator: 'is', value: '1', displayValue: 'Admin' },
            ]
          }
        ]
      }]
      const rql = buildRQL(groups)
      expect(rql).toBe('((priority = "high" AND state_id = 1) OR (issue_type_id = 1 AND assignee_id = 1))')
    })

    it('should build RQL with OR and AND at same level', () => {
      const groups: FilterGroup[] = [{
        operator: 'OR',
        children: [
          { field: 'title', operator: 'contains', value: 'bug', displayValue: 'bug' },
          { field: 'title', operator: 'contains', value: 'error', displayValue: 'error' },
        ]
      }]
      const rql = buildRQL(groups)
      expect(rql).toBe('(name LIKE "%bug%" OR name LIKE "%error%")')
    })
  })

  describe('parseRQL with OR groups', () => {
    it('should parse simple OR condition', () => {
      const rql = '(priority = "high" OR priority = "urgent")'
      const parsed = parseRQL(rql)
      expect(parsed.filterGroups).toHaveLength(1)
      expect(parsed.filterGroups[0].operator).toBe('OR')
      expect(parsed.filterGroups[0].children).toHaveLength(2)
    })

    it('should parse nested AND inside OR', () => {
      const rql = '(priority = "high" OR (state_id = 1 AND assignee_id IS NULL))'
      const parsed = parseRQL(rql)
      expect(parsed.filterGroups).toHaveLength(1)
      expect(parsed.filterGroups[0].operator).toBe('OR')
      expect(parsed.filterGroups[0].children).toHaveLength(2)
      expect(parsed.filterGroups[0].children[1]).toEqual({
        operator: 'AND',
        children: expect.arrayContaining([
          expect.objectContaining({ field: 'state_id', operator: 'is' }),
          expect.objectContaining({ field: 'assignee_id', operator: 'is empty' }),
        ])
      })
    })

    it('should parse nested OR inside AND', () => {
      const rql = '((state_id = 1 OR state_id = 2) AND priority = "high")'
      const parsed = parseRQL(rql)
      expect(parsed.filterGroups).toHaveLength(1)
      expect(parsed.filterGroups[0].operator).toBe('AND')
      expect(parsed.filterGroups[0].children).toHaveLength(2)
    })

    it('should parse complex nested groups', () => {
      const rql = '((priority = "high" AND state_id = 1) OR (issue_type_id = 1 AND assignee_id = 1))'
      const parsed = parseRQL(rql)
      expect(parsed.filterGroups).toHaveLength(1)
      expect(parsed.filterGroups[0].operator).toBe('OR')
      expect(parsed.filterGroups[0].children).toHaveLength(2)
    })
  })

  describe('OR groups round-trip', () => {
    it('should round-trip simple OR group', () => {
      const groups: FilterGroup[] = [{
        operator: 'OR',
        children: [
          { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
          { field: 'state_id', operator: 'is', value: '2', displayValue: 'In Progress' },
        ]
      }]
      const rql = buildRQL(groups)
      const parsed = parseRQL(rql)
      expect(parsed.filterGroups).toHaveLength(1)
      expect(parsed.filterGroups[0].operator).toBe('OR')
      expect(parsed.filterGroups[0].children).toHaveLength(2)
    })

    it('should round-trip nested AND inside OR', () => {
      const groups: FilterGroup[] = [{
        operator: 'OR',
        children: [
          { field: 'priority', operator: 'is', value: 'high', displayValue: 'High' },
          {
            operator: 'AND',
            children: [
              { field: 'type_id', operator: 'is', value: '1', displayValue: 'Bug' },
              { field: 'assignee_id', operator: 'is empty', value: '', displayValue: '' },
            ]
          }
        ]
      }]
      const rql = buildRQL(groups)
      const parsed = parseRQL(rql)
      expect(parsed.filterGroups).toHaveLength(1)
      expect(parsed.filterGroups[0].operator).toBe('OR')
    })
  })
})

describe('功能验收测试：内置函数 (Built-in Functions)', () => {
  it('should have all expected built-in functions defined', () => {
    const expectedFunctions = ['isOverdue', 'hasNoAssignee', 'hasNoLabel', 'isTopLevel', 'isSubWorkItem', 'hasChildren', 'hasStartAndDueDates']
    const actualFunctions = BUILT_IN_FUNCTIONS.map(f => f.name)
    expect(actualFunctions).toEqual(expect.arrayContaining(expectedFunctions))
    expect(actualFunctions.length).toBe(expectedFunctions.length)
  })

  it('should have valid label and description for each function', () => {
    for (const func of BUILT_IN_FUNCTIONS) {
      expect(func.label).toBeDefined()
      expect(func.label).toBeTruthy()
      expect(func.description).toBeDefined()
      expect(func.description).toBeTruthy()
    }
  })

  it('BUILT_IN_FUNCTIONS should have correct definitions', () => {
    const hasNoAssignee = BUILT_IN_FUNCTIONS.find(f => f.name === 'hasNoAssignee')
    expect(hasNoAssignee).toBeDefined()
    expect(hasNoAssignee?.label).toBe('filter.fnHasNoAssignee')
    expect(hasNoAssignee?.description).toBe('无负责人')
  })

  it('BUILT_IN_FUNCTIONS should have isOverdue function', () => {
    const isOverdue = BUILT_IN_FUNCTIONS.find(f => f.name === 'isOverdue')
    expect(isOverdue).toBeDefined()
    expect(isOverdue?.label).toBe('filter.fnIsOverdue')
    expect(isOverdue?.description).toBe('截止日期已过且状态为开放')
  })
})

describe('功能验收测试：flattenFilterGroups', () => {
  it('should flatten simple OR group', () => {
    const groups: FilterGroup[] = [{
      operator: 'OR',
      children: [
        { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
        { field: 'state_id', operator: 'is', value: '2', displayValue: 'In Progress' },
      ]
    }]
    const flattened = flattenFilterGroups(groups)
    expect(flattened).toHaveLength(2)
    expect(flattened[0]).toMatchObject({ field: 'state_id', operator: 'is', value: '1' })
    expect(flattened[1]).toMatchObject({ field: 'state_id', operator: 'is', value: '2' })
  })

  it('should flatten nested groups', () => {
    const groups: FilterGroup[] = [{
      operator: 'AND',
      children: [
        {
          operator: 'OR',
          children: [
            { field: 'priority', operator: 'is', value: 'high', displayValue: 'High' },
            { field: 'priority', operator: 'is', value: 'urgent', displayValue: 'Urgent' },
          ]
        },
        { field: 'state_id', operator: 'is', value: '1', displayValue: 'Open' },
      ]
    }]
    const flattened = flattenFilterGroups(groups)
    expect(flattened).toHaveLength(3)
  })

  it('should return empty array for empty groups', () => {
    const groups: FilterGroup[] = []
    const flattened = flattenFilterGroups(groups)
    expect(flattened).toEqual([])
  })
})