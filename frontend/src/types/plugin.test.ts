/**
 * Plugin types unit tests
 * Validates all type constants and interface shapes
 */
import { describe, it, expect } from 'vitest'
import {
  PLUGIN_TYPES,
  PLUGIN_TYPE_ICONS,
  SYSTEM_EVENTS,
  type PluginInfo,
  type Plugin,
  type PluginInstallRequest,
  type PluginUpdateRequest,
  type PluginEventLog,
} from './plugin'

// ============================================================
// PLUGIN_TYPES & PLUGIN_TYPE_ICONS
// ============================================================
describe('PLUGIN_TYPES', () => {
  it('should map all 5 plugin types', () => {
    expect(Object.keys(PLUGIN_TYPES)).toHaveLength(5)
    expect(PLUGIN_TYPES.webhook).toBe('Webhook')
    expect(PLUGIN_TYPES.notification).toBe('Notification')
    expect(PLUGIN_TYPES.importer).toBe('Import')
    expect(PLUGIN_TYPES.exporter).toBe('Export')
    expect(PLUGIN_TYPES.automation).toBe('Automation')
  })

  it('should have corresponding icons for all types', () => {
    for (const key of Object.keys(PLUGIN_TYPES)) {
      expect(PLUGIN_TYPE_ICONS[key]).toBeDefined()
      expect(typeof PLUGIN_TYPE_ICONS[key]).toBe('string')
      expect(PLUGIN_TYPE_ICONS[key].length).toBeGreaterThan(0)
    }
  })

  it('should have 5 icon entries', () => {
    expect(Object.keys(PLUGIN_TYPE_ICONS)).toHaveLength(5)
  })

  it('should map icons correctly', () => {
    expect(PLUGIN_TYPE_ICONS.webhook).toBe('🔗')
    expect(PLUGIN_TYPE_ICONS.notification).toBe('🔔')
    expect(PLUGIN_TYPE_ICONS.importer).toBe('📥')
    expect(PLUGIN_TYPE_ICONS.exporter).toBe('📤')
    expect(PLUGIN_TYPE_ICONS.automation).toBe('⚡')
  })

  it('should return undefined for unknown type', () => {
    expect(PLUGIN_TYPES.unknown).toBeUndefined()
    expect(PLUGIN_TYPE_ICONS.unknown).toBeUndefined()
  })
})

// ============================================================
// SYSTEM_EVENTS
// ============================================================
describe('SYSTEM_EVENTS', () => {
  it('should have 6 events', () => {
    expect(SYSTEM_EVENTS).toHaveLength(6)
  })

  it('should have value and label for each event', () => {
    for (const event of SYSTEM_EVENTS) {
      expect(event.value).toBeTruthy()
      expect(event.label).toBeTruthy()
      expect(typeof event.value).toBe('string')
      expect(typeof event.label).toBe('string')
    }
  })

  it('should include issue lifecycle events', () => {
    const values = SYSTEM_EVENTS.map(e => e.value)
    expect(values).toContain('issue.created')
    expect(values).toContain('issue.updated')
    expect(values).toContain('issue.deleted')
  })

  it('should include comment event', () => {
    const values = SYSTEM_EVENTS.map(e => e.value)
    expect(values).toContain('comment.created')
  })

  it('should include cycle events', () => {
    const values = SYSTEM_EVENTS.map(e => e.value)
    expect(values).toContain('cycle.started')
    expect(values).toContain('cycle.ended')
  })

  it('should have descriptive labels', () => {
    expect(SYSTEM_EVENTS[0].label).toBe('Issue Created')
    expect(SYSTEM_EVENTS[1].label).toBe('Issue Updated')
    expect(SYSTEM_EVENTS[2].label).toBe('Issue Deleted')
    expect(SYSTEM_EVENTS[3].label).toBe('Comment Created')
    expect(SYSTEM_EVENTS[4].label).toBe('Cycle Started')
    expect(SYSTEM_EVENTS[5].label).toBe('Cycle Ended')
  })
})

// ============================================================
// PluginInfo interface shape validation
// ============================================================
describe('PluginInfo interface', () => {
  const validInfo: PluginInfo = {
    slug: 'outgoing-webhook',
    name: 'Outgoing Webhook',
    description: 'Send webhook requests on events',
    author: 'ReqMango',
    version: '1.0.0',
    type: 'webhook',
    icon_url: '',
    entry_point: 'https://hooks.example.com',
    config_schema: { url: { type: 'string' } },
    subscribed_events: ['issue.created', 'issue.updated'],
  }

  it('should accept a valid PluginInfo object', () => {
    expect(validInfo.slug).toBe('outgoing-webhook')
    expect(validInfo.name).toBe('Outgoing Webhook')
    expect(validInfo.type).toBe('webhook')
  })

  it('should allow empty subscribed_events', () => {
    const info: PluginInfo = {
      ...validInfo,
      subscribed_events: undefined,
    }
    expect(info.subscribed_events).toBeUndefined()
  })

  it('should support all plugin types', () => {
    const types: PluginInfo['type'][] = ['webhook', 'notification', 'importer', 'exporter', 'automation']
    for (const t of types) {
      const info: PluginInfo = { ...validInfo, type: t }
      expect(info.type).toBe(t)
    }
  })

  it('should have required fields', () => {
    const keys: (keyof PluginInfo)[] = [
      'slug', 'name', 'description', 'author', 'version',
      'type', 'icon_url', 'entry_point', 'config_schema',
    ]
    const info = validInfo
    for (const key of keys) {
      expect(info[key]).toBeDefined()
    }
  })
})

// ============================================================
// Plugin interface shape validation
// ============================================================
describe('Plugin interface', () => {
  const validPlugin: Plugin = {
    id: 1,
    name: 'Outgoing Webhook',
    slug: 'outgoing-webhook',
    description: 'Send webhook requests on events',
    author: 'ReqMango',
    version: '1.0.0',
    icon_url: '',
    type: 'webhook',
    entry_point: 'https://hooks.example.com',
    config_schema: {},
    config: { url: 'https://myhook.com' },
    subscribed_events: ['issue.created'],
    enabled: true,
    workspace_id: 1,
    installed_by_id: 1,
    created_at: '2026-07-04T12:00:00Z',
    updated_at: '2026-07-04T12:00:00Z',
  }

  it('should accept a valid Plugin object', () => {
    expect(validPlugin.id).toBe(1)
    expect(validPlugin.enabled).toBe(true)
    expect(validPlugin.workspace_id).toBe(1)
  })

  it('should have all required fields', () => {
    const keys: (keyof Plugin)[] = [
      'id', 'name', 'slug', 'description', 'author', 'version',
      'icon_url', 'type', 'entry_point', 'config_schema', 'config',
      'subscribed_events', 'enabled', 'workspace_id', 'installed_by_id',
      'created_at', 'updated_at',
    ]
    for (const key of keys) {
      expect(validPlugin[key]).toBeDefined()
    }
  })

  it('should support disabled state', () => {
    const disabled: Plugin = { ...validPlugin, enabled: false }
    expect(disabled.enabled).toBe(false)
  })

  it('should allow null description', () => {
    const plugin: Plugin = { ...validPlugin, description: null }
    expect(plugin.description).toBeNull()
  })
})

// ============================================================
// PluginInstallRequest
// ============================================================
describe('PluginInstallRequest', () => {
  it('should require slug', () => {
    const req: PluginInstallRequest = { slug: 'test-plugin' }
    expect(req.slug).toBe('test-plugin')
  })

  it('should optionally include config', () => {
    const req: PluginInstallRequest = {
      slug: 'test-plugin',
      config: { url: 'https://example.com' },
    }
    expect(req.config).toEqual({ url: 'https://example.com' })
  })

  it('should optionally include subscribed_events', () => {
    const req: PluginInstallRequest = {
      slug: 'test-plugin',
      subscribed_events: ['issue.created'],
    }
    expect(req.subscribed_events).toEqual(['issue.created'])
  })
})

// ============================================================
// PluginUpdateRequest
// ============================================================
describe('PluginUpdateRequest', () => {
  it('should allow config only update', () => {
    const req: PluginUpdateRequest = {
      config: { url: 'https://newhook.com' },
    }
    expect(req.config).toEqual({ url: 'https://newhook.com' })
  })

  it('should allow subscribed_events only update', () => {
    const req: PluginUpdateRequest = {
      subscribed_events: ['issue.created', 'issue.updated'],
    }
    expect(req.subscribed_events).toHaveLength(2)
  })

  it('should allow empty object (no updates)', () => {
    const req: PluginUpdateRequest = {}
    expect(req.config).toBeUndefined()
    expect(req.subscribed_events).toBeUndefined()
  })
})

// ============================================================
// PluginEventLog
// ============================================================
describe('PluginEventLog', () => {
  const validLog: PluginEventLog = {
    id: 1,
    plugin_id: 1,
    event_type: 'issue.created',
    status: 'success',
    request_body: '{"test":true}',
    response_body: '{"ok":true}',
    status_code: 200,
    duration_ms: 150,
    created_at: '2026-07-04T12:00:00Z',
  }

  it('should accept a valid log entry', () => {
    expect(validLog.status).toBe('success')
    expect(validLog.status_code).toBe(200)
    expect(validLog.duration_ms).toBe(150)
  })

  it('should support all 3 statuses', () => {
    const statuses: PluginEventLog['status'][] = ['success', 'error', 'skipped']
    for (const status of statuses) {
      const log: PluginEventLog = { ...validLog, status }
      expect(log.status).toBe(status)
    }
  })

  it('should have all required fields', () => {
    const keys: (keyof PluginEventLog)[] = [
      'id', 'plugin_id', 'event_type', 'status',
      'request_body', 'response_body', 'status_code',
      'duration_ms', 'created_at',
    ]
    for (const key of keys) {
      expect(validLog[key]).toBeDefined()
    }
  })
})
