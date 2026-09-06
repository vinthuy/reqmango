<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @click.self="emit('cancel')">
      <div class="bg-white rounded-xl shadow-xl w-[720px] max-h-[90vh] flex flex-col" @click.stop>
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h3 class="text-lg font-semibold text-gray-800">{{ isEdit ? t('metrics.editChart') : t('metrics.newChart') }}</h3>
          <button @click="emit('cancel')" class="text-gray-400 hover:text-gray-600 transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-y-auto px-6 py-5 space-y-5">
          <!-- Chart Name -->
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('metrics.chartName') }}</label>
            <input v-model="form.name" type="text" :placeholder="t('metrics.chartNamePlaceholder')"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors" />
          </div>

          <!-- Chart Type -->
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('metrics.chartType') }}</label>
            <div class="inline-flex bg-gray-100 rounded-lg p-0.5 flex-wrap gap-0.5">
              <button v-for="c in chartTypeOptions" :key="c.value" @click="form.chart_type = c.value"
                :class="['px-2.5 py-1 text-xs rounded-md transition-colors', form.chart_type === c.value ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
              >{{ c.label }}</button>
            </div>
          </div>

          <!-- X Axis / Y Axis -->
          <div class="flex flex-wrap items-end gap-4">
            <!-- X-Axis -->
            <div class="flex-1 min-w-[200px]">
              <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('metrics.metricDimension') }}</label>
              <select v-model="form.x_axis" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                <optgroup :label="t('metrics.categoryDimensions')">
                  <option v-for="d in xAxisCategoryOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                </optgroup>
                <optgroup :label="t('metrics.timeCreated')">
                  <option v-for="d in xAxisTimeCreatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                </optgroup>
                <optgroup :label="t('metrics.timeCompleted')">
                  <option v-for="d in xAxisTimeCompletedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                </optgroup>
                <optgroup :label="t('metrics.timeUpdated')">
                  <option v-for="d in xAxisTimeUpdatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                </optgroup>
              </select>
            </div>

            <!-- Y-Axis -->
            <div class="flex-1 min-w-[200px]">
              <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('metrics.metricIndicator') }}</label>
              <select v-model="form.y_axis" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                <option v-for="m in yAxisOptions" :key="m.value" :value="m.value">{{ m.label }}</option>
              </select>
            </div>
          </div>

          <!-- Filters (RQL) -->
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('metrics.filterConditionsRql') }}</label>
            <textarea v-model="rqlQuery" rows="2" :placeholder="t('metrics.filterRqlPlaceholder')"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono bg-gray-50 focus:bg-white focus:outline-none focus:ring-1 focus:ring-blue-400 transition-colors resize-none"></textarea>
          </div>

          <!-- Advanced Config (collapsible) -->
          <div class="border border-gray-100 rounded-lg">
            <button @click="showAdvanced = !showAdvanced"
              class="w-full flex items-center justify-between px-4 py-2.5 text-sm text-gray-600 hover:bg-gray-50 transition-colors">
              <span class="font-medium">{{ t('metrics.advancedConfig.title') }}</span>
              <svg :class="['w-4 h-4 transition-transform', showAdvanced ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>
            <div v-show="showAdvanced" class="px-4 pb-4 space-y-4 border-t border-gray-100">
              <!-- Stack Mode -->
              <div v-if="form.chart_type === 'bar' || form.chart_type === 'area'" class="pt-3">
                <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('metrics.advancedConfig.stackMode') }}</label>
                <select v-model="advancedConfig.stack_mode" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option value="none">{{ t('metrics.advancedConfig.stackNone') }}</option>
                  <option value="stack">{{ t('metrics.advancedConfig.stack') }}</option>
                  <option value="percent_stack">{{ t('metrics.advancedConfig.percentStack') }}</option>
                </select>
              </div>

              <!-- Show Labels -->
              <div class="flex items-center gap-2 pt-2">
                <input type="checkbox" v-model="advancedConfig.show_labels" id="showLabels"
                  class="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500" />
                <label for="showLabels" class="text-sm text-gray-600">{{ t('metrics.advancedConfig.showLabels') }}</label>
              </div>

              <!-- Reference Lines -->
              <div>
                <div class="flex items-center justify-between mb-2">
                  <label class="text-[11px] font-medium text-gray-400 uppercase tracking-wide">{{ t('metrics.advancedConfig.referenceLine') }}</label>
                  <button @click="addReferenceLine" class="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/></svg>
                    {{ t('metrics.addReferenceLine') }}
                  </button>
                </div>
                <div v-for="(line, idx) in advancedConfig.reference_lines" :key="idx" class="flex items-center gap-2 mb-2">
                  <select v-model="line.type" class="w-28 px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <option value="average">{{ t('metrics.advancedConfig.refTypeAverage') }}</option>
                    <option value="max">{{ t('metrics.advancedConfig.refTypeMax') }}</option>
                    <option value="min">{{ t('metrics.advancedConfig.refTypeMin') }}</option>
                    <option value="median">{{ t('metrics.advancedConfig.refTypeMedian') }}</option>
                    <option value="constant">{{ t('metrics.advancedConfig.refTypeConstant') }}</option>
                  </select>
                  <input v-if="line.type === 'constant'" v-model.number="line.value" type="number" :placeholder="t('metrics.advancedConfig.refValuePlaceholder')"
                    class="w-24 px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" />
                  <input v-model="line.label" type="text" :placeholder="t('metrics.advancedConfig.refLabelPlaceholder')"
                    class="flex-1 px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" />
                  <select v-model="line.style" class="w-20 px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <option value="solid">{{ t('metrics.advancedConfig.refStyleSolid') }}</option>
                    <option value="dashed">{{ t('metrics.advancedConfig.refStyleDashed') }}</option>
                  </select>
                  <button @click="removeReferenceLine(idx)" class="text-gray-300 hover:text-red-500 text-lg leading-none px-1 transition-colors">&times;</button>
                </div>
              </div>
            </div>
          </div>

          <!-- Live Preview -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="text-[11px] font-medium text-gray-400 uppercase tracking-wide">{{ t('metrics.advancedConfig.livePreview') }}</label>
              <button @click="refreshPreview" :disabled="previewLoading" class="px-3 py-1 bg-neutral-900 text-white text-xs rounded-md hover:bg-neutral-800 disabled:opacity-50 transition-colors">
                {{ previewLoading ? t('metrics.loading') : t('metrics.advancedConfig.refreshPreview') }}
              </button>
            </div>
            <div class="bg-gray-50 border border-gray-100 rounded-xl p-4">
              <div v-if="previewLoading" class="flex items-center justify-center h-48 text-xs text-gray-400">
                {{ t('metrics.loading') }}
              </div>
              <div v-else-if="previewError" class="flex items-center justify-center h-48 text-xs text-red-400">
                {{ previewError }}
              </div>
              <div v-else-if="previewData" :class="['mx-auto', isPieType ? 'max-w-md' : isRadarType ? 'max-w-lg' : 'max-w-4xl']" style="height: 280px">
                <canvas ref="previewCanvas"></canvas>
              </div>
              <div v-else class="flex flex-col items-center justify-center h-48 text-xs text-gray-400">
                <svg class="w-8 h-8 mb-2 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
                <span>{{ t('metrics.advancedConfig.clickRefreshPreview') }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end gap-2 px-6 py-4 border-t border-gray-100">
          <button @click="emit('cancel')" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg text-sm transition-colors">{{ t('common.cancel') }}</button>
          <button @click="handleSave" :disabled="saving || !form.name.trim()" class="px-5 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-50 transition-colors">
            {{ saving ? t('common.saving') : (isEdit ? t('common.update') : t('common.create')) }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, nextTick, onUnmounted } from 'vue'
import type { MetricTemplate, MetricChart, CreateChartPayload, MetricChartConfig } from '@/types/metrics'
import { reportApi, type ReportResponse } from '@/api/report'
import { useReportChart } from '@/composables/useReportChart'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  projectId: number
  visible: boolean
  template?: MetricTemplate
  chart?: MetricChart
}>()

const emit = defineEmits<{
  (e: 'save', data: CreateChartPayload): void
  (e: 'cancel'): void
}>()

// ── State ──
const isEdit = computed(() => !!props.chart)
const showAdvanced = ref(false)
const saving = ref(false)

// Chart type mapping: metrics uses lowercase, ReportBuilder uses capitalized names
const chartTypeMap: Record<string, string> = {
  bar: 'Bar', line: 'Line', pie: 'Pie', doughnut: 'Doughnut',
  area: 'Area', radar: 'Radar', scatter: 'Scatter', bubble: 'Bubble',
  mixed: 'Mixed', table: 'Table',
}

const chartTypeOptions = [
  { value: 'bar', label: t('metrics.chartTypes.bar') },
  { value: 'line', label: t('metrics.chartTypes.line') },
  { value: 'pie', label: t('metrics.chartTypes.pie') },
  { value: 'doughnut', label: t('metrics.chartTypes.doughnut') },
  { value: 'area', label: t('metrics.chartTypes.area') },
  { value: 'radar', label: t('metrics.chartTypes.radar') },
  { value: 'scatter', label: t('metrics.chartTypes.scatter') },
  { value: 'bubble', label: t('metrics.chartTypes.bubble') },
  { value: 'mixed', label: t('metrics.chartTypes.mixed') },
  { value: 'table', label: t('metrics.chartTypes.table') },
]

// ── X Axis Options ──
const xAxisCategoryOptions = [
  { value: 'state', label: t('metrics.dimensions.state') },
  { value: 'priority', label: t('metrics.dimensions.priority') },
  { value: 'assignee', label: t('metrics.dimensions.assignee') },
  { value: 'type', label: t('metrics.dimensions.type') },
  { value: 'label', label: t('metrics.dimensions.label') },
  { value: 'cycle', label: t('metrics.dimensions.cycle') },
  { value: 'module', label: t('metrics.dimensions.module') },
  { value: 'state_group', label: t('metrics.dimensions.state_group') },
  { value: 'created_by', label: t('metrics.dimensions.created_by') },
]
const xAxisTimeCreatedOptions = [
  { value: 'created_day', label: t('metrics.dimensions.byDay') },
  { value: 'created_week', label: t('metrics.dimensions.byWeek') },
  { value: 'created_month', label: t('metrics.dimensions.byMonth') },
]
const xAxisTimeCompletedOptions = [
  { value: 'completed_day', label: t('metrics.dimensions.byDay') },
  { value: 'completed_week', label: t('metrics.dimensions.byWeek') },
  { value: 'completed_month', label: t('metrics.dimensions.byMonth') },
]
const xAxisTimeUpdatedOptions = [
  { value: 'updated_day', label: t('metrics.dimensions.byDay') },
  { value: 'updated_week', label: t('metrics.dimensions.byWeek') },
  { value: 'updated_month', label: t('metrics.dimensions.byMonth') },
]

// ── Y Axis Options ──
const yAxisOptions = [
  { value: 'count', label: t('metrics.yAxisOptions.count') },
  { value: 'avg_processing_time', label: t('metrics.yAxisOptions.avg_processing_time') },
  { value: 'current_retention', label: t('metrics.yAxisOptions.current_retention') },
  { value: 'created_vs_resolved', label: t('metrics.yAxisOptions.created_vs_resolved') },
]

// ── Form Data ──
const form = reactive({
  name: '',
  chart_type: 'bar',
  x_axis: 'state',
  y_axis: 'count',
})

const rqlQuery = ref('')

const advancedConfig = reactive<MetricChartConfig>({
  stack_mode: 'none',
  show_labels: false,
  reference_lines: [],
})

// ── Reference Lines ──
function addReferenceLine() {
  advancedConfig.reference_lines!.push({
    type: 'average',
    value: undefined,
    label: '',
    color: '#EF4444',
    style: 'dashed',
  })
}

function removeReferenceLine(idx: number) {
  advancedConfig.reference_lines!.splice(idx, 1)
}

// ── Preview ──
const previewCanvas = ref<HTMLCanvasElement | null>(null)
const previewData = ref<ReportResponse | null>(null)
const previewLoading = ref(false)
const previewError = ref<string | null>(null)

const { render: renderChart, destroy: destroyChart } = useReportChart(previewCanvas)

const isPieType = computed(() => ['pie', 'doughnut'].includes(form.chart_type))
const isRadarType = computed(() => form.chart_type === 'radar')

function chartTypeForRenderer(type: string): string {
  return chartTypeMap[type] || 'Bar'
}

async function refreshPreview() {
  previewLoading.value = true
  previewError.value = null
  try {
    const res = await reportApi.generateV2(props.projectId, {
      x_axis: form.x_axis,
      y_axis: form.y_axis,
      rql: rqlQuery.value || undefined,
    })
    previewData.value = res
    await nextTick()
    if (form.chart_type !== 'table') {
      await new Promise(r => setTimeout(r, 50))
      renderChart(res, chartTypeForRenderer(form.chart_type))
    }
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || e?.message || t('metrics.advancedConfig.loadingPreview')
    previewError.value = String(msg)
    previewData.value = null
  } finally {
    previewLoading.value = false
  }
}

// Re-render when chart type changes
watch(() => form.chart_type, async (newVal) => {
  if (newVal === 'table') {
    destroyChart()
  } else if (previewData.value) {
    await nextTick()
    await new Promise(r => setTimeout(r, 30))
    renderChart(previewData.value, chartTypeForRenderer(newVal))
  }
})

// ── Initialize from props ──
watch(() => props.visible, (v) => {
  if (v) {
    if (props.chart) {
      // Edit mode: fill form from existing chart
      form.name = props.chart.name
      form.chart_type = props.chart.chart_type
      form.x_axis = props.chart.x_axis
      form.y_axis = props.chart.y_axis
      rqlQuery.value = props.chart.filters || ''
      // Parse config JSON
      try {
        const cfg = JSON.parse(props.chart.config || '{}')
        advancedConfig.stack_mode = cfg.stack_mode || 'none'
        advancedConfig.show_labels = cfg.show_labels || false
        advancedConfig.reference_lines = cfg.reference_lines || []
      } catch {
        // ignore parse errors
      }
    } else if (props.template) {
      // Create from template
      form.name = props.template.name
      form.chart_type = props.template.chart_type
      form.x_axis = props.template.default_x_axis
      form.y_axis = props.template.default_y_axis
      rqlQuery.value = ''
      if (props.template.default_config) {
        advancedConfig.stack_mode = props.template.default_config.stack_mode || 'none'
        advancedConfig.show_labels = props.template.default_config.show_labels || false
        advancedConfig.reference_lines = props.template.default_config.reference_lines || []
      }
    } else {
      // Blank create
      form.name = ''
      form.chart_type = 'bar'
      form.x_axis = 'state'
      form.y_axis = 'count'
      rqlQuery.value = ''
      advancedConfig.stack_mode = 'none'
      advancedConfig.show_labels = false
      advancedConfig.reference_lines = []
    }
    showAdvanced.value = false
    previewData.value = null
    previewError.value = null
  }
})

// ── Save ──
async function handleSave() {
  if (!form.name.trim() || saving.value) return
  saving.value = true
  try {
    const payload: CreateChartPayload = {
      name: form.name.trim(),
      chart_type: form.chart_type,
      x_axis: form.x_axis,
      y_axis: form.y_axis,
      filters: rqlQuery.value ? { rql: rqlQuery.value } : undefined,
      config: {
        stack_mode: advancedConfig.stack_mode,
        show_labels: advancedConfig.show_labels,
        reference_lines: advancedConfig.reference_lines,
      },
    }
    if (props.template) {
      payload.template_id = props.template.id
    }
    emit('save', payload)
  } finally {
    saving.value = false
  }
}

// Cleanup on unmount
onUnmounted(() => { destroyChart() })
</script>
