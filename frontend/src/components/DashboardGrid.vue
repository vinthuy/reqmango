<template>
  <div class="grid-container p-5">
    <div v-if="widgets.length === 0 && editMode"
      class="add-widget-spot border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl min-h-[200px] flex flex-col items-center justify-center gap-3 cursor-pointer hover:border-indigo-400 dark:hover:border-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-900/10 transition-colors"
      @click="$emit('add')">
      <span class="text-3xl text-gray-300 dark:text-gray-600">+</span>
      <span class="text-sm font-medium text-gray-400 dark:text-gray-500">{{ t('dashboard.addFirstWidget') }}</span>
    </div>

    <div v-else-if="widgets.length === 0" class="flex items-center justify-center min-h-[200px] text-sm text-gray-400 dark:text-gray-500">
      {{ t('dashboard.noWidgetsYet') }}
    </div>

    <div v-else class="grid gap-4" :style="{ gridTemplateColumns: `repeat(${columns}, 1fr)`, gridAutoRows: 'minmax(160px, auto)' }">
      <div v-for="w in widgets" :key="w.id"
        class="widget-card bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden transition-shadow hover:shadow-md"
        :class="{ 'hover:border-indigo-300 dark:hover:border-indigo-500 cursor-grab': editMode }"
        :style="getSpanStyle(w)">
        <!-- Header -->
        <div class="widget-header px-4 py-2.5 border-b border-gray-100 dark:border-gray-700 flex items-center gap-2">
          <span class="text-xs font-semibold text-gray-600 dark:text-gray-400 flex-1 truncate">{{ w.title }}</span>
          <span class="text-[10px] text-gray-400 dark:text-gray-500 px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">{{ widgetTypeLabel(w.widget_type) }}</span>
          <div v-if="editMode" class="flex gap-1">
            <button @click="$emit('configure', w)" class="w-6 h-6 rounded flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors" :title="t('common.configure')">
              <svg class="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065zM15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
            <button @click="$emit('delete-widget', w.id)" class="w-6 h-6 rounded flex items-center justify-center hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors" :title="t('common.delete')">
              <svg class="w-3.5 h-3.5 text-gray-400 hover:text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Body -->
        <div class="widget-body p-4 min-h-[100px] flex items-center justify-center" v-loading="loading">
          <WidgetCard :widget="w" :data="getWidgetDataById(w.id)" @configure="$emit('configure', w)" />
        </div>
      </div>

      <!-- Add widget spot -->
      <div v-if="editMode"
        class="add-widget-spot border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl flex flex-col items-center justify-center gap-2 cursor-pointer hover:border-indigo-400 dark:hover:border-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-900/10 transition-colors min-h-[160px]"
        :style="{ gridColumn: `span 3` }"
        @click="$emit('add')">
        <span class="text-2xl text-gray-300 dark:text-gray-600">+</span>
        <span class="text-xs font-medium text-gray-400 dark:text-gray-500">{{ t('dashboard.addWidget') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '@/composables/useI18n'
import type { DashboardWidget, WidgetDataResponse, WidgetType } from '@/types/dashboard'
import WidgetCard from '@/components/WidgetCard.vue'

const { t } = useI18n()

const props = defineProps<{
  widgets: DashboardWidget[]
  widgetData: WidgetDataResponse[]
  editMode: boolean
  loading: boolean
  columns?: number
}>()

defineEmits<{
  add: []
  configure: [widget: DashboardWidget]
  'delete-widget': [widgetId: number]
}>()

const columns = computed(() => props.columns || 12)

function widgetTypeLabel(type: WidgetType): string {
  const map: Record<WidgetType, string> = {
    number_card: t('dashboard.numberCard'),
    bar_chart: t('dashboard.barChart'),
    pie_chart: t('dashboard.pieChart'),
    doughnut_chart: t('dashboard.doughnutChart'),
    line_chart: t('dashboard.lineChart'),
    bubble_chart: t('dashboard.bubbleChart'),
    scatter_chart: t('dashboard.scatterChart'),
    mixed_chart: t('dashboard.mixedChart'),
    burndown: t('dashboard.burndown'),
    table: t('dashboard.table'),
    recent_list: t('dashboard.recentList'),
    saved_report: t('dashboard.savedReport'),
  }
  return map[type] ?? type
}

function getWidgetDataById(widgetId: number): Record<string, any> | null {
  const entry = props.widgetData.find((w) => w.widget_id === widgetId)
  return entry?.data ?? null
}

function getSpanStyle(w: DashboardWidget) {
  const pos = w.position
  const colSpan = pos?.w ?? 4
  return { gridColumn: `span ${Math.min(colSpan, columns.value)}` }
}
</script>

<style scoped>
.grid-container {
  min-height: calc(100vh - 180px);
}
</style>
