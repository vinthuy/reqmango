<template>
  <div class="widget-content w-full h-full">
    <!-- Loading skeleton -->
    <div v-if="!data" class="flex items-center justify-center w-full h-full py-8">
      <svg class="animate-spin h-6 w-6 text-indigo-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
    </div>

    <!-- Number Card -->
    <div v-else-if="widget.widget_type === 'number_card'" class="flex flex-col items-center justify-center py-4">
      <span class="text-3xl font-bold text-gray-900 dark:text-gray-100">{{ data?.value ?? 0 }}</span>
      <span class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ data?.label ?? '' }}</span>
    </div>

    <!-- Bar Chart (mini bar) -->
    <div v-else-if="isChart && data?.labels" class="mini-bar flex flex-col gap-2 w-full">
      <div v-for="(label, i) in data.labels.slice(0, 6)" :key="i" class="flex items-center gap-2">
        <span class="text-[11px] text-gray-500 dark:text-gray-400 w-20 text-right truncate shrink-0">{{ label }}</span>
        <div class="flex-1 h-4 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
          <div class="h-full rounded-full transition-all duration-500"
            :style="{ width: barPercent(i, data.values) + '%', backgroundColor: barColor(i) }" />
        </div>
        <span class="text-[11px] text-gray-700 dark:text-gray-300 font-medium w-8 shrink-0">{{ data.values[i] ?? 0 }}</span>
      </div>
    </div>

    <!-- Pie / Doughnut Chart -->
    <div v-else-if="isPieOrDoughnut && data?.labels" class="flex items-center gap-4 justify-center">
      <svg viewBox="0 0 80 80" class="w-24 h-24 shrink-0">
        <circle v-for="(seg, i) in pieSegments(data.values)" :key="i"
          cx="40" cy="40" r="28" fill="none"
          :stroke="pieColors[i % pieColors.length]"
          :stroke-width="widget.widget_type === 'doughnut_chart' ? 12 : 18"
          :stroke-dasharray="seg.dash" :stroke-dashoffset="seg.offset"
          transform="rotate(-90 40 40)" />
        <text v-if="widget.widget_type === 'doughnut_chart'" x="40" y="44" text-anchor="middle" font-size="11" font-weight="bold" fill="currentColor" class="fill-gray-700 dark:fill-gray-300">
          {{ data.values.reduce((s: number, v: number) => s + v, 0) }}
        </text>
      </svg>
      <div class="flex flex-col gap-1.5">
        <div v-for="(label, i) in data.labels.slice(0, 5)" :key="i" class="flex items-center gap-1.5">
          <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: pieColors[i % pieColors.length] }" />
          <span class="text-[11px] text-gray-600 dark:text-gray-400 truncate max-w-[100px]">{{ label }}</span>
          <span class="text-[11px] text-gray-700 dark:text-gray-300 font-medium">{{ data.values[i] ?? 0 }}</span>
        </div>
      </div>
    </div>

    <!-- Burndown -->
    <div v-else-if="widget.widget_type === 'burndown' && data">
      <div v-if="data.error" class="text-sm text-gray-400 py-4 text-center">{{ data.error }}</div>
      <div v-else class="w-full">
        <svg viewBox="0 0 300 120" class="w-full h-28">
          <polyline :points="getBurndownLine(data)" fill="none" stroke="#3B82F6" stroke-width="2" stroke-linejoin="round" />
          <line x1="0" y1="120" x2="300" y2="0" stroke="#EF4444" stroke-width="1" stroke-dasharray="4,3" opacity="0.5" />
        </svg>
      </div>
    </div>

    <!-- Recent List -->
    <div v-else-if="widget.widget_type === 'recent_list' && Array.isArray(data)" class="w-full">
      <table class="w-full text-[11px]">
        <thead>
          <tr class="border-b border-gray-100 dark:border-gray-700">
            <th class="text-left py-1 px-2 text-gray-400 font-medium text-[10px] uppercase">#</th>
            <th class="text-left py-1 px-2 text-gray-400 font-medium text-[10px] uppercase">{{ t('common.name') }}</th>
            <th class="text-left py-1 px-2 text-gray-400 font-medium text-[10px] uppercase">{{ t('common.state') }}</th>
            <th class="text-right py-1 px-2 text-gray-400 font-medium text-[10px] uppercase">{{ t('common.updated') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data.slice(0, 8)" :key="item.id" class="border-b border-gray-50 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/30">
            <td class="py-1.5 px-2 text-gray-400">{{ item.sequence_id }}</td>
            <td class="py-1.5 px-2 text-gray-700 dark:text-gray-300 truncate max-w-[140px]">{{ item.name }}</td>
            <td class="py-1.5 px-2">
              <span class="inline-flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: item.state_color || '#6B7280' }" />
                <span class="text-gray-600 dark:text-gray-400">{{ item.state_name }}</span>
              </span>
            </td>
            <td class="py-1.5 px-2 text-right text-gray-400">{{ item.updated_at }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Table widget -->
    <div v-else-if="widget.widget_type === 'table' && data" class="w-full overflow-auto max-h-[300px]">
      <table class="w-full text-[11px]">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700">
            <th class="text-left py-1.5 px-2 text-gray-400 font-medium text-[10px] uppercase sticky top-0 bg-white dark:bg-gray-800">{{ t('report.groupBy') }}</th>
            <th class="text-right py-1.5 px-2 text-gray-400 font-medium text-[10px] uppercase sticky top-0 bg-white dark:bg-gray-800">{{ t('report.count') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(label, i) in (data.labels ?? [])" :key="i" class="border-b border-gray-50 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/30">
            <td class="py-1.5 px-2 text-gray-700 dark:text-gray-300">
              <span class="inline-flex items-center gap-1.5">
                <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: barColor(i) }" />
                {{ label }}
              </span>
            </td>
            <td class="py-1.5 px-2 text-right text-gray-700 dark:text-gray-300 font-medium">{{ (data.values ?? [])[i] ?? 0 }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Saved Report Widget -->
    <div v-else-if="widget.widget_type === 'saved_report' && data" class="w-full">
      <div v-if="data.error" class="text-sm text-gray-400 py-4 text-center">{{ data.error }}</div>
      <template v-else-if="data.labels">
        <!-- Chart rendering based on chart_type -->
        <div v-if="isSavedReportChart" class="w-full h-full">
          <canvas ref="savedReportCanvas" class="w-full h-full"></canvas>
        </div>
        <!-- Table fallback -->
        <div v-else class="w-full overflow-auto max-h-[300px]">
          <table class="w-full text-[11px]">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-1.5 px-2 text-gray-400 font-medium text-[10px] uppercase">Label</th>
                <th class="text-right py-1.5 px-2 text-gray-400 font-medium text-[10px] uppercase">Value</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(label, i) in data.labels" :key="i" class="border-b border-gray-50 dark:border-gray-700/50">
                <td class="py-1.5 px-2 text-gray-700 dark:text-gray-300">{{ label }}</td>
                <td class="py-1.5 px-2 text-right text-gray-700 dark:text-gray-300 font-medium">{{ data.values[i] ?? 0 }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="mt-2 text-[10px] text-gray-400 text-right">{{ t('report.total') }}: {{ data.total ?? 0 }}</div>
      </template>
    </div>

    <!-- Fallback -->
    <div v-else class="flex items-center justify-center py-6 text-xs text-gray-400 dark:text-gray-500">
      {{ t('dashboard.configureWidget') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import {
  Chart,
  BarController,
  PieController,
  DoughnutController,
  LineController,
  BubbleController,
  ScatterController,
  CategoryScale,
  LinearScale,
  ArcElement,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import type { DashboardWidget } from '@/types/dashboard'

// Register Chart.js components
Chart.register(
  BarController, PieController, DoughnutController, LineController,
  BubbleController, ScatterController,
  CategoryScale, LinearScale, ArcElement, PointElement, LineElement, BarElement,
  Title, Tooltip, Legend, Filler,
)

const CHART_COLORS = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
  '#EC4899', '#06B6D4', '#84CC16', '#F97316', '#6366F1',
]

const { t } = useI18n()

const props = defineProps<{
  widget: DashboardWidget
  data: Record<string, any> | null
}>()

const isChart = computed(() =>
  ['bar_chart', 'line_chart', 'bubble_chart', 'scatter_chart', 'mixed_chart'].includes(props.widget.widget_type)
)
const isPieOrDoughnut = computed(() =>
  ['pie_chart', 'doughnut_chart'].includes(props.widget.widget_type)
)

const barColors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899']
const pieColors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#F97316']

// Saved Report Chart.js rendering
const savedReportCanvas = ref<HTMLCanvasElement | null>(null)
const chartInstance = ref<Chart | null>(null)

const isSavedReportChart = computed(() => {
  if (props.widget.widget_type !== 'saved_report') return false
  const ct = props.data?.chart_type ?? 'bar'
  return ['bar', 'pie', 'doughnut', 'line', 'bubble', 'scatter'].includes(ct)
})

function getColors(data: Record<string, any>): string[] {
  if (!data.colors) return CHART_COLORS
  const prev = new Set<string>()
  return (data.labels || []).map((label: string) => {
    if (data.colors![label] && !prev.has(data.colors![label])) {
      prev.add(data.colors![label])
      return data.colors![label]
    }
    for (const c of CHART_COLORS) {
      if (!prev.has(c)) { prev.add(c); return c }
    }
    return CHART_COLORS[0]
  })
}

function destroyChart() {
  if (chartInstance.value) {
    chartInstance.value.destroy()
    chartInstance.value = null
  }
}

function renderSavedReportChart(data: Record<string, any>) {
  destroyChart()
  if (!savedReportCanvas.value) return
  const ctx = savedReportCanvas.value.getContext('2d')
  if (!ctx) return

  const chartType = data.chart_type ?? 'bar'
  const colors = getColors(data)
  const labels = data.labels || []
  const values = data.values || []

  const baseOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: true, position: 'bottom' as const, labels: { font: { size: 10 }, padding: 10 } },
      tooltip: { backgroundColor: '#1F2937', bodyColor: '#D1D5DB', titleFont: { size: 11 }, bodyFont: { size: 11 } },
    },
  }

  if (chartType === 'bar') {
    chartInstance.value = new Chart(ctx, {
      type: 'bar',
      data: { labels, datasets: [{ label: data.type ?? 'Count', data: values, backgroundColor: colors, borderRadius: 3 }] },
      options: {
        ...baseOptions,
        scales: {
          x: { grid: { display: false }, ticks: { maxRotation: 45, font: { size: 9 } } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 9 } } },
        },
      },
    })
  } else if (chartType === 'pie') {
    chartInstance.value = new Chart(ctx, {
      type: 'pie',
      data: { labels, datasets: [{ data: values, backgroundColor: colors, borderColor: '#FFFFFF', borderWidth: 1 }] },
      options: baseOptions,
    })
  } else if (chartType === 'doughnut') {
    chartInstance.value = new Chart(ctx, {
      type: 'doughnut',
      data: { labels, datasets: [{ data: values, backgroundColor: colors, borderColor: '#FFFFFF', borderWidth: 1 }] },
      options: { ...baseOptions, cutout: '60%' },
    })
  } else if (chartType === 'line') {
    chartInstance.value = new Chart(ctx, {
      type: 'line',
      data: { labels, datasets: [{ label: data.type ?? 'Count', data: values, borderColor: '#3B82F6', backgroundColor: 'rgba(59,130,246,0.06)', fill: true, tension: 0.3, pointRadius: 2 }] },
      options: {
        ...baseOptions,
        scales: {
          x: { grid: { display: false }, ticks: { maxRotation: 45, font: { size: 9 } } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 9 } } },
        },
      },
    })
  } else if (chartType === 'bubble') {
    const maxVal = Math.max(...values, 1)
    chartInstance.value = new Chart(ctx, {
      type: 'bubble',
      data: {
        datasets: [{ label: data.type ?? 'Count', data: values.map((v: number, i: number) => ({ x: i, y: v, r: Math.max((v / maxVal) * 12, 2) })), backgroundColor: colors.map((c: string) => c + '99'), borderColor: colors, borderWidth: 1 }],
      },
      options: {
        ...baseOptions,
        plugins: { ...baseOptions.plugins, legend: { display: false, position: 'bottom' as const, labels: { font: { size: 10 }, padding: 10 } } },
        scales: {
          x: { grid: { display: false }, ticks: { font: { size: 9 }, callback: (val: any) => labels[val] || '' } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 9 } } },
        },
      },
    })
  } else if (chartType === 'scatter') {
    chartInstance.value = new Chart(ctx, {
      type: 'scatter',
      data: {
        datasets: [{ label: data.type ?? 'Count', data: values.map((v: number, i: number) => ({ x: i, y: v })), backgroundColor: colors.map((c: string) => c + 'CC'), borderColor: colors, borderWidth: 1.5, pointRadius: 4 }],
      },
      options: {
        ...baseOptions,
        plugins: { ...baseOptions.plugins, legend: { display: false, position: 'bottom' as const, labels: { font: { size: 10 }, padding: 10 } } },
        scales: {
          x: { grid: { display: false }, ticks: { font: { size: 9 }, callback: (val: any) => labels[val] || '' } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 9 } } },
        },
      },
    })
  }
}

watch(() => props.data, async (newData) => {
  if (props.widget.widget_type === 'saved_report' && newData && isSavedReportChart.value) {
    await nextTick()
    renderSavedReportChart(newData)
  }
}, { immediate: true, deep: true })

onUnmounted(() => destroyChart())

function barPercent(i: number, values: number[]): number {
  if (!values || !values.length) return 0
  const max = Math.max(...values, 1)
  return Math.round((values[i] / max) * 100)
}

function barColor(i: number): string {
  return barColors[i % barColors.length]
}

function pieSegments(values: number[]) {
  if (!values || !values.length) return [{ dash: '0 175.93', offset: 0 }]
  const total = values.reduce((s, v) => s + v, 0) || 1
  const circumference = 2 * Math.PI * 28 // ≈ 175.93
  let offset = 0
  return values.map((v) => {
    const pct = v / total
    const dash = pct * circumference
    const seg = { dash: `${dash} ${circumference - dash}`, offset: -offset }
    offset += dash
    return seg
  })
}

function getBurndownLine(data: Record<string, any>): string {
  const points = data.points ?? data.data ?? []
  if (!points.length) return ''
  const maxVal = Math.max(...points.map((p: any) => p.remaining ?? p.value ?? 0), 1)
  return points
    .map((p: any, i: number) => {
      const x = (i / Math.max(points.length - 1, 1)) * 300
      const y = 120 - ((p.remaining ?? p.value ?? 0) / maxVal) * 110
      return `${x},${y}`
    })
    .join(' ')
}
</script>
