/**
 * Comment Types - 评论类型定义
 */

/** A work-item change an agent proposed instead of applying it directly. */
export interface IssueSuggestion {
  field: SuggestionField
  value: string | number
  label: string
  current?: string
  reason?: string
  /** True when the proposal is to keep the current value — show it, don't apply it. */
  noop?: boolean
}

export type SuggestionField =
  | 'title'
  | 'priority'
  | 'type'
  | 'state'
  | 'assignee'
  | 'description'

export interface Comment {
  id: number
  body: string
  content?: string
  html_content?: string
  issue_id: number
  author_id: number | null
  author?: UserLite
  agent_id?: number
  agent?: AgentLite
  parent_id?: number
  is_resolved: boolean
  resolved_by_id?: number
  resolved_at?: string
  reaction_count: number
  replies?: Comment[]
  /** Present on agent replies that proposed concrete field changes. */
  suggestions?: IssueSuggestion[]
  /** Set once the proposals have been adopted, so the actions stop rendering. */
  suggestions_applied_at?: string
  created_at: string
  updated_at: string
}

export interface SkippedSuggestion {
  field: SuggestionField
  label?: string
  reason: string
}

export interface ApplySuggestionsResult {
  applied: string[]
  /** Proposals that could not be written, with the reason each was left out. */
  skipped?: SkippedSuggestion[]
  issue?: Record<string, unknown>
}

export interface CommentCreate {
  issue_id: number
  body: string
  content?: string
  html_content?: string
  parent_id?: number
}

export interface CommentUpdate {
  body?: string
  content?: string
  html_content?: string
}

export interface CommentListResponse {
  comments: Comment[]
  items?: Comment[]
  total: number
  page: number
  page_size: number
}

export interface AgentLite {
  id: number
  name: string
  avatar?: string
}

export interface UserLite {
  id: number
  username: string
  display_name?: string
  email: string
  avatar_url?: string
}
