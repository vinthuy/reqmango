/**
 * Notification Types - 通知类型定义
 */

export interface Notification {
  id: string
  title: string
  message: string
  type: 'info' | 'warning' | 'error' | 'success'
  priority: 'low' | 'medium' | 'high' | 'urgent'
  is_read: boolean
  read_at?: string
  action_url?: string
  recipient_id: string
  sender_id?: string
  project_id?: string
  issue_id?: string
  created_at: string
  updated_at: string
}

export interface NotificationCreate {
  title: string
  message: string
  type?: 'info' | 'warning' | 'error' | 'success'
  priority?: 'low' | 'medium' | 'high' | 'urgent'
  action_url?: string
  recipient_id: string
  sender_id?: string
  project_id?: string
  issue_id?: string
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
