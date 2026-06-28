/**
 * AI Types — AI 智能助手类型定义
 */

export interface AIChatRequest {
  message: string
  thread_id?: number
  mode: 'ask' | 'build' | 'chart'
}

export interface StreamEvent {
  type: 'text' | 'tool_call' | 'tool_result' | 'thinking' | 'done' | 'error'
  content?: string
  tool_call?: ToolCall
  tool_result?: ToolResult
  thread_id?: number
  error?: string
}

export interface ToolCall {
  id: string
  name: string
  input: Record<string, any>
}

export interface ToolResult {
  tool_call_id: string
  content: string
}

export interface AISearchRequest {
  query: string
}

export interface AISearchResponse {
  rql: string
  explanation: string
  issues: IssueLite[]
}

export interface IssueLite {
  id: number
  name: string
  sequence_id: number
  priority: string
  state_id: number
}

export interface AICreateRequest {
  description: string
}

export interface AICreateResponse {
  preview: Record<string, any>
  explanation: string
}

export interface AIChatMessage {
  role: 'user' | 'assistant' | 'tool'
  content: string
  toolCalls?: ToolCall[]
  toolResults?: ToolResult[]
  chartConfig?: any
}
