<template>
  <div class="comment-list">
    <!-- Loading -->
    <div v-if="loading" class="text-center py-8">
      <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
      </svg>
    </div>

    <!-- Comment input -->
    <div v-else class="mb-6">
      <div class="flex items-start gap-3">
        <div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0"
          :style="{ backgroundColor: avatarColor(currentUserId) }"
        >{{ currentUserInitial }}</div>
        <div class="flex-1">
          <textarea
            ref="commentInput"
            v-model="newComment"
            :placeholder="t('comment.placeholder')"
            rows="3"
            class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 resize-none transition-shadow"
            @keydown.ctrl.enter="submitComment"
            @keydown.meta.enter="submitComment"
          ></textarea>
          <div class="flex items-center justify-between mt-2">
            <span class="text-[11px] text-gray-400">{{ t('comment.shortcutHint') }}</span>
            <button
              @click="submitComment"
              :disabled="!newComment.trim() || submitting"
              class="px-4 py-1.5 bg-indigo-600 text-white text-xs font-medium rounded-md hover:bg-indigo-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              {{ submitting ? t('comment.publishing') : t('comment.publish') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="!loading && comments.length === 0" class="text-center py-10">
      <svg class="h-10 w-10 text-gray-300 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
      </svg>
      <p class="mt-2 text-sm text-gray-400">{{ t('comment.noComments') }}</p>
    </div>

    <!-- Comment list -->
    <div v-else class="space-y-4">
      <div
        v-for="comment in comments"
        :key="comment.id"
        class="flex items-start gap-3 group"
      >
        <!-- Avatar -->
        <div class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0"
          :style="{ backgroundColor: avatarColor(comment.author_id) }"
          :title="comment.author?.display_name || t('comment.user')"
        >{{ getInitial(comment.author?.display_name || t('comment.user')) }}</div>

        <!-- Comment body -->
        <div class="flex-1 min-w-0">
          <div class="bg-gray-50 rounded-lg px-4 py-3">
            <!-- Header row -->
            <div class="flex items-center justify-between mb-1.5">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-gray-800">
                  {{ comment.author?.display_name || comment.author?.username || t('comment.user') }}
                </span>
                <span v-if="comment.is_resolved" class="text-[10px] px-1.5 py-0.5 rounded-full bg-green-100 text-green-700 font-medium">
                  {{ t('comment.resolved') }}
                </span>
                <span v-if="isEdited(comment)" class="text-[10px] text-gray-400">(edited)</span>
              </div>
              <span
                class="text-[11px] text-gray-400 shrink-0"
                :title="formatFullDate(comment.created_at)"
              >{{ formatRelativeTime(comment.created_at) }}</span>
            </div>

            <!-- View mode: rendered body -->
            <div v-if="editingId !== comment.id" class="text-sm text-gray-700 whitespace-pre-wrap break-words" v-html="renderBody(comment.body || comment.content)"></div>

            <!-- Edit mode: textarea -->
            <div v-else class="mt-1">
              <textarea
                v-model="editText"
                rows="3"
                class="w-full px-3 py-2 border border-indigo-300 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 resize-none"
                @keydown.ctrl.enter="saveEdit(comment)"
                @keydown.meta.enter="saveEdit(comment)"
                @keydown.escape="cancelEdit"
              ></textarea>
              <div class="flex items-center justify-end gap-2 mt-2">
                <button @click="cancelEdit" class="px-3 py-1 text-xs border border-gray-300 rounded-md hover:bg-gray-50 transition-colors">
                  {{ t('comment.cancel') }}
                </button>
                <button
                  @click="saveEdit(comment)"
                  :disabled="!editText.trim() || editSaving"
                  class="px-3 py-1 text-xs bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50 transition-colors"
                >
                  {{ editSaving ? t('comment.saving') : t('comment.save') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Action bar -->
          <div class="flex items-center gap-3 mt-1.5 px-1">
            <button
              v-if="!comment.is_resolved"
              @click="resolveComment(comment)"
              class="text-[11px] text-gray-400 hover:text-green-600 transition-colors"
            >{{ t('comment.markResolved') }}</button>

            <button
              @click="startReply(comment)"
              class="text-[11px] text-gray-400 hover:text-indigo-600 transition-colors"
            >{{ replyingTo?.id === comment.id ? t('comment.cancelReply') : t('comment.reply') }}</button>

            <button
              v-if="canEdit(comment)"
              @click="startEdit(comment)"
              class="text-[11px] text-gray-400 hover:text-indigo-600 transition-colors opacity-0 group-hover:opacity-100"
            >{{ t('comment.edit') }}</button>

            <button
              v-if="canDelete(comment)"
              @click="deleteComment(comment)"
              class="text-[11px] text-gray-400 hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100"
            >{{ t('comment.delete') }}</button>
          </div>

          <!-- Reply input -->
          <div v-if="replyingTo?.id === comment.id" class="mt-2 flex items-start gap-2">
            <div class="w-6 h-6 rounded-full flex items-center justify-center text-white text-[9px] font-bold shrink-0"
              :style="{ backgroundColor: avatarColor(currentUserId) }"
            >{{ currentUserInitial }}</div>
            <div class="flex-1">
              <textarea
                v-model="replyText"
                :placeholder="t('comment.replyPlaceholder')"
                rows="2"
                class="w-full px-3 py-1.5 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 resize-none"
                @keydown.ctrl.enter="submitReply"
                @keydown.meta.enter="submitReply"
                @keydown.escape="cancelReply"
              ></textarea>
              <div class="flex justify-end gap-2 mt-2">
                <button @click="cancelReply" class="px-3 py-1 text-xs border border-gray-300 rounded-md hover:bg-gray-50">{{ t('comment.cancel') }}</button>
                <button @click="submitReply" :disabled="!replyText.trim()" class="px-3 py-1 text-xs bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50">{{ t('comment.reply') }}</button>
              </div>
            </div>
          </div>

          <!-- Replies (nested) -->
          <div v-if="comment.replies && comment.replies.length > 0" class="mt-2 ml-6 pl-4 border-l-2 border-gray-100 space-y-3">
            <div
              v-for="reply in comment.replies"
              :key="reply.id"
              class="flex items-start gap-2"
            >
              <div class="w-6 h-6 rounded-full flex items-center justify-center text-white text-[9px] font-bold shrink-0"
                :style="{ backgroundColor: avatarColor(reply.author_id) }"
              >{{ getInitial(reply.author?.display_name || t('comment.user')) }}</div>
              <div class="flex-1 bg-gray-50 rounded-lg px-3 py-2">
                <div class="flex items-center justify-between mb-0.5">
                  <span class="text-xs font-semibold text-gray-800">{{ reply.author?.display_name || t('comment.user') }}</span>
                  <span
                    class="text-[10px] text-gray-400"
                    :title="formatFullDate(reply.created_at)"
                  >{{ formatRelativeTime(reply.created_at) }}</span>
                </div>
                <div class="text-xs text-gray-700 whitespace-pre-wrap break-words" v-html="renderBody(reply.body || reply.content)"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Load more -->
    <div v-if="hasMore" class="text-center mt-4">
      <button
        @click="loadMore"
        :disabled="loadingMore"
        class="text-sm text-indigo-600 hover:text-indigo-800 disabled:opacity-50 font-medium"
      >
        {{ loadingMore ? t('comment.loading') : t('comment.loadMore') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import commentApi from '@/api/comment'
import { useConfirm } from '@/composables/useConfirm'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from '@/composables/useI18n'
import type { Comment, CommentCreate } from '@/types/comment'

const AVATAR_COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316']

const props = defineProps<{ issueId: number; isAdmin?: boolean }>()

const { confirm } = useConfirm()
const authStore = useAuthStore()
const { t } = useI18n()

const comments = ref<Comment[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const submitting = ref(false)
const newComment = ref('')
const replyText = ref('')
const replyingTo = ref<Comment | null>(null)
const currentUserId = ref<number | null>(authStore.user?.id || null)
const page = ref(1)
const hasMore = ref(false)

// Edit state
const editingId = ref<number | null>(null)
const editText = ref('')
const editSaving = ref(false)

const currentUserInitial = computed(() => {
  const name = authStore.user?.display_name || '?'
  return name.charAt(0).toUpperCase()
})

// --- Data loading ---
async function loadComments() {
  loading.value = true
  try {
    const response = await commentApi.listIssueComments(props.issueId, page.value)
    const items = response.comments || response.items || []
    // Build reply nesting
    const roots: Comment[] = []
    const replies: Comment[] = []
    for (const c of items) {
      if (c.parent_id) replies.push(c)
      else roots.push(c)
    }
    for (const r of replies) {
      const parent = roots.find(p => p.id === r.parent_id)
      if (parent) {
        if (!parent.replies) parent.replies = []
        parent.replies.push(r)
      } else {
        roots.push(r)
      }
    }
    if (page.value === 1) {
      comments.value = roots
    } else {
      comments.value.push(...roots)
    }
    hasMore.value = items.length >= 20
  } catch (error) {
    console.error('Failed to load comments:', error)
  } finally { loading.value = false }
}

async function loadMore() {
  page.value++
  loadingMore.value = true
  await loadComments()
  loadingMore.value = false
}

// --- Submit ---
async function submitComment() {
  if (!newComment.value.trim()) return
  submitting.value = true
  try {
    const data: CommentCreate = { issue_id: props.issueId, body: newComment.value.trim() } as any
    const comment = await commentApi.createComment(data)
    comments.value.unshift(comment)
    newComment.value = ''
  } catch (error) { console.error('Failed to create comment:', error) }
  finally { submitting.value = false }
}

// --- Edit ---
function startEdit(comment: Comment) {
  editingId.value = comment.id
  editText.value = comment.body || comment.content || ''
}
function cancelEdit() {
  editingId.value = null
  editText.value = ''
}
async function saveEdit(comment: Comment) {
  if (!editText.value.trim()) return
  editSaving.value = true
  try {
    const updated = await commentApi.updateComment(comment.id, { body: editText.value.trim() })
    const idx = comments.value.findIndex(c => c.id === comment.id)
    if (idx !== -1) comments.value[idx] = updated
    editingId.value = null
  } catch (error) { console.error('Failed to update comment:', error) }
  finally { editSaving.value = false }
}
function canEdit(comment: Comment): boolean {
  if (props.isAdmin) return true
  return comment.author_id === currentUserId.value
}
function isEdited(comment: Comment): boolean {
  if (!comment.updated_at || !comment.created_at) return false
  return new Date(comment.updated_at).getTime() - new Date(comment.created_at).getTime() > 5000
}

// --- Resolve ---
async function resolveComment(comment: Comment) {
  try {
    const resolved = await commentApi.resolveComment(comment.id)
    const idx = comments.value.findIndex(c => c.id === comment.id)
    if (idx !== -1) comments.value[idx] = resolved
  } catch (error) { console.error('Failed to resolve comment:', error) }
}

// --- Reply ---
function startReply(comment: Comment) {
  if (replyingTo.value?.id === comment.id) { cancelReply(); return }
  replyingTo.value = comment
  replyText.value = ''
}
function cancelReply() { replyingTo.value = null; replyText.value = '' }
async function submitReply() {
  if (!replyText.value.trim() || !replyingTo.value) return
  submitting.value = true
  try {
    const data: CommentCreate = { issue_id: props.issueId, body: replyText.value.trim(), parent_id: replyingTo.value.id } as any
    const reply = await commentApi.createComment(data)
    const parent = comments.value.find(c => c.id === replyingTo.value!.id)
    if (parent) {
      if (!parent.replies) parent.replies = []
      parent.replies.push(reply)
    }
    cancelReply()
  } catch (error) { console.error('Failed to submit reply:', error) }
  finally { submitting.value = false }
}

// --- Delete ---
async function deleteComment(comment: Comment) {
  if (!(await confirm(t('comment.confirmDelete')))) return
  try {
    await commentApi.deleteComment(comment.id)
    comments.value = comments.value.filter(c => c.id !== comment.id)
  } catch (error) { console.error('Failed to delete comment:', error) }
}
function canDelete(comment: Comment): boolean {
  if (props.isAdmin) return true
  return comment.author_id === currentUserId.value
}

// --- Rendering ---
function avatarColor(id: number | null): string {
  if (!id) return '#94a3b8'
  return AVATAR_COLORS[id % AVATAR_COLORS.length]
}
function getInitial(name: string): string {
  return (name || '?').charAt(0).toUpperCase()
}

function renderBody(text: string | undefined): string {
  if (!text) return ''
  let html = escapeHtml(text)
  // Inline code
  html = html.replace(/`([^`]+)`/g, '<code class="bg-gray-200 text-red-600 px-1 py-0.5 rounded text-xs font-mono">$1</code>')
  // Bold
  html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  // Italic
  html = html.replace(/\*([^*]+)\*/g, '<em>$1</em>')
  // Links
  html = html.replace(/(https?:\/\/\S+)/g, '<a href="$1" target="_blank" rel="noopener" class="text-indigo-600 hover:underline">$1</a>')
  // @mentions
  html = html.replace(/@AI\s+\w[\w\s]*/g, '<span class="ai-mention px-1.5 py-0.5 rounded bg-gradient-to-r from-indigo-100 to-purple-100 text-indigo-700 font-medium">🤖 $&</span>')
  html = html.replace(/(^|\s)@(\w+)/g, '$1<span class="text-indigo-600 font-medium bg-indigo-50 px-1 rounded">@$2</span>')
  return html
}

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

// --- Time ---
function formatRelativeTime(timeStr: string): string {
  const date = new Date(timeStr).getTime()
  const now = Date.now()
  const diff = now - date
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  if (minutes < 1) return t('comment.justNow')
  if (minutes < 60) return t('comment.minutesAgo', { minutes })
  if (hours < 24) return t('comment.hoursAgo', { hours })
  if (days < 7) return t('comment.daysAgo', { days })
  return new Date(timeStr).toLocaleDateString()
}
function formatFullDate(timeStr: string): string {
  return new Date(timeStr).toLocaleString()
}

onMounted(() => loadComments())
</script>

<style scoped>
:deep(code) {
  font-family: ui-monospace, 'Cascadia Code', monospace;
}
:deep(.ai-mention) {
  display: inline;
}
</style>
