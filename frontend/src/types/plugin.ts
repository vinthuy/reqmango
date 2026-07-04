/**
 * Plugin system types
 */

export interface PluginInfo {
  slug: string
  name: string
  description: string
  author: string
  version: string
  type: 'webhook' | 'notification' | 'importer' | 'exporter' | 'automation'
  icon_url: string
  entry_point: string
  config_schema: Record<string, any>
  subscribed_events?: string[]
}

export interface Plugin {
  id: number
  name: string
  slug: string
  description: string | null
  author: string
  version: string
  icon_url: string
  type: string
  entry_point: string
  config_schema: Record<string, any>
  config: Record<string, any>
  subscribed_events: string[]
  enabled: boolean
  workspace_id: number
  installed_by_id: number
  created_at: string
  updated_at: string
}

export interface PluginInstallRequest {
  slug: string
  config?: Record<string, any>
  subscribed_events?: string[]
}

export interface PluginUpdateRequest {
  config?: Record<string, any>
  subscribed_events?: string[]
}

export interface PluginEventLog {
  id: number
  plugin_id: number
  event_type: string
  status: 'success' | 'error' | 'skipped'
  request_body: string
  response_body: string
  status_code: number
  duration_ms: number
  created_at: string
}

export const PLUGIN_TYPES: Record<string, string> = {
  webhook: 'Webhook',
  notification: 'Notification',
  importer: 'Import',
  exporter: 'Export',
  automation: 'Automation',
}

export const PLUGIN_TYPE_ICONS: Record<string, string> = {
  webhook: '🔗',
  notification: '🔔',
  importer: '📥',
  exporter: '📤',
  automation: '⚡',
}

export const SYSTEM_EVENTS = [
  { value: 'issue.created', label: 'Issue Created' },
  { value: 'issue.updated', label: 'Issue Updated' },
  { value: 'issue.deleted', label: 'Issue Deleted' },
  { value: 'comment.created', label: 'Comment Created' },
  { value: 'cycle.started', label: 'Cycle Started' },
  { value: 'cycle.ended', label: 'Cycle Ended' },
]
