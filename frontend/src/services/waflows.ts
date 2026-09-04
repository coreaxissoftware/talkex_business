import api from './api'

export type FlowStatus = 'draft' | 'published' | 'deprecated' | 'blocked'
export type FlowCategory =
  | 'SIGN_UP'
  | 'SIGN_IN'
  | 'APPOINTMENT_BOOKING'
  | 'LEAD_GENERATION'
  | 'SHOPPING'
  | 'CONTACT_US'
  | 'SURVEY'
  | 'OTHER'

export interface WAFlow {
  id: string
  owner_id: string
  name: string
  category: FlowCategory
  status: FlowStatus
  version: number
  flow_json: any
  meta_flow_id?: string
  published_at?: string | null
  endpoint: string
  created_at: string
}

export interface FlowResponse {
  id: string
  owner_id: string
  flow_id: string
  contact_id?: string
  screen_id?: string
  data: any
  created_at: string
}

export interface CreateFlowInput {
  name: string
  category?: FlowCategory
  flow_json: any
  endpoint?: string
}

export const waflowsService = {
  async list(): Promise<WAFlow[]> {
    const res = await api.get('/waflows')
    return res.data
  },
  async get(id: string): Promise<WAFlow> {
    const res = await api.get(`/waflows/${id}`)
    return res.data
  },
  async create(data: CreateFlowInput): Promise<WAFlow> {
    const res = await api.post('/waflows', data)
    return res.data
  },
  async update(id: string, data: Partial<CreateFlowInput>): Promise<WAFlow> {
    const res = await api.put(`/waflows/${id}`, data)
    return res.data
  },
  async publish(id: string): Promise<WAFlow> {
    const res = await api.post(`/waflows/${id}/publish`)
    return res.data
  },
  async listResponses(id: string): Promise<FlowResponse[]> {
    const res = await api.get(`/waflows/${id}/responses`)
    return res.data
  },
}
