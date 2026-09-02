import api from './api'

export interface Branding {
  id?: string
  owner_id?: string
  brand_name: string
  tagline: string
  primary_color: string
  accent_color: string
  logo_url: string
  logo_dark_url: string
  favicon_url: string
  custom_domain: string
  from_email: string
  support_url: string
  privacy_url: string
  terms_url: string
  hide_powered_by: boolean
}

export const whitelabelService = {
  async get(): Promise<Branding> {
    const res = await api.get('/branding/mine')
    return res.data
  },
  async update(patch: Partial<Branding>): Promise<Branding> {
    const res = await api.put('/branding/mine', patch)
    return res.data
  },
  async publicBranding(): Promise<Branding> {
    const res = await api.get('/branding')
    return res.data
  },
}
