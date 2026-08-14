import api from './api'

export interface ConsentRecord {
  id: string
  owner_id: string
  contact_id: string
  purpose: string
  channel: string
  consent_given: boolean
  consented_at: string | null
  revoked_at: string | null
  source: string
  ip_address: string | null
  created_at: string
}

export interface DSARRequest {
  id: string
  owner_id: string
  contact_id: string
  type: string
  status: string
  reason: string | null
  response: string | null
  completed_at: string | null
  created_at: string
}

export interface ProcessingRecord {
  id: string
  owner_id: string
  contact_id: string
  activity: string
  purpose: string
  data_category: string
  legal_basis: string
  details: string
  created_at: string
}

export interface ComplianceStats {
  active_consents: number
  revoked_consents: number
  pending_dsars: number
  completed_dsars: number
  processing_logs: number
}

export const complianceService = {
  async stats(): Promise<ComplianceStats> {
    const res = await api.get('/compliance/stats')
    return res.data
  },

  async listConsents(): Promise<ConsentRecord[]> {
    const res = await api.get('/compliance/consents')
    return res.data
  },

  async recordConsent(data: {
    contact_id: string
    purpose: string
    channel: string
    consent_given: boolean
    source: string
  }): Promise<ConsentRecord> {
    const res = await api.post('/compliance/consents', data)
    return res.data
  },

  async revokeAll(contactId: string): Promise<{ revoked: number }> {
    const res = await api.post(`/compliance/consents/${contactId}/revoke-all`)
    return res.data
  },

  async listDSARs(): Promise<DSARRequest[]> {
    const res = await api.get('/compliance/dsars')
    return res.data
  },

  async createDSAR(data: {
    contact_id: string
    type: string
    reason?: string
  }): Promise<DSARRequest> {
    const res = await api.post('/compliance/dsars', data)
    return res.data
  },

  async processDSAR(id: string): Promise<DSARRequest> {
    const res = await api.post(`/compliance/dsars/${id}/process`)
    return res.data
  },

  async completeDSAR(id: string, response: string): Promise<DSARRequest> {
    const res = await api.post(`/compliance/dsars/${id}/complete`, { response })
    return res.data
  },

  async rejectDSAR(id: string, reason: string): Promise<DSARRequest> {
    const res = await api.post(`/compliance/dsars/${id}/reject`, { reason })
    return res.data
  },

  async listProcessing(): Promise<ProcessingRecord[]> {
    const res = await api.get('/compliance/processing')
    return res.data
  },
}
