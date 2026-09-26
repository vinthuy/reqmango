<template>
  <div class="dashboard-page min-h-screen bg-gray-50 dark:bg-gray-900 flex flex-col">
    <!-- Top Bar -->
    <div class="topbar h-12 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center px-5 gap-3 sticky top-0 z-50">
      <button @click="goBack" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 flex items-center gap-1 text-sm shrink-0">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        {{ t('dashboard.backToProject') }}
      </button>
      <span class="text-gray-300 dark:text-gray-600">|</span>
      <span class="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">{{ t('dashboard.title') }}</span>
      <div class="flex-1"></div>
      <button v-if="editMode" class="btn btn-sm btn-ghost" @click="editMode = false; showConfig = false">
        {{ t('dashboard.done') }}
      </button>
      <button v-else class="btn btn-sm btn-primary" @click="editMode = true">
        {{ t('dashboard.edit') }}
      </button>
    </div>

    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar -->
      <DashboardSidebar
        :dashboards="dashboards"
        :currentId="currentId"
        @select="handleSelect"
        @create="handleCreate"
        @delete="handleDelete"
        @duplicate="handleDuplicate"
        @rename="handleRename"
      />

      <!-- Main Canvas -->
      <div class="flex-1 overflow-y-auto">
        <template v-if="current">
          <!-- Header -->
          <div class="border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 px-6 py-4 flex items-center gap-3">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 flex-1">{{ current.name }}</h2>
            <span class="text-xs px-2 py-1 rounded font-medium"
              :class="editMode ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/50 dark:text-amber-400' : 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-400'">
              {{ editMode ? t('dashboard.editing') : t('dashboard.viewing') }}
            </span>
          </div>

          <!-- Status bar -->
          <div class="px-6 py-2 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400 flex gap-4">
            <span>{{ t('dashboard.widgets') }}: <strong class="text-gray-700 dark:text-gray-300">{{ current.widgets?.length ?? 0 }}</strong></span>
          </div>

          <!-- Grid -->
          <DashboardGrid
            ref="gridRef"
            :widgets="current.widgets ?? []"
            :widgetData="widgetData"
            :editMode="editMode"
            :loading="loading"
            @add="showAddWidget = true"
            @configure="openConfig"
            @delete-widget="handleDeleteWidget"
            @refresh="loadWidgetData"
          />
        </template>

        <!-- Empty state -->
        <div v-else class="flex items-center justify-center h-full">
          <div class="text-center p-10">
            <svg class="w-16 h-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            <h3 class="text-base font-medium text-gray-500 dark:text-gray-400 mb-2">{{ t('dashboard.emptyTitle') }}</h3>
            <p class="text-sm text-gray-400 dark:text-gray-500 mb-4 max-w-xs mx-auto">{{ t('dashboard.emptyDesc') }}</p>
            <button class="btn btn-primary btn-sm" @click="handleCreate">
              <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              {{ t('dashboard.createFirst') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Widget Config Panel -->
    <WidgetConfigPanel
      v-if="configWidget"
      :widget="configWidget"
      :projectId="projectId"
      @close="closeConfig"
      @save="handleSaveWidget"
    />

    <!-- Add Widget Dialog -->
    <div v-if="showAddWidget" class="fixed inset-0 z-100 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/20" @click="showAddWidget = false"></div>
      <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-lg p-6 z-10 max-h-[85vh] overflow-y-auto">
        <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">{{ t('dashboard.addWidget') }}</h3>

        <!-- From Metrics -->
        <p class="text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-2">{{ t('dashboard.addFromMetrics') }}</p>
        <div v-if="metricChartsLoading" class="text-xs text-gray-400 py-4 text-center">…</div>
        <div v-else-if="metricCharts.length === 0" class="mb-4 p-3 rounded-lg bg-gray-50 dark:bg-gray-700/40 text-center">
          <p class="text-xs text-gray-500 mb-2">{{ t('dashboard.noMetricsCharts') }}</p>
          <button class="btn btn-sm btn-primary" @click="goToMetrics">{{ t('dashboard.goToMetrics') }}</button>
        </div>
        <div v-else class="grid grid-cols-1 gap-2 mb-5">
          <button
            v-for="c in metricCharts"
            :key="c.id"
            @click="handleAddMetricChart(c)"
            class="flex items-center gap-3 px-3 py-2.5 rounded-lg border border-gray-200 dark:border-gray-600 hover:border-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 text-left transition-colors"
          >
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 shrink-0">{{ c.chart_type }}</span>
            <span class="text-sm text-gray-800 dark:text-gray-200 truncate flex-1">{{ c.name }}</span>
          </button>
        </div>

        <!-- Layout widgets -->
        <p class="text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-2">{{ t('dashboard.addLayoutWidgets') }}</p>
        <div class="grid grid-cols-2 gap-2 mb-4">
          <button v-for="wt in layoutWidgetTypes" :key="wt.type"
            @click="handleAddWidget(wt.type)"
            class="flex flex-col items-center gap-1.5 p-3 rounded-lg border border-gray-200 dark:border-gray-600 hover:border-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 transition-colors">
            <span class="text-xs font-medium text-gray-700 dark:text-gray-300">{{ wt.label }}</span>
          </button>
        </div>

        <!-- Legacy advanced -->
        <details class="mb-4">
          <summary class="text-[11px] font-medium text-gray-400 cursor-pointer select-none">{{ t('dashboard.addAdvancedLegacy') }}</summary>
          <div class="grid grid-cols-2 gap-2 mt-2">
            <button v-for="wt in legacyWidgetTypes" :key="wt.type"
              @click="handleAddWidget(wt.type)"
              class="flex flex-col items-center gap-1 p-2.5 rounded-lg border border-dashed border-gray-200 dark:border-gray-600 hover:border-amber-400 text-xs text-gray-600 dark:text-gray-400">
              {{ wt.label }}
            </button>
          </div>
        </details>

        <div class="flex justify-end">
          <button class="btn btn-sm btn-ghost" @click="showAddWidget = false">{{ t('common.cancel') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useDashboard } from '@/composables/useDashboard'
import { metricsApi } from '@/api/metrics'
import type { MetricChart } from '@/types/metrics'
import type { DashboardWidget, WidgetCreate, WidgetType } from '@/types/dashboard'
import DashboardSidebar from '@/components/DashboardSidebar.vue'
import DashboardGrid from '@/components/DashboardGrid.vue'
import WidgetConfigPanel from '@/components/WidgetConfigPanel.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const projectId = computed(() => Number(route.params.id))
const {
  dashboards, current, currentId, editMode, loading, widgetData,
  loadDashboards, selectDashboard, createDashboard,
  updateDashboardMeta, deleteCurrentDashboard, duplicateCurrentDashboard,
  addWidgetToCurrent, updateWidgetOnCurrent, deleteWidgetOnCurrent,
  loadWidgetData,
} = useDashboard(projectId.value)

const gridRef = ref()
const showConfig = ref(false)
const showAddWidget = ref(false)
const configWidget = ref<DashboardWidget | null>(null)

const metricCharts = ref<MetricChart[]>([])
const metricChartsLoading = ref(false)

const layoutWidgetTypes = computed(() => [
  { type: 'number_card' as WidgetType, label: t('dashboard.numberCard') },
  { type: 'burndown' as WidgetType, label: t('dashboard.burndown') },
  { type: 'recent_list' as WidgetType, label: t('dashboard.recentList') },
  { type: 'ai_summary' as WidgetType, label: t('dashboard.aiSummary') },
])

const legacyWidgetTypes = computed(() => [
  { type: 'bar_chart' as WidgetType, label: t('dashboard.barChart') },
  { type: 'pie_chart' as WidgetType, label: t('dashboard.pieChart') },
  { type: 'doughnut_chart' as WidgetType, label: t('dashboard.doughnutChart') },
  { type: 'line_chart' as WidgetType, label: t('dashboard.lineChart') },
  { type: 'table' as WidgetType, label: t('dashboard.table') },
  { type: 'saved_report' as WidgetType, label: t('dashboard.savedReport') },
])

async function loadMetricCharts() {
  metricChartsLoading.value = true
  try {
    const data = await metricsApi.listCharts(projectId.value)
    metricCharts.value = Array.isArray(data) ? data : (data.charts || [])
  } catch {
    metricCharts.value = []
  } finally {
    metricChartsLoading.value = false
  }
}

watch(showAddWidget, (v) => {
  if (v) loadMetricCharts()
})

onMounted(() => {
  loadDashboards()
})

function goBack() {
  const slug = route.params.slug
  router.push({ name: 'Project', params: { slug, id: projectId.value } })
}

function goToMetrics() {
  showAddWidget.value = false
  router.push({ path: `/workspace/${route.params.slug}/project/${projectId.value}`, query: { tab: 'metrics' } })
}

function handleSelect(id: number) {
  selectDashboard(id)
}

async function handleCreate() {
  const base = t('dashboard.defaultName')
  const existing = new Set(dashboards.value.map((d) => d.name))
  let name = base
  let n = 2
  while (existing.has(name)) {
    name = `${base} (${n})`
    n += 1
  }
  await createDashboard(name)
}

async function handleDelete(id: number) {
  if (currentId.value === id) {
    await deleteCurrentDashboard()
  }
}

async function handleDuplicate(id: number) {
  if (currentId.value === id) {
    await duplicateCurrentDashboard()
  }
}

async function handleRename(id: number, name: string) {
  await updateDashboardMeta(id, { name })
}

function openConfig(widget: DashboardWidget | null) {
  configWidget.value = widget
}

function closeConfig() {
  configWidget.value = null
}

async function handleSaveWidget(updates: any) {
  if (!configWidget.value) return
  await updateWidgetOnCurrent(configWidget.value.id, updates)
  closeConfig()
}

async function handleDeleteWidget(widgetId: number) {
  await deleteWidgetOnCurrent(widgetId)
}

async function handleAddMetricChart(chart: MetricChart) {
  const data: WidgetCreate = {
    widget_type: 'metric_chart',
    title: chart.name,
    config: { metric_chart_id: chart.id },
    position: { x: 0, y: 0, w: 6, h: 4 },
  }
  await addWidgetToCurrent(data)
  showAddWidget.value = false
}

async function handleAddWidget(widgetType: WidgetType) {
  const titleMap: Partial<Record<WidgetType, string>> = {
    number_card: t('dashboard.numberCard'),
    bar_chart: t('dashboard.barChart'),
    pie_chart: t('dashboard.pieChart'),
    doughnut_chart: t('dashboard.doughnutChart'),
    line_chart: t('dashboard.lineChart'),
    bubble_chart: t('dashboard.bubbleChart'),
    scatter_chart: t('dashboard.scatterChart'),
    mixed_chart: t('dashboard.mixedChart'),
    metric_chart: t('dashboard.metricChart'),
    burndown: t('dashboard.burndown'),
    table: t('dashboard.table'),
    recent_list: t('dashboard.recentList'),
    saved_report: t('dashboard.savedReport'),
    ai_summary: t('dashboard.aiSummary'),
  }
  let config: Record<string, unknown> = {}
  if (widgetType === 'number_card') {
    config = { metric: 'total', label: t('dashboard.totalIssues') }
  } else if (widgetType === 'burndown') {
    config = { cycle_id: null }
  } else if (widgetType === 'recent_list') {
    config = { limit: 10 }
  } else if (widgetType === 'saved_report') {
    config = { saved_report_id: null }
  } else if (widgetType === 'ai_summary') {
    config = {}
  } else if (['line_chart', 'bubble_chart', 'scatter_chart', 'mixed_chart'].includes(widgetType)) {
    config = { report_type: 'created_trend', group_by: 'state', interval: 'week', rql: '' }
  } else if (['bar_chart', 'pie_chart', 'doughnut_chart', 'table'].includes(widgetType)) {
    config = { report_type: 'distribution', group_by: 'state', interval: 'week', rql: '' }
  }
  const data: WidgetCreate = {
    widget_type: widgetType,
    title: titleMap[widgetType] || widgetType,
    config,
    position: { x: 0, y: 0, w: widgetType === 'ai_summary' ? 6 : 4, h: widgetType === 'ai_summary' ? 4 : 3 },
  }
  await addWidgetToCurrent(data)
  showAddWidget.value = false
}
</script>

<style scoped>
.btn {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.15s;
}
.btn-ghost {
  background: transparent;
  color: #6b7280;
  border: 1px solid #d1d5db;
}
.btn-ghost:hover {
  background: #f3f4f6;
  color: #374151;
}
.btn-primary {
  background: #111827;
  color: #fff;
}
.btn-primary:hover {
  background: #1f2937;
}
.btn-sm {
  padding: 4px 10px;
  font-size: 12px;
}
</style>
