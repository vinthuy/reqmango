<template>
  <div class="h-full flex flex-col bg-gray-50/50">
    <!-- Header Bar -->
    <div class="flex items-center justify-between px-6 py-3 bg-white border-b border-gray-200 shrink-0">
      <div class="flex items-center gap-3">
        <h2 class="text-base font-semibold text-gray-800">度量</h2>
        <span class="text-xs text-gray-400">{{ charts.length }} 个图表</span>
      </div>
      <button @click="openPanel('new')"
        class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        添加图表
      </button>
    </div>

    <!-- Main Content -->
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
        <button @click="openPanel('new')"
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
          @edit="openPanel('edit', chart)"
          @delete="handleDeleteChart(chart.id)"
        />
      </div>
    </div>

    <!-- Side Panel -->
    <Teleport to="body">
      <div v-if="panel.visible" class="fixed inset-0 z-50 flex justify-end">
        <div class="absolute inset-0 bg-black/30" @click="closePanel"></div>
        <div class="relative w-[680px] bg-white shadow-2xl flex flex-col animate-slide-in">
          <!-- Header -->
          <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 shrink-0">
            <h3 class="text-sm font-semibold text-gray-800">
              {{ panel.mode === 'edit' ? '编辑图表' : '添加图表' }}
            </h3>
            <button @click="closePanel" class="p-1 text-gray-400 hover:text-gray-600 rounded">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <!-- Body: Two-column layout -->
          <div class="flex-1 flex overflow-hidden">
            <!-- Left: Config -->
            <div class="w-[340px] border-r border-gray-100 overflow-y-auto p-4 space-y-4">
              <!-- Step 1: Template Selection (new mode only) -->
              <div v-if="panel.mode === 'new' && !panel.selectedTemplate && !panel.useCustom">
                <p class="text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-2">选择模板</p>
                <div class="flex gap-1 mb-3 bg-gray-100 rounded-lg p-0.5">
                  <button v-for="cat in categories" :key="cat.id"
                    @click="panel.activeCategory = cat.id"
                    :class="['flex-1 px-2 py-1 text-[11px] font-medium rounded-md transition-colors',
                      panel.activeCategory === cat.id ? 'bg-white shadow-sm text-gray-800' : 'text-gray-500 hover:text-gray-700']"
                  >{{ cat.name }}</button>
                </div>
                <div class="space-y-1.5">
                  <button v-for="tpl in currentCategoryTemplates" :key="tpl.id"
                    @click="selectTemplate(tpl)"
                    class="w-full flex items-center gap-2.5 p-2.5 bg-gray-50 hover:bg-indigo-50 hover:border-indigo-200 border border-gray-100 rounded-lg text-left transition-colors group"
                  >
                    <span class="text-base">{{ getTemplateIcon(tpl.icon) }}</span>
                    <div class="min-w-0 flex-1">
                      <p class="text-xs font-medium text-gray-700 group-hover:text-indigo-700 truncate">{{ tpl.name }}</p>
                      <p class="text-[10px] text-gray-400 truncate">{{ tpl.description }}</p>
                    </div>
                    <svg class="w-3.5 h-3.5 text-gray-300 group-hover:text-indigo-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                    </svg>
                  </button>
                </div>
                <div class="mt-3 pt-3 border-t border-gray-100">
                  <button @click="panel.useCustom = true; panel.selectedTemplate = null; resetForm()"
                    class="w-full flex items-center justify-center gap-2 py-2 border border-dashed border-gray-300 rounded-lg text-xs text-gray-500 hover:text-indigo-600 hover:border-indigo-300 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                    </svg>
                    自定义创建
                  </button>
                </div>
              </div>

              <!-- Step 2: Configuration Form -->
              <div v-if="panel.mode === 'edit' || panel.selectedTemplate || panel.useCustom">
                <button v-if="panel.mode === 'new'" @click="backToTemplates"
                  class="flex items-center gap-1 text-[11px] text-gray-400 hover:text-indigo-600 transition-colors mb-2">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
                  </svg>
                  返回模板
                </button>

                <!-- Chart Name -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">图表名称</label>
                  <input v-model="form.name" type="text" placeholder="输入图表名称"
                    class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
                </div>

                <!-- Chart Type -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">
                    图表类型
                    <span v-if="isTemplateMode" class="text-indigo-400 ml-1">(模板锁定)</span>
                    <button v-if="isTemplateMode" @click="panel.selectedTemplate = null"
                      class="ml-2 text-[10px] text-gray-400 hover:text-indigo-500 underline">切换为自定义</button>
                  </label>
                  <div v-if="isTemplateMode" class="flex items-center gap-2 px-3 py-2 bg-indigo-50 rounded-lg border border-indigo-100">
                    <span class="text-sm">{{ chartTypeOptions.find(c => c.value === form.chart_type)?.icon }}</span>
                    <span class="text-[11px] text-indigo-700 font-medium">{{ chartTypeOptions.find(c => c.value === form.chart_type)?.label }}</span>
                    <svg class="w-3 h-3 text-indigo-400 ml-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                  </div>
                  <div v-else class="grid grid-cols-5 gap-1">
                    <button v-for="ct in chartTypeOptions" :key="ct.value" @click="updateChartType(ct.value)"
                      :class="['flex flex-col items-center gap-0.5 px-1 py-1.5 rounded-md border text-[10px] transition-all',
                        form.chart_type === ct.value
                          ? 'border-indigo-300 bg-indigo-50 text-indigo-700 font-medium'
                          : 'border-gray-100 bg-white text-gray-500 hover:border-gray-200']"
                    >
                      <span class="text-sm">{{ ct.icon }}</span>
                      <span>{{ ct.label }}</span>
                    </button>
                  </div>
                </div>

                <!-- X Axis -->
                <div>
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">
                    X 轴（维度）
                    <span v-if="isTemplateMode" class="text-indigo-400 ml-1">(模板锁定)</span>
                  </label>
                  <div v-if="isTemplateMode" class="flex items-center gap-2 px-3 py-2 bg-indigo-50 rounded-lg border border-indigo-100">
                    <span class="text-[11px] text-indigo-700 font-medium">{{ form.x_axis }}</span>
                    <svg class="w-3 h-3 text-indigo-400 ml-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                  </div>
                  <select v-else v-model="form.x_axis"
                    class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
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
                  <label class="block text-[11px] font-medium text-gray-500 mb-1">Y 轴（指标）</label>
                  <div v-if="isTemplateMode" class="flex items-center gap-2 px-3 py-2 bg-indigo-50 rounded-lg border border-indigo-100">
                    <span class="text-[11px] text-indigo-700 font-medium">{{ yAxisOptions.find(m => m.value === form.y_axis)?.label || form.y_axis }}</span>
                    <svg class="w-3 h-3 text-indigo-400 ml-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/></svg>
                  </div>
                  <select v-else v-model="form.y_axis"
                    class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent">
                    <option v-for="m in yAxisOptions" :key="m.value" :value="m.value">{{ m.label }}</option>
                  </select>
                </div>

                <!-- Jira-style Filters -->
                <div>
                  <div class="flex items-center justify-between mb-2">
                    <label class="text-[11px] font-medium text-gray-500">筛选条件</label>
                    <button @click="addFilter" class="inline-flex items-center gap-0.5 text-[11px] text-indigo-500 hover:text-indigo-700">
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
                      添加条件
                    </button>
                  </div>
                  <div v-if="filters.length === 0" class="flex items-center gap-2 px-3 py-2 bg-gray-50 rounded-lg border border-dashed border-gray-200">
                    <svg class="w-4 h-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/></svg>
                    <span class="text-[11px] text-gray-400">无筛选条件 — 显示全部数据</span>
                  </div>
                  <div v-for="(f, i) in filters" :key="i" class="relative mb-2">
                    <div class="flex items-center gap-1.5">
                      <!-- AND badge -->
                      <span v-if="i > 0" class="text-[9px] font-bold text-indigo-400 bg-indigo-50 px-1.5 py-0.5 rounded shrink-0">并且</span>
                      <!-- Field selector -->
                      <select v-model="f.field" @change="f.values = []; f.value = ''"
                        class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                        <option value="">选择字段</option>
                        <option value="state">状态</option>
                        <option value="priority">优先级</option>
                        <option value="assignee">负责人</option>
                        <option value="type">类型</option>
                        <option value="label">标签</option>
                        <option value="module">模块</option>
                        <option value="created_by">创建人</option>
                        <option value="title">标题</option>
                      </select>
                      <!-- Operator selector -->
                      <select v-model="f.operator" @change="f.values = []; f.value = ''"
                        class="w-24 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                        <option value="=">等于</option>
                        <option value="!=">不等于</option>
                        <option value="in">包含任一</option>
                        <option value="not_in">不包含任一</option>
                        <option value="contains">包含文本</option>
                        <option value="empty">为空</option>
                        <option value="not_empty">不为空</option>
                      </select>
                      <!-- Value: Multi-select for in/not_in -->
                      <div v-if="f.operator === 'in' || f.operator === 'not_in'" class="flex-1 relative">
                        <div @click="$event.stopPropagation()"
                          class="flex items-center gap-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white cursor-pointer hover:border-indigo-300 min-h-[30px] flex-wrap">
                          <span v-for="(v, vi) in f.values" :key="vi"
                            class="inline-flex items-center gap-0.5 bg-indigo-50 text-indigo-700 px-1.5 py-0.5 rounded text-[10px]">
                            {{ v }}
                            <button @click="f.values.splice(vi, 1)" class="text-indigo-400 hover:text-indigo-600">&times;</button>
                          </span>
                          <span v-if="f.values.length === 0" class="text-gray-400">选择多个值</span>
                        </div>
                        <div class="absolute z-50 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-40 overflow-y-auto">
                          <label v-for="v in getFilterFieldValues(f.field)" :key="v"
                            class="flex items-center gap-2 px-3 py-1.5 text-[11px] hover:bg-gray-50 cursor-pointer">
                            <input type="checkbox" :value="v" v-model="f.values"
                              class="w-3 h-3 rounded border-gray-300 text-indigo-500 focus:ring-indigo-400">
                            <span class="text-gray-700">{{ v }}</span>
                          </label>
                        </div>
                      </div>
                      <!-- Value: Single select/input for other operators -->
                      <template v-else-if="f.operator !== 'empty' && f.operator !== 'not_empty'">
                        <select v-if="getFilterFieldValues(f.field).length > 0"
                          v-model="f.value"
                          class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-indigo-400">
                          <option value="">选择值</option>
                          <option v-for="v in getFilterFieldValues(f.field)" :key="v" :value="v">{{ v }}</option>
                        </select>
                        <input v-else v-model="f.value" placeholder="输入值"
                          class="flex-1 px-2 py-1.5 border border-gray-200 rounded text-[11px] focus:outline-none focus:ring-1 focus:ring-indigo-400" />
                      </template>
                      <!-- Remove button -->
                      <button @click="removeFilter(i)" class="p-1 text-gray-300 hover:text-red-500 shrink-0 rounded hover:bg-red-50 transition-colors">
                        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Right: Preview -->
            <div class="flex-1 flex flex-col bg-gray-50">
              <div class="flex items-center justify-between px-4 py-2 border-b border-gray-100 bg-white shrink-0">
                <span class="text-[11px] font-medium text-gray-400">图表预览</span>
                <button @click="fetchPreview" :disabled="previewLoading"
                  class="px-3 py-1 text-[11px] font-medium text-indigo-600 bg-indigo-50 rounded-md hover:bg-indigo-100 disabled:opacity-50 transition-colors">
                  {{ previewLoading ? '加载中...' : '预览' }}
                </button>
              </div>
              <div class="flex-1 flex items-center justify-center p-4">
                <div v-if="!hasData && !previewLoading" class="text-center">
                  <svg class="w-10 h-10 text-gray-200 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                  </svg>
                  <p class="text-xs text-gray-400">{{ previewError || '配置完成后点击「预览」查看效果' }}</p>
                </div>
                <div v-else-if="previewLoading" class="text-center">
                  <svg class="animate-spin h-6 w-6 text-indigo-300 mx-auto mb-2" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  <p class="text-xs text-gray-400">加载预览数据...</p>
                </div>
                <div v-else class="w-full h-full">
                  <MetricsChartCard
                    :chart="previewChart"
                    :project-id="projectId"
                    class="shadow-none border-0 h-full"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-2 px-5 py-3 border-t border-gray-100 shrink-0">
            <button @click="closePanel"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
              取消
            </button>
            <button @click="handleSave" :disabled="saving || !canSave"
              class="px-5 py-2 text-sm font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
              {{ saving ? '保存中...' : (panel.mode === 'edit' ? '更新' : '创建') }}
            </button>
          </div>
        </div>
      </div>
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

// ── Side Panel ──
const panel = reactive({
  visible: false,
  mode: 'new' as 'new' | 'edit',
  activeCategory: 'agile',
  selectedTemplate: null as MetricTemplate | null,
  useCustom: false,
  editingChart: null as MetricChart | null,
})

// ── Form ──
const form = reactive({ name: '', chart_type: 'bar', x_axis: 'state', y_axis: 'count' })
const filters = reactive<Array<{ field: string; operator: string; value: string; values: string[] }>>([])
const advancedConfig = reactive<MetricChartConfig>({ stack_mode: 'none', show_labels: false, reference_lines: [] })
const canSave = computed(() => form.name.trim().length > 0)
const isTemplateMode = computed(() => !!panel.selectedTemplate && panel.useCustom)

// ── Filter Values (for dropdown) ──
const filterValues = ref<Record<string, string[]>>({})
async function loadFilterValues() {
  try {
    filterValues.value = await metricsApi.getFilterValues(props.projectId)
  } catch { /* ignore */ }
}

function getFilterFieldValues(field: string): string[] {
  return filterValues.value[field] || []
}

// ── Preview ──
const previewData = ref<{ labels: string[]; values: number[]; colors?: string[] } | null>(null)
const previewLoading = ref(false)
const previewError = ref('')
const hasData = computed(() => previewData.value && previewData.value.labels.length > 0)

const previewChart = computed<MetricChart>(() => ({
  id: 0, project_id: props.projectId, creator_id: 0, name: form.name || '预览',
  chart_type: form.chart_type, x_axis: form.x_axis, y_axis: form.y_axis,
  filters: '', config: '', template_id: '',
  data_labels: previewData.value?.labels || [],
  data_values: previewData.value?.values || [],
  data_colors: previewData.value?.colors || [],
  sort_order: 0, is_visible: true, created_at: '', updated_at: '',
}))

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
  { value: 'count', label: '数量（Issue 总数）' },
  { value: 'avg_processing_time', label: '平均处理时间（天）' },
  { value: 'current_retention', label: '当前留存时间（天）' },
  { value: 'created_vs_resolved', label: '创建 vs 解决' },
  { value: 'completion_rate', label: '完成率（%）' },
  { value: 'avg_cycle_time', label: '平均周期时间（天）' },
  { value: 'throughput', label: '吞吐量（完成数/周）' },
  { value: 'wip_count', label: '在制品数量' },
  { value: 'backlog_count', label: '待办数量' },
  { value: 'overdue_count', label: '逾期数量' },
]
const currentCategoryTemplates = computed(() =>
  categories.value.find(c => c.id === panel.activeCategory)?.templates || []
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

// ── Panel Actions ──
function openPanel(mode: 'new' | 'edit', chart?: MetricChart) {
  panel.mode = mode
  panel.visible = true
  panel.selectedTemplate = null
  panel.useCustom = false
  panel.editingChart = null
  previewData.value = null
  previewError.value = ''
  filters.splice(0)
  loadFilterValues()

  if (mode === 'edit' && chart) {
    panel.editingChart = chart
    panel.useCustom = true
    form.name = chart.name
    form.chart_type = chart.chart_type
    form.x_axis = chart.x_axis
    form.y_axis = chart.y_axis
    // Parse filters from JSON
    try {
      const f = JSON.parse(chart.filters || '{}')
      if (f.conditions) {
        f.conditions.forEach((c: any) => filters.push({ field: c.field || '', operator: c.operator || '=', value: c.value || '', values: [] }))
      } else if (f.rql) {
        // Convert legacy RQL to visual filters (best effort)
        const match = f.rql.match(/^(\w+)\s*(!?=)\s*"?([^"]+)"?$/)
        if (match) filters.push({ field: match[1], operator: match[2], value: match[3], values: [] })
      }
    } catch { /* ignore */ }
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

function closePanel() {
  panel.visible = false
  panel.selectedTemplate = null
  panel.useCustom = false
  panel.editingChart = null
  previewData.value = null
}

function selectTemplate(tpl: MetricTemplate) {
  panel.selectedTemplate = tpl
  panel.useCustom = true
  form.name = tpl.name
  form.chart_type = tpl.chart_type
  form.x_axis = tpl.default_x_axis
  form.y_axis = tpl.default_y_axis
  filters.splice(0)
  if (tpl.default_config) {
    advancedConfig.stack_mode = tpl.default_config.stack_mode || 'none'
    advancedConfig.show_labels = tpl.default_config.show_labels || false
    advancedConfig.reference_lines = tpl.default_config.reference_lines || []
  }
}

function backToTemplates() {
  panel.selectedTemplate = null
  panel.useCustom = false
  previewData.value = null
  resetForm()
}

function resetForm() {
  form.name = ''
  form.chart_type = 'bar'
  form.x_axis = 'state'
  form.y_axis = 'count'
  filters.splice(0)
  advancedConfig.stack_mode = 'none'
  advancedConfig.show_labels = false
  advancedConfig.reference_lines = []
}

function updateChartType(type: string) {
  form.chart_type = type
}

function addFilter() {
  filters.push({ field: '', operator: '=', value: '', values: [] })
}

function removeFilter(index: number) {
  filters.splice(index, 1)
}

// ── Preview ──

async function fetchPreview() {
  if (!form.x_axis || !form.y_axis) return
  previewLoading.value = true
  previewError.value = ''
  try {
    const payload: any = {
      name: form.name || 'preview',
      chart_type: form.chart_type,
      x_axis: form.x_axis,
      y_axis: form.y_axis,
    }
    // Build filters
    const validFilters = filters.filter(f => {
      if (!f.field) return false
      if (f.operator === 'empty' || f.operator === 'not_empty') return true
      if (f.operator === 'in' || f.operator === 'not_in') return f.values.length > 0
      return f.value !== ''
    })
    if (validFilters.length > 0) {
      const rqlParts = validFilters.map(f => {
        if (f.operator === 'empty') return `${f.field} = ""`
        if (f.operator === 'not_empty') return `${f.field} != ""`
        if (f.operator === 'in') return `(${f.values.map(v => `${f.field} = "${v}"`).join(' OR ')})`
        if (f.operator === 'not_in') return `(${f.values.map(v => `${f.field} != "${v}"`).join(' AND ')})`
        if (f.operator === 'contains') return `${f.field} ~ "${f.value}"`
        return `${f.field} ${f.operator} "${f.value}"`
      })
      payload.filters = { rql: rqlParts.join(' AND ') }
    }
    const res = await metricsApi.previewChart(props.projectId, payload)
    previewData.value = { labels: res.labels || [], values: res.values || [], colors: res.colors || [] }
  } catch (e: any) {
    previewError.value = e?.response?.data?.error || '预览失败'
    previewData.value = null
  } finally {
    previewLoading.value = false
  }
}

// ── Save ──
async function handleSave() {
  if (!canSave.value || saving.value) return
  saving.value = true
  try {
    const validFilters = filters.filter(f => f.field && (f.operator === 'empty' || f.operator === 'not_empty' || f.value))
    let filtersPayload: any = undefined
    if (validFilters.length > 0) {
      // Save as RQL for backend compatibility
      const rqlParts = validFilters.map(f => {
        if (f.operator === 'empty') return `${f.field} = ""`
        if (f.operator === 'not_empty') return `${f.field} != ""`
        return `${f.field} ${f.operator} "${f.value}"`
      })
      filtersPayload = { rql: rqlParts.join(' AND '), conditions: validFilters }
    }
    const payload: CreateChartPayload = {
      name: form.name.trim(),
      chart_type: form.chart_type,
      x_axis: form.x_axis,
      y_axis: form.y_axis,
      filters: filtersPayload,
      config: { stack_mode: advancedConfig.stack_mode, show_labels: advancedConfig.show_labels, reference_lines: advancedConfig.reference_lines },
    }
    if (panel.selectedTemplate) payload.template_id = panel.selectedTemplate.id

    if (panel.mode === 'edit' && panel.editingChart) {
      await metricsApi.updateChart(props.projectId, panel.editingChart.id, payload)
    } else {
      await metricsApi.createChart(props.projectId, payload)
    }
    closePanel()
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

onMounted(loadData)
</script>

<style scoped>
@keyframes slide-in { from { transform: translateX(100%); } to { transform: translateX(0); } }
.animate-slide-in { animation: slide-in 0.25s ease-out; }
</style>
