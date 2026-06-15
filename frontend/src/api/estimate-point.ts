/**
 * Estimate Point API - 估算点API模块
 */
import api from './index'
import type { EstimatePoint, EstimatePointCreate, EstimatePointUpdate, EstimatePointBulkCreate, EstimatePointReorder } from '@/types/estimate-point'

const BASE_URL = '/api/v1/projects'

/**
 * 创建估算点
 */
export async function createEstimatePoint(
  projectId: string,
  data: EstimatePointCreate
): Promise<EstimatePoint> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points`,
    data
  )
  return response.data
}

/**
 * 列出项目的估算点
 */
export async function listEstimatePoints(
  projectId: string
): Promise<EstimatePoint[]> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points`
  )
  return response.data
}

/**
 * 获取默认估算点
 */
export async function getDefaultEstimatePoint(
  projectId: string
): Promise<EstimatePoint> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points/default`
  )
  return response.data
}

/**
 * 获取估算点详情
 */
export async function getEstimatePoint(
  projectId: string,
  pointId: string
): Promise<EstimatePoint> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points/${pointId}`
  )
  return response.data
}

/**
 * 更新估算点
 */
export async function updateEstimatePoint(
  projectId: string,
  pointId: string,
  data: EstimatePointUpdate
): Promise<EstimatePoint> {
  const response = await api.patch(
    `${BASE_URL}/${projectId}/estimate-points/${pointId}`,
    data
  )
  return response.data
}

/**
 * 删除估算点
 */
export async function deleteEstimatePoint(
  projectId: string,
  pointId: string
): Promise<void> {
  await api.delete(
    `${BASE_URL}/${projectId}/estimate-points/${pointId}`
  )
}

/**
 * 重新排序估算点
 */
export async function reorderEstimatePoints(
  projectId: string,
  data: EstimatePointReorder
): Promise<EstimatePoint[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points/reorder`,
    data
  )
  return response.data
}

/**
 * 批量创建估算点
 */
export async function bulkCreateEstimatePoints(
  projectId: string,
  data: EstimatePointBulkCreate
): Promise<EstimatePoint[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points/bulk`,
    data
  )
  return response.data
}

/**
 * 创建默认估算点
 */
export async function createDefaultEstimatePoints(
  projectId: string
): Promise<EstimatePoint[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points/defaults`
  )
  return response.data
}

export default {
  createEstimatePoint,
  listEstimatePoints,
  getDefaultEstimatePoint,
  getEstimatePoint,
  updateEstimatePoint,
  deleteEstimatePoint,
  reorderEstimatePoints,
  bulkCreateEstimatePoints,
  createDefaultEstimatePoints
}
