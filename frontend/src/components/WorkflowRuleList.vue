<template>
  <div class="workflow-rule-list">
    <!-- 头部工具栏 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <h3 class="text-sm font-medium text-gray-700">{{ t('workflowRule.title') }}</h3>
          <!-- 启用/禁用筛选 -->
          <select
            v-model="filters.enabled"
            class="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">{{ t('workflowRule.all') }}</option>
            <option value="true">{{ t('workflowRule.enabled') }}</option>
            <option value="false">{{ t('workflowRule.disabled') }}</option>
          </select>
        </div>

        <div class="flex items-center space-x-3">
          <!-- 从模板创建 -->
          <button
            @click="$emit('open-templates')"
            class="px-3 py-1.5 text-sm text-indigo-600 hover:text-indigo-800 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
            </svg>
            <span>{{ t('workflowRule.fromTemplate') }}</span>
          </button>

          <!-- 新建规则 -->
          <button
            @click="$emit('create')"
            class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 flex items-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('workflowRule.newRule') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 列表内容 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="mt-2 text-gray-500">{{ t('workflowRule.loading') }}</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="rules.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <p class="mt-2 text-gray-500">{{ t('workflowRule.empty') }}</p>
        <button @click="$emit('create')" class="mt-3 text-indigo-600 hover:text-indigo-800 text-sm">
          {{ t('workflowRule.createFirst') }}
        </button>
      </div>

      <!-- 规则列表 -->
      <div v-else class="space-y-4">
        <WorkflowRuleCard
          v-for="rule in filteredRules"
          :key="rule.id"
          :rule="rule"
          @toggle="toggleRule(rule)"
          @edit="$emit('edit', rule)"
          @delete="$emit('delete', rule)"
          @view-logs="$emit('view-logs', rule)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import WorkflowRuleCard from './WorkflowRuleCard.vue'
import workflowApi from '@/api/workflow'
import type { AutomationRule } from '@/types/workflow'

const { t } = useI18n()

// Props
const props = defineProps<{
  projectId: number
  workspaceId: number
}>()

// Emits
defineEmits<{
  (e: 'create'): void
  (e: 'edit', rule: AutomationRule): void
  (e: 'delete', rule: AutomationRule): void
  (e: 'open-templates'): void
  (e: 'view-logs', rule: AutomationRule): void
}>()

// State
const rules = ref<AutomationRule[]>([])
const loading = ref(false)

const filters = ref({
  enabled: ''
})

// Filtered rules
const filteredRules = computed(() => {
  if (filters.value.enabled === '') return rules.value
  const enabled = filters.value.enabled === 'true'
  return rules.value.filter(r => r.is_enabled === enabled)
})

// Load rules
onMounted(() => {
  loadRules()
})

async function loadRules() {
  loading.value = true
  try {
    rules.value = await workflowApi.listAutomationRules(props.projectId)
  } catch (error) {
    console.error('Failed to load automation rules:', error)
  } finally {
    loading.value = false
  }
}

// Toggle rule enabled
async function toggleRule(rule: AutomationRule) {
  try {
    const updated = await workflowApi.toggleAutomationRule(props.projectId, rule.id, !rule.is_enabled)
    const index = rules.value.findIndex(r => r.id === rule.id)
    if (index !== -1) {
      rules.value[index] = updated
    }
  } catch (error) {
    console.error('Failed to toggle rule:', error)
  }
}
</script>

<style scoped>
.workflow-rule-list {
  @apply bg-white rounded-lg;
}
</style>