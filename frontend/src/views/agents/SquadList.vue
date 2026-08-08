<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('ai.squad.title') }}</h1>
        <p class="text-gray-500 mt-1">{{ t('ai.squad.description') }}</p>
      </div>
      <button @click="openCreateModal" class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg">
        {{ t('common.create') }}
      </button>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('ai.squad.total') }}</p>
            <p class="text-2xl font-bold text-gray-900">{{ squads.length }}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">👥</div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('ai.squad.active') }}</p>
            <p class="text-2xl font-bold text-green-600">{{ activeCount }}</p>
          </div>
          <div class="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">✅</div>
        </div>
      </div>
      <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ t('ai.squad.executions') }}</p>
            <p class="text-2xl font-bold text-purple-600">{{ executionCount }}</p>
          </div>
          <div class="w-10 h-10 bg-purple-100 rounded-full flex items-center justify-center">🚀</div>
        </div>
      </div>
    </div>

    <!-- Squad List -->
    <div class="bg-white rounded-lg shadow-sm border border-gray-200">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.name') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.squad.goal') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('ai.squad.members') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.createdAt') }}</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="squad in squads" :key="squad.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center mr-3">👥</div>
                  <div>
                    <p class="font-medium text-gray-900">{{ squad.name }}</p>
                    <p class="text-sm text-gray-500">{{ squad.description }}</p>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <p class="text-sm text-gray-600 line-clamp-2">{{ squad.goal }}</p>
              </td>
              <td class="px-6 py-4">
                <div class="flex -space-x-2">
                  <div v-for="member in (squad.members || []).slice(0, 3)" :key="member.id" class="w-7 h-7 bg-gray-200 rounded-full border-2 border-white flex items-center justify-center text-xs">
                    {{ member.role.charAt(0).toUpperCase() }}
                  </div>
                  <div v-if="(squad.members || []).length > 3" class="w-7 h-7 bg-gray-100 rounded-full border-2 border-white flex items-center justify-center text-xs text-gray-500">
                    +{{ (squad.members || []).length - 3 }}
                  </div>
                </div>
              </td>
              <td class="px-6 py-4">
                <span :class="getStatusClass(squad.status)" class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full">
                  {{ getStatusText(squad.status) }}
                </span>
              </td>
              <td class="px-6 py-4">
                <p class="text-sm text-gray-500">{{ formatDate(squad.created_at) }}</p>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end space-x-2">
                  <button @click="viewSquad(squad)" class="text-gray-600 hover:text-blue-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>
                  </button>
                  <button @click="openEditModal(squad)" class="text-gray-600 hover:text-blue-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                  </button>
                  <button @click="deleteSquadConfirm(squad)" class="text-gray-600 hover:text-red-600">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="squads.length === 0">
              <td colspan="6" class="px-6 py-12 text-center">
                <div class="text-gray-400">
                  <div class="text-4xl mb-2">👥</div>
                  <p>{{ t('ai.squad.empty') }}</p>
                  <button @click="openCreateModal" class="mt-4 text-blue-600 hover:text-blue-700 font-medium">{{ t('common.create') }}</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Transition name="modal">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="closeModal"></div>
        <div class="relative bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
          <div class="flex items-center justify-between p-4 border-b">
            <h2 class="text-lg font-semibold">{{ isEditing ? t('ai.squad.edit') : t('ai.squad.create') }}</h2>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.name') }}</label>
              <input v-model="form.name" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" :placeholder="t('ai.squad.namePlaceholder')" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('common.description') }}</label>
              <textarea v-model="form.description" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" rows="3" :placeholder="t('ai.squad.descriptionPlaceholder')"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('ai.squad.goal') }}</label>
              <textarea v-model="form.goal" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500" rows="3" :placeholder="t('ai.squad.goalPlaceholder')"></textarea>
            </div>
          </div>
          <div class="flex items-center justify-end p-4 border-t space-x-2">
            <button @click="closeModal" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">{{ t('common.cancel') }}</button>
            <button @click="saveSquad" :disabled="!form.name" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50">{{ t('common.save') }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useRoute, useRouter } from 'vue-router'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import * as squadApi from '@/api/squad'
import type { Squad } from '@/api/squad'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { getWorkspaceId } = useWorkspaceId()

const workspaceId = ref(0)

const squads = ref<Squad[]>([])
const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const form = ref({
  name: '',
  description: '',
  goal: '',
})

const activeCount = computed(() => squads.value.filter(s => s.status === 'active').length)
const executionCount = computed(() => {
  return squads.value.reduce((sum, s) => sum + (s.execution_count || 0), 0)
})

async function loadSquads() {
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    squads.value = await squadApi.listSquads(wsId)
  } catch (e) {
    console.error('Failed to load squads', e)
  }
}

function openCreateModal() {
  isEditing.value = false
  editingId.value = null
  form.value = { name: '', description: '', goal: '' }
  showModal.value = true
}

function openEditModal(squad: Squad) {
  isEditing.value = true
  editingId.value = squad.id
  form.value = { name: squad.name, description: squad.description || '', goal: squad.goal || '' }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function saveSquad() {
  if (!form.value.name) return
  try {
    if (isEditing.value && editingId.value) {
      await squadApi.updateSquad(getWorkspaceId(), editingId.value, form.value)
    } else {
      await squadApi.createSquad(getWorkspaceId(), form.value)
    }
    closeModal()
    await loadSquads()
  } catch (e) {
    console.error('Save failed', e)
  }
}

async function deleteSquadConfirm(squad: Squad) {
  if (confirm(t('common.deleteConfirm'))) {
    try {
      await squadApi.deleteSquad(workspaceId.value, squad.id)
      await loadSquads()
    } catch (e) {
      console.error('Delete failed', e)
    }
  }
}

function viewSquad(squad: Squad) {
  const slug = route.params.slug as string
  if (slug) {
    router.push(`/workspace/${slug}/agents/squads/${squad.id}`)
  }
}

function getStatusClass(status: string) {
  switch (status) {
    case 'active': return 'bg-green-100 text-green-800'
    case 'inactive': return 'bg-gray-100 text-gray-800'
    default: return 'bg-gray-100 text-gray-800'
  }
}

function getStatusText(status: string) {
  switch (status) {
    case 'active': return t('common.active')
    case 'inactive': return t('common.inactive')
    default: return status
  }
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleDateString()
}

onMounted(() => {
  loadSquads()
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>