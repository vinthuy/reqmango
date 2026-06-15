/**
 * Notification API - 通知API模块
 */
import api from './index'
import type { Notification, NotificationCreate, NotificationUpdate, NotificationSummary } from '@/types/notification'

const BASE_URL = '/api/v1/notifications'

/**
 * 获取通知列表
 */
export async function listNotifications(
  unreadOnly: boolean = false,
  limit: number = 50,
  offset: number = 0
): Promise<Notification[]> {
  const response = await api.get(BASE_URL, {
    params: { unread_only: unreadOnly, limit, offset }
  })
  return response.data
}

/**
 * 获取通知摘要
 */
export async function getNotificationSummary(): Promise<NotificationSummary> {
  const response = await api.get(`${BASE_URL}/summary`)
  return response.data
}

/**
 * 获取通知详情
 */
export async function getNotification(
  notificationId: string
): Promise<Notification> {
  const response = await api.get(`${BASE_URL}/${notificationId}`)
  return response.data
}

/**
 * 创建通知
 */
export async function createNotification(
  data: NotificationCreate
): Promise<Notification> {
  const response = await api.post(BASE_URL, data)
  return response.data
}

/**
 * 标记通知已读
 */
export async function markAsRead(
  notificationId: string
): Promise<Notification> {
  const response = await api.patch(`${BASE_URL}/${notificationId}/read`)
  return response.data
}

/**
 * 标记所有通知已读
 */
export async function markAllAsRead(): Promise<{ marked_count: number }> {
  const response = await api.post(`${BASE_URL}/read-all`)
  return response.data
}

/**
 * 删除通知
 */
export async function deleteNotification(
  notificationId: string
): Promise<void> {
  await api.delete(`${BASE_URL}/${notificationId}`)
}

/**
 * 批量创建通知
 */
export async function createBulkNotification(
  data: NotificationCreate,
  recipientIds: string[]
): Promise<Notification[]> {
  const response = await api.post(`${BASE_URL}/bulk`, data, {
    params: { recipient_ids: recipientIds }
  })
  return response.data
}

export default {
  listNotifications,
  getNotificationSummary,
  getNotification,
  createNotification,
  markAsRead,
  markAllAsRead,
  deleteNotification,
  createBulkNotification
}
