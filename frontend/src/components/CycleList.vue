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
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-2 text-gray-500">加载中...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="cycles.length === 0" class="text-center py-12">
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
          @start="$emit('start', cycle)"
          @end="$emit('end', cycle)"
          @cancel="$emit('cancel', cycle)"
          @delete="$emit('delete', cycle)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import CycleCard from './CycleCard.vue'
import cycleApi from '@/api/cycle'
import type { CycleResponse } from '@/types/cycle'

// Props
const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

// Emits
defineEmits<{
  (e: 'create'): void
  (e: 'select', cycle: CycleResponse): void
  (e: 'start', cycle: CycleResponse): void
  (e: 'end', cycle: CycleResponse): void
  (e: 'cancel', cycle: CycleResponse): void
  (e: 'delete', cycle: CycleResponse): void
}>()

// State
const cycles = ref<CycleResponse[]>([])
const loading = ref(false)

const filters = ref({
  status: ''
})

// Filtered cycles
const filteredCycles = computed(() => {
  if (!filters.value.status) return cycles.value
  return cycles.value.filter(c => c.status === filters.value.status)
})

// Load cycles
onMounted(() => {
  loadCycles()
})

watch(filters, () => {
  // Filters are computed, no need to reload
}, { deep: true })

async function loadCycles() {
  loading.value = true
  try {
    const result = await cycleApi.listCycles(props.projectId)
    cycles.value = result.items
  } catch (error) {
    console.error('Failed to load cycles:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.cycle-list {
  @apply bg-white rounded-lg;
}
</style>