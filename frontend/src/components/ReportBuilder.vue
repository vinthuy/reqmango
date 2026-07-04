<template>
  <div class="flex gap-4 min-h-[600px]">
    <!-- Left: Saved Reports -->
    <div class="w-56 shrink-0 border border-gray-100 rounded-xl bg-white h-fit">
      <div class="px-4 py-3 border-b border-gray-100">
        <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.savedReports') }}</h3>
      </div>
      <div class="divide-y divide-gray-50">
        <button
          v-for="r in savedReports"
          :key="r.id"
          @click="selectSavedReport(r)"
          class="w-full text-left px-4 py-2.5 text-sm hover:bg-gray-50 transition-colors flex items-center justify-between group"
          :class="{ 'bg-blue-50 text-blue-700': selectedId === r.id }"
        >
          <span class="truncate">{{ r.name }}</span>
          <button @click.stop="deleteSavedReport(r)" class="text-gray-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition">&times;</button>
        </button>
        <div v-if="savedReports.length === 0 && !loadingSaved" class="px-4 py-6 text-center text-xs text-gray-400">
          {{ t('report.noSavedReports') }}
        </div>
      </div>
    </div>

    <!-- Right: Report Builder -->
    <div class="flex-1 border border-gray-100 rounded-xl bg-white">
      <!-- Control Bar -->
      <div class="px-5 py-3 border-b border-gray-100 space-y-3">
        <!-- Row 1: Report Type + Group By + Interval + Chart + Date + Actions -->
        <div class="flex flex-wrap items-end gap-3">
          <!-- Report Type -->
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.reportType') }}</label>
            <select v-model="reportType" @change="onTypeChange" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
              <option v-for="(label, key) in reportTypeLabels" :key="key" :value="key">{{ label }}</option>
            </select>
          </div>

          <!-- Group By (only for non-trend) -->
          <div v-if="reportType !== 'created_vs_resolved' && reportType !== 'created_trend'">
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.groupBy') }}</label>
            <select v-model="groupBy" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
              <option v-for="d in dims" :key="d.value" :value="d.value">{{ d.label }}</option>
            </select>
          </div>

          <!-- Interval (for trend) -->
          <div v-if="reportType === 'created_vs_resolved' || reportType === 'created_trend'">
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.interval') }}</label>
            <select v-model="interval" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
              <option v-for="(label, key) in intervalLabels" :key="key" :value="key">{{ label }}</option>
            </select>
          </div>

          <!-- Chart Type -->
          <div>
            <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.chart') }}</label>
            <div class="inline-flex bg-gray-100 rounded-lg p-0.5">
              <button v-for="c in availableCharts" :key="c" @click="chartType = c"
                :class="['px-2.5 py-1 text-xs rounded-md transition-colors', chartType === c ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
              >{{ c }}</button>
            </div>
          </div>

          <!-- Date Range -->
          <div><label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.dateFrom') }}</label><input v-model="dateFrom" type="date" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" /></div>
          <div><label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.dateTo') }}</label><input v-model="dateTo" type="date" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" /></div>

          <button @click="generate" :disabled="loading" class="px-4 py-1.5 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 disabled:opacity-50 transition-colors self-end mb-0.5">
            {{ loading ? '...' : t('report.generate') }}
          </button>

          <button @click="showSaveDialog = true" class="px-3 py-1.5 border border-gray-200 text-gray-600 text-sm rounded-md hover:bg-gray-50 transition-colors self-end mb-0.5">
            {{ t('report.save') }}
          </button>
        </div>

        <!-- Row 2: Filter Mode Toggle + Filter Area -->
        <div>
          <div class="flex items-center gap-3 mb-2">
            <label class="text-[11px] font-medium text-gray-400 uppercase tracking-wide">{{ t('report.filterMode') }}</label>
            <div class="inline-flex bg-gray-100 rounded-lg p-0.5">
              <button @click="filterMode = 'basic'"
                :class="['px-2.5 py-1 text-xs rounded-md transition-colors', filterMode === 'basic' ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
              >{{ t('report.basicFilters') }}</button>
              <button @click="filterMode = 'rql'"
                :class="['px-2.5 py-1 text-xs rounded-md transition-colors', filterMode === 'rql' ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
              >{{ t('report.rqlAdvanced') }}</button>
            </div>
          </div>

          <!-- Basic Filters -->
          <div v-if="filterMode === 'basic'" class="space-y-2">
            <div class="flex flex-wrap items-center gap-2">
              <template v-for="(filter, idx) in filters" :key="idx">
                <select v-model="filter.field" @change="onFilterFieldChange(idx)" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option value="">{{ t('report.selectValue') }}</option>
                  <option v-for="f in filterFields" :key="f.value" :value="f.value">{{ f.label }}</option>
                </select>
                <select v-if="filter.field" v-model="filter.value" class="px-2 py-1.5 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400 min-w-[140px]">
                  <option value="">{{ t('report.selectValue') }}</option>
                  <option v-for="o in getFilterOptions(filter.field)" :key="o.value" :value="o.value">{{ o.label }}</option>
                </select>
                <button @click="removeFilter(idx)" class="text-gray-400 hover:text-red-500 text-lg leading-none">&times;</button>
              </template>
              <button @click="addFilter" class="px-2 py-1.5 text-xs text-gray-500 border border-dashed border-gray-300 rounded-md hover:border-gray-400 hover:text-gray-700 transition-colors">
                {{ t('report.addFilter') }}
              </button>
            </div>
            <div v-if="filters.length === 0" class="text-xs text-gray-400 italic">{{ t('report.noFilters') }}</div>
          </div>

          <!-- RQL Mode -->
          <div v-else class="relative">
            <div class="flex items-start gap-2">
              <div class="relative flex-1">
                <svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
                <input
                  v-model="rqlQuery"
                  type="text"
                  :placeholder="t('report.rqlPlaceholder')"
                  class="w-full pl-9 pr-3 py-1.5 border border-gray-200 rounded-md text-sm font-mono bg-gray-50 focus:bg-white focus:outline-none focus:ring-1 focus:ring-blue-400 transition-colors"
                  @keydown.enter="generate"
                />
              </div>
            </div>
            <p class="text-[11px] text-gray-400 mt-1 ml-1">{{ t('report.rqlHint') }}</p>
          </div>
        </div>
      </div>

      <!-- Status Bar -->
      <div v-if="data" class="flex items-center justify-between px-5 py-2 border-b border-gray-100 text-xs text-gray-500 bg-gray-50/50">
        <div class="flex items-center gap-3">
          <span>{{ t('report.matched') }}: <strong class="text-gray-800">{{ data.total }}</strong> {{ t('report.issues') }}</span>
          <span v-if="data.summary?.avg_days">{{ t('report.avg') }}: <strong class="text-gray-800">{{ data.summary.avg_days.toFixed(1) }}</strong> {{ t('report.days') }}</span>
        </div>
        <div class="flex items-center gap-2">
          <button @click="exportCSV" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded text-xs transition-colors" :title="t('report.exportCSV')">CSV</button>
          <button v-if="chartType !== 'Table'" @click="exportPNG" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded text-xs transition-colors" :title="t('report.exportPNG')">PNG</button>
        </div>
      </div>

      <!-- Chart Area -->
      <div v-if="loading" class="flex items-center justify-center py-20 text-gray-400 text-sm">{{ t('report.loading') }}</div>

      <template v-else-if="data">
        <!-- Chart.js Canvas -->
        <div v-if="chartType !== 'Table'" class="p-5">
          <div :class="['mx-auto', chartType === 'Pie' || chartType === 'Doughnut' ? 'max-w-md' : 'max-w-3xl']" style="height: 360px">
            <canvas ref="chartCanvas"></canvas>
          </div>
        </div>

        <!-- Table -->
        <div v-if="chartType === 'Table'" class="p-5">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100">
                <th class="text-left py-2 text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.groupBy') }}</th>
                <th class="text-right py-2 text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.count') }}</th>
                <th class="text-right py-2 text-xs font-medium text-gray-400 uppercase tracking-wide w-20">{{ t('report.percent') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(label, i) in data.labels" :key="label" class="border-b border-gray-50 hover:bg-gray-50/50">
                <td class="py-2 text-gray-700 flex items-center gap-2">
                  <span class="w-2.5 h-2.5 rounded-full shrink-0" :style="{backgroundColor: chartColors[i % chartColors.length]}"></span>
                  {{ label }}
                </td>
                <td class="text-right py-2 font-medium text-gray-800">{{ data.values[i] }}</td>
                <td class="text-right py-2 text-gray-400">{{ pct(data.values[i]) }}%</td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="border-t-2 border-gray-100">
                <td class="py-2 text-xs font-medium text-gray-500 uppercase">{{ t('report.total') }}</td>
                <td class="text-right py-2 font-semibold text-gray-800">{{ data.total }}</td>
                <td></td>
              </tr>
            </tfoot>
          </table>
        </div>
      </template>

      <!-- Empty State -->
      <div v-else class="flex items-center justify-center py-20 text-gray-400 text-sm">
        {{ t('report.emptyState') }}
      </div>
    </div>

    <!-- Save Dialog -->
    <Teleport to="body">
      <div v-if="showSaveDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/20" @click.self="showSaveDialog = false">
        <div class="bg-white rounded-xl shadow-xl p-6 w-96">
          <h3 class="text-base font-semibold text-gray-800 mb-4">{{ t('report.saveDialog.title') }}</h3>
          <label class="block text-xs font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.saveDialog.name') }}</label>
          <input v-model="saveName" type="text" class="w-full px-3 py-2 border border-gray-200 rounded-md text-sm focus:outline-none focus:ring-1 focus:ring-blue-400" :placeholder="t('report.saveDialog.namePlaceholder')" @keydown.enter="saveReport" />
          <div class="flex justify-end gap-2 mt-4">
            <button @click="showSaveDialog = false" class="px-4 py-1.5 text-sm text-gray-600 border border-gray-200 rounded-md hover:bg-gray-50">{{ t('report.saveDialog.cancel') }}</button>
            <button @click="saveReport" class="px-4 py-1.5 text-sm text-white bg-neutral-900 rounded-md hover:bg-neutral-800 disabled:opacity-50" :disabled="!saveName.trim() || saving">{{ saving ? '...' : t('report.save') }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { reportApi, savedReportApi } from '@/api/report'
import type { ReportResponse, SavedReport } from '@/api/report'
import { useReportChart, exportReportCSV, exportChartPNG } from '@/composables/useReportChart'
import api from '@/api'

const props = defineProps<{ projectId: number }>()
const { t } = useI18n()

// Report config
const reportType = ref('distribution')
const groupBy = ref('state')
const interval = ref('week')
const chartType = ref('Bar')
const dateFrom = ref('')
const dateTo = ref('')

// Filter mode: 'basic' | 'rql'
const filterMode = ref<'basic' | 'rql'>('basic')
const rqlQuery = ref('')

// Basic filter conditions
interface FilterCondition { field: string; value: string }
const filters = ref<FilterCondition[]>([{ field: '', value: '' }])

// Response data
const data = ref<ReportResponse | null>(null)
const loading = ref(false)

// Saved reports
const savedReports = ref<SavedReport[]>([])
const loadingSaved = ref(false)
const selectedId = ref<number | null>(null)
const showSaveDialog = ref(false)
const saveName = ref('')
const saving = ref(false)

// Chart.js
const chartCanvas = ref<HTMLCanvasElement | null>(null)
const { render: renderChart, destroy: destroyChart } = useReportChart(chartCanvas)

// Dropdown data for basic filters
const states = ref<{value:string;label:string}[]>([])
const priorities = ref<{value:string;label:string}[]>([])
const members = ref<{value:string;label:string}[]>([])
const issueTypes = ref<{value:string;label:string}[]>([])
const labels = ref<{value:string;label:string}[]>([])
const cycles = ref<{value:string;label:string}[]>([])
const modules = ref<{value:string;label:string}[]>([])

// i18n labels
const reportTypeLabels = computed(() => ({
  distribution: t('report.types.distribution'),
  created_vs_resolved: t('report.types.created_vs_resolved'),
  avg_age: t('report.types.avg_age'),
  current_age: t('report.types.current_age'),
  created_trend: t('report.types.created_trend'),
}))
const intervalLabels = computed(() => ({
  day: t('report.intervals.day'),
  week: t('report.intervals.week'),
  month: t('report.intervals.month'),
}))

const dims = [
  { value: 'state', label: t('report.state') },
  { value: 'priority', label: t('report.priority') },
  { value: 'assignee', label: t('report.assignee') },
  { value: 'type', label: t('report.type') },
  { value: 'label', label: t('report.label') },
  { value: 'cycle', label: t('report.cycle') },
  { value: 'module', label: t('report.module') },
]

const filterFields = [
  { value: 'state', label: t('report.state') },
  { value: 'priority', label: t('report.priority') },
  { value: 'assignee', label: t('report.assignee') },
  { value: 'type', label: t('report.type') },
  { value: 'label', label: t('report.label') },
  { value: 'cycle', label: t('report.cycle') },
  { value: 'module', label: t('report.module') },
]

const availableCharts = computed(() => {
  if (reportType.value === 'created_vs_resolved' || reportType.value === 'created_trend') return ['Bar', 'Line', 'Table']
  return ['Bar', 'Pie', 'Doughnut', 'Table']
})

const chartColors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#84CC16', '#F97316', '#6366F1']

function pct(v: number) { return data.value ? Math.round((v / data.value.total) * 100) : 0 }

// Build RQL string from basic filters
function buildRQLFromFilters(): string {
  const active = filters.value.filter(f => f.field && f.value)
  if (active.length === 0) return ''
  return active.map(f => {
    const val = f.value.includes(' ') ? `"${f.value}"` : f.value
    return `${f.field} = ${val}`
  }).join(' AND ')
}

// Load dropdown options
async function loadFilterOptions() {
  try {
    const [s, p, m, t, l, c, mod] = await Promise.all([
      api.get(`/projects/${props.projectId}/states`),
      Promise.resolve({ data: [
        { value: 'urgent', label: 'Urgent' },
        { value: 'high', label: 'High' },
        { value: 'medium', label: 'Medium' },
        { value: 'low', label: 'Low' },
        { value: 'none', label: 'None' },
      ]}),
      api.get(`/projects/${props.projectId}/members`),
      api.get(`/projects/issue-types`).catch(() => ({ data: [] })),
      api.get(`/projects/${props.projectId}/labels`),
      api.get(`/projects/${props.projectId}/cycles`),
      api.get(`/projects/${props.projectId}/modules`),
    ])
    states.value = (s.data || []).map((x: any) => ({ value: x.name, label: x.name || x.color }))
    priorities.value = p.data || []
    members.value = (m.data || []).map((x: any) => ({ value: x.user?.display_name || x.email, label: x.user?.display_name || x.email }))
    issueTypes.value = (t.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    labels.value = (l.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    cycles.value = (c.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    modules.value = (mod.data || []).map((x: any) => ({ value: x.name, label: x.name }))
  } catch (_) { /* ignore */ }
}

function getFilterOptions(field: string) {
  const map: Record<string, {value:string;label:string}[]> = {
    state: states.value,
    priority: priorities.value,
    assignee: members.value,
    type: issueTypes.value,
    label: labels.value,
    cycle: cycles.value,
    module: modules.value,
  }
  return map[field] || []
}

function onFilterFieldChange(idx: number) {
  filters.value[idx].value = ''
}

function addFilter() {
  filters.value.push({ field: '', value: '' })
}

function removeFilter(idx: number) {
  if (filters.value.length > 1) {
    filters.value.splice(idx, 1)
  } else {
    filters.value[0] = { field: '', value: '' }
  }
}

// Report generation
async function generate() {
  loading.value = true
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const res = await reportApi.generate(props.projectId, {
      report_type: reportType.value,
      group_by: groupBy.value,
      rql: rql || undefined,
      date_from: dateFrom.value || undefined,
      date_to: dateTo.value || undefined,
      interval: interval.value,
    })
    data.value = res
    await nextTick()
    if (chartType.value !== 'Table') {
      renderChart(res, chartType.value)
    }
  } catch (e) {
    console.error('Failed to generate report:', e)
    data.value = null
  } finally {
    loading.value = false
  }
}

function onTypeChange() {
  if (reportType.value === 'created_trend') interval.value = 'day'
  else if (reportType.value === 'created_vs_resolved') interval.value = 'week'
  if (reportType.value === 'created_vs_resolved' || reportType.value === 'created_trend') chartType.value = 'Bar'
}

// Re-render chart when chart type changes
watch(chartType, async () => {
  if (data.value && chartType.value !== 'Table') {
    await nextTick()
    renderChart(data.value, chartType.value)
  }
})

// Export
function exportCSV() {
  if (!data.value) return
  exportReportCSV(data.value, `report-${reportType.value}-${new Date().toISOString().slice(0, 10)}.csv`)
}

function exportPNG() {
  exportChartPNG(chartCanvas.value, `chart-${reportType.value}-${new Date().toISOString().slice(0, 10)}.png`)
}

// Saved Reports
async function loadSavedReports() {
  loadingSaved.value = true
  try {
    savedReports.value = await savedReportApi.list(props.projectId)
  } catch (e) {
    console.error('Failed to load saved reports:', e)
  } finally {
    loadingSaved.value = false
  }
}

function selectSavedReport(r: SavedReport) {
  selectedId.value = r.id!
  reportType.value = r.report_type
  groupBy.value = r.group_by
  chartType.value = r.chart_type
  interval.value = r.interval || 'week'
  dateFrom.value = r.date_from || ''
  dateTo.value = r.date_to || ''
  saveName.value = r.name

  if (r.rql) {
    filterMode.value = 'rql'
    rqlQuery.value = r.rql
  } else {
    filterMode.value = 'basic'
    rqlQuery.value = ''
  }
  generate()
}

async function saveReport() {
  if (!saveName.value.trim() || saving.value) return
  saving.value = true
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const payload = {
      name: saveName.value,
      report_type: reportType.value,
      group_by: groupBy.value,
      chart_type: chartType.value,
      rql: rql,
      date_from: dateFrom.value,
      date_to: dateTo.value,
      interval: interval.value,
    }

    if (selectedId.value) {
      await savedReportApi.update(props.projectId, selectedId.value, payload)
    } else {
      const created = await savedReportApi.create(props.projectId, payload)
      selectedId.value = created.id!
    }
    showSaveDialog.value = false
    await loadSavedReports()
  } catch (e) {
    console.error('Failed to save report:', e)
    alert(t('report.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function deleteSavedReport(r: SavedReport) {
  if (!r.id || !confirm(t('report.deleteConfirm').replace('{name}', r.name))) return
  try {
    await savedReportApi.delete(props.projectId, r.id)
    if (selectedId.value === r.id) {
      selectedId.value = null
      destroyChart()
    }
    await loadSavedReports()
  } catch (e) {
    console.error('Failed to delete report:', e)
  }
}

watch(() => props.projectId, () => {
  destroyChart()
  loadSavedReports()
  loadFilterOptions()
  selectedId.value = null
  data.value = null
})

onMounted(() => {
  loadSavedReports()
  loadFilterOptions()
  generate()
})
</script>
