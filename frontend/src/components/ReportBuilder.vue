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
      <!-- KPI Number Widgets -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <div class="p-4 bg-white rounded-xl border border-gray-100">
          <div class="text-xs text-gray-400 mb-1">{{ t('report.quick.totalIssues') }}</div>
          <div class="text-2xl font-bold text-gray-800">{{ quickStats.total }}</div>
        </div>
        <div class="p-4 bg-white rounded-xl border border-gray-100">
          <div class="text-xs text-gray-400 mb-1">{{ t('report.quick.avgAge') }}</div>
          <div class="text-2xl font-bold text-gray-800">{{ quickStats.avgAge }} <span class="text-sm font-normal text-gray-400">{{ t('report.days') }}</span></div>
        </div>
        <div class="p-4 bg-white rounded-xl border border-gray-100">
          <div class="text-xs text-gray-400 mb-1">{{ t('report.quick.stateGroups') }}</div>
          <div class="text-2xl font-bold text-gray-800">{{ quickStats.stateGroups }}</div>
        </div>
        <div class="p-4 bg-white rounded-xl border border-gray-100">
          <div class="text-xs text-gray-400 mb-1">{{ t('report.quick.completionRate') }}</div>
          <div class="text-2xl font-bold text-gray-800">{{ quickStats.completionRate }}<span class="text-sm font-normal text-gray-400">%</span></div>
        </div>
      </div>

      <!-- Chart Widgets -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div v-for="q in quickCharts" :key="q.title"
          @click="runQuickChart(q)"
          class="bg-white border border-gray-100 rounded-xl overflow-hidden cursor-pointer hover:shadow-md transition-shadow"
        >
          <!-- Widget Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b border-gray-50">
            <h4 class="text-sm font-medium text-gray-700">{{ q.title }}</h4>
            <div class="flex items-center gap-1" @click.stop>
              <button v-for="c in q.charts" :key="c" @click="setQuickChartStyle(q, c)"
                :class="['px-2 py-0.5 text-[10px] rounded transition-colors', q.style === c ? 'bg-gray-100 text-gray-800 font-medium' : 'text-gray-400 hover:text-gray-600']"
              >{{ chartLabel(c) }}</button>
            </div>
          </div>
          <!-- Widget Filter -->
          <div class="px-4 py-2 flex items-center gap-2" @click.stop>
            <select v-if="q.filterOptions.length > 0" :value="q.filters.value" @change="setQuickFilter(q, ($event.target as HTMLSelectElement).value)"
              class="text-xs px-2 py-1 border border-gray-200 rounded-md bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
              <option value="">{{ t('report.all') }}</option>
              <option v-for="o in q.filterOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
            </select>
            <input v-model="q.dateFrom" type="date" class="px-2 py-1 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" />
            <input v-model="q.dateTo" type="date" class="px-2 py-1 border border-gray-200 rounded-md text-xs bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" />
          </div>
          <!-- Widget Body -->
          <div class="px-4 py-4">
            <div v-if="q.loading" class="flex items-center justify-center h-48 text-xs text-gray-400">
              {{ t('report.loading') }}
            </div>
            <div v-else-if="q.data" class="h-48 relative">
              <canvas :ref="(el: any) => setQuickCanvas(q.title, el)"></canvas>
            </div>
            <div v-else class="flex flex-col items-center justify-center h-48 text-xs text-gray-400">
              <svg class="w-8 h-8 mb-2 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
              <span>{{ t('report.clickToGenerate') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Export -->
      <div v-if="quickData" class="flex items-center justify-end gap-2 px-5 py-2 border border-gray-100 rounded-xl bg-white text-xs text-gray-500">
        <button @click="exportCSV" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors">CSV</button>
        <button @click="exportPNG" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors">PNG</button>
      </div>
    </template>

    <!-- ═══ CUSTOM REPORTS TAB ═══ -->
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

        <!-- Right: Step-by-step Report Builder -->
        <div class="flex-1 space-y-4">

          <!-- ── Step 1: Filter Data ── -->
          <div class="bg-white border border-gray-100 rounded-xl px-5 py-4">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span class="flex items-center justify-center w-5 h-5 rounded-full bg-indigo-600 text-white text-[11px] font-bold">1</span>
                <h4 class="text-sm font-medium text-gray-700">{{ t('report.v2.step1') }}</h4>
              </div>
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
                <select v-model="filter.field" @change="onFilterFieldChange(idx)" class="w-40 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option value="">{{ t('report.selectField') }}</option>
                  <option v-for="f in filterFields" :key="f.value" :value="f.value">{{ f.label }}</option>
                </select>
                <select v-if="filter.field" v-model="filter.operator" class="w-28 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                  <option v-for="op in filterOperators" :key="op.value" :value="op.value">{{ op.label }}</option>
                </select>
                <template v-if="filter.field && filter.operator !== 'empty' && filter.operator !== 'not_empty'">
                  <select v-if="getFilterOptions(filter.field).length > 0" v-model="filter.value" class="w-44 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <option value="">{{ t('report.selectValue') }}</option>
                    <option v-for="o in getFilterOptions(filter.field)" :key="o.value" :value="o.value">{{ o.label }}</option>
                  </select>
                  <input v-else v-model="filter.value" :type="['start_date','target_date','created_at','updated_at'].includes(filter.field) ? 'date' : 'text'" :placeholder="t('report.selectValue')" class="w-44 px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400" />
                </template>
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

            <!-- Apply -->
            <div class="mt-3 flex items-center gap-3">
              <button @click="applyFilter" :disabled="filterLoading" class="px-4 py-1.5 bg-indigo-600 text-white text-sm rounded-md hover:bg-indigo-700 disabled:opacity-50 transition-colors">
                {{ filterLoading ? '...' : t('report.v2.applyFilter') }}
              </button>
              <span v-if="filterApplied" class="text-sm text-green-600 font-medium">
                {{ t('report.v2.filterApplied', { count: matchCount }) }}
              </span>
            </div>
          </div>

          <!-- ── Step 2: Configure Chart ── -->
          <div :class="['bg-white border rounded-xl px-5 py-4 transition-colors', filterApplied ? 'border-gray-100' : 'border-gray-100 opacity-60']">
            <div class="flex items-center gap-2 mb-4">
              <span :class="['flex items-center justify-center w-5 h-5 rounded-full text-[11px] font-bold', filterApplied ? 'bg-indigo-600 text-white' : 'bg-gray-200 text-gray-500']">2</span>
              <h4 class="text-sm font-medium text-gray-700">{{ t('report.v2.step2') }}</h4>
            </div>

            <template v-if="filterApplied">
              <!-- X/Y Axis Selectors -->
              <div class="flex flex-wrap items-end gap-4 mb-4">
                <!-- X-Axis -->
                <div class="flex-1 min-w-[200px]">
                  <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.v2.xAxis') }}</label>
                  <select v-model="xAxis" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <optgroup :label="t('report.v2.category')">
                      <option v-for="d in xAxisCategoryOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup :label="t('report.v2.timeCreated')">
                      <option v-for="d in xAxisTimeCreatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup :label="t('report.v2.timeCompleted')">
                      <option v-for="d in xAxisTimeCompletedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                    <optgroup :label="t('report.v2.timeUpdated')">
                      <option v-for="d in xAxisTimeUpdatedOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
                    </optgroup>
                  </select>
                </div>

                <!-- Y-Axis -->
                <div class="flex-1 min-w-[200px]">
                  <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.v2.yAxis') }}</label>
                  <select v-model="yAxis" class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <option v-for="m in yAxisOptions" :key="m.value" :value="m.value">{{ m.label }}</option>
                  </select>
                </div>
              </div>

              <!-- Chart Type + Actions -->
              <div class="flex flex-wrap items-end gap-3 mb-4">
                <div>
                  <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.chart') }}</label>
                  <div class="inline-flex bg-gray-100 rounded-lg p-0.5">
                    <button v-for="c in availableCharts" :key="c" @click="chartType = c"
                      :class="['px-2.5 py-1 text-xs rounded-md transition-colors', chartType === c ? 'bg-white shadow-sm font-medium text-gray-800' : 'text-gray-500 hover:text-gray-700']"
                    >{{ (chartLabels as Record<string, string>)[c] || c }}</button>
                  </div>
                </div>
                <button @click="generateV2" :disabled="chartLoading" class="px-4 py-1.5 bg-neutral-900 text-white text-sm rounded-md hover:bg-neutral-800 disabled:opacity-50 transition-colors self-end mb-0.5">
                  {{ chartLoading ? '...' : t('report.generate') }}
                </button>
                <button @click="showSaveFilterDialog = true" class="px-3 py-1.5 border border-gray-200 text-gray-600 text-sm rounded-md hover:bg-gray-50 transition-colors self-end mb-0.5">
                  {{ t('report.saveFilter') }}
                </button>
              </div>

              <!-- Results -->
              <div v-if="data" class="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
                <div class="p-4 bg-gray-50 rounded-xl border border-gray-100">
                  <div class="text-xs text-gray-400 mb-1">{{ t('report.quick.totalIssues') }}</div>
                  <div class="text-2xl font-bold text-gray-800">{{ data.total }}</div>
                </div>
                <div v-if="data.summary?.avg_days" class="p-4 bg-gray-50 rounded-xl border border-gray-100">
                  <div class="text-xs text-gray-400 mb-1">{{ t('report.avg') }}</div>
                  <div class="text-2xl font-bold text-gray-800">{{ data.summary.avg_days.toFixed(1) }} <span class="text-sm font-normal text-gray-400">{{ t('report.days') }}</span></div>
                </div>
              </div>
              <div v-if="data" class="flex items-center justify-between px-4 py-2 bg-gray-50 rounded-lg text-xs text-gray-500 mb-4">
                <span>{{ t('report.matched') }}: <strong class="text-gray-800">{{ data.total }}</strong> {{ t('report.issues') }}</span>
                <div class="flex items-center gap-2">
                  <button @click="exportCSV" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors">CSV</button>
                  <button v-if="chartType !== 'Table'" @click="exportPNG" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded transition-colors">PNG</button>
                </div>
              </div>

              <!-- Chart Area -->
              <div v-if="chartLoading" class="flex items-center justify-center py-16 text-sm text-gray-400">
                {{ t('report.loading') }}
              </div>
              <template v-else-if="data">
                <div v-show="chartType !== 'Table'" class="bg-gray-50 border border-gray-100 rounded-xl p-5">
                  <div :class="['mx-auto', (chartType === 'Pie' || chartType === 'Doughnut') ? 'max-w-md' : (chartType === 'Radar' || chartType === 'PolarArea') ? 'max-w-lg' : 'max-w-4xl']" style="height: 360px">
                    <canvas :ref="setChartCanvas"></canvas>
                  </div>
                </div>
                <div v-show="chartType === 'Table'" class="bg-gray-50 border border-gray-100 rounded-xl p-5">
                  <table class="w-full text-sm">
                    <thead><tr class="border-b border-gray-200">
                      <th class="text-left py-2 text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.groupBy') }}</th>
                      <th class="text-right py-2 text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.count') }}</th>
                      <th class="text-right py-2 text-xs font-medium text-gray-400 uppercase tracking-wide w-20">{{ t('report.percent') }}</th>
                    </tr></thead>
                    <tbody>
                      <tr v-for="(label, i) in data.labels" :key="i" class="border-b border-gray-50 hover:bg-gray-100/50">
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
            </template>

            <!-- Empty State -->
            <div v-else class="flex flex-col items-center justify-center py-12 text-gray-400">
              <svg class="w-10 h-10 mb-2 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
              <span class="text-sm">{{ t('report.v2.noData') }}</span>
            </div>
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

// ═══ V2 STATE ═══
const xAxis = ref('state')
const yAxis = ref('count')
const filterApplied = ref(false)
const matchCount = ref(0)
const filterLoading = ref(false)
const chartLoading = ref(false)

// ═══ QUICK CHART STATE ═══
const quickData = ref<ReportResponse | null>(null)
const quickLoading = ref(false)
const quickChartType = ref('Bar')
const quickCanvasMap = new Map<string, HTMLCanvasElement | null>()
const quickChartInstances = new Map<string, any>()
const quickStats = ref({ total: 0, avgAge: 0, stateGroups: 0, completionRate: 0 })
const quickCharts = ref([
  { title: t('report.quick.byState'), reportType: 'distribution', groupBy: 'state', style: 'Bar', charts: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], data: null as ReportResponse | null, loading: false, filters: { value: '' }, filterOptions: [] as {value:string;label:string}[], dateFrom: '', dateTo: '' },
  { title: t('report.quick.byPriority'), reportType: 'distribution', groupBy: 'priority', style: 'Pie', charts: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], data: null as ReportResponse | null, loading: false, filters: { value: '' }, filterOptions: [] as {value:string;label:string}[], dateFrom: '', dateTo: '' },
  { title: t('report.quick.byAssignee'), reportType: 'distribution', groupBy: 'assignee', style: 'Bar', charts: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], data: null as ReportResponse | null, loading: false, filters: { value: '' }, filterOptions: [] as {value:string;label:string}[], dateFrom: '', dateTo: '' },
  { title: t('report.quick.byType'), reportType: 'distribution', groupBy: 'type', style: 'Bar', charts: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], data: null as ReportResponse | null, loading: false, filters: { value: '' }, filterOptions: [] as {value:string;label:string}[], dateFrom: '', dateTo: '' },
  { title: t('report.quick.byTrend'), reportType: 'created_trend', groupBy: 'state', style: 'Area', charts: ['Area', 'Line', 'Bar', 'Table'] as string[], data: null as ReportResponse | null, loading: false, filters: { value: '' }, filterOptions: [] as {value:string;label:string}[], dateFrom: '', dateTo: '' },
])
function setQuickCanvas(key: string, el: any) { if (el) quickCanvasMap.set(key, el) }
function chartLabel(c: string) { return (chartLabels.value as Record<string,string>)[c] || c }
function setQuickChartStyle(q: any, c: string) { q.style = c }
function setQuickFilter(q: any, v: string) { q.filters.value = v; runQuickChart(q) }

async function runQuickChart(q: any) {
  q.loading = true
  quickLoading.value = true
  try {
    let rql = ''
    if (q.filters?.value) {
      rql = `${q.groupBy} IN ["${q.filters.value}"]`
    }
    const res = await reportApi.generate(props.projectId, {
      report_type: q.reportType, group_by: q.groupBy, chart: q.style.toLowerCase(),
      rql: rql || undefined, interval: q.reportType === 'created_trend' ? 'day' : undefined,
      date_from: q.dateFrom || undefined, date_to: q.dateTo || undefined,
    })
    q.data = res
    quickData.value = res
    quickChartType.value = q.style

    // Update KPI stats from distribution results
    if (q.reportType === 'distribution' && q.groupBy === 'state') {
      quickStats.value.total = res.total
      quickStats.value.stateGroups = res.labels?.length || 0
    }

    q.loading = false
    quickLoading.value = false
    await nextTick()
    const canvas = quickCanvasMap.get(q.title)
    if (canvas && q.style !== 'Table') {
      // Destroy old chart instance if exists
      const oldChart = quickChartInstances.get(q.title)
      if (oldChart) oldChart.destroy()
      await new Promise(r => setTimeout(r, 50))
      const { Chart: ChartJS } = await import('chart.js')
      const chartInstance = new ChartJS(canvas, {
        type: q.style === 'Area' ? 'line' : (q.style?.toLowerCase() as any),
        data: { labels: res.labels, datasets: [{ data: res.values, backgroundColor: ['#3B82F6','#10B981','#F59E0B','#EF4444','#8B5CF6','#EC4899','#06B6D4','#84CC16','#F97316','#6366F1'] }] },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { display: ['Pie','Doughnut'].includes(q.style) } }, scales: q.style !== 'Pie' && q.style !== 'Doughnut' ? { y: { beginAtZero: true } } : undefined },
      })
      quickChartInstances.set(q.title, chartInstance)
    }
  } catch (e) { console.error(e); q.loading = false; quickLoading.value = false }
}

// Load filter options for quick chart cards
async function loadQuickChartFilters() {
  try {
    const sRes = await api.get(`/projects/${props.projectId}/settings/states`).catch(() => ({ data: [] }))
    const statesList = (sRes.data || []).map((x: any) => ({ value: x.name, label: x.name || x.color }))
    const pRes = await api.get(`/projects/${props.projectId}/settings/priorities`).catch(() => ({ data: [] }))
    const prioritiesList = (pRes.data || []).map((x: any) => ({ value: x.name, label: x.name || x }))
    const tRes = await api.get(`/projects/issue-types`).catch(() => ({ data: [] }))
    const typesList = (tRes.data || []).map((x: any) => ({ value: x.name, label: x.name }))
    quickCharts.value.forEach((q: any) => {
      if (q.groupBy === 'state') q.filterOptions = statesList
      else if (q.groupBy === 'priority') q.filterOptions = prioritiesList
      else if (q.groupBy === 'type') q.filterOptions = typesList
    })
  } catch (_) { /* ignore */ }
}

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
  { value: '>=', label: '晚于' },
  { value: '<=', label: '早于' },
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

const dims = computed(() => [
  { value: 'state', label: t('report.state') },
  { value: 'priority', label: t('report.priority') },
  { value: 'assignee', label: t('report.assignee') },
  { value: 'type', label: t('report.type') },
  { value: 'label', label: t('report.label') },
  { value: 'cycle', label: t('report.cycle') },
  { value: 'module', label: t('report.module') },
])
const filterFields = computed(() => [
  ...dims.value,
  { value: 'name', label: '标题' },
  { value: 'description', label: '描述' },
  { value: 'sequence_id', label: '编号' },
  { value: 'start_date', label: '开始日期' },
  { value: 'target_date', label: '目标日期' },
  { value: 'created_at', label: '创建时间' },
  { value: 'updated_at', label: '更新时间' },
])
const availableCharts = computed(() => {
  return ['Bar', 'Pie', 'Doughnut', 'Line', 'Area', 'Radar', 'PolarArea', 'HorizontalBar', 'StackedBar', 'Bubble', 'Scatter', 'Mixed', 'Table']
})
const chartLabels = computed(() => ({
  Bar: t('report.charts.bar'), Pie: t('report.charts.pie'), Doughnut: t('report.charts.doughnut'),
  Line: t('report.charts.line'), Area: t('report.charts.area'), Table: t('report.charts.table'),
  Radar: t('report.charts.radar'), PolarArea: t('report.charts.polarArea'),
  HorizontalBar: t('report.charts.horizontalBar'), StackedBar: t('report.charts.stackedBar'),
  Bubble: t('report.charts.bubble'), Scatter: t('report.charts.scatter'), Mixed: t('report.charts.mixed'),
}))

// ═══ V2 X/Y AXIS OPTIONS ═══
const xAxisCategoryOptions = computed(() => [
  { value: 'state', label: t('report.state') },
  { value: 'priority', label: t('report.priority') },
  { value: 'assignee', label: t('report.assignee') },
  { value: 'type', label: t('report.type') },
  { value: 'label', label: t('report.label') },
  { value: 'cycle', label: t('report.cycle') },
  { value: 'module', label: t('report.module') },
])
const xAxisTimeCreatedOptions = computed(() => [
  { value: 'created_day', label: `${t('report.v2.timeCreated')} - ${t('report.v2.timeGranularity')}: Day` },
  { value: 'created_week', label: `${t('report.v2.timeCreated')} - ${t('report.v2.timeGranularity')}: Week` },
  { value: 'created_month', label: `${t('report.v2.timeCreated')} - ${t('report.v2.timeGranularity')}: Month` },
])
const xAxisTimeCompletedOptions = computed(() => [
  { value: 'completed_day', label: `${t('report.v2.timeCompleted')} - ${t('report.v2.timeGranularity')}: Day` },
  { value: 'completed_week', label: `${t('report.v2.timeCompleted')} - ${t('report.v2.timeGranularity')}: Week` },
  { value: 'completed_month', label: `${t('report.v2.timeCompleted')} - ${t('report.v2.timeGranularity')}: Month` },
])
const xAxisTimeUpdatedOptions = computed(() => [
  { value: 'updated_day', label: `${t('report.v2.timeUpdated')} - ${t('report.v2.timeGranularity')}: Day` },
  { value: 'updated_week', label: `${t('report.v2.timeUpdated')} - ${t('report.v2.timeGranularity')}: Week` },
  { value: 'updated_month', label: `${t('report.v2.timeUpdated')} - ${t('report.v2.timeGranularity')}: Month` },
])
const yAxisOptions = computed(() => [
  { value: 'count', label: t('report.v2.count') },
  { value: 'avg_processing_time', label: t('report.v2.avgProcessingTime') },
  { value: 'current_retention', label: t('report.v2.currentRetention') },
  { value: 'created_vs_resolved', label: t('report.v2.createdVsResolved') },
])

function pct(v: number) { return data.value ? Math.round((v / data.value.total) * 100) : 0 }
const builtRqlPreview = computed(() => buildRQLFromFilters())

function buildRQLFromFilters(): string {
  const noValueOps = ['empty', 'not_empty']
  const active = filters.value.filter(f => f.field && f.operator && (noValueOps.includes(f.operator) || f.value.trim() !== ''))
  if (active.length === 0) return ''
  return active.map(f => {
    const v = f.value.replace(/"/g, '\\"')
    switch (f.operator) {
      case '=': return `${f.field} = "${v}"`
      case '!=': return `${f.field} != "${v}"`
      case 'in': return `${f.field} IN ("${v}")`
      case 'not_in': return `${f.field} NOT IN ("${v}")`
      case '~': return `${f.field} LIKE "${v}"`
      case '!~': return `${f.field} NOT LIKE "${v}"`
      case '>=': return `${f.field} >= "${v}"`
      case '<=': return `${f.field} <= "${v}"`
      case 'empty': return `${f.field} IS NULL`
      case 'not_empty': return `${f.field} IS NOT NULL`
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

// ═══ V2: APPLY FILTER (Step 1) ═══
async function applyFilter() {
  filterLoading.value = true
  rqlError.value = null
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    // Use count query to validate filter and get match count
    const res = await reportApi.generateV2(props.projectId, {
      x_axis: 'state', y_axis: 'count',
      rql: rql || undefined, date_from: dateFrom.value || undefined, date_to: dateTo.value || undefined,
    })
    matchCount.value = res.total
    filterApplied.value = true
    data.value = null
    destroyChart()
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || e?.message || 'Unknown error'
    rqlError.value = String(msg)
    filterApplied.value = false
  } finally {
    filterLoading.value = false
  }
}

// ═══ V2: GENERATE CHART (Step 2) ═══
async function generateV2() {
  chartLoading.value = true
  rqlError.value = null
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const res = await reportApi.generateV2(props.projectId, {
      x_axis: xAxis.value, y_axis: yAxis.value,
      rql: rql || undefined, date_from: dateFrom.value || undefined, date_to: dateTo.value || undefined,
    })
    data.value = res
    matchCount.value = res.total
    chartLoading.value = false
    await nextTick()
    if (chartType.value !== 'Table') {
      await new Promise(r => setTimeout(r, 50))
      renderChart(res, chartType.value)
    }
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || e?.message || 'Unknown error'
    rqlError.value = String(msg)
    data.value = null
    chartLoading.value = false
  }
}

// ═══ GENERATE (legacy, for quick charts) ═══
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


watch(chartType, async (newVal) => {
  if (newVal === 'Table') destroyChart()
  else {
    const chartData = activeTab.value === 'quick' ? quickData.value : data.value
    if (chartData) { await nextTick(); await new Promise(r => setTimeout(r, 30)); renderChart(chartData, newVal) }
  }
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
  // V2: restore xAxis/yAxis from saved report
  if ((f as any).x_axis) xAxis.value = (f as any).x_axis
  if ((f as any).y_axis) yAxis.value = (f as any).y_axis
  if (f.rql) { filterMode.value = 'rql'; rqlQuery.value = f.rql }
  else { filterMode.value = 'basic'; rqlQuery.value = '' }
  applyFilter()
}

async function saveFilter() {
  if (!filterName.value.trim() || savingFilter.value) return
  savingFilter.value = true
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const payload = {
      name: filterName.value, report_type: reportType.value, group_by: groupBy.value,
      chart_type: chartType.value, rql, date_from: dateFrom.value, date_to: dateTo.value, interval: interval.value,
      x_axis: xAxis.value, y_axis: yAxis.value,
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
onMounted(() => { loadSavedFilters(); loadFilterOptions(); loadQuickChartFilters() })
</script>
