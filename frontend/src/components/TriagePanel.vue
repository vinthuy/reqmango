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
          <div class="flex items-center gap-2 ml-4">
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

const props = defineProps<{ projectId: number }>()
defineEmits<{ (e: 'showForm'): void }>()

const items = ref<any[]>([])
const loading = ref(false)

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
  } catch (e: any) { alert(e.response?.data?.message || 'Failed') }
}

function formatDate(d: string) { return new Date(d).toLocaleDateString() }
</script>
