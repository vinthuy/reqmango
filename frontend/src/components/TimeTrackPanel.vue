<template>
  <div class="border-t border-gray-100 pt-4 mt-4">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-sm font-semibold text-gray-700">⏱️ Time Tracking</h4>
      <div class="flex items-center gap-2">
        <span v-if="summary" class="text-xs text-gray-500">
          {{ summary.total_hours.toFixed(1) }}h / {{ summary.entry_count }} entries
        </span>
        <button v-if="!isRunning" @click="start" :disabled="loading" class="px-3 py-1 bg-green-600 text-white text-xs rounded hover:bg-green-700 disabled:opacity-50">
          ▶ Start
        </button>
        <button v-else @click="stop" :disabled="loading" class="px-3 py-1 bg-red-500 text-white text-xs rounded hover:bg-red-600 disabled:opacity-50">
          ⏹ Stop
        </button>
      </div>
    </div>

    <div v-if="isRunning" class="mb-3 p-2 bg-green-50 border border-green-200 rounded text-xs text-green-700 flex items-center gap-2">
      <span class="animate-pulse">🔴</span> Timer running...
    </div>

    <div v-if="entries.length" class="space-y-1 max-h-40 overflow-y-auto">
      <div v-for="e in entries" :key="e.id" class="flex items-center justify-between text-xs py-1.5 px-2 rounded hover:bg-gray-50">
        <div class="flex items-center gap-2">
          <span class="text-gray-400">{{ formatDate(e.started_at) }}</span>
          <span class="text-gray-600">{{ e.description || 'No description' }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-gray-500 font-mono">{{ formatDuration(e.duration) }}</span>
          <button @click="remove(e.id)" class="text-gray-400 hover:text-red-500">✕</button>
        </div>
      </div>
    </div>
    <div v-else class="text-xs text-gray-400 py-2">No time entries yet.</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as ttApi from '@/api/time-track'
import type { TimeTrack, TimeTrackSummary } from '@/types/time-track'

const props = defineProps<{ issueId: number }>()

const entries = ref<TimeTrack[]>([])
const summary = ref<TimeTrackSummary | null>(null)
const isRunning = ref(false)
const loading = ref(false)

onMounted(() => load())

async function load() {
  try {
    const [e, s] = await Promise.all([ttApi.listTimeTracks(props.issueId), ttApi.getTimeSummary(props.issueId)])
    entries.value = e; summary.value = s; isRunning.value = s.is_running
  } catch (_) { /* ignore */ }
}

async function start() {
  loading.value = true
  try { await ttApi.startTimer(props.issueId); await load() } catch (_) {}
  finally { loading.value = false }
}

async function stop() {
  loading.value = true
  try { await ttApi.stopTimer(props.issueId); await load() } catch (_) {}
  finally { loading.value = false }
}

async function remove(id: number) {
  try { await ttApi.deleteTimeTrack(props.issueId, id); await load() } catch (_) {}
}

function formatDuration(sec: number): string {
  if (!sec) return '0m'
  const h = Math.floor(sec/3600); const m = Math.floor((sec%3600)/60)
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

function formatDate(d: string): string {
  return new Date(d).toLocaleTimeString([], { hour:'2-digit', minute:'2-digit' })
}
</script>
