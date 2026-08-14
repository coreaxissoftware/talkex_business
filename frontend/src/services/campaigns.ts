import api from './api'
import type { Campaign, CampaignCreateInput } from '../types/campaign'

export const campaignsService = {
  async list(): Promise<Campaign[]> {
    const res = await api.get('/campaigns')
    return res.data
  },

  async get(id: string): Promise<Campaign> {
    const res = await api.get(`/campaigns/${id}`)
    return res.data
  },

  async create(data: CampaignCreateInput): Promise<Campaign> {
    const res = await api.post('/campaigns', data)
    return res.data
  },

  async launch(id: string): Promise<Campaign> {
    const res = await api.post(`/campaigns/${id}/launch`)
    return res.data
  },

  async cancel(id: string): Promise<Campaign> {
    const res = await api.post(`/campaigns/${id}/cancel`)
    return res.data
  },

  async approve(id: string): Promise<Campaign> {
    const res = await api.post(`/campaigns/${id}/approve`)
    return res.data
  },

  async reject(id: string, reason: string): Promise<Campaign> {
    const res = await api.post(`/campaigns/${id}/reject`, { reason })
    return res.data
  },

  async update(id: string, data: Partial<CampaignCreateInput>): Promise<Campaign> {
    const res = await api.patch(`/campaigns/${id}`, data)
    return res.data
  },

  async clone(id: string): Promise<Campaign> {
    const res = await api.post(`/campaigns/${id}/clone`)
    return res.data
  },

  async exportCsv(): Promise<void> {
    const res = await api.get('/campaigns/export', { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = 'campaigns.csv'
    a.click()
    URL.revokeObjectURL(url)
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/campaigns/${id}`)
  },
}
