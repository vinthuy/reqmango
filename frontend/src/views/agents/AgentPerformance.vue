<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h1 class="text-xl font-semibold text-gray-800">📊 Agent Performance Analytics</h1>
          <p class="text-sm text-gray-500 mt-0.5">Execution efficiency, success rate, and failure breakdown</p>
        </div>
        <div class="flex items-center gap-2">
          <select
            v-model="rangePreset"
            class="border border-gray-300 rounded-lg px-3 py-1.5 text-sm bg-white"
            @change="loadAll"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
            <option value="all">All time</option>
          </select>
          <button
            class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 bg-white hover:bg-gray-50"
            @click="loadAll"
          >
            Refresh
          </button>
        </div>
      </div>
    </header>

    <main class="p-6">
      <div class="max-w-7xl mx-auto space-y-8">
        <!-- Overview stats -->
        <section>
          <div v-if="loadingOverview" class="text-sm text-gray-400">Loading overview…</div>
          <div v-else-if="overview" class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="text-2xl font-bold text-indigo-600">{{ overview.total_tasks }}</div>
              <div class="text-xs text-gray-500 mt-1">Total Tasks</div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="text-2xl font-bold text-green-600">{{ overview.completed_tasks }}</div>
              <div class="text-xs text-gray-500 mt-1">Completed</div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="text-2xl font-bold text-red-600">{{ overview.failed_tasks }}</div>
              <div class="text-xs text-gray-500 mt-1">Failed</div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="text-2xl font-bold text-blue-600">{{ overview.running_tasks }}</div>
              <div class="text-xs text-gray-500 mt-1">Running</div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="text-2xl font-bold text-amber-600">{{ overview.pending_tasks }}</div>
              <div class="text-xs text-gray-500 mt-1">Pending</div>
            </div>
            <div class="bg-white border border-gray-200 rounded-xl p-4">
              <div class="text-2xl font-bold text-emerald-600">{{ overview.success_rate.toFixed(1) }}%</div>
              <div class="text-xs text-gray-500 mt-1">Success Rate</div>
            </div>
          </div>
        </section>

        <!-- Avg duration banner -->
        <section v-if="overview" class="bg-white border border-gray-200 rounded-xl p-4">
          <div class="flex items-center justify-between flex-wrap gap-2">
            <div>
              <div class="text-xs text-gray-500 uppercase tracking-wide">Avg execution duration</div>
              <div class="text-lg font-semibold text-gray-800 mt-0.5">
                {{ formatDuration(overview.avg_duration_seconds) }}
              </div>
            </div>
            <div class="text-right">
              <div class="text-xs text-gray-500 uppercase tracking-wide">Total time spent</div>
              <div class="text-lg font-semibold text-gray-800 mt-0.5">
                {{ formatDuration(overview.total_duration_seconds) }}
              </div>
            </div>
          </div>
        </section>

        <!-- Timeline -->
        <section class="bg-white border border-gray-200 rounded-xl p-5">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-base font-semibold text-gray-800">Task Volume &amp; Success Rate</h2>
            <div class="flex items-center gap-2">
              <select
                v-model="bucket"
                class="border border-gray-300 rounded-lg px-2 py-1 text-sm bg-white"
                @change="loadTimeline"
              >
                <option value="day">Daily</option>
                <option value="week">Weekly</option>
                <option value="month">Monthly</option>
              </select>
            </div>
          </div>
          <div v-if="loadingTimeline" class="text-sm text-gray-400 py-8 text-center">Loading timeline…</div>
          <div v-else-if="timeline.length === 0" class="text-sm text-gray-400 py-8 text-center">
            No task data for the selected period.
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="point in timeline"
              :key="point.bucket_start"
              class="flex items-center gap-3"
            >
              <div class="w-32 text-xs text-gray-500 shrink-0">{{ formatBucket(point.bucket_start) }}</div>
              <div class="flex-1 flex items-center gap-2">
                <div class="flex-1 h-6 bg-gray-100 rounded relative overflow-hidden">
                  <div
                    class="h-full bg-indigo-400"
                    :style="{ width: barWidth(point.total_tasks) + '%' }"
                  ></div>
                  <div class="absolute inset-0 flex items-center px-2 text-xs text-gray-700">
                    {{ point.total_tasks }} tasks · {{ point.completed_tasks }} ok · {{ point.failed_tasks }} fail
                  </div>
                </div>
                <div class="w-20 text-xs text-right" :class="successRateColor(point.success_rate)">
                  {{ point.success_rate.toFixed(1) }}%
                </div>
              </div>
            </div>
          </div>
        </section>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Per-template breakdown -->
          <section class="bg-white border border-gray-200 rounded-xl p-5">
            <h2 class="text-base font-semibold text-gray-800 mb-4">Performance by Agent Template</h2>
            <div v-if="loadingTemplates" class="text-sm text-gray-400 py-6 text-center">Loading…</div>
            <div v-else-if="templates.length === 0" class="text-sm text-gray-400 py-6 text-center">
              No template data yet.
            </div>
            <div v-else class="overflow-x-auto">
              <table class="min-w-full text-sm">
                <thead>
                  <tr class="text-left text-xs text-gray-500 uppercase tracking-wide border-b border-gray-200">
                    <th class="py-2 pr-3">Template</th>
                    <th class="py-2 px-3">Total</th>
                    <th class="py-2 px-3">Success</th>
                    <th class="py-2 px-3">Avg Dur</th>
                    <th class="py-2 pl-3">Last Run</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr
                    v-for="t in templates"
                    :key="`${t.agent_template_id ?? 'unassigned'}-${t.task_type}`"
                  >
                    <td class="py-2 pr-3">
                      <div class="font-medium text-gray-800">{{ t.template_name }}</div>
                      <div v-if="t.task_type" class="text-xs text-gray-400">{{ t.task_type }}</div>
                    </td>
                    <td class="py-2 px-3 text-gray-700">{{ t.total_tasks }}</td>
                    <td class="py-2 px-3" :class="successRateColor(t.success_rate)">
                      {{ t.success_rate.toFixed(1) }}%
                    </td>
                    <td class="py-2 px-3 text-gray-700">{{ formatDuration(t.avg_duration_seconds) }}</td>
                    <td class="py-2 pl-3 text-xs text-gray-500">
                      {{ t.last_run_at ? formatBucket(t.last_run_at) : '—' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <!-- Failure breakdown -->
          <section class="bg-white border border-gray-200 rounded-xl p-5">
            <h2 class="text-base font-semibold text-gray-800 mb-4">Failure Reasons</h2>
            <div v-if="loadingFailures" class="text-sm text-gray-400 py-6 text-center">Loading…</div>
            <div v-else-if="failures.length === 0" class="text-sm text-gray-400 py-6 text-center">
              No failed tasks in this period.
            </div>
            <div v-else class="space-y-3">
              <div v-for="f in failures" :key="f.failure_reason">
                <div class="flex items-center justify-between text-sm mb-1">
                  <span class="font-medium text-gray-700">{{ humanizeReason(f.failure_reason) }}</span>
                  <span class="text-gray-500">{{ f.count }} · {{ f.percentage.toFixed(1) }}%</span>
                </div>
                <div class="h-2 bg-gray-100 rounded overflow-hidden">
                  <div class="h-full bg-red-400" :style="{ width: f.percentage + '%' }"></div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  agentPerformanceApi,
  type AgentPerformanceOverview,
  type TemplatePerformance,
  type TimelinePoint,
  type FailureBreakdown,
  type BucketGranularity
} from '@/api/agent-performance'
import { useWorkspaceId } from '@/composables/useWorkspaceId'

const { getWorkspaceId } = useWorkspaceId()

const overview = ref<AgentPerformanceOverview | null>(null)
const templates = ref<TemplatePerformance[]>([])
const timeline = ref<TimelinePoint[]>([])
const failures = ref<FailureBreakdown[]>([])

const loadingOverview = ref(false)
const loadingTemplates = ref(false)
const loadingTimeline = ref(false)
const loadingFailures = ref(false)

const rangePreset = ref<'7d' | '30d' | '90d' | 'all'>('30d')
const bucket = ref<BucketGranularity>('day')

const period = computed(() => {
  if (rangePreset.value === 'all') return {}
  const days = rangePreset.value === '7d' ? 7 : rangePreset.value === '30d' ? 30 : 90
  const end = new Date()
  const start = new Date(end.getTime() - days * 24 * 60 * 60 * 1000)
  return { from: start.toISOString(), to: end.toISOString() }
})

const maxTimelineTotal = computed(() => {
  return timeline.value.reduce((m, p) => Math.max(m, p.total_tasks), 1)
})

function barWidth(total: number): number {
  return Math.max(2, Math.round((total / maxTimelineTotal.value) * 100))
}

function successRateColor(rate: number): string {
  if (rate >= 80) return 'text-green-600'
  if (rate >= 50) return 'text-amber-600'
  return 'text-red-600'
}

function formatDuration(seconds: number): string {
  if (!seconds || seconds <= 0) return '—'
  if (seconds < 60) return `${seconds.toFixed(1)}s`
  if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`
  return `${(seconds / 3600).toFixed(2)}h`
}

function formatBucket(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: bucket.value === 'day' ? '2-digit' : undefined,
    minute: bucket.value === 'day' ? '2-digit' : undefined
  })
}

function humanizeReason(reason: string): string {
  const map: Record<string, string> = {
    agent_error: 'Agent Error',
    timeout: 'Timeout',
    runtime_offline: 'Runtime Offline',
    invalid_input: 'Invalid Input',
    model_error: 'Model Error',
    rate_limit: 'Rate Limit',
    unknown: 'Unknown'
  }
  return map[reason] || reason
}

async function loadOverview() {
  loadingOverview.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    overview.value = await agentPerformanceApi.overview(wsId, period.value)
  } catch (err) {
    console.error('Failed to load overview:', err)
    overview.value = null
  } finally {
    loadingOverview.value = false
  }
}

async function loadTemplates() {
  loadingTemplates.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    templates.value = (await agentPerformanceApi.byTemplate(wsId, period.value)) || []
  } catch (err) {
    console.error('Failed to load templates:', err)
    templates.value = []
  } finally {
    loadingTemplates.value = false
  }
}

async function loadTimeline() {
  loadingTimeline.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    timeline.value = (await agentPerformanceApi.timeline(wsId, bucket.value, period.value)) || []
  } catch (err) {
    console.error('Failed to load timeline:', err)
    timeline.value = []
  } finally {
    loadingTimeline.value = false
  }
}

async function loadFailures() {
  loadingFailures.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    failures.value = (await agentPerformanceApi.failureBreakdown(wsId, period.value)) || []
  } catch (err) {
    console.error('Failed to load failures:', err)
    failures.value = []
  } finally {
    loadingFailures.value = false
  }
}

async function loadAll() {
  await Promise.all([loadOverview(), loadTemplates(), loadTimeline(), loadFailures()])
}

onMounted(loadAll)
</script>
