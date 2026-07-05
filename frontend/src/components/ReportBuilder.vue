<template>
  <div class="space-y-4">
    <!-- Tab Header -->
    <div class="flex items-center gap-1 border-b border-gray-200">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="onTabClick(tab.key)"
        :class="[
          'px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px',
          activeTab === tab.key
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
        ]"
      >{{ tab.label }}</button>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- TAB 1: Quick Charts (快速图表) -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'quick'" class="space-y-4">
      <!-- Number Widgets -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <div v-for="nw in numberWidgets" :key="nw.label" class="p-4 bg-white rounded-xl border border-gray-100">
          <div class="text-xs text-gray-400 mb-1">{{ nw.label }}</div>
          <div class="text-2xl font-bold text-gray-800">{{ nw.value }}</div>
          <div v-if="nw.sub" class="text-xs text-gray-400 mt-1">{{ nw.sub }}</div>
        </div>
      </div>

      <!-- Quick chart cards -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div v-for="q in quickCharts" :key="q.type + q.groupBy"
          @click="runQuickChart(q)"
          class="bg-white rounded-xl border border-gray-100 overflow-hidden cursor-pointer hover:shadow-md transition-shadow"
          :class="{ 'ring-2 ring-indigo-200 border-indigo-200': quickActive?.type === q.type && quickActive?.groupBy === q.groupBy }"
        >
          <!-- Card Header -->
          <div class="flex items-center justify-between px-4 py-3 border-b border-gray-50">
            <div class="flex items-center gap-2">
              <span class="w-7 h-7 rounded-lg flex items-center justify-center text-xs" :style="{ backgroundColor: q.color + '18', color: q.color }">
                <component :is="q.icon" />
              </span>
              <div>
                <span class="text-sm font-medium text-gray-800">{{ q.label }}</span>
                <p class="text-[11px] text-gray-400">{{ q.desc }}</p>
              </div>
            </div>
            <!-- Chart type selector -->
            <div @click.stop class="flex items-center bg-gray-100 rounded-lg p-0.5">
              <button v-for="ct in q.chartTypes" :key="ct" @click="switchQuickChartType(q, ct)"
                :class="['px-2 py-0.5 text-[11px] rounded-md transition-colors', getQuickChartType(q) === ct ? 'bg-white shadow-sm font-medium text-gray-700' : 'text-gray-400 hover:text-gray-600']"
                :title="(chartLabels as Record<string, string>)[ct] || ct"
              >{{ (chartLabels as Record<string, string>)[ct] || ct }}</button>
            </div>
          </div>

          <!-- Inline Filters -->
          <div @click.stop v-if="q.filters && q.filters.length > 0" class="flex items-center gap-2 px-4 py-2 bg-gray-50/50 border-b border-gray-50">
            <svg class="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"/></svg>
            <template v-for="(f, fi) in q.filters" :key="fi">
              <select :value="getQuickFilterVal(q, fi)" @change="onQuickFilterChange(q, fi, ($event.target as HTMLSelectElement).value)"
                class="px-2 py-0.5 border border-gray-200 rounded text-[11px] bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                <option value="">{{ f.label }}</option>
                <option v-for="opt in f.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </template>
            <button v-if="hasQuickFilters(q)" @click="clearQuickFilters(q)" class="text-[11px] text-gray-400 hover:text-red-500 ml-1">清除</button>
          </div>

          <!-- Chart Area -->
          <div class="p-4">
            <div v-if="quickLoading && quickActive?.type === q.type && quickActive?.groupBy === q.groupBy"
              class="flex items-center justify-center py-12 text-gray-400 text-sm">
              {{ t('report.loading') }}
            </div>
            <div v-else-if="quickActive?.type === q.type && quickActive?.groupBy === q.groupBy && quickData" class="space-y-3">
              <!-- Summary Stats -->
              <div class="flex items-center gap-4 text-xs text-gray-500">
                <span>{{ t('report.matched') }}: <strong class="text-gray-800">{{ quickData.total }}</strong></span>
                <span v-if="quickData.summary?.avg_days">{{ t('report.avg') }}: <strong class="text-gray-800">{{ quickData.summary.avg_days.toFixed(1) }}</strong> {{ t('report.days') }}</span>
              </div>
              <!-- Canvas -->
              <div :class="['mx-auto', quickChartType === 'Pie' || quickChartType === 'Doughnut' ? 'max-w-xs' : 'w-full']" style="height: 240px">
                <canvas :ref="setQuickCanvas"></canvas>
              </div>
              <!-- Data Table -->
              <table class="w-full text-xs">
                <thead>
                  <tr class="border-b border-gray-100">
                    <th class="text-left py-1.5 text-gray-400 font-medium">{{ t('report.groupBy') }}</th>
                    <th class="text-right py-1.5 text-gray-400 font-medium">{{ t('report.count') }}</th>
                    <th class="text-right py-1.5 text-gray-400 font-medium w-16">{{ t('report.percent') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(label, i) in quickData.labels" :key="label" class="border-b border-gray-50">
                    <td class="py-1.5 text-gray-700 flex items-center gap-1.5">
                      <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: chartColors[i % chartColors.length] }"></span>
                      {{ label }}
                    </td>
                    <td class="text-right py-1.5 font-medium text-gray-800">{{ quickData.values[i] }}</td>
                    <td class="text-right py-1.5 text-gray-400">{{ qPct(quickData.values[i]) }}%</td>
                  </tr>
                </tbody>
              </table>
              <!-- Export -->
              <div class="flex justify-end gap-2 pt-1">
                <button @click="exportQuickCSV" class="px-2.5 py-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded text-[11px] transition-colors">CSV</button>
                <button @click="exportQuickPNG" class="px-2.5 py-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded text-[11px] transition-colors">PNG</button>
              </div>
            </div>
            <div v-else class="flex items-center justify-center py-12 text-gray-400 text-xs">
              点击生成
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- TAB 2: Custom Reports (自定义报表) -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'custom'" class="flex gap-4 min-h-[600px]">
      <!-- Left: Saved Reports -->
      <div class="w-52 shrink-0 border border-gray-100 rounded-xl bg-white h-fit">
        <div class="px-4 py-3 border-b border-gray-100">
          <h3 class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ t('report.savedReports') }}</h3>
        </div>
        <div class="divide-y divide-gray-50 max-h-[400px] overflow-y-auto">
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
          <div class="flex flex-wrap items-end gap-3">
            <!-- Report Type -->
            <div>
              <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.reportType') }}</label>
              <select v-model="reportType" @change="onTypeChange" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                <option v-for="(label, key) in reportTypeLabels" :key="key" :value="key">{{ label }}</option>
              </select>
            </div>

            <!-- Group By -->
            <div v-if="reportType !== 'created_vs_resolved' && reportType !== 'created_trend'">
              <label class="block text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1">{{ t('report.groupBy') }}</label>
              <select v-model="groupBy" class="px-2.5 py-1.5 border border-gray-200 rounded-md text-sm bg-white focus:outline-none focus:ring-1 focus:ring-blue-400">
                <option v-for="d in dims" :key="d.value" :value="d.value">{{ d.label }}</option>
              </select>
            </div>

            <!-- Interval -->
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
                >{{ (chartLabels as Record<string, string>)[c] || c }}</button>
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

          <!-- Filter Mode -->
          <div class="flex items-center gap-3">
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
            <div v-if="activeFiltersCount > 0" class="text-[11px] text-gray-400 mt-1 font-mono truncate" :title="builtRqlPreview">
              RQL: {{ builtRqlPreview }}
            </div>
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

        <!-- Error -->
        <div v-if="rqlError" class="flex items-center gap-2 px-5 py-2 bg-red-50 border-b border-red-100 text-xs text-red-600">
          <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
          <span class="flex-1 truncate">{{ rqlError }}</span>
          <button @click="rqlError = null" class="text-red-400 hover:text-red-600">&times;</button>
        </div>

        <!-- Status Bar -->
        <div v-if="data" class="flex items-center justify-between px-5 py-2 border-b border-gray-100 text-xs text-gray-500 bg-gray-50/50">
          <div class="flex items-center gap-3">
            <span>{{ t('report.matched') }}: <strong class="text-gray-800">{{ data.total }}</strong> {{ t('report.issues') }}</span>
            <span v-if="data.summary?.avg_days">{{ t('report.avg') }}: <strong class="text-gray-800">{{ data.summary.avg_days.toFixed(1) }}</strong> {{ t('report.days') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button @click="exportCSV" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded text-xs transition-colors">CSV</button>
            <button v-if="chartType !== 'Table'" @click="exportPNG" class="px-2.5 py-1 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded text-xs transition-colors">PNG</button>
          </div>
        </div>

        <!-- Chart Area -->
        <div v-if="loading" class="flex items-center justify-center py-20 text-gray-400 text-sm">{{ t('report.loading') }}</div>
        <template v-else-if="data">
          <div v-show="chartType !== 'Table'" class="p-5">
            <div :class="['mx-auto', chartType === 'Pie' || chartType === 'Doughnut' ? 'max-w-md' : 'max-w-3xl']" style="height: 360px">
              <canvas :ref="setChartCanvas"></canvas>
            </div>
          </div>
          <div v-show="chartType === 'Table'" class="p-5">
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
                    <span class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ backgroundColor: chartColors[i % chartColors.length] }"></span>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, h, type FunctionalComponent } from 'vue'
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
function onTabClick(key: string) { activeTab.value = key as 'quick' | 'custom' }
const tabs = computed(() => [
  { key: 'quick', label: t('report.tabQuickCharts') },
  { key: 'custom', label: t('report.tabCustomReports') },
])

// ── Shared colors ──
const chartColors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899', '#06B6D4', '#84CC16', '#F97316', '#6366F1']

// ═══════════════════════════════════════════
// QUICK CHARTS
// ═══════════════════════════════════════════
const quickCanvas = ref<HTMLCanvasElement | null>(null)
function setQuickCanvas(el: any) {
  const el_ = Array.isArray(el) ? el[0] : el
  quickCanvas.value = el_ || null
}
const { render: quickRender, destroy: quickDestroy } = useReportChart(quickCanvas)
const quickData = ref<ReportResponse | null>(null)
const quickLoading = ref(false)
const quickActive = ref<{ type: string; groupBy: string } | null>(null)
const quickChartType = ref('Bar')
// Per-card chart type overrides: key = "type-groupBy"
const quickChartOverrides = ref<Record<string, string>>({})
// Per-card filter values: key = "type-groupBy" → { filterIndex: value }
const quickFilterValues = ref<Record<string, Record<number, string>>>({})

function getQuickChartType(q: { type: string; groupBy: string; defaultChart?: string }) {
  return quickChartOverrides.value[q.type + '-' + q.groupBy] || q.defaultChart || 'Bar'
}
function switchQuickChartType(q: { type: string; groupBy: string }, ct: string) {
  quickChartOverrides.value[q.type + '-' + q.groupBy] = ct
  if (quickActive.value?.type === q.type && quickActive.value?.groupBy === q.groupBy) {
    quickChartType.value = ct
    if (quickData.value) {
      nextTick().then(() => { setTimeout(() => quickRender(quickData.value!, ct), 50) })
    }
  }
}
function getQuickFilterRql(q: { type: string; groupBy: string; filters?: any[] }): string {
  if (!q.filters) return ''
  const key = q.type + '-' + q.groupBy
  const vals = quickFilterValues.value[key] || {}
  const parts: string[] = []
  q.filters.forEach((f: any, i: number) => {
    if (vals[i]) parts.push(`${f.field} = "${vals[i]}"`)
  })
  return parts.join(' AND ')
}
function onQuickFilterChange(q: { type: string; groupBy: string }, fi: number, val: string) {
  const key = q.type + '-' + q.groupBy
  if (!quickFilterValues.value[key]) quickFilterValues.value[key] = {}
  quickFilterValues.value[key][fi] = val
  if (quickActive.value?.type === q.type && quickActive.value?.groupBy === q.groupBy) runQuickChart(q)
}
function hasQuickFilters(q: { type: string; groupBy: string; filters?: any[] }) {
  if (!q.filters) return false
  const key = q.type + '-' + q.groupBy
  const vals = quickFilterValues.value[key] || {}
  return Object.values(vals).some(v => v)
}
function clearQuickFilters(q: { type: string; groupBy: string; filters?: any[] }) {
  const key = q.type + '-' + q.groupBy
  quickFilterValues.value[key] = {}
  if (quickActive.value?.type === q.type && quickActive.value?.groupBy === q.groupBy) runQuickChart(q)
}
function getQuickFilterVal(q: { type: string; groupBy: string }, fi: number): string {
  const key = q.type + '-' + q.groupBy
  return quickFilterValues.value[key]?.[fi] || ''
}

// SVG icons as render functions
const IconState: FunctionalComponent = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2, innerHTML: '<circle cx="12" cy="12" r="10"/><path d="M8 12l2 2 4-4"/>' })
const IconPriority: FunctionalComponent = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2, innerHTML: '<path d="M12 9v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>' })
const IconAssignee: FunctionalComponent = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2, innerHTML: '<path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87m-4-12a4 4 0 010 7.75"/>' })
const IconTrend: FunctionalComponent = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2, innerHTML: '<polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/>' })
const IconType: FunctionalComponent = () => h('svg', { width: 14, height: 14, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 2, innerHTML: '<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18M3 9h18"/>' })

const quickCharts = computed(() => {
  const priorityOpts = [
    { value: 'urgent', label: 'Urgent' }, { value: 'high', label: 'High' },
    { value: 'medium', label: 'Medium' }, { value: 'low', label: 'Low' },
  ]
  const stateOpts = states.value.length > 0
    ? states.value.map(s => ({ value: s.value, label: s.label }))
    : [{ value: '待处理', label: '待处理' }, { value: '进行中', label: '进行中' }, { value: '已完成', label: '已完成' }]
  return [
    { type: 'distribution', groupBy: 'state', label: t('report.quick.stateDistribution'), desc: t('report.quick.stateDistributionDesc'), color: '#3B82F6', icon: IconState, defaultChart: 'Bar', chartTypes: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], filters: [{ field: 'priority', label: t('report.priority'), options: priorityOpts }] },
    { type: 'distribution', groupBy: 'priority', label: t('report.quick.priorityDistribution'), desc: t('report.quick.priorityDistributionDesc'), color: '#F59E0B', icon: IconPriority, defaultChart: 'Pie', chartTypes: ['Pie', 'Doughnut', 'Bar', 'Table'] as string[], filters: [] },
    { type: 'distribution', groupBy: 'assignee', label: t('report.quick.assigneeWorkload'), desc: t('report.quick.assigneeWorkloadDesc'), color: '#10B981', icon: IconAssignee, defaultChart: 'Bar', chartTypes: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], filters: [{ field: 'state', label: t('report.state'), options: stateOpts }] },
    { type: 'distribution', groupBy: 'type', label: t('report.quick.typeDistribution'), desc: t('report.quick.typeDistributionDesc'), color: '#06B6D4', icon: IconType, defaultChart: 'Bar', chartTypes: ['Bar', 'Pie', 'Doughnut', 'Table'] as string[], filters: [] },
    { type: 'created_trend', groupBy: '', label: t('report.quick.creationTrend'), desc: t('report.quick.creationTrendDesc'), color: '#8B5CF6', icon: IconTrend, defaultChart: 'Area', chartTypes: ['Area', 'Line', 'Bar', 'Table'] as string[], filters: [{ field: 'priority', label: t('report.priority'), options: priorityOpts }] },
  ]
})

// Number widgets
const numberWidgets = computed(() => {
  const total = quickMetrics.value?.total ?? '—'
  const avgDays = quickMetrics.value?.avg_days
  const byState = quickMetrics.value?.byState
  const completed = byState?.find((s: any) => s.name === '已完成' || s.name === 'Done' || s.name === 'Completed')
  const completionRate = total !== '—' && completed ? Math.round((completed.count / total) * 100) + '%' : '—'
  return [
    { label: t('report.quick.totalIssues'), value: total, sub: '' },
    { label: t('report.quick.completionRate'), value: completionRate, sub: completed ? `${completed.count} / ${total}` : '' },
    { label: t('report.quick.avgAge'), value: avgDays ? avgDays.toFixed(1) + ' ' + t('report.days') : '—', sub: '' },
    { label: t('report.quick.stateGroups'), value: byState ? byState.length : '—', sub: '' },
  ]
})
const quickMetrics = ref<any>(null)

async function loadQuickMetrics() {
  try {
    const [distRes, avgRes] = await Promise.all([
      reportApi.generate(props.projectId, { report_type: 'distribution', group_by: 'state' }),
      reportApi.generate(props.projectId, { report_type: 'avg_age', group_by: 'state' }).catch(() => null),
    ])
    quickMetrics.value = { total: distRes.total, avg_days: avgRes?.summary?.avg_days, byState: distRes.labels.map((l: string, i: number) => ({ name: l, count: distRes.values[i] })) }
  } catch (_) { /* ignore */ }
}

async function runQuickChart(q: { type: string; groupBy: string }) {
  quickActive.value = { type: q.type, groupBy: q.groupBy }
  quickChartType.value = getQuickChartType(q)
  quickLoading.value = true
  try {
    const rql = getQuickFilterRql(q)
    const res = await reportApi.generate(props.projectId, {
      report_type: q.type,
      group_by: q.groupBy || undefined,
      interval: q.type === 'created_trend' ? 'week' : undefined,
      rql: rql || undefined,
    })
    quickData.value = res
    quickLoading.value = false
    await nextTick()
    await new Promise(r => setTimeout(r, 50))
    quickRender(res, quickChartType.value)
  } catch (e) {
    quickData.value = null
    quickLoading.value = false
  }
}

function qPct(v: number) { return quickData.value ? Math.round((v / quickData.value.total) * 100) : 0 }
function exportQuickCSV() { if (quickData.value) exportReportCSV(quickData.value, `quick-chart.csv`) }
function exportQuickPNG() { exportChartPNG(quickCanvas.value, `quick-chart.png`) }

// ═══════════════════════════════════════════
// CUSTOM REPORTS
// ═══════════════════════════════════════════
const reportType = ref('distribution')
const groupBy = ref('state')
const interval = ref('week')
const chartType = ref('Bar')
const dateFrom = ref('')
const dateTo = ref('')
const filterMode = ref<'basic' | 'rql'>('basic')
const rqlQuery = ref('')
interface FilterCondition { field: string; value: string }
const filters = ref<FilterCondition[]>([{ field: '', value: '' }])
const data = ref<ReportResponse | null>(null)
const loading = ref(false)
const rqlError = ref<string | null>(null)
const savedReports = ref<SavedReport[]>([])
const loadingSaved = ref(false)
const selectedId = ref<number | null>(null)
const showSaveDialog = ref(false)
const saveName = ref('')
const saving = ref(false)
const chartCanvas = ref<HTMLCanvasElement | null>(null)
function setChartCanvas(el: any) {
  chartCanvas.value = el || null
}
const { render: renderChart, destroy: destroyChart } = useReportChart(chartCanvas)

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
  day: t('report.intervals.day'),
  week: t('report.intervals.week'),
  month: t('report.intervals.month'),
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
const activeFiltersCount = computed(() => filters.value.filter(f => f.field && f.value).length)
const builtRqlPreview = computed(() => buildRQLFromFilters())

function buildRQLFromFilters(): string {
  const active = filters.value.filter(f => f.field && f.value)
  if (active.length === 0) return ''
  return active.map(f => `${f.field} = "${f.value.replace(/"/g, '\\"')}"`).join(' AND ')
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

function onFilterFieldChange(idx: number) { filters.value[idx].value = '' }
function addFilter() { filters.value.push({ field: '', value: '' }) }
function removeFilter(idx: number) {
  if (filters.value.length > 1) filters.value.splice(idx, 1)
  else filters.value[0] = { field: '', value: '' }
}

async function generate() {
  loading.value = true
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
    rqlError.value = null
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
  else if (data.value) {
    await nextTick()
    await new Promise(r => setTimeout(r, 30))
    renderChart(data.value, newVal)
  }
})

watch([filterMode, rqlQuery], () => { rqlError.value = null })
watch(filters, () => { rqlError.value = null }, { deep: true })

function exportCSV() { if (data.value) exportReportCSV(data.value, `report-${reportType.value}.csv`) }
function exportPNG() { exportChartPNG(chartCanvas.value, `chart-${reportType.value}.png`) }

async function loadSavedReports() {
  loadingSaved.value = true
  try { savedReports.value = await savedReportApi.list(props.projectId) }
  catch (e) { console.error(e) }
  finally { loadingSaved.value = false }
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
  if (r.rql) { filterMode.value = 'rql'; rqlQuery.value = r.rql }
  else { filterMode.value = 'basic'; rqlQuery.value = '' }
  generate()
}

async function saveReport() {
  if (!saveName.value.trim() || saving.value) return
  saving.value = true
  try {
    const rql = filterMode.value === 'basic' ? buildRQLFromFilters() : rqlQuery.value
    const payload = {
      name: saveName.value, report_type: reportType.value, group_by: groupBy.value,
      chart_type: chartType.value, rql, date_from: dateFrom.value, date_to: dateTo.value, interval: interval.value,
    }
    if (selectedId.value) await savedReportApi.update(props.projectId, selectedId.value, payload)
    else { const created = await savedReportApi.create(props.projectId, payload); selectedId.value = created.id! }
    showSaveDialog.value = false
    await loadSavedReports()
  } catch (e) {
    console.error(e)
    toast.error(t('report.saveFailed'))
  } finally { saving.value = false }
}

async function deleteSavedReport(r: SavedReport) {
  if (!r.id || !confirm(t('report.deleteConfirm', { name: r.name }))) return
  try {
    await savedReportApi.delete(props.projectId, r.id)
    if (selectedId.value === r.id) { selectedId.value = null; destroyChart() }
    await loadSavedReports()
  } catch (e) { console.error(e) }
}

watch(() => props.projectId, () => {
  destroyChart(); quickDestroy()
  loadSavedReports(); loadFilterOptions()
  selectedId.value = null; data.value = null; quickData.value = null; quickActive.value = null
})

onMounted(() => {
  loadSavedReports()
  loadFilterOptions()
  loadQuickMetrics()
})
</script>
