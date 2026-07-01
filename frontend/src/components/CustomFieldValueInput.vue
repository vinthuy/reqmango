<template>
  <div class="custom-field-value-input">
    <!-- 文本类型 -->
    <template v-if="field.field_type === 'text'">
      <input
        v-if="field.text_type === 'single'"
        type="text"
        :value="localValue.text_value"
        @input="updateTextValue(($event.target as HTMLInputElement).value)"
        :placeholder="field.placeholder || t('customField.enterValue')"
        :disabled="field.is_readonly"
        :required="field.is_required"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
      <textarea
        v-else
        :value="localValue.text_value"
        @input="updateTextValue(($event.target as HTMLTextAreaElement).value)"
        :placeholder="field.placeholder || t('customField.enterValue')"
        :disabled="field.is_readonly"
        :required="field.is_required"
        rows="3"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      ></textarea>
    </template>

    <!-- 数字类型 -->
    <template v-else-if="field.field_type === 'number'">
      <input
        type="number"
        :value="localValue.number_value"
        @input="updateNumberValue(($event.target as HTMLInputElement).value)"
        :min="field.number_min"
        :max="field.number_max"
        :disabled="field.is_readonly"
        :required="field.is_required"
        :placeholder="field.placeholder"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
      <p v-if="field.number_min || field.number_max" class="text-xs text-gray-500 mt-1">
        {{ t('customField.range', { min: field.number_min ?? '-', max: field.number_max ?? '-' }) }}
      </p>
    </template>

    <!-- 下拉类型 -->
    <template v-else-if="field.field_type === 'dropdown'">
      <!-- 单选 -->
      <select
        v-if="!field.is_multi_select"
        :value="localValue.json_value?.[0] || ''"
        @change="updateDropdownValue(($event.target as HTMLSelectElement).value)"
        :disabled="field.is_readonly"
        :required="field.is_required"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      >
        <option value="">{{ t('customField.selectPlaceholder') }}</option>
        <option
          v-for="option in activeOptions"
          :key="option.id"
          :value="option.id"
        >
          {{ option.value }}
        </option>
      </select>
      <!-- 多选 -->
      <div v-else class="space-y-2">
        <label
          v-for="option in activeOptions"
          :key="option.id"
          class="flex items-center space-x-2 cursor-pointer"
        >
          <input
            type="checkbox"
            :checked="localValue.json_value?.includes(option.id)"
            @change="toggleDropdownOption(option.id)"
            :disabled="field.is_readonly"
            class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
          />
          <span
            class="px-2 py-1 rounded text-sm"
            :style="{ backgroundColor: option.color || '#E5E7EB' }"
          >
            {{ option.value }}
          </span>
        </label>
      </div>
    </template>

    <!-- 布尔类型 -->
    <template v-else-if="field.field_type === 'boolean'">
      <label class="flex items-center space-x-2 cursor-pointer">
        <input
          type="checkbox"
          :checked="localValue.boolean_value"
          @change="updateBooleanValue(($event.target as HTMLInputElement).checked)"
          :disabled="field.is_readonly"
          class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
        />
        <span class="text-sm text-gray-700">{{ localValue.boolean_value ? t('customField.yes') : t('customField.no') }}</span>
      </label>
    </template>

    <!-- 日期类型 -->
    <template v-else-if="field.field_type === 'date'">
      <input
        type="date"
        :value="localValue.date_value"
        @input="updateDateValue(($event.target as HTMLInputElement).value)"
        :disabled="field.is_readonly"
        :required="field.is_required"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
    </template>

    <!-- URL 类型 -->
    <template v-else-if="field.field_type === 'url'">
      <input
        type="url"
        :value="localValue.url_value"
        @input="updateUrlValue(($event.target as HTMLInputElement).value)"
        placeholder="https://"
        :disabled="field.is_readonly"
        :required="field.is_required"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
      <a
        v-if="localValue.url_value"
        :href="localValue.url_value"
        target="_blank"
        class="text-sm text-indigo-600 hover:underline mt-1"
      >
        {{ t('customField.openLink') }}
      </a>
    </template>

    <!-- 成员选择类型 -->
    <template v-else-if="field.field_type === 'member'">
      <select
        v-if="!field.is_multi_select"
        :value="localValue.json_value?.[0] || ''"
        @change="updateMemberValue(($event.target as HTMLSelectElement).value)"
        :disabled="field.is_readonly"
        :required="field.is_required"
        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
      >
        <option value="">{{ t('customField.selectPlaceholder') }}</option>
        <option v-for="member in members" :key="member.id" :value="member.id">
          {{ member.display_name || member.email }}
        </option>
      </select>
      <div v-else class="space-y-2">
        <label
          v-for="member in members"
          :key="member.id"
          class="flex items-center space-x-2 cursor-pointer"
        >
          <input
            type="checkbox"
            :checked="localValue.json_value?.includes(member.id)"
            @change="toggleMemberOption(member.id)"
            :disabled="field.is_readonly"
            class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
          />
          <span class="text-sm text-gray-700">
            {{ member.display_name || member.email }}
          </span>
        </label>
      </div>
    </template>

    <!-- 必填标记 -->
    <span v-if="field.is_required" class="text-red-500 ml-1">*</span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
const { t } = useI18n()
import type {
  CustomFieldLite,
  IssueCustomFieldValueUpdate
} from '@/types/custom-field'

// Props
const props = defineProps<{
  field: CustomFieldLite
  value?: IssueCustomFieldValueUpdate
  members?: Array<{ id: number; display_name?: string; email: string }>
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:value', value: IssueCustomFieldValueUpdate): void
}>()

// Local value state
const localValue = ref<IssueCustomFieldValueUpdate>({
  text_value: props.value?.text_value,
  number_value: props.value?.number_value,
  boolean_value: props.value?.boolean_value,
  date_value: props.value?.date_value,
  url_value: props.value?.url_value,
  json_value: props.value?.json_value || []
})

// Watch props value changes
watch(() => props.value, (newValue) => {
  if (newValue) {
    localValue.value = {
      text_value: newValue.text_value,
      number_value: newValue.number_value,
      boolean_value: newValue.boolean_value,
      date_value: newValue.date_value,
      url_value: newValue.url_value,
      json_value: newValue.json_value || []
    }
  }
}, { deep: true })

// Active options for dropdown
const activeOptions = computed(() => {
  return (props.field.options || []).filter((opt: any) => opt.is_active !== false)
})

// Value update methods
function updateTextValue(value: string) {
  localValue.value.text_value = value
  emit('update:value', localValue.value)
}

function updateNumberValue(value: string) {
  const num = parseFloat(value)
  localValue.value.number_value = isNaN(num) ? undefined : num
  emit('update:value', localValue.value)
}

function updateDropdownValue(value: string) {
  const numValue = parseInt(value, 10)
  localValue.value.json_value = numValue ? [numValue] : []
  emit('update:value', localValue.value)
}

function toggleDropdownOption(optionId: number) {
  const current = localValue.value.json_value || []
  const index = current.indexOf(optionId)
  if (index > -1) {
    current.splice(index, 1)
  } else {
    current.push(optionId)
  }
  localValue.value.json_value = current
  emit('update:value', localValue.value)
}

function updateBooleanValue(value: boolean) {
  localValue.value.boolean_value = value
  emit('update:value', localValue.value)
}

function updateDateValue(value: string) {
  localValue.value.date_value = value
  emit('update:value', localValue.value)
}

function updateUrlValue(value: string) {
  localValue.value.url_value = value
  emit('update:value', localValue.value)
}

function updateMemberValue(value: string) {
  const numValue = parseInt(value, 10)
  localValue.value.json_value = numValue ? [numValue] : []
  emit('update:value', localValue.value)
}

function toggleMemberOption(memberId: number) {
  const current = localValue.value.json_value || []
  const index = current.indexOf(memberId)
  if (index > -1) {
    current.splice(index, 1)
  } else {
    current.push(memberId)
  }
  localValue.value.json_value = current
  emit('update:value', localValue.value)
}
</script>

<style scoped>
.custom-field-value-input {
  width: 100%;
}
</style>