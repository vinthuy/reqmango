<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6" data-testid="workspace-analytics">
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">{{ t('workspaceAnalytics.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('workspaceAnalytics.subtitle') }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <div class="inline-flex rounded-md border border-gray-200 dark:border-gray-700 overflow-hidden" data-testid="wa-range">
          <button
            v-for="d in rangeOptions"
            :key="d"
            type="button"
            :class="['px-3 py-1.5 text-xs font-medium transition-colors', days === d ? 'bg-indigo-600 text-white' : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700']"
            :data-testid="`wa-range-${d}`"
            @click="days = d"
          >{{ t('workspaceAnalytics.rangeDays', { n: String(d) }) }}</button>
        </div>
        <select v-model="projectId" class="form-input text-sm" style="width: 180px" data-testid="wa-project">
          <option :value="0">{{ t('workspaceAnalytics.allProjects') }}</option>
          <option v-for="p in projectOptions" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <button type="button" :class="btnClass" :disabled="loading" data-testid="wa-refresh" @click="load">{{ t('workspaceAnalytics.refresh') }}</button>
        <button type="button" :class="btnClass" :disabled="!data || data.projects.length === 0" data-testid="wa-export" @click="exportCsv">{{ t('workspaceAnalytics.exportCsv') }}</button>
      </div>
    </div>

    <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 flex items-center justify-between" data-testid="wa-error">
      <span>{{ t('workspaceAnalytics.loadFailed') }}</span>
      <button type="button" :class="btnClass" @click="load">{{ t('workspaceAnalytics.retry') }}</button>
    </div>

    <div v-if="loading && !data" class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="i in 8" :key="i" class="h-24 rounded-lg bg-gray-100 dark:bg-gray-800 animate-pulse" />
    </div>

    <template v-if="data">
      <div :class="['grid grid-cols-2 md:grid-cols-4 gap-4 transition-opacity', loading ? 'opacity-60' : '']" data-testid="wa-kpis">
        <div v-for="k in kpis" :key="k.key" class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4" :title="k.hint" :data-testid="`wa-kpi-${k.key}`">
          <div class="text-xs text-gray-500">{{ k.label }}</div>
          <div :class="['text-2xl font-semibold mt-1', k.tone]">{{ k.value }}</div>
          <div v-if="k.sub" class="text-xs text-gray-400 mt-1">{{ k.sub }}</div>
        </div>
      </div>

      <div v-if="data.summary.total === 0" class="rounded-lg border border-dashed border-gray-300 p-10 text-center text-sm text-gray-500" data-testid="wa-empty">
        {{ t('workspaceAnalytics.empty') }}
      </div>

      <template v-else>
        <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ t('workspaceAnalytics.trendTitle') }}</h2>
            <span class="text-xs text-gray-400">{{ data.granularity === 'week' ? t('workspaceAnalytics.byWeek') : t('workspaceAnalytics.byDay') }}</span>
          </div>
          <div class="h-64"><canvas ref="trendCanvas" data-testid="wa-trend" /></div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div v-for="dist in distributions" :key="dist.key" class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4" :data-testid="`wa-dist-${dist.key}`">
            <h2 class="text-sm font-semibold text-gray-800 dark:text-gray-200 mb-3">{{ dist.title }}</h2>
            <div class="space-y-2">
              <div v-for="b in dist.items" :key="b.key" class="text-xs">
                <div class="flex justify-between mb-0.5">
                  <span class="text-gray-600 dark:text-gray-300">{{ b.label }}</span>
                  <span class="text-gray-500 tabular-nums">{{ b.count }} · {{ pct(b.count / dist.total) }}</span>
                </div>
                <div class="h-1.5 rounded bg-gray-100 dark:bg-gray-700">
                  <div class="h-1.5 rounded" :style="{ width: `${dist.total ? (b.count / dist.total) * 100 : 0}%`, background: b.color }" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-x-auto">
          <h2 class="text-sm font-semibold text-gray-800 dark:text-gray-200 px-4 pt-4 pb-2">{{ t('workspaceAnalytics.projectsTitle') }}</h2>
          <table class="w-full text-sm" data-testid="wa-projects">
            <thead>
              <tr class="text-left text-xs text-gray-500 border-b border-gray-100 dark:border-gray-700">
                <th v-for="c in projectColumns" :key="c.key" class="px-4 py-2 font-medium whitespace-nowrap cursor-pointer select-none" :class="c.numeric ? 'text-right' : ''" :data-testid="`wa-sort-${c.key}`" @click="toggleSort(c.key)">
                  {{ c.label }}<span v-if="sortKey === c.key" class="ml-0.5">{{ sortDir === 'desc' ? '↓' : '↑' }}</span>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in sortedProjects" :key="p.id" class="border-b border-gray-50 dark:border-gray-700/50 hover:bg-gray-50 dark:hover:bg-gray-700/40" data-testid="wa-project-row">
                <td class="px-4 py-2">
                  <router-link :to="`/workspace/${slug}/project/${p.id}/analytics`" class="inline-flex items-center gap-2 text-gray-800 dark:text-gray-100 hover:text-indigo-600">
                    <span class="w-2 h-2 rounded-full" :style="{ background: p.color || '#6366f1' }" />
                    <span>{{ p.name }}</span>
                    <span class="text-xs text-gray-400">{{ p.identifier }}</span>
                  </router-link>
                </td>
                <td class="px-4 py-2 text-right tabular-nums">{{ p.total }}</td>
                <td class="px-4 py-2 text-right tabular-nums">{{ p.open }}</td>
                <td class="px-4 py-2 text-right tabular-nums">{{ p.completed }}</td>
                <td class="px-4 py-2 text-right tabular-nums" :class="p.overdue > 0 ? 'text-red-600' : ''">{{ p.overdue }}</td>
                <td class="px-4 py-2 text-right tabular-nums">{{ p.created_in_range }}</td>
                <td class="px-4 py-2 text-right tabular-nums">{{ p.completed_in_range }}</td>
                <td class="px-4 py-2">
                  <div class="flex items-center justify-end gap-2">
                    <div class="w-20 h-1.5 rounded bg-gray-100 dark:bg-gray-700"><div class="h-1.5 rounded bg-green-500" :style="{ width: `${p.completion_rate * 100}%` }" /></div>
                    <span class="text-xs tabular-nums w-10 text-right">{{ pct(p.completion_rate) }}</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-x-auto">
          <h2 class="text-sm font-semibold text-gray-800 dark:text-gray-200 px-4 pt-4 pb-2">{{ t('workspaceAnalytics.assigneesTitle') }}</h2>
          <div v-if="data.assignees.length === 0" class="px-4 pb-4 text-sm text-gray-500">{{ t('workspaceAnalytics.noAssignees') }}</div>
          <table v-else class="w-full text-sm" data-testid="wa-assignees">
            <thead>
              <tr class="text-left text-xs text-gray-500 border-b border-gray-100 dark:border-gray-700">
                <th class="px-4 py-2 font-medium">{{ t('workspaceAnalytics.col.member') }}</th>
                <th class="px-4 py-2 font-medium">{{ t('workspaceAnalytics.col.open') }}</th>
                <th class="px-4 py-2 font-medium text-right">{{ t('workspaceAnalytics.col.overdue') }}</th>
                <th class="px-4 py-2 font-medium text-right">{{ t('workspaceAnalytics.col.completedInRange') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="a in data.assignees" :key="a.user_id" class="border-b border-gray-50 dark:border-gray-700/50" data-testid="wa-assignee-row">
                <td class="px-4 py-2 text-gray-800 dark:text-gray-100">{{ a.name }}</td>
                <td class="px-4 py-2">
                  <div class="flex items-center gap-2">
                    <div class="w-32 h-1.5 rounded bg-gray-100 dark:bg-gray-700"><div class="h-1.5 rounded bg-indigo-500" :style="{ width: `${maxAssigneeOpen ? (a.open / maxAssigneeOpen) * 100 : 0}%` }" /></div>
                    <span class="text-xs tabular-nums">{{ a.open }}</span>
                  </div>
                </td>
                <td class="px-4 py-2 text-right tabular-nums" :class="a.overdue > 0 ? 'text-red-600' : ''">{{ a.overdue }}</td>
                <td class="px-4 py-2 text-right tabular-nums">{{ a.completed_in_range }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <p class="text-xs text-gray-400 text-right">{{ t('workspaceAnalytics.generatedAt', { time: new Date(data.generated_at).toLocaleString() }) }}</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  Chart, CategoryScale, LinearScale, BarElement, PointElement, LineElement,
  BarController, LineController, Filler, Tooltip, Legend,
} from 'chart.js'
import { useI18n } from '@/composables/useI18n'
import { getWorkspaceAnalytics, type WorkspaceAnalytics, type WorkspaceAnalyticsProject } from '@/api/workspace-analytics'

Chart.register(CategoryScale, LinearScale, BarElement, PointElement, LineElement, BarController, LineController, Filler, Tooltip, Legend)

const route = useRoute()
const { t } = useI18n()

const btnClass = 'px-3 py-1.5 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50'
const slug = computed(() => route.params.slug as string)
const rangeOptions = [7, 14, 30, 90, 180]
const days = ref(30)
const projectId = ref(0)
const data = ref<WorkspaceAnalytics | null>(null)
const loading = ref(false)
const error = ref(false)
const projectOptions = ref<{ id: number; name: string }[]>([])
const trendCanvas = ref<HTMLCanvasElement | null>(null)
let trendChart: Chart | null = null
let requestSeq = 0

const STATE_COLORS: Record<string, string> = {
  backlog: '#9ca3af', unstarted: '#60a5fa', started: '#f59e0b', completed: '#22c55e', cancelled: '#ef4444',
}
const PRIORITY_COLORS: Record<string, string> = {
  urgent: '#dc2626', high: '#f97316', medium: '#eab308', low: '#3b82f6', none: '#9ca3af',
}
const TYPE_PALETTE = ['#6366f1', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#0ea5e9', '#84cc16', '#a855f7']

function pct(v: number) {
  return `${Math.round((v || 0) * 100)}%`
}

async function load() {
  const seq = ++requestSeq
  loading.value = true
  error.value = false
  try {
    const res = await getWorkspaceAnalytics(slug.value, { days: days.value, project_id: projectId.value || undefined })
    if (seq !== requestSeq) return
    data.value = res
    if (!projectId.value) {
      projectOptions.value = res.projects.map(p => ({ id: p.id, name: p.name })).sort((a, b) => a.name.localeCompare(b.name))
    }
    await nextTick()
    renderTrend()
  } catch {
    if (seq === requestSeq) error.value = true
  } finally {
    if (seq === requestSeq) loading.value = false
  }
}

function renderTrend() {
  if (!data.value || !trendCanvas.value) {
    trendChart?.destroy()
    trendChart = null
    return
  }
  const labels = data.value.trend.map(p => p.date.slice(5))
  const created = data.value.trend.map(p => p.created)
  const completed = data.value.trend.map(p => p.completed)
  if (trendChart && trendChart.canvas === trendCanvas.value) {
    trendChart.data.labels = labels
    trendChart.data.datasets[0].data = created
    trendChart.data.datasets[1].data = completed
    trendChart.update()
    return
  }
  trendChart?.destroy()
  trendChart = new Chart(trendCanvas.value, {
    type: 'bar',
    data: {
      labels,
      datasets: [
        { label: t('workspaceAnalytics.trendCreated'), data: created, backgroundColor: '#818cf8', borderRadius: 3 },
        { label: t('workspaceAnalytics.trendCompleted'), data: completed, backgroundColor: '#34d399', borderRadius: 3 },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: 'index', intersect: false },
      plugins: { legend: { position: 'bottom', labels: { boxWidth: 10 } } },
      scales: { y: { beginAtZero: true, ticks: { precision: 0 } }, x: { grid: { display: false } } },
    },
  })
}

const kpis = computed(() => {
  const s = data.value?.summary
  if (!s) return []
  return [
    { key: 'total', label: t('workspaceAnalytics.kpi.total'), value: s.total, sub: t('workspaceAnalytics.kpi.projects', { n: String(s.project_count) }), tone: 'text-gray-900 dark:text-gray-100', hint: '' },
    { key: 'open', label: t('workspaceAnalytics.kpi.open'), value: s.open, sub: '', tone: 'text-gray-900 dark:text-gray-100', hint: '' },
    { key: 'overdue', label: t('workspaceAnalytics.kpi.overdue'), value: s.overdue, sub: s.open ? pct(s.overdue / s.open) : '', tone: s.overdue > 0 ? 'text-red-600' : 'text-gray-900 dark:text-gray-100', hint: t('workspaceAnalytics.kpi.overdueHint') },
    { key: 'rate', label: t('workspaceAnalytics.kpi.completionRate'), value: pct(s.completion_rate), sub: '', tone: 'text-green-600', hint: '' },
    { key: 'created', label: t('workspaceAnalytics.kpi.created'), value: s.created_in_range, sub: '', tone: 'text-indigo-600', hint: '' },
    { key: 'completed', label: t('workspaceAnalytics.kpi.completed'), value: s.completed_in_range, sub: '', tone: 'text-emerald-600', hint: '' },
    { key: 'cycle', label: t('workspaceAnalytics.kpi.avgCycle'), value: s.avg_cycle_days == null ? '—' : t('workspaceAnalytics.kpi.days', { n: s.avg_cycle_days.toFixed(1) }), sub: '', tone: 'text-gray-900 dark:text-gray-100', hint: t('workspaceAnalytics.kpi.avgCycleHint') },
    { key: 'net', label: `${t('workspaceAnalytics.trendCreated')} − ${t('workspaceAnalytics.trendCompleted')}`, value: s.created_in_range - s.completed_in_range, sub: t('workspaceAnalytics.rangeDays', { n: String(data.value!.days) }), tone: s.created_in_range > s.completed_in_range ? 'text-amber-600' : 'text-green-600', hint: '' },
  ]
})

const distributions = computed(() => {
  const d = data.value
  if (!d) return []
  const sum = (b: { count: number }[]) => b.reduce((n, x) => n + x.count, 0)
  return [
    {
      key: 'state', title: t('workspaceAnalytics.stateGroups'), total: sum(d.state_groups),
      items: d.state_groups.map(b => ({ key: b.key, count: b.count, label: t(`workspaceAnalytics.stateGroup.${b.key}`), color: STATE_COLORS[b.key] || '#9ca3af' })),
    },
    {
      key: 'priority', title: t('workspaceAnalytics.priorities'), total: sum(d.priorities),
      items: d.priorities.map(b => ({ key: b.key, count: b.count, label: t(`workspaceAnalytics.priority.${b.key}`), color: PRIORITY_COLORS[b.key] || '#9ca3af' })),
    },
    {
      key: 'type', title: t('workspaceAnalytics.issueTypes'), total: sum(d.issue_types),
      items: d.issue_types.slice(0, 8).map((b, i) => ({ key: b.key || '_none', count: b.count, label: b.key || t('workspaceAnalytics.untyped'), color: TYPE_PALETTE[i % TYPE_PALETTE.length] })),
    },
  ]
})

type SortKey = 'name' | 'total' | 'open' | 'completed' | 'overdue' | 'created_in_range' | 'completed_in_range' | 'completion_rate'
const sortKey = ref<SortKey>('open')
const sortDir = ref<'asc' | 'desc'>('desc')
const projectColumns = computed<{ key: SortKey; label: string; numeric: boolean }[]>(() => [
  { key: 'name', label: t('workspaceAnalytics.col.project'), numeric: false },
  { key: 'total', label: t('workspaceAnalytics.col.total'), numeric: true },
  { key: 'open', label: t('workspaceAnalytics.col.open'), numeric: true },
  { key: 'completed', label: t('workspaceAnalytics.col.completed'), numeric: true },
  { key: 'overdue', label: t('workspaceAnalytics.col.overdue'), numeric: true },
  { key: 'created_in_range', label: t('workspaceAnalytics.col.created'), numeric: true },
  { key: 'completed_in_range', label: t('workspaceAnalytics.col.completedInRange'), numeric: true },
  { key: 'completion_rate', label: t('workspaceAnalytics.col.rate'), numeric: true },
])

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'desc' ? 'asc' : 'desc'
  } else {
    sortKey.value = key
    sortDir.value = key === 'name' ? 'asc' : 'desc'
  }
}

const sortedProjects = computed<WorkspaceAnalyticsProject[]>(() => {
  const rows = [...(data.value?.projects || [])]
  const k = sortKey.value
  const dir = sortDir.value === 'desc' ? -1 : 1
  return rows.sort((a, b) => {
    if (k === 'name') return a.name.localeCompare(b.name) * dir
    return ((a[k] as number) - (b[k] as number)) * dir
  })
})

const maxAssigneeOpen = computed(() => Math.max(0, ...(data.value?.assignees || []).map(a => a.open)))

function exportCsv() {
  if (!data.value) return
  const header = projectColumns.value.map(c => c.label)
  const rows = sortedProjects.value.map(p => [
    p.name, p.total, p.open, p.completed, p.overdue, p.created_in_range, p.completed_in_range, pct(p.completion_rate),
  ])
  const esc = (v: unknown) => {
    const s = String(v)
    return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s
  }
  const csv = '\uFEFF' + [header, ...rows].map(r => r.map(esc).join(',')).join('\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
  const a = document.createElement('a')
  a.href = url
  a.download = `workspace-analytics-${slug.value}-${data.value.days}d.csv`
  a.click()
  URL.revokeObjectURL(url)
}

watch([days, projectId], load)
watch(slug, () => {
  projectId.value = 0
  load()
})
onMounted(load)
onUnmounted(() => {
  trendChart?.destroy()
  trendChart = null
})
</script>
