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
}
