import { reactive, computed, provide, inject, type ComputedRef } from 'vue'
import type { FilterCondition, SortOption, GroupOption } from '../types/filters'
import { buildRQL, parseRQL } from '../types/filters'

const FILTERS_KEY = Symbol('filters')
const SEARCH_HISTORY_KEY = 'reqmango_search_history'
const MAX_HISTORY_COUNT = 10

export interface FiltersState {
  filters: FilterCondition[]
  sortBy: SortOption | null
  groupBy: GroupOption | null
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
  restoreFromRQL: (rql: string) => void
  setSortBy: (sort: SortOption | null) => void
  setGroupBy: (group: GroupOption | null) => void
  setQuickSearch: (query: string) => void
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
  const state = reactive<FiltersState>({
    filters: [],
    sortBy: null,
    groupBy: null,
    quickSearch: '',
    searchHistory: loadSearchHistory()
  })

  const rql = computed<string>(() => {
    return buildRQL(state.filters, state.quickSearch)
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
    state.sortBy = null
    state.groupBy = null
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

  function restoreFromRQL(rqlStr: string): void {
    if (!rqlStr.trim()) {
      state.filters = []
      state.sortBy = null
      return
    }
    const parsed = parseRQL(rqlStr)
    state.filters = parsed.filters
    state.sortBy = parsed.sortBy || null
  }

  function setSortBy(sort: SortOption | null): void {
    state.sortBy = sort
  }

  function setGroupBy(group: GroupOption | null): void {
    state.groupBy = group
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
    setGroupBy,
    setQuickSearch,
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