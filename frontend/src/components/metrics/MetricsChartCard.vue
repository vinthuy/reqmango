<template>
  <div class="bg-white border border-gray-100 rounded-xl overflow-hidden hover:shadow-md transition-shadow">
    <!-- Title Bar -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-50">
      <div class="flex items-center gap-2 flex-1 mr-2 min-w-0">
        <h4 class="text-sm font-medium text-gray-700 truncate">{{ chart.name }}</h4>
        <span class="text-[10px] text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded shrink-0">{{ typeLabel }}</span>
      </div>
      <div class="flex items-center gap-1 shrink-0">
        <!-- Edit -->
        <button @click="emit('edit', chart)" class="p-1 text-gray-400 hover:text-indigo-600 rounded transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>
        <!-- Delete -->
        <button @click="emit('delete', chart.id)" class="p-1 text-gray-400 hover:text-red-500 rounded transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Chart Area -->
    <div class="px-4 py-4">
      <div v-if="loading" class="flex items-center justify-center h-48">
        <svg class="animate-spin h-6 w-6 text-gray-300" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>
      <div v-else-if="renderData" :class="['mx-auto', isPieType ? 'max-w-md' : 'max-w-full']" style="height: 280px">
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
  (e: 'type-change', chartId: number, newType: string): void
}>()

const chartCanvas = ref<HTMLCanvasElement | null>(null)
const { render: renderChart, destroy: destroyChart } = useReportChart(chartCanvas)

const loading = ref(false)
const renderData = ref<ReportResponse | null>(null)
const currentType = ref(props.chart.chart_type)

const typeLabel = computed(() => {
  const map: Record<string, string> = { bar: '柱状图', line: '折线图', pie: '饼图', doughnut: '环形图', area: '面积图', radar: '雷达图', scatter: '散点图', bubble: '气泡图', mixed: '混合图', table: '表格' }
  return map[currentType.value] || currentType.value
})

const chartTypeMap: Record<string, string> = {
  bar: 'Bar', line: 'Line', pie: 'Pie', doughnut: 'Doughnut',
}

const isPieType = computed(() => ['pie', 'doughnut'].includes(currentType.value))

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
  loading.value = true
  try {
    const res: RenderResult = await metricsApi.renderChart(props.projectId, props.chart.id)
    renderData.value = toReportResponse(res)
    await nextTick()
    await new Promise(r => setTimeout(r, 50))
    renderChart(renderData.value, chartTypeMap[currentType.value] || 'Bar')
  } catch (e) {
    console.error('Failed to render chart:', e)
    renderData.value = null
  } finally {
    loading.value = false
  }
}

watch(() => props.chart.id, () => {
  destroyChart()
  renderData.value = null
  currentType.value = props.chart.chart_type
  fetchAndRender()
})

onMounted(() => {
  fetchAndRender()
})

onUnmounted(() => {
  destroyChart()
})
</script>
