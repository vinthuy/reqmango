<template>
  <div class="project-detail">
    <!-- 加载状态 -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>

    <!-- 项目内容 -->
    <div v-else-if="project" class="space-y-6">
      <!-- 项目头部 -->
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        <div class="flex items-start justify-between">
          <div class="flex items-center space-x-4">
            <!-- 项目图标 -->
            <div
              class="w-12 h-12 rounded-lg flex items-center justify-center text-white text-xl font-bold"
              :style="{ backgroundColor: project.color || '#6366f1' }"
            >
              {{ project.name?.charAt(0)?.toUpperCase() || 'P' }}
            </div>
            <div>
              <h1 class="text-xl font-semibold text-gray-900">{{ project.name }}</h1>
              <p v-if="project.description" class="text-sm text-gray-500 mt-1">{{ project.description }}</p>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <!-- 归档按钮 -->
            <button
              v-if="!project.is_archived"
              @click="archiveProject"
              class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              归档
            </button>
            <button
              v-else
              @click="restoreProject"
              class="px-3 py-1.5 text-sm text-green-600 border border-green-300 rounded-md hover:bg-green-50"
            >
              恢复
            </button>

            <!-- 设置 -->
            <button
              @click="$emit('settings')"
              class="px-3 py-1.5 text-sm text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50"
            >
              设置
            </button>
          </div>
        </div>

        <!-- 项目状态 -->
        <div class="mt-4 flex items-center space-x-4">
          <span
            class="px-2 py-1 text-xs rounded"
            :class="project.is_archived ? 'bg-yellow-100 text-yellow-700' : 'bg-green-100 text-green-700'"
          >
            {{ project.is_archived ? '已归档' : '活跃' }}
          </span>
          <span class="text-xs text-gray-500">
            创建于 {{ formatDate(project.created_at) }}
          </span>
        </div>
      </div>

      <!-- 统计卡片 -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">总工作项</p>
              <p class="text-2xl font-semibold text-gray-900 mt-1">{{ statistics.total_issues }}</p>
            </div>
            <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">进行中</p>
              <p class="text-2xl font-semibold text-yellow-600 mt-1">{{ statistics.in_progress_issues }}</p>
            </div>
            <div class="w-10 h-10 bg-yellow-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">已完成</p>
              <p class="text-2xl font-semibold text-green-600 mt-1">{{ statistics.completed_issues }}</p>
            </div>
            <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500">成员数</p>
              <p class="text-2xl font-semibold text-purple-600 mt-1">{{ statistics.member_count }}</p>
            </div>
            <div class="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- 快捷入口 -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <button
          @click="$emit('navigate', 'issues')"
          class="bg-white rounded-lg border border-gray-200 p-4 hover:border-indigo-300 hover:shadow-sm transition-all text-left"
        >
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <div>
              <h3 class="font-medium text-gray-900">工作项</h3>
              <p class="text-sm text-gray-500">{{ statistics.total_issues }} 个工作项</p>
            </div>
          </div>
        </button>

        <button
          @click="$emit('navigate', 'cycles')"
          class="bg-white rounded-lg border border-gray-200 p-4 hover:border-indigo-300 hover:shadow-sm transition-all text-left"
        >
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h3 class="font-medium text-gray-900">周期</h3>
              <p class="text-sm text-gray-500">管理迭代</p>
            </div>
          </div>
        </button>

        <button
          @click="$emit('navigate', 'modules')"
          class="bg-white rounded-lg border border-gray-200 p-4 hover:border-indigo-300 hover:shadow-sm transition-all text-left"
        >
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <div>
              <h3 class="font-medium text-gray-900">模块</h3>
              <p class="text-sm text-gray-500">组织工作</p>
            </div>
          </div>
        </button>
      </div>

      <!-- 成员管理 -->
      <ProjectMemberList
        :project-id="projectId"
        :workspace-id="workspaceId"
        @refresh="loadProject"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import projectApi from '@/api/project'
import ProjectMemberList from './ProjectMemberList.vue'
import type { ProjectResponse } from '@/types/project'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
}>()

// Emits
defineEmits<{
  (e: 'settings'): void
  (e: 'navigate', tab: string): void
}>()

// State
const project = ref<ProjectResponse | null>(null)
const statistics = ref({
  total_issues: 0,
  in_progress_issues: 0,
  completed_issues: 0,
  member_count: 0
})
const loading = ref(false)

// Load project
onMounted(() => {
  loadProject()
})

async function loadProject() {
  loading.value = true
  try {
    project.value = await projectApi.getProject(props.projectId, props.workspaceId)
    const stats = await projectApi.getProjectStatistics(props.projectId)
    statistics.value = {
      total_issues: stats.total_issues || 0,
      in_progress_issues: stats.in_progress_issues || 0,
      completed_issues: stats.completed_issues || 0,
      member_count: stats.member_count || 0
    }
  } catch (error) {
    console.error('Failed to load project:', error)
  } finally {
    loading.value = false
  }
}

// Archive project
async function archiveProject() {
  try {
    await projectApi.archiveProject(props.projectId, props.workspaceId)
    await loadProject()
  } catch (error) {
    console.error('Failed to archive project:', error)
  }
}

// Restore project
async function restoreProject() {
  try {
    await projectApi.restoreProject(props.projectId, props.workspaceId)
    await loadProject()
  } catch (error) {
    console.error('Failed to restore project:', error)
  }
}

// Format date
function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()}`
}
</script>

<style scoped>
.project-detail {
  @apply space-y-4;
}
</style>