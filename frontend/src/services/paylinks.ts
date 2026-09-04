import api from './api'

export interface PayLink {
  id: string
  owner_id: string
  contact_id: string
  conversation_id?: string
  amount_paise: number
  currency: string
  description: string
  status: 'created' | 'sent' | 'paid' | 'expired' | 'cancelled'
  razorpay_id?: string
  url: string
  simulated: boolean
  sent_at?: string | null
  paid_at?: string | null
  expires_at?: string | null
  created_at: string
}

export interface CreatePayLinkInput {
  contact_id: string
  conversation_id?: string
  amount_paise: number
  description: string
  expire_hours?: number
}

export const paylinksService = {
  async list(status?: string): Promise<PayLink[]> {
    const res = await api.get('/pay-links', { params: status ? { status } : {} })
    return res.data
  },
  async create(data: CreatePayLinkInput): Promise<PayLink> {
    const res = await api.post('/pay-links', data)
    return res.data
  },
  async markSent(id: string): Promise<void> {
    await api.post(`/pay-links/${id}/sent`)
  },
}
