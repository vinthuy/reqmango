/**
 * Dashboard API - 仪表盘面板API模块
 */
import api from './index'
import type {
  Dashboard,
  DashboardCreate,
  DashboardUpdate,
  DashboardWidget,
  WidgetCreate,
  WidgetUpdate,
  ReorderWidgetsRequest,
  DashboardFullResponse,
} from '@/types/dashboard'

// ==================== Dashboard CRUD ====================

export async function listDashboards(projectId: number): Promise<Dashboard[]> {
  const response = await api.get(`/projects/${projectId}/dashboards`)
  return response.data
}

export async function getDashboard(projectId: number, id: number): Promise<Dashboard> {
  const response = await api.get(`/projects/${projectId}/dashboards/${id}`)
  return response.data
}

export async function createDashboard(projectId: number, data: DashboardCreate): Promise<Dashboard> {
  const response = await api.post(`/projects/${projectId}/dashboards`, data)
  return response.data
}

export async function updateDashboard(projectId: number, id: number, data: DashboardUpdate): Promise<Dashboard> {
  const response = await api.put(`/projects/${projectId}/dashboards/${id}`, data)
  return response.data
}

export async function deleteDashboard(projectId: number, id: number): Promise<void> {
  await api.delete(`/projects/${projectId}/dashboards/${id}`)
}

export async function setDefaultDashboard(projectId: number, id: number): Promise<Dashboard> {
  const response = await api.post(`/projects/${projectId}/dashboards/${id}/set-default`)
  return response.data
}

export async function duplicateDashboard(projectId: number, id: number): Promise<Dashboard> {
  const response = await api.post(`/projects/${projectId}/dashboards/${id}/duplicate`)
  return response.data
}

export async function getDashboardFull(projectId: number, id: number): Promise<DashboardFullResponse> {
  const response = await api.get(`/projects/${projectId}/dashboards/${id}/full`)
  return response.data
}

// ==================== Widget CRUD ====================

export async function addWidget(projectId: number, dashboardId: number, data: WidgetCreate): Promise<DashboardWidget> {
  const response = await api.post(`/projects/${projectId}/dashboards/${dashboardId}/widgets`, data)
  return response.data
}

export async function updateWidget(
  projectId: number,
  dashboardId: number,
  widgetId: number,
  data: WidgetUpdate
): Promise<DashboardWidget> {
  const response = await api.put(`/projects/${projectId}/dashboards/${dashboardId}/widgets/${widgetId}`, data)
  return response.data
}

export async function deleteWidget(projectId: number, dashboardId: number, widgetId: number): Promise<void> {
  await api.delete(`/projects/${projectId}/dashboards/${dashboardId}/widgets/${widgetId}`)
}

export async function reorderWidgets(
  projectId: number,
  dashboardId: number,
  data: ReorderWidgetsRequest
): Promise<void> {
  await api.put(`/projects/${projectId}/dashboards/${dashboardId}/widgets/reorder`, data)
}

export default {
  listDashboards,
  getDashboard,
  createDashboard,
  updateDashboard,
  deleteDashboard,
  setDefaultDashboard,
  duplicateDashboard,
  getDashboardFull,
  addWidget,
  updateWidget,
  deleteWidget,
  reorderWidgets,
}
