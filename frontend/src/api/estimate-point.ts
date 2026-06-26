import api from './index'
import type { 
  EstimatePoint, 
  EstimatePointCreate, 
  EstimatePointUpdate, 
  EstimatePointBulkCreate, 
  EstimatePointReorder,
  EstimateCategory,
  EstimateCategoryCreate,
  EstimateTime,
  EstimateTimeCreate,
  ProjectEstimateSettings,
  EstimateMode
} from '@/types/estimate-point'

const BASE_URL = '/projects'

export async function getEstimateSettings(
  projectId: number
): Promise<ProjectEstimateSettings> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points/settings`
  )
  return response.data
}

export async function updateEstimateSettings(
  projectId: number,
  mode: EstimateMode
): Promise<ProjectEstimateSettings> {
  const response = await api.put(
    `${BASE_URL}/${projectId}/estimate-points/settings`,
    { mode }
  )
  return response.data
}

export async function createEstimatePoint(
  projectId: number,
  data: EstimatePointCreate
): Promise<EstimatePoint> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points`,
    data
  )
  return response.data
}

export async function listEstimatePoints(
  projectId: number
): Promise<EstimatePoint[]> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points`
  )
  return response.data
}

export async function getDefaultEstimatePoint(
  projectId: number
): Promise<EstimatePoint> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points/default`
  )
  return response.data
}

export async function getEstimatePoint(
  projectId: number,
  pointId: number
): Promise<EstimatePoint> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-points/${pointId}`
  )
  return response.data
}

export async function updateEstimatePoint(
  projectId: number,
  pointId: number,
  data: EstimatePointUpdate
): Promise<EstimatePoint> {
  const response = await api.patch(
    `${BASE_URL}/${projectId}/estimate-points/${pointId}`,
    data
  )
  return response.data
}

export async function deleteEstimatePoint(
  projectId: number,
  pointId: number
): Promise<void> {
  await api.delete(
    `${BASE_URL}/${projectId}/estimate-points/${pointId}`
  )
}

export async function reorderEstimatePoints(
  projectId: number,
  data: EstimatePointReorder
): Promise<EstimatePoint[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points/reorder`,
    data
  )
  return response.data
}

export async function bulkCreateEstimatePoints(
  projectId: number,
  data: EstimatePointBulkCreate
): Promise<EstimatePoint[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points/bulk`,
    data
  )
  return response.data
}

export async function createDefaultEstimatePoints(
  projectId: number
): Promise<EstimatePoint[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-points/defaults`
  )
  return response.data
}

export async function listEstimateCategories(
  projectId: number
): Promise<EstimateCategory[]> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-categories`
  )
  return response.data
}

export async function createEstimateCategory(
  projectId: number,
  data: EstimateCategoryCreate
): Promise<EstimateCategory> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-categories`,
    data
  )
  return response.data
}

export async function createDefaultEstimateCategories(
  projectId: number
): Promise<EstimateCategory[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-categories/defaults`
  )
  return response.data
}

export async function listEstimateTime(
  projectId: number
): Promise<EstimateTime[]> {
  const response = await api.get(
    `${BASE_URL}/${projectId}/estimate-time`
  )
  return response.data
}

export async function createEstimateTime(
  projectId: number,
  data: EstimateTimeCreate
): Promise<EstimateTime> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-time`,
    data
  )
  return response.data
}

export async function createDefaultEstimateTime(
  projectId: number
): Promise<EstimateTime[]> {
  const response = await api.post(
    `${BASE_URL}/${projectId}/estimate-time/defaults`
  )
  return response.data
}

export default {
  getEstimateSettings,
  updateEstimateSettings,
  createEstimatePoint,
  listEstimatePoints,
  getDefaultEstimatePoint,
  getEstimatePoint,
  updateEstimatePoint,
  deleteEstimatePoint,
  reorderEstimatePoints,
  bulkCreateEstimatePoints,
  createDefaultEstimatePoints,
  listEstimateCategories,
  createEstimateCategory,
  createDefaultEstimateCategories,
  listEstimateTime,
  createEstimateTime,
  createDefaultEstimateTime
}