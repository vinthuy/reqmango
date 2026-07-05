<template>
  <div class="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200">
    <div class="flex items-center gap-3">
      <!-- Back button -->
      <button
        data-test="back-btn"
        class="p-1 -ml-1 text-gray-400 hover:text-gray-600 transition-colors"
        @click="$emit('back')"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>

      <!-- Issue type badge -->
      <span
        class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
        :style="{ backgroundColor: issue?.issue_type?.color + '20', color: issue?.issue_type?.color }"
      >
        {{ issue?.issue_type?.name }}
      </span>

      <!-- Sequence ID -->
      <span class="text-sm font-medium text-gray-700">
        #{{ issue?.sequence_id }}
      </span>
    </div>

    <!-- Save button -->
    <button
      data-test="save-btn"
      class="px-4 py-1.5 text-sm font-medium bg-neutral-900 text-white rounded-md hover:bg-neutral-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      :disabled="saving"
      @click="$emit('save')"
    >
      {{ saving ? t('issue.saving') : t('issue.save') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'

defineProps<{
  issue: any
  saving: boolean
}>()

defineEmits<{
  save: []
  back: []
}>()

const { t } = useI18n()
</script>
