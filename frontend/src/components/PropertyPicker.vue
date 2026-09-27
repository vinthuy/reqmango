<template>
  <div class="min-w-0">
    <button
      ref="triggerRef"
      type="button"
      class="picker-trigger"
      :class="{ 'picker-trigger--open': open }"
      :disabled="disabled"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggle"
      @keydown.down.prevent="openPicker"
    >
      <template v-if="selected">
        <span v-if="selected.color" class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: selected.color }" />
        <span v-else-if="selected.icon" class="shrink-0 text-xs leading-none">{{ selected.icon }}</span>
        <span class="truncate">{{ selected.label }}</span>
      </template>
      <span v-else class="truncate text-gray-400">{{ placeholder || t('common.none') }}</span>
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        ref="panelRef"
        class="fixed z-[1000] bg-white border border-gray-200 rounded-lg shadow-lg overflow-hidden"
        :style="panelStyle"
      >
        <div class="p-1.5 border-b border-gray-100">
          <input
            ref="searchRef"
            v-model="query"
            type="text"
            class="w-full px-2 py-1 text-sm rounded border border-transparent focus:outline-none focus:border-indigo-300 bg-gray-50"
            :placeholder="t('common.search')"
            @keydown="onSearchKeydown"
          />
        </div>
        <ul ref="listRef" role="listbox" class="max-h-60 overflow-y-auto py-1">
          <li
            v-for="(item, i) in items"
            :key="String(item.value)"
            role="option"
            :aria-selected="item.value === normalizedValue"
            class="flex items-center gap-2 px-2.5 py-1.5 text-sm cursor-pointer"
            :class="i === activeIndex ? 'bg-indigo-50 text-indigo-900' : 'text-gray-700'"
            @mouseenter="activeIndex = i"
            @mousedown.prevent="pick(item.value)"
          >
            <span v-if="item.color" class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: item.color }" />
            <span v-else-if="item.icon" class="shrink-0 text-xs leading-none">{{ item.icon }}</span>
            <span class="flex-1 truncate" :class="{ 'text-gray-400': item.value === null }">{{ item.label }}</span>
            <svg v-if="item.value === normalizedValue" class="w-3.5 h-3.5 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.7 5.3a1 1 0 010 1.4l-8 8a1 1 0 01-1.4 0l-4-4a1 1 0 011.4-1.4L8 12.6l7.3-7.3a1 1 0 011.4 0z" clip-rule="evenodd" />
            </svg>
          </li>
          <li v-if="!items.length" class="px-2.5 py-2 text-xs text-gray-400">{{ t('common.noResults') }}</li>
        </ul>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { useI18n } from '@/composables/useI18n'

export interface PickerOption {
  value: string | number
  label: string
  color?: string
  icon?: string
}

type PickerValue = string | number | null

const props = defineProps<{
  modelValue: PickerValue | undefined
  options: PickerOption[]
  placeholder?: string
  disabled?: boolean
  clearable?: boolean
  clearLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: PickerValue): void
}>()

const { t } = useI18n()

const PANEL_MAX_HEIGHT = 300
const PANEL_MIN_WIDTH = 220

const open = ref(false)
const query = ref('')
const activeIndex = ref(0)
const triggerRef = ref<HTMLButtonElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const searchRef = ref<HTMLInputElement | null>(null)
const listRef = ref<HTMLElement | null>(null)
const panelStyle = ref<Record<string, string>>({})

const normalizedValue = computed<PickerValue>(() =>
  props.modelValue === undefined || props.modelValue === '' ? null : props.modelValue
)

const selected = computed(() =>
  normalizedValue.value === null ? null : props.options.find((o) => o.value === normalizedValue.value) ?? null
)

const items = computed(() => {
  const q = query.value.trim().toLowerCase()
  const matched = q ? props.options.filter((o) => o.label.toLowerCase().includes(q)) : props.options
  const list: Array<Omit<PickerOption, 'value'> & { value: PickerValue }> = [...matched]
  if (props.clearable && !q) list.unshift({ value: null, label: props.clearLabel || t('common.none') })
  return list
})

watch(query, () => { activeIndex.value = 0 })

function positionPanel() {
  const el = triggerRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const width = Math.max(rect.width, PANEL_MIN_WIDTH)
  const left = Math.min(rect.left, window.innerWidth - width - 8)
  const style: Record<string, string> = { left: `${Math.max(8, left)}px`, width: `${width}px` }
  if (window.innerHeight - rect.bottom < PANEL_MAX_HEIGHT && rect.top > window.innerHeight - rect.bottom) {
    style.bottom = `${window.innerHeight - rect.top + 4}px`
  } else {
    style.top = `${rect.bottom + 4}px`
  }
  panelStyle.value = style
}

function openPicker() {
  if (props.disabled || open.value) return
  query.value = ''
  const idx = items.value.findIndex((i) => i.value === normalizedValue.value)
  activeIndex.value = idx >= 0 ? idx : 0
  positionPanel()
  open.value = true
  nextTick(() => {
    searchRef.value?.focus()
    scrollActiveIntoView()
  })
}

function close(refocus = false) {
  if (!open.value) return
  open.value = false
  if (refocus) nextTick(() => triggerRef.value?.focus())
}

function toggle() {
  if (open.value) close()
  else openPicker()
}

function pick(value: PickerValue) {
  close(true)
  if (value !== normalizedValue.value) emit('update:modelValue', value)
}

function scrollActiveIntoView() {
  const el = listRef.value?.children[activeIndex.value] as HTMLElement | undefined
  el?.scrollIntoView?.({ block: 'nearest' })
}

function move(delta: number) {
  const n = items.value.length
  if (!n) return
  activeIndex.value = (activeIndex.value + delta + n) % n
  nextTick(scrollActiveIntoView)
}

function onSearchKeydown(e: KeyboardEvent) {
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      move(1)
      break
    case 'ArrowUp':
      e.preventDefault()
      move(-1)
      break
    case 'Enter': {
      e.preventDefault()
      const item = items.value[activeIndex.value]
      if (item) pick(item.value)
      break
    }
    case 'Escape':
      e.preventDefault()
      e.stopPropagation()
      close(true)
      break
    case 'Tab':
      close()
      break
  }
}

function onDocumentPointerDown(e: MouseEvent) {
  const target = e.target as Node
  if (triggerRef.value?.contains(target) || panelRef.value?.contains(target)) return
  close()
}

function onViewportChange(e: Event) {
  if (e.type === 'scroll' && panelRef.value?.contains(e.target as Node)) return
  close()
}

watch(open, (isOpen) => {
  if (isOpen) {
    document.addEventListener('mousedown', onDocumentPointerDown, true)
    window.addEventListener('scroll', onViewportChange, true)
    window.addEventListener('resize', onViewportChange)
  } else {
    removeListeners()
  }
})

function removeListeners() {
  document.removeEventListener('mousedown', onDocumentPointerDown, true)
  window.removeEventListener('scroll', onViewportChange, true)
  window.removeEventListener('resize', onViewportChange)
}

onBeforeUnmount(removeListeners)
</script>

<style scoped>
.picker-trigger {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  width: 100%;
  min-height: 1.75rem;
  padding: 0.25rem 0.375rem;
  border: 1px solid transparent;
  border-radius: 0.25rem;
  font-size: 0.8125rem;
  line-height: 1.25rem;
  color: #374151;
  text-align: left;
  background: transparent;
  transition: background-color 0.1s, border-color 0.1s;
}
.picker-trigger:hover:not(:disabled),
.picker-trigger--open {
  background: #f3f4f6;
}
.picker-trigger:focus-visible {
  outline: none;
  border-color: #a5b4fc;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}
.picker-trigger:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
