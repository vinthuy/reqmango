/**
 * Estimate Point Types - 估算点类型定义
 */

export interface EstimatePoint {
  id: string
  name: string
  value: number
  is_default: boolean
  sequence: number
  project_id: string
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

export interface EstimatePointBulkCreate {
  points: EstimatePointCreate[]
}

export interface EstimatePointReorder {
  point_ids: string[]
}

// 常用估算点选项
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
