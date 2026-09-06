<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { releaseApi } from '@/api/release'
import type { Release, ReleaseCreateRequest, ReleaseUpdateRequest } from '@/types/release'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
}>()

const { confirm } = useConfirm()

const releases = ref<Release[]>([])
const loading = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingRelease = ref<Release | null>(null)
const createForm = ref<ReleaseCreateRequest>({ name: '', version: '', description: '', status: 'planned' })
const editForm = ref<ReleaseUpdateRequest>({ name: '', version: '', description: '', status: '' })

const statusColors: Record<string, string> = {
  planned: 'bg-gray-100 text-gray-700',
  in_progress: 'bg-yellow-100 text-yellow-700',
  released: 'bg-green-100 text-green-700',
  cancelled: 'bg-red-100 text-red-700'
}

const statusLabels: Record<string, string> = {
  planned: t('release.status.planned'),
  in_progress: t('release.status.inProgress'),
  released: t('release.status.released'),
  cancelled: t('release.status.cancelled')
}

async function loadReleases() {
  loading.value = true
  try {
    const result = await releaseApi.list(props.projectId)
    releases.value = result || []
  } catch (e) { console.error('Failed to load releases:', e) }
  finally { loading.value = false }
}

function openCreateModal() {
  createForm.value = { name: '', version: '', description: '', status: 'planned' }
  showCreateModal.value = true
}

function openEditModal(release: Release) {
  editingRelease.value = release
  editForm.value = { name: release.name, version: release.version, description: release.description, status: release.status }
  showEditModal.value = true
}

async function handleCreate() {
  if (!createForm.value.name || !createForm.value.version) return
  try {
    await releaseApi.create(props.projectId, createForm.value)
    showCreateModal.value = false
    await loadReleases()
  } catch (e) { console.error('Failed to create release:', e) }
}

async function handleUpdate() {
  if (!editingRelease.value) return
  try {
    await releaseApi.update(props.projectId, editingRelease.value.id, editForm.value)
    showEditModal.value = false
    await loadReleases()
  } catch (e) { console.error('Failed to update release:', e) }
}

async function handleDelete(release: Release) {
  if (!(await confirm({ title: t('common.delete'), message: t('release.deleteConfirm', { name: release.name }), danger: true, confirmText: t('common.delete') }))) return
  try {
    await releaseApi.delete(props.projectId, release.id)
    await loadReleases()
  } catch (e) { console.error('Failed to delete release:', e) }
}

onMounted(loadReleases)
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">{{ t('release.title') }}</h2>
        <p class="text-sm text-gray-500 mt-1">{{ t('release.desc') }}</p>
      </div>
      <button @click="openCreateModal" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors text-sm font-medium">
        {{ t('release.create') }}
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center h-32">
      <div class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"></div>
    </div>

    <div v-else-if="releases.length === 0" class="text-center text-gray-400 py-12">
      {{ t('release.noReleases') }}
    </div>

    <div v-else class="space-y-4">
      <div v-for="release in releases" :key="release.id" class="bg-white rounded-xl border border-gray-200 p-4 hover:border-gray-300 hover:shadow-md transition-all">
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center gap-3">
              <h3 class="font-medium text-gray-900">{{ release.name }}</h3>
              <span class="text-sm text-gray-500 font-mono">{{ release.version }}</span>
              <span :class="['px-2 py-0.5 rounded-full text-xs font-medium', statusColors[release.status] || statusColors.planned]">
                {{ statusLabels[release.status] || release.status }}
              </span>
            </div>
            <p v-if="release.description" class="mt-1 text-sm text-gray-500">{{ release.description }}</p>
            <div class="mt-2 flex items-center gap-4 text-xs text-gray-400">
              <span v-if="release.release_date">{{ t('release.releaseDate') }}: {{ (release.release_date as string).split('T')[0] }}</span>
              <span v-if="release.created_at">{{ t('common.createdAt') }}: {{ release.created_at.split('T')[0] }}</span>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button @click="openEditModal(release)" class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition">
              ✏️
            </button>
            <button @click="handleDelete(release)" class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition">
              🗑️
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('release.create') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.name') }}</label>
            <input v-model="createForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('release.namePlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.version') }}</label>
            <input v-model="createForm.version" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('release.versionPlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.description') }}</label>
            <textarea v-model="createForm.description" rows="2" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" :placeholder="t('release.descriptionPlaceholder')"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.statusLabel') }}</label>
            <select v-model="createForm.status" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
              <option value="planned">{{ t('release.status.planned') }}</option>
              <option value="in_progress">{{ t('release.status.inProgress') }}</option>
              <option value="released">{{ t('release.status.released') }}</option>
              <option value="cancelled">{{ t('release.status.cancelled') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.releaseDate') }}</label>
            <input v-model="createForm.release_date" type="date" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showCreateModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('common.cancel') }}</button>
          <button @click="handleCreate" :disabled="!createForm.name || !createForm.version" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50">{{ t('common.create') }}</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="showEditModal = false">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('release.edit') }}</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.name') }}</label>
            <input v-model="editForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.version') }}</label>
            <input v-model="editForm.version" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.description') }}</label>
            <textarea v-model="editForm.description" rows="2" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('release.statusLabel') }}</label>
            <select v-model="editForm.status" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
              <option value="planned">{{ t('release.status.planned') }}</option>
              <option value="in_progress">{{ t('release.status.inProgress') }}</option>
              <option value="released">{{ t('release.status.released') }}</option>
              <option value="cancelled">{{ t('release.status.cancelled') }}</option>
            </select>
          </div>
        </div>
        <div class="flex justify-end space-x-3 mt-6">
          <button @click="showEditModal = false" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('common.cancel') }}</button>
          <button @click="handleUpdate" :disabled="!editForm.name || !editForm.version" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50">{{ t('common.update') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>