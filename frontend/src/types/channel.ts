export type ChannelKind = 'talkex' | 'whatsapp' | 'telegram' | 'email' | 'sms' | 'rcs'

export interface ChannelCatalogItem {
  kind: ChannelKind
  display_name: string
  description: string
  implemented: boolean
  icon: string
}

export interface ChannelConfig {
  id: string
  owner_id: string
  kind: ChannelKind
  enabled: boolean
  config: Record<string, unknown>
  verified_at: string | null
  created_at: string
  updated_at: string
}
