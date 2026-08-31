import api from './api'

export interface GreenTickApplication {
  id: string
  owner_id: string
  status: 'not_started' | 'in_progress' | 'submitted' | 'approved' | 'rejected'
  notable_brand: boolean
  org_website: boolean
  meta_200_msg: boolean
  meta_tier2: boolean
  business_verified: boolean
  trademark_refs: boolean
  submitted_at?: string | null
  decided_at?: string | null
  meta_case_id?: string
  reject_reason?: string
}

export interface GreenTickResponse {
  application: GreenTickApplication
  progress: number
}

export const greenTickService = {
  async get(): Promise<GreenTickResponse> {
    const res = await api.get('/green-tick')
    return res.data
  },
  async update(patch: Partial<GreenTickApplication>): Promise<GreenTickResponse> {
    const res = await api.patch('/green-tick', patch)
    return res.data
  },
  async submit(metaCaseId: string): Promise<GreenTickApplication> {
    const res = await api.post('/green-tick/submit', { meta_case_id: metaCaseId })
    return res.data
  },
  async decision(approved: boolean, reason?: string): Promise<GreenTickApplication> {
    const res = await api.post('/green-tick/decision', { approved, reason: reason || '' })
    return res.data
  },
}
