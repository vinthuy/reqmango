<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl border-l border-gray-200 z-50 overflow-y-auto">
      <div class="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between z-10">
        <h3 class="text-lg font-semibold text-gray-900 truncate">{{ module?.name }}</h3>
        <div class="flex items-center space-x-1">
          <button @click="module && $emit('edit', module)" class="px-2 py-1 text-xs border border-gray-300 text-gray-600 rounded hover:bg-gray-50">编辑</button>
          <button @click="handleDelete" class="px-2 py-1 text-xs border border-red-300 text-red-600 rounded hover:bg-red-50">删除</button>
          <button @click="$emit('close')" class="p-1 text-gray-400 hover:text-gray-600 ml-1">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <svg class="animate-spin h-6 w-6 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
      </div>

      <div v-else-if="module" class="p-4 space-y-4">
        <p v-if="module.description" class="text-sm text-gray-500">{{ module.description }}</p>

        <div v-if="moduleStore.progress" class="grid grid-cols-3 gap-3">
          <div class="text-center p-3 bg-gray-50 rounded">
            <div class="text-xl font-bold text-gray-900">{{ moduleStore.progress.total_issues }}</div>
            <div class="text-xs text-gray-500">总数</div>
          </div>
          <div class="text-center p-3 bg-gray-50 rounded">
            <div class="text-xl font-bold text-green-600">{{ moduleStore.progress.completed_issues }}</div>
            <div class="text-xs text-gray-500">完成</div>
          </div>
          <div class="text-center p-3 bg-gray-50 rounded">
            <div class="text-xl font-bold text-indigo-600">{{ moduleStore.progress.progress }}%</div>
            <div class="text-xs text-gray-500">进度</div>
          </div>
        </div>

        <div>
          <div class="flex items-center justify-between mb-2">
            <h4 class="text-sm font-medium text-gray-700">工作项 ({{ moduleStore.moduleIssues.length }})</h4>
            <button @click="toggleAddIssue" class="px-2 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700">+ 添加</button>
          </div>
          <div v-if="showAddIssue" class="mb-3 border border-gray-200 rounded-md p-2">
            <input v-model="searchQuery" type="text" placeholder="搜索工作项..."
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded mb-2 focus:outline-none focus:ring-1 focus:ring-indigo-500" @input="searchIssues" />
            <div class="max-h-40 overflow-y-auto space-y-1">
              <div v-for="issue in availableIssues" :key="issue.id" @click="handleAddIssue(issue.id)" class="flex items-center p-1.5 hover:bg-indigo-50 rounded cursor-pointer text-sm">
                <span class="text-gray-900 truncate flex-1">{{ issue.name }}</span>
                <span class="text-xs text-gray-400 ml-2">#{{ issue.sequence_id }}</span>
              </div>
              <div v-if="availableIssues.length === 0 && searched" class="text-xs text-gray-400 py-2 text-center">没有可添加的工作项</div>
            </div>
          </div>
          <div v-if="moduleStore.moduleIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">暂无工作项</div>
          <div v-else class="space-y-2">
            <div v-for="issue in moduleStore.moduleIssues" :key="issue.id" class="flex items-center justify-between p-2 bg-gray-50 rounded text-sm">
              <span class="text-gray-900 truncate flex-1">{{ issue.name }}</span>
              <button @click="handleRemoveIssue(issue.id)" class="ml-2 text-gray-400 hover:text-red-500" title="移除">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useModuleStore } from '@/stores/module'
import { issueApi } from '@/api/issue'
import type { ModuleResponse } from '@/types/module'

const props = defineProps<{
  module: ModuleResponse | null
  visible: boolean
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{ close: []; edit: [module: ModuleResponse] }>()

const moduleStore = useModuleStore()
const loading = computed(() => moduleStore.isLoading)
const showAddIssue = ref(false)
const searchQuery = ref('')
const searched = ref(false)
const availableIssues = ref<any[]>([])

watch(() => props.visible, async (v) => {
  if (v && props.module) {
    showAddIssue.value = false
    await Promise.all([moduleStore.fetchProgress(props.module.id), moduleStore.fetchModuleIssues(props.module.id)])
  }
})

function toggleAddIssue() {
  showAddIssue.value = !showAddIssue.value
  if (showAddIssue.value) { searchQuery.value = ''; searchIssues() }
}

async function searchIssues() {
  if (!props.module) return
  try {
    const result = await issueApi.listIssues(props.projectId, props.workspaceId, { search: searchQuery.value || undefined } as any)
    const allIssues = (result as any)?.items || result || []
    const currentIds = new Set(moduleStore.moduleIssues.map((i: any) => i.id))
    availableIssues.value = allIssues.filter((i: any) => !currentIds.has(i.id) && i.state_group !== 'completed' && i.state_group !== 'cancelled')
    searched.value = true
  } catch { availableIssues.value = []; searched.value = true }
}

async function handleAddIssue(issueId: number) {
  if (!props.module) return
  await moduleStore.addIssueToModule(props.module.id, issueId)
  showAddIssue.value = false
}

async function handleRemoveIssue(issueId: number) {
  if (!props.module) return
  await moduleStore.removeIssueFromModule(props.module.id, issueId)
}

async function handleDelete() {
  if (!props.module) return
  if (!confirm(`确定要删除模块 "${props.module.name}" 吗？此操作不可撤销。`)) return
  await moduleStore.deleteModuleAction(props.module.id)
  emit('close')
}
</script>

<style scoped>
.slide-enter-active, .slide-leave-active { transition: transform 0.3s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(100%); }
</style>
