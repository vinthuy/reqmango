<template>
  <div class="relative" ref="dropdownRef">
    <button
      @click.stop="isOpen = !isOpen"
      class="text-sm bg-white border border-indigo-300 rounded px-2 py-1 outline-none min-w-[120px] text-left flex items-center justify-between gap-1"
      :class="{ 'border-indigo-500 ring-1 ring-indigo-500': isOpen }"
    >
      <span class="truncate">
        {{ selectedLabels.length > 0 ? selectedLabels.join(', ') : placeholder }}
      </span>
      <svg class="w-4 h-4 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
    
    <div
      v-if="isOpen"
      class="absolute top-full left-0 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg z-50 max-h-60 overflow-y-auto"
    >
      <div class="p-1">
        <label
          v-for="option in options"
          :key="option.value"
          class="flex items-center gap-2 px-2 py-1.5 hover:bg-gray-100 rounded cursor-pointer"
        >
          <input
            type="checkbox"
            :checked="modelValue.includes(option.value)"
            @change="toggleOption(option.value)"
            class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
          />
          <span class="text-sm text-gray-700">{{ option.label }}</span>
        </label>
      </div>
      
      <div v-if="options.length === 0" class="px-3 py-2 text-sm text-gray-500 text-center">
        {{ emptyText }}
      </div>
      
      <div v-if="modelValue.length > 0" class="border-t border-gray-100 p-1">
        <button
          @click="clearAll"
          class="w-full text-left px-2 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded"
        >
          {{ clearText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Option {
  value: string | number
  label: string
}

const props = withDefaults(defineProps<{
  modelValue: (string | number)[]
  options: Option[]
  placeholder?: string
  emptyText?: string
  clearText?: string
}>(), {
  placeholder: 'Select...',
  emptyText: 'No options',
  clearText: 'Clear all'
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: (string | number)[]): void
}>()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const selectedLabels = computed(() => {
  return props.modelValue
    .map(v => props.options.find(o => o.value === v)?.label)
    .filter(Boolean)
})

function toggleOption(value: string | number) {
  const newValue = [...props.modelValue]
  const index = newValue.indexOf(value)
  if (index >= 0) {
    newValue.splice(index, 1)
  } else {
    newValue.push(value)
  }
  emit('update:modelValue', newValue)
}

function clearAll() {
  emit('update:modelValue', [])
  isOpen.value = false
}

function handleClickOutside(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
