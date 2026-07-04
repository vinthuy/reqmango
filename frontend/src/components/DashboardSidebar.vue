<template>
  <div class="sidebar w-60 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col shrink-0">
    <div class="px-4 py-3 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
      <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">{{ t('dashboard.customAnalytics') }}</h3>
      <button @click="$emit('create')"
        class="w-6 h-6 rounded-md bg-gray-100 dark:bg-gray-700 flex items-center justify-center hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
        :title="t('dashboard.createFirst')">
        <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto py-1">
      <div v-if="dashboards.length === 0" class="px-4 py-6 text-xs text-gray-400 dark:text-gray-500 text-center">
        {{ t('dashboard.noDashboards') }}
      </div>
      <div v-for="d in dashboards" :key="d.id"
        class="dashboard-item group flex items-center gap-2 px-4 py-2.5 cursor-pointer text-sm transition-colors border-l-[3px]"
        :class="d.id === currentId
          ? 'border-l-indigo-500 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-400 font-medium'
          : 'border-l-transparent text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700/50'"
        @click="$emit('select', d.id)"
      >
        <span class="w-2 h-2 rounded-full shrink-0"
          :class="d.id === currentId ? 'bg-indigo-500' : 'bg-gray-300 dark:bg-gray-600'" />
        <span class="flex-1 truncate">{{ d.name }}</span>

        <!-- Actions dropdown -->
        <div class="relative" @click.stop>
          <button @click="toggleMenu(d.id)"
            class="w-6 h-6 rounded flex items-center justify-center opacity-0 group-hover:opacity-100 hover:bg-gray-200 dark:hover:bg-gray-600 transition-all">
            <svg class="w-3.5 h-3.5 text-gray-400 dark:text-gray-500" fill="currentColor" viewBox="0 0 20 20">
              <path d="M6 10a2 2 0 11-4 0 2 2 0 014 0zM12 10a2 2 0 11-4 0 2 2 0 014 0zM16 12a2 2 0 100-4 2 2 0 000 4z" />
            </svg>
          </button>
          <div v-if="menuOpen === d.id"
            class="absolute left-0 top-full mt-1 bg-white dark:bg-gray-700 rounded-lg shadow-lg border border-gray-200 dark:border-gray-600 py-1 w-36 z-50">
            <button @click="startRename(d); menuOpen = null"
              class="w-full text-left px-3 py-1.5 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600">
              {{ t('common.rename') }}
            </button>
            <button @click="$emit('duplicate', d.id); menuOpen = null"
              class="w-full text-left px-3 py-1.5 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600">
              {{ t('dashboard.duplicate') }}
            </button>
            <button @click="$emit('delete', d.id); menuOpen = null"
              class="w-full text-left px-3 py-1.5 text-xs text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20">
              {{ t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Rename modal -->
    <div v-if="renaming" class="fixed inset-0 z-[300] flex items-center justify-center"
      @click.self="cancelRename">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-5 w-80">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">{{ t('common.rename') }}</h4>
        <input v-model="renameValue" ref="renameInput" @keyup.enter="confirmRename" @keyup.escape="cancelRename"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          autofocus />
        <div class="flex justify-end gap-2 mt-4">
          <button class="btn btn-ghost btn-sm" @click="cancelRename">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary btn-sm" @click="confirmRename">{{ t('common.save') }}</button>
        </div>
      </div>
    </div>
    <div v-if="renaming" class="fixed inset-0 z-[299] bg-black/20" @click="cancelRename"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { Dashboard } from '@/types/dashboard'

const { t } = useI18n()

defineProps<{
  dashboards: Dashboard[]
  currentId: number | null
}>()

const emit = defineEmits<{
  select: [id: number]
  create: []
  delete: [id: number]
  duplicate: [id: number]
  rename: [id: number, name: string]
}>()

const menuOpen = ref<number | null>(null)
const renaming = ref(false)
const renameValue = ref('')
const renameTarget = ref<Dashboard | null>(null)
const renameInput = ref<HTMLInputElement>()

function toggleMenu(id: number) {
  menuOpen.value = menuOpen.value === id ? null : id
}

function startRename(d: Dashboard) {
  renameTarget.value = d
  renameValue.value = d.name
  renaming.value = true
  nextTick(() => renameInput.value?.focus())
}

function confirmRename() {
  if (renameTarget.value && renameValue.value.trim()) {
    emit('rename', renameTarget.value.id, renameValue.value.trim())
  }
  renaming.value = false
  renameTarget.value = null
}

function cancelRename() {
  renaming.value = false
  renameTarget.value = null
}
</script>

<style scoped>
.btn {
  padding: 6px 14px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none;
}
.btn-ghost { background: transparent; color: #6b7280; border: 1px solid #d1d5db; }
.btn-ghost:hover { background: #f3f4f6; color: #374151; }
.btn-primary { background: #111827; color: #fff; }
.btn-primary:hover { background: #1f2937; }
.btn-sm { padding: 4px 10px; font-size: 12px; }
</style>
