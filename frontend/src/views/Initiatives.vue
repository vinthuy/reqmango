<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useRoute } from 'vue-router'
import { initiativeApi, type Initiative } from '@/api/initiative'
import Roadmap from './Roadmap.vue'

const route = useRoute()
const { t } = useI18n()
const toast = useToast()
const slug = route.params.slug as string
const workspaceId = ref<number>(0)
const initiatives = ref<Initiative[]>([])
const loading = ref(true)
const showForm = ref(false)
const editing = ref<Initiative | null>(null)
const form = ref({ name: '', description: '', color: '#3b82f6', status: 'active', target_date: '', start_date: '', project_ids: [] as number[] })
const progressData = ref<Record<number, any>>({})
const viewMode = ref<'list' | 'roadmap'>('list')

onMounted(async () => {
  try {
    const wsResp = await fetch(`/api/v1/workspaces/${slug}`)
    if (!wsResp.ok) throw new Error(`Workspace fetch failed: ${wsResp.status}`)
    const body = await wsResp.json()
    const ws = body.data
    if (ws) { workspaceId.value = ws.id; await load(ws.id) }
  } catch (e) {
    console.error('Initiatives: failed to load workspace', e)
  } finally {
    loading.value = false
  }
})

async function load(wsId: number) {
  loading.value = true
  initiatives.value = await initiativeApi.list(wsId)
  for (const ini of initiatives.value) {
    try { progressData.value[ini.id] = await initiativeApi.getProgress(ini.id) } catch(e) {}
  }
  loading.value = false
}

function openCreate() {
  editing.value = null
  form.value = { name: '', description: '', color: '#3b82f6', status: 'active', target_date: '', start_date: '', project_ids: [] }
  showForm.value = true
}

async function save() {
  if (!form.value.name.trim()) {
    toast.warning(t('initiative.nameRequired'))
    return
  }
  if (editing.value) {
    await initiativeApi.update(editing.value.id, { ...form.value })
  } else {
    await initiativeApi.create(workspaceId.value, { ...form.value })
  }
  showForm.value = false; editing.value = null
  form.value = { name: '', description: '', color: '#3b82f6', status: 'active', target_date: '', start_date: '', project_ids: [] }
  await load(workspaceId.value)
}

function edit(ini: Initiative) {
  editing.value = ini
  form.value = { name: ini.name, description: ini.description || '', color: ini.color || '#3b82f6', status: ini.status, target_date: ini.target_date || '', start_date: ini.start_date || '', project_ids: (ini.projects || []).map((p: any) => p.id) }
  showForm.value = true
}

async function remove(ini: Initiative) {
  if (!confirm(t('initiative.confirmDelete', { name: ini.name }))) return
  await initiativeApi.delete(ini.id)
  await load(workspaceId.value)
}

function getStatusLabel(s: string) {
  const map: Record<string, string> = { active: t('initiative.statusActive'), completed: t('initiative.statusCompleted'), paused: t('initiative.statusPaused'), at_risk: t('initiative.statusAtRisk'), off_track: t('initiative.statusOffTrack') }
  return map[s] || s
}
function getStatusColor(s: string) {
  const map: Record<string, string> = { active: 'bg-green-100 text-green-700', completed: 'bg-blue-100 text-blue-700', paused: 'bg-gray-100 text-gray-600', at_risk: 'bg-yellow-100 text-yellow-700', off_track: 'bg-red-100 text-red-700' }
  return map[s] || 'bg-gray-100'
}
</script>

<template>
  <div class="p-6 max-w-6xl mx-auto">
    <!-- Header with view toggle -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <h1 class="text-2xl font-bold">{{ t('initiative.title') }}</h1>
        <!-- View toggle -->
        <div class="flex bg-gray-100 dark:bg-gray-700 rounded-lg p-0.5">
          <button
            @click="viewMode = 'list'"
            :class="[
              'px-3 py-1.5 text-sm rounded-md transition',
              viewMode === 'list' ? 'bg-white dark:bg-gray-600 shadow text-gray-900 dark:text-white font-medium' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
            ]"
          >
            <span class="flex items-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"/></svg>
              {{ t('initiative.listView') }}
            </span>
          </button>
          <button
            @click="viewMode = 'roadmap'"
            :class="[
              'px-3 py-1.5 text-sm rounded-md transition',
              viewMode === 'roadmap' ? 'bg-white dark:bg-gray-600 shadow text-gray-900 dark:text-white font-medium' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
            ]"
          >
            <span class="flex items-center gap-1">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l5.447 2.724A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"/></svg>
              {{ t('initiative.roadmapView') }}
            </span>
          </button>
        </div>
      </div>
      <button v-if="viewMode === 'list'" @click="openCreate" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ t('initiative.create') }}</button>
    </div>

    <!-- Roadmap view -->
    <Roadmap
      v-if="viewMode === 'roadmap'"
      :workspace-id="workspaceId"
      :slug="slug"
      @create-initiative="openCreate"
    />

    <!-- List view -->
    <template v-else>
      <div v-if="loading" class="text-gray-500">{{ t('initiative.loading') }}</div>

      <div v-else-if="initiatives.length === 0" class="text-center py-12 text-gray-400">
        <div class="text-4xl mb-2">🎯</div>
        <p>{{ t('initiative.empty') }}</p>
        <p class="text-sm mt-1">{{ t('initiative.emptyHint') }}</p>
      </div>

      <div v-else class="grid gap-4">
        <div v-for="ini in initiatives" :key="ini.id" class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 hover:shadow-md transition">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-3 h-3 rounded-full flex-shrink-0" :style="{ backgroundColor: ini.color || '#3b82f6' }"></div>
              <div>
                <h3 class="font-semibold text-lg">{{ ini.name }}</h3>
                <p v-if="ini.description" class="text-gray-500 text-sm mt-1">{{ ini.description }}</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-xs px-2 py-0.5 rounded-full" :class="getStatusColor(ini.status)">{{ getStatusLabel(ini.status) }}</span>
              <button @click="edit(ini)" class="text-gray-400 hover:text-gray-600 text-sm">{{ t('initiative.edit') }}</button>
              <button @click="remove(ini)" class="text-gray-400 hover:text-red-500 text-sm">{{ t('initiative.delete') }}</button>
            </div>
          </div>

          <div v-if="progressData[ini.id]" class="mt-4">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-500">{{ t('initiative.progress') }}</span>
              <span class="font-medium">{{ Math.round(progressData[ini.id].progress) }}%</span>
            </div>
            <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
              <div class="h-2 rounded-full transition-all" :style="{ width: progressData[ini.id].progress + '%', backgroundColor: ini.color || '#3b82f6' }"></div>
            </div>
            <div class="flex justify-between text-xs text-gray-400 mt-1">
              <span>{{ t('initiative.completedItems', { count: progressData[ini.id].completed_issues + '/' + progressData[ini.id].total_issues }) }}</span>
              <span>{{ t('initiative.projectCount', { count: progressData[ini.id].project_count }) }}</span>
            </div>
          </div>

          <div v-if="ini.projects && ini.projects.length > 0" class="mt-3 flex flex-wrap gap-1">
            <span v-for="p in ini.projects" :key="p.id" class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-0.5 rounded">{{ p.name }}</span>
          </div>

          <div v-if="ini.start_date || ini.target_date" class="mt-2 flex gap-3 text-xs text-gray-400">
            <span v-if="ini.start_date">{{ t('initiative.startDate') }} {{ ini.start_date }}</span>
            <span v-if="ini.target_date">{{ t('initiative.targetDate') }} {{ ini.target_date }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/30 z-50 flex items-center justify-center">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-lg mx-4">
        <h2 class="text-lg font-semibold mb-4">{{ editing ? t('initiative.formTitleEdit') : t('initiative.formTitleCreate') }}</h2>
        <div class="space-y-3">
          <div>
            <label class="text-sm text-gray-600 mb-1 block">{{ t('initiative.name') }}</label>
            <input v-model="form.name" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" :placeholder="t('initiative.namePlaceholder')">
          </div>
          <div>
            <label class="text-sm text-gray-600 mb-1 block">{{ t('initiative.description') }}</label>
            <textarea v-model="form.description" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600" rows="3" :placeholder="t('initiative.descriptionPlaceholder')"></textarea>
          </div>
          <div class="flex gap-3">
            <div class="flex-1">
              <label class="text-sm text-gray-600 mb-1 block">{{ t('initiative.startDate') }}</label>
              <input type="date" v-model="form.start_date" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
            </div>
            <div class="flex-1">
              <label class="text-sm text-gray-600 mb-1 block">{{ t('initiative.targetDate') }}</label>
              <input type="date" v-model="form.target_date" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
            </div>
          </div>
          <div>
            <label class="text-sm text-gray-600 mb-1 block">{{ t('initiative.color') }}</label>
            <input type="color" v-model="form.color" class="w-10 h-8 rounded cursor-pointer border">
          </div>
          <div>
            <label class="text-sm text-gray-600 mb-1 block">{{ t('initiative.status') }}</label>
            <select v-model="form.status" class="w-full border rounded-lg px-3 py-2 dark:bg-gray-700 dark:border-gray-600">
              <option value="active">{{ t('initiative.statusActive') }}</option>
              <option value="completed">{{ t('initiative.statusCompleted') }}</option>
              <option value="at_risk">{{ t('initiative.statusAtRisk') }}</option>
              <option value="off_track">{{ t('initiative.statusOffTrack') }}</option>
              <option value="paused">{{ t('initiative.statusPaused') }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end gap-2 mt-5">
          <button @click="showForm = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">{{ t('initiative.cancel') }}</button>
          <button @click="save" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">{{ editing ? t('initiative.save') : t('initiative.createBtn') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
