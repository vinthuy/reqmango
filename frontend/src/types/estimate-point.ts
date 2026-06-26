export type EstimateMode = 'points' | 'categories' | 'time'

export interface EstimatePoint {
  id: number
  name: string
  value: number
  mode: EstimateMode
  is_default: boolean
  sequence: number
  project_id: number
  workspace_id: number
  created_at: string
  updated_at: string
}

export interface EstimateCategory {
  id: number
  name: string
  mode: EstimateMode
  is_default: boolean
  sequence: number
  project_id: number
  workspace_id: number
  created_at: string
  updated_at: string
}

export interface EstimateTime {
  id: number
  name: string
  minutes: number
  mode: EstimateMode
  is_default: boolean
  sequence: number
  project_id: number
  workspace_id: number
  created_at: string
  updated_at: string
}

export interface ProjectEstimateSettings {
  id: number
  project_id: number
  workspace_id: number
  mode: EstimateMode
  points_enabled: boolean
  categories_enabled: boolean
  time_enabled: boolean
  created_at: string
  updated_at: string
}

export interface EstimatePointCreate {
  name: string
  value: number
  is_default?: boolean
  sequence?: number
}

export interface EstimatePointUpdate {
  name?: string
  value?: number
  is_default?: boolean
  sequence?: number
}

export interface EstimateCategoryCreate {
  name: string
  is_default?: boolean
  sequence?: number
}

export interface EstimateTimeCreate {
  name: string
  minutes: number
  is_default?: boolean
  sequence?: number
}

export interface EstimatePointBulkCreate {
  points: EstimatePointCreate[]
}

export interface EstimatePointReorder {
  point_ids: number[]
}

export const ESTIMATE_POINT_OPTIONS = [
  { name: '0 - 不需要估算', value: 0 },
  { name: '1 - 很小', value: 1 },
  { name: '2 - 小', value: 2 },
  { name: '3 - 中等', value: 3 },
  { name: '5 - 较大', value: 5 },
  { name: '8 - 大', value: 8 },
  { name: '13 - 很大', value: 13 },
  { name: '21 - 巨大', value: 21 }
]

export const ESTIMATE_CATEGORY_OPTIONS = [
  { name: 'XS - 极小', sequence: 1 },
  { name: 'S - 小', sequence: 2 },
  { name: 'M - 中等', sequence: 3 },
  { name: 'L - 大', sequence: 4 },
  { name: 'XL - 极大', sequence: 5 }
]

export const ESTIMATE_TIME_OPTIONS = [
  { name: '15 分钟', minutes: 15, sequence: 1 },
  { name: '30 分钟', minutes: 30, sequence: 2 },
  { name: '1 小时', minutes: 60, sequence: 3 },
  { name: '2 小时', minutes: 120, sequence: 4 },
  { name: '4 小时', minutes: 240, sequence: 5 },
  { name: '1 天', minutes: 480, sequence: 6 },
  { name: '2 天', minutes: 960, sequence: 7 }
]