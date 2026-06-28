<template>
  <div class="estimate-point-manager">
    <div class="bg-white rounded-lg border border-gray-200">
      <div class="px-4 py-3 border-b border-gray-200">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium text-gray-700">{{ t('estimatePoint.title') }}</h3>
        </div>
        
        <div class="mt-3 flex items-center space-x-4">
          <span class="text-xs text-gray-500">{{ t('estimatePoint.mode') }}</span>
          <div class="flex space-x-1 bg-gray-100 rounded-lg p-1">
            <button
              v-for="mode in modes"
              :key="mode.value"
              @click="switchMode(mode.value)"
              :class="[
                'px-3 py-1.5 text-xs rounded-md font-medium transition-colors',
                currentMode === mode.value
                  ? 'bg-white text-indigo-600 shadow-sm'
                  : 'text-gray-600 hover:text-gray-800'
              ]"
            >
              {{ mode.label }}
            </button>
          </div>
        </div>
      </div>

      <div class="p-4">
        <div v-if="loading" class="text-center py-8">
          <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>

        <div v-else-if="currentMode === 'points'" class="space-y-2">
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm font-medium text-gray-700">{{ t('estimatePoint.estimatePoints') }}</span>
            <div class="flex items-center space-x-2">
              <button
                @click="createDefaults('points')"
                class="px-2 py-1 text-xs text-indigo-600 hover:text-indigo-800"
              >
                {{ t('estimatePoint.useDefault') }}
              </button>
              <button
                @click="$emit('create', 'points')"
                class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                <span>{{ t('estimatePoint.add') }}</span>
              </button>
            </div>
          </div>
          
          <div v-if="points.length === 0" class="text-center py-8">
            <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
            </svg>
            <p class="mt-2 text-gray-500 text-sm">{{ t('estimatePoint.noPoints') }}</p>
            <button @click="createDefaults('points')" class="mt-2 text-indigo-600 hover:text-indigo-800 text-sm">
              {{ t('estimatePoint.createWithDefault') }}
            </button>
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="point in points"
              :key="point.id"
              class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100"
            >
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-indigo-100 text-indigo-700 rounded-lg flex items-center justify-center font-semibold">
                  {{ point.value }}
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ point.name }}</p>
                  <p class="text-xs text-gray-500">{{ t('estimatePoint.value') }} {{ point.value }}</p>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <span v-if="point.is_default" class="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded">{{ t('estimatePoint.default') }}</span>
                <button v-if="!point.is_default" @click="setDefault(point)" class="p-1 text-gray-400 hover:text-indigo-600" :title="t('estimatePoint.setDefault')">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                  </svg>
                </button>
                <button @click="$emit('edit', point, 'points')" class="p-1 text-gray-400 hover:text-indigo-600" :title="t('estimatePoint.edit')">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button @click="deleteItem(point, 'points')" class="p-1 text-gray-400 hover:text-red-600" :title="t('estimatePoint.delete')">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-else-if="currentMode === 'categories'" class="space-y-2">
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm font-medium text-gray-700">{{ t('estimatePoint.categories') }}</span>
            <div class="flex items-center space-x-2">
              <button
                @click="createDefaults('categories')"
                class="px-2 py-1 text-xs text-indigo-600 hover:text-indigo-800"
              >
                {{ t('estimatePoint.useDefault') }}
              </button>
              <button
                @click="$emit('create', 'categories')"
                class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                <span>{{ t('estimatePoint.add') }}</span>
              </button>
            </div>
          </div>
          
          <div v-if="categories.length === 0" class="text-center py-8">
            <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
            </svg>
            <p class="mt-2 text-gray-500 text-sm">{{ t('estimatePoint.noCategories') }}</p>
            <button @click="createDefaults('categories')" class="mt-2 text-indigo-600 hover:text-indigo-800 text-sm">
              {{ t('estimatePoint.createWithDefault') }}
            </button>
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="cat in categories"
              :key="cat.id"
              class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100"
            >
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-purple-100 text-purple-700 rounded-lg flex items-center justify-center font-semibold">
                  {{ cat.name.split(' ')[0] }}
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ cat.name }}</p>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <span v-if="cat.is_default" class="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded">{{ t('estimatePoint.default') }}</span>
                <button @click="deleteItem(cat, 'categories')" class="p-1 text-gray-400 hover:text-red-600" :title="t('estimatePoint.delete')">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-else-if="currentMode === 'time'" class="space-y-2">
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm font-medium text-gray-700">{{ t('estimatePoint.timeEstimates') }}</span>
            <div class="flex items-center space-x-2">
              <button
                @click="createDefaults('time')"
                class="px-2 py-1 text-xs text-indigo-600 hover:text-indigo-800"
              >
                {{ t('estimatePoint.useDefault') }}
              </button>
              <button
                @click="$emit('create', 'time')"
                class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                <span>{{ t('estimatePoint.add') }}</span>
              </button>
            </div>
          </div>
          
          <div v-if="timeEstimates.length === 0" class="text-center py-8">
            <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
            </svg>
            <p class="mt-2 text-gray-500 text-sm">{{ t('estimatePoint.noTimeOptions') }}</p>
            <button @click="createDefaults('time')" class="mt-2 text-indigo-600 hover:text-indigo-800 text-sm">
              {{ t('estimatePoint.createWithDefault') }}
            </button>
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="item in timeEstimates"
              :key="item.id"
              class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100"
            >
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-blue-100 text-blue-700 rounded-lg flex items-center justify-center font-semibold">
                  {{ formatMinutes(item.minutes) }}
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900">{{ item.name }}</p>
                  <p class="text-xs text-gray-500">{{ item.minutes }} {{ t('estimatePoint.minutes') }}</p>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <span v-if="item.is_default" class="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded">{{ t('estimatePoint.default') }}</span>
                <button @click="deleteItem(item, 'time')" class="p-1 text-gray-400 hover:text-red-600" :title="t('estimatePoint.delete')">
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import estimatePointApi from '@/api/estimate-point'
import { useConfirm } from '@/composables/useConfirm'
import { useI18n } from '@/composables/useI18n'
import type { EstimatePoint, EstimateCategory, EstimateTime, EstimateMode } from '@/types/estimate-point'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
}>()

defineEmits<{
  (e: 'create', mode: EstimateMode): void
  (e: 'edit', item: EstimatePoint | EstimateCategory | EstimateTime, mode: EstimateMode): void
}>()

const { confirm } = useConfirm()
const currentMode = ref<EstimateMode>('points')
const loading = ref(false)
const points = ref<EstimatePoint[]>([])
const categories = ref<EstimateCategory[]>([])
const timeEstimates = ref<EstimateTime[]>([])

const modes = [
  { value: 'points' as EstimateMode, label: t('estimatePoint.points') },
  { value: 'categories' as EstimateMode, label: t('estimatePoint.tshirt') },
  { value: 'time' as EstimateMode, label: t('estimatePoint.time') },
]

onMounted(() => {
  loadSettings()
  loadData()
})

watch(currentMode, () => {
  loadData()
})

async function loadSettings() {
  try {
    const settings = await estimatePointApi.getEstimateSettings(props.projectId)
    currentMode.value = settings.mode
  } catch {
  }
}

async function switchMode(mode: EstimateMode) {
  currentMode.value = mode
  try {
    await estimatePointApi.updateEstimateSettings(props.projectId, mode)
  } catch {
  }
}

async function loadData() {
  loading.value = true
  try {
    switch (currentMode.value) {
      case 'points':
        points.value = await estimatePointApi.listEstimatePoints(props.projectId)
        break
      case 'categories':
        categories.value = await estimatePointApi.listEstimateCategories(props.projectId)
        break
      case 'time':
        timeEstimates.value = await estimatePointApi.listEstimateTime(props.projectId)
        break
    }
  } catch (error) {
    console.error('Failed to load estimate data:', error)
  } finally {
    loading.value = false
  }
}

async function createDefaults(mode: EstimateMode) {
  try {
    switch (mode) {
      case 'points':
        points.value = await estimatePointApi.createDefaultEstimatePoints(props.projectId)
        break
      case 'categories':
        categories.value = await estimatePointApi.createDefaultEstimateCategories(props.projectId)
        break
      case 'time':
        timeEstimates.value = await estimatePointApi.createDefaultEstimateTime(props.projectId)
        break
    }
  } catch (error) {
    console.error('Failed to create defaults:', error)
  }
}

async function setDefault(point: EstimatePoint) {
  try {
    await estimatePointApi.updateEstimatePoint(props.projectId, point.id, { is_default: true })
    await loadData()
  } catch (error) {
    console.error('Failed to set default:', error)
  }
}

async function deleteItem(
  item: EstimatePoint | EstimateCategory | EstimateTime,
  mode: EstimateMode
) {
  if (!(await confirm(t('estimatePoint.confirmDelete').replace('{name}', (item as EstimatePoint).name)))) return

  try {
    if (mode === 'points') {
      await estimatePointApi.deleteEstimatePoint(props.projectId, item.id)
    }
    await loadData()
  } catch (error) {
    console.error('Failed to delete:', error)
  }
}

function formatMinutes(minutes: number): string {
  if (minutes < 60) return `${minutes}m`
  if (minutes < 480) return `${minutes / 60}h`
  return `${minutes / 480}d`
}
</script>

<style scoped>
.estimate-point-manager {
  @apply bg-white rounded-lg;
}
</style>