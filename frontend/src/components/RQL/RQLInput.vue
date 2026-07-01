<!-- frontend/src/components/RQL/RQLInput.vue -->

<template>
  <div class="rql-input">
    <div class="relative">
      <input
        ref="inputRef"
        v-model="rql"
        type="text"
        :placeholder="placeholder"
        class="w-full pl-8 pr-20 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        :class="inputClass"
        @keydown.enter="handleSearch"
        @focus="showHistoryPanel = true"
        @blur="onBlur"
      />
      <svg class="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <div class="absolute right-2 top-1/2 -translate-y-1/2 flex space-x-1">
        <button
          v-if="rql"
          @click="clearRQL"
          class="p-1 text-gray-400 hover:text-gray-600"
          :title="t('rql.clear')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        <button
          v-if="showHistory"
          @click="toggleHistory"
          class="p-1 text-gray-400 hover:text-gray-600"
          :title="t('rql.historyTooltip')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="mt-1 text-xs text-red-500">
      ✗ {{ error }}
    </div>

    <!-- 提示 -->
    <div v-if="showHints && !rql" class="mt-1 text-xs text-gray-400">
      {{ t('rql.example') }}: <code class="bg-gray-100 px-1 rounded">state = "待处理" AND priority = "high"</code>
    </div>

    <!-- 历史记录面板 -->
    <RQLHistory
      v-if="showHistoryPanel"
      @select="onSelectHistory"
      @close="showHistoryPanel = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import RQLHistory from './RQLHistory.vue'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
  showHistory?: boolean
  showHints?: boolean
  error?: string | null
}>(), {
  modelValue: '',
  placeholder: t('rql.placeholder'),
  showHistory: true,
  showHints: true,
  error: null
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'search': [value: string]
}>()

const inputRef = ref<HTMLInputElement | null>(null)
const showHistoryPanel = ref(false)
const rql = ref(props.modelValue)

const inputClass = computed(() => {
  if (!rql.value) return 'border-gray-300'
  if (props.error) return 'border-red-300'
  return 'border-green-300'
})

watch(() => props.modelValue, (val) => {
  rql.value = val
})

watch(rql, (val) => {
  emit('update:modelValue', val)
})

const handleSearch = () => {
  emit('search', rql.value)
}

const clearRQL = () => {
  rql.value = ''
  emit('update:modelValue', '')
}

const toggleHistory = () => {
  showHistoryPanel.value = !showHistoryPanel.value
}

const onSelectHistory = (item: any) => {
  rql.value = item.rql
  showHistoryPanel.value = false
  emit('update:modelValue', item.rql)
  emit('search', item.rql)
}

const onBlur = () => {
  setTimeout(() => {
    showHistoryPanel.value = false
  }, 200)
}

const focus = () => {
  inputRef.value?.focus()
}

defineExpose({ focus })
</script>