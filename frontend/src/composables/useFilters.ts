/**
 * useFilters — Plane-style filter composable (Provide/Inject)
 *
 * Single source of truth for all filter state across views.
 * Features:
 *  - Semantic operator labels (is / is not / contains / is any of ...)
 *  - Instant updates (no confirm step)
 *  - RQL auto-derivation from FilterCondition[]
 *  - localStorage history
 *  - Serialization for URL / SavedView
 */
import { reactive, computed, inject, provide } from 'vue'
import type { FilterCondition, FilterHistoryItem } from '@/types/filters'
import { buildRQL } from '@/types/filters'

const FILTERS_KEY = Symbol('filters')

// ── History (localStorage) ──
const HISTORY_KEY_PREFIX = 'reqmango:filter_history:'
const MAX_HISTORY = 30

function loadHistory(projectId: number): FilterHistoryItem[] {
  try {
    const raw = localStorage.getItem(HISTORY_KEY_PREFIX + projectId)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function saveHistory(projectId: number, items: FilterHistoryItem[]) {
  try {
    const trimmed = items.slice(0, MAX_HISTORY)
    localStorage.setItem(HISTORY_KEY_PREFIX + projectId, JSON.stringify(trimmed))
  } catch { /* ignore quota errors */ }
}

// ── Composable ──

export function useFilters(projectId: number) {
  const state = reactive<{
    filters: FilterCondition[]
  }>({
    filters: [],
  })

  // ── Computed ──

  const rql = computed(() => buildRQL(state.filters))
  const activeFilterCount = computed(() => state.filters.length)
  const isEmpty = computed(() => activeFilterCount.value === 0)

  // ── CRUD (instant — no confirm step) ──

  function addFilter(condition: FilterCondition): void {
    state.filters.push(condition)
  }

  function removeFilter(index: number): void {
    state.filters.splice(index, 1)
  }

  function updateFilter(index: number, updates: Partial<FilterCondition>): void {
    Object.assign(state.filters[index], updates)
  }

  function clearAll(): void {
    state.filters = []
  }

  function setFilters(filters: FilterCondition[]): void {
    state.filters = [...filters]
  }

  // ── Serialization ──

  function toJSON(): FilterCondition[] {
    return JSON.parse(JSON.stringify(state.filters))
  }

  // ── History ──

  function pushHistory(): void {
    const history = loadHistory(projectId)
    history.unshift({
      id: Date.now().toString(36) + Math.random().toString(36).slice(2, 6),
      timestamp: Date.now(),
      filters: toJSON(),
      rql: rql.value,
      projectId,
    })
    saveHistory(projectId, history)
  }

  function getHistory(): FilterHistoryItem[] {
    return loadHistory(projectId)
  }

  function clearHistory(): void {
    saveHistory(projectId, [])
  }

  return {
    state,
    rql,
    activeFilterCount,
    isEmpty,
    addFilter,
    removeFilter,
    updateFilter,
    clearAll,
    setFilters,
    toJSON,
    pushHistory,
    getHistory,
    clearHistory,
  }
}

// ── Provide / Inject ──

export function provideFilters(instance: ReturnType<typeof useFilters>) {
  provide(FILTERS_KEY, instance)
}

export function injectFilters(): ReturnType<typeof useFilters> {
  const instance = inject<ReturnType<typeof useFilters>>(FILTERS_KEY)
  if (!instance) {
    throw new Error('injectFilters() must be called within a component tree that has provideFilters()')
  }
  return instance
}
