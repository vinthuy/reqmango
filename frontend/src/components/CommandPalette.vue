<template>
  <Transition name="fade">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]" @click.self="close">
      <div class="bg-white rounded-xl shadow-2xl border border-gray-200 w-full max-w-lg overflow-hidden">
        <!-- Search input -->
        <div class="flex items-center px-4 py-3 border-b border-gray-200">
          <span class="text-gray-400 mr-3 text-lg">🔍</span>
          <input
            ref="inputRef"
            v-model="query"
            @keydown="handleKey"
            placeholder="Type a command or search..."
            class="flex-1 text-lg border-none focus:outline-none focus:ring-0"
          />
          <kbd class="text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded">Esc</kbd>
        </div>

        <!-- Results -->
        <div class="max-h-80 overflow-y-auto p-2">
          <div v-if="filteredItems.length === 0" class="text-center text-gray-400 py-8 text-sm">
            No matching commands
          </div>
          <button
            v-for="(item, idx) in filteredItems"
            :key="item.id"
            @click="execute(item)"
            class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-left text-sm transition-colors"
            :class="idx === selectedIdx ? 'bg-indigo-50 text-indigo-700' : 'text-gray-700 hover:bg-gray-50'"
          >
            <span class="text-lg">{{ item.icon }}</span>
            <div class="flex-1">
              <div class="font-medium">{{ item.label }}</div>
              <div class="text-xs text-gray-400">{{ item.description }}</div>
            </div>
            <kbd v-if="item.shortcut" class="text-xs text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded">{{ item.shortcut }}</kbd>
          </button>
        </div>

        <!-- Footer -->
        <div class="px-4 py-2 border-t border-gray-100 text-xs text-gray-400 flex gap-4">
          <span>↑↓ Navigate</span><span>↵ Select</span><span>Esc Close</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps<{
  visible: boolean
  workspaceSlug: string
  projectId: number
  workspaceId: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'navigate', path: string): void
}>()

const router = useRouter()
const query = ref('')
const selectedIdx = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)

interface CommandItem {
  id: string
  label: string
  description: string
  icon: string
  shortcut?: string
  action: () => void
}

const commands = computed<CommandItem[]>(() => {
  const s = props.workspaceSlug
  const pid = props.projectId
  const wid = props.workspaceId
  return [
    { id: 'new-issue', label: 'New Issue', description: 'Create a new work item', icon: '➕', shortcut: 'N', action: () => router.push(`/workspaces/${wid}/projects/${pid}/issues/new`) },
    { id: 'new-bug', label: 'New Bug', description: 'Quick create a bug', icon: '🐛', action: () => router.push(`/workspaces/${wid}/projects/${pid}/issues/new?type=bug`) },
    { id: 'ai-chat', label: 'AI Assistant', description: 'Open AI chat sidebar', icon: '🤖', shortcut: '⌘J', action: () => emit('navigate', '/ai-chat') },
    { id: 'ai-create', label: 'AI Create', description: 'Create issue from natural language', icon: '✨', action: () => emit('navigate', '/ai-create') },
    { id: 'go-issues', label: 'Go to Issues', description: 'View all work items', icon: '📋', action: () => router.push(`/workspace/${s}/project/${pid}`) },
    { id: 'go-pages', label: 'Go to Pages', description: 'View documentation', icon: '📄', action: () => router.push(`/workspace/${s}/project/${pid}/pages`) },
    { id: 'go-settings', label: 'Project Settings', description: 'Configure project', icon: '⚙️', action: () => router.push(`/workspace/${s}/project/${pid}/settings`) },
    { id: 'go-cycles', label: 'Cycles', description: 'View sprint cycles', icon: '🔄', action: () => router.push(`/workspace/${s}/project/${pid}?tab=cycles`) },
    { id: 'go-modules', label: 'Modules', description: 'View modules', icon: '📦', action: () => router.push(`/workspace/${s}/project/${pid}?tab=modules`) },
    { id: 'my-issues', label: 'My Issues', description: 'Issues assigned to me', icon: '👤', action: () => router.push(`/workspace/${s}/project/${pid}?filter=mine`) },
    { id: 'back-workspace', label: 'Back to Workspace', description: 'Return to workspace', icon: '🏠', action: () => router.push(`/workspace/${s}`) },
  ]
})

const filteredItems = computed(() => {
  if (!query.value.trim()) return commands.value
  const q = query.value.toLowerCase()
  return commands.value.filter(c => c.label.toLowerCase().includes(q) || c.description.toLowerCase().includes(q))
})

function handleKey(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') { e.preventDefault(); selectedIdx.value = Math.min(selectedIdx.value + 1, filteredItems.value.length - 1) }
  else if (e.key === 'ArrowUp') { e.preventDefault(); selectedIdx.value = Math.max(selectedIdx.value - 1, 0) }
  else if (e.key === 'Enter') { e.preventDefault(); if (filteredItems.value[selectedIdx.value]) execute(filteredItems.value[selectedIdx.value]) }
  else if (e.key === 'Escape') close()
}

function execute(item: CommandItem) { item.action(); close() }
function close() { query.value = ''; selectedIdx.value = 0; emit('close') }

onMounted(() => {
  document.addEventListener('keydown', globalHandler)
  nextTick(() => inputRef.value?.focus())
})
onUnmounted(() => document.removeEventListener('keydown', globalHandler))

function globalHandler(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    if (props.visible) close()
    else emit('close') // toggle handled by parent
  }
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
