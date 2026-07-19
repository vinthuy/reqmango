<template>
  <div class="p-6 max-w-7xl mx-auto">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">{{ t('approvals.listTitle') }}</h1>

    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
      <!-- Filter bar -->
      <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center gap-2">
        <label class="text-sm text-gray-600 dark:text-gray-300">{{ t('approvals.status') }}</label>
        <select v-model="filter.status" @change="load()"
          class="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
          <option value="">{{ t('common.all') }}</option>
          <option value="pending">{{ t('approvals.pending') }}</option>
          <option value="approved">{{ t('approvals.approved') }}</option>
          <option value="rejected">{{ t('approvals.rejected') }}</option>
          <option value="cancelled">{{ t('approvals.cancelled') }}</option>
        </select>
      </div>

      <!-- Table -->
      <table class="w-full">
        <thead class="bg-gray-50 dark:bg-gray-700/50 text-xs text-gray-500 dark:text-gray-400 uppercase">
          <tr>
            <th class="px-4 py-2 text-left">{{ t('approvals.issue') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.project') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.requester') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.submittedTime') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.status') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.action') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in approvals" :key="a.id"
            class="hover:bg-gray-50 dark:hover:bg-gray-700/30 border-t border-gray-100 dark:border-gray-700">
            <td class="px-4 py-2">
              <router-link :to="`/workspace/${slug}/project/${a.project_id}/issues/${a.issue_id}`"
                class="text-indigo-600 dark:text-indigo-400 hover:underline">
                {{ a.issue_key }}: {{ a.issue_title }}
              </router-link>
            </td>
            <td class="px-4 py-2 text-gray-900 dark:text-gray-100">{{ a.project_name }}</td>
            <td class="px-4 py-2 text-gray-900 dark:text-gray-100">{{ a.requester_name }}</td>
            <td class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">{{ new Date(a.created_at).toLocaleString() }}</td>
            <td class="px-4 py-2">
              <span :class="['inline-flex px-2 py-0.5 rounded text-xs font-medium', statusClass(a.status)]">
                {{ statusLabel(a.status) }}
              </span>
            </td>
            <td class="px-4 py-2">
              <div v-if="canDecide(a)" class="flex gap-2">
                <button @click="decide(a, 'approved')"
                  class="px-3 py-1 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded">
                  {{ t('approvals.approve') }}
                </button>
                <button @click="decide(a, 'rejected')"
                  class="px-3 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded">
                  {{ t('approvals.reject') }}
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="approvals.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
              {{ t('approvals.noApprovals') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <ApprovalDecisionDialog
      v-if="decisionData"
      :show="showDecisionDialog"
      :approval-id="decisionData.approvalId"
      :decision="decisionData.decision"
      @close="showDecisionDialog = false"
      @decided="onDecided"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useAuthStore } from '@/stores/auth'
import { workspaceApi } from '@/api/workspace'
import approvalApi, { type ApprovalResponse } from '@/api/approval'
import ApprovalDecisionDialog from '@/components/ApprovalDecisionDialog.vue'

const route = useRoute()
const { t } = useI18n()
const authStore = useAuthStore()

const slug = computed(() => route.params.slug as string)
const workspaceId = ref(0)
const approvals = ref<ApprovalResponse[]>([])
const filter = ref({ status: '' })
const showDecisionDialog = ref(false)
const decisionData = ref<{ approvalId: number; decision: 'approved' | 'rejected' } | null>(null)
const loading = ref(false)

async function resolveWorkspaceId() {
  try {
    const list = await workspaceApi.list()
    const ws = (list || []).find((w: any) => w.slug === slug.value)
    workspaceId.value = ws?.id || 0
  } catch (e) {
    console.error('Failed to resolve workspace', e)
  }
}

async function load() {
  if (!workspaceId.value) return
  loading.value = true
  try {
    const list = await approvalApi.listByWorkspace(workspaceId.value, {
      status: filter.value.status || undefined,
    })
    approvals.value = Array.isArray(list) ? list : []
  } catch (e) {
    console.error('Failed to load approvals', e)
    approvals.value = []
  } finally {
    loading.value = false
  }
}

function canDecide(a: ApprovalResponse): boolean {
  return a.status === 'pending' && a.approver_ids.includes(authStore.user?.id || 0)
}

function decide(a: ApprovalResponse, decision: 'approved' | 'rejected') {
  decisionData.value = { approvalId: a.id, decision }
  showDecisionDialog.value = true
}

async function onDecided() {
  showDecisionDialog.value = false
  await load()
}

function statusClass(s: string): string {
  return ({
    pending: 'bg-yellow-100 text-yellow-700',
    approved: 'bg-green-100 text-green-700',
    rejected: 'bg-red-100 text-red-700',
    cancelled: 'bg-gray-100 text-gray-500',
  } as Record<string, string>)[s] || 'bg-gray-100 text-gray-500'
}

function statusLabel(s: string): string {
  return t(`approvals.${s}`)
}

onMounted(async () => {
  await resolveWorkspaceId()
  await load()
})
</script>
