<template>
  <div class="min-h-screen bg-gray-50">
    <div class="bg-white border-b border-gray-200 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <button @click="goBack" class="text-gray-500 hover:text-gray-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1 class="text-lg font-semibold text-gray-900">{{ cycle?.name }}</h1>
          <span :class="statusBadgeClass">{{ cycle?.status }}</span>
        </div>
        <div class="flex items-center space-x-2">
          <button v-if="cycle?.status === 'upcoming'" @click="handleStart" class="px-3 py-1.5 bg-green-600 text-white text-sm rounded hover:bg-green-700">开始</button>
          <button v-if="cycle?.status === 'active'" @click="handleEnd" class="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700">结束</button>
          <button v-if="cycle?.status !== 'completed' && cycle?.status !== 'cancelled'" @click="handleCancel" class="px-3 py-1.5 border border-gray-300 text-sm text-gray-600 rounded hover:bg-gray-50">取消</button>
          <button @click="handleDelete" class="px-3 py-1.5 border border-red-300 text-sm text-red-600 rounded hover:bg-red-50">删除</button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="flex justify-center py-20">
      <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
      </svg>
    </div>

    <div v-else-if="cycle" class="max-w-5xl mx-auto px-6 py-6 space-y-6">
      <CycleProgressCard :progress="cycleStore.progress" />
      <CycleBurndownChart :data="cycleStore.burndown" />

      <!-- Issues section -->
      <div class="bg-white rounded-lg border border-gray-200 p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-medium text-gray-700">周期工作项 ({{ cycleStore.cycleIssues.length }})</h3>
          <button
            v-if="cycle?.status === 'active' || cycle?.status === 'upcoming'"
            @click="toggleAddIssue"
            class="px-3 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            + 添加工作项
          </button>
        </div>

        <!-- Add issue search -->
        <div v-if="showAddIssue" class="mb-4 border border-gray-200 rounded-md p-3 bg-gray-50">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索项目内的工作项..."
            class="w-full px-3 py-2 text-sm border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
            @input="searchIssues"
          />
          <div class="mt-2 max-h-60 overflow-y-auto space-y-1">
            <div
              v-for="issue in availableIssues"
              :key="issue.id"
              @click="handleAddIssue(issue.id)"
              class="flex items-center justify-between p-2 hover:bg-indigo-50 rounded cursor-pointer text-sm"
            >
              <div class="flex-1 min-w-0">
                <span class="text-gray-900 truncate block">{{ issue.name }}</span>
                <span class="text-xs text-gray-400">{{ issue.state_name }}</span>
              </div>
              <span class="text-xs text-gray-400 ml-2 shrink-0">#{{ issue.sequence_id }}</span>
            </div>
            <div v-if="availableIssues.length === 0 && searched" class="text-sm text-gray-400 py-4 text-center">
              没有可添加的工作项
            </div>
          </div>
        </div>

        <div v-if="cycleStore.cycleIssues.length === 0" class="text-sm text-gray-400 text-center py-8">暂无工作项</div>
        <div v-else class="space-y-2">
          <div v-for="issue in cycleStore.cycleIssues" :key="issue.id" class="flex items-center justify-between p-2 bg-gray-50 rounded text-sm">
            <span class="text-gray-900">{{ issue.name }}</span>
            <div class="flex items-center space-x-2">
              <span class="text-xs text-gray-400">#{{ issue.sequence_id }}</span>
              <button @click="handleRemoveIssue(issue.id)" class="text-gray-400 hover:text-red-500" title="移除">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCycleStore } from '@/stores/cycle'
import { issueApi } from '@/api/issue'
import CycleProgressCard from '@/components/CycleProgressCard.vue'
import CycleBurndownChart from '@/components/CycleBurndownChart.vue'

const route = useRoute()
const router = useRouter()
const cycleStore = useCycleStore()

const cycleId = Number(route.params.cycleId)
const cycle = computed(() => cycleStore.currentCycle)
const loading = computed(() => cycleStore.isLoading)

const showAddIssue = ref(false)
const searchQuery = ref('')
const searched = ref(false)
const availableIssues = ref<any[]>([])

const statusBadgeClass = computed(() => {
  const map: Record<string, string> = {
    upcoming: 'px-2 py-0.5 text-xs rounded bg-blue-100 text-blue-700',
    active: 'px-2 py-0.5 text-xs rounded bg-green-100 text-green-700',
    completed: 'px-2 py-0.5 text-xs rounded bg-gray-100 text-gray-600',
    cancelled: 'px-2 py-0.5 text-xs rounded bg-red-100 text-red-700',
  }
  return map[cycle.value?.status ?? ''] || ''
})

onMounted(async () => {
  await cycleStore.fetchCycle(cycleId)
  if (cycle.value) {
    await Promise.all([
      cycleStore.fetchProgress(cycleId),
      cycleStore.fetchBurndown(cycleId),
      cycleStore.fetchCycleIssues(cycleId),
    ])
  }
})

function toggleAddIssue() {
  showAddIssue.value = !showAddIssue.value
  if (showAddIssue.value) {
    searchQuery.value = ''
    searchIssues()
  }
}

async function searchIssues() {
  if (!cycle.value) return
  try {
    const result = await issueApi.listIssues(cycle.value.project_id, cycle.value.workspace_id, {
      search: searchQuery.value || undefined,
    })
    const allIssues = (result as any)?.items || result || []
    const currentIds = new Set(cycleStore.cycleIssues.map((i: any) => i.id))
    availableIssues.value = allIssues.filter(
      (i: any) => !currentIds.has(i.id) && i.state_group !== 'completed' && i.state_group !== 'cancelled'
    )
    searched.value = true
  } catch {
    availableIssues.value = []
    searched.value = true
  }
}

async function handleAddIssue(issueId: number) {
  await cycleStore.addIssueToCycle(cycleId, issueId)
  showAddIssue.value = false
  searchQuery.value = ''
  searched.value = false
}

async function handleRemoveIssue(issueId: number) {
  await cycleStore.removeIssueFromCycle(cycleId, issueId)
}

function goBack() { router.back() }

async function handleStart() {
  await cycleStore.startCycle(cycleId)
  await cycleStore.fetchCycle(cycleId)
}
async function handleEnd() {
  if (!confirm('确定要结束这个周期吗？')) return
  await cycleStore.endCycle(cycleId)
  await cycleStore.fetchCycle(cycleId)
}
async function handleCancel() {
  if (!confirm('确定要取消这个周期吗？')) return
  await cycleStore.cancelCycle(cycleId)
  await cycleStore.fetchCycle(cycleId)
}
async function handleDelete() {
  if (!confirm('确定要删除这个周期吗？此操作不可撤销。')) return
  await cycleStore.deleteCycleAction(cycleId)
  router.back()
}
</script>
