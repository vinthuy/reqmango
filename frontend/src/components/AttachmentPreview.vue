<template>
  <Transition name="fade">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80" @click.self="close" @keydown.esc="close">
      <!-- Close button -->
      <button @click="close" class="absolute top-4 right-4 text-white/70 hover:text-white text-2xl z-10">&times;</button>
      <!-- Nav -->
      <button v-if="items.length > 1" @click.stop="prev" class="absolute left-4 text-white/70 hover:text-white text-3xl z-10">&lsaquo;</button>
      <button v-if="items.length > 1" @click.stop="next" class="absolute right-12 text-white/70 hover:text-white text-3xl z-10">&rsaquo;</button>

      <!-- Image preview -->
      <div v-if="current && isImage(current)" class="max-w-[90vw] max-h-[90vh]">
        <img :src="fileURL(current)" :alt="current.name" class="max-w-full max-h-[85vh] object-contain rounded-lg shadow-2xl" />
        <p class="text-white/60 text-sm text-center mt-3">{{ current.name }} ({{ formatSize(current.file_size) }})</p>
      </div>

      <!-- PDF preview -->
      <div v-else-if="current && isPDF(current)" class="w-[90vw] h-[90vh]">
        <iframe :src="fileURL(current)" class="w-full h-full rounded-lg shadow-2xl" />
        <p class="text-white/60 text-sm text-center mt-1">{{ current.name }}</p>
      </div>

      <!-- Other file types -->
      <div v-else-if="current" class="text-center text-white">
        <div class="text-6xl mb-4">{{ fileIcon(current.mime_type) }}</div>
        <p class="text-lg font-medium">{{ current.name }}</p>
        <p class="text-sm text-white/60 mt-1">{{ formatSize(current.file_size) }}</p>
        <a :href="fileURL(current)" target="_blank" class="inline-block mt-4 px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700">Download</a>
      </div>

      <!-- Counter -->
      <div v-if="items.length > 1" class="absolute bottom-4 left-1/2 -translate-x-1/2 text-white/50 text-xs">{{ currentIdx + 1 }} / {{ items.length }}</div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Attachment { id: number; name: string; file_path: string; file_size: number; mime_type: string }

const props = defineProps<{ visible: boolean; items: Attachment[]; initialIdx?: number }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const currentIdx = ref(0)
const itemsArr = computed(() => {
  if (props.initialIdx !== undefined) currentIdx.value = props.initialIdx
  return props.items || []
})

const current = computed(() => itemsArr.value[currentIdx.value] || null)

function fileURL(a: Attachment) {
  if (a.file_path?.startsWith('http')) return a.file_path
  return `/api/uploads/${a.file_path || a.id}`
}

function isImage(a: Attachment) { return a.mime_type?.startsWith('image/') }
function isPDF(a: Attachment) { return a.mime_type === 'application/pdf' }

function fileIcon(mime: string) {
  if (!mime) return '📄'
  if (mime.startsWith('image/')) return '🖼️'
  if (mime.includes('pdf')) return '📕'
  if (mime.includes('zip') || mime.includes('rar')) return '📦'
  if (mime.includes('text') || mime.includes('json')) return '📝'
  return '📎'
}

function formatSize(bytes: number) {
  if (!bytes) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB']; let i = 0
  let s = bytes
  while (s >= 1024 && i < u.length - 1) { s /= 1024; i++ }
  return s.toFixed(1) + ' ' + u[i]
}

function prev() { if (currentIdx.value > 0) currentIdx.value-- }
function next() { if (currentIdx.value < itemsArr.value.length - 1) currentIdx.value++ }
function close() { emit('close') }
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
