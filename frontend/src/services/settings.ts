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
  sell_whatsapp: number
  sell_sms: number
  sell_talkex: number
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
