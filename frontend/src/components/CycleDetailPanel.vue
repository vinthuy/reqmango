<template>
  <Transition name="slide">
    <div v-if="visible" class="fixed inset-y-0 right-0 w-96 bg-white shadow-xl border-l border-gray-200 z-50 overflow-y-auto">
      <div class="sticky top-0 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between z-10">
        <h3 class="text-lg font-semibold text-gray-900 truncate">{{ cycle?.name }}</h3>
        <button @click="$emit('close')" class="p-1 text-gray-400 hover:text-gray-600">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
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

        <div>
          <h4 class="text-sm font-medium text-gray-700 mb-2">周期工作项 ({{ cycleStore.cycleIssues.length }})</h4>
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
import { computed, watch } from 'vue'
import { useCycleStore } from '@/stores/cycle'
import CycleProgressCard from './CycleProgressCard.vue'
import CycleBurndownChart from './CycleBurndownChart.vue'
import type { CycleResponse } from '@/types/cycle'

const props = defineProps<{
  cycle: CycleResponse | null
  visible: boolean
}>()

defineEmits<{
  close: []
}>()

const cycleStore = useCycleStore()
const loading = computed(() => cycleStore.isLoading)

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
    await Promise.all([
      cycleStore.fetchProgress(props.cycle.id),
      cycleStore.fetchBurndown(props.cycle.id),
      cycleStore.fetchCycleIssues(props.cycle.id),
    ])
  }
})

async function handleRemoveIssue(issueId: number) {
  if (props.cycle) {
    await cycleStore.removeIssueFromCycle(props.cycle.id, issueId)
  }
}
</script>

<style scoped>
.slide-enter-active, .slide-leave-active { transition: transform 0.3s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(100%); }
</style>
