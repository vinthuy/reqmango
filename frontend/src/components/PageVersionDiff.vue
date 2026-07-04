<template>
  <div class="page-version-diff">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-semibold text-gray-800">{{ t('pages.versionDiff') }}</h3>
      <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">✕</button>
    </div>
    
    <div class="flex gap-4 mb-4">
      <div class="flex-1">
        <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('pages.oldVersion') }}</label>
        <select v-model="oldVersionNumber" class="w-full px-3 py-2 border border-gray-300 rounded-lg">
          <option v-for="v in versions" :key="v.id" :value="v.version_number">
            v{{ v.version_number }} - {{ formatDate(v.created_at) }}
          </option>
        </select>
      </div>
      <div class="flex-1">
        <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('pages.newVersion') }}</label>
        <select v-model="newVersionNumber" class="w-full px-3 py-2 border border-gray-300 rounded-lg">
          <option v-for="v in versions" :key="v.id" :value="v.version_number">
            v{{ v.version_number }} - {{ formatDate(v.created_at) }}
          </option>
        </select>
      </div>
    </div>
    
    <div v-if="loading" class="flex items-center justify-center py-8">
      <div class="animate-spin h-6 w-6 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
    </div>
    
    <div v-else-if="diffLines.length > 0" class="border border-gray-200 rounded-lg overflow-hidden">
      <div class="bg-gray-50 px-4 py-2 border-b border-gray-200">
        <span class="text-sm text-gray-600">
          {{ t('pages.changes') }}: 
          <span class="text-green-600">+{{ addedCount }}</span>
          <span class="text-red-600 ml-2">-{{ removedCount }}</span>
        </span>
      </div>
      <div class="font-mono text-sm overflow-auto max-h-[500px]">
        <div v-for="(line, idx) in diffLines" :key="idx" 
             class="px-4 py-0.5 border-b border-gray-100"
             :class="{
               'bg-green-50': line.type === 'added',
               'bg-red-50': line.type === 'removed',
               'bg-gray-50': line.type === 'unchanged'
             }">
          <span class="inline-block w-8 text-gray-400 text-right mr-3">{{ line.lineNum || '' }}</span>
          <span v-if="line.type === 'added'" class="text-green-600">+ </span>
          <span v-else-if="line.type === 'removed'" class="text-red-600">- </span>
          <span v-else class="text-gray-400">  </span>
          <span>{{ line.content }}</span>
        </div>
      </div>
    </div>
    
    <div v-else class="text-center py-8 text-gray-500">
      {{ t('pages.noDiff') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import * as pageApi from '@/api/page'
import type { PageVersion } from '@/types/page'
import { useI18n } from '@/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  pageId: number
  projectId: number
}>()

defineEmits<{
  (e: 'close'): void
}>()

const versions = ref<PageVersion[]>([])
const oldVersionNumber = ref<number | null>(null)
const newVersionNumber = ref<number | null>(null)
const oldContent = ref('')
const newContent = ref('')
const loading = ref(false)

interface DiffLine {
  type: 'added' | 'removed' | 'unchanged'
  content: string
  lineNum?: number
}

const diffLines = computed(() => {
  const oldLines = oldContent.value.split('\n')
  const newLines = newContent.value.split('\n')
  const result: DiffLine[] = []
  
  // Simple line-by-line diff
  const maxLen = Math.max(oldLines.length, newLines.length)
  let lineNum = 1
  
  for (let i = 0; i < maxLen; i++) {
    const oldLine = oldLines[i]
    const newLine = newLines[i]
    
    if (oldLine === undefined) {
      result.push({ type: 'added', content: newLine, lineNum: lineNum++ })
    } else if (newLine === undefined) {
      result.push({ type: 'removed', content: oldLine })
    } else if (oldLine !== newLine) {
      result.push({ type: 'removed', content: oldLine })
      result.push({ type: 'added', content: newLine, lineNum: lineNum++ })
    } else {
      result.push({ type: 'unchanged', content: oldLine, lineNum: lineNum++ })
    }
  }
  
  return result
})

const addedCount = computed(() => diffLines.value.filter(l => l.type === 'added').length)
const removedCount = computed(() => diffLines.value.filter(l => l.type === 'removed').length)

onMounted(async () => {
  try {
    versions.value = await pageApi.getPageVersions(props.projectId, props.pageId)
    if (versions.value.length >= 2) {
      oldVersionNumber.value = versions.value[1].version_number
      newVersionNumber.value = versions.value[0].version_number
    } else if (versions.value.length === 1) {
      newVersionNumber.value = versions.value[0].version_number
    }
  } catch (e) {
    console.error('Failed to load versions:', e)
  }
})

watch([oldVersionNumber, newVersionNumber], async () => {
  if (!oldVersionNumber.value || !newVersionNumber.value) return

  loading.value = true
  try {
    const [oldVer, newVer] = await Promise.all([
      pageApi.getPageVersion(props.projectId, props.pageId, oldVersionNumber.value),
      pageApi.getPageVersion(props.projectId, props.pageId, newVersionNumber.value)
    ])
    oldContent.value = oldVer.content || ''
    newContent.value = newVer.content || ''
  } catch (e) {
    console.error('Failed to load versions:', e)
  } finally {
    loading.value = false
  }
})

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}
</script>

<style scoped>
.page-version-diff {
  background: white;
  border-radius: 0.5rem;
  padding: 1rem;
}
</style>
