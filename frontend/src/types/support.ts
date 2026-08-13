export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent'
export type TicketStatus = 'open' | 'in_progress' | 'resolved'

export interface Ticket {
  id: string
  owner_id: string
  subject: string
  body: string
  priority: TicketPriority
  status: TicketStatus
  resolved_at: string | null
  created_at: string
  updated_at: string
}

export interface TicketCreateInput {
  subject: string
  body: string
  priority?: TicketPriority
}
