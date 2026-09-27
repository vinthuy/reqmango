/**
 * Comment API - 评论API模块
 */
import api from './index'
import type { Comment, CommentCreate, CommentUpdate, CommentListResponse, SuggestionField, ApplySuggestionsResult } from '@/types/comment'

const BASE_URL = '/comments'

/**
 * 创建评论
 */
export async function createComment(
  data: CommentCreate
): Promise<Comment> {
  const response = await api.post(BASE_URL, data)
  return response.data
}

/**
 * 获取工作项的评论列表
 */
export async function listIssueComments(
  issueId: number,
  page: number = 1,
  pageSize: number = 20
): Promise<CommentListResponse> {
  const response = await api.get(`${BASE_URL}/issue/${issueId}`, {
    params: { page, page_size: pageSize }
  })
  return response.data
}

/**
 * 获取评论详情
 */
export async function getComment(
  commentId: number
): Promise<Comment> {
  const response = await api.get(`${BASE_URL}/${commentId}`)
  return response.data
}

/**
 * 更新评论
 */
export async function updateComment(
  commentId: number,
  data: CommentUpdate
): Promise<Comment> {
  const response = await api.patch(`${BASE_URL}/${commentId}`, data)
  return response.data
}

/**
 * 删除评论
 */
export async function deleteComment(
  commentId: number
): Promise<void> {
  await api.delete(`${BASE_URL}/${commentId}`)
}

/**
 * 标记评论为已解决
 */
export async function resolveComment(
  commentId: number
): Promise<Comment> {
  const response = await api.post(`${BASE_URL}/${commentId}/resolve`)
  return response.data
}

/**
 * 取消评论解决状态
 */
export async function unresolveComment(
  commentId: number
): Promise<Comment> {
  const response = await api.post(`${BASE_URL}/${commentId}/unresolve`)
  return response.data
}

/**
 * 采纳 Agent 在评论里提出的字段变更建议。
 * `fields` 为空表示采纳该评论上的全部建议。
 */
export async function applySuggestions(
  issueId: number,
  commentId: number,
  fields: SuggestionField[] = []
): Promise<ApplySuggestionsResult> {
  const response = await api.post(`/issues/${issueId}/suggestions/apply`, {
    comment_id: commentId,
    fields,
  })
  return response.data
}

export default {
  createComment,
  listIssueComments,
  getComment,
  updateComment,
  deleteComment,
  resolveComment,
  unresolveComment,
  applySuggestions,
}
