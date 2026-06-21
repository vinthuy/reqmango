export interface RQLSearchRequest {
  entity: 'issue' | 'cycle' | 'module'
  project_id: number
  rql: string
  page?: number
  page_size?: number
}

export interface RQLSearchResponse {
  success: boolean
  data?: {
    items: any[]
    total: number
    page: number
    page_size: number
  }
  error?: {
    code: string
    message: string
  }
}

export interface RQLHistoryItem {
  id: string
  rql: string
  timestamp: number
  entityType: 'issue' | 'cycle' | 'module'
}
