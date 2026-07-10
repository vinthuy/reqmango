/**
 * useReportChart — Chart.js 封装，用于 ReportBuilder 渲染图表
 */
import { ref, onUnmounted, type Ref } from 'vue'
import {
  Chart,
  BarController,
  PieController,
  DoughnutController,
  LineController,
  RadarController,
  PolarAreaController,
  BubbleController,
  ScatterController,
  CategoryScale,
  LinearScale,
  ArcElement,
  PointElement,
  LineElement,
  BarElement,
  RadialLinearScale,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import type { ReportResponse } from '@/api/report'

// 注册 Chart.js 组件
Chart.register(
  BarController, PieController, DoughnutController, LineController,
  RadarController, PolarAreaController, BubbleController, ScatterController,
  CategoryScale, LinearScale, ArcElement, PointElement, LineElement, BarElement,
  RadialLinearScale,
  Title, Tooltip, Legend, Filler,
)

const CHART_COLORS = [
  '#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6',
  '#EC4899', '#06B6D4', '#84CC16', '#F97316', '#6366F1',
  '#14B8A6', '#F43F5E', '#A855F7', '#0EA5E9', '#E11D48',
]

export function useReportChart(canvasRef: Ref<HTMLCanvasElement | null>) {
  const chartInstance = ref<Chart | null>(null)

  function destroy() {
    if (chartInstance.value) {
      chartInstance.value.destroy()
      chartInstance.value = null
    }
  }

  function getColors(data: ReportResponse): string[] {
    if (!data.colors) return CHART_COLORS
    const prev = new Set<string>()
    // First use colors from backend mapping
    return data.labels.map(label => {
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

  function renderBar(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    const datasets: any[] = [
      {
        label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
        data: data.values,
        backgroundColor: colors,
        borderRadius: 4,
        borderSkipped: false,
      },
    ]

    // For created_vs_resolved, add a second dataset
    if (data.values2 && data.values2.length > 0) {
      datasets[0].label = 'Created'
      datasets.push({
        label: 'Resolved',
        data: data.values2,
        backgroundColor: data.colors?.Resolved || '#10B981',
        borderRadius: 4,
        borderSkipped: false,
      })
    }

    chartInstance.value = new Chart(ctx, {
      type: 'bar',
      data: { labels: data.labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { intersect: false, mode: 'index' },
        plugins: {
          legend: { display: datasets.length > 1, position: 'bottom' as const },
          tooltip: { backgroundColor: '#1F2937', titleColor: '#F9FAFB', bodyColor: '#D1D5DB' },
        },
        scales: {
          x: { grid: { display: false }, ticks: { maxRotation: 45, font: { size: 11 } } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function renderPie(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    chartInstance.value = new Chart(ctx, {
      type: 'pie',
      data: {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: colors,
          borderColor: '#FFFFFF',
          borderWidth: 2,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'bottom' as const, labels: { padding: 16, usePointStyle: true, font: { size: 12 } } },
          tooltip: { backgroundColor: '#1F2937', bodyColor: '#D1D5DB' },
        },
      },
    })
  }

  function renderDoughnut(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    chartInstance.value = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: colors,
          borderColor: '#FFFFFF',
          borderWidth: 2,
          hoverBorderWidth: 3,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        cutout: '65%',
        plugins: {
          legend: { position: 'bottom' as const, labels: { padding: 16, usePointStyle: true, font: { size: 12 } } },
          tooltip: { backgroundColor: '#1F2937', bodyColor: '#D1D5DB' },
        },
      },
    })
  }

  function renderLine(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const datasets: any[] = [
      {
        label: data.type === 'created_vs_resolved' ? 'Created' : 'Count',
        data: data.values,
        borderColor: '#3B82F6',
        backgroundColor: 'rgba(59, 130, 246, 0.08)',
        fill: true,
        tension: 0.3,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#3B82F6',
      },
    ]

    if (data.values2 && data.values2.length > 0) {
      datasets.push({
        label: 'Resolved',
        data: data.values2,
        borderColor: '#10B981',
        backgroundColor: 'rgba(16, 185, 129, 0.08)',
        fill: true,
        tension: 0.3,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#10B981',
      })
    }

    chartInstance.value = new Chart(ctx, {
      type: 'line',
      data: { labels: data.labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { intersect: false, mode: 'index' },
        plugins: {
          legend: { position: 'bottom' as const },
          tooltip: { backgroundColor: '#1F2937', bodyColor: '#D1D5DB' },
        },
        scales: {
          x: { grid: { display: false }, ticks: { maxRotation: 45, font: { size: 11 } } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function renderRadar(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    chartInstance.value = new Chart(ctx, {
      type: 'radar',
      data: {
        labels: data.labels,
        datasets: [{
          label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
          data: data.values,
          backgroundColor: 'rgba(59, 130, 246, 0.15)',
          borderColor: '#3B82F6',
          pointBackgroundColor: '#3B82F6',
          pointRadius: 4,
          pointHoverRadius: 6,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'bottom' as const },
          tooltip: { backgroundColor: '#1F2937', bodyColor: '#D1D5DB' },
        },
        scales: {
          r: {
            beginAtZero: true,
            grid: { color: '#E5E7EB' },
            pointLabels: { font: { size: 11 } },
            ticks: { font: { size: 10 }, backdropColor: 'transparent' },
          },
        },
      },
    })
  }

  function renderPolarArea(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    chartInstance.value = new Chart(ctx, {
      type: 'polarArea',
      data: {
        labels: data.labels,
        datasets: [{
          data: data.values,
          backgroundColor: colors.map(c => c + 'CC'),
          borderColor: colors,
          borderWidth: 1,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'bottom' as const, labels: { padding: 16, usePointStyle: true, font: { size: 12 } } },
          tooltip: { backgroundColor: '#1F2937', bodyColor: '#D1D5DB' },
        },
        scales: {
          r: {
            beginAtZero: true,
            grid: { color: '#E5E7EB' },
            ticks: { font: { size: 10 }, backdropColor: 'transparent' },
          },
        },
      },
    })
  }

  function renderHorizontalBar(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    const datasets: any[] = [
      {
        label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
        data: data.values,
        backgroundColor: colors,
        borderRadius: 4,
        borderSkipped: false,
      },
    ]

    if (data.values2 && data.values2.length > 0) {
      datasets[0].label = 'Created'
      datasets.push({
        label: 'Resolved',
        data: data.values2,
        backgroundColor: data.colors?.Resolved || '#10B981',
        borderRadius: 4,
        borderSkipped: false,
      })
    }

    chartInstance.value = new Chart(ctx, {
      type: 'bar',
      data: { labels: data.labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: 'y' as const,
        interaction: { intersect: false, mode: 'index' },
        plugins: {
          legend: { display: datasets.length > 1, position: 'bottom' as const },
          tooltip: { backgroundColor: '#1F2937', titleColor: '#F9FAFB', bodyColor: '#D1D5DB' },
        },
        scales: {
          x: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
          y: { grid: { display: false }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function renderStackedBar(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    const datasets: any[] = [
      {
        label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
        data: data.values,
        backgroundColor: colors,
        borderRadius: 2,
        borderSkipped: false,
      },
    ]

    if (data.values2 && data.values2.length > 0) {
      datasets[0].label = 'Created'
      datasets.push({
        label: 'Resolved',
        data: data.values2,
        backgroundColor: data.colors?.Resolved || '#10B981',
        borderRadius: 2,
        borderSkipped: false,
      })
    }

    chartInstance.value = new Chart(ctx, {
      type: 'bar',
      data: { labels: data.labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { intersect: false, mode: 'index' },
        plugins: {
          legend: { display: datasets.length > 1, position: 'bottom' as const },
          tooltip: { backgroundColor: '#1F2937', titleColor: '#F9FAFB', bodyColor: '#D1D5DB' },
        },
        scales: {
          x: { stacked: true, grid: { display: false }, ticks: { maxRotation: 45, font: { size: 11 } } },
          y: { stacked: true, beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function renderBubble(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    const maxVal = Math.max(...data.values, 1)

    chartInstance.value = new Chart(ctx, {
      type: 'bubble',
      data: {
        datasets: [{
          label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
          data: data.values.map((v, i) => ({ x: i, y: v, r: Math.max((v / maxVal) * 20, 3) })),
          backgroundColor: colors.map(c => c + '99'),
          borderColor: colors,
          borderWidth: 1.5,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            backgroundColor: '#1F2937', titleColor: '#F9FAFB', bodyColor: '#D1D5DB',
            callbacks: {
              title: (items) => data.labels[items[0].dataIndex] || '',
              label: (item) => `${item.dataset.label}: ${(item.raw as any).y}`,
            },
          },
        },
        scales: {
          x: {
            grid: { display: false },
            ticks: {
              font: { size: 11 },
              callback: (val: any) => data.labels[val] || '',
            },
          },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function renderScatter(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)

    chartInstance.value = new Chart(ctx, {
      type: 'scatter',
      data: {
        datasets: [{
          label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
          data: data.values.map((v, i) => ({ x: i, y: v })),
          backgroundColor: colors.map(c => c + 'CC'),
          borderColor: colors,
          borderWidth: 1.5,
          pointRadius: 6,
          pointHoverRadius: 8,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            backgroundColor: '#1F2937', titleColor: '#F9FAFB', bodyColor: '#D1D5DB',
            callbacks: {
              title: (items) => data.labels[items[0].dataIndex] || '',
              label: (item) => `${item.dataset.label}: ${(item.raw as any).y}`,
            },
          },
        },
        scales: {
          x: {
            grid: { display: false },
            ticks: {
              font: { size: 11 },
              callback: (val: any) => data.labels[val] || '',
            },
          },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function renderMixed(data: ReportResponse) {
    destroy()
    if (!canvasRef.value) return
    const ctx = canvasRef.value.getContext('2d')
    if (!ctx) return

    const colors = getColors(data)
    const datasets: any[] = [
      {
        type: 'bar',
        label: data.type === 'avg_age' || data.type === 'current_age' ? 'Days' : 'Count',
        data: data.values,
        backgroundColor: colors.map(c => c + '99'),
        borderColor: colors,
        borderWidth: 1.5,
        borderRadius: 4,
        borderSkipped: false,
        order: 2,
      },
    ]

    // Add a line overlay for the same data
    datasets.push({
      type: 'line',
      label: data.values2 && data.values2.length > 0 ? 'Created' : 'Trend',
      data: data.values,
      borderColor: '#EF4444',
      backgroundColor: 'transparent',
      tension: 0.3,
      pointRadius: 4,
      pointHoverRadius: 6,
      pointBackgroundColor: '#EF4444',
      borderWidth: 2,
      order: 1,
    })

    // For created_vs_resolved, add second bar + line
    if (data.values2 && data.values2.length > 0) {
      datasets.push({
        type: 'bar',
        label: 'Resolved',
        data: data.values2,
        backgroundColor: 'rgba(16, 185, 129, 0.6)',
        borderColor: '#10B981',
        borderWidth: 1.5,
        borderRadius: 4,
        borderSkipped: false,
        order: 3,
      })
      datasets.push({
        type: 'line',
        label: 'Resolved Trend',
        data: data.values2,
        borderColor: '#10B981',
        backgroundColor: 'transparent',
        tension: 0.3,
        pointRadius: 4,
        pointHoverRadius: 6,
        pointBackgroundColor: '#10B981',
        borderWidth: 2,
        borderDash: [5, 3],
        order: 1,
      })
    }

    chartInstance.value = new Chart(ctx, {
      type: 'bar',
      data: { labels: data.labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { intersect: false, mode: 'index' },
        plugins: {
          legend: { display: true, position: 'bottom' as const },
          tooltip: { backgroundColor: '#1F2937', titleColor: '#F9FAFB', bodyColor: '#D1D5DB' },
        },
        scales: {
          x: { grid: { display: false }, ticks: { maxRotation: 45, font: { size: 11 } } },
          y: { beginAtZero: true, grid: { color: '#F3F4F6' }, ticks: { font: { size: 11 } } },
        },
      },
    })
  }

  function render(data: ReportResponse, chartType: string) {
    switch (chartType) {
      case 'Bar': return renderBar(data)
      case 'Pie': return renderPie(data)
      case 'Doughnut': return renderDoughnut(data)
      case 'Line': return renderLine(data)
      case 'Area': return renderLine(data)
      case 'Radar': return renderRadar(data)
      case 'PolarArea': return renderPolarArea(data)
      case 'HorizontalBar': return renderHorizontalBar(data)
      case 'StackedBar': return renderStackedBar(data)
      case 'Bubble': return renderBubble(data)
      case 'Scatter': return renderScatter(data)
      case 'Mixed': return renderMixed(data)
      default: return renderBar(data)
    }
  }

  onUnmounted(() => destroy())

  return { render, destroy }
}

/**
 * 导出报表数据为 CSV
 */
export function exportReportCSV(data: ReportResponse, filename = 'report.csv') {
  const rows: string[] = []

  if (data.values2 && data.values2.length > 0) {
    rows.push('Period,Created,Resolved')
    data.labels.forEach((label, i) => {
      rows.push(`"${label}",${data.values[i]},${data.values2![i]}`)
    })
  } else {
    rows.push('Label,Count')
    data.labels.forEach((label, i) => {
      rows.push(`"${label}",${data.values[i]}`)
    })
  }

  // Add BOM for Excel UTF-8 compatibility
  const blob = new Blob(['\uFEFF' + rows.join('\n')], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

/**
 * 导出图表为 PNG 图片
 */
export function exportChartPNG(canvas: HTMLCanvasElement | null, filename = 'chart.png') {
  if (!canvas) return
  const url = canvas.toDataURL('image/png')
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
}
