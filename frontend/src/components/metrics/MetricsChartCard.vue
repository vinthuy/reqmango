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
      <template v-if="chart.filters">
        <span class="axis-sep">|</span>
        <span>筛选: {{ chart.filters }}</span>
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
      <div v-else-if="hasData" :class="['mx-auto', isPieType ? 'max-w-md' : 'max-w-full']" style="height: 280px">
        <canvas ref="chartCanvas"></canvas>
      </div>
      <div v-else class="flex flex-col items-center justify-center h-48 text-xs text-gray-400">
        <svg class="w-8 h-8 mb-2 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <span>暂无数据</span>
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
  const map: Record<string, string> = { bar: '柱状图', line: '折线图', pie: '饼图', doughnut: '环形图', area: '面积图', radar: '雷达图', scatter: '散点图', bubble: '气泡图', mixed: '混合图', table: '表格' }
  return map[currentType.value] || currentType.value
})

const chartTypeMap: Record<string, string> = {
  bar: 'Bar', line: 'Line', pie: 'Pie', doughnut: 'Doughnut',
  area: 'Area', radar: 'Radar', scatter: 'Scatter', bubble: 'Bubble',
  mixed: 'Mixed', table: 'Table',
}

const isPieType = computed(() => ['pie', 'doughnut'].includes(currentType.value))

const xAxisLabelMap: Record<string, string> = {
  state: '状态', type: '类型', priority: '优先级', assignee: '负责人',
  reporter: '报告人', title: '标题', label: '标签', module: '模块',
  created_at: '创建日期', updated_at: '更新日期',
}

const yAxisLabelMap: Record<string, string> = {
  count: '数量', avg_days: '平均天数', throughput: '吞吐量', wip: '在制品',
  in_progress: '进行中', done: '已完成', avg_cycle_time: '平均周期时间',
  avg_resolution_days: '平均解决天数',
}

function axisLabel(field: string): string {
  if (field.startsWith('custom_field_avg:')) {
    return '自定义字段(平均)'
  }
  if (field.startsWith('custom_field_count:')) {
    return '自定义字段(计数)'
  }
  if (field.startsWith('custom_field:')) {
    return '自定义字段(' + field.split(':')[1] + ')'
  }
  return xAxisLabelMap[field] || yAxisLabelMap[field] || field
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

// Render chart when data changes
watch(renderData, async (data) => {
  if (data && data.labels.length > 0) {
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
