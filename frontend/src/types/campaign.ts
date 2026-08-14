export type CampaignStatus =
  | 'draft'
  | 'scheduled'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'paused'
  | 'pending_approval'
  | 'rejected'

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
  total_cost: number
  approval_required: boolean
  approved_by: string | null
  approved_at: string | null
  rejected_by: string | null
  rejected_at: string | null
  rejection_reason: string | null
  created_at: string
  updated_at: string
}

export interface CampaignCreateInput {
  name: string
  template_id: string
  contact_ids: string[]
  scheduled_at?: string | null
}
