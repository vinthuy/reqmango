<template>
  <div class="ai-result-actions flex flex-wrap gap-1.5 mt-2 pt-2 border-t border-gray-100">
    <!-- Create Issue — on text + chart + tool_result -->
    <button
      v-if="showCreateIssue"
      @click="handleCreateIssue"
      class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded bg-blue-50 text-blue-600 hover:bg-blue-100 transition-colors"
    >
      <span>📋</span>
      <span>{{ t('ai.createIssues') || 'Create Issues' }}</span>
    </button>

    <!-- Save to Dashboard — on chart -->
    <div v-if="showSaveToDashboard" class="relative">
      <button
        @click="showDashboardPicker = !showDashboardPicker"
        class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded bg-emerald-50 text-emerald-600 hover:bg-emerald-100 transition-colors"
      >
        <span>📊</span>
        <span>{{ t('ai.saveToDashboard') || 'Save to Dashboard' }}</span>
      </button>
      <!-- Dashboard picker dropdown -->
      <div
        v-if="showDashboardPicker"
        class="absolute bottom-full left-0 mb-1 w-48 bg-white border border-gray-200 rounded-lg shadow-lg z-20"
      >
        <div class="p-1.5 border-b border-gray-100 text-xs text-gray-500">
          {{ t('ai.selectDashboard') || 'Select Dashboard' }}
        </div>
        <div v-if="dashboardLoading" class="p-2 text-xs text-gray-400 text-center">...</div>
        <div v-else-if="dashboards.length === 0" class="p-2 text-xs text-gray-400 text-center">
          {{ t('ai.noDashboards') || 'No dashboards available' }}
        </div>
        <button
          v-for="d in dashboards"
          :key="d.id"
          @click="saveChartToDashboard(d.id)"
          class="block w-full text-left px-2 py-1.5 text-xs text-gray-700 hover:bg-gray-50 transition-colors"
        >{{ d.name }}</button>
      </div>
    </div>

    <!-- Save as Page — on text + agent results -->
    <button
      v-if="showSaveAsPage"
      @click="$emit('save-as-page', content)"
      class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded bg-violet-50 text-violet-600 hover:bg-violet-100 transition-colors"
    >
      <span>📝</span>
      <span>{{ t('ai.saveAsPage') || 'Save as Page' }}</span>
    </button>

    <!-- Batch Create Subtasks — on search results -->
    <button
      v-if="showBatchCreate"
      @click="handleBatchCreate"
      class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded bg-amber-50 text-amber-600 hover:bg-amber-100 transition-colors"
    >
      <span>📦</span>
      <span>{{ t('ai.batchCreateSubtasks') || 'Batch Create Subtasks' }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { listDashboards, addWidget } from '@/api/dashboard'
import type { Dashboard } from '@/types/dashboard'

const { t } = useI18n()
const toast = useToast()

const props = withDefaults(defineProps<{
  messageType: 'text' | 'chart' | 'tool_result'
  projectId: number
  workspaceId: number
  chartConfig?: any
  content?: string
  toolResult?: { toolName?: string; rows?: any[] }
}>(), {
  messageType: 'text',
})

const emit = defineEmits<{
  'create-issue': [suggestion: Record<string, any>]
  'save-as-page': [content: string | undefined]
  'saved-to-dashboard': [dashboardId: number]
}>()

// Show logic
const showCreateIssue = computed(() =>
  ['text', 'chart', 'tool_result'].includes(props.messageType)
)
const showSaveToDashboard = computed(() =>
  props.messageType === 'chart' && !!props.chartConfig
)
const showSaveAsPage = computed(() =>
  ['text', 'tool_result'].includes(props.messageType) && !!props.content
)
const showBatchCreate = computed(() =>
  props.messageType === 'tool_result' &&
  props.toolResult?.toolName === 'search_issues' &&
  (props.toolResult?.rows?.length || 0) > 0
)

// Dashboard picker
const showDashboardPicker = ref(false)
const dashboards = ref<Dashboard[]>([])
const dashboardLoading = ref(false)

async function loadDashboards() {
  dashboardLoading.value = true
  try {
    dashboards.value = await listDashboards(props.projectId)
  } catch {
    // dashboards stay empty
  } finally {
    dashboardLoading.value = false
  }
}

async function saveChartToDashboard(dashboardId: number) {
  showDashboardPicker.value = false
  try {
    const chartCfg = props.chartConfig
    const widgetType = (chartCfg?.chart_type || 'bar_chart') + '_chart'
    await addWidget(props.projectId, dashboardId, {
      widget_type: widgetType as any,
      title: chartCfg?.title || 'AI Generated Chart',
      config: {
        chart_type: chartCfg?.chart_type || 'bar',
        labels: chartCfg?.labels || [],
        datasets: chartCfg?.datasets || [],
        options: chartCfg?.options || {},
      },
      sort_order: 999,
    })
    toast.success(t('ai.chartSavedToDashboard') || 'Chart added to dashboard!')
    emit('saved-to-dashboard', dashboardId)
  } catch (e: any) {
    toast.error(e?.response?.data?.message || e?.message || 'Failed to save chart')
  }
}

function handleCreateIssue() {
  const suggestion: Record<string, any> = {}
  if (props.chartConfig) {
    suggestion.name = props.chartConfig.title || ''
    suggestion.description = JSON.stringify(props.chartConfig, null, 2)
  } else if (props.content) {
    // Extract first line as title suggestion
    const lines = props.content.trim().split('\n')
    suggestion.name = lines[0].replace(/^#+\s*/, '').slice(0, 120)
    suggestion.description = props.content
  }
  emit('create-issue', suggestion)
}

function handleBatchCreate() {
  const rows = props.toolResult?.rows || []
  const items = rows.map((row: any) => ({
    name: row.name || row.title || 'Untitled',
    description: row.description || '',
    priority: row.priority || 'medium',
  }))
  emit('create-issue', { batch: items })
}

// Close dashboard picker on outside click
function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.ai-result-actions')) {
    showDashboardPicker.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
  if (showSaveToDashboard.value) loadDashboards()
})
onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
})
</script>
