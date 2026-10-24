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
            @click="switchMode('csv')"
            :class="['px-4 py-2 text-sm font-medium rounded-md', mode === 'csv' ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200']"
          >
            {{ t('import.csvImport') }}
          </button>
          <button
            @click="switchMode('json')"
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

          <div v-if="previewing" class="text-sm text-gray-500">{{ t('import.previewing') }}</div>
          <div v-if="previewError" class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">
            {{ previewError }}
          </div>

          <!-- Mapping stays available even when auto-detection failed, so a
               file with unrecognised headers can still be imported. -->
          <div v-if="selectedFile && headerOptions.length">
            <h4 class="text-sm font-medium text-gray-700">{{ t('import.mapping') }}</h4>
            <p class="text-xs text-gray-500 mb-2">{{ t('import.mappingHint') }}</p>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <label v-for="field in IMPORT_FIELDS" :key="field" class="flex items-center gap-2 text-sm">
                <span class="w-32 shrink-0 text-gray-600">{{ fieldLabel(field) }}</span>
                <select
                  :value="mapping[field] ?? ''"
                  :data-testid="`import-map-${field}`"
                  class="flex-1 px-2 py-1 border border-gray-300 rounded text-sm"
                  @change="onMappingChange(field, ($event.target as HTMLSelectElement).value)"
                >
                  <option value="">{{ t('import.unmapped') }}</option>
                  <option v-for="h in headerOptions" :key="h" :value="h">{{ h }}</option>
                </select>
              </label>
            </div>
          </div>

          <div v-if="preview" class="space-y-4">
            <div class="flex flex-wrap items-center gap-3 text-sm">
              <span class="font-medium text-gray-700">
                {{ t('import.previewSummary', { total: preview.total_rows ?? 0, ok: preview.success_count }) }}
              </span>
              <span v-if="preview.fail_count > 0" class="text-red-600">
                {{ t('import.previewErrors', { count: preview.fail_count }) }}
              </span>
              <span v-if="(preview.warnings?.length ?? 0) > 0" class="text-amber-600">
                {{ t('import.previewWarnings', { count: preview.warnings?.length ?? 0 }) }}
              </span>
            </div>

            <div v-if="fileExcerpt" class="bg-gray-50 border border-gray-200 rounded-md p-3">
              <h4 class="text-xs font-medium text-gray-600 mb-1">{{ t('import.fileExcerpt') }}</h4>
              <pre class="text-xs text-gray-600 whitespace-pre-wrap break-all max-h-32 overflow-y-auto">{{ fileExcerpt }}</pre>
            </div>

            <div v-if="preview.errors.length" class="bg-red-50 border border-red-200 rounded-md p-3">
              <ul class="space-y-1 max-h-32 overflow-y-auto">
                <li v-for="(err, idx) in preview.errors" :key="idx" class="text-xs text-red-700">
                  {{ t('import.row') }} {{ err.row }}: {{ err.title || t('import.noTitle') }} - {{ err.message }}
                </li>
              </ul>
            </div>

            <div v-if="preview.warnings?.length" class="bg-amber-50 border border-amber-200 rounded-md p-3">
              <h4 class="text-xs font-medium text-amber-800 mb-1">{{ t('import.warnings') }}</h4>
              <ul class="space-y-1 max-h-32 overflow-y-auto">
                <li v-for="(w, idx) in preview.warnings" :key="idx" class="text-xs text-amber-800">
                  {{ t('import.row') }} {{ w.row }}: {{ w.message }}
                </li>
              </ul>
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
            <button
              @click="runJsonPreview"
              :disabled="previewing || !jsonText.trim()"
              class="mt-2 px-3 py-1.5 text-sm border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              {{ t('import.preview') }}
            </button>
            <div v-if="previewing" class="mt-2 text-sm text-gray-500">{{ t('import.previewing') }}</div>
            <div v-else-if="previewError" class="mt-2 bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">
              {{ previewError }}
            </div>
            <div v-else-if="preview" class="mt-2 bg-gray-50 border border-gray-200 rounded-md p-3">
              <div class="flex flex-wrap items-center gap-3 text-sm">
                <span class="font-medium text-gray-700">
                  {{ t('import.previewSummary', { total: preview.total_rows ?? 0, ok: preview.success_count }) }}
                </span>
                <span v-if="preview.fail_count > 0" class="text-red-600">
                  {{ t('import.previewErrors', { count: preview.fail_count }) }}
                </span>
                <span v-if="(preview.warnings?.length ?? 0) > 0" class="text-amber-600">
                  {{ t('import.previewWarnings', { count: preview.warnings?.length ?? 0 }) }}
                </span>
              </div>
              <ul v-if="preview.errors.length" class="mt-2 space-y-1 max-h-32 overflow-y-auto">
                <li v-for="(err, idx) in preview.errors" :key="idx" class="text-xs text-red-700">
                  {{ t('import.row') }} {{ err.row }}: {{ err.title || t('import.noTitle') }} - {{ err.message }}
                </li>
              </ul>
              <ul v-if="preview.warnings?.length" class="mt-2 space-y-1 max-h-32 overflow-y-auto">
                <li v-for="(w, idx) in preview.warnings" :key="idx" class="text-xs text-amber-800">
                  {{ t('import.row') }} {{ w.row }}: {{ w.message }}
                </li>
              </ul>
            </div>
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
import issueApi, { IMPORT_FIELDS } from '@/api/issue'
import { useI18n } from '@/composables/useI18n'
import type { ImportField, ImportResult } from '@/api/issue'

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
const previewing = ref(false)
const previewError = ref('')
const preview = ref<ImportResult | null>(null)
const mapping = ref<Record<string, string>>({})
const fileExcerpt = ref('')
// Headers parsed from the file itself. The preview response also carries them,
// but it is unavailable when auto-detection fails and the mapping UI must
// still work then.
const fileHeaders = ref<string[]>([])
const headerOptions = computed(() => preview.value?.headers?.length ? preview.value.headers : fileHeaders.value)

const canImport = computed(() => {
  if (mode.value === 'csv') return !!selectedFile.value && !previewing.value && !!mapping.value.name
  return jsonText.value.trim().length > 0
})

function fieldLabel(field: ImportField): string {
  const labels: Record<ImportField, string> = {
    name: t('import.fieldName'),
    description: t('import.fieldDescription'),
    priority: t('import.fieldPriority'),
    state: t('import.fieldState'),
    type: t('import.fieldType'),
    assignees: t('import.fieldAssignees'),
    labels: t('import.fieldLabels'),
    start_date: t('import.fieldStartDate'),
    target_date: t('import.fieldTargetDate'),
    parent_title: t('import.fieldParentTitle'),
    module: t('import.fieldModule'),
    cycle: t('import.fieldCycle'),
    estimate: t('import.fieldEstimate'),
  }
  return labels[field]
}

function resetPreview() {
  preview.value = null
  previewError.value = ''
  mapping.value = {}
  fileExcerpt.value = ''
  fileHeaders.value = []
}

/**
 * Ask the server to validate the file. The first pass lets the server detect
 * the mapping; later passes send the mapping back verbatim so what the user
 * sees is exactly what gets imported.
 */
async function runPreview(withMapping: boolean) {
  if (mode.value !== 'csv' || !selectedFile.value) return
  previewing.value = true
  previewError.value = ''
  try {
    const result = await issueApi.previewImportCSV(
      props.projectId,
      props.workspaceId,
      selectedFile.value,
      withMapping ? mapping.value : undefined
    )
    preview.value = result
    mapping.value = { ...(result.mapping || {}) }
  } catch (e: any) {
    preview.value = null
    // Keep what the user picked when a re-preview fails, otherwise every failed
    // attempt would wipe their mapping and force them to start over.
    if (!withMapping) mapping.value = {}
    previewError.value = e.response?.data?.message || t('import.importFailed')
  } finally {
    previewing.value = false
  }
}

function onMappingChange(field: ImportField, header: string) {
  const next = { ...mapping.value }
  if (header) next[field] = header
  else delete next[field]
  mapping.value = next
  void runPreview(true)
}

function switchMode(next: 'csv' | 'json') {
  if (mode.value === next) return
  mode.value = next
  importResult.value = null
  resetPreview()
  if (next === 'csv' && selectedFile.value) void runPreview(false)
}

async function runJsonPreview() {
  previewError.value = ''
  preview.value = null
  let parsed: unknown
  try {
    parsed = JSON.parse(jsonText.value)
  } catch {
    previewError.value = t('import.invalidJson')
    return
  }
  if (!Array.isArray(parsed) || parsed.length === 0) {
    previewError.value = t('import.invalidJson')
    return
  }
  previewing.value = true
  try {
    preview.value = await issueApi.previewImportJSON(props.projectId, props.workspaceId, parsed)
  } catch (e: any) {
    previewError.value = e.response?.data?.message || t('import.importFailed')
  } finally {
    previewing.value = false
  }
}

async function loadFileExcerpt(file: File) {
  try {
    const text = await file.text()
    const lines = text.replace(/^\uFEFF/, '').split(/\r?\n/)
    fileExcerpt.value = lines.slice(0, 4).join('\n')
    fileHeaders.value = splitCSVLine(lines[0] || '').map(h => h.trim()).filter(Boolean)
  } catch {
    fileExcerpt.value = ''
    fileHeaders.value = []
  }
}

/** Minimal RFC4180 field split, enough to read a header row with quotes. */
function splitCSVLine(line: string): string[] {
  const out: string[] = []
  let cur = ''
  let quoted = false
  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (quoted) {
      if (ch === '"' && line[i + 1] === '"') {
        cur += '"'
        i++
      } else if (ch === '"') {
        quoted = false
      } else {
        cur += ch
      }
    } else if (ch === '"') {
      quoted = true
    } else if (ch === ',') {
      out.push(cur)
      cur = ''
    } else {
      cur += ch
    }
  }
  out.push(cur)
  return out
}

function handleClose() {
  if (!importing.value) {
    importResult.value = null
    selectedFile.value = null
    jsonText.value = ''
    resetPreview()
    emit('close')
  }
}

function triggerFileInput() {
  fileInput.value?.click()
}

function handleFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectFile(input.files[0])
  }
}

function handleDrop(e: DragEvent) {
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    const file = e.dataTransfer.files[0]
    if (file.name.endsWith('.csv')) {
      selectFile(file)
    }
  }
}

function selectFile(file: File) {
  selectedFile.value = file
  importResult.value = null
  resetPreview()
  void loadFileExcerpt(file)
  void runPreview(false)
}

function clearFile() {
  selectedFile.value = null
  importResult.value = null
  resetPreview()
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
      result = await issueApi.importIssuesCSV(props.projectId, props.workspaceId, selectedFile.value, {
        mapping: mapping.value,
      })
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
