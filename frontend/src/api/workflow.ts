import api from './index'

export async function listWorkflows(projectId: number) { const r = await api.get(`/projects/${projectId}/workflows`); return r.data }
export async function createWorkflow(projectId: number, data: any) { const r = await api.post(`/projects/${projectId}/workflows`, data); return r.data }
export async function updateWorkflow(projectId: number, wid: number, data: any) { const r = await api.put(`/projects/${projectId}/workflows/${wid}`, data); return r.data }
export async function deleteWorkflow(projectId: number, wid: number) { await api.delete(`/projects/${projectId}/workflows/${wid}`) }

export async function addTransition(projectId: number, wid: number, data: any) { const r = await api.post(`/projects/${projectId}/workflows/${wid}/transitions`, data); return r.data }
export async function updateTransition(projectId: number, wid: number, tid: number, data: any) { const r = await api.put(`/projects/${projectId}/workflows/${wid}/transitions/${tid}`, data); return r.data }
export async function deleteTransition(projectId: number, wid: number, tid: number) { await api.delete(`/projects/${projectId}/workflows/${wid}/transitions/${tid}`) }

export async function listAutomations(projectId: number) { const r = await api.get(`/projects/${projectId}/automations`); return r.data }
export async function createAutomation(projectId: number, data: any) { const r = await api.post(`/projects/${projectId}/automations`, data); return r.data }
export async function updateAutomation(projectId: number, aid: number, data: any) { const r = await api.put(`/projects/${projectId}/automations/${aid}`, data); return r.data }
export async function deleteAutomation(projectId: number, aid: number) { await api.delete(`/projects/${projectId}/automations/${aid}`) }

export async function listStateTransitions(projectId: number) { const r = await api.get(`/projects/${projectId}/workflows/transitions`); return r.data }
export async function createStateTransition(projectId: number, _workspaceId: number, data: any) { const r = await api.post(`/projects/${projectId}/workflows/transitions`, data); return r.data }
export async function updateStateTransition(tid: number, data: any) { const r = await api.put(`/projects/0/workflows/transitions/${tid}`, data); return r.data }
export async function deleteStateTransition(tid: number) { await api.delete(`/projects/0/workflows/transitions/${tid}`) }

export async function listAutomationTemplates() { const r = await api.get('/automation-templates'); return r.data }
export async function listAutomationRules(projectId: number) { const r = await api.get(`/projects/${projectId}/automations`); return r.data }
export async function toggleAutomationRule(ruleId: number, enabled: boolean) { const r = await api.patch(`/projects/0/automations/${ruleId}`, { is_enabled: enabled }); return r.data }

const workflowApi = {
  listWorkflows, createWorkflow, updateWorkflow, deleteWorkflow,
  addTransition, updateTransition, deleteTransition,
  listAutomations, createAutomation, updateAutomation, deleteAutomation,
  listStateTransitions, createStateTransition, updateStateTransition, deleteStateTransition,
  listAutomationTemplates, listAutomationRules, toggleAutomationRule
}
export { workflowApi }
export default workflowApi
