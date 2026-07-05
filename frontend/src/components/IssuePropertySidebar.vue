<template>
  <div class="bg-white rounded-lg border border-gray-200 p-4 w-[240px] space-y-4">
    <h3 class="text-sm font-semibold text-gray-700 mb-3">{{ t('issue.properties') }}</h3>

    <!-- Relations summary -->
    <div v-if="relationSummary && relationSummary.total > 0" class="pb-3 border-b border-gray-100">
      <div class="flex items-center gap-2 mb-2">
        <span class="text-xs text-gray-500">🔗</span>
        <span class="text-xs font-medium text-gray-600">{{ t('issue.relations') }} ({{ relationSummary.total }})</span>
      </div>
      <div class="space-y-0.5">
        <template v-for="(counts, typeName) in relationSummary.byType" :key="typeName">
          <div v-if="counts.outbound > 0 || counts.inbound > 0" class="flex items-center justify-between text-[11px]">
            <span class="text-gray-500">{{ typeName }}</span>
            <span class="text-gray-600">
              <span v-if="counts.outbound > 0" class="text-blue-500">→{{ counts.outbound }}</span>
              <span v-if="counts.outbound > 0 && counts.inbound > 0" class="text-gray-300">·</span>
              <span v-if="counts.inbound > 0" class="text-amber-500">←{{ counts.inbound }}</span>
            </span>
          </div>
        </template>
      </div>
    </div>

    <!-- State -->
    <div>
      <label class="block text-xs text-gray-500 mb-1">{{ t('issue.state') }}</label>
      <select
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="issue.state_id"
        @change="emitStateUpdate"
      >
        <option v-for="s in states" :key="s.id" :value="s.id">{{ s.name }}</option>
      </select>
    </div>

    <!-- Priority -->
    <div>
      <label class="block text-xs text-gray-500 mb-1">{{ t('issue.priority') }}</label>
      <select
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="issue.priority"
        @change="emitPriorityUpdate"
      >
        <option value="urgent">{{ t('issue.priorityUrgent') }}</option>
        <option value="high">{{ t('issue.priorityHigh') }}</option>
        <option value="medium">{{ t('issue.priorityMedium') }}</option>
        <option value="low">{{ t('issue.priorityLow') }}</option>
        <option value="none">{{ t('issue.priorityNone') }}</option>
      </select>
    </div>

    <!-- Assignee -->
    <div>
      <label class="block text-xs text-gray-500 mb-1">{{ t('issue.assignee') }}</label>
      <select
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="issue.assignees?.[0]?.id ?? ''"
        @change="emitAssigneeUpdate"
      >
        <option value=""></option>
        <option v-for="m in members" :key="m.id" :value="m.id">{{ m.display_name }}</option>
      </select>
    </div>

    <!-- Cycle -->
    <div>
      <label class="block text-xs text-gray-500 mb-1">{{ t('issue.cycle') }}</label>
      <select
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="issue.cycle_id ?? ''"
        @change="emitCycleUpdate"
      >
        <option value=""></option>
        <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
    </div>

    <!-- Module -->
    <div>
      <label class="block text-xs text-gray-500 mb-1">{{ t('issue.module') }}</label>
      <select
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="issue.module_ids?.[0] ?? ''"
        @change="emitModuleUpdate"
      >
        <option value=""></option>
        <option v-for="m in modules" :key="m.id" :value="m.id">{{ m.name }}</option>
      </select>
    </div>

    <!-- Start Date + Target Date -->
    <div class="grid grid-cols-2 gap-2">
      <div>
        <label class="block text-xs text-gray-500 mb-1">{{ t('issue.startDate') }}</label>
        <input
          type="date"
          class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
          :value="issue.start_date?.split('T')[0] ?? ''"
          @input="emitStartDateUpdate"
        />
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">{{ t('issue.targetDate') }}</label>
        <input
          type="date"
          class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
          :value="issue.target_date?.split('T')[0] ?? ''"
          @input="emitTargetDateUpdate"
        />
      </div>
    </div>

    <!-- AI Agent -->
    <div class="pt-3 border-t border-gray-100">
      <label class="block text-xs text-gray-500 mb-1">{{ t('agent.title') }}</label>
      <AgentSelector v-model="localAgentId" :workspace-id="workspaceId" />
      <button
        v-if="localAgentId"
        @click="$emit('dispatch-agent', localAgentId)"
        :disabled="agentDispatching"
        class="mt-2 w-full px-3 py-1.5 text-xs font-medium rounded-md bg-violet-500 hover:bg-violet-600 text-white disabled:opacity-50 transition-colors"
      >
        {{ agentDispatching ? t('agent.dispatching') : t('agent.dispatchAgent') }}
      </button>
    </div>

    <!-- Custom Fields -->
    <div v-for="cf in customFields" :key="cf.field.id">
      <label class="block text-xs text-gray-500 mb-1">{{ cf.field.name }}</label>
      <input
        v-if="cf.field.field_type === 'text' || cf.field.field_type === 'url'"
        type="text"
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="cf.value ?? ''"
        @input="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).value)"
      />
      <input
        v-else-if="cf.field.field_type === 'number'"
        type="number"
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="cf.value ?? ''"
        @input="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).value)"
      />
      <input
        v-else-if="cf.field.field_type === 'date'"
        type="date"
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="cf.value ?? ''"
        @input="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).value)"
      />
      <label v-else-if="cf.field.field_type === 'boolean'" class="flex items-center gap-2 cursor-pointer">
        <input
          type="checkbox"
          class="w-4 h-4 rounded border-gray-300 text-indigo-600"
          :checked="cf.value === 'true'"
          @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLInputElement).checked ? 'true' : 'false')"
        />
        <span class="text-sm text-gray-700">{{ cf.value === 'true' ? t('customField.yes') : t('customField.no') }}</span>
      </label>
      <select
        v-else-if="cf.field.field_type === 'dropdown'"
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="cf.value ?? ''"
        @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, (e.target as HTMLSelectElement).value)"
      >
        <option value=""></option>
        <option v-for="opt in cf.field.options" :key="opt.id" :value="opt.value">{{ opt.value }}</option>
      </select>
      <select
        v-else-if="cf.field.field_type === 'member'"
        class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm"
        :value="cf.value ? JSON.parse(cf.value)[0] ?? '' : ''"
        @change="(e: Event) => emitCustomFieldUpdate(cf.field.id, JSON.stringify([Number((e.target as HTMLSelectElement).value)]))"
      >
        <option value=""></option>
        <option v-for="m in members" :key="m.id" :value="m.id">{{ m.display_name }}</option>
      </select>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import AgentSelector from '@/components/AgentSelector.vue'

interface StateOption { id: number; name: string }
interface MemberOption { id: number; display_name: string }
interface CycleOption { id: number; name: string }
interface ModuleOption { id: number; name: string }
interface CustomFieldEntry {
  field: {
    id: number
    name: string
    field_type: string
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
  customFields: CustomFieldEntry[]
  workspaceId: number
  agentDispatching?: boolean
  relationSummary?: {
    total: number
    outbound: number
    inbound: number
    byType: Record<string, { outbound: number; inbound: number }>
  } | null
}>()

const emit = defineEmits<{
  (e: 'update:state', stateId: number): void
  (e: 'update:priority', priority: string): void
  (e: 'update:assignee', userId: number | null): void
  (e: 'update:cycle', cycleId: number | null): void
  (e: 'update:module', moduleId: number | null): void
  (e: 'update:startDate', date: string): void
  (e: 'update:targetDate', date: string): void
  (e: 'update:customField', fieldId: number, value: string): void
  (e: 'dispatch-agent', agentId: string): void
}>()

// Local agent model to bind AgentSelector v-model
const localAgentId = ref('')
watch(() => props.issue, () => { localAgentId.value = '' })

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

function emitStartDateUpdate(event: Event) {
  emit('update:startDate', (event.target as HTMLInputElement).value)
}

function emitTargetDateUpdate(event: Event) {
  emit('update:targetDate', (event.target as HTMLInputElement).value)
}
</script>
