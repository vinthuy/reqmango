<template>
  <Transition name="fade">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @click.self="$emit('close')">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-md max-h-[70vh] overflow-hidden">
        <div class="flex items-center justify-between px-5 py-3 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">⌨️ {{ t('shortcuts.title') }}</h3>
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
import { useI18n } from '@/composables/useI18n'

defineProps<{ visible: boolean }>()
defineEmits<{ (e: 'close'): void }>()

const { t } = useI18n()

interface Shortcut { key: string; label: string }
interface Group { name: string; shortcuts: Shortcut[] }

const groups: Group[] = [
  { name: t('shortcuts.navigation'), shortcuts: [
    { key: '⌘K / Ctrl+K', label: t('shortcuts.commandPalette') },
    { key: '⌘J / Ctrl+J', label: t('shortcuts.aiChat') },
    { key: '?', label: t('shortcuts.showShortcuts') },
    { key: 'Esc', label: t('shortcuts.close') },
  ]},
  { name: t('shortcuts.issues'), shortcuts: [
    { key: 'N', label: t('shortcuts.newIssue') },
    { key: 'Ctrl+Enter', label: t('shortcuts.aiSmartCreate') },
    { key: 'Enter', label: t('shortcuts.submitForm') },
  ]},
  { name: t('shortcuts.views'), shortcuts: [
    { key: 'L', label: t('shortcuts.listView') },
    { key: 'K', label: t('shortcuts.kanbanView') },
    { key: 'T', label: t('shortcuts.treeView') },
    { key: 'C', label: t('shortcuts.calendarView') },
    { key: 'G', label: t('shortcuts.ganttView') },
  ]},
  { name: t('shortcuts.ai'), shortcuts: [
    { key: '⌘J / Ctrl+J', label: t('shortcuts.toggleAiChat') },
    { key: '⌘K / Ctrl+K', label: t('shortcuts.aiCommands') },
  ]},
  { name: t('shortcuts.pages'), shortcuts: [
    { key: 'Ctrl+S', label: t('shortcuts.savePage') },
    { key: 'Ctrl+Enter', label: t('shortcuts.createPage') },
  ]},
]

const filteredGroups = computed(() => groups)
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
