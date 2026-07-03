<template>
  <div v-if="show" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white rounded-xl p-6 w-full max-w-lg max-h-[80vh] flex flex-col">
      <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('pages.fromTemplate') }}</h3>

      <div v-if="loading" class="flex justify-center py-8">
        <div class="animate-spin h-6 w-6 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
      </div>

      <div v-else class="flex-1 overflow-y-auto space-y-2">
        <!-- Blank page option -->
        <div
          @click="selectBlank"
          class="p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-indigo-300 hover:bg-indigo-50 transition-colors"
          :class="{ 'border-indigo-500 bg-indigo-50': selectedId === 0 }"
        >
          <div class="font-medium text-gray-900 text-sm">{{ t('pages.blankPage') }}</div>
          <div class="text-xs text-gray-500 mt-0.5">{{ t('pages.blankPageDesc') }}</div>
        </div>

        <div
          v-for="tmpl in templates"
          :key="tmpl.id"
          @click="selectTemplate(tmpl)"
          class="p-3 border border-gray-200 rounded-lg cursor-pointer hover:border-indigo-300 hover:bg-indigo-50 transition-colors"
          :class="{ 'border-indigo-500 bg-indigo-50': selectedId === tmpl.id }"
        >
          <div class="flex items-center justify-between">
            <span class="font-medium text-gray-900 text-sm">{{ tmpl.name }}</span>
            <span v-if="tmpl.is_default" class="text-xs bg-indigo-100 text-indigo-700 px-1.5 py-0.5 rounded-full">{{ t('pages.default') }}</span>
          </div>
          <div v-if="tmpl.description" class="text-xs text-gray-500 mt-0.5">{{ tmpl.description }}</div>
        </div>
      </div>

      <div class="flex justify-end space-x-3 mt-4 pt-4 border-t border-gray-200">
        <button @click="$emit('close')" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">{{ t('common.cancel') }}</button>
        <button @click="confirmSelection" class="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700">
          {{ t('pages.useTemplate') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import type { PageTemplate } from '@/types/page'
import * as pageApi from '@/api/page'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', template: PageTemplate | null): void
}>()

const templates = ref<PageTemplate[]>([])
const loading = ref(false)
const selectedId = ref<number>(0)
const selectedTemplate = ref<PageTemplate | null>(null)

watch(() => props.show, (val) => {
  if (val) {
    selectedId.value = 0
    selectedTemplate.value = null
    loadTemplates()
  }
})

async function loadTemplates() {
  loading.value = true
  try {
    templates.value = await pageApi.listPageTemplates(props.projectId, props.workspaceId)
  } catch (e) {
    console.error('Failed to load templates:', e)
  } finally {
    loading.value = false
  }
}

function selectBlank() {
  selectedId.value = 0
  selectedTemplate.value = null
}

function selectTemplate(tmpl: PageTemplate) {
  selectedId.value = tmpl.id
  selectedTemplate.value = tmpl
}

function confirmSelection() {
  emit('select', selectedTemplate.value)
}
</script>
