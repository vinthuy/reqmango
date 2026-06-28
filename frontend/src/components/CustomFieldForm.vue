<template>
  <div class="custom-field-form">
    <form @submit.prevent="handleSubmit" class="space-y-6">
      <!-- 基本信息 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('customField.basicInfo') }}</h3>

        <!-- 字段名称 -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('customField.fieldName') }}
          </label>
          <input
            v-model="form.name"
            type="text"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :placeholder="t('customField.fieldNamePlaceholder')"
          />
        </div>

        <!-- 描述 -->
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('customField.description') }}
          </label>
          <textarea
            v-model="form.description"
            rows="2"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :placeholder="t('customField.descriptionPlaceholder')"
          />
        </div>

        <!-- 字段类型 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            {{ t('customField.fieldType') }}
          </label>
          <select
            v-model="form.field_type"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :disabled="isEdit"
          >
            <option value="">{{ t('customField.selectFieldType') }}</option>
            <option value="text">{{ t('customField.typeText') }}</option>
            <option value="number">{{ t('customField.typeNumber') }}</option>
            <option value="select">{{ t('customField.typeSelect') }}</option>
            <option value="multi_select">{{ t('customField.typeMultiSelect') }}</option>
            <option value="date">{{ t('customField.typeDate') }}</option>
            <option value="checkbox">{{ t('customField.typeCheckbox') }}</option>
            <option value="radio">{{ t('customField.typeRadio') }}</option>
            <option value="url">{{ t('customField.typeUrl') }}</option>
            <option value="email">{{ t('customField.typeEmail') }}</option>
            <option value="phone">{{ t('customField.typePhone') }}</option>
            <option value="user">{{ t('customField.typeUser') }}</option>
          </select>
          <p v-if="isEdit" class="mt-1 text-xs text-gray-500">
            {{ t('customField.cantChangeTypeInEdit') }}
          </p>
        </div>
      </div>

      <!-- 选项配置（选择类型） -->
      <div v-if="hasOptions" class="bg-white border border-gray-200 rounded-lg p-4">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-medium text-gray-700">{{ t('customField.options') }}</h3>
          <button
            type="button"
            @click="addOption"
            class="text-sm text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('customField.addOption') }}</span>
          </button>
        </div>

        <div class="space-y-2">
          <div
            v-for="(option, index) in form.options"
            :key="index"
            class="flex items-center space-x-2"
          >
            <input
              v-model="option.label"
              type="text"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              :placeholder="t('customField.optionLabel')"
            />
            <input
              v-model="option.value"
              type="text"
              class="w-24 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              :placeholder="t('customField.optionValue')"
            />
            <input
              v-model="option.color"
              type="color"
              class="w-8 h-8 border border-gray-300 rounded cursor-pointer"
            />
            <button
              type="button"
              @click="removeOption(index)"
              class="p-1 text-gray-400 hover:text-red-600"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <p v-if="form.options.length === 0" class="text-center py-4 text-gray-500 text-sm">
          {{ t('customField.noOptions') }}
        </p>
      </div>

      <!-- 验证设置 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('customField.validationSettings') }}</h3>

        <!-- 必填 -->
        <div class="mb-4">
          <label class="flex items-center space-x-2">
            <input
              v-model="form.is_required"
              type="checkbox"
              class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
            />
            <span class="text-sm text-gray-700">{{ t('customField.isRequired') }}</span>
          </label>
          <p class="mt-1 text-xs text-gray-500 ml-6">
            {{ t('customField.requiredHint') }}
          </p>
        </div>

        <!-- 唯一值 -->
        <div class="mb-4">
          <label class="flex items-center space-x-2">
            <input
              v-model="form.is_unique"
              type="checkbox"
              class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
            />
            <span class="text-sm text-gray-700">{{ t('customField.isUnique') }}</span>
          </label>
          <p class="mt-1 text-xs text-gray-500 ml-6">
            {{ t('customField.uniqueHint') }}
          </p>
        </div>

        <!-- 数值范围（数字类型） -->
        <div v-if="form.field_type === 'number'" class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('customField.minValue') }}
            </label>
            <input
              v-model.number="form.min_value"
              type="number"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('customField.maxValue') }}
            </label>
            <input
              v-model.number="form.max_value"
              type="number"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
        </div>

        <!-- 文本长度（文本类型） -->
        <div v-if="form.field_type === 'text'" class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('customField.minLength') }}
            </label>
            <input
              v-model.number="form.min_length"
              type="number"
              min="0"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              {{ t('customField.maxLength') }}
            </label>
            <input
              v-model.number="form.max_length"
              type="number"
              min="1"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
        </div>
      </div>

      <!-- 默认值 -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-700 mb-4">{{ t('customField.defaultValue') }}</h3>

        <!-- 文本默认值 -->
        <div v-if="form.field_type === 'text' || form.field_type === 'url' || form.field_type === 'email' || form.field_type === 'phone'">
          <input
            v-model="form.default_value"
            type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :placeholder="t('customField.defaultValue')"
          />
        </div>

        <!-- 数字默认值 -->
        <div v-if="form.field_type === 'number'">
          <input
            v-model.number="form.default_value"
            type="number"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
            :placeholder="t('customField.defaultValue')"
          />
        </div>

        <!-- 日期默认值 -->
        <div v-if="form.field_type === 'date'">
          <input
            v-model="form.default_value"
            type="date"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>

        <!-- 复选框默认值 -->
        <div v-if="form.field_type === 'checkbox'">
          <label class="flex items-center space-x-2">
            <input
              v-model="form.default_value"
              type="checkbox"
              class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
            />
            <span class="text-sm text-gray-700">{{ t('customField.defaultChecked') }}</span>
          </label>
        </div>

        <!-- 下拉/单选/多选默认值 -->
        <div v-if="hasOptions && form.options.length > 0">
          <select
            v-if="form.field_type === 'select' || form.field_type === 'radio'"
            v-model="form.default_value"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">{{ t('customField.noDefaultValue') }}</option>
            <option v-for="option in form.options" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>

          <div v-if="form.field_type === 'multi_select'" class="space-y-2">
            <label
              v-for="option in form.options"
              :key="option.value"
              class="flex items-center space-x-2"
            >
              <input
                v-model="form.default_value"
                type="checkbox"
                :value="option.value"
                class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
              />
              <span class="text-sm text-gray-700">{{ option.label }}</span>
            </label>
          </div>
        </div>
      </div>

      <!-- 提交按钮 -->
      <div class="flex justify-end space-x-3">
        <button
          type="button"
          @click="$emit('cancel')"
          class="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          :disabled="submitting || !isValid"
          class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
        >
          {{ submitting ? t('customField.saving') : (isEdit ? t('common.save') : t('common.create')) }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import customFieldApi from '@/api/custom-field'
import type { CustomFieldCreate, CustomFieldUpdate, CustomField } from '@/types/custom-field'

const { t } = useI18n()

// Props
const props = defineProps<{
  projectId: number
  workspaceId: number
  field?: CustomField
}>()

// Emits
const emit = defineEmits<{
  (e: 'submit', data: CustomFieldCreate | CustomFieldUpdate): void
  (e: 'cancel'): void
}>()

// State
const submitting = ref(false)

// Form data
const form = ref<{
  name: string
  description: string
  field_type: string
  is_required: boolean
  is_unique: boolean
  options: Array<{ id?: number; label: string; value: string; color: string; sequence?: number; is_default?: boolean; is_active?: boolean }>
  default_value: any
  min_value?: number
  max_value?: number
  min_length?: number
  max_length?: number
}>({
  name: '',
  description: '',
  field_type: '',
  is_required: false,
  is_unique: false,
  options: [],
  default_value: null
})

// Check if edit mode
const isEdit = computed(() => !!props.field)

// Check if field type has options
const hasOptions = computed(() => {
  return ['select', 'multi_select', 'radio', 'checkbox'].includes(form.value.field_type)
})

// Validate form
const isValid = computed(() => {
  if (!form.value.name.trim()) return false
  if (!form.value.field_type) return false
  if (hasOptions.value && form.value.options.length === 0) return false
  return true
})

// Initialize form with existing data
if (props.field) {
  form.value = {
    name: props.field.name,
    description: props.field.description || '',
    field_type: props.field.field_type,
    is_required: props.field.is_required || false,
    is_unique: props.field.is_unique || false,
    options: props.field.options ? props.field.options.map(opt => ({
      id: opt.id,
      label: opt.value,
      value: opt.value,
      color: opt.color || '#6366f1',
      sequence: opt.sequence,
      is_default: opt.is_default,
      is_active: opt.is_active
    })) : [],
    default_value: props.field.default_value
  }
}

// Watch field type changes
watch(() => form.value.field_type, (newType, oldType) => {
  // Reset options when changing field type
  if (newType !== oldType && !isEdit.value) {
    form.value.options = []
    form.value.default_value = null
  }
})

// Add option
function addOption() {
  form.value.options.push({
    label: '',
    value: '',
    color: '#6366f1'
  })
}

// Remove option
function removeOption(index: number) {
  form.value.options.splice(index, 1)
}

// Handle submit
async function handleSubmit() {
  if (!isValid.value) return

  submitting.value = true
  try {
    const data: CustomFieldCreate | CustomFieldUpdate = {
      name: form.value.name,
      description: form.value.description || undefined,
      field_type: form.value.field_type as any,
      is_required: form.value.is_required,
      is_unique: form.value.is_unique,
      options: hasOptions.value ? form.value.options : undefined,
      default_value: form.value.default_value || undefined,
      min_value: form.value.min_value,
      max_value: form.value.max_value,
      min_length: form.value.min_length,
      max_length: form.value.max_length
    }

    if (isEdit.value && props.field) {
      await customFieldApi.updateCustomField(props.field.id, data as CustomFieldUpdate)
    } else {
      await customFieldApi.createCustomField(props.workspaceId, data as CustomFieldCreate)
    }

    emit('submit', data)
  } catch (e: any) {
    alert(e.response?.data?.message || t('customField.operationFailed'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.custom-field-form {
  @apply bg-gray-50;
}
</style>
