<template>
  <div class="bg-white rounded-lg border border-gray-200">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200">
      <button @click="prevMonth" class="p-1 hover:bg-gray-100 rounded">←</button>
      <h3 class="text-sm font-semibold text-gray-800">{{ monthLabel }}</h3>
      <button @click="nextMonth" class="p-1 hover:bg-gray-100 rounded">→</button>
    </div>

    <!-- Day headers -->
    <div class="grid grid-cols-7 text-center text-xs font-medium text-gray-500 border-b border-gray-100">
      <div v-for="d in ['Sun','Mon','Tue','Wed','Thu','Fri','Sat']" :key="d" class="py-2">{{ d }}</div>
    </div>

    <!-- Calendar grid -->
    <div class="grid grid-cols-7 auto-rows-fr">
      <div
        v-for="(day, idx) in days"
        :key="idx"
        class="min-h-[80px] border-b border-r border-gray-50 p-1 text-xs"
        :class="[day.isCurrentMonth ? 'bg-white' : 'bg-gray-50 text-gray-400', day.isToday ? 'bg-indigo-50' : '']"
      >
        <div class="font-medium mb-0.5" :class="day.isToday ? 'text-indigo-600' : ''">{{ day.dayNum }}</div>
        <div v-for="issue in day.issues" :key="issue.id" class="mb-0.5">
          <div
            @click="$emit('select', issue)"
            class="truncate px-1 py-0.5 rounded cursor-pointer hover:opacity-80 text-white text-[10px]"
            :style="{ backgroundColor: priorityColor(issue.priority) }"
            :title="issue.name"
          >
            {{ issue.name }}
          </div>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex items-center gap-3 px-4 py-2 border-t border-gray-100 text-[10px] text-gray-400">
      <span>● Urgent</span><span>● High</span><span>● Medium</span><span>● Low</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import api from '@/api'

const props = defineProps<{ projectId: number; workspaceId: number }>()
defineEmits<{ (e: 'select', issue: any): void }>()

interface CalendarDay {
  dayNum: number; isCurrentMonth: boolean; isToday: boolean; date: string
  issues: any[]
}

const currentDate = ref(new Date())
const issues = ref<any[]>([])

const monthLabel = computed(() => {
  return currentDate.value.toLocaleDateString('zh-CN', { year:'numeric', month:'long' })
})

const days = computed<CalendarDay[]>(() => {
  const year = currentDate.value.getFullYear()
  const month = currentDate.value.getMonth()
  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)
  const startOffset = firstDay.getDay() // 0=Sun
  const today = new Date().toDateString()

  const result: CalendarDay[] = []
  // Previous month fill
  const prevLast = new Date(year, month, 0).getDate()
  for (let i = startOffset - 1; i >= 0; i--) {
    const d = new Date(year, month - 1, prevLast - i)
    result.push({ dayNum: prevLast - i, isCurrentMonth: false, isToday: false, date: fmt(d), issues: issuesForDate(d) })
  }
  // Current month
  for (let i = 1; i <= lastDay.getDate(); i++) {
    const d = new Date(year, month, i)
    result.push({ dayNum: i, isCurrentMonth: true, isToday: d.toDateString() === today, date: fmt(d), issues: issuesForDate(d) })
  }
  // Next month fill
  const remaining = 42 - result.length
  for (let i = 1; i <= remaining; i++) {
    const d = new Date(year, month + 1, i)
    result.push({ dayNum: i, isCurrentMonth: false, isToday: false, date: fmt(d), issues: issuesForDate(d) })
  }
  return result
})

function fmt(d: Date) { return d.toISOString().slice(0, 10) }

function issuesForDate(date: Date) {
  const ds = fmt(date)
  return issues.value.filter(i => i.target_date && i.target_date.slice(0, 10) === ds)
}

function prevMonth() { currentDate.value = new Date(currentDate.value.getFullYear(), currentDate.value.getMonth() - 1, 1) }
function nextMonth() { currentDate.value = new Date(currentDate.value.getFullYear(), currentDate.value.getMonth() + 1, 1) }

function priorityColor(p: string) {
  const m: Record<string,string> = { urgent:'#EF4444', high:'#F59E0B', medium:'#3B82F6', low:'#6B7280', none:'#9CA3AF' }
  return m[p] || '#9CA3AF'
}

onMounted(() => load())
watch(() => props.projectId, () => load())

async function load() {
  try {
    // Fetch issues for a wide date range around current month
    const r = await api.get(`/issues?project_id=${props.projectId}&limit=500`)
    issues.value = Array.isArray(r.data) ? r.data : []
  } catch (_) { issues.value = [] }
}
</script>
