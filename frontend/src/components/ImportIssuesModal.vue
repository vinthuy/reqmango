<template>
  <div v-if="visible" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="handleClose">
    <div class="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] flex flex-col">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-800">{{ t('import.title') }}</h2>
        <button @click="handleClose" class="text-gray-400 hover:text-gray-600 text-xl">&times;</button>
      </div>

      <div class="flex-1 overflow-y-auto p-6">
        <div class="flex space-x-2 mb-6">
          <button
            @click="mode = 'csv'"
            :class="['px-4 py-2 text-sm font-medium rounded-md', mode === 'csv' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200']"
          >
            {{ t('import.csvImport') }}
          </button>
          <button
            @click="mode = 'json'"
            :class="['px-4 py-2 text-sm font-medium rounded-md', mode === 'json' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200']"
          >
            {{ t('import.jsonImport') }}
          </button>
        </div>

        <div v-if="mode === 'csv'" class="space-y-4">
          <div class="bg-blue-50 border border-blue-200 rounded-md p-4">
            <h4 class="text-sm font-medium text-blue-800 mb-2">{{ t('import.csvFormat') }}</h4>
            <p class="text-xs text-blue-700 mb-2">{{ t('import.csvFormatHint') }}</p>
            <ul class="text-xs text-blue-700 space-y-1 list-disc list-inside">
              <li><code>name</code> / {{ t('import.colTitle') }} - {{ t('import.colTitleHint') }}</li>
              <li><code>description</code> / {{ t('import.colDescription') }} - {{ t('import.colDescriptionHint') }}</li>
              <li><code>priority</code> / {{ t('import.colPriority') }} - urgent/high/medium/low/none</li>
              <li><code>state</code> / {{ t('import.colState') }} - {{ t('import.colStateHint') }}</li>
              <li><code>type</code> / {{ t('import.colType') }} - {{ t('import.colTypeHint') }}</li>
              <li><code>assignees</code> / {{ t('import.colAssignees') }} - {{ t('import.colAssigneesHint') }}</li>
              <li><code>labels</code> / {{ t('import.colLabels') }} - {{ t('import.colLabelsHint') }}</li>
              <li><code>start_date</code> / {{ t('import.colStartDate') }} - YYYY-MM-DD</li>
              <li><code>target_date</code> / {{ t('import.colTargetDate') }} - YYYY-MM-DD</li>
              <li><code>parent_title</code> / {{ t('import.colParentTitle') }} - {{ t('import.colParentTitleHint') }}</li>
              <li><code>module</code> / {{ t('import.colModule') }} - {{ t('import.colModuleHint') }}</li>
              <li><code>cycle</code> / {{ t('import.colCycle') }} - {{ t('import.colCycleHint') }}</li>
              <li><code>estimate</code> / {{ t('import.colEstimate') }} - {{ t('import.colEstimateHint') }}</li>
            </ul>
            <button @click="handleDownloadTemplate" class="mt-3 text-sm text-indigo-600 hover:text-indigo-800 underline">
              {{ t('import.downloadTemplate') }}
            </button>
          </div>

          <div class="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-indigo-400 transition-colors cursor-pointer"
               @click="triggerFileInput"
               @dragover.prevent
               @drop.prevent="handleDrop">
            <input ref="fileInput" type="file" accept=".csv" class="hidden" @change="handleFileChange">
            <div v-if="!selectedFile">
              <div class="text-4xl mb-3">📄</div>
              <p class="text-sm text-gray-600 mb-2">{{ t('import.csvDropHint') }}</p>
              <p class="text-xs text-gray-400">{{ t('import.csvSupport') }}</p>
            </div>
            <div v-else class="flex items-center justify-center space-x-3">
              <span class="text-2xl">📎</span>
              <div class="text-left">
                <p class="text-sm font-medium text-gray-800">{{ selectedFile.name }}</p>
                <p class="text-xs text-gray-500">{{ formatFileSize(selectedFile.size) }}</p>
              </div>
              <button @click.stop="clearFile" class="text-red-500 hover:text-red-700 text-sm ml-2">{{ t('import.remove') }}</button>
            </div>
          </div>
        </div>

        <div v-else class="space-y-4">
          <div class="bg-green-50 border border-green-200 rounded-md p-4">
            <h4 class="text-sm font-medium text-green-800 mb-2">{{ t('import.jsonFormat') }}</h4>
            <pre class="text-xs text-green-700 bg-green-100/50 p-3 rounded overflow-x-auto">[
  {
    "name": "{{ t('import.exampleTitle') }}",
    "description": "{{ t('import.exampleDescription') }}",
    "priority": "high",
    "state_name": "{{ t('import.exampleState') }}",
    "type_name": "{{ t('import.exampleType') }}",
    "assignee_emails": ["user@example.com"],
    "label_names": ["bug"],
    "start_date": "2024-01-01",
    "target_date": "2024-01-31",
    "parent_title": "{{ t('import.exampleParentTitle') }}"
  }
]</pre>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('import.jsonPaste') }}</label>
            <textarea
              v-model="jsonText"
              rows="10"
              class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              :placeholder='`[{"name":"${t("import.examplePlaceholderTitle")}","priority":"high"},{"name":"${t("import.examplePlaceholderTitle2")}"}]`'
            ></textarea>
          </div>
        </div>

        <!-- Progress Bar -->
        <div v-if="importing" class="mt-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-sm text-gray-600">{{ t('import.importing') }}</span>
            <span class="text-sm text-indigo-600">{{ importProgress }}%</span>
          </div>
          <div class="w-full bg-gray-200 rounded-full h-2">
            <div class="bg-indigo-600 h-2 rounded-full transition-all duration-300" :style="{ width: importProgress + '%' }"></div>
          </div>
          <p class="text-xs text-gray-500 mt-1">{{ t('import.pleaseWait') }}</p>
        </div>

        <div v-if="importResult" class="mt-6">
          <div :class="['rounded-md p-4', importResult.fail_count > 0 ? 'bg-yellow-50 border border-yellow-200' : 'bg-green-50 border border-green-200']">
            <div class="flex items-center justify-between mb-2">
              <h4 :class="['text-sm font-medium', importResult.fail_count > 0 ? 'text-yellow-800' : 'text-green-800']">
                {{ t('import.result') }}
              </h4>
              <div class="text-sm">
                <span class="text-green-600 font-medium">{{ t('import.success') }} {{ importResult.success_count }}</span>
                <span v-if="importResult.fail_count > 0" class="text-red-600 font-medium ml-3">{{ t('import.fail') }} {{ importResult.fail_count }}</span>
              </div>
            </div>
            <div v-if="importResult.errors.length > 0" class="mt-3 max-h-40 overflow-y-auto">
              <ul class="space-y-1">
                <li v-for="(err, idx) in importResult.errors" :key="idx" class="text-xs text-red-700 flex items-start">
                  <span class="mr-2">{{ t('import.row') }} {{ err.row }}:</span>
                  <span>{{ err.title || t('import.noTitle') }} - {{ err.message }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-end space-x-3 px-6 py-4 border-t border-gray-200">
        <button
          @click="handleClose"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
        >
          {{ t('import.cancel') }}
        </button>
        <button
          @click="handleImport"
          :disabled="importing || !canImport"
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ importing ? t('import.importing') : t('import.startImport') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import issueApi from '@/api/issue'
import { useI18n } from '@/composables/useI18n'
import type { ImportResult } from '@/api/issue'

const props = defineProps<{
  visible: boolean
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success', result: ImportResult): void
}>()

const { t } = useI18n()

const mode = ref<'csv' | 'json'>('csv')
const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const jsonText = ref('')
const importing = ref(false)
const importProgress = ref(0)
const importResult = ref<ImportResult | null>(null)

const canImport = computed(() => {
  if (mode.value === 'csv') return !!selectedFile.value
  return jsonText.value.trim().length > 0
})

function handleClose() {
  if (!importing.value) {
    importResult.value = null
    selectedFile.value = null
    jsonText.value = ''
    emit('close')
  }
}

function triggerFileInput() {
  fileInput.value?.click()
}

function handleFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectedFile.value = input.files[0]
    importResult.value = null
  }
}

function handleDrop(e: DragEvent) {
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    const file = e.dataTransfer.files[0]
    if (file.name.endsWith('.csv')) {
      selectedFile.value = file
      importResult.value = null
    }
  }
}

function clearFile() {
  selectedFile.value = null
  importResult.value = null
  if (fileInput.value) fileInput.value.value = ''
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

async function handleDownloadTemplate() {
  try {
    await issueApi.downloadImportTemplate()
  } catch (e) {
    console.error('Failed to download template:', e)
  }
}

async function handleImport() {
  importing.value = true
  importProgress.value = 0
  importResult.value = null
  
  // Simulate progress
  const progressInterval = setInterval(() => {
    if (importProgress.value < 90) {
      importProgress.value += 10
    }
  }, 200)
  
  try {
    let result: ImportResult
    if (mode.value === 'csv' && selectedFile.value) {
      result = await issueApi.importIssuesCSV(props.projectId, props.workspaceId, selectedFile.value)
    } else {
      const items = JSON.parse(jsonText.value)
      result = await issueApi.importIssuesJSON(props.projectId, props.workspaceId, items)
    }
    importProgress.value = 100
    importResult.value = result
    if (result.success_count > 0) {
      emit('success', result)
    }
  } catch (e: any) {
    importProgress.value = 100
    importResult.value = {
      success_count: 0,
      fail_count: 1,
      errors: [{ row: 0, title: '', message: e.response?.data?.message || t('import.importFailed') }],
      imported_ids: []
    }
  } finally {
    clearInterval(progressInterval)
    importing.value = false
  }
}
</script>
