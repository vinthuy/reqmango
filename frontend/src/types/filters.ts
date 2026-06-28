/**
 * Filter Types — Plane-aligned filter condition definition
 *
 * Maps Plane semantic operators to RQL expressions internally.
 * Visual labels use Plane-style "is" / "contains" / "is any of" etc.
 */
export interface FilterCondition {
  /** 后端字段名 (state_id, priority, assignee_id, etc.) */
  field: string
  /** 语义操作符: is / is not / is any of / is not any of / contains / does not contain / is empty / is not empty / before / after / before or on / after or on / between / not between */
  operator: string
  /** 原始值 (string | number | string[] 多选) */
  value: any
  /** 显示值 (中文/英文标签) */
  displayValue: string
}

export interface FilterField {
  /** 后端字段名 */
  key: string
  /** i18n key (filter.field*) */
  labelKey: string
  /** 字段类型决定值选择器样式 */
  type: 'select' | 'multi' | 'date' | 'text' | 'date_range'
  /** 可用操作符 */
  operators: string[]
}

export interface FilterOperator {
  value: string
  labelKey: string
  /** 是否需要输入值 */
  needsValue: boolean
}

export interface FilterHistoryItem {
  id: string
  timestamp: number
  filters: FilterCondition[]
  rql: string
  projectId: number
}

/**
 * Complete Plane-aligned field definitions.
 * Ordered by relevance. Maps to Plane's filter dropdown.
 */
export const FILTER_FIELDS: FilterField[] = [
  { key: 'title',        labelKey: 'filter.fieldTitle',       type: 'text',       operators: ['is', 'is not', 'contains', 'does not contain'] },
  { key: 'state_id',     labelKey: 'filter.fieldState',       type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'state_group',  labelKey: 'filter.fieldStateGroup',  type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'priority',     labelKey: 'filter.fieldPriority',    type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'assignee_id',  labelKey: 'filter.fieldAssignee',    type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'label',        labelKey: 'filter.fieldLabel',       type: 'multi',      operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'cycle_id',     labelKey: 'filter.fieldCycle',       type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'module_id',    labelKey: 'filter.fieldModule',      type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of', 'is empty'] },
  { key: 'type_id',      labelKey: 'filter.fieldType',        type: 'select',     operators: ['is', 'is any of', 'is not', 'is not any of'] },
  { key: 'start_date',   labelKey: 'filter.fieldStartDate',   type: 'date',       operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty'] },
  { key: 'target_date',  labelKey: 'filter.fieldTargetDate',  type: 'date',       operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty'] },
  { key: 'created_at',   labelKey: 'filter.fieldCreatedAt',   type: 'date',       operators: ['is', 'is not', 'before', 'after', 'before or on', 'after or on', 'between', 'not between', 'is empty'] },
]

/** No-value operators (don't need a value input) */
export const NO_VALUE_OPERATORS = ['is empty', 'is not empty']

/** Multi-value operators (value is string[]) */
export const MULTI_VALUE_OPERATORS = ['is any of', 'is not any of']

/** Date range operator (needs two date inputs) */
export const DATE_RANGE_OPERATORS = ['between', 'not between']

/** Operator → RQL symbol mapping */
export const OPERATOR_TO_RQL: Record<string, string> = {
  'is': '=',
  'is not': '!=',
  'is any of': 'IN',
  'is not any of': 'NOT IN',
  'contains': '~',
  'does not contain': '!~',
  'is empty': 'IS EMPTY',
  'is not empty': 'IS NOT EMPTY',
  'before': '<',
  'after': '>',
  'before or on': '<=',
  'after or on': '>=',
  'between': '>=',
  'not between': '<',
}

/** State Group values */
export const STATE_GROUPS = {
  backlog: 'settings.stateGroupBacklog',
  unstarted: 'settings.stateGroupUnstarted',
  started: 'settings.stateGroupStarted',
  completed: 'settings.stateGroupCompleted',
  cancelled: 'settings.stateGroupCancelled',
} as const

/**
 * Build an RQL string from a list of FilterConditions.
 * Each condition maps to its RQL equivalent and combined with AND.
 */
export function buildRQL(filters: FilterCondition[]): string {
  return filters
    .map((c) => conditionToRQL(c))
    .filter(Boolean)
    .join(' AND ')
}

function conditionToRQL(c: FilterCondition): string | null {
  const { field, operator, value } = c
  if (NO_VALUE_OPERATORS.includes(operator)) {
    return `${field} ${OPERATOR_TO_RQL[operator] || ''}`
  }
  if (!value && value !== 0 && value !== false) return null

  if (MULTI_VALUE_OPERATORS.includes(operator)) {
    const vals = Array.isArray(value) ? value : [value]
    const formatted = vals.map((v: any) => typeof v === 'string' ? `"${v}"` : v).join(', ')
    return `${field} ${OPERATOR_TO_RQL[operator] || ''} [${formatted}]`
  }

  if (DATE_RANGE_OPERATORS.includes(operator)) {
    const [start, end] = Array.isArray(value) ? value : [value]
    if (operator === 'between') {
      return `${field} >= "${start}" AND ${field} <= "${end}"`
    }
    return `${field} < "${start}" OR ${field} > "${end}"`
  }

  const formatted = typeof value === 'string' ? `"${value}"` : value
  return `${field} ${OPERATOR_TO_RQL[operator] || '='} ${formatted}`
}
