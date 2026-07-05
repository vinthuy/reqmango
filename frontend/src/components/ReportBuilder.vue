<template>
  <div class="space-y-4">
    <!-- Tab Header -->
    <div class="flex items-center gap-1 border-b border-gray-200">
      <button v-for="tab in tabs" :key="tab.key"
        @click="activeTab = tab.key as 'quick' | 'custom'"
        :class="['px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px',
          activeTab === tab.key ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300']"
      >{{ tab.label }}</button>
    </div>

    <!-- ═══ QUICK CHARTS TAB ═══ -->
    <template v-if="activeTab === 'quick'">
      <!-- Config Bar -->
      <div class="bg-white border border-gray-100 rounded-xl px-5 py-3">
        <div class="flex flex-wrap items-end gap-3">
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.reportType') }}</label>
            <select v-model="reportType" @change="onTypeChange" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
              <option v-for="(label, key) in reportTypeLabels" :key="key" :value="key">{{ label }}</option>
            </select>
          </div>
          <div v-if="reportType === 'created_vs_resolved' || reportType === 'created_trend'">
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.interval') }}</label>
            <select v-model="interval" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
              <option v-for="(label, key) in intervalLabels" :key="key" :value="key">{{ label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.chart') }}</label>
            <div class="inline-flex bg-gray-100 rounded-lg p-0.5">
              <button v-for="c in availableCharts" :key="c" @click="chartType = c"
                :class="['px-2.5 py-1 text-xs rounded-md transition-colors', chartType === c ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
              >{{ (chartLabels as Record<string, string>)[c] || c }}</button>
            </div>
          </div>
          <div><label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.dateFrom') }}</label><input v-model="dateFrom" type="date" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" /></div>
          <div><label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.dateTo') }}</label><input v-model="dateTo" type="date" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" /></div>
          <button @click="generate" :disabled="loading" class="px-4 py-1.5 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 disabled:opacity-50 transition-colors self-end mb-0.5">
            {{ loading ? '...' : t('report.generate') }}
          </button>
          <button @click="showSaveFilterDialog = true" class="px-3 py-1.5 border border-gray-200 text-gray-600 text-sm rounded-md hover:bg-gray-50 transition-colors self-end mb-0.5">
            {{ t('report.save') }}
          </button>
        </div>
      </div>
    </template>

    <!-- ═══ CUSTOM REPORTS TAB (Jira-style) ═══ -->
    <template v-if="activeTab === 'custom'">
      <div class="flex gap-4 min-h-[600px]">
        <!-- Left: Saved Filters -->
        <div class="w-52 shrink-0 border border-gray-100 rounded-xl bg-white h-fit">
          <div class="px-4 py-3 border-b border-gray-100">
            <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.savedFilters') }}</h3>
          </div>
          <div class="divide-y divide-gray-50 max-h-[400px] overflow-y-auto">
            <button v-for="f in savedFilters" :key="f.id" @click="loadSavedFilter(f)"
              class="w-full text-left px-4 py-2.5 text-sm hover:bg-gray-50 transition-colors flex items-center justify-between group"
              :class="{ 'bg-blue-50 text-blue-700': selectedFilterId === f.id }">
              <span class="truncate">{{ f.name }}</span>
              <button @click.stop="deleteSavedFilter(f)" class="text-gray-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition">&times;</button>
            </button>
            <div v-if="savedFilters.length === 0 && !loadingSaved" class="px-4 py-6 text-center text-xs text-gray-400">
              {{ t('report.noSavedFilters') }}
            </div>
          </div>
        </div>

        <!-- Right: Filter + Report Config + Results -->
        <div class="flex-1 space-y-4">

          <!-- ── Filter Builder (Jira-style) ── -->
          <div class="bg-white border border-gray-100 rounded-xl px-5 py-4">
            <div class="flex items-center justify-between mb-3">
              <h4 class="text-sm font-medium text-gray-700">{{ t('report.filterConditions') }}</h4>
              <div class="flex items-center gap-2">
                <button @click="filterMode = 'basic'" :class="['text-xs px-2 py-0.5 rounded', filterMode === 'basic' ? 'bg-gray-100 text-gray-800 font-medium' : 'text-gray-400 hover:text-gray-600']">{{ t('report.visualFilter') }}</button>
                <span class="text-gray-300">|</span>
                <button @click="filterMode = 'rql'" :class="['text-xs px-2 py-0.5 rounded', filterMode === 'rql' ? 'bg-gray-100 text-gray-800 font-medium' : 'text-gray-400 hover:text-gray-600']">{{ t('report.rqlAdvanced') }}</button>
              </div>
            </div>

            <!-- Visual Filter -->
            <div v-if="filterMode === 'basic'" class="space-y-2">
              <div v-for="(filter, idx) in filters" :key="idx" class="flex items-center gap-2">
                <span v-if="idx > 0" class="text-[11px] font-medium text-gray-400 w-8 shrink-0">{{ t('report.and') }}</span>
                <span v-else class="w-8 shrink-0"></span>
                <select v-model="filter.field" @change="onFilterFieldChange(idx)" class="flex-1 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option value="">{{ t('report.selectField') }}</option>
                  <option v-for="f in filterFields" :key="f.value" :value="f.value">{{ f.label }}</option>
                </select>
                <select v-if="filter.field" v-model="filter.operator" class="w-28 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option v-for="op in filterOperators" :key="op.value" :value="op.value">{{ op.label }}</option>
                </select>
                <select v-if="filter.field && filter.operator !== 'empty' && filter.operator !== 'not_empty'" v-model="filter.value" class="flex-1 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option value="">{{ t('report.selectValue') }}</option>
                  <option v-for="o in getFilterOptions(filter.field)" :key="o.value" :value="o.value">{{ o.label }}</option>
                </select>
                <span v-else class="flex-1"></span>
                <button @click="removeFilter(idx)" class="text-gray-300 hover:text-red-500 text-lg leading-none px-1 transition-colors">&times;</button>
              </div>
              <button @click="addFilter" class="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 mt-2 transition-colors">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/></svg>
                {{ t('report.addCondition') }}
              </button>
            </div>

            <!-- RQL Editor -->
            <div v-else>
              <textarea v-model="rqlQuery" rows="3" :placeholder="t('report.rqlPlaceholder')"
                class="w-full px-3 py-2 border border-gray-200 rounded-md text-sm font-mono bg-gray-50 focus:bg-white focus:outline-none focus:ring-1 focus:ring-blue-400 transition-colors resize-none"></textarea>
            </div>

            <!-- RQL Preview -->
            <div v-if="filterMode === 'basic' && builtRqlPreview" class="mt-3 px-3 py-2 bg-gray-50 rounded-md text-xs font-mono text-gray-500">
              <span class="text-gray-400">RQL:</span> {{ builtRqlPreview }}
            </div>

            <!-- Error -->
            <div v-if="rqlError" class="mt-3 flex items-center gap-2 px-3 py-2 bg-red-50 border border-red-100 rounded-md text-xs text-red-600">
              <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
              <span class="flex-1">{{ rqlError }}</span>
              <button @click="rqlError = null" class="text-red-400 hover:text-red-600">&times;</button>
            </div>
          </div>

          <!-- ── Report Config ── -->
          <div class="bg-white border border-gray-100 rounded-xl px-5 py-3">
            <div class="flex flex-wrap items-end gap-3">
              <div>
                <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.reportType') }}</label>
                <select v-model="reportType" @change="onTypeChange" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option v-for="(label, key) in reportTypeLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
              <div v-if="reportType !== 'created_vs_resolved' && reportType !== 'created_trend'">
                <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.groupBy') }}</label>
                <select v-model="groupBy" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option v-for="d in dims" :key="d.value" :value="d.value">{{ d.label }}</option>
                </select>
              </div>
              <div v-if="reportType === 'created_vs_resolved' || reportType === 'created_trend'">
                <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.interval') }}</label>
                <select v-model="interval" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option v-for="(label, key) in intervalLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
              <div>
                <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.chart') }}</label>
                <div class="inline-flex bg-gray-100 rounded-lg p-0.5">
                  <button v-for="c in availableCharts" :key="c" @click="chartType = c"
                    :class="['px-2.5 py-1 text-xs rounded-md transition-colors', chartType === c ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
                  >{{ (chartLabels as Record<string, string>)[c] || c }}</button>
                </div>
              </div>
              <div><label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.dateFrom') }}</label><input v-model="dateFrom" type="date" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" /></div>
              <div><label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.dateTo') }}</label><input v-model="dateTo" type="date" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" /></div>
              <button @click="generate" :disabled="loading" class="px-4 py-1.5 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 disabled:opacity-50 transition-colors self-end mb-0.5">
                {{ loading ? '...' : t('report.generate') }}
              </button>
              <button @click="showSaveFilterDialog = true" class="px-3 py-1.5 border border-gray-200 text-gray-600 text-sm rounded-md hover:bg-gray-50 transition-colors self-end mb-0.5">
                {{ t('report.saveFilter') }}
              </button>
            </div>
          </div>

          <!-- ── Results ── -->
          <div v-if="data" class="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <div class="p-4 bg-white rounded-xl border border-gray-100">
              <div class="text-xs text-gray-400 mb-1">{{ t('report.quick.totalIssues') }}</div>
              <div class="text-2xl font-bold text-gray-800">{{ data.total }}</div>
            </div>
            <div v-if="data.summary?.avg_days" class="p-4 bg-white rounded-xl border border-gray-100">
              <div class="text-xs text-gray-400 mb-1">{{ t('report.avg') }}</div>
              <div class="text-2xl font-bold text-gray-800">{{ data.summary.avg_days.toFixed(1) }} <span class="text-sm font-normal text-gray-400">{{ t('report.days') }}</span></div>
            </div>
          </div>
          <div v-if="data" class="flex items-center justify-between px-5 py-2 border border-gray-100 rounded-xl bg-white text-xs text-gray-500">
            <span>{{ t('report.matched') }}: <strong class="text-gray-800">{{ data.total }}</strong> {{ t('report.issues') }}</span>
            <div class="flex items-center gap-2">
              <button @click="exportCSV" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors">CSV</button>
              <button v-if="chartType !== 'Table'" @click="exportPNG" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors">PNG</button>
            </div>
          </div>
          <div v-if="loading" class="flex items-center justify-center py-20 bg-white border border-gray-100 rounded-xl text-gray-400 text-sm">
            {{ t('report.loading') }}
          </div>
          <template v-else-if="data">
            <div v-show="chartType !== 'Table'" class="bg-white border border-gray-100 rounded-xl p-5">
              <div :class="['mx-auto', chartType === 'Pie' || chartType === 'Doughnut' ? 'max-w-md' : 'max-w-3xl']" style="height: 360px">
                <canvas :ref="setChartCanvas"></canvas>
              </div>
            </div>
            <div v-show="chartType === 'Table'" class="bg-white border border-gray-100 rounded-xl p-5">
              <table class="w-full text-sm">
                <thead><tr class="border-b border-gray-100">
                  <th class="text-left py-2 text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.groupBy') }}</th>
                  <th class="text-right py-2 text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.count') }}</th>
                  <th class="text-right py-2 text-xs font-medium text-gray-400 uppercase tracking-wide w-20">{{ t('report.percent') }}</th>
                </tr></thead>
                <tbody>
                  <tr v-for="(label, i) in data.labels" :key="i" class="border-b border-gray-50 hover:bg-gray-50/50">
                    <td class="py-2 flex items-center gap-2">
                      <span class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ backgroundColor: chartColors[i % chartColors.length] }"></span>{{ label }}
                    </td>
                    <td class="text-right py-2 font-medium">{{ data.values[i] }}</td>
                    <td class="text-right py-2 text-gray-500">{{ pct(data.values[i]) }}%</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>
          <div v-if="!loading && !data" class="flex flex-col items-center justify-center py-20 bg-white border border-gray-100 rounded-xl text-gray-400">
            <svg class="w-12 h-12 mb-3 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
            <span class="text-sm">{{ t('report.emptyState') }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- Save Filter Dialog -->
    <div v-if="showSaveFilterDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @click.self="showSaveFilterDialog = false">
      <div class="bg-white rounded-xl shadow-xl w-96 p-6">
        <h3 class="text-lg font-semibold mb-4">{{ t('report.saveFilter') }}</h3>
        <input v-model="filterName" type="text" :placeholder="t('report.filterNamePlaceholder')"
          class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 mb-4" />
        <div class="flex justify-end gap-2">
          <button @click="showSaveFilterDialog = false" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg text-sm">{{ t('common.cancel') }}</button>
          <button @click="saveFilter" :disabled="savingFilter || !filterName.trim()" class="px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-50">{{ savingFilter ? '...' : t('common.save') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'
import { reportApi, savedReportApi } from '@/api/report'
import type { ReportResponse, SavedReport } from '@/api/report'
import { useReportChart, exportReportCSV, exportChartPNG } from '@/composables/useReportChart'
import api from '@/api'

const props = defineProps<{ projectId: number }>()
const { t } = useI18n()
const toast = useToast()

// ── Tabs ──
const activeTab = ref<'quick' | 'custom'>('quick')
const tabs = computed(() => [
  { key: 'quick', label: t('report.tabQuickCharts') },
  { key: 'custom', label: t('report.tabCustomReports') },
])

const chartColors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#84CC16', '#F97316', '#6366F1']

// ═══ REPORT STATE ═══
const reportType = ref('distribution')
const groupBy = ref('state')
const interval = ref('week')
const chartType = ref('Bar')
const dateFrom = ref('')
const dateTo = ref('')

// ═══ FILTER STATE ═══
const filterMode = ref<'basic' | 'rql'>('basic')
const rqlQuery = ref('')
interface FilterCondition { field: string; operator: string; value: string }
const filters = ref<FilterCondition[]>([{ field: '', operator: '=', value: '' }])
const filterOperators = computed(() => [
  { value: '=', label: '等于' },
  { value: '!=', label: '不等于' },
  { value: 'in', label: '属于' },
  { value: 'not_in', label: '不属于' },
  { value: '~', label: '包含' },
  { value: '!~', label: '不包含' },
  { value: 'empty', label: '为空' },
  { value: 'not_empty', label: '不为空' },
])

// ═══ DATA STATE ═══
const data = ref<ReportResponse | null>(null)
const loading = ref(false)
const rqlError = ref<string | null>(null)
const chartCanvas = ref<HTMLCanvasElement | null>(null)
function setChartCanvas(el: any) { chartCanvas.value = el || null }
const { render: renderChart, destroy: destroyChart } = useReportChart(chartCanvas)

// ═══ SAVED FILTERS ═══
const savedFilters = ref<SavedReport[]>([])
const loadingSaved = ref(false)
const selectedFilterId = ref<number | null>(null)
const showSaveFilterDialog = ref(false)
const filterName = ref('')
const savingFilter = ref(false)

// ── Filter Options ──
const states = ref<{value:string;label:string}[]>([])
const priorities = ref<{value:string;label:string}[]>([])
const members = ref<{value:string;label:string}[]>([])
const issueTypes = ref<{value:string;label:string}[]>([])
const labels = ref<{value:string;label:string}[]>([])
const cycles = ref<{value:string;label:string}[]>([])
const modules = ref<{value:string;label:string}[]>([])

const reportTypeLabels = computed(() => ({
  distribution: t('report.types.distribution'),
  created_vs_resolved: t('report.types.created_vs_resolved'),
  avg_age: t('report.types.avg_age'),
  current_age: t('report.types.current_age'),
  created_trend: t('report.types.created_trend'),
}))
const intervalLabels = computed(() => ({
  day: t('report.intervals.day'), week: t('report.intervals.week'), month: t('report.intervals.month'),
}))
const dims = computed(() => [
  { value: 'state', label: t('report.state') },
  { value: 'priority', label: t('report.priority') },
  { value: 'assignee', label: t('report.assignee') },
  { value: 'type', label: t('report.type') },
  { value: 'label', label: t('report.label') },
  { value: 'cycle', label: t('report.cycle') },
  { value: 'module', label: t('report.module') },
])
const filterFields = computed(() => dims.value)
const availableCharts = computed(() => {
  if (reportType.value === 'created_vs_resolved' || reportType.value === 'created_trend') return ['Bar', 'Line', 'Table']
  return ['Bar', 'Pie', 'Doughnut', 'Table']
})
const chartLabels = computed(() => ({
  Bar: t('report.charts.bar'), Pie: t('report.charts.pie'), Doughnut: t('report.charts.doughnut'),
  Line: t('report.charts.line'), Area: t('report.charts.area'), Table: t('report.charts.table'),
}))

function pct(v: number) { return data.value ? Math.round((v / data.value.total) * 100) : 0 }
const builtRqlPreview = computed(() => buildRQLFromFilters())

function buildRQLFromFilters(): string {
  const active = filters.value.filter(f => f.field && f.operator)
  if (active.length === 0) return ''
  return active.map(f => {
    const v = f.value.replace(/"/g, '\\"')
    switch (f.operator) {
      case '=': return `${f.field} = "${v}"`
      case '!=': return `${f.field} != "${v}"`
      case 'in': return `${f.field} IN ["${v}"]`
      case 'not_in': return `${f.field} NOT IN ["${v}"]`
      case '~': return `${f.field} ~ "${v}"`
      case '!~': return `${f.field} !~ "${v}"`
      case 'empty': return `${f.field} IS EMPTY`
      case 'not_empty': return `${f.field} IS NOT EMPTY`
      default: return `${f.field} = "${v}"`
    }
  }).join(' AND ')
}

async function loadFilterOptions() {
  try {
    const [s, p, m, tp, l, c, mod] = await Promise.all([
      api.get(`/projects/${props.projectId}/settings/states`).catch(() => ({ data: [] })),
      Promise.resolve({ data: [
        { value: 'urgent', label: 'Urgent' }, { value: 'high', label: 'High' },
        { value: 'medium', label: 'Medium' }, { value: 'low', label: 'Low' }, { value: 'none', label: 'None' },
      ]}),
      api.get(`/projects/${props.projectId}/members`).catch(() => ({ data: [] })),
      api.get(`/projects/issue-types`).catch(() => ({ data: [] })),
      api.get(`/projects/${props.projectId}/labels`).catch(() => ({ data: [] })),
      api.get(`/projects/${props.projectId}/cycles`).catch(() => ({ data: [] })),
      api.get(`/projects/${props.projectId}/modules`).catch(() => ({ data: [] })),
    ])
    states.value = (s.data || []).map((x: any) => ({ value: x.name, label: x.name || x.color }))
    priorities.value = p.data || []
    members.value = (m.data || []).map((x: any) => ({ value: x.user?.display_name || x.email, label: x.user?.display_name || x.email }))
    issueTypes.value = (tp.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    labels.value = (l.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    cycles.value = (c.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    modules.value = (mod.data || []).map((x: any) => ({ value: x.name, label: x.name }))
  } catch (_) { /* ignore */ }
}

function getFilterOptions(field: string) {
  const map: Record<string, {value:string;label:string}[]> = {
    state: states.value, priority: priorities.value, assignee: members.value,
    type: issueTypes.value, label: labels.value, cycle: cycles.value, module: modules.value,
  }
  return map[field] || []
}

function onFilterFieldChange(idx: number) { filters.value[idx].value = ''; filters.value[idx].operator = '=' }
function addFilter() { filters.value.push({ field: '', operator: '=', value: '' }) }
function removeFilter(idx: number) {
  if (filters.value.length > 1) filters.value.splice(idx, 1)
  else filters.value[0] = { field: '', operator: '=', value: '' }
}

// ═══ GENERATE ═══
async function generate() {
  loading.value = true
  rqlError.value = null
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const res = await reportApi.generate(props.projectId, {
      report_type: reportType.value, group_by: groupBy.value, chart: chartType.value.toLowerCase(),
      rql: rql || undefined, date_from: dateFrom.value || undefined, date_to: dateTo.value || undefined,
      interval: interval.value,
    })
    data.value = res
    loading.value = false
    await nextTick()
    if (chartType.value !== 'Table') {
      await new Promise(r => setTimeout(r, 50))
      renderChart(res, chartType.value)
    }
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || e?.message || 'Unknown error'
    rqlError.value = String(msg)
    data.value = null
    loading.value = false
  }
}

function onTypeChange() {
  if (reportType.value === 'created_trend') interval.value = 'day'
  else if (reportType.value === 'created_vs_resolved') interval.value = 'week'
  if (reportType.value === 'created_vs_resolved' || reportType.value === 'created_trend') chartType.value = 'Bar'
}

watch(chartType, async (newVal) => {
  if (newVal === 'Table') destroyChart()
  else if (data.value) { await nextTick(); await new Promise(r => setTimeout(r, 30)); renderChart(data.value, newVal) }
})
watch([filterMode, rqlQuery], () => { rqlError.value = null })
watch(filters, () => { rqlError.value = null }, { deep: true })

function exportCSV() { if (data.value) exportReportCSV(data.value, `report-${reportType.value}.csv`) }
function exportPNG() { exportChartPNG(chartCanvas.value, `chart-${reportType.value}.png`) }

// ═══ SAVED FILTERS ═══
async function loadSavedFilters() {
  loadingSaved.value = true
  try { savedFilters.value = (await savedReportApi.list(props.projectId)) || [] }
  catch (e) { console.error(e) }
  finally { loadingSaved.value = false }
}

function loadSavedFilter(f: SavedReport) {
  selectedFilterId.value = f.id!
  reportType.value = f.report_type
  groupBy.value = f.group_by
  chartType.value = f.chart_type
  interval.value = f.interval || 'week'
  dateFrom.value = f.date_from || ''
  dateTo.value = f.date_to || ''
  filterName.value = f.name
  if (f.rql) { filterMode.value = 'rql'; rqlQuery.value = f.rql }
  else { filterMode.value = 'basic'; rqlQuery.value = '' }
  generate()
}

async function saveFilter() {
  if (!filterName.value.trim() || savingFilter.value) return
  savingFilter.value = true
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const payload = {
      name: filterName.value, report_type: reportType.value, group_by: groupBy.value,
      chart_type: chartType.value, rql, date_from: dateFrom.value, date_to: dateTo.value, interval: interval.value,
    }
    if (selectedFilterId.value) await savedReportApi.update(props.projectId, selectedFilterId.value, payload)
    else { const created = await savedReportApi.create(props.projectId, payload); selectedFilterId.value = created.id! }
    showSaveFilterDialog.value = false
    await loadSavedFilters()
  } catch (e) { console.error(e); toast.error(t('report.saveFailed')) }
  finally { savingFilter.value = false }
}

async function deleteSavedFilter(f: SavedReport) {
  if (!f.id || !confirm(t('report.deleteFilterConfirm', { name: f.name }))) return
  try {
    await savedReportApi.delete(props.projectId, f.id)
    if (selectedFilterId.value === f.id) { selectedFilterId.value = null; data.value = null; destroyChart() }
    await loadSavedFilters()
  } catch (e) { console.error(e) }
}

// ═══ LIFECYCLE ═══
watch(() => props.projectId, () => {
  destroyChart(); loadSavedFilters(); loadFilterOptions()
  selectedFilterId.value = null; data.value = null
})
onMounted(() => { loadSavedFilters(); loadFilterOptions() })
</script>
