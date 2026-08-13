export type PlanID = 'starter' | 'growth' | 'scale'

export interface Plan {
  id: PlanID
  name: string
  price_inr_per_month: number
  included_messages: number
  overage_per_msg_inr: number
  max_contacts: number
  max_team_members: number
  features: string[]
  recommended: boolean
}

export interface Subscription {
  id: string
  owner_id: string
  plan: PlanID
  period_start: string
  messages_used: number
  status: 'active' | 'past_due' | 'cancelled'
  current_invoice_id: string | null
  created_at: string
  updated_at: string
}

export interface SubscriptionResponse {
  subscription: Subscription
  plan: Plan
}

export interface Invoice {
  id: string
  owner_id: string
  plan: PlanID
  period_start: string
  period_end: string
  messages_used: number
  amount_inr: number
  status: 'paid' | 'pending' | 'failed'
  paid_at: string | null
  created_at: string
}
