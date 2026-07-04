<template>
  <div class="agent-selector">
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
      class="w-full px-2 py-1.5 border border-gray-200 rounded-md text-sm focus:outline-none focus:ring-1 focus:ring-blue-400 focus:border-blue-400"
    >
      <option value="">{{ placeholder }}</option>
      <optgroup v-if="members.length > 0" :label="t('agent.members') || 'Members'">
        <option v-for="m in members" :key="'m' + m.id" :value="m.id">
          {{ m.display_name || m.username || m.email || 'User #' + m.id }}
        </option>
      </optgroup>
      <optgroup v-if="activeAgents.length > 0" :label="t('agent.aiAgents') || 'AI Agents'">
        <option v-for="a in activeAgents" :key="'a' + a.id" :value="'agent:' + a.id">
          {{ a.avatar || '🤖' }} {{ a.name }}
        </option>
      </optgroup>
    </select>

    <!-- Agent capability info on hover -->
    <div v-if="selectedAgent" class="mt-2 p-2 bg-violet-50 border border-violet-200 rounded-md text-xs">
      <div class="flex items-center gap-1.5 mb-1">
        <span>{{ selectedAgent.avatar || '🤖' }}</span>
        <span class="font-medium text-violet-800">{{ selectedAgent.name }}</span>
        <span class="px-1.5 py-0.5 rounded text-xs" :class="selectedAgent.agent_type === 'builtin' ? 'bg-violet-100 text-violet-600' : 'bg-amber-100 text-amber-600'">
          {{ selectedAgent.agent_type === 'builtin' ? (t('agent.builtin') || 'Built-in') : (t('agent.custom') || 'Custom') }}
        </span>
      </div>
      <div v-if="selectedAgent.capabilities?.length" class="flex flex-wrap gap-1">
        <span
          v-for="cap in selectedAgent.capabilities"
          :key="cap"
          class="px-1.5 py-0.5 rounded-full bg-violet-100 text-violet-700 text-xs"
        >{{ formatCapability(cap) }}</span>
      </div>
      <div v-if="selectedAgent.model_override" class="mt-1 text-gray-500">
        {{ t('agent.model') || 'Model' }}: {{ selectedAgent.model_override }}
      </div>
    </div>
    <div v-else-if="activeAgents.length === 0 && !loading" class="mt-1 text-xs text-gray-400">
      {{ t('agent.noAgentsAvailable') || 'No AI agents available in this workspace' }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { agentApi } from '@/api/agent'
import type { Agent } from '@/types/agent'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  modelValue: string
  workspaceId: number
  members?: Array<{ id: number; display_name?: string; username?: string; email?: string }>
  placeholder?: string
}>(), {
  members: () => [],
  placeholder: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const agents = ref<Agent[]>([])
const loading = ref(false)

const activeAgents = computed(() => agents.value.filter(a => a.status === 'active'))

const selectedAgent = computed(() => {
  if (!props.modelValue || !props.modelValue.startsWith('agent:')) return null
  const agentId = parseInt(props.modelValue.replace('agent:', ''))
  return agents.value.find(a => a.id === agentId) || null
})

function formatCapability(cap: string): string {
  const labels: Record<string, string> = {
    create_issue: t('agent.capCreate') || 'Create',
    update_issue: t('agent.capUpdate') || 'Update',
    search: t('agent.capSearch') || 'Search',
    comment: t('agent.capComment') || 'Comment',
    label: t('agent.capLabel') || 'Label',
    assign: t('agent.capAssign') || 'Assign',
    triage: t('agent.capTriage') || 'Triage',
    summarize: t('agent.capSummarize') || 'Summarize',
    plan: t('agent.capPlan') || 'Plan',
  }
  return labels[cap] || cap
}

async function fetchAgents() {
  if (!props.workspaceId) return
  loading.value = true
  try {
    agents.value = await agentApi.list(props.workspaceId)
  } catch (e) {
    console.error('Failed to fetch agents', e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchAgents)
watch(() => props.workspaceId, fetchAgents)
</script>
