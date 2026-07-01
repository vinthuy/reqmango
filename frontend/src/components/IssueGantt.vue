<template>
  <div class="bg-white rounded-lg border border-gray-200 overflow-hidden">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200">
      <div class="flex items-center gap-3">
        <button @click="zoomOut" class="p-1 hover:bg-gray-100 rounded text-sm">🔍-</button>
        <span class="text-sm text-gray-500">{{ daysPerTick }}d/col</span>
        <button @click="zoomIn" class="p-1 hover:bg-gray-100 rounded text-sm">🔍+</button>
      </div>
      <div class="flex items-center gap-2">
        <select v-model="groupBy" class="text-xs border border-gray-300 rounded px-2 py-1">
          <option value="assignee">{{ t('issueGantt.groupByAssignee') }}</option>
          <option value="state">{{ t('issueGantt.groupByState') }}</option>
          <option value="none">{{ t('issueGantt.groupByFlat') }}</option>
        </select>
        <span class="text-xs text-gray-400">{{ issues.length }} issues</span>
      </div>
    </div>

    <!-- Timeline header -->
    <div class="flex border-b border-gray-200 overflow-x-auto" ref="scrollContainer">
      <div class="w-48 shrink-0 border-r border-gray-200 px-3 py-2 text-xs font-medium text-gray-500 bg-gray-50">Task</div>
      <div class="flex">
        <div v-for="tick in ticks" :key="tick.date" class="border-r border-gray-100 text-center text-[10px] text-gray-400" :style="{ width: colWidth + 'px' }">
          {{ tick.label }}
        </div>
      </div>
    </div>

    <!-- Timeline body -->
    <div class="overflow-x-auto overflow-y-auto max-h-[600px]" @scroll="syncScroll">
      <div v-for="group in groups" :key="group.label">
        <!-- Group header -->
        <div class="flex border-b border-gray-100 bg-gray-50">
          <div class="w-48 shrink-0 border-r border-gray-200 px-3 py-1.5 text-xs font-semibold text-gray-600">{{ group.label }} ({{ group.issues.length }})</div>
          <div class="flex" :style="{ width: ticks.length * colWidth + 'px' }">
            <div v-for="tick in ticks" :key="tick.date" class="border-r border-gray-50" :style="{ width: colWidth + 'px' }"></div>
          </div>
        </div>
        <!-- Issue rows -->
        <div v-for="issue in group.issues" :key="issue.id" class="flex border-b border-gray-50 hover:bg-gray-50 transition">
          <div class="w-48 shrink-0 border-r border-gray-200 px-3 py-1.5 flex items-center gap-2">
            <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: priorityColor(issue.priority) }"></span>
            <span class="text-xs text-gray-700 truncate cursor-pointer hover:text-indigo-600" @click="$emit('select', issue)">{{ issue.name }}</span>
          </div>
          <div class="relative flex items-center" :style="{ width: ticks.length * colWidth + 'px', height: '28px' }">
            <!-- Today line -->
            <div v-if="showTodayLine(ticks)" class="absolute top-0 bottom-0 w-px bg-red-400 z-10" :style="{ left: todayPosition + 'px' }"></div>
            <!-- Issue bar -->
            <div
              v-if="issue.start_date || issue.target_date"
              class="absolute top-1.5 h-4 rounded-full opacity-80 hover:opacity-100 cursor-pointer"
              :style="{ left: barLeft(issue) + 'px', width: Math.max(barWidth(issue), 8) + 'px', backgroundColor: priorityColor(issue.priority) }"
              :title="issue.name + ': ' + (issue.start_date || '?') + ' → ' + (issue.target_date || '?')"
              @click="$emit('select', issue)"
            ></div>
            <!-- Dot for issues without date -->
            <div v-else class="w-1.5 h-1.5 rounded-full bg-gray-300 ml-2"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex items-center gap-4 px-4 py-2 border-t border-gray-100 text-[10px] text-gray-400">
      <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-red-400"></span>Today</span>
      <span>● Urgent</span><span>● High</span><span>● Medium</span><span>● Low</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import api from '@/api'

const { t } = useI18n()

const props = defineProps<{ projectId: number; workspaceId: number; rql?: string; filterSortBy?: string; filterSortDir?: string }>()
defineEmits<{ (e: 'select', issue: any): void }>()

const issues = ref<any[]>([])
const groupBy = ref('assignee')
const daysPerTick = ref(7)
const colWidth = 40
const scrollContainer = ref<HTMLElement | null>(null)

// Compute date range
const dateRange = computed(() => {
  const dates = issues.value
    .flatMap(i => [i.start_date, i.target_date])
    .filter(Boolean)
    .map((d: string) => new Date(d))
  if (dates.length === 0) {
    const now = new Date()
    return { start: new Date(now.getFullYear(), now.getMonth() - 1, 1), end: new Date(now.getFullYear(), now.getMonth() + 2, 0) }
  }
  const min = new Date(Math.min(...dates.map(d => d.getTime())))
  const max = new Date(Math.max(...dates.map(d => d.getTime())))
  return { start: new Date(min.getFullYear(), min.getMonth() - 1, 1), end: new Date(max.getFullYear(), max.getMonth() + 2, 0) }
})

const ticks = computed(() => {
  const result: { date: string; label: string }[] = []
  const d = new Date(dateRange.value.start)
  while (d <= dateRange.value.end) {
    result.push({ date: d.toISOString().slice(0, 10), label: `${d.getMonth() + 1}/${d.getDate()}` })
    d.setDate(d.getDate() + daysPerTick.value)
  }
  return result
})

const groups = computed(() => {
  const map = new Map<string, any[]>()
  for (const issue of issues.value) {
    let key = 'All'
    if (groupBy.value === 'assignee') {
      if (issue.assignees?.length) {
        issue.assignees.forEach((a: any) => {
          const k = a.display_name || a.email || `User#${a.id}`
          if (!map.has(k)) map.set(k, [])
          map.get(k)!.push(issue)
        })
        continue
      }
      key = 'Unassigned'
    } else if (groupBy.value === 'state') {
      key = issue.state_name || issue.state_id || 'Unknown'
    }
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(issue)
  }
  return Array.from(map.entries()).map(([label, issues]) => ({ label, issues }))
})

function barLeft(issue: any): number {
  const start = issue.start_date ? new Date(issue.start_date) : new Date(issue.target_date)
  const offset = (start.getTime() - dateRange.value.start.getTime()) / (1000 * 86400)
  return (offset / daysPerTick.value) * colWidth
}

function barWidth(issue: any): number {
  if (!issue.start_date || !issue.target_date) return 8
  const s = new Date(issue.start_date).getTime()
  const e = new Date(issue.target_date).getTime()
  const days = Math.max((e - s) / (1000 * 86400), 0.5)
  return (days / daysPerTick.value) * colWidth
}

const todayPosition = computed(() => {
  const now = new Date()
  const offset = (now.getTime() - dateRange.value.start.getTime()) / (1000 * 86400)
  return (offset / daysPerTick.value) * colWidth
})

function showTodayLine(ticks: any[]) {
  if (ticks.length === 0) return false
  const now = new Date()
  return now >= dateRange.value.start && now <= dateRange.value.end
}

function zoomIn() { if (daysPerTick.value > 1) daysPerTick.value = Math.max(1, daysPerTick.value - 2) }
function zoomOut() { daysPerTick.value = Math.min(30, daysPerTick.value + 2) }

function syncScroll(e: Event) {
  const el = e.target as HTMLElement
  const header = el.parentElement?.querySelector('.flex.overflow-x-auto') as HTMLElement
  if (header) header.scrollLeft = el.scrollLeft
}

function priorityColor(p: string) {
  const m: Record<string,string> = { urgent:'#EF4444', high:'#F59E0B', medium:'#3B82F6', low:'#6B7280', none:'#9CA3AF' }
  return m[p] || '#9CA3AF'
}

onMounted(() => load())
watch(() => [props.projectId, props.rql], () => load())

async function load() {
  try {
    let url = `/issues?project_id=${props.projectId}&limit=500`
    if (props.rql) url += `&rql=${encodeURIComponent(props.rql)}`
    if (props.filterSortBy) {
      url += `&sort_by=${props.filterSortBy}&sort_dir=${props.filterSortDir || 'desc'}`
    }
    const r = await api.get(url)
    issues.value = Array.isArray(r.data) ? r.data : []
  } catch (_) { issues.value = [] }
}
</script>
