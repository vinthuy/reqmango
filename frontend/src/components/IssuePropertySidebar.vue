<template>
  <div class="bg-white rounded-lg border border-gray-200 p-4 w-[240px] space-y-4">
    <h3 class="text-sm font-semibold text-gray-700 mb-3">{{ t('issue.properties') }}</h3>

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
        :value="issue.module_id ?? ''"
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
    <div>
      <label class="block text-xs text-gray-500 mb-1">{{ t('agent.dispatchAgent') }}</label>
      <slot name="agent" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/composables/useI18n'

interface StateOption { id: number; name: string }
interface MemberOption { id: number; display_name: string }
interface CycleOption { id: number; name: string }
interface ModuleOption { id: number; name: string }

defineProps<{
  issue: any
  states: StateOption[]
  members: MemberOption[]
  cycles: CycleOption[]
  modules: ModuleOption[]
  selectedAgentId: any
  agentDispatching: boolean
}>()

const emit = defineEmits<{
  (e: 'update:state', stateId: number): void
  (e: 'update:priority', priority: string): void
  (e: 'update:assignee', userId: number | null): void
  (e: 'update:cycle', cycleId: number | null): void
  (e: 'update:module', moduleId: number | null): void
  (e: 'update:startDate', date: string): void
  (e: 'update:targetDate', date: string): void
}>()

const { t } = useI18n()

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
