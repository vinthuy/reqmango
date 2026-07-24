<template>
  <div v-if="approval" class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <div class="flex items-center space-x-2 mb-1">
          <span class="text-yellow-700 font-medium">{{ t('approvals.pendingBanner') }}</span>
          <span class="px-2 py-0.5 bg-yellow-100 text-yellow-700 rounded text-xs">
            {{ t('approvals.pending') }}
          </span>
        </div>
        <p class="text-sm text-yellow-700">
          {{ t('approvals.requestedBy', { name: approval.requester_name }) }}
        </p>
        <p class="text-sm text-yellow-700">
          {{ t('approvals.submittedAt', { time: new Date(approval.created_at).toLocaleString() }) }}
        </p>
        <p class="text-sm text-yellow-700 mt-1">
          {{ t('approvals.approvers') }}: {{ approval.approver_names.join(', ') }}
        </p>
        <p v-if="approval.request_note" class="text-sm text-yellow-700 mt-1 italic">
          "{{ approval.request_note }}"
        </p>
      </div>
      <div class="flex items-center space-x-2">
        <button v-if="canDecide" @click="$emit('decide', approval, 'approved')"
          class="px-3 py-1.5 bg-green-600 text-white rounded text-sm hover:bg-green-700">
          {{ t('approvals.approve') }}
        </button>
        <button v-if="canDecide" @click="$emit('decide', approval, 'rejected')"
          class="px-3 py-1.5 bg-red-600 text-white rounded text-sm hover:bg-red-700">
          {{ t('approvals.reject') }}
        </button>
        <button v-if="canCancel" @click="$emit('cancel', approval)"
          class="px-3 py-1.5 border border-yellow-300 text-yellow-700 rounded text-sm hover:bg-yellow-100">
          {{ t('approvals.cancelApproval') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { ApprovalResponse } from '@/api/approval'

const props = defineProps<{
  approval: ApprovalResponse | null
  currentUserId: number
}>()

defineEmits<{
  decide: [approval: ApprovalResponse, decision: 'approved' | 'rejected']
  cancel: [approval: ApprovalResponse]
}>()

const { t } = useI18n()

const canDecide = computed(() =>
  props.approval?.status === 'pending' &&
  props.approval.approver_ids.includes(props.currentUserId)
)

const canCancel = computed(() =>
  props.approval?.status === 'pending' &&
  props.approval.requester_id === props.currentUserId
)
</script>
