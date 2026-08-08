export type QualityStatus = 'green' | 'yellow' | 'red'

export interface User {
  id: string
  email: string
  full_name: string
  role: 'owner' | 'admin' | 'agent' | 'developer'
  is_active: boolean
  is_business_verified: boolean
  business_category: string | null
  quality_flagged_at: string | null
  quality_status: QualityStatus
  created_at: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  token_type: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  full_name: string
}
