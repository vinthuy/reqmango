/**
 * useIssueFilters — Shared filter state across all issue views
 */
import { ref, computed, reactive } from 'vue'

export interface IssueFilter {
  key: string
  label: string
  value: string
  displayValue: string
}

export function useIssueFilters() {
  const search = ref('')
  const filters = reactive<Record<string, any>>({})
  const activeFilterObj = ref<IssueFilter[]>([])
  const viewMode = ref<'list' | 'kanban' | 'tree' | 'calendar' | 'gantt'>('list')

  function buildFilterChips(states: any[], cycles: any[], labels: any[], members: any[]): IssueFilter[] {
    const chips: IssueFilter[] = []
    if (filters.state_id > 0) {
      const s = states.find((x: any) => x.id === filters.state_id)
      chips.push({ key: 'state_id', label: '状态', value: String(filters.state_id), displayValue: s?.name || String(filters.state_id) })
    }
    if (filters.priority) {
      const map: Record<string, string> = { urgent: '紧急', high: '高', medium: '中', low: '低', none: '无' }
      chips.push({ key: 'priority', label: '优先级', value: filters.priority, displayValue: map[filters.priority] || filters.priority })
    }
    if (filters.cycle_id > 0) {
      const c = cycles.find((x: any) => x.id === filters.cycle_id)
      chips.push({ key: 'cycle_id', label: '周期', value: String(filters.cycle_id), displayValue: c?.name || String(filters.cycle_id) })
    }
    if (filters.assignee_id > 0) {
      const m = members.find((x: any) => x.user_id === filters.assignee_id || x.id === filters.assignee_id)
      chips.push({ key: 'assignee_id', label: '负责人', value: String(filters.assignee_id), displayValue: m?.display_name || m?.user?.display_name || String(filters.assignee_id) })
    }
    if (filters.label_ids?.length) {
      const names = filters.label_ids.map((id: number) => labels.find((l: any) => l.id === id)?.name || String(id))
      chips.push({ key: 'label_ids', label: '标签', value: filters.label_ids.join(','), displayValue: names.join(', ') })
    }
    if (filters.assignee_id === 'me') {
      chips.push({ key: 'assignee_id', label: '负责人', value: 'me', displayValue: '我的' })
    }
    return chips
  }

  function setFilter(key: string, value: any) {
    if (value === '' || value === 0 || value === null || value === undefined || (Array.isArray(value) && value.length === 0)) {
      delete filters[key]
    } else {
      filters[key] = value
    }
  }

  function removeFilter(key: string) {
    delete filters[key]
  }

  function clearAllFilters() {
    Object.keys(filters).forEach(k => delete filters[k])
    search.value = ''
  }

  function getQueryParams(): Record<string, any> {
    const params: Record<string, any> = {}
    if (filters.state_id > 0) params.state_id = filters.state_id
    if (filters.priority) params.priority = filters.priority
    if (filters.cycle_id > 0) params.cycle_id = filters.cycle_id
    if (filters.assignee_id > 0) params.assignee_id = filters.assignee_id
    if (filters.start_date) params.start_date = filters.start_date
    if (filters.end_date) params.target_date = filters.end_date
    if (search.value) params.search = search.value
    if (filters.label_ids?.length) params.label_ids = filters.label_ids.join(',')
    return params
  }

  const activeFilterCount = computed(() => Object.keys(filters).length + (search.value ? 1 : 0))

  return {
    search, filters, activeFilterObj, viewMode, activeFilterCount,
    setFilter, removeFilter, clearAllFilters, buildFilterChips, getQueryParams,
  }
}
