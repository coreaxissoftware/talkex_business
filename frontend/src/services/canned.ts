import api from './api'

export interface CannedResponse {
  id: string
  owner_id: string
  shortcut: string
  title: string
  body: string
  category: string
  usage_count: number
  created_at: string
  updated_at: string
}

export interface CannedInput {
  shortcut: string
  title: string
  body: string
  category?: string
}

export const cannedService = {
  async list(): Promise<CannedResponse[]> {
    const res = await api.get('/canned-responses')
    return res.data
  },
  async create(data: CannedInput): Promise<CannedResponse> {
    const res = await api.post('/canned-responses', data)
    return res.data
  },
  async update(id: string, data: Partial<CannedInput>): Promise<CannedResponse> {
    const res = await api.patch(`/canned-responses/${id}`, data)
    return res.data
  },
  async remove(id: string): Promise<void> {
    await api.delete(`/canned-responses/${id}`)
  },
  async bumpUsage(id: string): Promise<void> {
    await api.post(`/canned-responses/${id}/use`)
  },
}
