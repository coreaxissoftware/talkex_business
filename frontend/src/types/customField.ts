export interface CustomFieldDefinition {
  id: string
  owner_id: string
  name: string
  label: string
  type: 'text' | 'number' | 'date' | 'boolean' | 'dropdown'
  required: boolean
  options: string
  created_at: string
  updated_at: string
}

export interface CustomFieldCreateInput {
  name: string
  label: string
  type: string
  required?: boolean
  options?: string
}
