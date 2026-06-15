<template>
  <div class="notification-center">
    <!-- 通知图标 -->
    <div class="relative">
      <button
        @click="togglePanel"
        class="relative p-2 text-gray-600 hover:text-gray-900 transition-colors"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        <!-- 未读计数 -->
        <span
          v-if="unreadCount > 0"
          class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center"
        >
          {{ unreadCount > 9 ? '9+' : unreadCount }}
        </span>
      </button>

      <!-- 通知面板 -->
      <div
        v-if="isOpen"
        class="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border border-gray-200 z-50"
      >
        <!-- 头部 -->
        <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
          <h3 class="font-semibold text-gray-900">通知</h3>
          <div class="flex items-center space-x-2">
            <button
              v-if="unreadCount > 0"
              @click="markAllRead"
              class="text-xs text-indigo-600 hover:text-indigo-800"
            >
              全部已读
            </button>
            <button
              @click="togglePanel"
              class="text-gray-400 hover:text-gray-600"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 通知列表 -->
        <div class="max-h-96 overflow-y-auto">
          <!-- 加载状态 -->
          <div v-if="loading" class="p-4 text-center">
            <svg class="animate-spin h-6 w-6 text-indigo-600 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <!-- 空状态 -->
          <div v-else-if="notifications.length === 0" class="p-8 text-center">
            <svg class="h-10 w-10 text-gray-400 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <p class="mt-2 text-gray-500 text-sm">暂无通知</p>
          </div>

          <!-- 通知列表 -->
          <div v-else>
            <div
              v-for="notif in notifications"
              :key="notif.id"
              class="px-4 py-3 hover:bg-gray-50 border-b border-gray-100 cursor-pointer transition-colors"
              :class="{ 'bg-indigo-50': !notif.is_read }"
              @click="handleNotificationClick(notif)"
            >
              <div class="flex items-start space-x-3">
                <!-- 图标 -->
                <div
                  class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0"
                  :class="getNotificationIconClass(notif.type)"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                  </svg>
                </div>

                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-gray-900 truncate">{{ notif.title }}</p>
                  <p class="text-xs text-gray-500 mt-0.5 line-clamp-2">{{ notif.message }}</p>
                  <p class="text-xs text-gray-400 mt-1">{{ formatTime(notif.created_at) }}</p>
                </div>

                <!-- 未读标记 -->
                <div v-if="!notif.is_read" class="w-2 h-2 bg-indigo-600 rounded-full flex-shrink-0 mt-2"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- 查看全部 -->
        <div v-if="notifications.length > 0" class="px-4 py-3 border-t border-gray-200 text-center">
          <a href="/notifications" class="text-sm text-indigo-600 hover:text-indigo-800">
            查看全部通知
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import notificationApi from '@/api/notification'
import type { Notification } from '@/types/notification'

// State
const notifications = ref<Notification[]>([])
const loading = ref(false)
const isOpen = ref(false)

// Computed
const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.is_read).length
})

// Methods
function togglePanel() {
  isOpen.value = !isOpen.value
  if (isOpen.value && notifications.value.length === 0) {
    loadNotifications()
  }
}

async function loadNotifications() {
  loading.value = true
  try {
    notifications.value = await notificationApi.listNotifications(false, 10)
  } catch (error) {
    console.error('Failed to load notifications:', error)
  } finally {
    loading.value = false
  }
}

async function markAllRead() {
  try {
    await notificationApi.markAllAsRead()
    notifications.value.forEach(n => {
      n.is_read = true
      n.read_at = new Date().toISOString()
    })
  } catch (error) {
    console.error('Failed to mark all as read:', error)
  }
}

async function handleNotificationClick(notif: Notification) {
  if (!notif.is_read) {
    try {
      await notificationApi.markAsRead(notif.id)
      notif.is_read = true
      notif.read_at = new Date().toISOString()
    } catch (error) {
      console.error('Failed to mark as read:', error)
    }
  }

  if (notif.action_url) {
    window.location.href = notif.action_url
  }
}

function getNotificationIconClass(type: string) {
  const classes: Record<string, string> = {
    info: 'bg-blue-100 text-blue-600',
    warning: 'bg-yellow-100 text-yellow-600',
    error: 'bg-red-100 text-red-600',
    success: 'bg-green-100 text-green-600'
  }
  return classes[type] || classes.info
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
  loadNotifications()
})
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
