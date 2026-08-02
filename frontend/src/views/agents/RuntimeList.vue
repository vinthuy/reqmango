<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useWorkspaceId } from '@/composables/useWorkspaceId'
import { runtimeApi, type RuntimeResponse, type RuntimeCreate, type RuntimeUpdate } from '@/api/runtime'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { getWorkspaceId } = useWorkspaceId()

const workspaceId = ref(0)
const runtimes = ref<RuntimeResponse[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const saving = ref(false)

const newRuntime = ref<RuntimeCreate>({
  name: '',
  runtime_type: 'local',
  runtime_mode: 'daemon',
  endpoint: '',
  capacity: 4,
  metadata: {}
})

const editingRuntime = ref<RuntimeResponse | null>(null)
const editForm = ref<RuntimeUpdate>({})

const runtimeTypes = [
  { value: 'local', label: 'Local' },
  { value: 'cloud', label: 'Cloud' },
  { value: 'container', label: 'Container' }
]

const runtimeModes = [
  { value: 'daemon', label: 'Daemon' },
  { value: 'serverless', label: 'Serverless' },
  { value: 'k8s', label: 'Kubernetes' }
]

const statusColors: Record<string, string> = {
  active: 'bg-green-100 text-green-800',
  offline: 'bg-gray-100 text-gray-800',
  error: 'bg-red-100 text-red-800',
  idle: 'bg-blue-100 text-blue-800'
}

const filteredRuntimes = computed(() => {
  return runtimes.value.sort((a, b) => {
    if (a.status === 'active' && b.status !== 'active') return -1
    if (a.status !== 'active' && b.status === 'active') return 1
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  })
})

async function loadRuntimes() {
  loading.value = true
  try {
    const wsId = await getWorkspaceId()
    if (!wsId) return
    workspaceId.value = wsId
    runtimes.value = await runtimeApi.list(wsId) || []
  } catch (err) {
    console.error('Failed to load runtimes:', err)
    runtimes.value = []
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  newRuntime.value = {
    name: '',
    runtime_type: 'local',
    runtime_mode: 'daemon',
    endpoint: '',
    capacity: 4,
    metadata: {}
  }
  showCreateModal.value = true
}

function openEditModal(runtime: RuntimeResponse) {
  editingRuntime.value = runtime
  editForm.value = {
    name: runtime.name,
    runtime_type: runtime.runtime_type,
    runtime_mode: runtime.runtime_mode,
    endpoint: runtime.endpoint || undefined,
    capacity: runtime.capacity,
    metadata: runtime.metadata || undefined
  }
  showEditModal.value = true
}

async function handleCreate() {
  if (!newRuntime.value.name || saving.value) return
  saving.value = true
  try {
    await runtimeApi.create(workspaceId.value, newRuntime.value)
    showCreateModal.value = false
    await loadRuntimes()
  } catch (err) {
    console.error('Failed to create runtime:', err)
  } finally {
    saving.value = false
  }
}

async function handleUpdate() {
  if (!editingRuntime.value || saving.value) return
  saving.value = true
  try {
    await runtimeApi.update(workspaceId.value, editingRuntime.value.id, editForm.value)
    showEditModal.value = false
    await loadRuntimes()
  } catch (err) {
    console.error('Failed to update runtime:', err)
  } finally {
    saving.value = false
  }
}

async function handleDelete(runtime: RuntimeResponse) {
  if (!confirm(t('common.confirmDelete', { name: runtime.name }))) return
  try {
    await runtimeApi.delete(workspaceId.value, runtime.id)
    await loadRuntimes()
  } catch (err) {
    console.error('Failed to delete runtime:', err)
  }
}

async function handleHeartbeat(runtime: RuntimeResponse) {
  try {
    await runtimeApi.heartbeat(workspaceId.value, runtime.id, {
      version: runtime.version || '1.0',
      current_load: runtime.current_load
    })
    await loadRuntimes()
  } catch (err) {
    console.error('Failed to send heartbeat:', err)
  }
}

function getLoadPercent(capacity: number, currentLoad: number): number {
  if (capacity === 0) return 0
  return Math.round((currentLoad / capacity) * 100)
}

function getLoadColor(capacity: number, currentLoad: number): string {
  const percent = getLoadPercent(capacity, currentLoad)
  if (percent >= 90) return 'bg-red-500'
  if (percent >= 70) return 'bg-yellow-500'
  return 'bg-green-500'
}

onMounted(() => {
  loadRuntimes()
})
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('ai.runtime.title') }}</h1>
        <p class="text-gray-500 mt-1">{{ t('ai.runtime.description') }}</p>
      </div>
      <button
        @click="openCreateModal"
        class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors flex items-center gap-2"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
        </svg>
        {{ t('common.create') }}
      </button>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <div class="text-gray-500 text-sm">{{ t('ai.runtime.total') }}</div>
        <div class="text-2xl font-bold text-gray-900 mt-1">{{ runtimes.length }}</div>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <div class="text-gray-500 text-sm">{{ t('ai.runtime.active') }}</div>
        <div class="text-2xl font-bold text-green-600 mt-1">
          {{ runtimes.filter(r => r.status === 'active').length }}
        </div>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <div class="text-gray-500 text-sm">{{ t('ai.runtime.totalCapacity') }}</div>
        <div class="text-2xl font-bold text-blue-600 mt-1">
          {{ runtimes.reduce((sum, r) => sum + r.capacity, 0) }}
        </div>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <div class="text-gray-500 text-sm">{{ t('ai.runtime.currentLoad') }}</div>
        <div class="text-2xl font-bold text-purple-600 mt-1">
          {{ runtimes.reduce((sum, r) => sum + r.current_load, 0) }}
        </div>
      </div>
    </div>

    <!-- Runtime List -->
    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('common.name') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('ai.runtime.type') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('ai.runtime.mode') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('ai.runtime.status') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('ai.runtime.load') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('common.lastHeartbeat') }}
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                {{ t('common.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="runtime in filteredRuntimes" :key="runtime.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center mr-3">
                    <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                    </svg>
                  </div>
                  <div>
                    <div class="text-sm font-medium text-gray-900">{{ runtime.name }}</div>
                    <div class="text-xs text-gray-500">{{ runtime.endpoint || '-' }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">
                  {{ runtime.runtime_type }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="text-sm text-gray-600">{{ runtime.runtime_mode }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="['px-2 inline-flex text-xs leading-5 font-semibold rounded-full', statusColors[runtime.status] || statusColors.idle]">
                  {{ runtime.status }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="w-16 bg-gray-200 rounded-full h-2 mr-2">
                    <div :class="['h-2 rounded-full', getLoadColor(runtime.capacity, runtime.current_load)]" :style="{ width: getLoadPercent(runtime.capacity, runtime.current_load) + '%' }"></div>
                  </div>
                  <span class="text-sm text-gray-600">{{ runtime.current_load }}/{{ runtime.capacity }}</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="text-xs text-gray-500">
                  {{ runtime.last_heartbeat ? new Date(runtime.last_heartbeat).toLocaleString() : '-' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  @click="handleHeartbeat(runtime)"
                  class="text-green-600 hover:text-green-700 mr-3"
                  title="Heartbeat"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path>
                  </svg>
                </button>
                <button
                  @click="openEditModal(runtime)"
                  class="text-blue-600 hover:text-blue-700 mr-3"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  @click="handleDelete(runtime)"
                  class="text-red-600 hover:text-red-700"
                >
                  {{ t('common.delete') }}
                </button>
              </td>
            </tr>
            <tr v-if="!loading && runtimes.length === 0">
              <td colspan="7" class="px-6 py-12 text-center">
                <div class="text-gray-400">
                  <svg class="w-12 h-12 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                  </svg>
                  <p>{{ t('ai.runtime.empty') }}</p>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Modal -->
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" v-if="showCreateModal">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-900">{{ t('ai.runtime.create') }}</h2>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('common.name') }}</label>
            <input
              v-model="newRuntime.name"
              type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="Enter runtime name"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.type') }}</label>
            <select
              v-model="newRuntime.runtime_type"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
            >
              <option v-for="type in runtimeTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.mode') }}</label>
            <select
              v-model="newRuntime.runtime_mode"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
            >
              <option v-for="mode in runtimeModes" :key="mode.value" :value="mode.value">{{ mode.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.endpoint') }}</label>
            <input
              v-model="newRuntime.endpoint"
              type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="http://localhost:8080"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.capacity') }}</label>
            <input
              v-model.number="newRuntime.capacity"
              type="number"
              min="1"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="4"
            />
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showCreateModal = false"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            @click="handleCreate"
            :disabled="!newRuntime.name || saving"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-blue-300"
          >
            {{ saving ? t('common.saving') : t('common.create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" v-if="showEditModal && editingRuntime">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-lg font-semibold text-gray-900">{{ t('ai.runtime.edit') }}: {{ editingRuntime.name }}</h2>
        </div>
        <div class="px-6 py-4 space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('common.name') }}</label>
            <input
              v-model="editForm.name"
              type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.type') }}</label>
            <select
              v-model="editForm.runtime_type"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
            >
              <option v-for="type in runtimeTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.mode') }}</label>
            <select
              v-model="editForm.runtime_mode"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
            >
              <option v-for="mode in runtimeModes" :key="mode.value" :value="mode.value">{{ mode.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.endpoint') }}</label>
            <input
              v-model="editForm.endpoint"
              type="text"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="http://localhost:8080"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('ai.runtime.capacity') }}</label>
            <input
              v-model.number="editForm.capacity"
              type="number"
              min="1"
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
          <button
            @click="showEditModal = false"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            @click="handleUpdate"
            :disabled="!editForm.name || saving"
            class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-blue-300"
          >
            {{ saving ? t('common.saving') : t('common.update') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
