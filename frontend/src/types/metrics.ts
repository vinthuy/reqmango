export interface MetricTemplate {
  id: string
  category: string
  name: string
  description: string
  chart_type: string
  default_x_axis: string
  default_y_axis: string
  default_filters?: Record<string, any>
  default_config?: Record<string, any>
  icon: string
}

export interface TemplateCategory {
  id: string
  name: string
  templates: MetricTemplate[]
}

export interface ReferenceLine {
  type: 'constant' | 'max' | 'min' | 'average' | 'median' | 'sum'
  value?: number
  label?: string
  color?: string
  style?: string
}

export interface MetricChartConfig {
  stack_mode?: string
  reference_lines?: ReferenceLine[]
  show_labels?: boolean
  dual_y_axis?: boolean
  legend_position?: string
}

export interface MetricChart {
  id: number
  project_id: number
  creator_id: number
  name: string
  template_id: string
  chart_type: string
  x_axis: string
  y_axis: string
  filters: string
  config: string
  sort_order: number
  is_visible: boolean
  created_at: string
  updated_at: string
  // Preview-only fields (not persisted)
  data_labels?: string[]
  data_values?: number[]
  data_colors?: string[]
}

export interface RenderResult {
  labels: string[]
  values: number[]
  total: number
  colors: Record<string, string>
  reference_lines: Array<{
    type: string
    value: number
    label: string
    color: string
    style: string
  }>
  chart_type: string
  config: MetricChartConfig
}

export interface CreateChartPayload {
  name: string
  template_id?: string
  chart_type: string
  x_axis: string
  y_axis: string
  filters?: Record<string, any>
  config?: MetricChartConfig
  sort_order?: number
}
