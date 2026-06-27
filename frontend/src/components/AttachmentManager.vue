<template>
  <div class="attachment-manager">
    <!-- 上传区域 -->
    <div
      class="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-indigo-400 transition-colors cursor-pointer"
      :class="{ 'border-indigo-500 bg-indigo-50': isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
      @click="triggerFileInput"
    >
      <input
        ref="fileInput"
        type="file"
        multiple
        class="hidden"
        @change="handleFileSelect"
      />
      <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
      </svg>
      <p class="mt-2 text-sm text-gray-600">
        拖拽文件到此处，或 <span class="text-indigo-600">点击上传</span>
      </p>
      <p class="mt-1 text-xs text-gray-500">
        支持任意文件类型
      </p>
    </div>

    <!-- 上传进度 -->
    <div v-if="uploadingFiles.length > 0" class="mt-4 space-y-2">
      <div
        v-for="file in uploadingFiles"
        :key="file.name"
        class="flex items-center space-x-3 bg-gray-50 rounded-lg px-3 py-2"
      >
        <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <div class="flex-1 min-w-0">
          <p class="text-sm text-gray-700 truncate">{{ file.name }}</p>
          <p class="text-xs text-gray-500">{{ formatSize(file.size) }}</p>
        </div>
        <svg class="animate-spin h-4 w-4 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>
    </div>

    <!-- 附件列表 -->
    <div v-if="attachments.length > 0" class="mt-6">
      <h4 class="text-sm font-medium text-gray-700 mb-3">已上传 ({{ attachments.length }})</h4>
      <div class="space-y-2">
        <div
          v-for="attachment in attachments"
          :key="attachment.id"
          class="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 cursor-pointer"
          @click="previewAttachment(attachment, idx)"
        >
          <div class="flex items-center space-x-3 min-w-0">
            <!-- 文件图标 -->
            <div class="w-10 h-10 bg-indigo-100 rounded-lg flex items-center justify-center flex-shrink-0">
              <svg class="h-5 w-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>

            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-gray-900 truncate">{{ attachment.name }}</p>
              <p class="text-xs text-gray-500">
                {{ formatSize(attachment.file_size) }} · {{ formatTime(attachment.created_at) }}
              </p>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <!-- 下载 -->
            <a
              :href="getDownloadUrl(attachment.id)"
              target="_blank"
              class="p-1.5 text-gray-400 hover:text-indigo-600 rounded"
              title="下载"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
            </a>

            <!-- 删除 -->
            <button
              @click="deleteAttachment(attachment)"
              class="p-1.5 text-gray-400 hover:text-red-600 rounded"
              title="删除"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
  <AttachmentPreview :visible="showPreview" :items="attachments" :initial-idx="previewIdx" @close="showPreview = false" />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import attachmentApi from '@/api/attachment'
import { useConfirm } from '@/composables/useConfirm'
import AttachmentPreview from '@/components/AttachmentPreview.vue'
import type { Attachment } from '@/types/attachment'

// Props
const props = defineProps<{
  issueId?: number
  projectId?: number
}>()

// State
const { confirm } = useConfirm()
const attachments = ref<Attachment[]>([])
const uploadingFiles = ref<File[]>([])
const showPreview = ref(false)
const previewIdx = ref(0)

function previewAttachment(_att: Attachment, idx: number) {
  previewIdx.value = idx; showPreview.value = true
}
const isDragging = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

// Methods
async function loadAttachments() {
  try {
    if (props.issueId) {
      attachments.value = await attachmentApi.listIssueAttachments(props.issueId)
    }
  } catch (error) {
    console.error('Failed to load attachments:', error)
  }
}

function triggerFileInput() {
  fileInput.value?.click()
}

async function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files) {
    await uploadFiles(Array.from(target.files))
  }
}

async function handleDrop(event: DragEvent) {
  isDragging.value = false
  if (event.dataTransfer?.files) {
    await uploadFiles(Array.from(event.dataTransfer.files))
  }
}

async function uploadFiles(files: File[]) {
  if (!props.issueId) return

  for (const file of files) {
    uploadingFiles.value.push(file)
  }

  try {
    for (const file of files) {
      const attachment = await attachmentApi.uploadAttachment(props.issueId!, file)
      attachments.value.unshift(attachment)
    }
  } catch (error) {
    console.error('Failed to upload file:', error)
  } finally {
    uploadingFiles.value = []
  }
}

async function deleteAttachment(attachment: Attachment) {
  if (!(await confirm(`确定要删除 "${attachment.name}" 吗？`))) return
  if (!props.issueId) return

  try {
    await attachmentApi.deleteAttachment(props.issueId, attachment.id)
    attachments.value = attachments.value.filter(a => a.id !== attachment.id)
  } catch (error) {
    console.error('Failed to delete attachment:', error)
  }
}

function getDownloadUrl(attachmentId: number): string {
  return attachmentApi.getAttachmentDownloadUrl(props.issueId || 0, attachmentId)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
}

function formatTime(timeStr: string) {
  const date = new Date(timeStr)
  return date.toLocaleDateString('zh-CN')
}

// Load on mount
onMounted(() => {
  loadAttachments()
})
</script>
