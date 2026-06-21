import { ref } from 'vue'
import { rqlApi } from '../api/rql'
import type { RQLSearchRequest, RQLHistoryItem } from '../utils/rql/types'

const HISTORY_KEY = 'reqman:rql:history'
const MAX_HISTORY = 50

export function useRQL() {
  const rql = ref('')
  const loading = ref(false)
  const error = ref<string | null>(null)
  const results = ref<any[]>([])
  const total = ref(0)

  // 历史记录
  const getHistory = (): RQLHistoryItem[] => {
    try {
      const data = localStorage.getItem(HISTORY_KEY)
      return data ? JSON.parse(data) : []
    } catch {
      return []
    }
  }

  const addToHistory = (rqlText: string, entityType: 'issue' | 'cycle' | 'module' = 'issue') => {
    if (!rqlText.trim()) return

    const history = getHistory()
    const filtered = history.filter(h => h.rql !== rqlText)

    filtered.unshift({
      id: Date.now().toString(),
      rql: rqlText,
      timestamp: Date.now(),
      entityType
    })

    const trimmed = filtered.slice(0, MAX_HISTORY)
    localStorage.setItem(HISTORY_KEY, JSON.stringify(trimmed))
  }

  const clearHistory = () => {
    localStorage.removeItem(HISTORY_KEY)
  }

  // 执行搜索
  const search = async (projectId: number, entity: 'issue' | 'cycle' | 'module' = 'issue', page = 1, pageSize = 20) => {
    if (!rql.value.trim()) {
      error.value = null
      results.value = []
      total.value = 0
      return
    }

    loading.value = true
    error.value = null

    try {
      const request: RQLSearchRequest = {
        entity,
        project_id: projectId,
        rql: rql.value,
        page,
        page_size: pageSize
      }

      const response = await rqlApi.search(request)

      if (response.data.success) {
        results.value = response.data.data?.items || []
        total.value = response.data.data?.total || 0
        addToHistory(rql.value, entity)
      } else {
        error.value = response.data.error?.message || '查询失败'
      }
    } catch (err: any) {
      error.value = err.response?.data?.error?.message || err.message || '网络错误'
    } finally {
      loading.value = false
    }
  }

  return {
    rql,
    loading,
    error,
    results,
    total,
    search,
    getHistory,
    addToHistory,
    clearHistory
  }
}
