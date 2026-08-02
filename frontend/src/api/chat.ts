/**
 * Chat API - 聊天API模块
 */
import api from './index'
import type {
  Chat,
  ChatMessage,
  ListMessagesResponse,
  SendMessagePayload,
} from '@/types/chat'

/**
 * 懒获取/创建 issue 关联的 chat（返回最近 50 条消息）
 */
export async function getIssueChat(issueId: number): Promise<Chat> {
  const response = await api.get(`/issues/${issueId}/chat`)
  return response.data
}

/**
 * 获取 chat 详情
 */
export async function getChat(chatId: number): Promise<Chat> {
  const response = await api.get(`/chats/${chatId}`)
  return response.data
}

/**
 * 分页加载历史消息（游标分页，加载更老的消息）
 */
export async function listMessages(
  chatId: number,
  cursor: string = '',
  limit: number = 20
): Promise<ListMessagesResponse> {
  const response = await api.get(`/chats/${chatId}/messages`, {
    params: { cursor, limit },
  })
  return response.data
}

/**
 * 发送消息
 */
export async function sendMessage(
  chatId: number,
  payload: SendMessagePayload
): Promise<ChatMessage> {
  const response = await api.post(`/chats/${chatId}/messages`, payload)
  return response.data
}

/**
 * 编辑消息（30 分钟窗口内）
 */
export async function editMessage(
  messageId: number,
  content: string
): Promise<ChatMessage> {
  const response = await api.put(`/messages/${messageId}`, { content })
  return response.data
}

/**
 * 软删除消息
 */
export async function deleteMessage(messageId: number): Promise<void> {
  await api.delete(`/messages/${messageId}`)
}

/**
 * 添加表情反应（幂等）
 */
export async function addReaction(
  messageId: number,
  emoji: string
): Promise<void> {
  await api.post(`/messages/${messageId}/reactions`, { emoji })
}

/**
 * 删除表情反应（幂等）
 */
export async function removeReaction(
  messageId: number,
  emoji: string
): Promise<void> {
  await api.delete(`/messages/${messageId}/reactions`, { data: { emoji } })
}

export default {
  getIssueChat,
  getChat,
  listMessages,
  sendMessage,
  editMessage,
  deleteMessage,
  addReaction,
  removeReaction,
}
