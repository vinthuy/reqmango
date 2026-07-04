/**
 * Dashboard Types - 单元测试
 */
import { describe, it, expect } from 'vitest'
import type {
  Dashboard,
  DashboardWidget,
  DashboardCreate,
  DashboardUpdate,
  WidgetCreate,
  WidgetUpdate,
  WidgetType,
  DashboardFullResponse,
  WidgetDataResponse,
  NumberCardData,
  ChartData,
} from './dashboard'

describe('Dashboard Types', () => {
  it('should accept a valid Dashboard object', () => {
    const d: Dashboard = {
      id: 1,
      name: 'Sprint Dashboard',
      is_default: true,
      is_shared: false,
      owner_id: 10,
      project_id: 5,
      columns: 12,
      widgets: [],
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-06-01T00:00:00Z',
    }
    expect(d.id).toBe(1)
    expect(d.name).toBe('Sprint Dashboard')
    expect(d.columns).toBe(12)
    expect(d.is_default).toBe(true)
  })

  it('should accept optional description and date fields', () => {
    const d: Dashboard = {
      id: 2,
      name: 'Release Dashboard',
      description: 'Tracks release progress',
      is_default: false,
      is_shared: true,
      owner_id: 10,
      project_id: 5,
      date_from: '2024-01-01',
      date_to: '2024-12-31',
      columns: 6,
      widgets: [],
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(d.description).toBe('Tracks release progress')
    expect(d.date_from).toBe('2024-01-01')
    expect(d.date_to).toBe('2024-12-31')
  })

  it('should handle Dashboard with widgets', () => {
    const widget: DashboardWidget = {
      id: 100,
      dashboard_id: 1,
      widget_type: 'number_card',
      title: 'Total Issues',
      config: { metric: 'total', label: 'Total' },
      position: { x: 0, y: 0, w: 6, h: 2 },
      sort_order: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    const d: Dashboard = {
      id: 1,
      name: 'My Dash',
      is_default: false,
      is_shared: false,
      owner_id: 10,
      project_id: 5,
      columns: 12,
      widgets: [widget],
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    expect(d.widgets).toHaveLength(1)
    expect(d.widgets[0].widget_type).toBe('number_card')
    expect(d.widgets[0].title).toBe('Total Issues')
  })
})

describe('DashboardWidget Types', () => {
  it('should accept all WidgetType values', () => {
    const validTypes: WidgetType[] = [
      'number_card',
      'bar_chart',
      'pie_chart',
      'doughnut_chart',
      'line_chart',
      'burndown',
      'table',
      'recent_list',
    ]

    validTypes.forEach((type) => {
      const w: DashboardWidget = {
        id: 1,
        dashboard_id: 1,
        widget_type: type,
        title: 'Test',
        config: {},
        position: {},
        sort_order: 0,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }
      expect(w.widget_type).toBe(type)
    })
  })

  it('should store JSON config and position', () => {
    const w: DashboardWidget = {
      id: 1,
      dashboard_id: 1,
      widget_type: 'bar_chart',
      title: 'Distribution',
      config: { report_type: 'distribution', group_by: 'state', chart_type: 'bar' },
      position: { x: 0, y: 1, w: 12, h: 4 },
      sort_order: 2,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(w.config.report_type).toBe('distribution')
    expect(w.config.group_by).toBe('state')
    expect(w.position.x).toBe(0)
    expect(w.position.w).toBe(12)
    expect(w.position.h).toBe(4)
    expect(w.sort_order).toBe(2)
  })

  it('should handle optional description', () => {
    const w: DashboardWidget = {
      id: 2,
      dashboard_id: 1,
      widget_type: 'table',
      title: 'Issues Table',
      description: 'Shows all open issues',
      config: {},
      position: {},
      sort_order: 0,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }
    expect(w.description).toBe('Shows all open issues')
  })
})

describe('DashboardCreate Type', () => {
  it('should require only name', () => {
    const dc: DashboardCreate = { name: 'New Dashboard' }
    expect(dc.name).toBe('New Dashboard')
    expect(dc.columns).toBeUndefined()
    expect(dc.is_default).toBeUndefined()
  })

  it('should accept full create payload', () => {
    const dc: DashboardCreate = {
      name: 'Full Dashboard',
      description: 'A complete dashboard',
      is_default: true,
      is_shared: true,
      date_from: '2024-01-01',
      date_to: '2024-06-30',
      columns: 6,
      widgets: [
        { widget_type: 'number_card', title: 'Total', config: { metric: 'total' }, position: { x: 0, y: 0, w: 6, h: 2 } },
      ],
    }
    expect(dc.widgets).toHaveLength(1)
    expect(dc.columns).toBe(6)
    expect(dc.date_from).toBe('2024-01-01')
    expect(dc.date_to).toBe('2024-06-30')
  })
})

describe('DashboardUpdate Type', () => {
  it('should allow partial updates', () => {
    const du: DashboardUpdate = { name: 'Renamed' }
    expect(du.name).toBe('Renamed')
    expect(du.is_default).toBeUndefined()
  })

  it('should accept all optional fields', () => {
    const du: DashboardUpdate = {
      name: 'Updated',
      description: 'New description',
      is_default: true,
      is_shared: false,
      columns: 12,
    }
    expect(du.name).toBe('Updated')
    expect(du.is_default).toBe(true)
    expect(du.is_shared).toBe(false)
    expect(du.columns).toBe(12)
  })
})

describe('WidgetCreate Type', () => {
  it('should require widget_type and title', () => {
    const wc: WidgetCreate = { widget_type: 'bar_chart', title: 'Chart' }
    expect(wc.widget_type).toBe('bar_chart')
    expect(wc.title).toBe('Chart')
  })

  it('should accept optional fields', () => {
    const wc: WidgetCreate = {
      widget_type: 'number_card',
      title: 'Count',
      description: 'Total issue count',
      config: { metric: 'total' },
      position: { x: 0, y: 0, w: 3, h: 2 },
      sort_order: 5,
    }
    expect(wc.sort_order).toBe(5)
  })
})

describe('WidgetUpdate Type', () => {
  it('should allow partial widget updates', () => {
    const wu: WidgetUpdate = { title: 'New Title' }
    expect(wu.title).toBe('New Title')
  })

  it('should accept all widget update fields', () => {
    const wu: WidgetUpdate = {
      title: 'Updated Widget',
      description: 'Updated desc',
      config: { metric: 'completed' },
      position: { x: 0, y: 1, w: 6, h: 3 },
      sort_order: 10,
    }
    expect(wu.sort_order).toBe(10)
  })
})

describe('DashboardFullResponse Type', () => {
  it('should combine dashboard with widget data', () => {
    const widgetData: WidgetDataResponse[] = [
      { widget_id: 100, data: { metric: 'total', value: 42 } },
      { widget_id: 101, data: { type: 'bar', labels: ['A'], values: [10] } },
    ]

    const dashboard: Dashboard = {
      id: 1,
      name: 'Main',
      is_default: true,
      is_shared: false,
      owner_id: 1,
      project_id: 1,
      columns: 12,
      widgets: [],
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    const full: DashboardFullResponse = { dashboard, widget_data: widgetData }
    expect(full.dashboard.name).toBe('Main')
    expect(full.widget_data).toHaveLength(2)
    expect(full.widget_data[0].widget_id).toBe(100)
    expect(full.widget_data[1].widget_id).toBe(101)
  })
})

describe('NumberCardData Shape', () => {
  it('should match metric/value/label shape', () => {
    const card: NumberCardData = { metric: 'total', label: 'Total Issues', value: 128 }
    expect(card.metric).toBe('total')
    expect(card.label).toBe('Total Issues')
    expect(card.value).toBe(128)
  })
})

describe('ChartData Shape', () => {
  it('should match chart data shape', () => {
    const chart: ChartData = {
      type: 'bar',
      chart_type: 'bar',
      labels: ['Open', 'In Progress', 'Done'],
      values: [10, 5, 20],
      colors: { Open: '#ff0000', 'In Progress': '#00ff00', Done: '#0000ff' },
    }
    expect(chart.labels).toHaveLength(3)
    expect(chart.values).toEqual([10, 5, 20])
    expect(chart.type).toBe('bar')
  })

  it('should accept datasets array', () => {
    const chart: ChartData = {
      type: 'line',
      labels: ['Jan', 'Feb'],
      values: [5, 10],
      datasets: [{ label: 'Created', data: [5, 10] }, { label: 'Closed', data: [2, 7] }],
    }
    expect(chart.datasets).toHaveLength(2)
  })
})
