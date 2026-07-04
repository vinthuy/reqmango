/**
 * useDashboard Composable - 单元测试
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useDashboard } from '@/composables/useDashboard'
import type { Dashboard, DashboardFullResponse, DashboardWidget } from '@/types/dashboard'

// Mock the dashboard API module
vi.mock('@/api/dashboard', () => ({
  listDashboards: vi.fn(),
  createDashboard: vi.fn(),
  updateDashboard: vi.fn(),
  deleteDashboard: vi.fn(),
  duplicateDashboard: vi.fn(),
  getDashboardFull: vi.fn(),
  addWidget: vi.fn(),
  updateWidget: vi.fn(),
  deleteWidget: vi.fn(),
  reorderWidgets: vi.fn(),
}))

import * as api from '@/api/dashboard'

const mockDashboard = (overrides: Partial<Dashboard> = {}): Dashboard => ({
  id: 1,
  name: 'Test Dashboard',
  is_default: false,
  is_shared: false,
  owner_id: 10,
  project_id: 5,
  columns: 12,
  widgets: [],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  ...overrides,
})

const mockWidget = (overrides: Partial<DashboardWidget> = {}): DashboardWidget => ({
  id: 100,
  dashboard_id: 1,
  widget_type: 'number_card',
  title: 'Count',
  config: {},
  position: {},
  sort_order: 0,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  ...overrides,
})

beforeEach(() => {
  vi.clearAllMocks()
})

describe('useDashboard - Initial State', () => {
  it('should initialize with empty state', () => {
    const { dashboards, currentId, editMode, loading, saving, widgetData, hasDashboards } = useDashboard(5)

    expect(dashboards.value).toEqual([])
    expect(currentId.value).toBeNull()
    expect(editMode.value).toBe(false)
    expect(loading.value).toBe(false)
    expect(saving.value).toBe(false)
    expect(widgetData.value).toEqual([])
    expect(hasDashboards.value).toBe(false)
  })

  it('should compute current as null when no dashboard selected', () => {
    const { current } = useDashboard(5)
    expect(current.value).toBeNull()
  })

  it('should compute canEdit based on current dashboard', () => {
    const { canEdit } = useDashboard(5)
    // current is null, so canEdit should be false (requires dashboard to exist)
    expect(canEdit.value).toBe(false)
  })
})

describe('useDashboard - loadDashboards', () => {
  it('should load dashboards and select first one if none selected', async () => {
    const mocks = [mockDashboard({ id: 1, name: 'Dash 1' }), mockDashboard({ id: 2, name: 'Dash 2' })]
    vi.mocked(api.listDashboards).mockResolvedValueOnce(mocks)

    const { dashboards, currentId, loadDashboards } = useDashboard(5)
    await loadDashboards()

    expect(dashboards.value).toEqual(mocks)
    expect(currentId.value).toBe(1)
  })

  it('should prefer default dashboard when loading', async () => {
    const mocks = [
      mockDashboard({ id: 1, name: 'Regular' }),
      mockDashboard({ id: 2, name: 'Default', is_default: true }),
    ]
    vi.mocked(api.listDashboards).mockResolvedValueOnce(mocks)

    const { currentId, loadDashboards } = useDashboard(5)
    await loadDashboards()

    expect(currentId.value).toBe(2)
  })

  it('should select first dashboard if no default', async () => {
    const mocks = [mockDashboard({ id: 5, name: 'Only Dash', is_default: false })]
    vi.mocked(api.listDashboards).mockResolvedValueOnce(mocks)

    const { currentId, loadDashboards } = useDashboard(5)
    await loadDashboards()

    expect(currentId.value).toBe(5)
  })

  it('should handle empty dashboard list', async () => {
    vi.mocked(api.listDashboards).mockResolvedValueOnce([])

    const { dashboards, currentId, loadDashboards } = useDashboard(5)
    await loadDashboards()

    expect(dashboards.value).toEqual([])
    expect(currentId.value).toBeNull()
  })

  it('should handle API errors gracefully', async () => {
    vi.mocked(api.listDashboards).mockRejectedValueOnce(new Error('Network error'))

    const { loadDashboards, loading } = useDashboard(5)
    // Should not throw
    await loadDashboards()
    expect(loading.value).toBe(false)
  })
})

describe('useDashboard - createDashboard', () => {
  it('should create dashboard and set as current', async () => {
    const newDash = mockDashboard({ id: 3, name: 'Created' })
    vi.mocked(api.createDashboard).mockResolvedValueOnce(newDash)

    const { currentId, dashboards, createDashboard } = useDashboard(5)
    const result = await createDashboard('Created')

    expect(result).toEqual(newDash)
    expect(currentId.value).toBe(3)
    expect(dashboards.value).toContainEqual(newDash)
  })
})

describe('useDashboard - updateDashboardMeta', () => {
  it('should update dashboard in local state', async () => {
    const { dashboards, updateDashboardMeta } = useDashboard(5)
    // Set up initial state
    dashboards.value = [mockDashboard({ id: 1, name: 'Old' })]

    const updated = mockDashboard({ id: 1, name: 'Renamed' })
    vi.mocked(api.updateDashboard).mockResolvedValueOnce(updated)

    const result = await updateDashboardMeta(1, { name: 'Renamed' })
    expect(result.name).toBe('Renamed')
    expect(dashboards.value[0].name).toBe('Renamed')
  })
})

describe('useDashboard - deleteCurrentDashboard', () => {
  it('should delete current and fallback to next', async () => {
    vi.mocked(api.deleteDashboard).mockResolvedValueOnce()

    const { dashboards, currentId, deleteCurrentDashboard } = useDashboard(5)
    dashboards.value = [
      mockDashboard({ id: 1, name: 'First' }),
      mockDashboard({ id: 2, name: 'Second' }),
    ]
    currentId.value = 1

    await deleteCurrentDashboard()

    expect(dashboards.value).toHaveLength(1)
    expect(currentId.value).toBe(2)
  })

  it('should set current to null when deleting last dashboard', async () => {
    vi.mocked(api.deleteDashboard).mockResolvedValueOnce()

    const { dashboards, currentId, deleteCurrentDashboard } = useDashboard(5)
    dashboards.value = [mockDashboard({ id: 1 })]
    currentId.value = 1

    await deleteCurrentDashboard()

    expect(dashboards.value).toEqual([])
    expect(currentId.value).toBeNull()
  })
})

describe('useDashboard - duplicateCurrentDashboard', () => {
  it('should duplicate and set as current', async () => {
    const cloned = mockDashboard({ id: 3, name: 'Original (Copy)' })
    vi.mocked(api.duplicateDashboard).mockResolvedValueOnce(cloned)
    vi.mocked(api.getDashboardFull).mockResolvedValueOnce({
      dashboard: cloned,
      widget_data: [],
    } as DashboardFullResponse)

    const { dashboards, currentId, duplicateCurrentDashboard } = useDashboard(5)
    dashboards.value = [mockDashboard({ id: 1, name: 'Original' })]
    currentId.value = 1

    await duplicateCurrentDashboard()

    expect(dashboards.value).toHaveLength(2)
    expect(currentId.value).toBe(3)
  })
})

describe('useDashboard - loadWidgetData', () => {
  it('should load widget data for current dashboard', async () => {
    const dash = mockDashboard({ id: 1, name: 'Main' })
    const full: DashboardFullResponse = {
      dashboard: dash,
      widget_data: [
        { widget_id: 100, data: { value: 42 } },
        { widget_id: 101, data: { labels: ['A'], values: [1] } },
      ],
    }
    vi.mocked(api.getDashboardFull).mockResolvedValueOnce(full)

    const { dashboards, currentId, widgetData, loadWidgetData } = useDashboard(5)
    dashboards.value = [dash]
    currentId.value = 1

    await loadWidgetData()

    expect(widgetData.value).toHaveLength(2)
    expect(widgetData.value[0].widget_id).toBe(100)
  })
})

describe('useDashboard - getWidgetData', () => {
  it('should return widget data by id', () => {
    const { widgetData, getWidgetData } = useDashboard(5)
    widgetData.value = [
      { widget_id: 100, data: { metric: 'total', value: 10 } },
      { widget_id: 200, data: { metric: 'completed', value: 5 } },
    ]

    expect(getWidgetData(100)).toEqual({ metric: 'total', value: 10 })
    expect(getWidgetData(200)).toEqual({ metric: 'completed', value: 5 })
    expect(getWidgetData(999)).toBeNull()
  })
})

describe('useDashboard - Widget CRUD', () => {
  it('should add widget to current dashboard', async () => {
    const newWidget = mockWidget({ id: 300, title: 'New Card' })
    vi.mocked(api.addWidget).mockResolvedValueOnce(newWidget)
    vi.mocked(api.getDashboardFull).mockResolvedValueOnce({
      dashboard: mockDashboard({ id: 1, widgets: [newWidget] }),
      widget_data: [{ widget_id: 300, data: {} }],
    } as DashboardFullResponse)

    const { dashboards, currentId, addWidgetToCurrent } = useDashboard(5)
    dashboards.value = [mockDashboard({ id: 1 })]
    currentId.value = 1

    await addWidgetToCurrent({ widget_type: 'number_card', title: 'New Card' })

    expect(dashboards.value[0].widgets.some((w) => w.id === 300)).toBe(true)
  })

  it('should delete widget from current dashboard', async () => {
    vi.mocked(api.deleteWidget).mockResolvedValueOnce()

    const { dashboards, currentId, widgetData, deleteWidgetOnCurrent } = useDashboard(5)
    const widget = mockWidget({ id: 100 })
    dashboards.value = [mockDashboard({ id: 1, widgets: [widget, mockWidget({ id: 200 })] })]
    currentId.value = 1
    widgetData.value = [{ widget_id: 100, data: {} }, { widget_id: 200, data: {} }]

    await deleteWidgetOnCurrent(100)

    expect(dashboards.value[0].widgets).toHaveLength(1)
    expect(dashboards.value[0].widgets[0].id).toBe(200)
    expect(widgetData.value).toHaveLength(1)
    expect(widgetData.value[0].widget_id).toBe(200)
  })

  it('should reorder widgets', async () => {
    vi.mocked(api.reorderWidgets).mockResolvedValueOnce()

    const { currentId, dashboards, reorderWidgetsOnCurrent } = useDashboard(5)
    dashboards.value = [mockDashboard({ id: 1 })]
    currentId.value = 1

    await reorderWidgetsOnCurrent([30, 10, 20])
    // Should not throw
    expect(api.reorderWidgets).toHaveBeenCalledWith(5, 1, { widget_ids: [30, 10, 20] })
  })
})

describe('useDashboard - selectDashboard', () => {
  it('should switch dashboard and load widget data', async () => {
    const d2 = mockDashboard({ id: 2, name: 'Second' })
    vi.mocked(api.getDashboardFull).mockResolvedValueOnce({
      dashboard: d2,
      widget_data: [],
    } as DashboardFullResponse)

    const { currentId, selectDashboard } = useDashboard(5)
    currentId.value = 1

    await selectDashboard(2)

    expect(currentId.value).toBe(2)
  })
})
