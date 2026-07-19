<template>
  <div v-if="show" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white rounded-xl p-6 w-full max-w-md">
      <h3 class="text-lg font-semibold mb-2">{{ t('approvals.submitTitle') }}</h3>
      <p class="text-sm text-gray-500 mb-4">
        {{ t('approvals.submitDesc', { from: fromStateName, to: approveStateName }) }}
      </p>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('approvals.approvers') }}</label>
          <div class="text-sm text-gray-700 bg-gray-50 rounded px-3 py-2">
            {{ approverNames.join(', ') || '-' }}
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('approvals.requestNote') }}</label>
          <textarea v-model="note" rows="3" class="w-full px-3 py-2 border rounded-lg"
            :placeholder="t('approvals.requestNotePlaceholder')"></textarea>
        </div>
      </div>
      <div class="flex justify-end space-x-3 mt-6">
        <button @click="$emit('close')" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
        <button @click="submit" :disabled="submitting"
          class="px-4 py-2 bg-indigo-600 text-white rounded-lg disabled:opacity-50">
          {{ submitting ? t('common.saving') : t('approvals.submit') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import approvalApi from '@/api/approval'

const props = defineProps<{
  show: boolean
  issueId: number
  transitionId: number
  fromStateName: string
  approveStateName: string
  approverNames: string[]
}>()

const emit = defineEmits<{ close: []; submitted: [] }>()
const { t } = useI18n()
const note = ref('')
const submitting = ref(false)

async function submit() {
  if (submitting.value) return
  submitting.value = true
  try {
    await approvalApi.submit(props.issueId, {
      transition_id: props.transitionId,
      request_note: note.value,
    })
    emit('submitted')
    emit('close')
  } catch (e: any) {
    alert(e?.response?.data?.message || e?.response?.data?.error || 'Failed to submit approval')
  } finally {
    submitting.value = false
  }
}
</script>
