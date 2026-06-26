<template>
  <div class="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4">
    <div class="bg-white rounded-xl shadow-lg p-8 w-full max-w-lg">
      <div class="text-center mb-6">
        <h1 class="text-2xl font-bold text-gray-900">{{ projectName }}</h1>
        <p class="text-sm text-gray-500 mt-1">Submit a new work item</p>
      </div>

      <div v-if="submitted" class="text-center py-8">
        <div class="text-4xl mb-3">✅</div>
        <h2 class="text-lg font-semibold text-gray-800">Submitted!</h2>
        <p class="text-sm text-gray-500 mt-1">Your item has been received and will be reviewed.</p>
      </div>

      <div v-else class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
          <input v-model="form.name" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="Brief summary of the issue" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea v-model="form.description" rows="4" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm resize-none" placeholder="Detailed description..."></textarea>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
            <select v-model="form.priority" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
              <option value="none">None</option>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="urgent">Urgent</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Your Name</label>
            <input v-model="form.submitter" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="Optional" />
          </div>
        </div>
        <button @click="submit" :disabled="submitting" class="w-full py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 font-medium">
          {{ submitting ? 'Submitting...' : 'Submit' }}
        </button>
        <p v-if="error" class="text-red-500 text-xs text-center">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api'

const route = useRoute()
const projectId = ref(0)
const projectName = ref('Intake Form')
const submitted = ref(false)
const submitting = ref(false)
const error = ref('')
const form = reactive({ name: '', description: '', priority: 'none', submitter: '' })

onMounted(async () => {
  projectId.value = parseInt((route.params as any).projectId as string, 10)
  try {
    const r = await api.get(`/projects/${projectId.value}`)
    projectName.value = r.data?.name || 'Intake Form'
  } catch (_) {}
})

async function submit() {
  if (!form.name.trim()) { error.value = 'Title is required'; return }
  submitting.value = true; error.value = ''
  try {
    await api.post(`/intake/${projectId.value}`, { ...form })
    submitted.value = true
  } catch (e: any) { error.value = e.response?.data?.message || 'Failed to submit' }
  finally { submitting.value = false }
}
</script>
