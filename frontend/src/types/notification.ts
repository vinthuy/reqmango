/**
 * Notification Types - 通知类型定义
 */

export type NotificationType = 'info' | 'warning' | 'error' | 'success' | 'assignment' | 'status' | 'comment' | 'reminder'
export type NotificationPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface Notification {
  id: number
  title: string
  message: string
  type: NotificationType
  priority: NotificationPriority
  is_read: boolean
  read_at?: string
  action_url?: string
  recipient_id: number
  sender_id?: number
  project_id?: number
  issue_id?: number
  created_at: string
  updated_at: string
}

export interface NotificationCreate {
  title: string
  message: string
  type?: NotificationType
  priority?: NotificationPriority
  action_url?: string
  recipient_id: number
  sender_id?: number
  project_id?: number
  issue_id?: number
}

export interface NotificationUpdate {
  is_read?: boolean
  title?: string
  message?: string
}

export interface NotificationSummary {
  total: number
  unread: number
  unread_by_type: Record<string, number>
}
