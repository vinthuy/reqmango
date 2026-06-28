<template>
  <div class="custom-field-list">
    <!-- 头部 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-gray-700">{{ t('customField.customFields') }}</h3>
        <button
          @click="$emit('create')"
          class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span>{{ t('customField.createField') }}</span>
        </button>
      </div>
    </div>

    <!-- 字段列表 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <!-- 空状态 -->
      <div v-else-if="fields.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <p class="mt-2 text-gray-500">{{ t('customField.noCustomFields') }}</p>
        <button @click="$emit('create')" class="mt-3 text-indigo-600 hover:text-indigo-800 text-sm">
          {{ t('customField.createFirstField') }}
        </button>
      </div>

      <!-- 字段列表 -->
      <div v-else class="space-y-3">
        <div
          v-for="field in fields"
          :key="field.id"
          class="bg-white border border-gray-200 rounded-lg p-4 hover:border-gray-300"
        >
          <div class="flex items-start justify-between">
            <!-- 字段信息 -->
            <div class="flex items-start space-x-3 flex-1">
              <!-- 字段类型图标 -->
              <div class="mt-0.5" :class="getFieldTypeColor(field.field_type)">
                <component :is="getFieldTypeIcon(field.field_type)" class="w-5 h-5" />
              </div>

              <div class="flex-1 min-w-0">
                <div class="flex items-center space-x-2">
                  <h4 class="text-sm font-medium text-gray-900">{{ field.name }}</h4>
                  <span class="px-1.5 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">
                    {{ getFieldTypeName(field.field_type) }}
                  </span>
                  <span
                    v-if="field.is_required"
                    class="px-1.5 py-0.5 text-xs bg-red-100 text-red-600 rounded"
                  >
                    {{ t('customField.required') }}
                  </span>
                </div>

                <p v-if="field.description" class="text-xs text-gray-500 mt-0.5">
                  {{ field.description }}
                </p>

                <!-- 选项预览（如果是选择类型） -->
                <div v-if="hasOptions(field.field_type) && field.options" class="mt-2 flex flex-wrap gap-1">
                  <span
                    v-for="(option, index) in field.options.slice(0, 5)"
                    :key="index"
                    class="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded"
                  >
                    {{ option.value }}
                  </span>
                  <span v-if="field.options.length > 5" class="text-xs text-gray-400">
                    +{{ field.options.length - 5 }} {{ t('customField.more') }}
                  </span>
                </div>
              </div>
            </div>

            <!-- 操作 -->
            <div class="flex items-center space-x-2 ml-4">
              <button
                @click="$emit('edit', field)"
                class="p-1.5 text-gray-400 hover:text-indigo-600 rounded"
                :title="t('common.edit')"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                @click="deleteField(field)"
                class="p-1.5 text-gray-400 hover:text-red-600 rounded"
                :title="t('common.delete')"
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
import { ref, onMounted, h } from 'vue'
import { useI18n } from '@/composables/useI18n'
import customFieldApi from '@/api/custom-field'
import { useConfirm } from '@/composables/useConfirm'
import type { CustomField } from '@/types/custom-field'

const { t } = useI18n()

// Props
const props = defineProps<{
  projectId: number
}>()

// Emits
defineEmits<{
  (e: 'create'): void
  (e: 'edit', field: CustomField): void
}>()

// State
const { confirm } = useConfirm()
const fields = ref<CustomField[]>([])
const loading = ref(false)

// Load fields
onMounted(() => {
  loadFields()
})

async function loadFields() {
  loading.value = true
  try {
    fields.value = await customFieldApi.listCustomFields(props.projectId)
  } catch (error) {
    console.error('Failed to load custom fields:', error)
  } finally {
    loading.value = false
  }
}

// Delete field
async function deleteField(field: CustomField) {
  if (!(await confirm(t('customField.confirmDeleteField').replace('{name}', field.name)))) return

  try {
    await customFieldApi.deleteCustomField(field.id)
    await loadFields()
  } catch (error) {
    console.error('Failed to delete field:', error)
  }
}

// Field type helpers
function hasOptions(fieldType: string): boolean {
  return ['select', 'multi_select', 'radio', 'checkbox'].includes(fieldType)
}

function getFieldTypeName(fieldType: string): string {
  const names: Record<string, string> = {
    text: t('customField.typeText'),
    number: t('customField.typeNumber'),
    select: t('customField.typeSelectShort'),
    multi_select: t('customField.typeMultiSelect'),
    date: t('customField.typeDate'),
    checkbox: t('customField.typeCheckbox'),
    radio: t('customField.typeRadio'),
    url: t('customField.typeUrl'),
    email: t('customField.typeEmail'),
    phone: t('customField.typePhone'),
    user: t('customField.typeUser')
  }
  return names[fieldType] || fieldType
}

function getFieldTypeColor(fieldType: string): string {
  const colors: Record<string, string> = {
    text: 'text-gray-500',
    number: 'text-blue-500',
    select: 'text-purple-500',
    multi_select: 'text-pink-500',
    date: 'text-green-500',
    checkbox: 'text-orange-500',
    radio: 'text-indigo-500',
    url: 'text-cyan-500',
    email: 'text-teal-500',
    phone: 'text-yellow-500',
    user: 'text-red-500'
  }
  return colors[fieldType] || 'text-gray-500'
}

function getFieldTypeIcon(fieldType: string) {
  const icons: Record<string, any> = {
    text: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z' })
    ]),
    number: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M7 20l4-16m2 16l4-16M6 9h14M4 15h14' })
    ]),
    select: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M19 9l-7 7-7-7' })
    ]),
    date: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' })
    ]),
    checkbox: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' })
    ])
  }
  return icons[fieldType] || icons.text
}
</script>

<style scoped>
.custom-field-list {
  @apply bg-white rounded-lg;
}
</style>
