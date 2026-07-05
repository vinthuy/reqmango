import api from './index'

export const metricsApi = {
  listTemplates: async (projectId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/templates`)
    return res.data
  },
  listCharts: async (projectId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/charts`)
    return res.data
  },
  getChart: async (projectId: number, chartId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/charts/${chartId}`)
    return res.data
  },
  createChart: async (projectId: number, data: any) => {
    const res = await api.post(`/projects/${projectId}/metrics/charts`, data)
    return res.data
  },
  updateChart: async (projectId: number, chartId: number, data: any) => {
    const res = await api.put(`/projects/${projectId}/metrics/charts/${chartId}`, data)
    return res.data
  },
  deleteChart: async (projectId: number, chartId: number) => {
    const res = await api.delete(`/projects/${projectId}/metrics/charts/${chartId}`)
    return res.data
  },
  renderChart: async (projectId: number, chartId: number) => {
    const res = await api.post(`/projects/${projectId}/metrics/charts/${chartId}/render`)
    return res.data
  },
  reorderCharts: async (projectId: number, chartIds: number[]) => {
    const res = await api.post(`/projects/${projectId}/metrics/charts/reorder`, { chart_ids: chartIds })
    return res.data
  },
  previewChart: async (projectId: number, data: any) => {
    const res = await api.post(`/projects/${projectId}/metrics/preview`, data)
    return res.data
  },
  getFilterValues: async (projectId: number) => {
    const res = await api.get(`/projects/${projectId}/metrics/filter-values`)
    return res.data
  },
}
