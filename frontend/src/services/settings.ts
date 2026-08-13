import api from './api'

export interface UserPrefs {
  notif_campaigns: boolean
  notif_messages: boolean
  notif_system: boolean
  email_digest: boolean
  timezone: string
  language: string
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
