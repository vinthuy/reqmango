/**
 * Dashboard Types - 仪表盘面板类型定义
 */

export interface Dashboard {
  id: number
  name: string
  description?: string
  is_default: boolean
  is_shared: boolean
  owner_id: number
  project_id: number
  date_from?: string
  date_to?: string
  columns: number
  widgets: DashboardWidget[]
  created_at: string
  updated_at: string
}

export interface DashboardWidget {
  id: number
  dashboard_id: number
  widget_type: WidgetType
  title: string
  description?: string
  config: Record<string, any>
  position: Record<string, any>
  sort_order: number
  created_at: string
  updated_at: string
}

export type WidgetType =
  | 'number_card'
  | 'bar_chart'
  | 'pie_chart'
  | 'doughnut_chart'
  | 'line_chart'
  | 'bubble_chart'
  | 'scatter_chart'
  | 'mixed_chart'
  | 'burndown'
  | 'table'
  | 'recent_list'
  | 'saved_report'

export type WidgetIconMap = Record<WidgetType, string>

export interface DashboardCreate {
  name: string
  description?: string
  is_default?: boolean
  is_shared?: boolean
  date_from?: string
  date_to?: string
  columns?: number
  widgets?: WidgetCreate[]
}

export interface DashboardUpdate {
  name?: string
  description?: string
  is_default?: boolean
  is_shared?: boolean
  date_from?: string
  date_to?: string
  columns?: number
}

export interface WidgetCreate {
  widget_type: WidgetType
  title: string
  description?: string
  config?: Record<string, any>
  position?: Record<string, any>
  sort_order?: number
}

export interface WidgetUpdate {
  title?: string
  description?: string
  config?: Record<string, any>
  position?: Record<string, any>
  sort_order?: number
}

export interface ReorderWidgetsRequest {
  widget_ids: number[]
}

export interface WidgetDataResponse {
  widget_id: number
  data: Record<string, any>
}

export interface DashboardFullResponse {
  dashboard: Dashboard
  widget_data: WidgetDataResponse[]
}

/** Number card data shape */
export interface NumberCardData {
  metric: string
  label: string
  value: number
}

/** Chart data shape (bar/pie/doughnut/line/table) */
export interface ChartData {
  type: string
  chart_type?: string
  labels: string[]
  values: number[]
  colors?: Record<string, string>
  datasets?: any[]
}
