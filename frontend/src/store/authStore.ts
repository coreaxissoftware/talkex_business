import { create } from 'zustand'
import type { User } from '../types/auth'
import { authService } from '../services/auth'

interface AuthState {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, fullName: string) => Promise<void>
  setTokens: (accessToken: string, refreshToken: string) => void
  fetchUser: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoading: true,
  isAuthenticated: authService.isAuthenticated(),

  login: async (email, password) => {
    await authService.login({ email, password })
    const user = await authService.getMe()
    set({ user, isAuthenticated: true, isLoading: false })
  },

  register: async (email, password, fullName) => {
    await authService.register({ email, password, full_name: fullName })
    await authService.login({ email, password })
    const user = await authService.getMe()
    set({ user, isAuthenticated: true, isLoading: false })
  },

  setTokens: (accessToken, refreshToken) => {
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
    set({ isAuthenticated: true })
  },

  fetchUser: async () => {
    try {
      const user = await authService.getMe()
      set({ user, isAuthenticated: true, isLoading: false })
    } catch {
      set({ user: null, isAuthenticated: false, isLoading: false })
    }
  },

  logout: () => {
    set({ user: null, isAuthenticated: false })
    authService.logout()
  },
}))
