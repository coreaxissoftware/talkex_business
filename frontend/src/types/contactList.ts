export interface ContactList {
  id: string
  owner_id: string
  name: string
  description: string
  member_count: number
  created_at: string
  updated_at: string
}

export interface ContactListCreateInput {
  name: string
  description?: string
}

export interface ContactListUpdateInput {
  name?: string
  description?: string
}
