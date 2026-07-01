<template>
  <div class="automation-template-list">
    <!-- 头部 -->
    <div class="bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-gray-700">{{ t('automationTemplate.title') }}</h3>
        <select
          v-model="selectedCategory"
          class="px-3 py-1.5 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          <option value="">{{ t('automationTemplate.allCategories') }}</option>
          <option value="issue">{{ t('automationTemplate.categoryIssues') }}</option>
          <option value="notification">{{ t('automationTemplate.categoryNotifications') }}</option>
          <option value="workflow">{{ t('automationTemplate.categoryWorkflow') }}</option>
        </select>
      </div>
    </div>

    <!-- 模板列表 -->
    <div class="p-4">
      <!-- 加载状态 -->
      <div v-if="loading" class="text-center py-12">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <!-- 模板网格 -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="template in filteredTemplates"
          :key="template.id"
          class="bg-white border border-gray-200 rounded-lg p-4 hover:border-indigo-300 transition-colors"
        >
          <div class="flex items-start justify-between mb-2">
            <div>
              <h4 class="text-sm font-medium text-gray-900">{{ template.name }}</h4>
              <span class="text-xs text-gray-500">{{ template.category }}</span>
            </div>
            <span
              v-if="template.is_system"
              class="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded"
            >
              {{ t('automationTemplate.system') }}
            </span>
          </div>

          <p v-if="template.description" class="text-sm text-gray-500 mb-3 line-clamp-2">
            {{ template.description }}
          </p>

          <!-- 触发器和动作预览 -->
          <div class="mb-3">
            <div class="flex items-center space-x-1 mb-1">
              <svg class="w-3 h-3 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <span class="text-xs text-gray-600">{{ getTriggerName(template.trigger.type) }}</span>
            </div>
            <div v-if="template.conditions && template.conditions.length > 0" class="text-xs text-gray-500">
              {{ t('workflowRule.conditionsCount', { count: template.conditions.length }) }}
            </div>
            <div class="text-xs text-gray-500">
              {{ t('workflowRule.actionsCount', { count: template.actions?.length || 0 }) }}
            </div>
          </div>

          <!-- 使用统计 -->
          <div class="text-xs text-gray-400 mb-3">
            {{ t('workflowRule.executedCount', { count: template.usage_count }) }}
          </div>

          <!-- 操作 -->
          <button
            @click="$emit('apply', template)"
            class="w-full px-3 py-1.5 bg-indigo-50 text-indigo-600 text-sm rounded-md hover:bg-indigo-100 flex items-center justify-center space-x-1"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>{{ t('automationTemplate.useTemplate') }}</span>
          </button>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && filteredTemplates.length === 0" class="text-center py-12">
        <svg class="h-12 w-12 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
        </svg>
        <p class="mt-2 text-gray-500">{{ t('automationTemplate.empty') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import workflowApi from '@/api/workflow'
import { useI18n } from '@/composables/useI18n'
import type { AutomationTemplate, TriggerTypeEnum } from '@/types/workflow'
import { getTriggerDisplayName } from '@/types/workflow'

const { t } = useI18n()

// Props
defineProps<{
  projectId: string
  workspaceId: string
}>()

// Emits
defineEmits<{
  (e: 'apply', template: AutomationTemplate): void
}>()

// State
const templates = ref<AutomationTemplate[]>([])
const loading = ref(false)
const selectedCategory = ref('')

// Filtered templates
const filteredTemplates = computed(() => {
  if (!selectedCategory.value) return templates.value
  return templates.value.filter(t => t.category === selectedCategory.value)
})

// Load templates
onMounted(() => {
  loadTemplates()
})

async function loadTemplates() {
  loading.value = true
  try {
    templates.value = await workflowApi.listAutomationTemplates()
  } catch (error) {
    console.error('Failed to load templates:', error)
  } finally {
    loading.value = false
  }
}

// Get trigger name
function getTriggerName(type: string): string {
  return getTriggerDisplayName(type as TriggerTypeEnum) || type
}
</script>

<style scoped>
.automation-template-list {
  @apply bg-white rounded-lg;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>