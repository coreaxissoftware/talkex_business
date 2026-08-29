import api from './api'

export interface UserPrefs {
  notif_campaigns: boolean
  notif_messages: boolean
  notif_system: boolean
  email_digest: boolean
  timezone: string
  language: string
  auto_pause_enabled: boolean
  min_balance: number
  sandbox_mode: boolean
  approval_threshold: number
  cost_whatsapp: number
  cost_sms: number
  cost_talkex: number
  cost_telegram: number
  cost_email: number
  cost_rcs: number
  cost_instagram: number
  cost_messenger: number
  sell_whatsapp: number
  sell_sms: number
  sell_talkex: number
  sell_telegram: number
  sell_email: number
  sell_rcs: number
  sell_instagram: number
  sell_messenger: number

  // Business hours + away message + SLA + AI auto-tag
  business_hours_enabled: boolean
  business_days: number[]
  business_open_time: string
  business_close_time: string
  away_message: string
  sla_first_response_mins: number
  ai_auto_tag_enabled: boolean
}

export const settingsService = {
  async get(): Promise<UserPrefs> {
    const res = await api.get('/settings')
    return res.data
  },

  async save(prefs: UserPrefs): Promise<UserPrefs> {
    const res = await api.put('/settings', prefs)
    return res.data
  },
}
