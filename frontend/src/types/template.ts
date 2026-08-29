export type TemplateCategory = 'marketing' | 'utility' | 'authentication'
export type TemplateStatus = 'draft' | 'pending_review' | 'approved' | 'rejected'

export interface TemplateButton {
  type: 'quick_reply' | 'url' | 'phone'
  text: string
  url?: string
  phone?: string
}

export interface TemplateListRow {
  id: string
  title: string
  description?: string
}

export interface MessageTemplate {
  id: string
  owner_id: string
  name: string
  category: TemplateCategory
  channel: string
  body: string
  variables: string[]
  status: TemplateStatus
  buttons: TemplateButton[]
  list_rows: TemplateListRow[]
  header: string
  footer: string
  media_type: string
  media_url: string
  submitted_at?: number
  external_ref: string
  external_status: string
  reject_reason: string
  created_at: string
  updated_at: string
}

export interface TemplateCreateInput {
  name: string
  category: TemplateCategory
  channel: string
  body: string
  variables?: string[]
  buttons?: TemplateButton[]
  list_rows?: TemplateListRow[]
  header?: string
  footer?: string
  media_type?: string
  media_url?: string
}

export interface TemplateUpdateInput {
  name?: string
  body?: string
  variables?: string[]
  status?: TemplateStatus
  buttons?: TemplateButton[]
  list_rows?: TemplateListRow[]
  header?: string
  footer?: string
  media_type?: string
  media_url?: string
}
