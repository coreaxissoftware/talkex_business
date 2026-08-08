export type CampaignStatus =
  | 'draft'
  | 'scheduled'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'

export interface Campaign {
  id: string
  owner_id: string
  name: string
  template_id: string
  channel: string
  status: CampaignStatus
  scheduled_at: string | null
  started_at: string | null
  completed_at: string | null
  contact_ids: string[]
  total_count: number
  sent_count: number
  delivered_count: number
  read_count: number
  failed_count: number
  created_at: string
  updated_at: string
}

export interface CampaignCreateInput {
  name: string
  template_id: string
  contact_ids: string[]
  scheduled_at?: string | null
}
