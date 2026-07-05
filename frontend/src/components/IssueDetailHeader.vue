<template>
  <div class="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200">
    <div class="flex items-center gap-3 min-w-0 flex-1">
      <!-- Back button -->
      <button
        data-test="back-btn"
        class="p-1 -ml-1 text-gray-400 hover:text-gray-600 transition-colors shrink-0"
        @click="$emit('back')"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>

      <!-- Issue type badge -->
      <span
        class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium shrink-0"
        :style="{ backgroundColor: issue?.issue_type?.color + '20', color: issue?.issue_type?.color }"
      >
        {{ issue?.issue_type?.name }}
      </span>

      <!-- Sequence ID -->
      <span class="text-xs text-gray-400 font-mono shrink-0">{{ projectIdentifier }}-{{ issue?.sequence_id }}</span>

      <!-- Title -->
      <span class="text-sm font-semibold text-gray-800 truncate">{{ issue?.name }}</span>

      <!-- Priority -->
      <span
        v-if="issue?.priority && issue.priority !== 'none'"
        class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium shrink-0"
        :class="priorityClass(issue.priority)"
      >
        <span class="w-1.5 h-1.5 rounded-full" :class="priorityDot(issue.priority)"></span>
        {{ priorityLabel(issue.priority) }}
      </span>
    </div>

    <div class="flex items-center gap-2 shrink-0">
      <!-- Copy link -->
      <button
        class="p-1.5 text-gray-400 hover:text-indigo-500 hover:bg-indigo-50 rounded transition-colors"
        :title="t('issue.copyLink')"
        @click="copyLink"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/></svg>
      </button>
      <span v-if="copied" class="text-[10px] text-green-600">{{ t('issue.copied') }}</span>

      <!-- Delete button -->
      <button
        data-test="delete-btn"
        class="px-3 py-1.5 text-xs font-medium text-red-500 border border-red-200 rounded-md hover:bg-red-50 transition-colors"
        @click="$emit('delete')"
      >
        {{ t('common.delete') }}
      </button>

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
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  issue: any
  saving: boolean
  projectIdentifier?: string
}>()

defineEmits<{
  save: []
  back: []
  delete: []
}>()

const { t } = useI18n()
const toast = useToast()
const copied = ref(false)

function copyLink() {
  const url = window.location.href
  navigator.clipboard.writeText(url).then(() => {
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  }).catch(() => {
    toast.error(t('issue.copyFailed'))
  })
}

function priorityClass(p: string) {
  const m: Record<string, string> = {
    urgent: 'bg-red-50 text-red-700', high: 'bg-orange-50 text-orange-700',
    medium: 'bg-yellow-50 text-yellow-700', low: 'bg-green-50 text-green-700',
    none: 'bg-gray-50 text-gray-600'
  }
  return m[p] || m.none
}
function priorityDot(p: string) {
  const m: Record<string, string> = {
    urgent: 'bg-red-500', high: 'bg-orange-500', medium: 'bg-yellow-500',
    low: 'bg-green-500', none: 'bg-gray-400'
  }
  return m[p] || m.none
}
function priorityLabel(p: string) {
  const m: Record<string, string> = {
    urgent: t('issue.priorityUrgent'), high: t('issue.priorityHigh'),
    medium: t('issue.priorityMedium'), low: t('issue.priorityLow'), none: t('issue.priorityNone')
  }
  return m[p] || p
}
</script>
