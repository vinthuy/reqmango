<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import templateApi from '@/api/template'
import type { Workspace, ProjectCreate } from '@/types'
import type { ProjectResponse } from '@/types/project'
import type { ProjectTemplate } from '@/types/template'
import AICopilot from '@/components/AICopilot.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const workspace = ref<Workspace | null>(null)
const projects = ref<ProjectResponse[]>([])
const showCreateModal = ref(false)
const newProject = ref<ProjectCreate>({ name: '', identifier: '', description: '' })
const loading = ref(false)
const createLoading = ref(false)
const error = ref('')
const templates = ref<ProjectTemplate[]>([])
const templatesLoading = ref(false)

const fetchWorkspace = async () => {
  loading.value = true
  const slug = route.params.slug as string
  try {
    workspace.value = await workspaceApi.getBySlug(slug)
    await fetchProjects()
  } catch (err) {
    console.error('Failed to fetch workspace:', err)
  } finally {
    loading.value = false
  }
}

const fetchTemplates = async () => {
  if (!workspace.value) return
  templatesLoading.value = true
  try {
    templates.value = await templateApi.listTemplates(workspace.value.id)
  } catch (err) {
    console.error('Failed to fetch templates:', err)
    templates.value = []
  } finally {
    templatesLoading.value = false
  }
}

watch(showCreateModal, (val) => {
  if (val) {
    fetchTemplates()
  }
})

const fetchProjects = async () => {
  if (!workspace.value) return
  try {
    projects.value = await projectApi.listProjects(workspace.value.id)
  } catch (err) {
    console.error('Failed to fetch projects:', err)
    projects.value = []
  }
}

const goToProject = (projectId: number) => {
  router.push(`/workspace/${workspace.value?.slug}/project/${projectId}`)
}

const createProject = async () => {
  error.value = ''
  if (!newProject.value.name || !newProject.value.identifier) {
    error.value = !newProject.value.name ? t('workspace.nameRequiredMsg') : t('workspace.slugRequiredMsg')
    return
  }
  if (!workspace.value) return
  createLoading.value = true
  try {
    const project = await projectApi.createProject(workspace.value.id, newProject.value)
    projects.value.push(project)
    showCreateModal.value = false
    newProject.value = { name: '', identifier: '', description: '' }
  } catch (err: any) {
    error.value = err.response?.data?.detail || t('workspace.createFailedGeneric')
  } finally {
    createLoading.value = false
  }
}

const showAIChat = ref(false)

function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'j') {
    e.preventDefault()
    showAIChat.value = !showAIChat.value
  }
}

onMounted(() => {
  fetchWorkspace()
  document.addEventListener('keydown', handleKeydown)
})
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div class="p-6">
    <!-- Top bar -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">{{ workspace?.name || 'Workspace' }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          {{ t('workspace.projectCount', { count: projects.length }) }} ·
          <router-link :to="`/workspace/${workspace?.slug}/overview`" class="text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300 font-medium">
            {{ t('workspace.overview') }}
          </router-link>
        </p>
      </div>
      <button
        @click="showCreateModal = true"
        class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        {{ t('workspace.createProject') }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <div class="animate-spin rounded-full h-8 w-8 border-2 border-indigo-600 border-t-transparent"></div>
    </div>

    <!-- Empty -->
    <div v-else-if="projects.length === 0" class="text-center py-20">
      <div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
        <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      </div>
      <h3 class="text-base font-medium text-gray-900 dark:text-gray-100 mb-1">{{ t('workspace.noProjects') }}</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{{ t('workspace.noProjectsHint') }}</p>
      <button
        @click="showCreateModal = true"
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
      >
        {{ t('workspace.createFirst') }}
      </button>
    </div>

    <!-- Project Grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="project in projects"
        :key="project.id"
        class="group bg-white dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700/60 rounded-xl p-5 hover:border-indigo-300 dark:hover:border-indigo-500/40 hover:shadow-md transition-all cursor-pointer"
        @click="goToProject(project.id)"
      >
        <div class="flex items-start justify-between mb-3">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 text-white text-sm font-bold" 
                 :style="{ backgroundColor: project.color || '#6366f1' }">
              {{ project.name.charAt(0).toUpperCase() }}
            </div>
            <div class="min-w-0">
              <h3 class="font-semibold text-gray-900 dark:text-gray-100 truncate text-[15px]">{{ project.name }}</h3>
              <span class="text-xs text-gray-400 dark:text-gray-500 font-mono">{{ project.identifier }}</span>
            </div>
          </div>
          <svg class="w-4 h-4 text-gray-300 dark:text-gray-600 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </div>
        <p v-if="project.description" class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2">{{ project.description }}</p>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50" @click.self="showCreateModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-5">{{ t('workspace.createProject') }}</h3>
        <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-600 dark:text-red-400 text-sm">{{ error }}</div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{{ t('workspace.name') }}</label>
            <input v-model="newProject.name" type="text" class="w-full px-3.5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm" :placeholder="t('workspace.name')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{{ t('auth.identifier', '标识符') }}</label>
            <input v-model="newProject.identifier" type="text" class="w-full px-3.5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm uppercase" :placeholder="t('auth.identifierHint', 'PROJ')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{{ t('common.description', '描述') }}</label>
            <textarea v-model="newProject.description" class="w-full px-3.5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm" rows="3" :placeholder="t('common.descriptionHint', '项目描述')" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">{{ t('settings.templates', '模板') }}</label>
            <select v-model.number="newProject.template_id" :disabled="templatesLoading" class="w-full px-3.5 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition text-sm">
              <option :value="null">{{ t('common.none', '无') }}</option>
              <option v-for="template in templates" :key="template.id" :value="template.id">
                {{ template.name }}
                <span v-if="template.is_default" class="text-indigo-500">*</span>
              </option>
            </select>
          </div>
        </div>
        <div class="flex gap-3 mt-6">
          <button @click="showCreateModal = false" class="flex-1 px-4 py-2.5 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition text-sm font-medium">{{ t('workspace.cancel') }}</button>
          <button @click="createProject" :disabled="createLoading" class="flex-1 px-4 py-2.5 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition text-sm font-medium">{{ createLoading ? t('workspace.creating') : t('workspace.create') }}</button>
        </div>
      </div>
    </div>

    <!-- AI Copilot -->
    <AICopilot
      :visible="showAIChat"
      :project-id="0"
      :workspace-id="workspace?.id || 0"
      :project-name="workspace?.name || ''"
      @close="showAIChat = false"
    />

    <!-- FAB: AI Copilot -->
    <button
      @click="showAIChat = !showAIChat"
      class="fixed bottom-6 right-6 w-12 h-12 rounded-full bg-indigo-600 text-white shadow-lg hover:bg-indigo-700 transition-colors flex items-center justify-center text-xl z-30"
      :title="'AI (Ctrl+J)'"
    >🤖</button>
  </div>
</template>
