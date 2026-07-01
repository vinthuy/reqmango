import { reactive, computed, provide, inject, type ComputedRef } from 'vue'
import type { FilterCondition, SortOption, GroupOption } from '../types/filters'
import { buildRQL, parseRQL } from '../types/filters'

const FILTERS_KEY = Symbol('filters')

export interface FiltersState {
  filters: FilterCondition[]
  sortBy: SortOption | null
  groupBy: GroupOption | null
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
}

export function useFilters() {
  const state = reactive<FiltersState>({
    filters: [],
    sortBy: null,
    groupBy: null
  })

  const rql = computed<string>(() => {
    return buildRQL(state.filters)
  })
  
  const activeFilterCount = computed<number>(() => state.filters.length)
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
    setGroupBy
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