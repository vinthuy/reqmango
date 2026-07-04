/** Represents an AI agent that can be assigned to work items. */
export interface Agent {
  id: number
  workspace_id: number
  name: string
  avatar: string                           // emoji, e.g. "🤖"
  agent_type: 'builtin' | 'custom'
  capabilities: string[]                   // JSON array of tool names
  status: 'active' | 'inactive'
  model_override?: string
  system_prompt?: string
  created_by_id: number
  created_at: string
  updated_at: string
}

/** Records every action an AI agent performs — used for audit trail. */
export interface AgentActivity {
  id: number
  agent_id: number
  issue_id?: number
  action: string                           // "dispatch" | "auto_triage" | "auto_assign" | "mention" | "summarize" | "custom"
  result_summary: string
  executed_at: string
  agent_name: string
  task_context?: string
  rating?: number | null                     // 1=positive, -1=negative, null=no feedback
  created_at: string
}

/** Request body for dispatching a task to an agent. */
export interface AgentDispatchRequest {
  task: string
  issue_id?: number
  project_id?: number
}

/** Request body for creating a new agent. */
export interface AgentCreateRequest {
  name: string
  avatar?: string
  agent_type?: 'builtin' | 'custom'
  capabilities?: string[]
  status?: 'active' | 'inactive'
  model_override?: string
  system_prompt?: string
}

/** Request body for updating an existing agent. */
export interface AgentUpdateRequest {
  name?: string
  avatar?: string
  agent_type?: 'builtin' | 'custom'
  capabilities?: string[]
  status?: 'active' | 'inactive'
  model_override?: string
  system_prompt?: string
}
