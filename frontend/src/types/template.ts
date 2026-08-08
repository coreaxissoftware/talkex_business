export type TemplateCategory = 'marketing' | 'utility' | 'authentication'
export type TemplateStatus = 'draft' | 'pending_review' | 'approved' | 'rejected'

export interface MessageTemplate {
  id: string
  owner_id: string
  name: string
  category: TemplateCategory
  channel: string
  body: string
  variables: string[]
  status: TemplateStatus
  created_at: string
  updated_at: string
}

export interface TemplateCreateInput {
  name: string
  category: TemplateCategory
  channel: string
  body: string
  variables?: string[]
}

export interface TemplateUpdateInput {
  name?: string
  body?: string
  variables?: string[]
  status?: TemplateStatus
}
