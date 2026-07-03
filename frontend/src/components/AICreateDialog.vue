<template>
  <div v-if="visible" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click.self="cancel">
    <div class="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
        <div class="flex items-center gap-2">
          <span class="text-xl">🤖</span>
          <div>
            <h3 class="text-lg font-semibold text-gray-900">{{ t('ai.smartCreate') }}</h3>
            <p class="text-xs text-gray-500">{{ t('ai.smartCreateDesc') }}</p>
          </div>
        </div>
        <button @click="cancel" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
      </div>

      <!-- Input -->
      <div class="p-4 border-b border-gray-100">
        <textarea
          v-model="description"
          :disabled="loading"
          rows="3"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 resize-none disabled:bg-gray-100"
          :placeholder="t('ai.promptExample')"
          @keydown.ctrl.enter="generate"
        ></textarea>
        <div class="flex justify-between items-center mt-2">
          <span class="text-xs text-gray-400">{{ t('ai.smartCreateHint') }}</span>
          <button
            @click="generate"
            :disabled="loading || !description.trim()"
            class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition"
          >
            {{ loading ? t('ai.generating') : t('ai.generatePreview') }}
          </button>
        </div>
      </div>

      <!-- Preview -->
      <div class="flex-1 overflow-y-auto p-4">
        <div v-if="loading" class="text-center py-8 text-gray-400">
          <div class="animate-spin h-6 w-6 border-2 border-indigo-500 border-t-transparent rounded-full mx-auto mb-2"></div>
          {{ t('ai.analyzing') }}
        </div>

        <div v-else-if="preview" class="space-y-4">
          <div class="bg-green-50 border border-green-200 rounded-lg p-3 text-sm text-green-700">
            {{ explanation }}
          </div>

          <div class="bg-gray-50 rounded-lg p-4 space-y-3">
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">{{ t('issue.title') }}</label>
              <input v-model="preview.name" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-medium text-gray-500 mb-1">{{ t('issue.priority') }}</label>
                <select v-model="preview.priority" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                  <option v-for="p in ['urgent','high','medium','low','none']" :key="p" :value="p">{{ t(`issue.priority${p.charAt(0).toUpperCase() + p.slice(1)}`) || p }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-500 mb-1">{{ t('issue.type') }} ID</label>
                <input v-model.number="preview.type_id" type="number" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">{{ t('issue.description') }}</label>
              <textarea v-model="preview.description" rows="4" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm resize-none"></textarea>
            </div>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div v-if="preview && !loading" class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
        <button @click="cancel" class="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition">{{ t('common.cancel') }}</button>
        <button @click="confirm" :disabled="creating" class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 transition">
          {{ creating ? t('issue.creating') : t('issue.create') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '@/composables/useI18n'
import { createPreviewWithAI } from '@/api/ai'
import { issueApi } from '@/api/issue'
import type { AICreateResponse } from '@/types/ai'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', issue: any): void
}>()

const description = ref('')
const loading = ref(false)
const creating = ref(false)
const preview = ref<Record<string, any> | null>(null)
const explanation = ref('')

async function generate() {
  if (!description.value.trim()) return
  loading.value = true
  preview.value = null
  try {
    const result: AICreateResponse = await createPreviewWithAI(props.projectId, {
      description: description.value.trim(),
    })
    preview.value = result.preview
    explanation.value = result.explanation
  } catch (e: any) {
    explanation.value = 'Failed to generate: ' + (e.message || 'Unknown error')
  } finally {
    loading.value = false
  }
}

async function confirm() {
  if (!preview.value?.name) return
  creating.value = true
  try {
    const payload = {
      name: preview.value.name,
      description: preview.value.description || '',
      priority: preview.value.priority || 'none',
      issue_type_id: preview.value.type_id || undefined,
      state_id: preview.value.state_id || undefined,
      project_id: props.projectId,
      workspace_id: props.workspaceId,
    }
    const created = await issueApi.createIssue(props.projectId, props.workspaceId, payload)
    emit('created', created)
    cancel()
  } catch (e: any) {
    alert('Creation failed: ' + (e.response?.data?.message || e.message))
  } finally {
    creating.value = false
  }
}

function cancel() {
  description.value = ''
  preview.value = null
  loading.value = false
  emit('close')
}
</script>
