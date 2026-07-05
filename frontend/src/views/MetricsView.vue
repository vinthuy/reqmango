<template>
  <div class="h-full flex flex-col bg-gray-50/50">
    <!-- Header Bar -->
    <div class="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200 shrink-0">
      <div class="flex items-center gap-3">
        <h2 class="text-base font-semibold text-gray-800">度量</h2>
        <span class="text-xs text-gray-400">{{ charts.length }} 个图表</span>
      </div>
      <button @click="openSidePanel('new')"
        class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        添加图表
      </button>
    </div>

    <!-- Main Content: Chart Grid -->
    <div class="flex-1 overflow-y-auto p-6">
      <!-- Loading -->
      <div v-if="loading && charts.length === 0" class="flex items-center justify-center py-20">
        <svg class="animate-spin h-6 w-6 text-gray-300" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>

      <!-- Empty State -->
      <div v-else-if="charts.length === 0" class="flex flex-col items-center justify-center py-20">
        <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
          <svg class="w-8 h-8 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
          </svg>
        </div>
        <p class="text-sm text-gray-500 mb-1">暂无度量图表</p>
        <p class="text-xs text-gray-400 mb-4">点击「添加图表」开始创建，或从预置模板快速开始</p>
        <button @click="openSidePanel('new')"
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          添加图表
        </button>
      </div>

      <!-- Chart Grid -->
      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <MetricsChartCard
          v-for="chart in charts"
          :key="chart.id"
          :chart="chart"
          :project-id="projectId"
          @edit="openSidePanel('edit', chart)"
          @delete="handleDeleteChart(chart.id)"
        />
      </div>
    </div>

    <!-- Side Panel: Add/Edit Chart -->
    <Teleport to="body">
      <Transition name="slide">
        <div v-if="sidePanel.visible" class="fixed inset-0 z-50 flex justify-end">
          <!-- Backdrop -->
          <div class="absolute inset-0 bg-black/30" @click="closeSidePanel"></div>
          <!-- Panel -->
          <div class="relative w-[520px] bg-white shadow-2xl flex flex-col animate-slide-in">
            <!-- Panel Header -->
            <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 shrink-0">
              <h3 class="text-sm font-semibold text-gray-800">
                {{ sidePanel.mode === 'edit' ? '编辑图表' : '添加图表' }}
              </h3>
              <button @click="closeSidePanel" class="p-1 text-gray-400 hover:text-gray-600 rounded transition-colors">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>

            <!-- Panel Body -->
            <div class="flex-1 overflow-y-auto">
              <!-- Template Selection (only in new mode) -->
              <div v-if="sidePanel.mode === 'new' && !sidePanel.selectedTemplate" class="p-5">
                <p class="text-xs font-medium text-gray-400 uppercase tracking-wide mb-3">选择模板</p>
                <!-- Category Tabs -->
                <div class="flex gap-1 mb-4 bg-gray-100 rounded-lg p-0.5">
                  <button v-for="cat in categories" :key="cat.id"
                    @click="sidePanel.activeCategory = cat.id"
                    :class="[
                      'flex-1 px-3 py-1.5 text-xs font-medium rounded-md transition-colors',
                      sidePanel.activeCategory === cat.id
                        ? 'bg-white shadow-sm text-gray-800'
                        : 'text-gray-500 hover:text-gray-700'
                    ]"
                  >{{ cat.name }}</button>
                </div>
                <!-- Template Grid -->
                <div class="grid grid-cols-2 gap-2">
                  <button v-for="tpl in currentCategoryTemplates" :key="tpl.id"
                    @click="selectTemplate(tpl)"
                    class="flex items-start gap-2.5 p-3 bg-gray-50 hover:bg-indigo-50 hover:border-indigo-200 border border-gray-100 rounded-lg text-left transition-colors group"
                  >
                    <span class="text-lg mt-0.5">{{ getTemplateIcon(tpl.icon) }}</span>
                    <div class="min-w-0">
                      <p class="text-sm font-medium text-gray-700 group-hover:text-indigo-700 truncate">{{ tpl.name }}</p>
                      <p class="text-[11px] text-gray-400 mt-0.5 line-clamp-2">{{ tpl.description }}</p>
                    </div>
                  </button>
                </div>
                <!-- Or create custom -->
                <div class="mt-4 pt-4 border-t border-gray-100">
                  <button @click="sidePanel.useCustom = true"
                    class="w-full flex items-center justify-center gap-2 py-2.5 border border-dashed border-gray-300 rounded-lg text-sm text-gray-500 hover:text-indigo-600 hover:border-indigo-300 transition-colors">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                    </svg>
                    自定义创建
                  </button>
                </div>
              </div>

              <!-- Chart Configuration Form -->
              <div v-if="sidePanel.mode === 'edit' || sidePanel.selectedTemplate || sidePanel.useCustom" class="p-5 space-y-4">
                <!-- Back to templates (only in new mode) -->
                <button v-if="sidePanel.mode === 'new'" @click="backToTemplates"
                  class="flex items-center gap-1 text-xs text-gray-400 hover:text-indigo-600 transition-colors mb-1">
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
                  </svg>
                  返回模板选择
                </button>

                <!-- Chart Name -->
                <div>
                  <label class="block text-xs font-medium text-gray-500 mb-1.5">图表名称</label>
                  <input v-model="form.name" type="text" placeholder="输入图表名称"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all" />
                </div>

                <!-- Chart Type -->
                <div>
                  <label class="block text-xs font-medium text-gray-500 mb-1.5">图表类型</label>
                  <div class="grid grid-cols-5 gap-1.5">
                    <button v-for="ct in chartTypeOptions" :key="ct.value" @click="form.chart_type = ct.value"
                      :class="[
                        'flex flex-col items-center gap-1 px-2 py-2.5 rounded-lg border text-[11px] transition-all',
                        form.chart_type === ct.value
                          ? 'border-indigo-300 bg-indigo-50 text-indigo-700 font-medium'
                          : 'border-gray-100 bg-white text-gray-500 hover:border-gray-200 hover:bg-gray-50'
                      ]"
                    >
                      <span class="text-base">{{ ct.icon }}</span>
                      <span>{{ ct.label }}</span>
                    </button>
                  </div>
                </div>

                <!-- X Axis -->
                <div>
                  <label class="block text-xs font-medium text-gray-500 mb-1.5">X 轴（维度）</label>
                  <select v-model="form.x_axis"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all">
                    <optgroup label="分类维度">
                      <option v-for="d in xAxisCategoryOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup label="创建时间">
                      <option v-for="d in xAxisTimeCreatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup label="完成时间">
                      <option v-for="d in xAxisTimeCompletedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup label="更新时间">
                      <option v-for="d in xAxisTimeUpdatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                  </select>
                </div>

                <!-- Y Axis -->
                <div>
                  <label class="block text-xs font-medium text-gray-500 mb-1.5">Y 轴（指标）</label>
                  <select v-model="form.y_axis"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all">
                    <option v-for="m in yAxisOptions" :key="m.value" :value="m.value">{{ m.label }}</option>
                  </select>
                </div>

                <!-- Filters -->
                <div>
                  <label class="block text-xs font-medium text-gray-500 mb-1.5">筛选条件（可选）</label>
                  <textarea v-model="rqlQuery" rows="2" placeholder="例如: state = &quot;open&quot;"
                    class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono bg-gray-50 focus:bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all resize-none"></textarea>
                </div>

                <!-- Advanced Config -->
                <details class="group">
                  <summary class="flex items-center gap-1 text-xs font-medium text-gray-400 cursor-pointer hover:text-gray-600 transition-colors">
                    <svg class="w-3.5 h-3.5 transition-transform group-open:rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                    </svg>
                    高级配置
                  </summary>
                  <div class="mt-3 pl-5 space-y-3 border-l-2 border-gray-100">
                    <!-- Stack Mode -->
                    <div v-if="form.chart_type === 'bar' || form.chart_type === 'area'">
                      <label class="block text-[11px] text-gray-400 mb-1">堆叠模式</label>
                      <select v-model="advancedConfig.stack_mode" class="w-full px-2.5 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                        <option value="none">无</option>
                        <option value="stack">堆叠</option>
                        <option value="percent_stack">百分比堆叠</option>
                      </select>
                    </div>
                    <!-- Show Labels -->
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input type="checkbox" v-model="advancedConfig.show_labels"
                        class="w-3.5 h-3.5 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500" />
                      <span class="text-xs text-gray-500">显示数据标签</span>
                    </label>
                  </div>
                </details>
              </div>
            </div>

            <!-- Panel Footer -->
            <div class="flex items-center justify-end gap-2 px-5 py-3 border-t border-gray-100 shrink-0">
              <button @click="closeSidePanel"
                class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
                取消
              </button>
              <button @click="handleSave" :disabled="saving || !canSave"
                class="px-5 py-2 text-sm font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
                {{ saving ? '保存中...' : (sidePanel.mode === 'edit' ? '更新' : '创建') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import type { TemplateCategory, MetricTemplate, MetricChart, CreateChartPayload, MetricChartConfig } from '@/types/metrics'
import { metricsApi } from '@/api/metrics'
import { useConfirm } from '@/composables/useConfirm'
import MetricsChartCard from '@/components/metrics/MetricsChartCard.vue'

const props = defineProps<{ projectId: number }>()
const { confirm } = useConfirm()

// ── Data ──
const categories = ref<TemplateCategory[]>([])
const charts = ref<MetricChart[]>([])
const loading = ref(false)
const saving = ref(false)

// ── Side Panel State ──
const sidePanel = reactive({
  visible: false,
  mode: 'new' as 'new' | 'edit',
  activeCategory: 'agile',
  selectedTemplate: null as MetricTemplate | null,
  useCustom: false,
  editingChart: null as MetricChart | null,
})

// ── Form ──
const form = reactive({ name: '', chart_type: 'bar', x_axis: 'state', y_axis: 'count' })
const rqlQuery = ref('')
const advancedConfig = reactive<MetricChartConfig>({ stack_mode: 'none', show_labels: false, reference_lines: [] })

const canSave = computed(() => form.name.trim().length > 0)

// ── Options ──
const chartTypeOptions = [
  { value: 'bar', label: '柱状图', icon: '📊' },
  { value: 'line', label: '折线图', icon: '📈' },
  { value: 'pie', label: '饼图', icon: '🥧' },
  { value: 'doughnut', label: '环形图', icon: '🍩' },
  { value: 'area', label: '面积图', icon: '🏔' },
  { value: 'radar', label: '雷达图', icon: '🕷' },
  { value: 'scatter', label: '散点图', icon: '⚬' },
  { value: 'bubble', label: '气泡图', icon: '🫧' },
  { value: 'mixed', label: '混合图', icon: '📋' },
  { value: 'table', label: '表格', icon: '📑' },
]

const xAxisCategoryOptions = [
  { value: 'state', label: '状态' }, { value: 'priority', label: '优先级' },
  { value: 'assignee', label: '负责人' }, { value: 'type', label: '类型' },
  { value: 'label', label: '标签' }, { value: 'cycle', label: '周期' },
  { value: 'module', label: '模块' }, { value: 'state_group', label: '状态分组' },
  { value: 'created_by', label: '创建者' },
]
const xAxisTimeCreatedOptions = [
  { value: 'created_day', label: '按日' }, { value: 'created_week', label: '按周' }, { value: 'created_month', label: '按月' },
]
const xAxisTimeCompletedOptions = [
  { value: 'completed_day', label: '按日' }, { value: 'completed_week', label: '按周' }, { value: 'completed_month', label: '按月' },
]
const xAxisTimeUpdatedOptions = [
  { value: 'updated_day', label: '按日' }, { value: 'updated_week', label: '按周' }, { value: 'updated_month', label: '按月' },
]
const yAxisOptions = [
  { value: 'count', label: '数量' }, { value: 'avg_processing_time', label: '平均处理时间' },
  { value: 'current_retention', label: '当前留存' }, { value: 'created_vs_resolved', label: '创建 vs 解决' },
]

const currentCategoryTemplates = computed(() =>
  categories.value.find(c => c.id === sidePanel.activeCategory)?.templates || []
)

const templateIconMap: Record<string, string> = {
  flame: '🔥', 'trending-up': '📈', layers: '📚', clock: '⏱', timer: '⏲', columns: '📊',
  package: '📦', target: '🎯', 'alert-triangle': '⚠️', 'bar-chart-2': '📊', users: '👥',
  search: '🔍', shield: '🛡', 'check-circle': '✅', 'pause-circle': '⏸',
}

function getTemplateIcon(icon: string) { return templateIconMap[icon] || '📊' }

// ── Data Loading ──
async function loadData() {
  loading.value = true
  try {
    const [tplData, chartData] = await Promise.all([
      metricsApi.listTemplates(props.projectId),
      metricsApi.listCharts(props.projectId),
    ])
    categories.value = tplData.categories || tplData
    charts.value = chartData.charts || chartData
  } catch (e) {
    console.error('Failed to load metrics data:', e)
  } finally {
    loading.value = false
  }
}

// ── Side Panel Actions ──
function openSidePanel(mode: 'new' | 'edit', chart?: MetricChart) {
  sidePanel.mode = mode
  sidePanel.visible = true
  sidePanel.selectedTemplate = null
  sidePanel.useCustom = false
  sidePanel.editingChart = null

  if (mode === 'edit' && chart) {
    sidePanel.editingChart = chart
    form.name = chart.name
    form.chart_type = chart.chart_type
    form.x_axis = chart.x_axis
    form.y_axis = chart.y_axis
    rqlQuery.value = (() => { try { return JSON.parse(chart.filters || '{}').rql || '' } catch { return '' } })()
    try {
      const cfg = JSON.parse(chart.config || '{}')
      advancedConfig.stack_mode = cfg.stack_mode || 'none'
      advancedConfig.show_labels = cfg.show_labels || false
      advancedConfig.reference_lines = cfg.reference_lines || []
    } catch { /* ignore */ }
  } else {
    resetForm()
  }
}

function closeSidePanel() {
  sidePanel.visible = false
  sidePanel.selectedTemplate = null
  sidePanel.useCustom = false
  sidePanel.editingChart = null
}

function selectTemplate(tpl: MetricTemplate) {
  sidePanel.selectedTemplate = tpl
  form.name = tpl.name
  form.chart_type = tpl.chart_type
  form.x_axis = tpl.default_x_axis
  form.y_axis = tpl.default_y_axis
  if (tpl.default_config) {
    advancedConfig.stack_mode = tpl.default_config.stack_mode || 'none'
    advancedConfig.show_labels = tpl.default_config.show_labels || false
    advancedConfig.reference_lines = tpl.default_config.reference_lines || []
  }
}

function backToTemplates() {
  sidePanel.selectedTemplate = null
  sidePanel.useCustom = false
  resetForm()
}

function resetForm() {
  form.name = ''
  form.chart_type = 'bar'
  form.x_axis = 'state'
  form.y_axis = 'count'
  rqlQuery.value = ''
  advancedConfig.stack_mode = 'none'
  advancedConfig.show_labels = false
  advancedConfig.reference_lines = []
}

// ── Save ──
async function handleSave() {
  if (!canSave.value || saving.value) return
  saving.value = true
  try {
    const payload: CreateChartPayload = {
      name: form.name.trim(),
      chart_type: form.chart_type,
      x_axis: form.x_axis,
      y_axis: form.y_axis,
      filters: rqlQuery.value ? { rql: rqlQuery.value } : undefined,
      config: { stack_mode: advancedConfig.stack_mode, show_labels: advancedConfig.show_labels, reference_lines: advancedConfig.reference_lines },
    }
    if (sidePanel.selectedTemplate) payload.template_id = sidePanel.selectedTemplate.id

    if (sidePanel.mode === 'edit' && sidePanel.editingChart) {
      await metricsApi.updateChart(props.projectId, sidePanel.editingChart.id, payload)
    } else {
      await metricsApi.createChart(props.projectId, payload)
    }
    closeSidePanel()
    await loadData()
  } catch (e) {
    console.error('Failed to save chart:', e)
  } finally {
    saving.value = false
  }
}

// ── Delete ──
async function handleDeleteChart(chartId: number) {
  const ok = await confirm({ title: '删除图表', message: '确定要删除该图表吗？', confirmText: '删除', danger: true })
  if (!ok) return
  try {
    await metricsApi.deleteChart(props.projectId, chartId)
    await loadData()
  } catch (e) {
    console.error('Failed to delete chart:', e)
  }
}

// ── Init ──
onMounted(loadData)
</script>

<style scoped>
@keyframes slide-in { from { transform: translateX(100%); } to { transform: translateX(0); } }
.animate-slide-in { animation: slide-in 0.25s ease-out; }
.slide-enter-active, .slide-leave-active { transition: opacity 0.2s; }
.slide-enter-from, .slide-leave-to { opacity: 0; }
.line-clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
</style>
