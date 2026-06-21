import api from './index'
import type {
  ProjectTemplate,
  ProjectTemplateCreate,
  ProjectTemplateUpdate,
  ProjectTemplateType,
  TemplateTypeAdd
} from '@/types/template'

export async function listTemplates(workspaceId: number): Promise<ProjectTemplate[]> {
  const response = await api.get('/templates', { params: { workspace_id: workspaceId } })
  return response.data
}

export async function getTemplate(templateId: number): Promise<ProjectTemplate> {
  const response = await api.get(`/templates/${templateId}`)
  return response.data
}

export async function createTemplate(workspaceId: number, data: ProjectTemplateCreate): Promise<ProjectTemplate> {
  const response = await api.post('/templates', data, { params: { workspace_id: workspaceId } })
  return response.data
}

export async function updateTemplate(templateId: number, data: ProjectTemplateUpdate): Promise<ProjectTemplate> {
  const response = await api.put(`/templates/${templateId}`, data)
  return response.data
}

export async function deleteTemplate(templateId: number): Promise<void> {
  await api.delete(`/templates/${templateId}`)
}

export async function addTypeToTemplate(templateId: number, data: TemplateTypeAdd): Promise<ProjectTemplateType> {
  const response = await api.post(`/templates/${templateId}/types`, data)
  return response.data
}

export async function removeTypeFromTemplate(templateId: number, typeId: number): Promise<void> {
  await api.delete(`/templates/${templateId}/types/${typeId}`)
}

export async function applyTemplate(templateId: number, projectId: number): Promise<void> {
  await api.post(`/templates/${templateId}/apply`, { project_id: projectId })
}

// Type Template APIs
export async function listTypeTemplates(workspaceId: number): Promise<any[]> {
  const r = await api.get('/type-templates', { params: { workspace_id: workspaceId } }); return r.data
}
export async function getTypeTemplate(id: number): Promise<any> {
  const r = await api.get(`/type-templates/${id}`); return r.data
}
export async function createTypeTemplate(workspaceId: number, data: any): Promise<any> {
  const r = await api.post('/type-templates', data, { params: { workspace_id: workspaceId } }); return r.data
}
export async function updateTypeTemplate(id: number, data: any): Promise<any> {
  const r = await api.put(`/type-templates/${id}`, data); return r.data
}
export async function deleteTypeTemplate(id: number): Promise<void> {
  await api.delete(`/type-templates/${id}`)
}
export async function bindFieldToTypeTemplate(typeTemplateId: number, data: {field_id:number,is_required?:boolean,sequence?:number}): Promise<any> {
  const r = await api.post(`/type-templates/${typeTemplateId}/fields`, data); return r.data
}
export async function unbindFieldFromTypeTemplate(typeTemplateId: number, fieldId: number): Promise<void> {
  await api.delete(`/type-templates/${typeTemplateId}/fields/${fieldId}`)
}

export const templateApi = {
  listTemplates, getTemplate, createTemplate, updateTemplate, deleteTemplate,
  addTypeToTemplate, removeTypeFromTemplate, applyTemplate,
  listTypeTemplates, getTypeTemplate, createTypeTemplate, updateTypeTemplate, deleteTypeTemplate,
  bindFieldToTypeTemplate, unbindFieldFromTypeTemplate,
}
export default templateApi
