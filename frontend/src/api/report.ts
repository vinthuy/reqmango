/**
 * Report API — 报告生成模块
 */
import api from './index'

export const reportApi = {
  generate: async (projectId: number, data: { type: string; filters?: Record<string, any> }): Promise<any> => {
    const res = await api.post(`/projects/${projectId}/reports`, data)
    return res.data
  },
}
