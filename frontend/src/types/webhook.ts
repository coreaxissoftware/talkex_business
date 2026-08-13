export interface WebhookEndpoint {
  id: string
  owner_id: string
  name: string
  url: string
  events: string[]
  active: boolean
  last_fired_at: string | null
  created_at: string
  updated_at: string
}

export interface WebhookEndpointCreateResult {
  endpoint: WebhookEndpoint
  plaintext_secret: string
}

export interface WebhookDelivery {
  id: string
  endpoint_id: string
  event: string
  payload: string
  status_code: number
  success: boolean
  attempts: number
  error_message?: string
  delivered_at?: string
  created_at: string
}

export const WEBHOOK_EVENTS = [
  'inbound.message',
  'message.status',
  'campaign.completed',
  'contact.created',
] as const
export type WebhookEvent = (typeof WEBHOOK_EVENTS)[number]
