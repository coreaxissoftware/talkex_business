export interface TeamMember {
  id: string
  owner_id: string
  email: string
  name: string
  role: 'admin' | 'agent' | 'viewer'
  status: 'pending' | 'active'
  user_id: string
  created_at: string
  updated_at: string
}

export interface InviteInput {
  email: string
  name?: string
  role: string
}
