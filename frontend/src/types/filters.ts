export interface FilterCondition {
  field: string
  operator: string
  value: any
  displayValue: string
}

export type FilterNode = FilterCondition | FilterGroup

export interface FilterGroup {
  operator: 'AND' | 'OR'
  children: FilterNode[]
}

export interface FilterField {
  key: string
  dbKey: string
  labelKey: string
  type: 'select' | 'multi' | 'date' | 'text' | 'number' | 'date_range'
  valueType: 'string' | 'number' | 'date' | 'boolean'
  operators: string[]
}

export interface SortOption {
  key: string
  labelKey: string
  direction: 'asc' | 'desc'
}

export interface SortConfigEntry {
  field: string
  dir: 'asc' | 'desc'
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
  { key: 'updated_at', dbKey: 'updated_at', labelKey: 'filter.fieldUpdatedAt', type: 'date', valueType: 'date', operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty', 'is not empty'] },
  { key: 'created_by', dbKey: 'created_by', labelKey: 'filter.fieldCreatedBy', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty'] },
  { key: 'milestone', dbKey: 'milestone', labelKey: 'filter.fieldMilestone', type: 'select', valueType: 'number', operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty', 'is not empty'] },
]

export const BUILT_IN_FUNCTIONS = [
  { name: 'isOverdue', label: 'filter.fnIsOverdue', description: '截止日期已过且状态为开放' },
  { name: 'hasNoAssignee', label: 'filter.fnHasNoAssignee', description: '无负责人' },
  { name: 'hasNoLabel', label: 'filter.fnHasNoLabel', description: '无标签' },
  { name: 'isTopLevel', label: 'filter.fnIsTopLevel', description: '非子工作项' },
  { name: 'isSubWorkItem', label: 'filter.fnIsSubWorkItem', description: '是子工作项' },
  { name: 'hasChildren', label: 'filter.fnHasChildren', description: '有子工作项' },
  { name: 'hasStartAndDueDates', label: 'filter.fnHasStartAndDueDates', description: '同时设置了开始和截止日期' },
]

function formatValue(value: any, valueType: 'string' | 'number' | 'date' | 'boolean'): string {
  if (valueType === 'number' || valueType === 'boolean') {
    return String(value)
  }
  if (valueType === 'string') {
    if (/^\d+$/.test(String(value))) {
      return String(value)
    }
    if (value === 'true' || value === 'false') {
      return value
    }
  }
  return `"${value}"`
}

function buildFunctionRQL(funcName: string, currentUserId?: number | null): string {
  const today = new Date().toISOString().split('T')[0]
  switch (funcName) {
    case 'isOverdue':
      return `(target_date < "${today}" AND state_group IN ("backlog", "in_progress"))`
    case 'hasNoAssignee':
      return 'assignee_id IS NULL'
    case 'hasNoLabel':
      return 'label IS NULL'
    case 'isTopLevel':
      return 'parent_id IS NULL'
    case 'isSubWorkItem':
      return 'parent_id IS NOT NULL'
    case 'hasChildren':
      return 'has_children = true'
    case 'hasStartAndDueDates':
      return 'start_date IS NOT NULL AND target_date IS NOT NULL'
    default:
      return ''
  }
}

function isFilterCondition(node: FilterNode): node is FilterCondition {
  return 'field' in node && 'operator' in node && 'value' in node
}

function buildNodeRQL(node: FilterNode, currentUserId?: number | null): string {
  if (isFilterCondition(node)) {
    const { field, operator, value: rawValue } = node

    if (Array.isArray(rawValue) && rawValue.length === 0) return ''
    if ((rawValue === '' || rawValue === null || rawValue === undefined) && 
        !['is empty', 'is not empty'].includes(operator)) return ''

    const resolveValue = (val: any): any => {
      if (val === '$CURRENT_USER' && currentUserId != null) {
        return currentUserId
      }
      return val
    }
    const value = Array.isArray(rawValue) ? rawValue.map(v => resolveValue(v)) : resolveValue(rawValue)

    const fieldDef = FILTER_FIELDS.find(f => f.key === field)
    const dbKey = fieldDef?.dbKey || field
    const valueType = fieldDef?.valueType || 'string'

    switch (operator) {
      case 'is':
        return `${dbKey} = ${formatValue(value, valueType)}`
      case 'is not':
        return `${dbKey} != ${formatValue(value, valueType)}`
      case 'is any of': {
        const anyOfValues = Array.isArray(value) ? value : [value]
        const formattedAnyOf = anyOfValues.map(v => formatValue(v, valueType)).join(', ')
        return `${dbKey} IN (${formattedAnyOf})`
      }
      case 'is not any of': {
        const notAnyOfValues = Array.isArray(value) ? value : [value]
        const formattedNotAnyOf = notAnyOfValues.map(v => formatValue(v, valueType)).join(', ')
        return `${dbKey} NOT IN (${formattedNotAnyOf})`
      }
      case 'contains': {
        const escaped = String(value).replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/%/g, '\\%').replace(/_/g, '\\_')
        return `${dbKey} LIKE "%${escaped}%"`
      }
      case 'does not contain': {
        const escaped = String(value).replace(/\\/g, '\\\\').replace(/"/g, '\\"').replace(/%/g, '\\%').replace(/_/g, '\\_')
        return `${dbKey} NOT LIKE "%${escaped}%"`
      }
      case 'is empty':
        return `${dbKey} IS NULL`
      case 'is not empty':
        return `${dbKey} IS NOT NULL`
      case 'before':
        return `${dbKey} < "${value}"`
      case 'after':
        return `${dbKey} > "${value}"`
      case 'before or on':
        return `${dbKey} <= "${value}"`
      case 'after or on':
        return `${dbKey} >= "${value}"`
      case 'between': {
        if (Array.isArray(value) && value.length >= 2 && value[0] !== '' && value[1] !== '') {
          return `${dbKey} >= ${formatValue(value[0], valueType)} AND ${dbKey} <= ${formatValue(value[1], valueType)}`
        }
        return ''
      }
      case 'not between': {
        if (Array.isArray(value) && value.length >= 2 && value[0] !== '' && value[1] !== '') {
          return `${dbKey} < ${formatValue(value[0], valueType)} OR ${dbKey} > ${formatValue(value[1], valueType)}`
        }
        return ''
      }
      default:
        return ''
    }
  } else {
    const childClauses = node.children
      .map(child => buildNodeRQL(child, currentUserId))
      .filter(clause => clause)
    
    if (childClauses.length === 0) return ''
    if (childClauses.length === 1) return childClauses[0]
    
    return `(${childClauses.join(` ${node.operator} `)})`
  }
}

export interface RQLResult {
  filters: FilterCondition[]
  filterGroups: FilterGroup[]
  sortBy?: SortOption[]
}

export function buildRQL(filters: FilterCondition[], quickSearchValue?: string, currentUserId?: number | null, sortBy?: SortOption[]): string
export function buildRQL(filterGroups: FilterGroup[], quickSearchValue?: string, currentUserId?: number | null, sortBy?: SortOption[]): string
export function buildRQL(filtersOrGroups: FilterCondition[] | FilterGroup[], quickSearchValue?: string, currentUserId?: number | null, sortBy?: SortOption[]): string {
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
        const escaped = qs
          .replace(/\\/g, '\\\\')
          .replace(/"/g, '\\"')
          .replace(/%/g, '\\%')
          .replace(/_/g, '\\_')
        clauses.push(`(name LIKE "%${escaped}%" OR description LIKE "%${escaped}%")`)
      }
    }
  }

  let hasFilterContent = false

  if ((filtersOrGroups as FilterGroup[]).some((f: any) => 'operator' in f)) {
    const groups = filtersOrGroups as FilterGroup[]
    for (const group of groups) {
      const rql = buildNodeRQL(group, currentUserId)
      if (rql) {
        clauses.push(rql)
        hasFilterContent = true
      }
    }
  } else {
    const filters = filtersOrGroups as FilterCondition[]
    for (const filter of filters) {
      const rql = buildNodeRQL(filter, currentUserId)
      if (rql) {
        clauses.push(rql)
        hasFilterContent = true
      }
    }
  }

  if (sortBy && sortBy.length > 0) {
    for (const sort of sortBy) {
      const fieldDef = FILTER_FIELDS.find(f => f.key === sort.key)
      const dbKey = fieldDef?.dbKey || sort.key
      clauses.push(`orderby ${dbKey} ${sort.direction}`)
    }
    hasFilterContent = true
  }

  if (!hasFilterContent && !quickSearchValue) return ''

  return clauses.join(' AND ')
}

function parseSimpleCondition(clause: string): FilterCondition | null {
  const fieldMatch = clause.match(/^(\w+)/)
  if (!fieldMatch) return null
  const field = fieldMatch[1]

  const likeMatch = clause.match(/^(\w+) (NOT LIKE|LIKE) "((?:[^"\\]|\\.)*)"$/)
  if (likeMatch) {
    let rawValue = (likeMatch[3] || '').replace(/\\(.)/g, '$1')
    const value = rawValue.replace(/^%/, '').replace(/%$/, '')
    return {
      field: likeMatch[1],
      operator: likeMatch[2] === 'LIKE' ? 'contains' : 'does not contain',
      value,
      displayValue: value
    }
  }

  const inMatch = clause.match(/^(\w+) (NOT IN|IN) \((.*)\)$/)
  if (inMatch) {
    const valuesStr = inMatch[3]
    const values = valuesStr.split(',').map(v => {
      const trimmedV = v.trim()
      const quotedMatch = trimmedV.match(/^"([^"]+)"$/) || trimmedV.match(/^'([^']+)'$/)
      return quotedMatch ? quotedMatch[1] : trimmedV
    })
    return {
      field: inMatch[1],
      operator: inMatch[2] === 'IN' ? 'is any of' : 'is not any of',
      value: values,
      displayValue: values.join(', ')
    }
  }

  const nullMatch = clause.match(/^(\w+) (IS NOT NULL|IS NULL)$/)
  if (nullMatch) {
    return {
      field: nullMatch[1],
      operator: nullMatch[2] === 'IS NULL' ? 'is empty' : 'is not empty',
      value: '',
      displayValue: ''
    }
  }

  const betweenMatch = clause.match(/^(\w+)\s+>=\s+"([^"]+)"\s+AND\s+\1\s+<=\s+"([^"]+)"$/i) || 
                       clause.match(/^(\w+)\s+>=\s+'([^']+)'\s+AND\s+\1\s+<=\s+'([^']+)'$/i)
  if (betweenMatch) {
    return {
      field: betweenMatch[1],
      operator: 'between',
      value: [betweenMatch[2], betweenMatch[3]],
      displayValue: `${betweenMatch[2]} - ${betweenMatch[3]}`
    }
  }

  const notBetweenMatch = clause.match(/^(\w+)\s+<\s+"([^"]+)"\s+OR\s+\1\s+>\s+"([^"]+)"$/i) ||
                         clause.match(/^(\w+)\s+<\s+'([^']+)'\s+OR\s+\1\s+>\s+'([^']+)'$/i)
  if (notBetweenMatch) {
    return {
      field: notBetweenMatch[1],
      operator: 'not between',
      value: [notBetweenMatch[2], notBetweenMatch[3]],
      displayValue: `not ${notBetweenMatch[2]} - ${notBetweenMatch[3]}`
    }
  }

  const quotedMatch = clause.match(/^(\w+) ([=<>!]+) "([^"]+)"$/) || clause.match(/^(\w+) ([=<>!]+) '([^']+)'$/)
  if (quotedMatch) {
    const op = quotedMatch[2]
    let operator = ''
    switch (op) {
      case '=':  operator = 'is'; break
      case '!=': operator = 'is not'; break
      case '>':  operator = 'after'; break
      case '<':  operator = 'before'; break
      case '>=': operator = 'after or on'; break
      case '<=': operator = 'before or on'; break
    }
    if (operator) {
      return {
        field: quotedMatch[1],
        operator,
        value: quotedMatch[3],
        displayValue: quotedMatch[3]
      }
    }
  }

  const unquotedMatch = clause.match(/^(\w+) ([=<>!]+) (\S+)$/)
  if (unquotedMatch) {
    const op = unquotedMatch[2]
    let operator = ''
    switch (op) {
      case '=':  operator = 'is'; break
      case '!=': operator = 'is not'; break
      case '>':  operator = 'after'; break
      case '<':  operator = 'before'; break
      case '>=': operator = 'after or on'; break
      case '<=': operator = 'before or on'; break
    }
    if (operator) {
      return {
        field: unquotedMatch[1],
        operator,
        value: unquotedMatch[3],
        displayValue: unquotedMatch[3]
      }
    }
  }

  return null
}

function parseRQLExpression(rqlStr: string): FilterNode | null {
  const trimmed = rqlStr.trim()
  
  if (trimmed.startsWith('(') && trimmed.endsWith(')')) {
    const inner = trimmed.slice(1, -1).trim()
    return parseRQLGroup(inner)
  }

  const betweenMatch = trimmed.match(/^(\w+)\s+>=\s+"([^"]+)"\s+AND\s+(\w+)\s+<=\s+"([^"]+)"$/i) || 
                       trimmed.match(/^(\w+)\s+>=\s+'([^']+)'\s+AND\s+(\w+)\s+<=\s+'([^']+)'$/i)
  if (betweenMatch && betweenMatch[1] === betweenMatch[3]) {
    return {
      field: betweenMatch[1],
      operator: 'between',
      value: [betweenMatch[2], betweenMatch[4]],
      displayValue: `${betweenMatch[2]} - ${betweenMatch[4]}`
    }
  }

  const notBetweenMatch = trimmed.match(/^(\w+)\s+<\s+"([^"]+)"\s+OR\s+(\w+)\s+>\s+"([^"]+)"$/i) ||
                         trimmed.match(/^(\w+)\s+<\s+'([^']+)'\s+OR\s+(\w+)\s+>\s+'([^']+)'$/i)
  if (notBetweenMatch && notBetweenMatch[1] === notBetweenMatch[3]) {
    return {
      field: notBetweenMatch[1],
      operator: 'not between',
      value: [notBetweenMatch[2], notBetweenMatch[4]],
      displayValue: `not ${notBetweenMatch[2]} - ${notBetweenMatch[4]}`
    }
  }

  const simple = parseSimpleCondition(trimmed)
  if (simple) {
    return simple
  }

  return parseRQLGroup(trimmed)
}

function parseRQLGroup(rqlStr: string): FilterGroup | null {
  const trimmed = rqlStr.trim()
  
  const betweenMatch = trimmed.match(/^(\w+)\s+>=\s+"([^"]+)"\s+AND\s+\1\s+<=\s+"([^"]+)"$/i) || 
                       trimmed.match(/^(\w+)\s+>=\s+'([^']+)'\s+AND\s+\1\s+<=\s+'([^']+)'$/i)
  if (betweenMatch) {
    return null
  }

  const notBetweenMatch = trimmed.match(/^(\w+)\s+<\s+"([^"]+)"\s+OR\s+\1\s+>\s+"([^"]+)"$/i) ||
                         trimmed.match(/^(\w+)\s+<\s+'([^']+)'\s+OR\s+\1\s+>\s+'([^']+)'$/i)
  if (notBetweenMatch) {
    return null
  }

  const parts: string[] = []
  let current = ''
  let depth = 0
  let inString = false
  let stringChar = ''
  let topLevelOperator: 'AND' | 'OR' | null = null

  for (let i = 0; i < rqlStr.length; i++) {
    const char = rqlStr[i]
    const nextChar = rqlStr[i + 1]

    if (!inString && char === '(') {
      depth++
      current += char
    } else if (!inString && char === ')') {
      depth--
      current += char
    } else if (!inString && (char === '"' || char === "'")) {
      inString = true
      stringChar = char
      current += char
    } else if (inString && char === stringChar && nextChar !== '\\') {
      inString = false
      current += char
    } else if (!inString && depth === 0) {
      const potentialAnd = rqlStr.slice(i, i + 4).toUpperCase()
      const potentialOr = rqlStr.slice(i, i + 3).toUpperCase()

      if (potentialAnd === ' AND') {
        if (topLevelOperator === null) topLevelOperator = 'AND'
        parts.push(current.trim())
        current = ''
        i += 3
      } else if (potentialOr === ' OR') {
        if (topLevelOperator === null) topLevelOperator = 'OR'
        parts.push(current.trim())
        current = ''
        i += 2
      } else {
        current += char
      }
    } else {
      current += char
    }
  }

  if (current.trim()) {
    parts.push(current.trim())
  }

  if (parts.length < 2) return null

  const operator: 'AND' | 'OR' = topLevelOperator || 'AND'

  const children: FilterNode[] = []
  for (const part of parts) {
    const node = parseRQLExpression(part)
    if (node) {
      children.push(node)
    }
  }

  if (children.length === 0) return null
  if (children.length === 1) return children[0]

  return { operator, children }
}

export function parseRQL(rqlStr: string): RQLResult {
  if (!rqlStr.trim()) return { filters: [], filterGroups: [] }

  const conditions: FilterCondition[] = []
  const filterGroups: FilterGroup[] = []
  const sortBy: SortOption[] = []

  let rqlWithoutOrderby = rqlStr
  let orderbyMatch = rqlWithoutOrderby.match(/orderby (\w+) (asc|desc)$/i)
  while (orderbyMatch) {
    sortBy.push({
      key: orderbyMatch[1],
      labelKey: SORT_OPTION_MAP[orderbyMatch[1]]?.labelKey || '',
      direction: orderbyMatch[2].toLowerCase() as 'asc' | 'desc'
    })
    const newRql = rqlWithoutOrderby.replace(/\s+(AND|OR)\s+orderby \w+ (asc|desc)$/i, '').trim()
    if (newRql !== rqlWithoutOrderby) {
      rqlWithoutOrderby = newRql
    } else {
      rqlWithoutOrderby = rqlWithoutOrderby.replace(/\s*orderby \w+ (asc|desc)$/i, '').trim()
    }
    orderbyMatch = rqlWithoutOrderby.match(/orderby (\w+) (asc|desc)$/i)
  }
  sortBy.reverse()

  const node = parseRQLExpression(rqlWithoutOrderby)
  if (!node) return { filters: [], filterGroups: [], sortBy: sortBy.length > 0 ? sortBy : undefined }

  if (isFilterCondition(node)) {
    conditions.push(node)
  } else {
    if (node.operator === 'AND' && node.children.every(c => isFilterCondition(c))) {
      for (const child of node.children) {
        if (isFilterCondition(child)) {
          conditions.push(child)
        }
      }
    } else {
      filterGroups.push(node)
    }
  }

  return { filters: conditions, filterGroups, sortBy: sortBy.length > 0 ? sortBy : undefined }
}

export function flattenFilterGroups(groups: FilterGroup[]): FilterCondition[] {
  const conditions: FilterCondition[] = []

  function traverse(node: FilterNode): void {
    if (isFilterCondition(node)) {
      conditions.push(node)
    } else {
      for (const child of node.children) {
        traverse(child)
      }
    }
  }

  for (const group of groups) {
    traverse(group)
  }

  return conditions
}

export const SORT_OPTIONS: SortOption[] = [
  { key: 'created_at', labelKey: 'filter.orderLastCreated', direction: 'desc' },
  { key: 'updated_at', labelKey: 'filter.orderLastUpdated', direction: 'desc' },
  { key: 'priority', labelKey: 'filter.orderPriority', direction: 'desc' },
  { key: 'start_date', labelKey: 'filter.orderStartDate', direction: 'asc' },
  { key: 'target_date', labelKey: 'filter.orderDueDate', direction: 'asc' },
  { key: 'sequence_id', labelKey: 'filter.orderSequenceId', direction: 'desc' },
  { key: 'name', labelKey: 'filter.orderName', direction: 'asc' },
  { key: 'state', labelKey: 'filter.orderState', direction: 'asc' },
  { key: 'issue_type', labelKey: 'filter.orderType', direction: 'asc' },
]

export const SORT_OPTION_MAP: Record<string, SortOption> = {}
SORT_OPTIONS.forEach(s => { SORT_OPTION_MAP[s.key] = s })

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

export type SubGroupOption = GroupOption

export const SUB_GROUP_OPTIONS: SubGroupOption[] = [
  { key: 'none', labelKey: 'filter.subGroupByNone' },
  { key: 'state_id', labelKey: 'filter.subGroupByState' },
  { key: 'priority', labelKey: 'filter.subGroupByPriority' },
  { key: 'assignee_id', labelKey: 'filter.subGroupByAssignee' },
  { key: 'label', labelKey: 'filter.subGroupByLabel' },
  { key: 'cycle_id', labelKey: 'filter.subGroupByCycle' },
  { key: 'module_id', labelKey: 'filter.subGroupByModule' },
  { key: 'type_id', labelKey: 'filter.subGroupByType' },
]