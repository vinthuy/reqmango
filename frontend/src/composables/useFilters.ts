import { reactive, computed, provide, inject, type ComputedRef } from 'vue'
import type { FilterCondition, SortOption, GroupOption, SubGroupOption } from '../types/filters'
import { buildRQL, parseRQL, FILTER_FIELDS } from '../types/filters'
import { useAuthStore } from '../stores/auth'

const FILTERS_KEY = Symbol('filters')
const SEARCH_HISTORY_KEY = 'reqmango_search_history'
const MAX_HISTORY_COUNT = 10

// DB key → UI key reverse mapping (for restoring from RQL)
const DB_TO_UI_KEY: Record<string, string> = {}
FILTER_FIELDS.forEach(f => { DB_TO_UI_KEY[f.dbKey] = f.key })

export interface FiltersState {
  filters: FilterCondition[]
  sortBy: SortOption[]  // multi-sort, empty array = default order
  groupBy: GroupOption | null
  subGroupBy: SubGroupOption | null
  quickSearch: string
  searchHistory: string[]
}

export interface FiltersContext {
  state: FiltersState
  rql: ComputedRef<string>
  activeFilterCount: ComputedRef<number>
  isEmpty: ComputedRef<boolean>
  addFilter: (field: string, operator: string, value: any, displayValue: string) => void
  removeFilter: (index: number) => void
  updateFilter: (index: number, updates: Partial<FilterCondition>) => void
  clearAll: () => void
  restoreFromRQL: (rql: string, extractQuickSearch?: boolean) => void
  setSortBy: (sorts: SortOption[]) => void
  addSortBy: (sort: SortOption) => void
  removeSortBy: (index: number) => void
  setGroupBy: (group: GroupOption | null) => void
  setQuickSearch: (query: string) => void
  setSubGroupBy: (group: SubGroupOption | null) => void
  addToHistory: (query: string) => void
  removeFromHistory: (index: number) => void
  clearHistory: () => void
}

function loadSearchHistory(): string[] {
  try {
    const stored = localStorage.getItem(SEARCH_HISTORY_KEY)
    if (stored) {
      return JSON.parse(stored)
    }
  } catch { /* */ }
  return []
}

function saveSearchHistory(history: string[]): void {
  try {
    localStorage.setItem(SEARCH_HISTORY_KEY, JSON.stringify(history))
  } catch { /* */ }
}

export function useFilters() {
  const auth = useAuthStore()
  
  const state = reactive<FiltersState>({
    filters: [],
    sortBy: [],
    groupBy: null,
    subGroupBy: null,
    quickSearch: '',
    searchHistory: loadSearchHistory()
  })

  const rql = computed<string>(() => {
    return buildRQL(state.filters, state.quickSearch, auth.user?.id, state.sortBy)
  })
  
  const activeFilterCount = computed<number>(() => state.filters.length + (state.quickSearch ? 1 : 0))
  const isEmpty = computed<boolean>(() => activeFilterCount.value === 0)

  function addFilter(field: string, operator: string, value: any, displayValue: string): void {
    state.filters.push({ field, operator, value, displayValue })
  }

  function removeFilter(index: number): void {
    state.filters.splice(index, 1)
  }

  function updateFilter(index: number, updates: Partial<FilterCondition>): void {
    const cond = state.filters[index]
    Object.assign(cond, updates)
  }

  function clearAll(): void {
    state.filters = []
    state.sortBy = []
    state.groupBy = null
    state.subGroupBy = null
    state.quickSearch = ''
  }

  function setQuickSearch(query: string): void {
    state.quickSearch = query
  }

  function addToHistory(query: string): void {
    const trimmed = query.trim()
    if (!trimmed) return
    
    const existingIndex = state.searchHistory.indexOf(trimmed)
    if (existingIndex !== -1) {
      state.searchHistory.splice(existingIndex, 1)
    }
    
    state.searchHistory.unshift(trimmed)
    if (state.searchHistory.length > MAX_HISTORY_COUNT) {
      state.searchHistory.pop()
    }
    
    saveSearchHistory(state.searchHistory)
  }

  function removeFromHistory(index: number): void {
    state.searchHistory.splice(index, 1)
    saveSearchHistory(state.searchHistory)
  }

  function clearHistory(): void {
    state.searchHistory = []
    saveSearchHistory(state.searchHistory)
  }

  function resolveTemplateVars(rql: string, currentUserId?: number | null): string {
    let result = rql
    // Resolve $CURRENT_USER
    if (currentUserId != null) {
      result = result.replace(/\$CURRENT_USER/g, String(currentUserId))
    }
    // Resolve date variables using local time (not UTC)
    const now = new Date()
    const pad = (n: number) => String(n).padStart(2, '0')
    const today = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
    const dayOfWeek = now.getDay()
    const daysUntilSunday = (7 - dayOfWeek) % 7
    const endOfWeek = new Date(now)
    endOfWeek.setDate(now.getDate() + daysUntilSunday)
    const endOfWeekStr = `${endOfWeek.getFullYear()}-${pad(endOfWeek.getMonth() + 1)}-${pad(endOfWeek.getDate())}`
    const oneWeekAgo = new Date(now)
    oneWeekAgo.setDate(now.getDate() - 7)
    const oneWeekAgoStr = `${oneWeekAgo.getFullYear()}-${pad(oneWeekAgo.getMonth() + 1)}-${pad(oneWeekAgo.getDate())}`

    // Wrap dates in double quotes so parseRQL's between detection works.
    // Between patterns (field >= "v1" AND field <= "v2") require quoted values
    // to be recognized and merged into a single 'between' filter condition.
    result = result.replace(/\$TODAY/g, `"${today}"`)
    result = result.replace(/\$END_OF_WEEK/g, `"${endOfWeekStr}"`)
    result = result.replace(/\$ONE_WEEK_AGO/g, `"${oneWeekAgoStr}"`)

    return result
  }

  function restoreFromRQL(rqlStr: string, extractQuickSearch?: boolean): void {
    if (!rqlStr.trim()) {
      state.filters = []
      state.sortBy = []
      if (extractQuickSearch) state.quickSearch = ''
      return
    }
    let cleaned = rqlStr
    // Resolve template variables before parsing
    cleaned = resolveTemplateVars(cleaned, auth.user?.id)
    if (extractQuickSearch) {
      const likeMatch = cleaned.match(/\(name\s+LIKE\s+"%(.+?)%"\s+OR\s+description\s+LIKE\s+"%(.+?)%"\)/i)
      state.quickSearch = likeMatch ? likeMatch[1] : ''
      cleaned = cleaned.replace(/\(name\s+LIKE\s+"%.+?%"\s+OR\s+description\s+LIKE\s+"%.+?%"\)/i, '')
      cleaned = cleaned.replace(/\s*AND\s*AND\s*/gi, ' AND ').replace(/^\s*AND\s*/i, '').replace(/\s*AND\s*$/i, '').trim()
    }
    if (cleaned) {
      const parsed = parseRQL(cleaned)
      // Map DB keys back to UI keys for FilterBar display
      state.filters = parsed.filters.map(f => ({
        ...f,
        field: DB_TO_UI_KEY[f.field] || f.field
      }))
      state.sortBy = parsed.sortBy || []
    } else {
      state.filters = []
      state.sortBy = []
    }
  }

  function setSortBy(sorts: SortOption[]): void {
    state.sortBy = sorts
  }

  function addSortBy(sort: SortOption): void {
    // Don't add duplicate sort fields
    const idx = state.sortBy.findIndex(s => s.key === sort.key)
    if (idx >= 0) {
      state.sortBy[idx] = sort
    } else {
      state.sortBy.push(sort)
    }
  }

  function removeSortBy(index: number): void {
    state.sortBy.splice(index, 1)
  }

  function setGroupBy(group: GroupOption | null): void {
    state.groupBy = group
  }

  function setSubGroupBy(group: SubGroupOption | null): void {
    state.subGroupBy = group
  }

  const context: FiltersContext = {
    state,
    rql,
    activeFilterCount,
    isEmpty,
    addFilter,
    removeFilter,
    updateFilter,
    clearAll,
    restoreFromRQL,
    setSortBy,
    addSortBy,
    removeSortBy,
    setGroupBy,
    setQuickSearch,
    setSubGroupBy,
    addToHistory,
    removeFromHistory,
    clearHistory
  }

  provide(FILTERS_KEY, context)

  return context
}

export function injectFilters(): FiltersContext {
  const context = inject<FiltersContext>(FILTERS_KEY)
  if (!context) {
    throw new Error('useFilters must be called within a provider')
  }
  return context
}