/**
 * useDashboard - 仪表盘状态管理 composable
 *
 * 管理 Dashboard 列表、当前 dashboard、widget 数据、编辑/查看模式切换
 */
import { ref, computed } from 'vue'
import type { Dashboard, DashboardFullResponse, WidgetDataResponse } from '@/types/dashboard'
import * as dashboardApi from '@/api/dashboard'

export function useDashboard(projectId: number) {
  const dashboards = ref<Dashboard[]>([])
  const currentId = ref<number | null>(null)
  const editMode = ref(false)
  const loading = ref(false)
  const saving = ref(false)
  const widgetData = ref<WidgetDataResponse[]>([])

  const current = computed<Dashboard | null>(() =>
    dashboards.value.find((d) => d.id === currentId.value) ?? null
  )

  const hasDashboards = computed(() => dashboards.value.length > 0)

  // ==================== Dashboard CRUD ====================

  async function loadDashboards() {
    loading.value = true
    try {
      dashboards.value = await dashboardApi.listDashboards(projectId)
      // Select first or default
      if (!currentId.value) {
        const def = dashboards.value.find((d) => d.is_default)
        if (def) currentId.value = def.id
        else if (dashboards.value.length > 0) currentId.value = dashboards.value[0].id
      }
    } catch (e) {
      console.error('Failed to load dashboards', e)
    } finally {
      loading.value = false
    }
  }

  async function selectDashboard(id: number) {
    currentId.value = id
    await loadWidgetData()
  }

  async function createDashboard(name: string) {
    saving.value = true
    try {
      const d = await dashboardApi.createDashboard(projectId, { name, columns: 12 })
      dashboards.value.push(d)
      currentId.value = d.id
      widgetData.value = []
      return d
    } finally {
      saving.value = false
    }
  }

  async function updateDashboardMeta(id: number, updates: Record<string, any>) {
    saving.value = true
    try {
      const d = await dashboardApi.updateDashboard(projectId, id, updates)
      const idx = dashboards.value.findIndex((x) => x.id === id)
      if (idx >= 0) dashboards.value[idx] = d
      return d
    } finally {
      saving.value = false
    }
  }

  async function deleteCurrentDashboard() {
    if (!currentId.value) return
    saving.value = true
    try {
      await dashboardApi.deleteDashboard(projectId, currentId.value)
      dashboards.value = dashboards.value.filter((d) => d.id !== currentId.value)
      currentId.value = dashboards.value.length > 0 ? dashboards.value[0].id : null
      widgetData.value = []
    } finally {
      saving.value = false
    }
  }

  async function duplicateCurrentDashboard() {
    if (!currentId.value) return
    saving.value = true
    try {
      const d = await dashboardApi.duplicateDashboard(projectId, currentId.value)
      dashboards.value.push(d)
      currentId.value = d.id
      await loadWidgetData()
      return d
    } finally {
      saving.value = false
    }
  }

  // ==================== Widget Data ====================

  async function loadWidgetData() {
    if (!currentId.value) return
    loading.value = true
    try {
      const full: DashboardFullResponse = await dashboardApi.getDashboardFull(projectId, currentId.value)
      // Update local dashboard with fresh widgets
      const idx = dashboards.value.findIndex((d) => d.id === currentId.value)
      if (idx >= 0) dashboards.value[idx] = full.dashboard
      widgetData.value = full.widget_data
    } catch (e) {
      console.error('Failed to load widget data', e)
    } finally {
      loading.value = false
    }
  }

  function getWidgetData(widgetId: number): Record<string, any> | null {
    const entry = widgetData.value.find((w) => w.widget_id === widgetId)
    return entry?.data ?? null
  }

  // ==================== Widget CRUD ====================

  async function addWidgetToCurrent(data: any) {
    if (!currentId.value) return
    saving.value = true
    try {
      const w = await dashboardApi.addWidget(projectId, currentId.value, data)
      if (current.value) {
        current.value.widgets = [...(current.value.widgets ?? []), w]
      }
      // Reload data for the new widget
      await loadWidgetData()
      return w
    } finally {
      saving.value = false
    }
  }

  async function updateWidgetOnCurrent(widgetId: number, updates: any) {
    if (!currentId.value) return
    saving.value = true
    try {
      const w = await dashboardApi.updateWidget(projectId, currentId.value, widgetId, updates)
      if (current.value) {
        const idx = current.value.widgets.findIndex((x) => x.id === widgetId)
        if (idx >= 0) current.value.widgets[idx] = w
      }
      await loadWidgetData()
      return w
    } finally {
      saving.value = false
    }
  }

  async function deleteWidgetOnCurrent(widgetId: number) {
    if (!currentId.value) return
    saving.value = true
    try {
      await dashboardApi.deleteWidget(projectId, currentId.value, widgetId)
      if (current.value) {
        current.value.widgets = current.value.widgets.filter((w) => w.id !== widgetId)
      }
      widgetData.value = widgetData.value.filter((wd) => wd.widget_id !== widgetId)
    } finally {
      saving.value = false
    }
  }

  async function reorderWidgetsOnCurrent(widgetIds: number[]) {
    if (!currentId.value) return
    try {
      await dashboardApi.reorderWidgets(projectId, currentId.value, { widget_ids: widgetIds })
    } catch (e) {
      console.error('Failed to reorder widgets', e)
    }
  }

  // ==================== Editable flag ====================
  const canEdit = computed(() => {
    if (!current.value) return false
    return true // owner-only logic if needed
  })

  return {
    dashboards,
    current,
    currentId,
    editMode,
    loading,
    saving,
    widgetData,
    hasDashboards,
    canEdit,
    loadDashboards,
    selectDashboard,
    createDashboard,
    updateDashboardMeta,
    deleteCurrentDashboard,
    duplicateCurrentDashboard,
    loadWidgetData,
    getWidgetData,
    addWidgetToCurrent,
    updateWidgetOnCurrent,
    deleteWidgetOnCurrent,
    reorderWidgetsOnCurrent,
  }
}
