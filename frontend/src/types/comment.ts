/**
 * Comment Types - 评论类型定义
 */

export interface Comment {
  id: number
  content: string
  html_content?: string
  issue_id: number
  author_id: number
  author?: UserLite
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
  content: string
  html_content?: string
  parent_id?: number
}

export interface CommentUpdate {
  content?: string
  html_content?: string
}

export interface CommentListResponse {
  items: Comment[]
  total: number
  page: number
  page_size: number
}

export interface UserLite {
  id: number
  username: string
  display_name?: string
  email: string
  avatar_url?: string
}
