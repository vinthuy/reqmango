<template>
  <div class="p-6">
    <div class="mb-6">
      <h2 class="text-lg font-semibold text-gray-900">AI Settings</h2>
      <p class="text-sm text-gray-500 mt-1">Configure the AI provider for workspace-wide features</p>
    </div>

    <div v-if="loading" class="text-center py-8 text-gray-400">Loading...</div>

    <div v-else class="bg-white rounded-xl border border-gray-200 p-6 max-w-2xl space-y-4">
      <div class="flex items-center gap-3 mb-4 pb-4 border-b border-gray-100">
        <span class="text-2xl">🤖</span>
        <div>
          <span v-if="configured" class="px-2 py-0.5 bg-green-100 text-green-700 text-xs rounded-full font-medium">Configured</span>
          <span v-else class="px-2 py-0.5 bg-amber-100 text-amber-700 text-xs rounded-full font-medium">Not Configured</span>
        </div>
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Provider</label>
        <select v-model="form.provider" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
          <option value="deepseek">DeepSeek</option>
          <option value="anthropic">Anthropic (Claude)</option>
          <option value="openai">OpenAI / Compatible</option>
        </select>
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">API Key</label>
        <input
          v-model="form.api_key"
          type="password"
          :placeholder="configured ? '•••••••• (leave blank to keep current)' : 'sk-...'"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500"
        />
        <p class="text-xs text-gray-400 mt-1">Your API key is encrypted at rest and never exposed in responses.</p>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Model</label>
          <select v-model="form.model" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
            <optgroup label="DeepSeek">
              <option value="deepseek-chat">deepseek-chat</option>
              <option value="deepseek-reasoner">deepseek-reasoner</option>
            </optgroup>
            <optgroup label="Anthropic">
              <option value="claude-sonnet-4-6">claude-sonnet-4-6</option>
              <option value="claude-opus-4-8">claude-opus-4-8</option>
              <option value="claude-haiku-4-5">claude-haiku-4-5</option>
            </optgroup>
            <optgroup label="OpenAI">
              <option value="gpt-4o">gpt-4o</option>
              <option value="gpt-4-turbo">gpt-4-turbo</option>
            </optgroup>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Max Tokens</label>
          <input v-model.number="form.max_tokens" type="number" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" />
        </div>
      </div>

      <div class="flex items-center gap-2 pt-2">
        <input v-model="form.is_active" type="checkbox" class="rounded" id="ai-active" />
        <label for="ai-active" class="text-sm text-gray-700">Enable AI features</label>
      </div>

      <div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
        <button @click="testConnection" :disabled="testing" class="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50">
          {{ testing ? 'Testing...' : 'Test Connection' }}
        </button>
        <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50">
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
      </div>

      <div v-if="testResult" :class="['mt-2 p-3 rounded-lg text-sm', testResult.ok ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-600']">
        {{ testResult.msg }}
      </div>
      <div v-if="saveResult" class="mt-2 p-3 bg-green-50 text-green-700 rounded-lg text-sm">{{ saveResult }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import api from '@/api'

const props = defineProps<{ workspaceId: number }>()

const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const configured = ref(false)
const testResult = ref<{ ok: boolean; msg: string } | null>(null)
const saveResult = ref('')

const form = reactive({
  provider: 'deepseek',
  model: 'deepseek-chat',
  api_key: '',
  max_tokens: 4096,
  is_active: true,
})

onMounted(() => load())

async function load() {
  loading.value = true
  try {
    const ws = await api.get(`/workspaces/${props.workspaceId}/ai-config`)
    const d = ws.data
    configured.value = d.configured
    form.provider = d.provider || 'deepseek'
    form.model = d.model || 'deepseek-chat'
    form.max_tokens = d.max_tokens || 4096
    form.is_active = d.is_active !== false
  } catch (e: any) {
    if (e.response?.status !== 404) console.error(e)
  } finally { loading.value = false }
}

async function save() {
  saving.value = true
  saveResult.value = ''
  try {
    const payload: any = {
      provider: form.provider,
      model: form.model,
      max_tokens: form.max_tokens,
      is_active: form.is_active,
    }
    if (form.api_key) payload.api_key = form.api_key
    await api.put(`/workspaces/${props.workspaceId}/ai-config`, payload)
    configured.value = true
    form.api_key = ''
    saveResult.value = 'Settings saved successfully!'
  } catch (e: any) {
    saveResult.value = 'Failed: ' + (e.response?.data?.message || e.message)
  } finally { saving.value = false }
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  try {
    await api.post(`/projects/1/ai/chat?workspace_id=${props.workspaceId}`, {
      message: 'hi, respond with just "OK"', mode: 'ask',
    })
    testResult.value = { ok: true, msg: 'Connection successful! AI is responding.' }
  } catch (e: any) {
    const msg = e.response?.data?.message || e.message || 'Unknown error'
    testResult.value = { ok: false, msg: 'Connection failed: ' + msg.substring(0, 150) }
  } finally { testing.value = false }
}
</script>
