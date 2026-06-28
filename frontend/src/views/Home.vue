<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { workspaceApi } from '@/api/workspace'
import type { Workspace, WorkspaceCreate } from '@/types'

const router = useRouter()
const { t } = useI18n()
const workspaces = ref<Workspace[]>([])
const showCreateModal = ref(false)
const newWorkspace = ref<WorkspaceCreate>({ name: '', slug: '' })
const loading = ref(false)
const createLoading = ref(false)
const error = ref('')

const fetchWorkspaces = async () => {
  loading.value = true
  try {
    workspaces.value = await workspaceApi.list()
  } catch (err) {
    workspaces.value = []
  } finally {
    loading.value = false
  }
}

const goToWorkspace = (slug: string) => router.push(`/workspace/${slug}`)

const validateSlug = (slug: string): boolean => /^[a-z0-9-]+$/.test(slug)

const createWorkspace = async () => {
  error.value = ''
  if (!newWorkspace.value.name) { error.value = t('home.nameRequired'); return }
  if (!newWorkspace.value.slug) { error.value = t('home.slugRequired'); return }
  if (!validateSlug(newWorkspace.value.slug)) { error.value = t('home.slugInvalid'); return }
  createLoading.value = true
  try {
    const workspace = await workspaceApi.create(newWorkspace.value)
    workspaces.value.push(workspace)
    showCreateModal.value = false
    newWorkspace.value = { name: '', slug: '' }
  } catch (err: any) {
    error.value = err.response?.data?.detail || t('home.createFailed')
  } finally {
    createLoading.value = false
  }
}

fetchWorkspaces()
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">{{ t('home.title') }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{{ t('home.workspaceCount', { count: workspaces.length }) }}</p>
      </div>
      <button
        @click="showCreateModal = true"
        class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        {{ t('home.createWorkspace') }}
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="animate-spin rounded-full h-8 w-8 border-2 border-indigo-600 border-t-transparent"></div>
    </div>

    <div v-else-if="workspaces.length === 0" class="text-center py-20">
      <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
        <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      </div>
      <h3 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-1">{{ t('home.noWorkspaces') }}</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ t('home.noWorkspacesHint') }}</p>
      <button @click="showCreateModal = true" class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">{{ t('home.createFirst') }}</button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="ws in workspaces"
        :key="ws.id"
        class="group bg-white dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700/60 rounded-xl p-5 hover:border-indigo-300 dark:hover:border-indigo-500/40 hover:shadow-md transition-all cursor-pointer"
        @click="goToWorkspace(ws.slug)"
      >
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-9 h-9 rounded-lg bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center shrink-0">
              <span class="text-indigo-600 dark:text-indigo-400 text-sm font-bold">{{ ws.name.charAt(0).toUpperCase() }}</span>
            </div>
            <div class="min-w-0">
              <h3 class="font-semibold text-gray-900 dark:text-gray-100 truncate text-[15px]">{{ ws.name }}</h3>
              <span class="text-xs text-gray-400 dark:text-gray-500 font-mono">{{ ws.slug }}</span>
            </div>
          </div>
          <svg class="w-4 h-4 text-gray-300 dark:text-gray-600 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </div>
      </div>
    </div>

    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-5">{{ t('home.createWorkspace') }}</h3>
        <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-600 dark:text-red-400 text-sm">{{ error }}</div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{{ t('home.name') }}</label>
            <input v-model="newWorkspace.name" type="text" class="w-full px-3.5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm" :placeholder="t('home.namePlaceholder')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{{ t('home.slug') }}</label>
            <input v-model="newWorkspace.slug" type="text" class="w-full px-3.5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm" placeholder="url-slug" />
          </div>
        </div>
        <div class="flex gap-3 mt-6">
          <button @click="showCreateModal = false" class="flex-1 px-4 py-2.5 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition text-sm font-medium">{{ t('home.cancel') }}</button>
          <button @click="createWorkspace" :disabled="createLoading" class="flex-1 px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium">{{ createLoading ? t('home.creating') : t('home.create') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>
