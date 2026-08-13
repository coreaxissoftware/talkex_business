export interface Customer {
  id: string
  owner_id: string
  business_name: string
  business_category: string
  gstin: string | null
  website: string | null
  address: string | null
  city: string | null
  state: string | null
  country: string
  phone: string | null
  logo_url: string | null
  verification_status: 'pending' | 'verified' | 'rejected'
  verification_note: string | null
  created_at: string
  updated_at: string
}

export interface CustomerUpsertInput {
  business_name: string
  business_category: string
  gstin?: string | null
  website?: string | null
  address?: string | null
  city?: string | null
  state?: string | null
  country?: string
  phone?: string | null
  logo_url?: string | null
}
