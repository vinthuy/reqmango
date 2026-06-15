<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { Project, ProjectCreate } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const workspaceSlug = ref('')
const projects = ref<Project[]>([])
const showCreateModal = ref(false)
const newProject = ref<ProjectCreate>({ name: '', identifier: '' })

const logout = () => {
  authStore.logout()
}

const fetchWorkspace = async () => {
  workspaceSlug.value = route.params.slug as string
}

const createProject = async () => {
  if (!newProject.value.name || !newProject.value.identifier) return
  showCreateModal.value = false
  newProject.value = { name: '', identifier: '' }
}

fetchWorkspace()
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
          <h1 class="text-xl font-bold text-gray-800">Reqman AI - {{ workspaceSlug }}</h1>
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
        <h2 class="text-lg font-semibold text-gray-800">项目</h2>
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
          @click="() => {}"
        >
          <div class="flex items-center gap-2">
            <span class="px-2 py-1 bg-gray-100 text-gray-600 text-xs font-medium rounded">{{ project.identifier }}</span>
            <h3 class="font-semibold text-gray-800">{{ project.name }}</h3>
          </div>
          <p v-if="project.description" class="text-sm text-gray-500 mt-2">{{ project.description }}</p>
        </div>
      </div>
    </main>
    
    <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl shadow-lg p-6 w-full max-w-md mx-4">
        <h3 class="text-lg font-semibold text-gray-800 mb-4">创建项目</h3>
        
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
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
          >
            创建
          </button>
        </div>
      </div>
    </div>
  </div>
</template>