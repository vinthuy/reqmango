<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl border-l border-gray-200 z-50 overflow-y-auto">
      <div class="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between z-10">
        <h3 class="text-lg font-semibold text-gray-900 truncate">{{ cycle?.name }}</h3>
        <div class="flex items-center space-x-1">
          <button v-if="cycle?.status === 'upcoming'" @click="handleStart" class="px-2 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700">开始</button>
          <button v-if="cycle?.status === 'active'" @click="handleEnd" class="px-2 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700">结束</button>
          <button v-if="cycle?.status !== 'completed' && cycle?.status !== 'cancelled'" @click="handleCancel" class="px-2 py-1 text-xs border border-gray-300 text-gray-600 rounded hover:bg-gray-50">取消</button>
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

      <div v-else-if="cycle" class="p-4 space-y-4">
        <!-- Status badge and date -->
        <div class="flex items-center justify-between">
          <span :class="statusBadgeClass">{{ cycle.status }}</span>
          <span class="text-xs text-gray-500">{{ cycle.start_date }} {{ cycle.end_date ? '~ ' + cycle.end_date : '' }}</span>
        </div>

        <CycleProgressCard :progress="cycleStore.progress" />
        <CycleBurndownChart :data="cycleStore.burndown" />

        <!-- Issues section -->
        <div>
          <div class="flex items-center justify-between mb-2">
            <h4 class="text-sm font-medium text-gray-700">周期工作项 ({{ cycleStore.cycleIssues.length }})</h4>
            <button
              v-if="cycle?.status === 'active' || cycle?.status === 'upcoming'"
              @click="toggleAddIssue"
              class="px-2 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700"
            >
              + 添加
            </button>
          </div>

          <!-- Add issue dropdown -->
          <div v-if="showAddIssue" class="mb-3 border border-gray-200 rounded-md p-2">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索工作项..."
              class="w-full px-2 py-1 text-sm border border-gray-300 rounded mb-2 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              @input="searchIssues"
            />
            <div class="max-h-40 overflow-y-auto space-y-1">
              <div
                v-for="issue in availableIssues"
                :key="issue.id"
                @click="handleAddIssue(issue.id)"
                class="flex items-center p-1.5 hover:bg-indigo-50 rounded cursor-pointer text-sm"
              >
                <span class="text-gray-900 truncate flex-1">{{ issue.name }}</span>
                <span class="text-xs text-gray-400 ml-2">#{{ issue.sequence_id }}</span>
              </div>
              <div v-if="availableIssues.length === 0 && searchIssuesDone" class="text-xs text-gray-400 py-2 text-center">
                没有可添加的工作项
              </div>
            </div>
          </div>

          <div v-if="cycleStore.cycleIssues.length === 0" class="text-sm text-gray-400 py-4 text-center">暂无工作项</div>
          <div v-else class="space-y-2">
            <div v-for="issue in cycleStore.cycleIssues" :key="issue.id" class="flex items-center justify-between p-2 bg-gray-50 rounded text-sm">
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
import { useCycleStore } from '@/stores/cycle'
import { issueApi } from '@/api/issue'
import CycleProgressCard from './CycleProgressCard.vue'
import CycleBurndownChart from './CycleBurndownChart.vue'
import { useConfirm } from '@/composables/useConfirm'
import type { CycleResponse } from '@/types/cycle'

const props = defineProps<{
  cycle: CycleResponse | null
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
  closed: []
}>()

const cycleStore = useCycleStore()
const { confirm } = useConfirm()
const loading = computed(() => cycleStore.isLoading)

const showAddIssue = ref(false)
const searchQuery = ref('')
const searchIssuesDone = ref(false)
const availableIssues = ref<any[]>([])

const statusBadgeClass = computed(() => {
  const map: Record<string, string> = {
    upcoming: 'px-2 py-0.5 text-xs rounded bg-blue-100 text-blue-700',
    active: 'px-2 py-0.5 text-xs rounded bg-green-100 text-green-700',
    completed: 'px-2 py-0.5 text-xs rounded bg-gray-100 text-gray-600',
    cancelled: 'px-2 py-0.5 text-xs rounded bg-red-100 text-red-700',
  }
  return map[props.cycle?.status ?? ''] || ''
})

watch(() => props.visible, async (v) => {
  if (v && props.cycle) {
    showAddIssue.value = false
    searchQuery.value = ''
    await Promise.all([
      cycleStore.fetchProgress(props.cycle.id),
      cycleStore.fetchBurndown(props.cycle.id),
      cycleStore.fetchCycleIssues(props.cycle.id),
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
  if (!props.cycle) return
  try {
    const result = await issueApi.listIssues(props.cycle.project_id, props.cycle.workspace_id, {
      search: searchQuery.value || undefined,
    })
    const allIssues = (result as any)?.items || result || []
    const currentIds = new Set(cycleStore.cycleIssues.map((i: any) => i.id))
    availableIssues.value = allIssues.filter(
      (i: any) => !currentIds.has(i.id) && i.state_group !== 'completed' && i.state_group !== 'cancelled'
    )
    searchIssuesDone.value = true
  } catch {
    availableIssues.value = []
    searchIssuesDone.value = true
  }
}

async function handleAddIssue(issueId: number) {
  if (!props.cycle) return
  await cycleStore.addIssueToCycle(props.cycle.id, issueId)
  showAddIssue.value = false
  searchQuery.value = ''
}

async function handleRemoveIssue(issueId: number) {
  if (props.cycle) {
    await cycleStore.removeIssueFromCycle(props.cycle.id, issueId)
  }
}

async function handleStart() {
  if (!props.cycle) return
  await cycleStore.startCycle(props.cycle.id)
  await cycleStore.fetchCycle(props.cycle.id)
}

async function handleEnd() {
  if (!props.cycle) return
  if (!(await confirm('确定要结束这个周期吗？'))) return
  await cycleStore.endCycle(props.cycle.id)
  await cycleStore.fetchCycle(props.cycle.id)
}

async function handleCancel() {
  if (!props.cycle) return
  if (!(await confirm('确定要取消这个周期吗？'))) return
  await cycleStore.cancelCycle(props.cycle.id)
  await cycleStore.fetchCycle(props.cycle.id)
}

async function handleDelete() {
  if (!props.cycle) return
  if (!(await confirm('确定要删除这个周期吗？此操作不可撤销。'))) return
  await cycleStore.deleteCycleAction(props.cycle.id)
  emit('close')
}
</script>

<style scoped>
.slide-enter-active, .slide-leave-active { transition: transform 0.3s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(100%); }
</style>
