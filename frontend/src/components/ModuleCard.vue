<template>
  <div
    class="module-card bg-white border border-gray-200 rounded-lg p-4 hover:border-indigo-300 hover:shadow-sm cursor-pointer transition-all"
    @click="$emit('click')"
  >
    <!-- 头部：标题 -->
    <div class="flex items-start justify-between mb-3">
      <div class="flex items-center space-x-2">
        <svg class="w-5 h-5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <h3 class="text-base font-semibold text-gray-900">{{ module.name }}</h3>
      </div>
    </div>

    <!-- 描述 -->
    <p v-if="module.description" class="text-sm text-gray-500 mb-3 line-clamp-2">
      {{ module.description }}
    </p>

    <!-- Footer -->
    <div class="flex items-center justify-between pt-3 border-t border-gray-100 text-xs text-gray-400">
      <span v-if="module.parent_id">{{ t('moduleCard.submodule') }}</span>
      <span v-else>{{ t('moduleCard.topLevel') }}</span>
      <div class="relative" @click.stop>
        <button
          @click="showMenu = !showMenu"
          class="p-1 text-gray-400 hover:text-gray-600 rounded"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
          </svg>
        </button>

        <div
          v-if="showMenu"
          class="absolute right-0 mt-1 w-28 bg-white border border-gray-200 rounded-md shadow-lg z-10"
        >
          <button
            @click="$emit('click'); showMenu = false"
            class="w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
          >
            {{ t('moduleCard.viewDetails') }}
          </button>
          <button
            @click="$emit('delete', module); showMenu = false"
            class="w-full px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
          >
            {{ t('moduleCard.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ModuleResponse } from '@/types/module'

const { t } = useI18n()

// Props
defineProps<{
  module: ModuleResponse
}>()

// Emits
defineEmits<{
  (e: 'click'): void
  (e: 'delete', module: ModuleResponse): void
}>()

// State
const showMenu = ref(false)
</script>

<style scoped>
.module-card:hover {
  transform: translateY(-2px);
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>