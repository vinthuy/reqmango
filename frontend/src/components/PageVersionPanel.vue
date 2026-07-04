<template>
  <div class="version-panel">
    <div class="version-panel-header">
      <h3 class="text-base font-semibold text-gray-900">{{ t('pages.versionHistory') }}</h3>
      <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600 text-lg leading-none">&times;</button>
    </div>

    <div v-if="loading" class="flex justify-center py-8">
      <div class="animate-spin h-6 w-6 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
    </div>

    <div v-else-if="versions.length === 0" class="text-center text-gray-400 py-8 text-sm">
      {{ t('pages.noVersions') }}
    </div>

    <div v-else class="version-list">
      <div
        v-for="v in versions"
        :key="v.id"
        class="version-item"
        :class="{ 'is-selected': selectedVersion === v.version_number }"
        @click="previewVersion(v)"
      >
        <div class="version-item-main">
          <span class="version-number">v{{ v.version_number }}</span>
          <span class="version-meta">
            <span class="version-author">{{ v.created_by_name || 'Unknown' }}</span>
            <span class="version-date">{{ formatDate(v.created_at) }}</span>
          </span>
        </div>
        <button
          v-if="selectedVersion === v.version_number"
          @click.stop="restoreVersion(v.version_number)"
          class="restore-btn"
          :disabled="restoring"
        >
          {{ restoring ? t('pages.restoring') : t('pages.restore') }}
        </button>
      </div>
    </div>

    <!-- Preview -->
    <div v-if="previewContent !== null" class="version-preview border-t border-gray-200 p-3">
      <div class="text-xs text-gray-500 mb-2">{{ t('pages.previewVersion', { num: selectedVersion ?? 0 }) }}</div>
      <div class="preview-content prose prose-sm max-w-none" v-html="previewContent"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { PageVersion } from '@/types/page'
import * as pageApi from '@/api/page'
import { useI18n } from '@/composables/useI18n'
import { useToast } from '@/composables/useToast'

const { t } = useI18n()
const toast = useToast()

const props = defineProps<{
  projectId: number
  pageId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'restored'): void
}>()

const versions = ref<PageVersion[]>([])
const loading = ref(true)
const selectedVersion = ref<number | null>(null)
const previewContent = ref<string | null>(null)
const restoring = ref(false)

onMounted(() => loadVersions())

async function loadVersions() {
  loading.value = true
  try {
    versions.value = await pageApi.getPageVersions(props.projectId, props.pageId)
  } catch (e) {
    console.error('Failed to load versions:', e)
  } finally {
    loading.value = false
  }
}

async function previewVersion(v: PageVersion) {
  selectedVersion.value = v.version_number
  if (v.content_json) {
    try {
      // Validate JSON parseable, but use v.content for display
      JSON.parse(typeof v.content_json === 'string' ? v.content_json : JSON.stringify(v.content_json))
      previewContent.value = v.content || `<p>${t('pages.versionContentAvailable')}</p>`
    } catch {
      previewContent.value = v.content
    }
  } else {
    previewContent.value = v.content
  }
}

async function restoreVersion(versionNumber: number) {
  if (!confirm(t('pages.restoreConfirm', { num: versionNumber }))) return
  restoring.value = true
  try {
    await pageApi.restorePageVersion(props.projectId, props.pageId, versionNumber)
    emit('restored')
  } catch (e) {
    console.error('Failed to restore version:', e)
    toast.error(t('pages.restoreFailed'))
  } finally {
    restoring.value = false
  }
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleString()
}
</script>

<style scoped>
.version-panel {
  background: #fff;
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
  max-height: 500px;
  display: flex;
  flex-direction: column;
}

.version-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.version-list {
  overflow-y: auto;
  flex: 1;
}

.version-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  cursor: pointer;
  transition: background 0.15s;
  border-bottom: 1px solid #f3f4f6;
}

.version-item:hover {
  background: #f9fafb;
}

.version-item.is-selected {
  background: #eef2ff;
}

.version-item-main {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.version-number {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #4f46e5;
}

.version-meta {
  font-size: 0.75rem;
  color: #9ca3af;
  display: flex;
  gap: 0.5rem;
}

.restore-btn {
  font-size: 0.75rem;
  padding: 0.25rem 0.625rem;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 0.25rem;
  cursor: pointer;
}

.restore-btn:hover { background: #4338ca; }
.restore-btn:disabled { opacity: 0.5; }

.version-preview {
  max-height: 200px;
  overflow-y: auto;
  background: #fafbfc;
}

.preview-content {
  font-size: 0.8125rem;
  color: #4b5563;
  line-height: 1.6;
}
</style>
