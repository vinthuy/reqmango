<template>
  <div class="p-6">
    <div class="mb-6"><h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">📊 Custom Reports</h2><p class="text-sm text-gray-500 mt-1">Build your own reports with flexible grouping and chart types</p></div>

    <!-- Controls -->
    <div class="flex flex-wrap items-center gap-3 mb-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
      <div>
        <label class="block text-xs font-medium text-gray-500 mb-1">Group By</label>
        <select v-model="groupBy" @change="generate" class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 dark:text-gray-200">
          <option v-for="o in dims" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-gray-500 mb-1">Chart</label>
        <div class="inline-flex bg-gray-200 dark:bg-gray-700 rounded-lg p-0.5">
          <button v-for="c in charts" :key="c" @click="chart=c;generate()" :class="['px-2 py-1 text-xs rounded',chart===c?'bg-white dark:bg-gray-600 shadow font-medium':'text-gray-500']">{{ c }}</button>
        </div>
      </div>
      <div class="flex gap-2">
        <div><label class="block text-xs font-medium text-gray-500 mb-1">From</label><input v-model="dateFrom" type="date" @change="generate" class="px-2 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-700 dark:text-gray-200" /></div>
        <div><label class="block text-xs font-medium text-gray-500 mb-1">To</label><input v-model="dateTo" type="date" @change="generate" class="px-2 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-700 dark:text-gray-200" /></div>
      </div>
      <button @click="generate" :disabled="loading" class="px-3 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50 mt-5">Generate</button>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <!-- Total -->
    <div v-if="data" class="mb-4 text-sm text-gray-600 dark:text-gray-400">Total: <strong>{{ data.total }}</strong> issues</div>

    <!-- Bar Chart -->
    <div v-if="data && chart === 'Bar'" class="space-y-2 max-w-2xl">
      <div v-for="(label, i) in data.labels" :key="label" class="flex items-center gap-2">
        <span class="text-xs text-gray-600 dark:text-gray-400 w-24 truncate text-right">{{ label }}</span>
        <div class="flex-1 bg-gray-100 dark:bg-gray-700 rounded-full h-5 overflow-hidden">
          <div class="h-full rounded-full transition-all duration-500 flex items-center justify-end pr-2" :style="{width: pct(data.values[i])+'%', backgroundColor: data.colors?.[label] || '#6366F1'}">
            <span v-if="pct(data.values[i]) > 15" class="text-[10px] text-white font-medium">{{ data.values[i] }}</span>
          </div>
        </div>
        <span v-if="pct(data.values[i]) <= 15" class="text-xs text-gray-500">{{ data.values[i] }}</span>
      </div>
    </div>

    <!-- Pie Chart -->
    <div v-if="data && chart === 'Pie'" class="max-w-md">
      <svg viewBox="0 0 200 200" class="w-64 h-64 mx-auto">
        <circle v-for="(seg, i) in pieSegments" :key="i" :cx="100" :cy="100" :r="80"
          fill="none" stroke-width="40"
          :stroke="data.colors?.[data.labels[i]] || '#6366F1'"
          :stroke-dasharray="seg.dash"
          :stroke-dashoffset="seg.offset"
          :transform="'rotate(-90 100 100)'"
          class="transition-all duration-500" />
      </svg>
      <div class="flex flex-wrap justify-center gap-3 mt-4">
        <div v-for="(label, i) in data.labels" :key="label" class="flex items-center gap-1 text-xs">
          <span class="w-3 h-3 rounded" :style="{backgroundColor: data.colors?.[label] || '#6366F1'}"></span>
          {{ label }} ({{ data.values[i] }})
        </div>
      </div>
    </div>

    <!-- Table -->
    <table v-if="data && chart === 'Table'" class="w-full max-w-md text-sm">
      <thead><tr class="border-b border-gray-200 dark:border-gray-700"><th class="text-left py-2 text-gray-500">Group</th><th class="text-right py-2 text-gray-500">Count</th><th class="text-right py-2 text-gray-500">%</th></tr></thead>
      <tbody>
        <tr v-for="(label, i) in data.labels" :key="label" class="border-b border-gray-100 dark:border-gray-800">
          <td class="py-2 text-gray-800 dark:text-gray-200">{{ label }}</td>
          <td class="text-right py-2 font-medium">{{ data.values[i] }}</td>
          <td class="text-right py-2 text-gray-400">{{ pct(data.values[i]) }}%</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api'

const props = defineProps<{ projectId: number }>()
const groupBy = ref('state'); const chart = ref('Bar')
const dateFrom = ref(''); const dateTo = ref('')
const data = ref<any>(null); const loading = ref(false)

const dims = [
  {value:'state',label:'State'},{value:'priority',label:'Priority'},
  {value:'assignee',label:'Assignee'},{value:'type',label:'Type'},{value:'label',label:'Label'}
]
const charts = ['Bar','Pie','Table']

const pieSegments = computed(() => {
  if (!data.value) return []
  const total = data.value.values.reduce((a:number,b:number)=>a+b,0) || 1
  const circumference = 2 * Math.PI * 80
  let offset = 0
  return data.value.values.map((v:number) => {
    const len = (v/total) * circumference
    const seg = { dash: `${len} ${circumference-len}`, offset: -offset }
    offset += len
    return seg
  })
})

function pct(v: number) { return data.value ? Math.round(v/data.value.total*100) : 0 }

onMounted(() => generate())
async function generate() {
  loading.value = true
  try {
    const r = await api.post(`/projects/${props.projectId}/reports`, {
      group_by: groupBy.value, metric: 'count',
      date_from: dateFrom.value || undefined, date_to: dateTo.value || undefined,
    })
    data.value = r.data
  } catch (_) { data.value = null }
  finally { loading.value = false }
}
</script>
