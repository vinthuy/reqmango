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

export async function listWorkspaceAutomations(workspaceId: number) { const r = await api.get(`/workspaces/${workspaceId}/automations`); return r.data }
export async function createWorkspaceAutomation(workspaceId: number, data: any) { const r = await api.post(`/workspaces/${workspaceId}/automations`, data); return r.data }
export async function updateWorkspaceAutomation(workspaceId: number, aid: number, data: any) { const r = await api.put(`/workspaces/${workspaceId}/automations/${aid}`, data); return r.data }
export async function deleteWorkspaceAutomation(workspaceId: number, aid: number) { await api.delete(`/workspaces/${workspaceId}/automations/${aid}`) }

export async function listWorkspaceWorkflows(workspaceId: number) { const r = await api.get(`/workspaces/${workspaceId}/workflows`); return r.data }
export async function createWorkspaceWorkflow(workspaceId: number, data: any) { const r = await api.post(`/workspaces/${workspaceId}/workflows`, data); return r.data }
export async function updateWorkspaceWorkflow(workspaceId: number, wid: number, data: any) { const r = await api.put(`/workspaces/${workspaceId}/workflows/${wid}`, data); return r.data }
export async function deleteWorkspaceWorkflow(workspaceId: number, wid: number) { await api.delete(`/workspaces/${workspaceId}/workflows/${wid}`) }

export async function addWorkspaceTransition(workspaceId: number, wid: number, data: any) { const r = await api.post(`/workspaces/${workspaceId}/workflows/${wid}/transitions`, data); return r.data }
export async function updateWorkspaceTransition(workspaceId: number, wid: number, tid: number, data: any) { const r = await api.put(`/workspaces/${workspaceId}/workflows/${wid}/transitions/${tid}`, data); return r.data }
export async function deleteWorkspaceTransition(workspaceId: number, wid: number, tid: number) { await api.delete(`/workspaces/${workspaceId}/workflows/${wid}/transitions/${tid}`) }

export async function listStateTransitions(projectId: number, wid: number) { const r = await api.get(`/projects/${projectId}/workflows/${wid}/transitions`); return r.data }
export async function createStateTransition(projectId: number, wid: number, data: any) { const r = await api.post(`/projects/${projectId}/workflows/${wid}/transitions`, data); return r.data }
export async function updateStateTransition(projectId: number, wid: number, tid: number, data: any) { const r = await api.put(`/projects/${projectId}/workflows/${wid}/transitions/${tid}`, data); return r.data }
export async function deleteStateTransition(projectId: number, wid: number, tid: number) { await api.delete(`/projects/${projectId}/workflows/${wid}/transitions/${tid}`) }

export async function listAutomationTemplates() { const r = await api.get('/automation-templates'); return r.data }
export async function listAutomationRules(projectId: number) { const r = await api.get(`/projects/${projectId}/automations`); return r.data }
export async function toggleAutomationRule(projectId: number, ruleId: number, enabled: boolean) { const r = await api.put(`/projects/${projectId}/automations/${ruleId}`, { is_enabled: enabled }); return r.data }

const workflowApi = {
  listWorkflows, createWorkflow, updateWorkflow, deleteWorkflow,
  addTransition, updateTransition, deleteTransition,
  listAutomations, createAutomation, updateAutomation, deleteAutomation,
  listWorkspaceAutomations, createWorkspaceAutomation, updateWorkspaceAutomation, deleteWorkspaceAutomation,
  listWorkspaceWorkflows, createWorkspaceWorkflow, updateWorkspaceWorkflow, deleteWorkspaceWorkflow,
  addWorkspaceTransition, updateWorkspaceTransition, deleteWorkspaceTransition,
  listStateTransitions, createStateTransition, updateStateTransition, deleteStateTransition,
  listAutomationTemplates, listAutomationRules, toggleAutomationRule
}
export { workflowApi }
export default workflowApi
