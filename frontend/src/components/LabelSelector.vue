<template>
  <div class="label-selector">
    <div class="flex flex-wrap gap-1.5 mb-2">
      <span
        v-for="label in selectedLabels"
        :key="label.id"
        class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium cursor-pointer"
        :style="{ backgroundColor: label.color + '20', color: label.color, border: '1px solid ' + label.color }"
        @click="toggleLabel(label)"
      >
        {{ label.name }}
        <button @click.stop="toggleLabel(label)" class="ml-1 hover:opacity-70">&times;</button>
      </span>
      <span v-if="selectedLabels.length === 0" class="text-xs text-gray-400">{{ t('labelSelector.noLabel') }}</span>
    </div>

    <!-- Dropdown toggle -->
    <button
      @click="open = !open"
      class="w-full px-3 py-1.5 text-xs text-gray-500 border border-dashed border-gray-300 rounded-md hover:border-gray-400 hover:text-gray-700 transition"
    >
      {{ t('labelSelector.addLabel') }}
    </button>

    <!-- Dropdown panel -->
    <div v-if="open" class="mt-2 p-3 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto z-10">
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="label in availableLabels"
          :key="label.id"
          @click="toggleLabel(label)"
          class="px-2 py-0.5 rounded-full text-xs font-medium border transition hover:opacity-80"
          :style="{ backgroundColor: label.color + '15', color: label.color, borderColor: label.color + '40' }"
        >
          {{ label.name }}
        </button>
      </div>
      <div v-if="availableLabels.length === 0" class="text-xs text-gray-400 text-center py-2">
        {{ t('labelSelector.noAvailable') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

interface LabelItem {
  id: number
  name: string
  color: string
}

const props = defineProps<{
  projectId?: number
  labels?: LabelItem[]           // preloaded labels list
  modelValue?: number[]          // selected label IDs
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', ids: number[]): void
  (e: 'change', ids: number[]): void
}>()

const open = ref(false)
const allLabels = ref<LabelItem[]>([])
const selectedIds = ref<number[]>([...props.modelValue || []])

const selectedLabels = computed(() => allLabels.value.filter(l => selectedIds.value.includes(l.id)))
const availableLabels = computed(() => allLabels.value.filter(l => !selectedIds.value.includes(l.id)))

function toggleLabel(label: LabelItem) {
  const idx = selectedIds.value.indexOf(label.id)
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(label.id)
  }
  emit('update:modelValue', [...selectedIds.value])
  emit('change', [...selectedIds.value])
}

async function loadLabels() {
  if (props.labels) {
    allLabels.value = props.labels
    return
  }
  if (!props.projectId) return
  try {
    const api = (await import('@/api/index')).default
    const resp = await api.get(`/projects/${props.projectId}/settings/labels`)
    const data = resp.data
    allLabels.value = Array.isArray(data) ? data : (data?.items || [])
  } catch (e) {
    console.error('Failed to load labels:', e)
  }
}

onMounted(loadLabels)
watch(() => props.labels, (val) => { if (val && val.length > 0) allLabels.value = val })
watch(() => props.modelValue, (val) => { selectedIds.value = [...(val || [])] })
</script>
