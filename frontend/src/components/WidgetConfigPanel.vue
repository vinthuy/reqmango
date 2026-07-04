<template>
  <div class="config-overlay fixed inset-0 z-[200] flex justify-end">
    <div class="absolute inset-0 bg-black/20" @click="$emit('close')"></div>
    <div class="config-panel relative w-[400px] h-full bg-white dark:bg-gray-800 shadow-xl flex flex-col z-[201]">
      <!-- Header -->
      <div class="px-5 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center gap-3">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 flex-1">{{ t('dashboard.configureWidget') }}</h3>
        <button @click="$emit('close')" class="w-7 h-7 rounded flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700">
          <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Body -->
      <div class="flex-1 overflow-y-auto p-5 flex flex-col gap-4">
        <!-- Title -->
        <div class="form-group">
          <label class="form-label">{{ t('dashboard.widgetTitle') }}</label>
          <input v-model="form.title" class="form-input" :placeholder="t('dashboard.widgetTitle')" />
        </div>

        <!-- Chart config -->
        <template v-if="isChartWidget">
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.reportType') }}</label>
            <select v-model="form.config.report_type" class="form-input">
              <option value="distribution">{{ t('dashboard.distribution') }}</option>
              <option value="created_vs_resolved">{{ t('dashboard.createdVsResolved') }}</option>
              <option value="created_trend">{{ t('dashboard.createdTrend') }}</option>
              <option value="avg_age">{{ t('dashboard.avgAge') }}</option>
              <option value="current_age">{{ t('dashboard.currentAge') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.groupBy') }}</label>
            <select v-model="form.config.group_by" class="form-input">
              <option value="state">{{ t('common.state') }}</option>
              <option value="priority">{{ t('common.priority') }}</option>
              <option value="assignees">{{ t('common.assignees') }}</option>
              <option value="labels">{{ t('common.labels') }}</option>
              <option value="state_group">{{ t('dashboard.stateGroup') }}</option>
              <option value="cycle">{{ t('common.cycle') }}</option>
              <option value="module">{{ t('common.module') }}</option>
              <option value="issue_type">{{ t('common.issueType') }}</option>
              <option value="created_by">{{ t('common.createdBy') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.interval') }}</label>
            <select v-model="form.config.interval" class="form-input">
              <option value="day">{{ t('dashboard.daily') }}</option>
              <option value="week">{{ t('dashboard.weekly') }}</option>
              <option value="month">{{ t('dashboard.monthly') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">RQL</label>
            <textarea v-model="form.config.rql" rows="3" class="form-input font-mono text-xs" placeholder="priority = 'high'" />
          </div>
        </template>

        <!-- Number card config -->
        <template v-else-if="widget.widget_type === 'number_card'">
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.metric') }}</label>
            <select v-model="form.config.metric" class="form-input">
              <option value="total">{{ t('dashboard.totalIssues') }}</option>
              <option value="completed">{{ t('dashboard.completed') }}</option>
              <option value="in_progress">{{ t('dashboard.inProgress') }}</option>
              <option value="overdue">{{ t('dashboard.overdue') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.label') }}</label>
            <input v-model="form.config.label" class="form-input" :placeholder="t('dashboard.labelPlaceholder')" />
          </div>
        </template>

        <!-- Burndown config -->
        <template v-else-if="widget.widget_type === 'burndown'">
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.selectCycle') }}</label>
            <select v-model="form.config.cycle_id" class="form-input">
              <option :value="null">{{ t('dashboard.selectCycle') }}</option>
              <option v-for="c in cycles" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
        </template>

        <!-- Saved Report config -->
        <template v-else-if="widget.widget_type === 'saved_report'">
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.selectReport') }}</label>
            <select v-model="form.config.saved_report_id" class="form-input">
              <option :value="null">{{ t('dashboard.selectReport') }}</option>
              <option v-for="r in savedReports" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>
        </template>

        <!-- Recent list config -->
        <template v-else-if="widget.widget_type === 'recent_list'">
          <div class="form-group">
            <label class="form-label">{{ t('dashboard.maxItems') }}</label>
            <input v-model.number="form.config.limit" type="number" min="1" max="50" class="form-input" />
          </div>
        </template>

        <!-- Dimensions -->
        <div class="form-group">
          <label class="form-label">{{ t('dashboard.position') }}</label>
          <div class="grid grid-cols-2 gap-2">
            <div>
              <span class="text-[10px] text-gray-400">{{ t('dashboard.width') }}</span>
              <input v-model.number="form.position.w" type="number" min="1" max="12" class="form-input mt-0.5" />
            </div>
            <div>
              <span class="text-[10px] text-gray-400">{{ t('dashboard.height') }}</span>
              <input v-model.number="form.position.h" type="number" min="1" max="6" class="form-input mt-0.5" />
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-5 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2">
        <button class="btn btn-ghost btn-sm" @click="$emit('close')">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary btn-sm" @click="save">{{ t('common.save') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { listCycles } from '@/api/cycle'
import { savedReportApi } from '@/api/report'
import type { DashboardWidget } from '@/types/dashboard'
import type { SavedReport } from '@/api/report'

const { t } = useI18n()

const props = defineProps<{
  widget: DashboardWidget
  projectId: number
}>()

const emit = defineEmits<{
  close: []
  save: [updates: Record<string, any>]
}>()

const isChartWidget = computed(() =>
  ['bar_chart', 'pie_chart', 'doughnut_chart', 'line_chart', 'table'].includes(props.widget.widget_type)
)

const cycles = reactive<{ id: number; name: string }[]>([])
const savedReports = reactive<SavedReport[]>([])

const form = reactive({
  title: props.widget.title,
  config: { ...(props.widget.config ?? {}) },
  position: { w: 4, h: 3, ...(props.widget.position ?? {}) },
})

// Initialize default config values
if (isChartWidget.value && !form.config.report_type) {
  form.config.report_type = 'distribution'
  form.config.group_by = form.config.group_by || 'state'
  form.config.interval = form.config.interval || 'week'
  form.config.rql = form.config.rql || ''
}

// Fetch cycles when configuring a burndown widget
async function fetchCycles() {
  if (props.widget.widget_type !== 'burndown') return
  try {
    const result = await listCycles(props.projectId)
    cycles.splice(0, cycles.length, ...result.items.map(c => ({ id: c.id, name: c.name })))
  } catch {
    // silently ignore
  }
}

// Fetch saved reports when configuring a saved_report widget
async function fetchSavedReports() {
  if (props.widget.widget_type !== 'saved_report') return
  try {
    const reports = await savedReportApi.list(props.projectId)
    savedReports.splice(0, savedReports.length, ...reports)
  } catch {
    // silently ignore
  }
}

onMounted(() => {
  fetchCycles()
  fetchSavedReports()
})
watch(() => props.widget.widget_type, () => {
  fetchCycles()
  fetchSavedReports()
})

function save() {
  emit('save', {
    title: form.title,
    config: form.config,
    position: form.position,
  })
}
</script>

<style scoped>
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-label { font-size: 10px; font-weight: 600; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.04em; }
.form-input {
  padding: 7px 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  width: 100%;
  background: #fff;
  color: #374151;
}
html.dark .form-input {
  background: #374151;
  border-color: #4b5563;
  color: #e5e7eb;
}
.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,.15);
}
.btn { padding: 6px 14px; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; border: none; }
.btn-ghost { background: transparent; color: #6b7280; border: 1px solid #d1d5db; }
.btn-ghost:hover { background: #f3f4f6; color: #374151; }
.btn-primary { background: #111827; color: #fff; }
.btn-primary:hover { background: #1f2937; }
.btn-sm { padding: 4px 10px; font-size: 12px; }
</style>
