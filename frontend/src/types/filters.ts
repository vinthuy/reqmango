export interface FilterCondition {
  field: string
  operator: string
  value: any
  displayValue: string
}

export interface FilterField {
  key: string
  dbKey: string
  labelKey: string
  type: 'select' | 'multi' | 'date' | 'text' | 'number' | 'date_range'
  valueType: 'string' | 'number' | 'date'
  operators: string[]
}

export interface SortOption {
  key: string
  labelKey: string
  direction: 'asc' | 'desc'
}

export interface GroupOption {
  key: string
  labelKey: string
}

export const FILTER_FIELDS: FilterField[] = [
  { key: 'sequence_id', dbKey: 'sequence_id', labelKey: 'filter.fieldSequenceId', type: 'number', valueType: 'number', operators: ['is', 'is not'] },
  { key: 'title', dbKey: 'name', labelKey: 'filter.fieldTitle', type: 'text', valueType: 'string', operators: ['is', 'is not', 'contains', 'does not contain'] },
  { key: 'state_id', dbKey: 'state_id', labelKey: 'filter.fieldState', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'state_group', dbKey: 'state_group', labelKey: 'filter.fieldStateGroup', type: 'select', valueType: 'string', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'priority', dbKey: 'priority', labelKey: 'filter.fieldPriority', type: 'select', valueType: 'string', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'assignee_id', dbKey: 'assignee_id', labelKey: 'filter.fieldAssignee', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty'] },
  { key: 'label', dbKey: 'label', labelKey: 'filter.fieldLabel', type: 'multi', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty'] },
  { key: 'cycle_id', dbKey: 'cycle_id', labelKey: 'filter.fieldCycle', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty'] },
  { key: 'module_id', dbKey: 'module_id', labelKey: 'filter.fieldModule', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty'] },
  { key: 'type_id', dbKey: 'issue_type_id', labelKey: 'filter.fieldType', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'start_date', dbKey: 'start_date', labelKey: 'filter.fieldStartDate', type: 'date', valueType: 'date', operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty'] },
  { key: 'target_date', dbKey: 'target_date', labelKey: 'filter.fieldTargetDate', type: 'date', valueType: 'date', operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty'] },
  { key: 'created_at', dbKey: 'created_at', labelKey: 'filter.fieldCreatedAt', type: 'date', valueType: 'date', operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty'] },
]

function formatValue(value: any, valueType: 'string' | 'number' | 'date'): string {
  if (valueType === 'number') {
    return String(value)
  }
  return `"${value}"`
}

export interface RQLResult {
  filters: FilterCondition[]
  sortBy?: SortOption
}

export function buildRQL(filters: FilterCondition[], quickSearchValue?: string): string {
  const clauses: string[] = []

  if (quickSearchValue) {
    const qs = quickSearchValue.trim()
    if (qs) {
      const issueKeyMatch = qs.match(/^[A-Z]+-\d+$/)
      if (issueKeyMatch) {
        const parts = qs.split('-')
        const sequenceId = parts[1]
        clauses.push(`sequence_id = ${sequenceId}`)
      } else {
        clauses.push(`(name LIKE "${qs}" OR description LIKE "${qs}")`)
      }
    }
  }

  if (filters.length === 0 && !quickSearchValue) return ''

  for (const filter of filters) {
    let clause = ''
    const { field, operator, value } = filter

    const fieldDef = FILTER_FIELDS.find(f => f.key === field)
    const dbKey = fieldDef?.dbKey || field
    const valueType = fieldDef?.valueType || 'string'

    switch (operator) {
      case 'is':
        clause = `${dbKey} = ${formatValue(value, valueType)}`
        break
      case 'is not':
        clause = `${dbKey} != ${formatValue(value, valueType)}`
        break
      case 'is any of':
        const anyOfValues = Array.isArray(value) ? value : [value]
        const formattedAnyOf = anyOfValues.map(v => formatValue(v, valueType)).join(', ')
        clause = `${dbKey} IN (${formattedAnyOf})`
        break
      case 'is not any of':
        const notAnyOfValues = Array.isArray(value) ? value : [value]
        const formattedNotAnyOf = notAnyOfValues.map(v => formatValue(v, valueType)).join(', ')
        clause = `${dbKey} NOT IN (${formattedNotAnyOf})`
        break
      case 'contains':
        clause = `${dbKey} LIKE "%${value}%"`
        break
      case 'does not contain':
        clause = `${dbKey} NOT LIKE "%${value}%"`
        break
      case 'is empty':
        clause = `${dbKey} IS NULL`
        break
      case 'is not empty':
        clause = `${dbKey} IS NOT NULL`
        break
      case 'before':
        clause = `${dbKey} < "${value}"`
        break
      case 'after':
        clause = `${dbKey} > "${value}"`
        break
      case 'before or on':
        clause = `${dbKey} <= "${value}"`
        break
      case 'after or on':
        clause = `${dbKey} >= "${value}"`
        break
      case 'between':
        clause = `${dbKey} >= "${value[0]}" AND ${dbKey} <= "${value[1]}"`
        break
      case 'not between':
        clause = `${dbKey} < "${value[0]}" OR ${dbKey} > "${value[1]}"`
        break
      }

    if (clause) clauses.push(clause)
  }

  return clauses.join(' AND ')
}

export function parseRQL(rqlStr: string): RQLResult {
  if (!rqlStr.trim()) return { filters: [] }

  const conditions: FilterCondition[] = []
  let sortBy: SortOption | undefined

  const clauses = rqlStr.split(' AND ')

  for (const clause of clauses) {
    const trimmed = clause.trim()
    if (!trimmed) continue

    const orderbyMatch = trimmed.match(/^orderby (\w+) (asc|desc)$/i)
    if (orderbyMatch) {
      sortBy = {
        key: orderbyMatch[1],
        labelKey: SORT_OPTIONS.find(s => s.key === orderbyMatch[1])?.labelKey || '',
        direction: orderbyMatch[2].toLowerCase() as 'asc' | 'desc'
      }
      continue
    }

    let field: string = ''
    let operator: string = ''
    let value: any = ''
    let displayValue: string = ''

    const likeMatch = trimmed.match(/^(\w+) (NOT LIKE|LIKE) "%([^"]+)"$/)
    if (likeMatch) {
      field = likeMatch[1]
      operator = likeMatch[2] === 'LIKE' ? 'contains' : 'does not contain'
      // Strip trailing % from value (buildRQL wraps value with %...%)
      let rawValue = likeMatch[3] || ''
      value = rawValue.replace(/%$/, '')
      displayValue = value
      conditions.push({ field, operator, value, displayValue })
      continue
    }

    const inMatch = trimmed.match(/^(\w+) (NOT IN|IN) \((.*)\)$/)
    if (inMatch) {
      field = inMatch[1]
      operator = inMatch[2] === 'IN' ? 'is any of' : 'is not any of'
      const valuesStr = inMatch[3]
      const values = valuesStr.split(',').map(v => {
        const trimmedV = v.trim()
        const quotedMatch = trimmedV.match(/^"([^"]+)"$/)
        return quotedMatch ? quotedMatch[1] : trimmedV
      })
      value = values
      displayValue = values.join(', ')
      conditions.push({ field, operator, value, displayValue })
      continue
    }

    const nullMatch = trimmed.match(/^(\w+) (IS NOT NULL|IS NULL)$/)
    if (nullMatch) {
      field = nullMatch[1]
      operator = nullMatch[2] === 'IS NULL' ? 'is empty' : 'is not empty'
      value = ''
      displayValue = ''
      conditions.push({ field, operator, value, displayValue })
      continue
    }

    // Try quoted value (strings, dates)
    const quotedMatch = trimmed.match(/^(\w+) ([=<>!]+) "([^"]+)"$/)
    if (quotedMatch) {
      field = quotedMatch[1]
      const op = quotedMatch[2]
      value = quotedMatch[3]
      displayValue = value

      switch (op) {
        case '=':  operator = 'is'; break
        case '!=': operator = 'is not'; break
        case '>':  operator = 'after'; break
        case '<':  operator = 'before'; break
        case '>=': operator = 'after or on'; break
        case '<=': operator = 'before or on'; break
      }

      if (operator) {
        conditions.push({ field, operator, value, displayValue })
      }
      continue
    }

    // Try unquoted value (numbers)
    const unquotedMatch = trimmed.match(/^(\w+) ([=<>!]+) (\S+)$/)
    if (unquotedMatch) {
      field = unquotedMatch[1]
      const op = unquotedMatch[2]
      value = unquotedMatch[3]
      displayValue = value

      switch (op) {
        case '=':  operator = 'is'; break
        case '!=': operator = 'is not'; break
        case '>':  operator = 'after'; break
        case '<':  operator = 'before'; break
        case '>=': operator = 'after or on'; break
        case '<=': operator = 'before or on'; break
      }

      if (operator) {
        conditions.push({ field, operator, value, displayValue })
      }
      continue
    }
  }

  return { filters: conditions, sortBy }
}

export const SORT_OPTIONS: SortOption[] = [
  { key: 'created_at', labelKey: 'filter.orderLastCreated', direction: 'desc' },
  { key: 'updated_at', labelKey: 'filter.orderLastUpdated', direction: 'desc' },
  { key: 'priority', labelKey: 'filter.orderPriority', direction: 'desc' },
  { key: 'start_date', labelKey: 'filter.orderStartDate', direction: 'asc' },
  { key: 'target_date', labelKey: 'filter.orderDueDate', direction: 'asc' },
]

export const GROUP_OPTIONS: GroupOption[] = [
  { key: 'none', labelKey: 'filter.groupByNone' },
  { key: 'state_id', labelKey: 'filter.groupByState' },
  { key: 'priority', labelKey: 'filter.groupByPriority' },
  { key: 'assignee_id', labelKey: 'filter.groupByAssignee' },
  { key: 'label', labelKey: 'filter.groupByLabel' },
  { key: 'cycle_id', labelKey: 'filter.groupByCycle' },
  { key: 'module_id', labelKey: 'filter.groupByModule' },
  { key: 'type_id', labelKey: 'filter.groupByType' },
]