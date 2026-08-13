import api from './api'
import type { User } from '../types/auth'

export interface UpdateProfileInput {
  full_name?: string
  business_category?: string | null
}

export interface ChangePasswordInput {
  current_password: string
  new_password: string
}

export const usersService = {
  async updateProfile(data: UpdateProfileInput): Promise<User> {
    const res = await api.patch('/users/me', data)
    return res.data
  },

  async changePassword(data: ChangePasswordInput): Promise<void> {
    await api.post('/users/me/change-password', data)
  },

  async deactivateAccount(password: string): Promise<void> {
    await api.post('/users/me/deactivate', { password })
  },

  async setup2FA(): Promise<{ secret: string; provisioning_uri: string }> {
    const res = await api.post('/users/me/2fa/setup')
    return res.data
  },

  async verify2FA(code: string): Promise<void> {
    await api.post('/users/me/2fa/verify', { code })
  },

  async disable2FA(password: string, code: string): Promise<void> {
    await api.post('/users/me/2fa/disable', { password, code })
  },

  async listSessions(): Promise<Session[]> {
    const res = await api.get('/users/me/sessions')
    return res.data
  },

  async revokeSession(id: string): Promise<void> {
    await api.delete(`/users/me/sessions/${id}`)
  },

  async revokeAllSessions(): Promise<void> {
    await api.post('/users/me/sessions/revoke-all')
  },
}

export interface Session {
  id: string
  ip_address: string
  user_agent: string
  expires_at: string
  created_at: string
}
