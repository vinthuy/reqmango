<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import type { Workspace, ProjectCreate } from '@/types'
import type { ProjectResponse } from '@/types/project'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const workspace = ref<Workspace | null>(null)
const projects = ref<ProjectResponse[]>([])
const showCreateModal = ref(false)
const newProject = ref<ProjectCreate>({ name: '', identifier: '', description: '' })
const loading = ref(false)
const createLoading = ref(false)
const error = ref('')

const logout = () => {
  authStore.logout()
  router.push('/login')
}

const goToSettings = () => {
  router.push(`/workspace/${route.params.slug}/settings`)
}

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
  
  if (!newProject.value.name) {
    error.value = '请输入项目名称'
    return
  }
  
  if (!newProject.value.identifier) {
    error.value = '请输入项目标识符'
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
    error.value = err.response?.data?.detail || '创建项目失败'
    console.error('Failed to create project:', err)
  } finally {
    createLoading.value = false
  }
}

onMounted(fetchWorkspace)
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <a href="/" class="text-gray-600 hover:text-gray-800">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
            </svg>
          </a>
          <h1 class="text-xl font-bold text-gray-800">
            {{ workspace?.name || 'Workspace' }}
          </h1>
        </div>
        <div class="flex items-center gap-4">
          <span class="text-gray-600">{{ authStore.user?.email }}</span>
          <button
            @click="goToSettings"
            class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
          >
            设置
          </button>
          <button
            @click="logout"
            class="px-4 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-lg transition"
          >
            退出登录
          </button>
        </div>
      </div>
    </header>
    
    <main class="max-w-4xl mx-auto px-6 py-8">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
      
      <template v-else>
        <div class="flex items-center justify-between mb-6">
          <div>
            <h2 class="text-lg font-semibold text-gray-800">项目</h2>
            <p class="text-sm text-gray-500 mt-1">{{ projects.length }} 个项目</p>
          </div>
          <button
            @click="showCreateModal = true"
            class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition"
          >
            创建项目
          </button>
        </div>
        
        <div v-if="projects.length === 0" class="text-center py-12">
          <div class="text-gray-400 mb-4">暂无项目</div>
          <button
            @click="showCreateModal = true"
            class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition"
          >
            创建第一个项目
          </button>
        </div>
        
        <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="project in projects"
            :key="project.id"
            class="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-md transition cursor-pointer"
            @click="goToProject(project.id)"
          >
            <div class="flex items-center gap-2">
              <span class="px-2 py-1 bg-gray-100 text-gray-600 text-xs font-medium rounded">{{ project.identifier }}</span>
              <h3 class="font-semibold text-gray-800">{{ project.name }}</h3>
            </div>
            <p v-if="project.description" class="text-sm text-gray-500 mt-2">{{ project.description }}</p>
          </div>
        </div>
      </template>
    </main>
    
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">创建项目</h3>
        
        <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-600 text-sm">
          {{ error }}
        </div>
        
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">名称</label>
          <input
            v-model="newProject.name"
            type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
            placeholder="项目名称"
          />
        </div>
        
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">标识符</label>
          <input
            v-model="newProject.identifier"
            type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition uppercase"
            placeholder="PROJ"
          />
        </div>
        
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 mb-2">描述</label>
          <textarea
            v-model="newProject.description"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
            rows="3"
            placeholder="项目描述"
          ></textarea>
        </div>
        
        <div class="flex gap-3">
          <button
            @click="showCreateModal = false"
            class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition"
          >
            取消
          </button>
          <button
            @click="createProject"
            :disabled="createLoading"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
          >
            {{ createLoading ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>