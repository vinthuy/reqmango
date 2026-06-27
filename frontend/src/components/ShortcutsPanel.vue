<template>
  <Transition name="fade">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @click.self="$emit('close')">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-md max-h-[70vh] overflow-hidden">
        <div class="flex items-center justify-between px-5 py-3 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">⌨️ Keyboard Shortcuts</h3>
          <kbd class="text-xs text-gray-400 bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded" @click="$emit('close')">Esc</kbd>
        </div>
        <div class="overflow-y-auto max-h-[60vh] p-4 space-y-3">
          <div v-for="group in filteredGroups" :key="group.name">
            <h4 class="text-xs font-semibold text-gray-400 uppercase mb-2">{{ group.name }}</h4>
            <div v-for="s in group.shortcuts" :key="s.key" class="flex items-center justify-between py-1.5 text-sm">
              <span class="text-gray-700 dark:text-gray-300">{{ s.label }}</span>
              <kbd class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 px-2 py-1 rounded font-mono">{{ s.key }}</kbd>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed } from 'vue'

defineProps<{ visible: boolean }>()
defineEmits<{ (e: 'close'): void }>()

interface Shortcut { key: string; label: string }
interface Group { name: string; shortcuts: Shortcut[] }

const groups: Group[] = [
  { name: 'Navigation', shortcuts: [
    { key: '⌘K / Ctrl+K', label: 'Command Palette' },
    { key: '⌘J / Ctrl+J', label: 'AI Chat Sidebar' },
    { key: '?', label: 'Show Shortcuts' },
    { key: 'Esc', label: 'Close Panel / Modal' },
  ]},
  { name: 'Issues', shortcuts: [
    { key: 'N', label: 'New Issue (from list)' },
    { key: 'Ctrl+Enter', label: 'AI Smart Create' },
    { key: 'Enter', label: 'Submit Form' },
  ]},
  { name: 'Views', shortcuts: [
    { key: 'L', label: 'List View' },
    { key: 'K', label: 'Kanban View' },
    { key: 'T', label: 'Tree View' },
    { key: 'C', label: 'Calendar View' },
    { key: 'G', label: 'Gantt View' },
  ]},
  { name: 'AI', shortcuts: [
    { key: '⌘J / Ctrl+J', label: 'Toggle AI Chat' },
    { key: '⌘K / Ctrl+K', label: 'AI Commands' },
  ]},
  { name: 'Pages', shortcuts: [
    { key: 'Ctrl+S', label: 'Save Page (auto)' },
    { key: 'Ctrl+Enter', label: 'Create Page' },
  ]},
]

const filteredGroups = computed(() => groups)
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
