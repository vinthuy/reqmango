export interface TimeTrack {
  id: number
  issue_id: number
  user_id: number
  description?: string
  started_at: string
  ended_at?: string
  duration: number // seconds
  created_at: string
}

export interface TimeTrackSummary {
  issue_id: number
  total_seconds: number
  total_hours: number
  entry_count: number
  is_running: boolean
}
