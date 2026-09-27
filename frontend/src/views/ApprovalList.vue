<template>
  <div class="p-6 max-w-7xl mx-auto" data-testid="approval-center">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ t('approvals.listTitle') }}</h1>
      <button @click="load()"
        class="px-3 py-1.5 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700">
        {{ t('common.refresh') }}
      </button>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
      <button v-for="s in stats" :key="s.scope" @click="setScope(s.scope)"
        :data-testid="`approval-stat-${s.scope}`"
        :class="['text-left p-4 rounded-xl border transition-colors',
          scope === s.scope ? 'border-indigo-400 bg-indigo-50 dark:bg-indigo-900/20' : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 hover:border-indigo-300']">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ s.label }}</div>
        <div class="text-2xl font-semibold text-gray-900 dark:text-gray-100 mt-1">{{ s.value }}</div>
      </button>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
      <!-- Scope tabs + filters -->
      <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex flex-wrap items-center gap-3">
        <div class="flex rounded-lg bg-gray-100 dark:bg-gray-700 p-0.5">
          <button v-for="sc in scopes" :key="sc.value" @click="setScope(sc.value)"
            :data-testid="`approval-scope-${sc.value}`"
            :class="['px-3 py-1 text-sm rounded-md transition-colors',
              scope === sc.value ? 'bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 shadow-sm font-medium' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700']">
            {{ sc.label }}
          </button>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-sm text-gray-600 dark:text-gray-300">{{ t('approvals.status') }}</label>
          <select v-model="filter.status" @change="load()" data-testid="approval-status-filter"
            class="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
            <option value="">{{ t('common.all') }}</option>
            <option value="pending">{{ t('approvals.pending') }}</option>
            <option value="approved">{{ t('approvals.approved') }}</option>
            <option value="rejected">{{ t('approvals.rejected') }}</option>
            <option value="cancelled">{{ t('approvals.cancelled') }}</option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-sm text-gray-600 dark:text-gray-300">{{ t('approvals.project') }}</label>
          <select v-model.number="filter.projectId" @change="load()" data-testid="approval-project-filter"
            class="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
            <option :value="0">{{ t('approvals.allProjects') }}</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
      </div>

      <!-- Table -->
      <table class="w-full">
        <thead class="bg-gray-50 dark:bg-gray-700/50 text-xs text-gray-500 dark:text-gray-400 uppercase">
          <tr>
            <th class="px-4 py-2 text-left">{{ t('approvals.issue') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.project') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.transition') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.requester') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.submittedTime') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.status') }}</th>
            <th class="px-4 py-2 text-left">{{ t('approvals.action') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</td>
          </tr>
          <template v-else v-for="a in visibleApprovals" :key="a.id">
            <tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30 border-t border-gray-100 dark:border-gray-700" data-testid="approval-row">
              <td class="px-4 py-2">
                <router-link :to="`/workspace/${slug}/project/${a.project_id}/issues/${a.issue_id}`"
                  class="text-indigo-600 dark:text-indigo-400 hover:underline">
                  {{ a.issue_key }}: {{ a.issue_title }}
                </router-link>
              </td>
              <td class="px-4 py-2 text-gray-900 dark:text-gray-100">{{ a.project_name }}</td>
              <td class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 whitespace-nowrap">
                {{ a.source_state_name || '—' }} → {{ a.approve_target_state_name || '—' }}
              </td>
              <td class="px-4 py-2 text-gray-900 dark:text-gray-100">{{ a.requester_name }}</td>
              <td class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">{{ new Date(a.created_at).toLocaleString() }}</td>
              <td class="px-4 py-2">
                <span :class="['inline-flex px-2 py-0.5 rounded text-xs font-medium', statusClass(a.status)]">
                  {{ statusLabel(a.status) }}
                </span>
              </td>
              <td class="px-4 py-2">
                <div class="flex flex-wrap gap-2">
                  <button @click="toggleExpand(a.id)" data-testid="approval-toggle-details"
                    class="px-3 py-1 text-xs font-medium text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 rounded transition-colors">
                    {{ expanded.has(a.id) ? t('approvals.hideDetails') : t('approvals.details') }}
                  </button>
                  <template v-if="canDecide(a)">
                    <button @click="decide(a, 'approved')" data-testid="approval-approve"
                      class="px-3 py-1 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded transition-colors"
                      :title="t('approvals.approveHint')">
                      {{ t('approvals.approve') }}
                    </button>
                    <button @click="decide(a, 'rejected')" data-testid="approval-reject"
                      class="px-3 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 rounded transition-colors"
                      :title="t('approvals.rejectHint')">
                      {{ t('approvals.reject') }}
                    </button>
                  </template>
                  <button v-if="canCancel(a)" @click="cancel(a)" data-testid="approval-cancel"
                    class="px-3 py-1 text-xs font-medium text-red-600 border border-red-200 hover:bg-red-50 rounded transition-colors">
                    {{ t('approvals.cancelApproval') }}
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="expanded.has(a.id)" class="bg-gray-50/60 dark:bg-gray-900/30" data-testid="approval-details">
              <td colspan="7" class="px-6 py-4">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                  <div>
                    <div class="text-xs uppercase text-gray-400 mb-1">{{ t('approvals.requestNote') }}</div>
                    <div class="text-gray-800 dark:text-gray-200 whitespace-pre-wrap">{{ a.request_note || t('approvals.noRequestNote') }}</div>
                  </div>
                  <div>
                    <div class="text-xs uppercase text-gray-400 mb-1">{{ t('approvals.approvers') }}</div>
                    <div class="text-gray-800 dark:text-gray-200">{{ (a.approver_names || []).join('、') || '—' }}</div>
                    <div class="text-xs uppercase text-gray-400 mt-3 mb-1">{{ t('approvals.transition') }}</div>
                    <div class="text-gray-800 dark:text-gray-200">
                      {{ a.source_state_name || '—' }} → {{ a.approve_target_state_name || '—' }}
                      <span v-if="a.reject_target_state_name" class="text-gray-400">（{{ t('approvals.reject') }} → {{ a.reject_target_state_name }}）</span>
                    </div>
                  </div>
                  <div>
                    <div class="text-xs uppercase text-gray-400 mb-1">{{ t('approvals.records') }}</div>
                    <ul v-if="recordsOf(a).length" class="space-y-2">
                      <li v-for="r in recordsOf(a)" :key="r.id" class="text-gray-800 dark:text-gray-200">
                        <span :class="['inline-flex px-1.5 py-0.5 rounded text-xs font-medium mr-1', statusClass(r.decision)]">{{ statusLabel(r.decision) }}</span>
                        {{ t('approvals.decidedBy', { name: r.approver_name, time: new Date(r.decided_at).toLocaleString() }) }}
                        <div v-if="r.note" class="text-gray-500 dark:text-gray-400 mt-0.5">{{ r.note }}</div>
                      </li>
                    </ul>
                    <div v-else class="text-gray-500 dark:text-gray-400">{{ t('approvals.noRecords') }}</div>
                  </div>
                </div>
              </td>
            </tr>
          </template>
          <tr v-if="!loading && visibleApprovals.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
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
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { useAuthStore } from '@/stores/auth'
import { workspaceApi } from '@/api/workspace'
import { projectApi } from '@/api/project'
import approvalApi, { type ApprovalResponse } from '@/api/approval'
import ApprovalDecisionDialog from '@/components/ApprovalDecisionDialog.vue'

type Scope = 'mine' | 'requested' | 'all'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const toast = useToast()
const { confirm } = useConfirm()
const authStore = useAuthStore()

const slug = computed(() => route.params.slug as string)
const workspaceId = ref(0)
const approvals = ref<ApprovalResponse[]>([])
const projects = ref<{ id: number; name: string }[]>([])
const allForStats = ref<ApprovalResponse[]>([])
const scope = ref<Scope>((['mine', 'requested', 'all'] as Scope[]).includes(route.query.scope as Scope) ? route.query.scope as Scope : 'mine')
const filter = ref({ status: scope.value === 'mine' ? 'pending' : '', projectId: 0 })
const expanded = ref(new Set<number>())
const showDecisionDialog = ref(false)
const decisionData = ref<{ approvalId: number; decision: 'approved' | 'rejected' } | null>(null)
const loading = ref(false)

const userId = computed(() => authStore.user?.id || 0)

const scopes = computed(() => [
  { value: 'mine' as Scope, label: t('approvals.scopeMine') },
  { value: 'requested' as Scope, label: t('approvals.scopeRequested') },
  { value: 'all' as Scope, label: t('approvals.scopeAll') },
])

const stats = computed(() => [
  { scope: 'mine' as Scope, label: t('approvals.statPendingMine'),
    value: allForStats.value.filter(a => a.status === 'pending' && a.approver_ids.includes(userId.value)).length },
  { scope: 'requested' as Scope, label: t('approvals.statRequestedPending'),
    value: allForStats.value.filter(a => a.status === 'pending' && a.requester_id === userId.value).length },
  { scope: 'all' as Scope, label: t('approvals.statDecided'),
    value: allForStats.value.filter(a => a.status === 'approved' || a.status === 'rejected').length },
])

const visibleApprovals = computed(() =>
  scope.value === 'requested' ? approvals.value.filter(a => a.requester_id === userId.value) : approvals.value,
)

async function resolveWorkspaceId() {
  try {
    const list = await workspaceApi.list()
    const ws = (list || []).find((w: any) => w.slug === slug.value)
    workspaceId.value = ws?.id || 0
  } catch (e) {
    console.error('Failed to resolve workspace', e)
  }
}

async function loadProjects() {
  if (!workspaceId.value) return
  try {
    const list = await projectApi.listProjects(workspaceId.value)
    projects.value = (list || []).map((p: any) => ({ id: p.id, name: p.name }))
  } catch {
    projects.value = []
  }
}

async function load() {
  if (!workspaceId.value) return
  loading.value = true
  try {
    const [list, all] = await Promise.all([
      approvalApi.listByWorkspace(workspaceId.value, {
        status: filter.value.status || undefined,
        project_id: filter.value.projectId || undefined,
        approver_id: scope.value === 'mine' ? userId.value : undefined,
      }),
      approvalApi.listByWorkspace(workspaceId.value, {
        project_id: filter.value.projectId || undefined,
      }),
    ])
    approvals.value = Array.isArray(list) ? list : []
    allForStats.value = Array.isArray(all) ? all : []
  } catch (e) {
    console.error('Failed to load approvals', e)
    approvals.value = []
  } finally {
    loading.value = false
  }
}

function setScope(s: Scope) {
  if (scope.value === s) return
  scope.value = s
  filter.value.status = s === 'mine' ? 'pending' : ''
  router.replace({ query: { ...route.query, scope: s } })
  load()
}

function toggleExpand(id: number) {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}

function recordsOf(a: ApprovalResponse) {
  if (a.records?.length) return a.records
  if ((a.status === 'approved' || a.status === 'rejected') && a.decided_at) {
    return [{
      id: 0,
      approver_id: a.decided_by ?? 0,
      approver_name: a.decided_by_name || '—',
      decision: a.status,
      note: a.decision_note || '',
      decided_at: a.decided_at,
    }]
  }
  return []
}

function canDecide(a: ApprovalResponse): boolean {
  return a.status === 'pending' && a.approver_ids.includes(userId.value)
}

function canCancel(a: ApprovalResponse): boolean {
  return a.status === 'pending' && a.requester_id === userId.value
}

function decide(a: ApprovalResponse, decision: 'approved' | 'rejected') {
  decisionData.value = { approvalId: a.id, decision }
  showDecisionDialog.value = true
}

async function cancel(a: ApprovalResponse) {
  if (!(await confirm(t('approvals.cancelConfirm')))) return
  try {
    await approvalApi.cancel(a.id)
    toast.success(t('approvals.cancelSuccess'))
    window.dispatchEvent(new Event('approvals:changed'))
    await load()
  } catch (e: any) {
    toast.error(e?.response?.data?.message || t('approvals.cancelFailed'))
  }
}

async function onDecided() {
  showDecisionDialog.value = false
  window.dispatchEvent(new Event('approvals:changed'))
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
  await Promise.all([loadProjects(), load()])
})
</script>
