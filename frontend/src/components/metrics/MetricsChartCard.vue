<template>
  <div class="bg-white border border-gray-100 rounded-xl overflow-hidden hover:shadow-md transition-shadow">
    <!-- Title Bar -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-50">
      <div class="flex items-center gap-2 flex-1 mr-2 min-w-0">
        <h4 class="text-sm font-medium text-gray-700 truncate">{{ chart.name }}</h4>
        <span class="text-[10px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded shrink-0">{{ typeLabel }}</span>
      </div>
      <div v-if="chart.id > 0" class="flex items-center gap-1 shrink-0">
        <button @click="emit('edit', chart)" class="p-1 text-gray-400 hover:text-indigo-600 rounded transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>
        <button @click="emit('delete', chart.id)" class="p-1 text-gray-400 hover:text-red-500 rounded transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Axis Info -->
    <div v-if="chart.x_axis || chart.y_axis" class="card-axis-info px-4">
      <span v-if="chart.x_axis">X: {{ axisLabel(chart.x_axis) }}</span>
      <span v-if="chart.x_axis && chart.y_axis" class="axis-sep">|</span>
      <span v-if="chart.y_axis">Y: {{ axisLabel(chart.y_axis) }}</span>
      <template v-if="formatFilters(chart.filters)">
        <span class="axis-sep">|</span>
        <span>{{ t('metrics.filter') }}: {{ formatFilters(chart.filters) }}</span>
      </template>
    </div>

    <!-- Chart Area -->
    <div class="px-4 py-4">
      <div v-if="loading" class="flex items-center justify-center h-48">
        <svg class="animate-spin h-6 w-6 text-gray-300" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>
      <!-- Table View -->
      <div v-else-if="hasData && isTableType" class="overflow-x-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="border-b border-gray-100">
              <th class="text-left py-2 px-3 text-gray-500 font-medium">{{ axisLabel(chart.x_axis) }}</th>
              <th class="text-right py-2 px-3 text-gray-500 font-medium">{{ axisLabel(chart.y_axis) }}</th>
              <th class="text-right py-2 px-3 text-gray-500 font-medium">{{ t('metrics.tableHeader.ratio') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, i) in paginatedItems" :key="i" class="border-b border-gray-50 hover:bg-gray-50/50">
              <td class="py-1.5 px-3 text-gray-700">{{ item.label }}</td>
              <td class="py-1.5 px-3 text-right font-medium text-gray-800">{{ item.value.toLocaleString() }}</td>
              <td class="py-1.5 px-3 text-right text-gray-500">{{ item.pct }}%</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="border-t border-gray-200 font-medium">
              <td class="py-1.5 px-3 text-gray-600">{{ t('metrics.tableHeader.total') }}</td>
              <td class="py-1.5 px-3 text-right text-gray-800">{{ renderData!.total.toLocaleString() }}</td>
              <td class="py-1.5 px-3 text-right text-gray-500">100%</td>
            </tr>
          </tfoot>
        </table>
        <!-- Pagination -->
        <div v-if="tablePageCount > 1" class="flex items-center justify-between py-2 px-3 border-t border-gray-100">
          <span class="text-[11px] text-gray-400">{{ t('metrics.tableHeader.pagination', { total: tableItems.length, page: tablePage, totalPages: tablePageCount }) }}</span>
          <div class="flex items-center gap-1">
            <button @click="tablePage = Math.max(1, tablePage - 1)" :disabled="tablePage <= 1"
              class="px-2 py-0.5 text-[11px] rounded border border-gray-200 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed">{{ t('metrics.tableHeader.prevPage') }}</button>
            <button v-for="p in visiblePages" :key="p" @click="tablePage = p"
              :class="['px-2 py-0.5 text-[11px] rounded border', p === tablePage ? 'bg-indigo-50 border-indigo-200 text-indigo-600 font-medium' : 'border-gray-200 hover:bg-gray-50']">{{ p }}</button>
            <button @click="tablePage = Math.min(tablePageCount, tablePage + 1)" :disabled="tablePage >= tablePageCount"
              class="px-2 py-0.5 text-[11px] rounded border border-gray-200 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed">{{ t('metrics.tableHeader.nextPage') }}</button>
          </div>
        </div>
      </div>
      <!-- Canvas Chart -->
      <div v-else-if="hasData" :class="['mx-auto', isPieType ? 'max-w-md' : 'max-w-full']" style="height: 280px">
        <canvas ref="chartCanvas"></canvas>
      </div>
      <div v-else class="flex flex-col items-center justify-center h-48 text-xs text-gray-400">
        <svg class="w-8 h-8 mb-2 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <span>{{ t('metrics.tableHeader.noData') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, onUnmounted } from 'vue'
import type { MetricChart } from '@/types/metrics'
import type { RenderResult } from '@/types/metrics'
import type { ReportResponse } from '@/api/report'
import { metricsApi } from '@/api/metrics'
import { useReportChart } from '@/composables/useReportChart'

import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  chart: MetricChart
  projectId: number
}>()

const emit = defineEmits<{
  (e: 'edit', chart: MetricChart): void
  (e: 'delete', chartId: number): void
}>()

const chartCanvas = ref<HTMLCanvasElement | null>(null)
const { render: renderChart, destroy: destroyChart } = useReportChart(chartCanvas)

const loading = ref(false)
const fetchedData = ref<ReportResponse | null>(null)
const currentType = ref(props.chart.chart_type)

const typeLabel = computed(() => {
  const map: Record<string, string> = {
    bar: t('metrics.chartTypes.bar'), line: t('metrics.chartTypes.line'), pie: t('metrics.chartTypes.pie'), doughnut: t('metrics.chartTypes.doughnut'),
    area: t('metrics.chartTypes.area'), radar: t('metrics.chartTypes.radar'), scatter: t('metrics.chartTypes.scatter'), bubble: t('metrics.chartTypes.bubble'),
    mixed: t('metrics.chartTypes.mixed'), table: t('metrics.chartTypes.table'),
  }
  return map[currentType.value] || currentType.value
})

const chartTypeMap: Record<string, string> = {
  bar: 'Bar', line: 'Line', pie: 'Pie', doughnut: 'Doughnut',
  area: 'Area', radar: 'Radar', scatter: 'Scatter', bubble: 'Bubble',
  mixed: 'Mixed', table: 'Table',
}

const isPieType = computed(() => ['pie', 'doughnut'].includes(currentType.value))
const isTableType = computed(() => currentType.value === 'table')

// ── Table Pagination ──
const TABLE_PAGE_SIZE = 10
const tablePage = ref(1)

const tableItems = computed(() => {
  if (!renderData.value) return []
  return renderData.value.labels.map((label, i) => {
    const value = renderData.value!.values[i] || 0
    const pct = renderData.value!.total > 0 ? Math.round((value / renderData.value!.total) * 1000) / 10 : 0
    return { label, value, pct }
  })
})

const tablePageCount = computed(() => Math.max(1, Math.ceil(tableItems.value.length / TABLE_PAGE_SIZE)))

const paginatedItems = computed(() => {
  const start = (tablePage.value - 1) * TABLE_PAGE_SIZE
  return tableItems.value.slice(start, start + TABLE_PAGE_SIZE)
})

const visiblePages = computed(() => {
  const total = tablePageCount.value
  const cur = tablePage.value
  const pages: number[] = []
  const start = Math.max(1, cur - 2)
  const end = Math.min(total, start + 4)
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

const xAxisLabelMap: Record<string, string> = {
  state: t('metrics.dimensions.state'), type: t('metrics.dimensions.type'), priority: t('metrics.dimensions.priority'), assignee: t('metrics.dimensions.assignee'),
  reporter: t('metrics.dimensions.reporter'), title: t('metrics.fieldTitle'), label: t('metrics.dimensions.label'), module: t('metrics.dimensions.module'),
  created_at: t('metrics.dimensions.created_at'), updated_at: t('metrics.dimensions.updated_at'), created_by: t('metrics.fieldCreatedBy'),
  created_day: t('metrics.dimensions.createdByDay'), created_week: t('metrics.dimensions.createdByWeek'), created_month: t('metrics.dimensions.createdByMonth'),
  completed_day: t('metrics.dimensions.completedByDay'), completed_week: t('metrics.dimensions.completedByWeek'), completed_month: t('metrics.dimensions.completedByMonth'),
  updated_day: t('metrics.dimensions.updatedByDay'), updated_week: t('metrics.dimensions.updatedByWeek'), updated_month: t('metrics.dimensions.updatedByMonth'),
  state_group: t('metrics.dimensions.state_group'), cycle: t('metrics.dimensions.cycle'),
}

const yAxisLabelMap: Record<string, string> = {
  count: t('metrics.yAxisOptions.count'), avg_days: t('metrics.yAxisOptions.avgDays'), avg_processing_time: t('metrics.yAxisOptions.avg_processing_time'),
  current_retention: t('metrics.yAxisOptions.current_retention'), avg_cycle_time: t('metrics.yAxisOptions.avg_cycle_time'),
  completion_rate: t('metrics.yAxisOptions.completion_rate'), throughput: t('metrics.yAxisOptions.throughput'),
  wip_count: t('metrics.yAxisOptions.wip_count'), backlog_count: t('metrics.yAxisOptions.backlog_count'), overdue_count: t('metrics.yAxisOptions.overdue_count'),
  avg_resolution_days: t('metrics.yAxisOptions.avgResolutionDays'),
}

function axisLabel(field: string): string {
  if (field.startsWith('custom_field_avg:')) {
    return t('metrics.axisLabels.customFieldAvg')
  }
  if (field.startsWith('custom_field_count:')) {
    return t('metrics.axisLabels.customFieldCount')
  }
  if (field.startsWith('custom_field:')) {
    return t('metrics.axisLabels.customFieldPrefix', { id: field.split(':')[1] })
  }
  return xAxisLabelMap[field] || yAxisLabelMap[field] || field
}

function formatFilters(filters: string | Record<string, any> | null | undefined): string {
  if (!filters) return ''
  try {
    const obj = typeof filters === 'string' ? JSON.parse(filters) : filters
    // Prefer conditions array (human-readable)
    if (obj.conditions && Array.isArray(obj.conditions) && obj.conditions.length > 0) {
      const parts: string[] = []
      for (const c of obj.conditions) {
        if (!c || !c.field) continue
        const field = axisLabel(c.field)
        const op = c.operator || '='
        if (op === 'empty') { parts.push(`${field} ${t('metrics.opEmpty')}`); continue }
        if (op === 'not_empty') { parts.push(`${field} ${t('metrics.opNotEmpty')}`); continue }
        let val = ''
        if (c.values && c.values.length > 0) val = c.values.join(', ')
        else if (c.value) val = String(c.value)
        if (val) parts.push(`${field} ${op} ${val}`)
      }
      return parts.join(' AND ')
    }
    // Fallback to RQL string
    if (obj.rql && typeof obj.rql === 'string' && obj.rql.trim()) return obj.rql
    return ''
  } catch {
    return ''
  }
}

// Preview mode: chart.id === 0 means preview with inline data
const isPreview = computed(() => props.chart.id === 0)

// Build report response from either fetched data or preview inline data
const renderData = computed(() => {
  const data: ReportResponse | null = isPreview.value
    ? (props.chart.data_labels && props.chart.data_labels.length > 0
      ? { type: currentType.value, labels: props.chart.data_labels, values: props.chart.data_values || [], total: (props.chart.data_values || []).reduce((a, b) => a + b, 0) }
      : null)
    : fetchedData.value
  return data
})

// Reset page when data changes
watch(renderData, () => { tablePage.value = 1 })

const hasData = computed(() => renderData.value && renderData.value.labels.length > 0)

function toReportResponse(data: RenderResult): ReportResponse {
  return {
    type: data.chart_type || 'distribution',
    labels: data.labels,
    values: data.values,
    total: data.total,
    colors: data.colors,
  }
}

async function fetchAndRender() {
  if (isPreview.value) return // Preview mode: data comes from props
  loading.value = true
  try {
    const res: RenderResult = await metricsApi.renderChart(props.projectId, props.chart.id)
    fetchedData.value = toReportResponse(res)
  } catch (e) {
    console.error('Failed to render chart:', e)
    fetchedData.value = null
  } finally {
    loading.value = false
  }
}

// Render chart when data changes (skip for table type)
watch(renderData, async (data) => {
  if (data && data.labels.length > 0 && !isTableType.value) {
    await nextTick()
    await new Promise(r => setTimeout(r, isPreview.value ? 100 : 50))
    renderChart(data, chartTypeMap[currentType.value] || 'Bar')
  }
}, { immediate: true })

watch(() => props.chart.id, () => {
  destroyChart()
  fetchedData.value = null
  currentType.value = props.chart.chart_type
  fetchAndRender()
})

// Re-fetch when chart config changes (edit without ID change)
watch(() => [props.chart.chart_type, props.chart.x_axis, props.chart.y_axis, props.chart.filters, props.chart.config], () => {
  if (props.chart.id > 0) {
    destroyChart()
    currentType.value = props.chart.chart_type
    fetchAndRender()
  }
})

onMounted(() => {
  fetchAndRender()
})

onUnmounted(() => {
  destroyChart()
})
</script>

<style scoped>
.card-axis-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  font-size: 11px;
  color: #64748b;
}
.axis-sep {
  color: #cbd5e1;
}
</style>
