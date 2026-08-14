import api from './api'

export interface QualityStats {
  status: 'green' | 'yellow' | 'red'
  blocks_last_7d: number
  reports_last_7d: number
  total_blocks: number
  total_reports: number
  flagged_at: string | null
  threshold: number
  health_score: number
}

export interface QualityEvent {
  id: string
  owner_id: string
  contact_id: string
  channel: string
  type: 'block' | 'report' | 'unblock'
  reason: string | null
  created_at: string
}

export const qualityService = {
  async stats(): Promise<QualityStats> {
    const res = await api.get('/quality/stats')
    return res.data
  },

  async events(): Promise<QualityEvent[]> {
    const res = await api.get('/quality/events')
    return res.data
  },
}
