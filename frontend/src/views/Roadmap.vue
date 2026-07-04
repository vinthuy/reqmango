<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import api from '@/api'
import { releaseApi } from '@/api/release'
import type { Initiative } from '@/api/initiative'

const props = defineProps<{
  workspaceId: number
  slug: string
}>()

const emit = defineEmits<{
  'create-initiative': []
}>()

const { t } = useI18n()
const initiatives = ref<Initiative[]>([])
const cycles = ref<any[]>([])
const modules = ref<any[]>([])
const releases = ref<any[]>([])
const loading = ref(true)
const selectedProjectId = ref<number | null>(null)
const projects = ref<any[]>([])

// For timeline calculation
const timelineStart = ref('')
const timelineEnd = ref('')
const timelineMonths = ref<string[]>([])

onMounted(async () => {
  if (props.workspaceId) {
    await loadAll(props.workspaceId)
  } else {
    try {
      const wsResp = await api.get(`/workspaces/${props.slug}`)
      const ws = wsResp.data?.data
      if (ws) await loadAll(ws.id)
    } catch (e) {
      console.error('Roadmap: failed to load workspace', e)
    }
  }
  loading.value = false
})

watch(() => props.workspaceId, async (id) => {
  if (id) await loadAll(id)
})

async function loadAll(wsId: number) {
  loading.value = true
  // Load initiatives (authenticated)
  try { const r = await api.get(`/workspaces/${wsId}/initiatives`); initiatives.value = r.data.data || [] } catch(e) {}
  
  // Load projects (authenticated)
  try { const r = await api.get(`/projects?workspace_id=${wsId}`); projects.value = r.data || [] } catch(e) {}
  
  // Load cycles for first project (or all)
  if (projects.value.length > 0) {
    selectedProjectId.value = projects.value[0].id
    await loadProjectData(projects.value[0].id)
  }
  loading.value = false
}

async function loadProjectData(projectId: number) {
  selectedProjectId.value = projectId
  try { const r = await api.get(`/projects/${projectId}/cycles`); cycles.value = r.data.data || [] } catch(e) { cycles.value = [] }
  try { const r = await api.get(`/modules?project_id=${projectId}`); modules.value = r.data.data || [] } catch(e) { modules.value = [] }
  try { releases.value = await releaseApi.list(projectId) || [] } catch(e) { releases.value = [] }
  computeTimeline()
}

function computeTimeline() {
  const allDates: Date[] = []
  const now = new Date()
  
  initiatives.value.forEach(i => {
    if (i.start_date) allDates.push(new Date(i.start_date))
    if (i.target_date) allDates.push(new Date(i.target_date))
  })
  cycles.value.forEach(c => {
    if (c.start_date) allDates.push(new Date(c.start_date))
    if (c.end_date) allDates.push(new Date(c.end_date))
  })
  releases.value.forEach(r => {
    if (r.release_date) allDates.push(new Date(r.release_date))
  })
  
  const start = allDates.length > 0 ? new Date(Math.min(...allDates.map(d => d.getTime()))) : new Date(now.getFullYear(), now.getMonth() - 3, 1)
  const end = allDates.length > 0 ? new Date(Math.max(...allDates.map(d => d.getTime()))) : new Date(now.getFullYear(), now.getMonth() + 6, 1)
  
  timelineStart.value = start.toISOString().split('T')[0]
  timelineEnd.value = end.toISOString().split('T')[0]
  
  // Generate months
  const months: string[] = []
  const d = new Date(start.getFullYear(), start.getMonth(), 1)
  while (d <= end) {
    months.push(`${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`)
    d.setMonth(d.getMonth() + 1)
  }
  timelineMonths.value = months
}

function getTimelinePosition(startDate: string | undefined, endDate: string | undefined): { left: string; width: string } | null {
  if (!startDate) return null
  const totalDays = (new Date(timelineEnd.value).getTime() - new Date(timelineStart.value).getTime()) / 86400000
  const startDays = (new Date(startDate).getTime() - new Date(timelineStart.value).getTime()) / 86400000
  const endDays = endDate ? (new Date(endDate).getTime() - new Date(timelineStart.value).getTime()) / 86400000 : startDays + 30
  
  return {
    left: Math.max(0, (startDays / totalDays) * 100) + '%',
    width: Math.max(2, ((endDays - startDays) / totalDays) * 100) + '%'
  }
}

function getIniStatusColor(s: string) {
  const m: Record<string, string> = { active: '#22c55e', completed: '#3b82f6', at_risk: '#eab308', off_track: '#ef4444', paused: '#6b7280' }
  return m[s] || '#6b7280'
}

function getCycleStatusColor(cycle: any): string {
  if (cycle.status === 'completed') return '#3b82f6'
  if (cycle.status === 'current') return '#22c55e'
  return '#6b7280'
}

function getReleaseStatusColor(status: string): string {
  const m: Record<string, string> = { planned: '#6b7280', in_progress: '#eab308', released: '#22c55e', cancelled: '#ef4444' }
  return m[status] || '#6b7280'
}

const monthLabels = computed(() => timelineMonths.value.map(m => {
  const [, mo] = m.split('-')
  return t('roadmap.monthFormat', { month: parseInt(mo) })
}))
</script>

<template>
  <div class="p-6 max-w-full mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">{{ t('roadmap.title') }}</h1>
      <div v-if="projects.length > 1" class="flex gap-2">
        <select v-model="selectedProjectId" @change="loadProjectData(selectedProjectId!)" class="border rounded-lg px-3 py-1.5 text-sm dark:bg-gray-700 dark:border-gray-600">
          <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="text-center py-16 text-gray-400">{{ t('roadmap.loading') }}</div>

    <div v-else>
      <!-- When data exists, show timeline -->
      <template v-if="initiatives.length > 0 || cycles.length > 0 || modules.length > 0 || releases.length > 0">
        <div class="overflow-x-auto">
          <!-- Timeline header -->
          <div class="flex border-b border-gray-200 dark:border-gray-700 mb-2">
            <div class="w-48 flex-shrink-0 px-2 py-1 text-xs font-medium text-gray-400"></div>
            <div class="flex-1 flex">
              <div v-for="(month, idx) in timelineMonths" :key="month" class="flex-1 text-center text-xs text-gray-400 py-1 border-l border-gray-100 dark:border-gray-700">
                {{ monthLabels[idx] }}
              </div>
            </div>
          </div>

          <!-- Initiatives Section -->
          <div v-if="initiatives.length > 0" class="mb-8">
            <div class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-2">
              <span>🎯</span> Initiatives
            </div>
            <div v-for="ini in initiatives" :key="ini.id" class="flex items-center mb-2 group">
              <div class="w-48 flex-shrink-0 pr-3">
                <div class="flex items-center gap-2">
                  <div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{ backgroundColor: ini.color || '#3b82f6' }"></div>
                  <span class="text-sm truncate">{{ ini.name }}</span>
                </div>
              </div>
              <div class="flex-1 relative h-7">
                <div v-if="getTimelinePosition(ini.start_date, ini.target_date)" class="absolute top-1 h-5 rounded-full opacity-80 group-hover:opacity-100 transition" :style="{ left: getTimelinePosition(ini.start_date, ini.target_date)!.left, width: getTimelinePosition(ini.start_date, ini.target_date)!.width, backgroundColor: getIniStatusColor(ini.status) }"></div>
              </div>
            </div>
          </div>

          <!-- Cycles Section -->
          <div v-if="cycles.length > 0" class="mb-8">
            <div class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-2">
              <span>🔄</span> Cycles
            </div>
            <div v-for="cycle in cycles" :key="cycle.id" class="flex items-center mb-2 group">
              <div class="w-48 flex-shrink-0 pr-3">
                <span class="text-sm truncate">{{ cycle.name }}</span>
              </div>
              <div class="flex-1 relative h-7">
                <div v-if="getTimelinePosition(cycle.start_date, cycle.end_date)" class="absolute top-1 h-5 rounded-full opacity-80 group-hover:opacity-100 transition" :style="{ left: getTimelinePosition(cycle.start_date, cycle.end_date)!.left, width: getTimelinePosition(cycle.start_date, cycle.end_date)!.width, backgroundColor: getCycleStatusColor(cycle) }"></div>
              </div>
            </div>
          </div>

          <!-- Releases Section -->
          <div v-if="releases.length > 0" class="mb-8">
            <div class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-2">
              <span>🚀</span> Releases
            </div>
            <div v-for="release in releases" :key="release.id" class="flex items-center mb-2 group">
              <div class="w-48 flex-shrink-0 pr-3">
                <span class="text-sm truncate">{{ release.name }} ({{ release.version }})</span>
              </div>
              <div class="flex-1 relative h-7">
                <div v-if="release.release_date" class="absolute top-1 h-5 rounded-full opacity-80 group-hover:opacity-100 transition" :style="{ left: getTimelinePosition(release.release_date, release.release_date)!.left, width: '4px', backgroundColor: getReleaseStatusColor(release.status) }"></div>
              </div>
            </div>
          </div>

          <!-- Modules Section -->
          <div v-if="modules.length > 0">
            <div class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-2">
              <span>📦</span> Modules
            </div>
            <div v-for="mod in modules" :key="mod.id" class="flex items-center mb-2 group">
              <div class="w-48 flex-shrink-0 pr-3">
                <span class="text-sm truncate">{{ mod.name }}</span>
              </div>
              <div class="flex-1 relative h-7">
                <div v-if="getTimelinePosition(mod.start_date, mod.target_date)" class="absolute top-1 h-5 rounded-full bg-indigo-400 opacity-70 group-hover:opacity-100 transition" :style="{ left: getTimelinePosition(mod.start_date, mod.target_date)!.left, width: getTimelinePosition(mod.start_date, mod.target_date)!.width }"></div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Empty state - show when no data AND not loading -->
      <div v-else class="text-center py-16 text-gray-400">
        <div class="text-5xl mb-3">🗺️</div>
        <p class="text-lg font-medium text-gray-500 dark:text-gray-400">{{ t('roadmap.empty') }}</p>
        <p class="text-sm mt-1 mb-6">{{ t('roadmap.emptyHint') }}</p>
        <div class="flex items-center justify-center gap-3 flex-wrap">
          <button @click="emit('create-initiative')" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 transition">
            {{ t('roadmap.createInitiative') }}
          </button>
          <router-link v-if="selectedProjectId" :to="`/workspace/${props.slug}/project/${selectedProjectId}/cycles/new`" class="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition">
            {{ t('roadmap.createCycle') }}
          </router-link>
          <router-link v-if="selectedProjectId" :to="`/workspace/${props.slug}/project/${selectedProjectId}?tab=modules`" class="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition">
            {{ t('roadmap.createModule') }}
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>
