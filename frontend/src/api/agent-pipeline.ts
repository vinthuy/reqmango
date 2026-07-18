import api from './index'

export interface Pipeline {
  id: number; workspace_id: number; name: string; description?: string
  pipeline_def: any; version: string; status: string
  created_by_id?: number; created_at: string; updated_at: string
}

export interface PipelineRun {
  id: number; pipeline_id: number; trigger_type: string
  status: 'pending'|'running'|'completed'|'failed'
  stages_result?: any[]; tokens_used: number; cost_usd: number
  started_at?: string; completed_at?: string; error_message?: string
  created_at: string
}

export const pipelineApi = {
  list(ws: number): Promise<Pipeline[]> { return api.get(`/workspaces/${ws}/pipelines`).then(r=>r.data) },
  get(ws: number, id: number): Promise<Pipeline> { return api.get(`/workspaces/${ws}/pipelines/${id}`).then(r=>r.data) },
  create(ws: number, data: {name:string;description?:string;pipeline_def:any}): Promise<Pipeline> { return api.post(`/workspaces/${ws}/pipelines`,data).then(r=>r.data) },
  update(ws: number, id: number, data: Partial<Pipeline>): Promise<Pipeline> { return api.put(`/workspaces/${ws}/pipelines/${id}`,data).then(r=>r.data) },
  delete(ws: number, id: number): Promise<void> { return api.delete(`/workspaces/${ws}/pipelines/${id}`) },
  run(ws: number, id: number): Promise<PipelineRun> { return api.post(`/workspaces/${ws}/pipelines/${id}/run`).then(r=>r.data) },
  getRuns(ws: number, id: number): Promise<PipelineRun[]> { return api.get(`/workspaces/${ws}/pipelines/${id}/runs`).then(r=>r.data) },
  getRun(ws: number, runId: number): Promise<PipelineRun> { return api.get(`/workspaces/${ws}/pipelines/runs/${runId}`).then(r=>r.data) },
}
