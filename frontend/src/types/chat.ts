export interface Mention {
  type: 'user' | 'agent'
  id: number
  name: string
}

export interface ReactionGroup {
  emoji: string
  count: number
  user_ids: number[]
}

export interface ChatMessage {
  id: number
  chat_id: number
  sender_id: number
  sender_type: 'user' | 'agent'
  content: string
  reply_to_id: number | null
  mentions: Mention[]
  edited_at: string | null
  deleted_at: string | null
  created_at: string
  reactions: ReactionGroup[]
}

export interface Chat {
  id: number
  workspace_id: number
  project_id: number | null
  issue_id: number | null
  type: string
  title: string
  created_at: string
  messages: ChatMessage[]
}

export interface ListMessagesResponse {
  messages: ChatMessage[]
  next_cursor: string
}

export interface SendMessagePayload {
  content: string
  reply_to_id?: number | null
}
