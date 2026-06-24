<template>
  <div class="comment-list">
    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-8">
      <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>

    <!-- 评论输入 -->
    <div v-else class="mb-6">
      <div class="flex items-start space-x-3">
        <div class="w-8 h-8 bg-indigo-100 rounded-full flex items-center justify-center text-indigo-600 flex-shrink-0">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </div>
        <div class="flex-1">
          <textarea
            v-model="newComment"
            placeholder="添加评论..."
            rows="3"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 resize-none"
            @keydown.ctrl.enter="submitComment"
          ></textarea>
          <div class="flex justify-end mt-2">
            <button
              @click="submitComment"
              :disabled="!newComment.trim() || submitting"
              class="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ submitting ? '发布中...' : '发布' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 评论列表 -->
    <div v-if="!loading && comments.length === 0" class="text-center py-8">
      <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>
      <p class="mt-2 text-gray-500 text-sm">暂无评论</p>
    </div>

    <!-- 评论项 -->
    <div v-else class="space-y-4">
      <div
        v-for="comment in comments"
        :key="comment.id"
        class="flex items-start space-x-3"
      >
        <div class="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center text-gray-600 flex-shrink-0">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </div>

        <div class="flex-1 min-w-0">
          <div class="bg-gray-50 rounded-lg px-4 py-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-gray-900">{{ comment.author?.display_name || '用户' }}</span>
              <span class="text-xs text-gray-500">{{ formatTime(comment.created_at) }}</span>
            </div>
            <div class="mt-1 text-sm text-gray-700 whitespace-pre-wrap">{{ comment.body || comment.content }}</div>
          </div>

          <!-- 操作按钮 -->
          <div class="flex items-center space-x-4 mt-1 pl-2">
            <button
              v-if="!comment.is_resolved"
              @click="resolveComment(comment)"
              class="text-xs text-gray-500 hover:text-indigo-600"
            >
              标记为已解决
            </button>
            <span v-else class="text-xs text-green-600">已解决</span>

            <button
              @click="replyTo(comment)"
              class="text-xs text-gray-500 hover:text-indigo-600"
            >
              回复
            </button>

            <button
              v-if="canDelete(comment)"
              @click="deleteComment(comment)"
              class="text-xs text-gray-500 hover:text-red-600"
            >
              删除
            </button>
          </div>

          <!-- 回复列表 -->
          <div v-if="comment.replies && comment.replies.length > 0" class="mt-3 pl-4 border-l-2 border-gray-200 space-y-3">
            <div
              v-for="reply in comment.replies"
              :key="reply.id"
              class="flex items-start space-x-2"
            >
              <div class="w-6 h-6 bg-gray-100 rounded-full flex items-center justify-center text-gray-600 flex-shrink-0">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>
              <div class="flex-1 bg-gray-50 rounded-lg px-3 py-2">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-medium text-gray-900">{{ reply.author?.display_name || '用户' }}</span>
                  <span class="text-xs text-gray-500">{{ formatTime(reply.created_at) }}</span>
                </div>
                <div class="mt-0.5 text-xs text-gray-700 whitespace-pre-wrap">{{ reply.body || reply.content }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore" class="text-center mt-4">
      <button
        @click="loadMore"
        :disabled="loadingMore"
        class="text-sm text-indigo-600 hover:text-indigo-800 disabled:opacity-50"
      >
        {{ loadingMore ? '加载中...' : '加载更多' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import commentApi from '@/api/comment'
import { useConfirm } from '@/composables/useConfirm'
import { useAuthStore } from '@/stores/auth'
import type { Comment, CommentCreate } from '@/types/comment'

// Props
const props = defineProps<{
  issueId: number
}>()

// State
const { confirm } = useConfirm()
const authStore = useAuthStore()
const comments = ref<Comment[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const submitting = ref(false)
const newComment = ref('')
const currentUserId = ref<number | null>(authStore.user?.id || null)
const page = ref(1)
const hasMore = ref(false)

// Methods
async function loadComments() {
  loading.value = true
  try {
    const response = await commentApi.listIssueComments(props.issueId, page.value)
    const items = response.comments || response.items || []
    if (page.value === 1) {
      comments.value = items
    } else {
      comments.value.push(...items)
    }
    hasMore.value = comments.value.length < response.total
  } catch (error) {
    console.error('Failed to load comments:', error)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  page.value++
  loadingMore.value = true
  await loadComments()
  loadingMore.value = false
}

async function submitComment() {
  if (!newComment.value.trim()) return

  submitting.value = true
  try {
    const data: CommentCreate = {
      issue_id: props.issueId,
      body: newComment.value.trim()
    } as any
    const comment = await commentApi.createComment(data)
    comments.value.unshift(comment)
    newComment.value = ''
  } catch (error) {
    console.error('Failed to create comment:', error)
  } finally {
    submitting.value = false
  }
}

async function resolveComment(comment: Comment) {
  try {
    const resolved = await commentApi.resolveComment(comment.id)
    const index = comments.value.findIndex(c => c.id === comment.id)
    if (index !== -1) {
      comments.value[index] = resolved
    }
  } catch (error) {
    console.error('Failed to resolve comment:', error)
  }
}

function replyTo(comment: Comment) {
  // TODO: 实现回复功能
  console.log('Reply to comment:', comment.id)
}

async function deleteComment(comment: Comment) {
  if (!(await confirm('确定要删除这条评论吗？'))) return

  try {
    await commentApi.deleteComment(comment.id)
    comments.value = comments.value.filter(c => c.id !== comment.id)
  } catch (error) {
    console.error('Failed to delete comment:', error)
  }
}

function canDelete(comment: Comment): boolean {
  // TODO: 实现权限检查
  return comment.author_id === currentUserId.value
}

function formatTime(timeStr: string) {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  return date.toLocaleDateString('zh-CN')
}

// Load on mount
onMounted(() => {
  loadComments()
})
</script>
