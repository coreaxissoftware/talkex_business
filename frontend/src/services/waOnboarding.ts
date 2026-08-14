import api from './api'

export interface WAOnboarding {
  id: string
  owner_id: string
  current_step: string
  business_name: string
  business_website: string
  business_category: string
  business_address: string
  fb_business_manager_id: string | null
  verification_status: string
  phone_number: string | null
  phone_verified: boolean
  display_name: string | null
  display_name_status: string
  waba_id: string | null
  phone_number_id: string | null
  created_at: string
  updated_at: string
}

export const waOnboardingService = {
  async get(): Promise<WAOnboarding | null> {
    const res = await api.get('/channels/whatsapp/onboarding')
    if (res.data.status === 'not_started') return null
    return res.data
  },

  async start(): Promise<WAOnboarding> {
    const res = await api.post('/channels/whatsapp/onboarding/start')
    return res.data
  },

  async saveBusinessInfo(data: {
    business_name: string
    business_website?: string
    business_category?: string
    business_address?: string
  }): Promise<WAOnboarding> {
    const res = await api.put('/channels/whatsapp/onboarding/business-info', data)
    return res.data
  },

  async saveVerification(fb_business_manager_id: string): Promise<WAOnboarding> {
    const res = await api.put('/channels/whatsapp/onboarding/verification', { fb_business_manager_id })
    return res.data
  },

  async savePhoneRegistration(phone_number: string): Promise<WAOnboarding> {
    const res = await api.put('/channels/whatsapp/onboarding/phone-registration', { phone_number })
    return res.data
  },

  async saveDisplayName(display_name: string): Promise<WAOnboarding> {
    const res = await api.put('/channels/whatsapp/onboarding/display-name', { display_name })
    return res.data
  },
}
