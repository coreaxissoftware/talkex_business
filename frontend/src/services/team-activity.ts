import api from './api'

export interface AgentActivity {
  user_id: string
  name: string
  email: string
  role: string
  open_assigned: number
  messages_sent_30d: number
  avg_csat_30d: number
  csat_count_30d: number
}

export const teamActivityService = {
  async list(): Promise<AgentActivity[]> {
    const res = await api.get('/team/activity')
    return res.data
  },
}
