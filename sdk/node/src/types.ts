/** Data types mirroring the Go client's DTO definitions. */

// -- Auth --

export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_at: string;
}

export interface User {
  id: number;
  email: string;
  username: string;
  display_name: string;
  avatar_url: string;
}

export interface PATResponse {
  id: number;
  name: string;
  token_prefix: string;
  scopes: string;
  last_used_at: string | null;
  expires_at: string | null;
  revoked_at: string | null;
  created_at: string;
}

export interface CreatePATResponse extends PATResponse {
  token: string;
}

// -- Workspaces --

export interface Workspace {
  id: number;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
}

export interface UserLite {
  id: number;
  display_name: string;
  email: string;
  avatar_url: string;
}

export interface Member {
  id: number;
  workspace_id: number;
  user_id: number;
  role: number;
  is_active: boolean;
  user: UserLite;
  created_at: string;
  updated_at: string;
}

// -- Projects --

export interface Project {
  id: number;
  name: string;
  identifier: string;
  description: string;
  workspace_id: number;
  archived_at: string | null;
  total_issues: number;
  total_members: number;
  is_favorite: boolean;
  created_at: string;
  updated_at: string;
}

// -- Issues --

export interface Issue {
  id: number;
  name: string;
  description_html: string;
  priority: string;
  sequence_id: number;
  sort_order: number;
  start_date: string | null;
  target_date: string | null;
  completed_at: string | null;
  is_draft: boolean;
  archived_at: string | null;
  project_id: number;
  workspace_id: number;
  state_id: number;
  state_name: string;
  state_group: string;
  parent_id: number | null;
  depth: number;
  assignees: UserLite[];
  labels: number[];
  sub_issues_count: number;
  link_count: number;
  estimate_point_id: number | null;
  cycle_id: number | null;
  module_ids: number[];
  release_id: number | null;
  created_at: string;
  updated_at: string;
}

export interface IssueListResult {
  items: Issue[];
  total: number;
}

export interface IssueSearchResult {
  id: number;
  name: string;
  sequence_id: number;
  project_identifier: string;
  project_id: number;
  workspace_slug: string;
}

export interface Comment {
  id: number;
  issue_id: number;
  author_id: number;
  body: string;
  is_resolved: boolean;
  parent_id: number | null;
  created_at: string;
  updated_at: string;
}

// -- Cycles --

export interface Cycle {
  id: number;
  name: string;
  description: string | null;
  status: string;
  progress: number;
  total_issues: number;
  completed_issues: number;
  start_date: string;
  end_date: string | null;
  project_id: number;
  workspace_id: number;
  created_at: string;
  updated_at: string;
}

export interface CycleListResult {
  items: Cycle[];
  total: number;
  limit: number;
  offset: number;
}

export interface StateBreakdown {
  state: string;
  group: string;
  count: number;
}

export interface CycleProgress {
  cycle_id: number;
  cycle_name: string;
  total_issues: number;
  completed_issues: number;
  progress: number;
  state_breakdown: StateBreakdown[];
}

export interface BurndownDayPoint {
  day_index: number;
  date: string;
  ideal_remaining: number;
  actual_completed: number;
  actual_remaining: number;
}

export interface BurndownData {
  cycle_id: number;
  cycle_name: string;
  start_date: string;
  end_date: string;
  total_issues: number;
  total_days: number;
  days_elapsed: number;
  ideal_daily_burn: number;
  ideal_remaining: number;
  actual_completed: number;
  actual_remaining: number;
  is_on_track: boolean;
  daily_points: BurndownDayPoint[];
}

// -- Meta --

export interface State {
  id: number;
  name: string;
  color: string;
  group: string;
  sequence: number;
  is_default: boolean;
  is_active: boolean;
  project_id: number;
  workspace_id: number;
}

export interface Label {
  id: number;
  name: string;
  color: string;
  description: string;
  project_id: number;
  workspace_id: number;
}

export interface IssueType {
  id: number;
  name: string;
  color: string;
  icon: string;
  description: string;
  level: string;
  parent_type_id: number | null;
  is_default: boolean;
  sequence: number;
  is_active: boolean;
  project_id: number;
  workspace_id: number;
}

export interface Page {
  id: number;
  title: string;
  content: string;
  published: boolean;
  sequence: number;
  parent_id: number | null;
  depth: number;
  project_id: number;
  workspace_id: number;
  created_at: string;
  updated_at: string;
}

export interface Notification {
  id: number;
  title: string;
  message: string;
  type: string;
  priority: string;
  is_read: boolean;
  read_at: string | null;
  action_url: string;
  issue_id: number | null;
  created_at: string;
}

// -- AI --

export interface AISearchResponse {
  rql: string;
  explanation: string;
  issues: Record<string, unknown>[];
}

export interface AIChatReply {
  text: string;
  thread_id: number;
  tool_calls: string[];
}

// -- Agents --

export interface Agent {
  id: number;
  workspace_id: number;
  name: string;
  avatar: string;
  agent_type: string;
  capabilities: string[];
  status: string;
  model_override: string | null;
  system_prompt: string | null;
  template_id: number | null;
  created_at: string;
  updated_at: string;
}

export interface AgentActivity {
  id: number;
  agent_id: number;
  issue_id: number | null;
  action: string;
  result_summary: string;
  rating: number | null;
  executed_at: string | null;
  agent_name: string;
  task_context: string;
  created_at: string;
  updated_at: string;
}

export interface AgentTask {
  id: number;
  title: string;
  description: string;
  status: string;
  priority: number;
  progress: number;
  task_type: string;
  output_data: unknown;
  error_info: string;
  failure_reason: string;
  workspace_id: number;
  project_id: number | null;
  issue_id: number | null;
  enqueued_at: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface TaskLog {
  id: number;
  task_id: number;
  level: string;
  message: string;
  created_at: string;
}

// -- Chat --

export interface Message {
  id: number;
  chat_id: number;
  sender_id: number;
  sender_type: string;
  content: string;
  reply_to_id: number | null;
  edited_at: string | null;
  deleted_at: string | null;
  created_at: string;
}

export interface Chat {
  id: number;
  workspace_id: number;
  project_id: number | null;
  issue_id: number | null;
  type: string;
  title: string;
  created_at: string;
  messages: Message[];
}

// -- Request types --

export interface IssueListOptions {
  project_id?: number;
  workspace_id?: number;
  rql?: string;
  state_id?: number;
  priority?: string;
  assignee_id?: number;
  cycle_id?: number;
  issue_type_id?: number;
  search?: string;
  sort_by?: string;
  sort_dir?: string;
  limit?: number;
  offset?: number;
}

export interface CreateIssueRequest {
  name: string;
  description_html?: string;
  priority?: string;
  state_id?: number;
  assignee_ids?: number[];
  label_ids?: number[];
  start_date?: string;
  target_date?: string;
  parent_id?: number;
  type_id?: number;
  cycle_id?: number;
}

export interface UpdateIssueRequest {
  name?: string;
  description_html?: string;
  priority?: string;
  state_id?: number;
  assignee_ids?: number[];
  label_ids?: number[];
  target_date?: string;
  parent_id?: number;
  type_id?: number;
  cycle_id?: number;
}

export interface ProjectCreateRequest {
  name: string;
  identifier: string;
  description?: string;
}

export interface WorkspaceCreateRequest {
  name: string;
  slug: string;
}

export interface CreatePATRequest {
  name: string;
  expires_at?: string;
}
