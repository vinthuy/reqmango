/**
 * Comment Types - 评论类型定义
 */

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
  created_at: string
  updated_at: string
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
