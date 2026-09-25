import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import notificationApi from '@/api/notification'
import type { Notification, NotificationSummary } from '@/types/notification'

export const useNotificationStore = defineStore('notification', () => {
  // ==================== State ====================
  const notifications = ref<Notification[]>([])
  const summary = ref<NotificationSummary | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // ==================== Computed ====================
  const unreadCount = computed(() => summary.value?.unread ?? 0)

  // ==================== Fetch ====================
  async function fetchNotifications(unreadOnly: boolean = false, limit: number = 50, offset: number = 0) {
    isLoading.value = true
    error.value = null
    try {
      notifications.value = await notificationApi.listNotifications(unreadOnly, limit, offset)
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    } finally {
      isLoading.value = false
    }
  }

  async function fetchSummary() {
    error.value = null
    try {
      summary.value = await notificationApi.getNotificationSummary()
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Mark As Read ====================
  async function markAsRead(notificationId: number) {
    error.value = null
    try {
      const updated = await notificationApi.markAsRead(notificationId)
      const idx = notifications.value.findIndex(n => n.id === notificationId)
      if (idx !== -1) notifications.value[idx] = updated
      await fetchSummary()
      return updated
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
      return null
    }
  }

  async function markAllAsRead() {
    error.value = null
    try {
      await notificationApi.markAllAsRead()
      notifications.value = notifications.value.map(n => ({ ...n, is_read: true }))
      await fetchSummary()
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Delete ====================
  async function deleteNotification(notificationId: number) {
    error.value = null
    try {
      await notificationApi.deleteNotification(notificationId)
      notifications.value = notifications.value.filter(n => n.id !== notificationId)
      await fetchSummary()
    } catch (e: any) {
      error.value = e.response?.data?.message || e.message
    }
  }

  // ==================== Return ====================
  return {
    notifications, summary, unreadCount, isLoading, error,
    fetchNotifications, fetchSummary,
    markAsRead, markAllAsRead,
    deleteNotification,
  }
})
