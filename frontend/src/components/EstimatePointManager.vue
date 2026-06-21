<template>
  <div class="estimate-point-manager">
    <div class="bg-white rounded-lg border border-gray-200">
      <!-- 头部 -->
      <div class="px-4 py-3 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium text-gray-700">估算点</h3>
          <div class="flex items-center space-x-2">
            <button
              @click="createDefaults"
              class="px-2 py-1 text-xs text-indigo-600 hover:text-indigo-800"
              :disabled="loading"
            >
              使用默认模板
            </button>
            <button
              @click="$emit('create')"
              class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              <span>添加</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 估算点列表 -->
      <div class="p-4">
        <!-- 加载状态 -->
        <div v-if="loading" class="text-center py-8">
          <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>

        <!-- 空状态 -->
        <div v-else-if="points.length === 0" class="text-center py-8">
          <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
          </svg>
          <p class="mt-2 text-gray-500 text-sm">暂无估算点</p>
          <button
            @click="createDefaults"
            class="mt-2 text-indigo-600 hover:text-indigo-800 text-sm"
          >
            使用默认模板创建
          </button>
        </div>

        <!-- 估算点列表 -->
        <div v-else class="space-y-2">
          <div
            v-for="point in points"
            :key="point.id"
            class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100"
          >
            <div class="flex items-center space-x-3">
              <!-- 值显示 -->
              <div class="w-10 h-10 bg-indigo-100 text-indigo-700 rounded-lg flex items-center justify-center font-semibold">
                {{ point.value }}
              </div>
              <div>
                <p class="text-sm font-medium text-gray-900">{{ point.name }}</p>
                <p class="text-xs text-gray-500">值: {{ point.value }}</p>
              </div>
            </div>

            <div class="flex items-center space-x-2">
              <!-- 默认标记 -->
              <span
                v-if="point.is_default"
                class="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded"
              >
                默认
              </span>

              <!-- 设为默认 -->
              <button
                v-if="!point.is_default"
                @click="setDefault(point)"
                class="p-1 text-gray-400 hover:text-indigo-600"
                title="设为默认"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                </svg>
              </button>

              <!-- 编辑 -->
              <button
                @click="$emit('edit', point)"
                class="p-1 text-gray-400 hover:text-indigo-600"
                title="编辑"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>

              <!-- 删除 -->
              <button
                @click="deletePoint(point)"
                class="p-1 text-gray-400 hover:text-red-600"
                title="删除"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
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
import { ref, onMounted } from 'vue'
import estimatePointApi from '@/api/estimate-point'
import { useConfirm } from '@/composables/useConfirm'
import type { EstimatePoint } from '@/types/estimate-point'

// Props
const props = defineProps<{
  projectId: number
}>()

// Emits
defineEmits<{
  (e: 'create'): void
  (e: 'edit', point: EstimatePoint): void
}>()

// State
const { confirm } = useConfirm()
const points = ref<EstimatePoint[]>([])
const loading = ref(false)

// Load points
onMounted(() => {
  loadPoints()
})

async function loadPoints() {
  loading.value = true
  try {
    points.value = await estimatePointApi.listEstimatePoints(props.projectId)
  } catch (error) {
    console.error('Failed to load estimate points:', error)
  } finally {
    loading.value = false
  }
}

// Create default points
async function createDefaults() {
  try {
    const newPoints = await estimatePointApi.createDefaultEstimatePoints(props.projectId)
    points.value = newPoints
  } catch (error) {
    console.error('Failed to create default points:', error)
  }
}

// Set default
async function setDefault(point: EstimatePoint) {
  try {
    await estimatePointApi.updateEstimatePoint(props.projectId, point.id, {
      is_default: true
    })
    await loadPoints()
  } catch (error) {
    console.error('Failed to set default:', error)
  }
}

// Delete point
async function deletePoint(point: EstimatePoint) {
  if (!(await confirm(`确定要删除估算点 "${point.name}" 吗？`))) return

  try {
    await estimatePointApi.deleteEstimatePoint(props.projectId, point.id)
    await loadPoints()
  } catch (error) {
    console.error('Failed to delete point:', error)
  }
}
</script>

<style scoped>
.estimate-point-manager {
  @apply bg-white rounded-lg;
}
</style>