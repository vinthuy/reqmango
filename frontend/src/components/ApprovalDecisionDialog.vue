<template>
  <div v-if="show" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white rounded-xl p-6 w-full max-w-md">
      <h3 class="text-lg font-semibold mb-2">
        {{ decision === 'approved' ? t('approvals.approveTitle') : t('approvals.rejectTitle') }}
      </h3>
      <p class="text-sm text-gray-500 mb-4">{{ t('approvals.decisionDesc') }}</p>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('approvals.decisionNote') }}</label>
          <textarea v-model="note" rows="3" class="w-full px-3 py-2 border rounded-lg"
            :placeholder="t('approvals.decisionNotePlaceholder')"></textarea>
        </div>
      </div>
      <div class="flex justify-end space-x-3 mt-6">
        <button @click="$emit('close')" class="px-4 py-2 border rounded-lg">{{ t('common.cancel') }}</button>
        <button @click="confirm" :disabled="submitting"
          :class="['px-4 py-2 text-white rounded-lg disabled:opacity-50',
            decision === 'approved' ? 'bg-green-600 hover:bg-green-700' : 'bg-red-600 hover:bg-red-700']">
          {{ submitting ? t('common.saving') : (decision === 'approved' ? t('approvals.approve') : t('approvals.reject')) }}
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
  approvalId: number
  decision: 'approved' | 'rejected'
}>()

const emit = defineEmits<{ close: []; decided: [] }>()
const { t } = useI18n()
const note = ref('')
const submitting = ref(false)

async function confirm() {
  if (submitting.value) return
  submitting.value = true
  try {
    await approvalApi.decide(props.approvalId, {
      decision: props.decision,
      note: note.value,
    })
    emit('decided')
    emit('close')
  } catch (e: any) {
    alert(e?.response?.data?.message || e?.response?.data?.error || 'Failed to decide approval')
  } finally {
    submitting.value = false
  }
}
</script>
