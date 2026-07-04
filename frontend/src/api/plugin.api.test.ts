/**
 * Plugin API client unit tests
 * Tests all Plugin API methods with mocked axios
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

// Mock axios before importing
const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()

vi.mock('@/api/index', () => ({
  default: {
    get: (...args: any[]) => mockGet(...args),
    post: (...args: any[]) => mockPost(...args),
    put: (...args: any[]) => mockPut(...args),
    delete: (...args: any[]) => mockDelete(...args),
  },
}))

import { pluginApi } from '@/api/plugin'

const WS_ID = 1
const PLUGIN_ID = 42

describe('pluginApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // ----------------------------------------------------------
  // catalog()
  // ----------------------------------------------------------
  describe('catalog()', () => {
    it('should call GET /workspaces/:ws/plugins/catalog', async () => {
      const mockData = [
        { slug: 'webhook', name: 'Outgoing Webhook', type: 'webhook', version: '1.0.0' },
      ]
      mockGet.mockResolvedValue({ data: mockData })

      const result = await pluginApi.catalog(WS_ID)

      expect(mockGet).toHaveBeenCalledTimes(1)
      expect(mockGet).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/catalog`)
      expect(result).toEqual(mockData)
    })

    it('should handle empty catalog', async () => {
      mockGet.mockResolvedValue({ data: [] })
      const result = await pluginApi.catalog(WS_ID)
      expect(result).toEqual([])
    })

    it('should propagate network errors', async () => {
      mockGet.mockRejectedValue(new Error('Network Error'))
      await expect(pluginApi.catalog(WS_ID)).rejects.toThrow('Network Error')
    })
  })

  // ----------------------------------------------------------
  // list()
  // ----------------------------------------------------------
  describe('list()', () => {
    it('should call GET /workspaces/:ws/plugins', async () => {
      const mockData = [
        { id: 1, slug: 'webhook', name: 'Webhook', enabled: true },
      ]
      mockGet.mockResolvedValue({ data: mockData })

      const result = await pluginApi.list(WS_ID)

      expect(mockGet).toHaveBeenCalledTimes(1)
      expect(mockGet).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins`)
      expect(result).toEqual(mockData)
    })

    it('should return empty array when no plugins installed', async () => {
      mockGet.mockResolvedValue({ data: [] })
      const result = await pluginApi.list(WS_ID)
      expect(result).toEqual([])
    })

    it('should handle API errors', async () => {
      mockGet.mockRejectedValue({ response: { status: 500, data: { message: 'Server Error' } } })
      await expect(pluginApi.list(WS_ID)).rejects.toEqual(expect.objectContaining({ response: expect.anything() }))
    })
  })

  // ----------------------------------------------------------
  // install()
  // ----------------------------------------------------------
  describe('install()', () => {
    it('should call POST /workspaces/:ws/plugins with slug', async () => {
      const mockPlugin = { id: 1, slug: 'webhook', name: 'Webhook', enabled: true }
      mockPost.mockResolvedValue({ data: mockPlugin })

      const result = await pluginApi.install(WS_ID, { slug: 'webhook' })

      expect(mockPost).toHaveBeenCalledTimes(1)
      expect(mockPost).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins`, { slug: 'webhook' })
      expect(result).toEqual(mockPlugin)
    })

    it('should include optional config in install request', async () => {
      mockPost.mockResolvedValue({ data: { id: 2 } })
      await pluginApi.install(WS_ID, { slug: 'webhook', config: { url: 'https://x.com' } })
      expect(mockPost).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins`, {
        slug: 'webhook',
        config: { url: 'https://x.com' },
      })
    })

    it('should include optional subscribed_events', async () => {
      mockPost.mockResolvedValue({ data: { id: 3 } })
      await pluginApi.install(WS_ID, { slug: 'webhook', subscribed_events: ['issue.created'] })
      expect(mockPost).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins`, {
        slug: 'webhook',
        subscribed_events: ['issue.created'],
      })
    })

    it('should handle duplicate install error (409)', async () => {
      mockPost.mockRejectedValue({
        response: { status: 409, data: { message: 'Plugin already installed' } },
      })
      await expect(pluginApi.install(WS_ID, { slug: 'webhook' })).rejects.toEqual(
        expect.objectContaining({ response: expect.objectContaining({ status: 409 }) })
      )
    })
  })

  // ----------------------------------------------------------
  // get()
  // ----------------------------------------------------------
  describe('get()', () => {
    it('should call GET /workspaces/:ws/plugins/:id', async () => {
      const mockPlugin = { id: PLUGIN_ID, slug: 'webhook', name: 'Webhook' }
      mockGet.mockResolvedValue({ data: mockPlugin })

      const result = await pluginApi.get(WS_ID, PLUGIN_ID)

      expect(mockGet).toHaveBeenCalledTimes(1)
      expect(mockGet).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}`)
      expect(result).toEqual(mockPlugin)
    })

    it('should throw on not found (404)', async () => {
      mockGet.mockRejectedValue({ response: { status: 404, data: { message: 'Not found' } } })
      await expect(pluginApi.get(WS_ID, 9999)).rejects.toEqual(
        expect.objectContaining({ response: expect.objectContaining({ status: 404 }) })
      )
    })
  })

  // ----------------------------------------------------------
  // update()
  // ----------------------------------------------------------
  describe('update()', () => {
    it('should call PUT /workspaces/:ws/plugins/:id with config', async () => {
      const mockUpdated = { id: PLUGIN_ID, config: { url: 'https://new.com' } }
      mockPut.mockResolvedValue({ data: mockUpdated })

      const result = await pluginApi.update(WS_ID, PLUGIN_ID, { config: { url: 'https://new.com' } })

      expect(mockPut).toHaveBeenCalledTimes(1)
      expect(mockPut).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}`, {
        config: { url: 'https://new.com' },
      })
      expect(result).toEqual(mockUpdated)
    })

    it('should allow updating subscribed_events', async () => {
      mockPut.mockResolvedValue({ data: {} })
      await pluginApi.update(WS_ID, PLUGIN_ID, { subscribed_events: ['issue.created'] })
      expect(mockPut).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}`, {
        subscribed_events: ['issue.created'],
      })
    })

    it('should allow updating both config and subscribed_events', async () => {
      mockPut.mockResolvedValue({ data: {} })
      await pluginApi.update(WS_ID, PLUGIN_ID, {
        config: { url: 'https://x.com' },
        subscribed_events: ['issue.created', 'issue.updated'],
      })
      expect(mockPut).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}`, {
        config: { url: 'https://x.com' },
        subscribed_events: ['issue.created', 'issue.updated'],
      })
    })
  })

  // ----------------------------------------------------------
  // uninstall()
  // ----------------------------------------------------------
  describe('uninstall()', () => {
    it('should call DELETE /workspaces/:ws/plugins/:id', async () => {
      mockDelete.mockResolvedValue({ data: {} })

      await pluginApi.uninstall(WS_ID, PLUGIN_ID)

      expect(mockDelete).toHaveBeenCalledTimes(1)
      expect(mockDelete).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}`)
    })

    it('should not throw on successful uninstall', async () => {
      mockDelete.mockResolvedValue({ data: {} })
      await expect(pluginApi.uninstall(WS_ID, PLUGIN_ID)).resolves.toBeUndefined()
    })

    it('should propagate error on uninstall failure', async () => {
      mockDelete.mockRejectedValue(new Error('Forbidden'))
      await expect(pluginApi.uninstall(WS_ID, PLUGIN_ID)).rejects.toThrow('Forbidden')
    })
  })

  // ----------------------------------------------------------
  // enable()
  // ----------------------------------------------------------
  describe('enable()', () => {
    it('should call POST /workspaces/:ws/plugins/:id/enable', async () => {
      const mockPlugin = { id: PLUGIN_ID, enabled: true }
      mockPost.mockResolvedValue({ data: mockPlugin })

      const result = await pluginApi.enable(WS_ID, PLUGIN_ID)

      expect(mockPost).toHaveBeenCalledTimes(1)
      expect(mockPost).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}/enable`)
      expect(result).toEqual(mockPlugin)
    })
  })

  // ----------------------------------------------------------
  // disable()
  // ----------------------------------------------------------
  describe('disable()', () => {
    it('should call POST /workspaces/:ws/plugins/:id/disable', async () => {
      const mockPlugin = { id: PLUGIN_ID, enabled: false }
      mockPost.mockResolvedValue({ data: mockPlugin })

      const result = await pluginApi.disable(WS_ID, PLUGIN_ID)

      expect(mockPost).toHaveBeenCalledTimes(1)
      expect(mockPost).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}/disable`)
      expect(result).toEqual(mockPlugin)
    })
  })

  // ----------------------------------------------------------
  // logs()
  // ----------------------------------------------------------
  describe('logs()', () => {
    it('should call GET /workspaces/:ws/plugins/:id/logs with limit param', async () => {
      const mockLogs = [
        { id: 1, event_type: 'issue.created', status: 'success', duration_ms: 100 },
      ]
      mockGet.mockResolvedValue({ data: mockLogs })

      const result = await pluginApi.logs(WS_ID, PLUGIN_ID, 50)

      expect(mockGet).toHaveBeenCalledTimes(1)
      expect(mockGet).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}/logs`, {
        params: { limit: 50 },
      })
      expect(result).toEqual(mockLogs)
    })

    it('should omit limit param when not provided', async () => {
      mockGet.mockResolvedValue({ data: [] })
      await pluginApi.logs(WS_ID, PLUGIN_ID)
      expect(mockGet).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}/logs`, {
        params: undefined,
      })
    })

    it('should return empty array for no logs', async () => {
      mockGet.mockResolvedValue({ data: [] })
      const result = await pluginApi.logs(WS_ID, PLUGIN_ID)
      expect(result).toEqual([])
    })
  })

  // ----------------------------------------------------------
  // test()
  // ----------------------------------------------------------
  describe('test()', () => {
    it('should call POST /workspaces/:ws/plugins/:id/test', async () => {
      const mockResult = { message: 'Test completed', payload: '{"ok":true}' }
      mockPost.mockResolvedValue({ data: mockResult })

      const result = await pluginApi.test(WS_ID, PLUGIN_ID)

      expect(mockPost).toHaveBeenCalledTimes(1)
      expect(mockPost).toHaveBeenCalledWith(`/workspaces/${WS_ID}/plugins/${PLUGIN_ID}/test`)
      expect(result).toEqual(mockResult)
    })

    it('should propagate test errors', async () => {
      mockPost.mockRejectedValue({ response: { status: 502, data: { message: 'Test failed: upstream unreachable' } } })
      await expect(pluginApi.test(WS_ID, PLUGIN_ID)).rejects.toEqual(
        expect.objectContaining({ response: expect.anything() })
      )
    })
  })

  // ----------------------------------------------------------
  // URL correctness across workspace IDs
  // ----------------------------------------------------------
  describe('URL construction', () => {
    it('should handle numeric workspace ID', async () => {
      mockGet.mockResolvedValue({ data: [] })
      await pluginApi.catalog(99)
      expect(mockGet).toHaveBeenCalledWith('/workspaces/99/plugins/catalog')
    })

    it('should handle large workspace ID', async () => {
      mockGet.mockResolvedValue({ data: [] })
      await pluginApi.list(999999)
      expect(mockGet).toHaveBeenCalledWith('/workspaces/999999/plugins')
    })
  })
})
