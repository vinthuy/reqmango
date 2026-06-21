<template>
  <div class="cycle-list">
    <!-- 头部工具栏 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <!-- 状态筛选 -->
          <select
            v-model="filters.status"
            class="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">所有状态</option>
            <option value="upcoming">即将开始</option>
            <option value="active">进行中</option>
            <option value="completed">已完成</option>
            <option value="cancelled">已取消</option>
          </select>
        </div>

        <div class="flex items-center space-x-3">
          <!-- 新建按钮 -->
          <button
            @click="$emit('create')"
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>新建周期</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 列表内容 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="cycleStore.isLoading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="cycleStore.cycles.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="mt-2 text-gray-500">暂无周期</p>
        <button @click="$emit('create')" class="mt-3 text-indigo-600 hover:text-indigo-800 text-sm">
          创建第一个周期
        </button>
      </div>

      <!-- 周期网格 -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <CycleCard
          v-for="cycle in filteredCycles"
          :key="cycle.id"
          :cycle="cycle"
          @click="$emit('select', cycle)"
          @start="handleStart(cycle)"
          @end="handleEnd(cycle)"
          @cancel="handleCancel(cycle)"
          @delete="handleDelete(cycle)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useCycleStore } from '@/stores/cycle'
import CycleCard from './CycleCard.vue'
import { useConfirm } from '@/composables/useConfirm'
import type { CycleResponse } from '@/types/cycle'

const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

// Use cycle_ prefix to avoid conflict with store's own functions
defineEmits<{
  'create': []
  'select': [cycle: CycleResponse]
}>()

const cycleStore = useCycleStore()
const { confirm } = useConfirm()

const filters = ref({
  status: ''
})

const filteredCycles = computed(() => {
  // Server-side filtering via store — the store's cycles ref is already filtered
  return cycleStore.cycles
})

async function handleStart(cycle: CycleResponse) {
  await cycleStore.startCycle(cycle.id)
}

async function handleEnd(cycle: CycleResponse) {
  if (!(await confirm(`确定要结束周期 "${cycle.name}" 吗？`))) return
  await cycleStore.endCycle(cycle.id)
}

async function handleCancel(cycle: CycleResponse) {
  if (!(await confirm(`确定要取消周期 "${cycle.name}" 吗？`))) return
  await cycleStore.cancelCycle(cycle.id)
}

async function handleDelete(cycle: CycleResponse) {
  if (!(await confirm(`确定要删除周期 "${cycle.name}" 吗？此操作不可撤销。`))) return
  await cycleStore.deleteCycleAction(cycle.id)
}

onMounted(async () => {
  await cycleStore.fetchCycles(props.projectId)
})

// Reload from backend when status filter changes
watch(() => filters.value.status, async (newStatus) => {
  await cycleStore.fetchCycles(props.projectId, newStatus || undefined)
})
</script>

<style scoped>
.cycle-list {
  @apply bg-white rounded-lg;
}
</style>
