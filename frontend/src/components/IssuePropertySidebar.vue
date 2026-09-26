<template>
  <div class="bg-white rounded-lg border border-gray-200 w-[280px] shrink-0 overflow-hidden">
    <div class="px-3 py-2.5 border-b border-gray-100">
      <h3 class="text-xs font-semibold text-gray-600 uppercase tracking-wide">{{ t('issue.properties') }}</h3>
    </div>

    <div class="px-3 py-2 space-y-0">
      <div v-if="isLocked" class="mb-2 px-2.5 py-2 bg-amber-50 border border-amber-200 rounded text-xs text-amber-800">
        <div class="font-medium">{{ t('approvals.pending') }}</div>
        <div class="text-amber-700/80 mt-0.5">{{ t('approvals.pendingLockHint') }}</div>
      </div>

      <!-- Relations summary -->
      <div v-if="relationSummary && relationSummary.total > 0" class="pb-2 mb-1 border-b border-gray-100">
        <div class="flex items-center gap-2 mb-1.5">
          <span class="text-xs font-medium text-gray-600">{{ t('issue.relations') }} ({{ relationSummary.total }})</span>
        </div>
        <div class="space-y-0.5">
          <template v-for="(counts, typeName) in relationSummary.byType" :key="typeName">
            <div v-if="counts.outbound > 0 || counts.inbound > 0" class="flex items-center justify-between text-[11px]">
              <span class="text-gray-500 truncate">{{ typeName }}</span>
              <span class="text-gray-600 shrink-0">
                <span v-if="counts.outbound > 0" class="text-blue-500">→{{ counts.outbound }}</span>
                <span v-if="counts.outbound > 0 && counts.inbound > 0" class="text-gray-300">·</span>
                <span v-if="counts.inbound > 0" class="text-amber-500">←{{ counts.inbound }}</span>
              </span>
            </div>
          </template>
        </div>
      </div>

      <!-- Status & people -->
      <p class="pt-1 pb-1 text-[10px] font-semibold text-gray-400 uppercase tracking-wide">{{ t('issue.state') }} / {{ t('issue.assignee') }}</p>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.state') }}</label>
        <div class="prop-control">
          <div v-if="issue.approval_status === 'pending'" class="flex items-center gap-1.5 w-full">
            <select class="prop-input bg-gray-100 cursor-not-allowed" :value="issue.state_id" disabled>
              <option v-for="s in visibleStates" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 rounded text-[10px] font-medium whitespace-nowrap">{{ t('approvals.pending') }}</span>
          </div>
          <select v-else class="prop-input" :value="issue.state_id" @change="emitStateUpdate">
            <option v-for="s in visibleStates" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
      </div>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.priority') }}</label>
        <div class="prop-control">
          <select class="prop-input" :value="issue.priority" :disabled="isLocked" @change="emitPriorityUpdate">
            <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
            <option value="high">{{ t('issue.priorityHigh') }}</option>
            <option value="medium">{{ t('issue.priorityMedium') }}</option>
            <option value="low">{{ t('issue.priorityLow') }}</option>
            <option value="none">{{ t('issue.priorityNone') }}</option>
          </select>
        </div>
      </div>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.assignee') }}</label>
        <div class="prop-control">
          <select class="prop-input" :value="issue.assignees?.[0]?.id ?? ''" :disabled="isLocked" @change="emitAssigneeUpdate">
            <option value=""></option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.display_name }}</option>
          </select>
        </div>
      </div>

      <!-- Planning -->
      <p class="pt-2 pb-1 text-[10px] font-semibold text-gray-400 uppercase tracking-wide border-t border-gray-100 mt-1">{{ t('issue.cycle') }}</p>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.cycle') }}</label>
        <div class="prop-control">
          <select class="prop-input" :value="issue.cycle_id ?? ''" :disabled="isLocked" @change="emitCycleUpdate">
            <option value=""></option>
            <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
      </div>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.module') }}</label>
        <div class="prop-control">
          <select class="prop-input" :value="issue.module_ids?.[0] ?? ''" :disabled="isLocked" @change="emitModuleUpdate">
            <option value=""></option>
            <option v-for="m in modules" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </div>
      </div>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.release') }}</label>
        <div class="prop-control">
          <select class="prop-input" :value="issue.release_id ?? ''" :disabled="isLocked" @change="emitReleaseUpdate">
            <option value=""></option>
            <option v-for="r in releases" :key="r.id" :value="r.id">{{ r.name }} ({{ r.version }})</option>
          </select>
        </div>
      </div>

      <!-- Dates & labels -->
      <p class="pt-2 pb-1 text-[10px] font-semibold text-gray-400 uppercase tracking-wide border-t border-gray-100 mt-1">{{ t('issue.startDate') }}</p>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.startDate') }}</label>
        <div class="prop-control">
          <input type="date" class="prop-input" :value="issue.start_date?.split('T')[0] ?? ''" :disabled="isLocked" @input="emitStartDateUpdate" />
        </div>
      </div>

      <div class="prop-row">
        <label class="prop-label">{{ t('issue.targetDate') }}</label>
        <div class="prop-control">
          <input type="date" class="prop-input" :value="issue.target_date?.split('T')[0] ?? ''" :disabled="isLocked" @input="emitTargetDateUpdate" />
        </div>
      </div>

      <div class="prop-row items-start py-2" :class="{ 'pointer-events-none opacity-60': isLocked }">
        <label class="prop-label pt-1">{{ t('issue.labels') }}</label>
        <div class="prop-control">
          <LabelSelector
            :labels="labels"
            :model-value="issue.labels || issue.label_ids || []"
            @change="(ids: number[]) => $emit('update:labels', ids)"
          />
        </div>
      </div>

      <!-- Agent -->
      <div class="pt-2 mt-1 border-t border-gray-100" :class="{ 'pointer-events-none opacity-60': isLocked }">
        <div class="prop-row items-start">
          <label class="prop-label pt-1">{{ t('agent.title') }}</label>
          <div class="prop-control space-y-1.5">
            <AgentSelector v-model="localAgentId" :workspace-id="workspaceId" />
            <div v-if="agentStatus?.agent_id" class="flex items-center gap-1.5 text-xs text-gray-500">
              <span class="font-medium text-violet-700">{{ agentStatus.agent_name }}</span>
              <span v-if="agentStatus.task_status" class="px-1.5 py-0.5 rounded bg-gray-100 text-gray-600">{{ agentStatus.task_status }}</span>
            </div>
            <button
              type="button"
              @click="emitAssign"
              :disabled="!hasAgentSelected || agentAssigning"
              class="w-full px-2 py-1 text-xs font-medium rounded-md bg-violet-500 hover:bg-violet-600 text-white disabled:opacity-50 transition-colors"
            >
              {{ agentAssigning ? t('agent.assigning') : t('agent.assignAgent') }}
            </button>
            <button
              v-if="hasAgentSelected"
              type="button"
              @click="$emit('dispatch-agent', localAgentId)"
              :disabled="agentDispatching"
              class="w-full px-2 py-1 text-xs font-medium rounded-md border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-50 transition-colors"
            >
              {{ agentDispatching ? t('agent.dispatching') : t('agent.dispatchAgent') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Custom fields -->
      <template v-if="customFields.length">
        <p class="pt-2 pb-1 text-[10px] font-semibold text-gray-400 uppercase tracking-wide border-t border-gray-100 mt-1">{{ t('issue.customFields') }}</p>
        <div
          v-for="cf in customFields"
          :key="cf.field.id"
          class="prop-row"
          :class="{ 'pointer-events-none opacity-60': isLocked && cf.field.field_type !== 'boolean' }"
        >
          <label class="prop-label">
            {{ cf.field.name }}
            <span v-if="cf.field.is_required" class="text-red-500">*</span>
          </label>
          <div class="prop-control">
            <input
              v-if="cf.field.field_type === 'text' || cf.field.field_type === 'url'"
              type="text"
              class="prop-input"
              :value="cf.value ?? ''"
              :disabled="isLocked"
              @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).value)"
            />
            <input
              v-else-if="cf.field.field_type === 'number'"
              type="number"
              class="prop-input"
              :value="cf.value ?? ''"
              :disabled="isLocked"
              @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).value)"
            />
            <input
              v-else-if="cf.field.field_type === 'date'"
              type="date"
              class="prop-input"
              :value="cf.value ?? ''"
              :disabled="isLocked"
              @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).value)"
            />
            <label v-else-if="cf.field.field_type === 'boolean'" class="flex items-center gap-2" :class="isLocked ? 'cursor-not-allowed' : 'cursor-pointer'">
              <input
                type="checkbox"
                class="w-4 h-4 rounded border-gray-300 text-indigo-600"
                :checked="cf.value === 'true'"
                :disabled="isLocked"
                @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).checked ? 'true' : 'false')"
              />
              <span class="text-sm text-gray-700">{{ cf.value === 'true' ? t('customField.yes') : t('customField.no') }}</span>
            </label>
            <select
              v-else-if="cf.field.field_type === 'dropdown'"
              class="prop-input"
              :value="cf.value ?? ''"
              :disabled="isLocked"
              @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLSelectElement).value)"
            >
              <option value=""></option>
              <option v-for="opt in cf.field.options" :key="opt.id" :value="opt.value">{{ opt.value }}</option>
            </select>
            <select
              v-else-if="cf.field.field_type === 'member'"
              class="prop-input"
              :value="cf.value ? JSON.parse(cf.value)[0] ?? '' : ''"
              :disabled="isLocked"
              @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, JSON.stringify([Number((e.target as HTMLSelectElement).value)]))"
            >
              <option value=""></option>
              <option v-for="m in members" :key="m.id" :value="m.id">{{ m.display_name }}</option>
            </select>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import AgentSelector from '@/components/AgentSelector.vue'
import LabelSelector from '@/components/LabelSelector.vue'
import type { AgentStatus } from '@/api/issue-agent'

interface StateOption { id: number; name: string }
interface MemberOption { id: number; display_name: string }
interface CycleOption { id: number; name: string }
interface ModuleOption { id: number; name: string }
interface CustomFieldEntry {
  field: {
    id: number
    name: string
    field_type: string
    is_required?: boolean
    options?: Array<{ id: number; value: string; color?: string }>
  }
  value: string | null
}

const props = defineProps<{
  issue: any
  states: StateOption[]
  members: MemberOption[]
  cycles: CycleOption[]
  modules: ModuleOption[]
  releases: Array<{ id: number; name: string; version: string }>
  customFields: CustomFieldEntry[]
  workspaceId: number
  agentDispatching?: boolean
  agentAssigning?: boolean
  agentStatus?: AgentStatus | null
  labels?: Array<{ id: number; name: string; color: string }>
  relationSummary?: {
    total: number
    outbound: number
    inbound: number
    byType: Record<string, { outbound: number; inbound: number }>
  } | null
}>()

const isLocked = computed(() => props.issue?.approval_status === 'pending')
const visibleStates = computed(() =>
  (props.states || []).filter((s: any) => {
    if (s?.is_active === false) return false
    if (/^E2E\s+Test/i.test(String(s?.name || ''))) return false
    return true
  })
)

const emit = defineEmits<{
  (e: 'update:state', stateId: number): void
  (e: 'update:priority', priority: string): void
  (e: 'update:assignee', userId: number | null): void
  (e: 'update:cycle', cycleId: number | null): void
  (e: 'update:module', moduleId: number | null): void
  (e: 'update:release', releaseId: number | null): void
  (e: 'update:startDate', date: string): void
  (e: 'update:targetDate', date: string): void
  (e: 'update:customField', fieldId: number, value: string): void
  (e: 'dispatch-agent', agentId: string): void
  (e: 'assign-agent', agentId: string): void
  (e: 'unassign-agent'): void
  (e: 'update:labels', labelIds: number[]): void
}>()

const localAgentId = ref('')
const syncingAgent = ref(false)
const hasAgentSelected = computed(() => !!localAgentId.value && localAgentId.value.startsWith('agent:'))

watch(() => props.agentStatus, (status) => {
  syncingAgent.value = true
  localAgentId.value = status?.agent_id ? `agent:${status.agent_id}` : ''
  nextTick(() => { syncingAgent.value = false })
}, { immediate: true })

// Only clear selection unassigns; assign requires the Assign button (no auto-assign on select).
watch(localAgentId, (val, oldVal) => {
  if (syncingAgent.value) return
  if (oldVal && oldVal.startsWith('agent:') && (!val || !val.startsWith('agent:'))) {
    emit('unassign-agent')
  }
})

function emitAssign() {
  if (!hasAgentSelected.value) return
  emit('assign-agent', localAgentId.value)
}

const { t } = useI18n()

function emitCustomFieldUpdate(fieldId: number, value: string) {
  emit('update:customField', fieldId, value)
}
function emitStateUpdate(event: Event) {
  emit('update:state', Number((event.target as HTMLSelectElement).value))
}
function emitPriorityUpdate(event: Event) {
  emit('update:priority', (event.target as HTMLSelectElement).value)
}
function emitAssigneeUpdate(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('update:assignee', value ? Number(value) : null)
}
function emitCycleUpdate(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('update:cycle', value ? Number(value) : null)
}
function emitModuleUpdate(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('update:module', value ? Number(value) : null)
}
function emitReleaseUpdate(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('update:release', value ? Number(value) : null)
}
function emitStartDateUpdate(event: Event) {
  emit('update:startDate', (event.target as HTMLInputElement).value)
}
function emitTargetDateUpdate(event: Event) {
  emit('update:targetDate', (event.target as HTMLInputElement).value)
}
</script>

<style scoped>
.prop-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2rem;
  padding: 0.2rem 0;
}
.prop-label {
  width: 4.5rem;
  flex-shrink: 0;
  font-size: 0.75rem;
  line-height: 1rem;
  color: #6b7280;
}
.prop-control {
  flex: 1;
  min-width: 0;
}
.prop-input {
  width: 100%;
  padding: 0.25rem 0.375rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.25rem;
  font-size: 0.8125rem;
  line-height: 1.25rem;
  background: #fff;
}
.prop-input:disabled {
  background: #f3f4f6;
  cursor: not-allowed;
}
.prop-input:focus {
  outline: none;
  border-color: #a5b4fc;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}
</style>
