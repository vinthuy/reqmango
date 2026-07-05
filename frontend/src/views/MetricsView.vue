<template>
  <div class="space-y-6">
    <!-- Tabs -->
    <div class="flex items-center gap-1 border-b border-gray-200">
      <button
        @click="activeTab = 'templates'"
        :class="[
          'px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px',
          activeTab === 'templates'
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
        ]"
      >预置模板</button>
      <button
        @click="activeTab = 'charts'"
        :class="[
          'px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px',
          activeTab === 'charts'
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
        ]"
      >我的图表</button>
    </div>

    <!-- Templates Tab -->
    <div v-if="activeTab === 'templates'">
      <MetricsTemplateGallery
        v-if="categories.length > 0"
        :categories="categories"
        @use-template="handleUseTemplate"
      />
      <div v-else-if="loading" class="flex items-center justify-center py-12 text-sm text-gray-400">
        加载中...
      </div>
      <div v-else class="flex flex-col items-center justify-center py-12 text-sm text-gray-400">
        暂无模板
      </div>
    </div>

    <!-- Charts Tab -->
    <div v-if="activeTab === 'charts'">
      <!-- New Chart Button -->
      <div class="flex items-center justify-end">
        <button
          @click="handleNewChart"
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          新建图表
        </button>
      </div>

      <!-- Chart Grid -->
      <div v-if="charts.length > 0" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <MetricsChartCard
          v-for="chart in charts"
          :key="chart.id"
          :chart="chart"
          :project-id="projectId"
          @edit="handleEditChart"
          @delete="handleDeleteChart"
          @type-change="handleTypeChange"
        />
      </div>
      <div v-else-if="loading" class="flex items-center justify-center py-12 text-sm text-gray-400">
        加载中...
      </div>
      <div v-else class="flex flex-col items-center justify-center py-12 text-sm text-gray-400">
        暂无图表，点击"新建图表"开始创建
      </div>
    </div>

    <!-- MetricsChartConfig Dialog -->
    <MetricsChartConfig
      :project-id="projectId"
      :visible="showConfig"
      :template="configTemplate"
      :chart="configChart"
      @save="handleSave"
      @cancel="showConfig = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { TemplateCategory, MetricTemplate, MetricChart, CreateChartPayload } from '@/types/metrics'
import { metricsApi } from '@/api/metrics'
import { useConfirm } from '@/composables/useConfirm'
import MetricsTemplateGallery from '@/components/metrics/MetricsTemplateGallery.vue'
import MetricsChartCard from '@/components/metrics/MetricsChartCard.vue'
import MetricsChartConfig from '@/components/metrics/MetricsChartConfig.vue'

const props = defineProps<{
  projectId: number
}>()

const { confirm } = useConfirm()

// ── State ──
const activeTab = ref<'templates' | 'charts'>('templates')
const categories = ref<TemplateCategory[]>([])
const charts = ref<MetricChart[]>([])
const showConfig = ref(false)
const configTemplate = ref<MetricTemplate | undefined>()
const configChart = ref<MetricChart | undefined>()
const loading = ref(false)

// ── Data Fetching ──
async function fetchTemplates() {
  try {
    const data = await metricsApi.listTemplates(props.projectId)
    categories.value = data
  } catch (e) {
    console.error('Failed to load templates:', e)
  }
}

async function fetchCharts() {
  try {
    const data = await metricsApi.listCharts(props.projectId)
    charts.value = data
  } catch (e) {
    console.error('Failed to load charts:', e)
  }
}

async function loadData() {
  loading.value = true
  try {
    await Promise.all([fetchTemplates(), fetchCharts()])
  } finally {
    loading.value = false
  }
}

// ── Handlers ──
function handleUseTemplate(template: MetricTemplate) {
  configTemplate.value = template
  configChart.value = undefined
  showConfig.value = true
}

function handleNewChart() {
  configTemplate.value = undefined
  configChart.value = undefined
  showConfig.value = true
}

function handleEditChart(chart: MetricChart) {
  configTemplate.value = undefined
  configChart.value = chart
  showConfig.value = true
}

async function handleDeleteChart(chartId: number) {
  const ok = await confirm({
    title: '删除图表',
    message: '确定要删除该图表吗？此操作不可撤销。',
    confirmText: '删除',
    danger: true,
  })
  if (!ok) return
  try {
    await metricsApi.deleteChart(props.projectId, chartId)
    await fetchCharts()
  } catch (e) {
    console.error('Failed to delete chart:', e)
  }
}

async function handleTypeChange(chartId: number, newType: string) {
  try {
    await metricsApi.updateChart(props.projectId, chartId, { chart_type: newType })
    await fetchCharts()
  } catch (e) {
    console.error('Failed to update chart type:', e)
  }
}

async function handleSave(payload: CreateChartPayload) {
  try {
    if (configChart.value) {
      await metricsApi.updateChart(props.projectId, configChart.value.id, payload)
    } else {
      await metricsApi.createChart(props.projectId, payload)
    }
    showConfig.value = false
    await fetchCharts()
  } catch (e) {
    console.error('Failed to save chart:', e)
  }
}

// ── Lifecycle ──
onMounted(() => {
  loadData()
})
</script>
