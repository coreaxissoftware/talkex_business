export type MatchType = 'contains' | 'exact' | 'starts_with'

export interface AutomationRule {
  id: string
  owner_id: string
  name: string
  trigger_keywords: string[]
  match_type: MatchType
  response_body: string
  template_id: string | null
  active: boolean
  fire_count: number
  created_at: string
  updated_at: string
}

export interface AutomationRuleInput {
  name: string
  trigger_keywords: string[]
  match_type: MatchType
  response_body: string
  template_id?: string | null
  active: boolean
}
