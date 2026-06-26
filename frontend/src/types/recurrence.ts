export interface RecurrenceRule {
  id: number
  issue_id: number
  frequency: 'daily' | 'weekly' | 'monthly' | 'cron'
  interval: number
  cron_expr?: string
  next_run: string
  end_date?: string
  is_active: boolean
  created_at: string
}

export interface RecurrenceCreate {
  frequency: string
  interval?: number
  cron_expr?: string
  next_run: string
  end_date?: string
}

export interface RecurrenceUpdate {
  frequency?: string
  interval?: number
  cron_expr?: string
  next_run?: string
  end_date?: string
  is_active?: boolean
}
