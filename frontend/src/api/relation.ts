import api from './index'

// ---- Relation Types ----
export async function listRelationTypes(workspaceId: number): Promise<any[]> {
  const r = await api.get('/relations/types', { params: { workspace_id: workspaceId } }); return r.data
}
export async function createRelationType(workspaceId: number, data: {name:string,inward_name:string,outward_name:string}): Promise<any> {
  const r = await api.post('/relations/types', data, { params: { workspace_id: workspaceId } }); return r.data
}
export async function updateRelationType(id: number, data: any): Promise<any> {
  const r = await api.put(`/relations/types/${id}`, data); return r.data
}
export async function deleteRelationType(id: number): Promise<void> {
  await api.delete(`/relations/types/${id}`)
}

// ---- Issue Relations ----
export async function listIssueRelations(issueId: number): Promise<any[]> {
  const r = await api.get(`/issues/${issueId}/relations`); return r.data
}
export async function createIssueRelation(issueId: number, data: {related_issue_id:number,relation_type_id:number,comment?:string}): Promise<any> {
  const r = await api.post(`/issues/${issueId}/relations`, data); return r.data
}
export async function deleteIssueRelation(relationId: number): Promise<void> {
  await api.delete(`/relations/${relationId}`)
}

export const relationApi = {
  listRelationTypes, createRelationType, updateRelationType, deleteRelationType,
  listIssueRelations, createIssueRelation, deleteIssueRelation,
}
export default relationApi
