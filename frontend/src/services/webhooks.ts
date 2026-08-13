import api from './api'
import type { WebhookEndpoint, WebhookEndpointCreateResult, WebhookDelivery } from '../types/webhook'

export interface CreateWebhookInput {
  name: string
  url: string
  events: string[]
  active: boolean
}

export const webhooksService = {
  async list(): Promise<WebhookEndpoint[]> {
    const res = await api.get('/webhooks')
    return res.data
  },

  async create(data: CreateWebhookInput): Promise<WebhookEndpointCreateResult> {
    const res = await api.post('/webhooks', data)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/webhooks/${id}`)
  },

  async deliveries(id: string): Promise<WebhookDelivery[]> {
    const res = await api.get(`/webhooks/${id}/deliveries`)
    return res.data
  },

  async retryDelivery(deliveryId: string): Promise<void> {
    await api.post(`/webhooks/deliveries/${deliveryId}/retry`)
  },
}
