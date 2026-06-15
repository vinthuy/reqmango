/**
 * Comment Types - 评论类型定义
 */

export interface Comment {
  id: string
  content: string
  html_content?: string
  issue_id: string
  author_id: string
  author?: UserLite
  parent_id?: string
  is_resolved: boolean
  resolved_by_id?: string
  resolved_at?: string
  reaction_count: number
  replies?: Comment[]
  created_at: string
  updated_at: string
}

export interface CommentCreate {
  issue_id: string
  content: string
  html_content?: string
  parent_id?: string
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
  id: string
  username: string
  display_name?: string
  email: string
  avatar_url?: string
}
