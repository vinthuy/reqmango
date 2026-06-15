<template>
  <div class="issue-list">
    <!-- 头部工具栏 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <!-- 状态筛选 -->
          <select
            v-model="filters.state_id"
            class="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">所有状态</option>
            <option v-for="state in states" :key="state.id" :value="state.id">
              {{ state.name }}
            </option>
          </select>

          <!-- 优先级筛选 -->
          <select
            v-model="filters.priority"
            class="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">所有优先级</option>
            <option value="urgent">紧急</option>
            <option value="high">高</option>
            <option value="medium">中</option>
            <option value="low">低</option>
            <option value="none">无</option>
          </select>

          <!-- 周期筛选 -->
          <select
            v-model="filters.cycle_id"
            class="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">所有周期</option>
            <option v-for="cycle in cycles" :key="cycle.id" :value="cycle.id">
              {{ cycle.name }}
            </option>
          </select>
        </div>

        <div class="flex items-center space-x-3">
          <!-- 搜索框 -->
          <div class="relative">
            <input
              v-model="filters.search"
              type="text"
              placeholder="搜索工作项..."
              class="pl-8 pr-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 w-64"
            />
            <svg class="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 transform -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>

          <!-- 新建按钮 -->
          <button
            @click="$emit('create')"
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>新建</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 列表内容 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="issues.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <p class="mt-2 text-gray-500">暂无工作项</p>
        <button @click="$emit('create')" class="mt-3 text-indigo-600 hover:text-indigo-800 text-sm">
          创建第一个工作项
        </button>
      </div>

      <!-- 工作项列表 -->
      <div v-else class="space-y-2">
        <IssueCard
          v-for="issue in issues"
          :key="issue.id"
          :issue="issue"
          @click="$emit('select', issue)"
          @archive="$emit('archive', issue)"
          @delete="$emit('delete', issue)"
        />
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="mt-4 flex justify-center">
        <nav class="flex items-center space-x-1">
          <button
            @click="page--"
            :disabled="page <= 1"
            class="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            上一页
          </button>
          <span class="px-3 py-1 text-sm text-gray-600">
            第 {{ page }} / {{ totalPages }} 页
          </span>
          <button
            @click="page++"
            :disabled="page >= totalPages"
            class="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            下一页
          </button>
        </nav>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import IssueCard from './IssueCard.vue'
import issueApi from '@/api/issue'

// Props
const props = defineProps<{
  projectId: string
  workspaceId: string
}>()

// Emits
defineEmits<{
  (e: 'create'): void
  (e: 'select', issue: any): void
  (e: 'archive', issue: any): void
  (e: 'delete', issue: any): void
}>()

// State
const issues = ref<any[]>([])
const states = ref<any[]>([])
const cycles = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const limit = ref(20)
const totalPages = ref(1)

const filters = ref({
  state_id: '',
  priority: '',
  cycle_id: '',
  search: ''
})

// Load data
onMounted(async () => {
  await Promise.all([
    loadIssues(),
    loadStates(),
    loadCycles()
  ])
})

// Watch filters and reload
watch(filters, () => {
  page.value = 1
  loadIssues()
}, { deep: true })

watch(page, () => {
  loadIssues()
})

// Load issues
async function loadIssues() {
  loading.value = true
  try {
    const params: any = {
      limit: limit.value,
      offset: (page.value - 1) * limit.value
    }

    if (filters.value.state_id) params.state_id = filters.value.state_id
    if (filters.value.priority) params.priority = filters.value.priority
    if (filters.value.cycle_id) params.cycle_id = filters.value.cycle_id
    if (filters.value.search) params.search = filters.value.search

    const response = await issueApi.listIssues(props.projectId, params)
    issues.value = response.items || response
    totalPages.value = Math.ceil((response.total || issues.value.length) / limit.value)
  } catch (error) {
    console.error('Failed to load issues:', error)
  } finally {
    loading.value = false
  }
}

// Load states
async function loadStates() {
  try {
    const response = await fetch(`/api/v1/projects/${props.projectId}/states`)
    states.value = await response.json()
  } catch (error) {
    console.error('Failed to load states:', error)
  }
}

// Load cycles
async function loadCycles() {
  try {
    const response = await fetch(`/api/v1/projects/${props.projectId}/cycles`)
    cycles.value = await response.json()
  } catch (error) {
    console.error('Failed to load cycles:', error)
  }
}
</script>

<style scoped>
.issue-list {
  @apply bg-white rounded-lg;
}
</style>