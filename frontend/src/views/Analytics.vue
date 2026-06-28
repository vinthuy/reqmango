<template>
  <div class="analytics-page min-h-screen bg-gray-50 p-6">
    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
      <span class="ml-3 text-gray-500">加载分析数据...</span>
    </div>

    <div v-else class="space-y-6">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <button @click="goBack" class="text-gray-400 hover:text-gray-600 mb-1 flex items-center gap-1 text-sm">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            返回项目
          </button>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">分析看板</h1>
          <p class="text-sm text-gray-500 mt-1">{{ projectName }}</p>
        </div>
      </div>

      <!-- Section 1: Overview Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-lg bg-indigo-100 dark:bg-indigo-900/50 flex items-center justify-center shrink-0">
            <svg class="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
            </svg>
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">总工作项</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ stats.total_issues ?? overviewData.total ?? 0 }}</p>
          </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-lg bg-green-100 dark:bg-green-900/50 flex items-center justify-center shrink-0">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">已完成</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ stats.completed_issues ?? overviewData.completed ?? 0 }}</p>
          </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-lg bg-blue-100 dark:bg-blue-900/50 flex items-center justify-center shrink-0">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
            </svg>
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">进行中</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ overviewData.inProgress ?? 0 }}</p>
          </div>
        </div>

        <div class="bg-white rounded-lg border border-gray-200 p-5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-lg bg-red-100 flex items-center justify-center shrink-0">
            <svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <div>
            <p class="text-sm text-gray-500">已逾期</p>
            <p class="text-2xl font-bold text-gray-900">{{ overdueCount }}</p>
          </div>
        </div>
      </div>

      <!-- Section 2: Issue Distribution -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Priority Distribution -->
        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
          <h3 class="text-sm font-semibold text-gray-700 mb-4">优先级分布</h3>
          <div class="space-y-3">
            <div v-for="item in priorityDistribution" :key="item.key">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm text-gray-600">{{ item.label }}</span>
                <span class="text-sm font-medium text-gray-800">{{ item.count }}</span>
              </div>
              <div class="w-full bg-gray-100 rounded-full h-2.5">
                <div class="h-2.5 rounded-full transition-all" :style="{ width: item.percent + '%', backgroundColor: item.color }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- State Distribution -->
        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-4">状态分布</h3>
          <div class="space-y-3">
            <div v-for="item in stateDistribution" :key="item.key">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm text-gray-600">{{ item.label }}</span>
                <span class="text-sm font-medium text-gray-800">{{ item.count }}</span>
              </div>
              <div class="w-full bg-gray-100 rounded-full h-2.5">
                <div class="h-2.5 rounded-full transition-all" :style="{ width: item.percent + '%', backgroundColor: item.color }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Section 2.5: Flow Metrics (Cycle Time / Lead Time / WIP Aging) -->
      <div v-if="flowMetrics" class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-4">流程指标</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- State Groups -->
          <div>
            <h4 class="text-xs font-medium text-gray-500 mb-3">状态分布</h4>
            <div class="space-y-2">
              <div v-for="sg in flowMetrics.state_groups" :key="sg.group" class="flex items-center justify-between">
                <span class="text-sm text-gray-600">{{ sg.group }}</span>
                <div class="flex items-center gap-2">
                  <div class="w-24 bg-gray-100 rounded-full h-2">
                    <div class="h-2 rounded-full bg-indigo-500" :style="{ width: flowMetrics.state_groups.length ? Math.round(sg.count / Math.max(...flowMetrics.state_groups.map((g: any) => g.count)) * 100) + '%' : '0%' }"></div>
                  </div>
                  <span class="text-sm font-medium text-gray-800 w-8 text-right">{{ sg.count }}</span>
                </div>
              </div>
            </div>
          </div>
          <!-- Issue Types -->
          <div>
            <h4 class="text-xs font-medium text-gray-500 mb-3">工作项类型分布</h4>
            <div class="space-y-2">
              <div v-for="it in flowMetrics.issue_types" :key="it.type_name" class="flex items-center justify-between">
                <span class="text-sm text-gray-600">{{ it.type_name }}</span>
                <span class="text-sm font-medium text-gray-800">{{ it.count }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Trend 30 Days -->
        <div class="mt-6" v-if="flowMetrics.trend_30d && flowMetrics.trend_30d.length > 0">
          <h4 class="text-xs font-medium text-gray-500 mb-3">近30天创建趋势</h4>
          <div class="flex items-end gap-1 h-32">
            <div
              v-for="t in flowMetrics.trend_30d"
              :key="t.date"
              class="flex-1 bg-indigo-100 hover:bg-indigo-200 rounded-t transition cursor-default relative group"
              :style="{ height: getTrendHeight(t.count) + '%' }"
              :title="`${t.date}: ${t.count}`"
            >
              <span class="absolute -bottom-5 left-1/2 -translate-x-1/2 text-[9px] text-gray-400 whitespace-nowrap">{{ t.date.slice(5) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Section 3: Sprint Burndown -->
      <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-4">冲刺燃尽图</h3>
          <select
            v-model="selectedCycleId"
            class="border border-gray-300 rounded-md text-sm px-3 py-1.5 bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
            @change="loadBurndown"
          >
            <option :value="0">请选择冲刺周期</option>
            <option v-for="cycle in cycles" :key="cycle.id" :value="cycle.id">{{ cycle.name }}</option>
          </select>
        </div>

        <div v-if="!selectedCycleId" class="text-center py-12 text-gray-400 text-sm">
          请选择一个冲刺周期以查看燃尽图
        </div>

        <div v-else-if="burndownLoading" class="text-center py-12">
          <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
          </svg>
        </div>

        <div v-else-if="burndownData" class="burndown-chart">
          <div class="flex items-center gap-6 mb-4 text-xs text-gray-500">
            <span class="flex items-center gap-1.5">
              <span class="w-3 h-0.5 bg-indigo-500 inline-block rounded"></span> 理想线
            </span>
            <span class="flex items-center gap-1.5">
              <span class="w-3 h-0.5 bg-orange-500 inline-block rounded"></span> 实际线
            </span>
            <span>{{ burndownData.total_issues }} 工作项 | 已过去 {{ burndownData.days_elapsed }}/{{ burndownData.total_days }} 天</span>
          </div>

          <!-- SVG Burndown Chart -->
          <svg :viewBox="`0 0 600 ${chartHeight + 60}`" class="w-full" preserveAspectRatio="xMidYMid meet">
            <!-- Y-axis grid lines -->
            <line
              v-for="tick in yTicks"
              :key="'g-' + tick"
              :x1="40" :y1="chartMargin + (chartHeight - (tick / maxY) * chartHeight)"
              :x2="580" :y2="chartMargin + (chartHeight - (tick / maxY) * chartHeight)"
              stroke="#f3f4f6" stroke-width="1"
            />
            <!-- Y-axis labels -->
            <text
              v-for="tick in yTicks"
              :key="'yl-' + tick"
              :x="34" :y="chartMargin + (chartHeight - (tick / maxY) * chartHeight) + 4"
              text-anchor="end" class="text-[10px]" fill="#9ca3af"
            >{{ tick }}</text>

            <!-- X-axis labels (days) -->
            <text
              v-for="(day, idx) in burndownDays"
              :key="'xl-' + idx"
              :x="chartMargin + (idx / (burndownDays.length - 1)) * chartWidth"
              :y="chartMargin + chartHeight + 20"
              text-anchor="middle" class="text-[10px]" fill="#9ca3af"
            >{{ day.label }}</text>

            <!-- Ideal line -->
            <polyline
              :points="idealLinePoints"
              fill="none" stroke="#6366f1" stroke-width="2" stroke-dasharray="6,3"
            />
            <!-- Ideal dots -->
            <circle
              v-for="(pt, idx) in idealPoints"
              :key="'id-' + idx"
              :cx="pt.x" :cy="pt.y" r="3"
              fill="#6366f1"
            />

            <!-- Actual line -->
            <polyline
              :points="actualLinePoints"
              fill="none" stroke="#f97316" stroke-width="2.5"
            />
            <!-- Actual dots -->
            <circle
              v-for="(pt, idx) in actualPoints"
              :key="'ad-' + idx"
              :cx="pt.x" :cy="pt.y" r="3.5"
              fill="#f97316"
            />
          </svg>

          <div class="mt-2 text-center text-xs" :class="burndownData.is_on_track ? 'text-green-600' : 'text-orange-600'">
            {{ burndownData.is_on_track ? '✓ 进度正常' : '⚠ 进度落后' }}
          </div>
        </div>
      </div>

      <!-- Section 4: Recent Activity -->
      <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-4">最近更新的工作项</h3>
        <div v-if="recentIssues.length === 0" class="text-center py-8 text-gray-400 text-sm">
          暂无近期活动
        </div>
        <div v-else class="divide-y divide-gray-100">
          <div v-for="issue in recentIssues" :key="issue.id" class="flex items-center gap-3 py-3">
            <span
              class="w-2 h-2 rounded-full shrink-0"
              :style="{ backgroundColor: stateColor(issue.state_group || '') }"
            ></span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-800 truncate">{{ issue.name }}</span>
                <span :class="priorityClass(issue.priority)" class="text-xs px-1.5 py-0.5 rounded whitespace-nowrap">{{ priorityLabel(issue.priority) }}</span>
              </div>
              <p class="text-xs text-gray-400 mt-0.5">
                <span class="text-gray-500">{{ issue.state_name }}</span>
                &middot; 更新于 {{ formatDate(issue.updated_at) }}
              </p>
            </div>
            <button
              @click="openIssue(issue.id)"
              class="text-xs text-indigo-600 hover:text-indigo-800 shrink-0"
            >查看</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { projectApi } from '@/api/project'
import { issueApi } from '@/api/issue'
import { cycleApi } from '@/api/cycle'
import api from '@/api/index'
import type { ProjectStatistics, ProjectIssuesSummary } from '@/types/project'
import type { IssueStatistics, IssueResponse } from '@/types/issue'
import type { CycleResponse, BurndownData } from '@/types/cycle'


const route = useRoute()
const router = useRouter()

const loading = ref(true)
const projectName = ref('')
const workspaceId = ref(0)
const projectId = ref(0)
const slug = ref('')

// Overview
const stats = ref<ProjectStatistics>({
  project_id: 0, project_name: '', total_issues: 0, completed_issues: 0,
  progress: 0, state_breakdown: [], active_members: 0, is_archived: false,
  in_progress_issues: 0, member_count: 0
})
const issueStats = ref<IssueStatistics>({ total: 0, by_state: {}, by_priority: {} })
const issuesSummary = ref<ProjectIssuesSummary>({ project_id: 0, project_name: '', issues: { todo: 0, in_progress: 0, done: 0, cancelled: 0 } })
const overdueCount = ref(0)

// Cycles
const cycles = ref<CycleResponse[]>([])
const selectedCycleId = ref(0)
const burndownData = ref<BurndownData | null>(null)
const burndownLoading = ref(false)

// Flow metrics
const flowMetrics = ref<any>(null)
const flowMetricsLoading = ref(false)

async function loadFlowMetrics() {
  if (!projectId.value) return
  flowMetricsLoading.value = true
  try {
    const r = await api.get(`/issues/flow-metrics?project_id=${projectId.value}`)
    flowMetrics.value = r.data
  } catch {
    flowMetrics.value = null
  } finally {
    flowMetricsLoading.value = false
  }
}

// Recent issues
const recentIssues = ref<IssueResponse[]>([])

// ---- Computed overview data ----
const overviewData = computed(() => ({
  total: stats.value.total_issues || issueStats.value.total || 0,
  completed: stats.value.completed_issues || 0,
  inProgress: stats.value.in_progress_issues || issuesSummary.value.issues?.in_progress || 0,
}))

// ---- Priority distribution ----
const priorityConfig: Record<string, { label: string; color: string }> = {
  urgent: { label: '紧急', color: '#ef4444' },
  high: { label: '高', color: '#f59e0b' },
  medium: { label: '中', color: '#3b82f6' },
  low: { label: '低', color: '#10b981' },
  none: { label: '无', color: '#9ca3af' },
}

const priorityDistribution = computed(() => {
  const byPriority = issueStats.value.by_priority || {}
  const total = issueStats.value.total || 1
  return Object.entries(priorityConfig).map(([key, cfg]) => ({
    key,
    label: cfg.label,
    color: cfg.color,
    count: byPriority[key] || 0,
    percent: Math.round(((byPriority[key] || 0) / total) * 100)
  }))
})

// ---- State distribution ----
const stateConfig: Record<string, { label: string; color: string }> = {
  backlog: { label: '待办', color: '#9ca3af' },
  todo: { label: '计划中', color: '#3b82f6' },
  in_progress: { label: '进行中', color: '#f59e0b' },
  done: { label: '已完成', color: '#10b981' },
  cancelled: { label: '已取消', color: '#ef4444' },
}

const stateDistribution = computed(() => {
  const byState = issueStats.value.by_state || {}
  const total = issueStats.value.total || 1
  return Object.entries(stateConfig).map(([key, cfg]) => ({
    key,
    label: cfg.label,
    color: cfg.color,
    count: byState[key] || 0,
    percent: Math.round(((byState[key] || 0) / total) * 100)
  }))
})

// ---- Burndown chart ----
const chartWidth = 540
const chartHeight = 220
const chartMargin = 30

const maxY = computed(() => {
  if (!burndownData.value) return 10
  return Math.max(burndownData.value.total_issues, 10)
})

const yTicks = computed(() => {
  const max = maxY.value
  const step = Math.ceil(max / 5)
  const ticks: number[] = []
  for (let i = 0; i <= 5; i++) {
    ticks.push(i * step)
  }
  return ticks
})

const burndownDays = computed(() => {
  if (!burndownData.value) return []
  const days: { label: string }[] = []
  const total = burndownData.value.total_days
  const start = new Date(burndownData.value.start_date)
  for (let i = 0; i < total; i++) {
    const d = new Date(start)
    d.setDate(d.getDate() + i)
    days.push({ label: `${d.getMonth() + 1}/${d.getDate()}` })
  }
  return days
})

const idealPoints = computed(() => {
  if (!burndownData.value) return []
  const totalDays = burndownData.value.total_days
  const totalIssues = burndownData.value.total_issues
  const points: { x: number; y: number }[] = []

  for (let i = 0; i < totalDays; i++) {
    const x = chartMargin + (i / (totalDays - 1)) * chartWidth
    const remaining = totalIssues - (i * burndownData.value.ideal_daily_burn)
    const y = chartMargin + chartHeight - (Math.max(0, remaining) / maxY.value) * chartHeight
    points.push({ x, y })
  }
  return points
})

const idealLinePoints = computed(() => {
  return idealPoints.value.map(p => `${p.x},${p.y}`).join(' ')
})

const actualPoints = computed(() => {
  if (!burndownData.value) return []
  const totalDays = burndownData.value.total_days
  const totalIssues = burndownData.value.total_issues
  const daysElapsed = burndownData.value.days_elapsed
  const actualCompleted = burndownData.value.actual_completed
  const points: { x: number; y: number }[] = []

  const stepX = chartWidth / (totalDays - 1)
  for (let i = 0; i < daysElapsed; i++) {
    const x = chartMargin + i * stepX
    const ratio = (i + 1) / daysElapsed
    const completed = Math.round(actualCompleted * ratio)
    const remaining = totalIssues - completed
    const y = chartMargin + chartHeight - (Math.max(0, remaining) / maxY.value) * chartHeight
    points.push({ x, y })
  }
  return points
})

const actualLinePoints = computed(() => {
  return actualPoints.value.map(p => `${p.x},${p.y}`).join(' ')
})

// ---- Helpers ----
function priorityClass(p: string) {
  const m: Record<string, string> = {
    urgent: 'bg-red-100 text-red-700',
    high: 'bg-orange-100 text-orange-700',
    medium: 'bg-yellow-100 text-yellow-700',
    low: 'bg-green-100 text-green-700',
    none: 'bg-gray-100 text-gray-500'
  }
  return m[p] || m.none
}

function priorityLabel(p: string) {
  const m: Record<string, string> = { urgent: '紧急', high: '高', medium: '中', low: '低', none: '无' }
  return m[p] || p
}

function stateColor(group: string) {
  const m: Record<string, string> = {
    backlog: '#9ca3af', todo: '#3b82f6', in_progress: '#f59e0b',
    done: '#10b981', cancelled: '#ef4444'
  }
  return m[group] || '#9ca3af'
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

function getTrendHeight(count: number) {
  if (!flowMetrics.value?.trend_30d?.length) return 0
  const maxVal = Math.max(...flowMetrics.value.trend_30d.map((t: any) => t.count), 1)
  return Math.round((count / maxVal) * 100)
}

function goBack() {
  router.push(`/workspace/${slug.value}/project/${projectId.value}`)
}

function openIssue(issueId: number) {
  router.push(`/workspace/${slug.value}/project/${projectId.value}/issues/${issueId}`)
}

async function loadBurndown() {
  if (!selectedCycleId.value) {
    burndownData.value = null
    return
  }
  burndownLoading.value = true
  try {
    burndownData.value = await cycleApi.getBurndownData(selectedCycleId.value)
  } catch {
    burndownData.value = null
  } finally {
    burndownLoading.value = false
  }
}

onMounted(async () => {
  slug.value = route.params.slug as string
  projectId.value = parseInt(route.params.id as string)

  try {
    // Load workspace for workspaceId
    const { workspaceApi } = await import('@/api/workspace')
    const ws = await workspaceApi.getBySlug(slug.value)
    workspaceId.value = ws.id

    const project = await projectApi.getProject(projectId.value)
    projectName.value = project.name

    const [projStats, issStats, issSummary, cyclesResult, recentResult] = await Promise.all([
      projectApi.getProjectStatistics(projectId.value).catch(() => null),
      issueApi.getIssueStatistics(projectId.value).catch(() => null),
      projectApi.getProjectIssuesSummary(projectId.value).catch(() => null),
      cycleApi.listCycles(projectId.value, { limit: 20 }).catch(() => ({ items: [], total: 0, limit: 0, offset: 0 })),
      issueApi.listIssues(projectId.value, workspaceId.value, { limit: 10 }).catch(() => ({ items: [], total: 0 })),
    ])

    if (projStats) stats.value = projStats
    if (issStats) issueStats.value = issStats
    if (issSummary) issuesSummary.value = issSummary
    cycles.value = cyclesResult.items || []
    recentIssues.value = recentResult.items || []

    // Calculate overdue count from backend's in_progress_issues or use total - completed
    overdueCount.value = (projStats?.total_issues || 0) - (projStats?.completed_issues || 0)

    // Auto-load burndown for active cycles
    const activeCycle = cycles.value.find(c => c.status === 'active')
    if (activeCycle) {
      selectedCycleId.value = activeCycle.id
      await loadBurndown()
    }

    // Load flow metrics
    await loadFlowMetrics()
  } catch (err) {
    console.error('Failed to load analytics:', err)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.analytics-page {
  max-width: 1200px;
  margin: 0 auto;
}

.burndown-chart svg {
  width: 100%;
  max-height: 320px;
}
</style>
