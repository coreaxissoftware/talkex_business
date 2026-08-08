export type MessageDirection = 'inbound' | 'outbound'
export type MessageStatus = 'queued' | 'sent' | 'delivered' | 'read' | 'failed'

export interface Conversation {
  id: string
  owner_id: string
  contact_id: string
  channel: string
  last_inbound_at: string | null
  last_outbound_at: string | null
  last_message_at: string | null
  unread_count: number
  created_at: string
  updated_at: string
}

// Inbox row — Conversation + joined contact fields.
export interface ConversationRow extends Conversation {
  contact_name: string | null
  contact_phone_number: string
}

export interface Message {
  id: string
  conversation_id: string
  direction: MessageDirection
  body: string
  status: MessageStatus
  template_id?: string | null
  delivered_at?: string | null
  read_at?: string | null
  error_reason?: string | null
  created_at: string
  updated_at: string
}

export interface ConversationThread {
  conversation: Conversation
  window_open: boolean
  messages: Message[]
}

export interface SendInput {
  contact_id: string
  channel: string
  body: string
  template_id?: string | null
}

export interface InboundInput {
  contact_id: string
  channel: string
  body: string
}
