import api from './api'
import type { LoginRequest, RegisterRequest, TokenResponse, User } from '../types/auth'

export const authService = {
  async register(data: RegisterRequest): Promise<User> {
    const res = await api.post('/auth/register', data)
    return res.data
  },

  async login(data: LoginRequest): Promise<TokenResponse> {
    const res = await api.post('/auth/login', data)
    const tokens: TokenResponse = res.data
    localStorage.setItem('access_token', tokens.access_token)
    localStorage.setItem('refresh_token', tokens.refresh_token)
    return tokens
  },

  async getMe(): Promise<User> {
    const res = await api.get('/users/me')
    return res.data
  },

  logout() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    window.location.href = '/login'
  },

  isAuthenticated(): boolean {
    return !!localStorage.getItem('access_token')
  },
}
