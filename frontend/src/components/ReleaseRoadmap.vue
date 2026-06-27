<template>
  <div class="p-6">
    <div class="mb-6"><h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">🚀 Release Roadmap</h2><p class="text-sm text-gray-500 mt-1">Timeline of planned releases</p></div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else-if="releases.length === 0" class="text-center py-12 bg-gray-50 dark:bg-gray-800 rounded-lg border text-gray-400 text-sm">
      No releases configured. Go to Project Settings → Releases to create one.
    </div>

    <div v-else class="relative">
      <!-- Timeline line -->
      <div class="absolute left-8 top-0 bottom-0 w-0.5 bg-indigo-200 dark:bg-indigo-800"></div>

      <div v-for="(rel, idx) in sortedReleases" :key="rel.id" class="relative pl-16 pb-8">
        <!-- Timeline dot -->
        <div class="absolute left-6 w-5 h-5 rounded-full border-2 border-indigo-500 bg-white dark:bg-gray-800 z-10"
          :class="statusDot(rel.status)"></div>

        <!-- Card -->
        <div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md transition">
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center gap-2">
              <span class="font-semibold text-gray-900 dark:text-gray-100">{{ rel.name }}</span>
              <span class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-2 py-0.5 rounded-full font-mono">{{ rel.version }}</span>
            </div>
            <span :class="statusBadge(rel.status)" class="px-2 py-0.5 text-xs rounded-full font-medium">{{ rel.status }}</span>
          </div>
          <div class="flex items-center gap-4 text-xs text-gray-500">
            <span v-if="rel.release_date">📅 {{ formatDate(rel.release_date) }}</span>
            <span>{{ rel.issue_count || 0 }} issues</span>
          </div>
          <!-- Progress bar -->
          <div v-if="rel.issue_count" class="mt-3">
            <div class="flex justify-between text-xs text-gray-400 mb-1">
              <span>{{ rel.completed_count || 0 }}/{{ rel.issue_count }} done</span>
              <span>{{ pct(rel) }}%</span>
            </div>
            <div class="h-2 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
              <div class="h-full bg-green-500 rounded-full transition-all" :style="{width: pct(rel)+'%'}"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api'

const props = defineProps<{ projectId: number }>()
const releases = ref<any[]>([])
const loading = ref(false)

const sortedReleases = computed(() =>
  [...releases.value].sort((a, b) => new Date(a.release_date||0).getTime() - new Date(b.release_date||0).getTime())
)

onMounted(() => load())
async function load() {
  loading.value = true
  try {
    const r = await api.get(`/projects/${props.projectId}/releases`)
    const list = r.data?.data || r.data || []
    // Load issue counts for each release
    for (const rel of list) {
      try {
        const pr = await api.get(`/projects/${props.projectId}/releases/${rel.id}/progress`)
        const pdata = pr.data?.data || pr.data || {}
        rel.issue_count = pdata.total_issues || pdata.total || 0
        rel.completed_count = pdata.completed_issues || pdata.completed || 0
      } catch (_) { rel.issue_count = 0 }
    }
    releases.value = list
  } catch (_) { releases.value = [] }
  finally { loading.value = false }
}

function statusDot(s: string) {
  if (s === 'released') return 'bg-green-500 border-green-500'
  if (s === 'in_progress') return 'bg-amber-500 border-amber-500 animate-pulse'
  return 'bg-gray-300'
}

function statusBadge(s: string) {
  if (s === 'released') return 'bg-green-100 text-green-700'
  if (s === 'in_progress') return 'bg-amber-100 text-amber-700'
  if (s === 'cancelled') return 'bg-red-100 text-red-600'
  return 'bg-gray-100 text-gray-600'
}

function pct(rel: any) { return rel.issue_count ? Math.round((rel.completed_count||0)/rel.issue_count*100) : 0 }
function formatDate(d: string) { return new Date(d).toLocaleDateString('zh-CN') }
</script>
