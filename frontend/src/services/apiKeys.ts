import api from './api'
import type { ApiKey, ApiKeyCreateResult } from '../types/apiKey'

export const apiKeysService = {
  async list(): Promise<ApiKey[]> {
    const res = await api.get('/api-keys')
    return res.data
  },

  async create(name: string): Promise<ApiKeyCreateResult> {
    const res = await api.post('/api-keys', { name })
    return res.data
  },

  async revoke(id: string): Promise<ApiKey> {
    const res = await api.post(`/api-keys/${id}/revoke`)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/api-keys/${id}`)
  },
}
