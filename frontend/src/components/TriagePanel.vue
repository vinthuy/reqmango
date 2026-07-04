<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">Triage</h2>
        <p class="text-sm text-gray-500 mt-1">Review and manage incoming work items</p>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-sm text-gray-500">{{ items.length }} pending</span>
        <button @click="$emit('showForm')" class="px-3 py-1.5 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700">
          + Intake Form Link
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else-if="items.length === 0" class="text-center py-12 bg-gray-50 rounded-lg border border-gray-200">
      <p class="text-gray-500">No items to triage</p>
      <p class="text-xs text-gray-400 mt-1">Share the intake form link to receive submissions</p>
    </div>

    <div v-else class="space-y-3">
      <div v-for="item in items" :key="item.id" class="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 transition">
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-1">
              <span class="text-sm font-medium text-gray-900">{{ item.name }}</span>
              <span class="px-1.5 py-0.5 bg-amber-100 text-amber-700 text-xs rounded">Pending</span>
            </div>
            <p v-if="item.description_html && item.description_html !== '<p></p>'" class="text-xs text-gray-500 line-clamp-2" v-html="item.description_html"></p>
            <div class="flex items-center gap-3 mt-2 text-xs text-gray-400">
              <span>{{ item.priority }}</span>
              <span>{{ formatDate(item.created_at) }}</span>
            </div>
          </div>
          <!-- AI Suggestion -->
          <div v-if="aiResults[item.id]" class="mt-2 p-2 bg-indigo-50 rounded text-xs">
            <div class="flex items-center gap-2 mb-1">
              <span class="font-medium text-indigo-700">🤖 AI:</span>
              <span class="text-indigo-600">{{ aiResults[item.id].suggested_type }}</span>
              <span :class="'px-1 rounded text-white '+(aiResults[item.id].suggested_priority==='urgent'?'bg-red-500':'bg-amber-500')">{{ aiResults[item.id].suggested_priority }}</span>
            </div>
            <div class="text-gray-600">{{ aiResults[item.id].summary }}</div>
            <div v-if="aiResults[item.id].has_duplicates" class="text-amber-600 mt-1">⚠ Possible duplicates: #{{ aiResults[item.id].duplicate_ids?.join(', #') }}</div>
          </div>

          <div class="flex items-center gap-2 ml-4">
            <button @click="analyzeAI(item.id)" class="px-2 py-1.5 text-xs border border-indigo-300 text-indigo-600 rounded hover:bg-indigo-50" :disabled="analyzing[item.id]">
              {{ analyzing[item.id] ? '...' : '🤖' }}
            </button>
            <button @click="triage(item.id, 'accept')" class="px-3 py-1.5 bg-green-600 text-white text-xs rounded hover:bg-green-700">Accept</button>
            <button @click="triage(item.id, 'reject')" class="px-3 py-1.5 bg-red-500 text-white text-xs rounded hover:bg-red-600">Reject</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ projectId: number }>()
const toast = useToast()
defineEmits<{ (e: 'showForm'): void }>()

const items = ref<any[]>([])
const loading = ref(false)
const aiResults = ref<Record<number, any>>({})
const analyzing = ref<Record<number, boolean>>({})

async function analyzeAI(issueId: number) {
  analyzing.value[issueId] = true
  try {
    const r = await api.post(`/projects/${props.projectId}/intake/${issueId}/ai-analyze`)
    aiResults.value[issueId] = r.data
  } catch (_) {}
  finally { analyzing.value[issueId] = false }
}

onMounted(() => load())

async function load() {
  loading.value = true
  try { const r = await api.get(`/projects/${props.projectId}/intake`); items.value = r.data || [] }
  catch (_) {}
  finally { loading.value = false }
}

async function triage(issueId: number, action: string) {
  try {
    await api.post(`/projects/${props.projectId}/intake/${issueId}/triage`, { action })
    load()
  } catch (e: any) { toast.error(e.response?.data?.message || 'Failed') }
}

function formatDate(d: string) { return new Date(d).toLocaleDateString() }
</script>
