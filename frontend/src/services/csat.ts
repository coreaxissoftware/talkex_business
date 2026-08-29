import api from './api'

export interface Rating {
  id: string
  owner_id: string
  conversation_id: string
  contact_id: string
  agent_user_id?: string
  score: number
  comment: string
  channel: string
  created_at: string
}

export interface CsatSummary {
  total: number
  average: number
  distribution: Record<string, number>
  per_agent: Array<{ agent_user_id: string; count: number; average: number }>
  per_channel: Array<{ channel: string; count: number; average: number }>
}

export const csatService = {
  async list(limit = 50): Promise<Rating[]> {
    const res = await api.get('/csat', { params: { limit } })
    return res.data
  },
  async summary(): Promise<CsatSummary> {
    const res = await api.get('/csat/summary')
    return res.data
  },
  async submit(data: {
    conversation_id: string
    contact_id?: string
    agent_user_id?: string
    score: number
    comment?: string
    channel?: string
  }): Promise<Rating> {
    const res = await api.post('/csat', data)
    return res.data
  },
}
