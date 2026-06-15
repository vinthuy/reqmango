<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { workspaceApi } from '@/api/workspace'
import type { Workspace, WorkspaceCreate } from '@/types'

const authStore = useAuthStore()
const workspaces = ref<Workspace[]>([])
const showCreateModal = ref(false)
const newWorkspace = ref<WorkspaceCreate>({ name: '', slug: '' })
const loading = ref(false)
const createLoading = ref(false)
const error = ref('')

const logout = () => {
  authStore.logout()
}

const fetchWorkspaces = async () => {
  loading.value = true
  try {
    workspaces.value = await workspaceApi.list()
  } catch (err) {
    console.error('Failed to fetch workspaces:', err)
    workspaces.value = []
  } finally {
    loading.value = false
  }
}

const validateSlug = (slug: string): boolean => {
  const pattern = /^[a-z0-9-]+$/
  return pattern.test(slug)
}

const createWorkspace = async () => {
  error.value = ''
  
  if (!newWorkspace.value.name) {
    error.value = '请输入工作空间名称'
    return
  }
  
  if (!newWorkspace.value.slug) {
    error.value = '请输入Slug'
    return
  }
  
  if (!validateSlug(newWorkspace.value.slug)) {
    error.value = 'Slug只能包含小写字母、数字和连字符'
    return
  }
  
  createLoading.value = true
  try {
    const workspace = await workspaceApi.create(newWorkspace.value)
    workspaces.value.push(workspace)
    showCreateModal.value = false
    newWorkspace.value = { name: '', slug: '' }
  } catch (err: any) {
    error.value = err.response?.data?.detail || '创建工作空间失败，请检查信息是否正确'
    console.error('Failed to create workspace:', err)
  } finally {
    createLoading.value = false
  }
}

fetchWorkspaces()
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <h1 class="text-xl font-bold text-gray-800">Reqman AI</h1>
        </div>
        <div class="flex items-center gap-4">
          <span class="text-gray-600">{{ authStore.user?.email }}</span>
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
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-lg font-semibold text-gray-800">工作空间</h2>
        <button
          @click="showCreateModal = true"
          class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition"
        >
          创建工作空间
        </button>
      </div>
      
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
      
      <div v-else-if="workspaces.length === 0" class="text-center py-12">
        <div class="text-gray-400 mb-4">暂无工作空间</div>
        <button
          @click="showCreateModal = true"
          class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition"
        >
          创建第一个工作空间
        </button>
      </div>
      
      <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="workspace in workspaces"
          :key="workspace.id"
          class="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-md transition cursor-pointer"
          @click="() => {}"
        >
          <h3 class="font-semibold text-gray-800">{{ workspace.name }}</h3>
          <p class="text-sm text-gray-500 mt-1">{{ workspace.slug }}</p>
        </div>
      </div>
    </main>
    
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">创建工作空间</h3>
        
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">名称</label>
          <input
            v-model="newWorkspace.name"
            type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
            placeholder="工作空间名称"
          />
        </div>
        
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 mb-2">Slug</label>
          <input
            v-model="newWorkspace.slug"
            type="text"
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition"
            placeholder="url-slug"
          />
        </div>
        
        <div class="flex gap-3">
          <button
            @click="showCreateModal = false"
            class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition"
          >
            取消
          </button>
          <button
            @click="createWorkspace"
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
          >
            创建
          </button>
        </div>
      </div>
    </div>
  </div>
</template>