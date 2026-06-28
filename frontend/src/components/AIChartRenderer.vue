<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import {
  Chart, CategoryScale, LinearScale, BarElement, PointElement, LineElement,
  ArcElement, RadialLinearScale, Filler, Title, Tooltip, Legend, BarController,
  LineController, PieController, DoughnutController, PolarAreaController
} from 'chart.js'
import type { AIChartData } from '@/api/ai'

// Register all chart types
Chart.register(
  CategoryScale, LinearScale, BarElement, PointElement, LineElement,
  ArcElement, RadialLinearScale, Filler, Title, Tooltip, Legend,
  BarController, LineController, PieController, DoughnutController, PolarAreaController
)

const props = defineProps<{
  config: AIChartData
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null

function createChart() {
  if (!canvasRef.value || !props.config) return

  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }

  const cfg = props.config
  const ctx = canvasRef.value.getContext('2d')
  if (!ctx) return

  chartInstance = new Chart(ctx, {
    type: cfg.chart_type,
    data: {
      labels: cfg.labels,
      datasets: cfg.datasets.map(ds => ({
        label: ds.label,
        data: ds.data,
        backgroundColor: ds.backgroundColor || [
          '#6366f1', '#8b5cf6', '#ec4899', '#f43f5e',
          '#f97316', '#eab308', '#22c55e', '#06b6d4'
        ],
        borderColor: ds.borderColor || [
          '#6366f1', '#8b5cf6', '#ec4899', '#f43f5e',
          '#f97316', '#eab308', '#22c55e', '#06b6d4'
        ],
        fill: ds.fill ?? (cfg.chart_type === 'line'),
        tension: ds.tension ?? 0.3,
        borderWidth: 2,
        borderRadius: 4,
      })),
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      indexAxis: (cfg.options?.indexAxis as 'x' | 'y') || 'x',
      plugins: {
        title: {
          display: !!cfg.title,
          text: cfg.title,
          font: { size: 14, weight: 'bold' as const },
          color: document.documentElement.classList.contains('dark') ? '#e2e8f0' : '#1e293b',
        },
        legend: {
          display: cfg.options?.showLegend !== false,
          position: 'bottom' as const,
          labels: {
            usePointStyle: true,
            padding: 16,
            font: { size: 11 },
            color: document.documentElement.classList.contains('dark') ? '#94a3b8' : '#64748b',
          },
        },
        tooltip: {
          backgroundColor: document.documentElement.classList.contains('dark') ? '#1e293b' : '#1e293b',
          titleFont: { size: 12 },
          bodyFont: { size: 12 },
        },
      },
      scales: cfg.chart_type !== 'pie' && cfg.chart_type !== 'doughnut' && cfg.chart_type !== 'polarArea' ? {
        x: {
          grid: { color: document.documentElement.classList.contains('dark') ? '#334155' : '#e2e8f0' },
          ticks: { font: { size: 11 }, color: document.documentElement.classList.contains('dark') ? '#94a3b8' : '#64748b' },
        },
        y: {
          beginAtZero: true,
          grid: { color: document.documentElement.classList.contains('dark') ? '#334155' : '#e2e8f0' },
          ticks: { font: { size: 11 }, color: document.documentElement.classList.contains('dark') ? '#94a3b8' : '#64748b' },
        },
      } : undefined,
    },
  })
}

watch(() => props.config, () => {
  setTimeout(createChart, 50)
}, { deep: true })

onMounted(() => {
  setTimeout(createChart, 50)
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
})
</script>

<template>
  <div class="w-full" style="min-height: 280px; max-height: 400px;">
    <canvas ref="canvasRef"></canvas>
  </div>
</template>
