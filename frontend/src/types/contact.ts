export interface Contact {
  id: string
  owner_id: string
  phone_number: string
  name: string | null
  email: string | null
  tags: string[]
  custom_fields: Record<string, unknown>
  opted_in: boolean
  opted_in_at: string | null
  last_inbound_at: string | null
  created_at: string
  updated_at: string
}

export interface ContactCreateInput {
  phone_number: string
  name?: string | null
  email?: string | null
  tags?: string[]
  custom_fields?: Record<string, unknown>
}

export interface ContactUpdateInput {
  name?: string | null
  email?: string | null
  tags?: string[]
  custom_fields?: Record<string, unknown>
}
