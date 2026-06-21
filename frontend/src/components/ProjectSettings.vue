<template>
  <div class="flex h-screen bg-gray-50">
    <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center text-white text-lg font-bold"
            :style="{ backgroundColor: project?.color || '#6366f1' }"
          >
            {{ project?.name?.charAt(0)?.toUpperCase() || 'P' }}
          </div>
          <div>
            <h2 class="font-semibold text-gray-800">{{ project?.name }}</h2>
            <p class="text-xs text-gray-500">{{ project?.identifier }}</p>
          </div>
        </div>
      </div>

      <nav class="flex-1 p-2 overflow-y-auto">
        <button
          v-for="item in menuItems"
          :key="item.id"
          @click="activeMenu = item.id"
          class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors"
          :class="activeMenu === item.id ? 'bg-indigo-50 text-indigo-600' : 'text-gray-600 hover:bg-gray-50'"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="item.iconPath"></svg>
          {{ item.label }}
        </button>
      </nav>
    </aside>

    <main class="flex-1 overflow-auto">
      <header class="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
        <h1 class="text-xl font-semibold text-gray-800">{{ currentMenuLabel }}</h1>
        <button @click="goBack" class="px-4 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
          </svg>
          Back
        </button>
      </header>

      <div class="p-6">
        <div v-if="activeMenu === 'overview'" class="bg-white rounded-lg border border-gray-200">
          <div class="p-6">
            <h2 class="text-lg font-semibold text-gray-800 mb-4">Project Overview</h2>
            <div class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div class="p-4 bg-gray-50 rounded-lg">
                  <p class="text-sm text-gray-500">Total Issues</p>
                  <p class="text-2xl font-bold text-gray-800">{{ stats.issuesCount || 0 }}</p>
                </div>
                <div class="p-4 bg-gray-50 rounded-lg">
                  <p class="text-sm text-gray-500">Members</p>
                  <p class="text-2xl font-bold text-gray-800">{{ stats.membersCount || 0 }}</p>
                </div>
              </div>
              <div class="pt-4">
                <label class="block text-sm font-medium text-gray-700 mb-2">Description</label>
                <p class="text-gray-600">{{ project?.description || 'No description' }}</p>
              </div>
            </div>
          </div>
          <div class="px-6 py-4 border-t border-gray-200">
            <button @click="showEditModal = true" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 transition">
              Edit Project
            </button>
          </div>
        </div>

        <div v-else-if="activeMenu === 'members'" class="bg-white rounded-lg border border-gray-200">
          <ProjectMemberList :project-id="projectId" :workspace-id="workspaceId" />
        </div>

        <div v-else-if="activeMenu === 'modules'" class="bg-white rounded-lg border border-gray-200">
          <ModuleList :project-id="projectId" :workspace-id="workspaceId" @create="showModuleModal = true" @select="selectModule" />
        </div>

        <div v-else-if="activeMenu === 'cycles'" class="bg-white rounded-lg border border-gray-200">
          <CycleList :project-id="projectId" :workspace-id="workspaceId" @create="showCycleModal = true" @select="selectCycle" />
        </div>

        <div v-else-if="activeMenu === 'custom-fields'" class="bg-white rounded-lg border border-gray-200">
          <CustomFieldList :project-id="projectId" @create="showFieldForm = true" />
        </div>

        <div v-else-if="activeMenu === 'estimate-points'" class="bg-white rounded-lg border border-gray-200">
          <EstimatePointManager :project-id="projectId" @create="showEstimateForm = true" />
        </div>

        <div v-else-if="activeMenu === 'workflow'" class="bg-white rounded-lg border border-gray-200">
          <WorkflowManager :project-id="projectId" />
        </div>

        <div v-else-if="activeMenu === 'automation'" class="bg-white rounded-lg border border-gray-200">
          <AutomationManager :project-id="projectId" />
        </div>

        <div v-else-if="activeMenu === 'delete'" class="bg-white rounded-lg border border-gray-200 p-6">
          <h2 class="text-lg font-semibold text-gray-800 mb-4">Delete Project</h2>
          <div class="bg-red-50 border border-red-200 rounded-lg p-4">
            <div class="flex items-start gap-3">
              <svg class="h-6 w-6 text-red-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <div>
                <h3 class="font-medium text-red-800">Delete Project</h3>
                <p class="text-sm text-red-700 mt-1">This action will permanently delete the project and all related data. This action cannot be undone.</p>
              </div>
            </div>
            <button @click="showDeleteConfirm = true" class="mt-4 w-full px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 transition">
              Delete Project
            </button>
          </div>
        </div>
      </div>
    </main>

    <div v-if="showEditModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-800">Edit Project</h3>
          <button @click="showEditModal = false" class="text-gray-400 hover:text-gray-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div v-if="settingsError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">{{ settingsError }}</div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input v-model="editForm.name" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="Project name" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Identifier</label>
            <input v-model="editForm.identifier" type="text" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 uppercase" placeholder="PROJ" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea v-model="editForm.description" rows="3" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="Project description"></textarea>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
            <div class="flex items-center gap-2">
              <input v-model="editForm.color" type="color" class="w-10 h-10 rounded border border-gray-300 cursor-pointer" />
              <input v-model="editForm.color" type="text" class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="#6366f1" />
            </div>
          </div>
        </div>
        <div class="flex gap-3 mt-6">
          <button @click="showEditModal = false" class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition">Cancel</button>
          <button @click="saveProject" :disabled="settingsLoading" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition">
            {{ settingsLoading ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
            <svg class="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <h3 class="text-lg font-semibold text-gray-800">Confirm Delete</h3>
            <p class="text-sm text-gray-500">This action cannot be undone</p>
          </div>
        </div>
        <p class="text-gray-600 mb-6">Are you sure you want to delete <strong>{{ project?.name }}</strong>? This will delete all related issues, comments, and attachments.</p>
        <div class="flex gap-3">
          <button @click="showDeleteConfirm = false" class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition">Cancel</button>
          <button @click="deleteProject" :disabled="deleteLoading" class="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 transition">
            {{ deleteLoading ? 'Deleting...' : 'Delete Project' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import projectApi from '@/api/project'
import type { ProjectResponse, ProjectUpdate } from '@/types/project'
import ProjectMemberList from './ProjectMemberList.vue'
import ModuleList from './ModuleList.vue'
import CycleList from './CycleList.vue'
import CustomFieldList from './CustomFieldList.vue'
import EstimatePointManager from './EstimatePointManager.vue'
import WorkflowRuleList from './WorkflowRuleList.vue'
import WorkflowManager from './WorkflowManager.vue'
import AutomationManager from './AutomationManager.vue'

const route = useRoute()
const router = useRouter()

const projectId = ref(Number(route.params.id))
const workspaceId = ref(Number(route.params.workspaceId) || 0)
const project = ref<ProjectResponse | null>(null)

const activeMenu = ref('overview')
const showEditModal = ref(false)
const showDeleteConfirm = ref(false)
const settingsLoading = ref(false)
const deleteLoading = ref(false)
const settingsError = ref('')

const editForm = reactive({ name: '', identifier: '', description: '', color: '#6366f1' })
const stats = ref({ issuesCount: 0, membersCount: 0 })

const menuItems = [
  { id: 'overview', label: '项目概况', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />' },
  { id: 'members', label: '成员管理', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />' },
  { id: 'modules', label: '模块', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />' },
  { id: 'cycles', label: '周期', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />' },
  { id: 'custom-fields', label: '自定义字段', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />' },
  { id: 'workflow', label: '工作流', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />' },
  { id: 'estimate-points', label: '估算点', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />' },
  { id: 'automation', label: '自动化', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />' },
  { id: 'delete', label: '删除项目', iconPath: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />' }
]

const currentMenuLabel = computed(() => {
  const item = menuItems.find(i => i.id === activeMenu.value)
  return item?.label || ''
})

onMounted(() => {
  loadProject()
})

async function loadProject() {
  try {
    project.value = await projectApi.getProject(projectId.value)
    editForm.name = project.value.name
    editForm.identifier = project.value.identifier || ''
    editForm.description = project.value.description || ''
    editForm.color = project.value.color || '#6366f1'
    await loadStats()
  } catch (error) {
    console.error('Failed to load project:', error)
  }
}

async function loadStats() {
  try {
    const members = await projectApi.listProjectMembers(projectId.value)
    stats.value.membersCount = members.length
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

async function saveProject() {
  if (!editForm.name) {
    settingsError.value = 'Please enter project name'
    return
  }
  settingsLoading.value = true
  settingsError.value = ''
  try {
    const data: ProjectUpdate = {
      name: editForm.name,
      identifier: editForm.identifier || undefined,
      description: editForm.description || undefined,
      color: editForm.color
    }
    const updated = await projectApi.updateProject(projectId.value, data)
    project.value = { ...project.value, ...updated } as ProjectResponse
    showEditModal.value = false
  } catch (error: any) {
    settingsError.value = error.response?.data?.message || 'Save failed'
  } finally {
    settingsLoading.value = false
  }
}

async function deleteProject() {
  deleteLoading.value = true
  try {
    await projectApi.deleteProject(projectId.value)
    router.push(`/workspace/${route.params.slug}`)
  } catch (error: any) {
    alert(error.response?.data?.message || 'Delete failed')
    showDeleteConfirm.value = false
  } finally {
    deleteLoading.value = false
  }
}

function goBack() {
  router.push(`/workspace/${route.params.slug}/project/${projectId.value}`)
}

function selectModule(module: any) { console.log('Selected module:', module) }
function selectCycle(cycle: any) { console.log('Selected cycle:', cycle) }
</script>
