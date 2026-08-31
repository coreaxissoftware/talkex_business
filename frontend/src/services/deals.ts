import api from './api'

export interface Pipeline {
  id: string
  owner_id: string
  name: string
  stages: string[]
  is_default: boolean
  created_at: string
}

export interface Deal {
  id: string
  owner_id: string
  pipeline_id: string
  title: string
  stage: string
  value: number
  currency: string
  contact_id?: string | null
  assigned_to?: string | null
  notes: string
  expected_close_at?: string | null
  closed_at?: string | null
  stage_changed_at: string
  created_at: string
}

export interface KanbanColumn {
  stage: string
  deals: Deal[]
  total_value: number
}

export interface DealInput {
  pipeline_id: string
  title: string
  stage: string
  value?: number
  currency?: string
  contact_id?: string
  assigned_to?: string
  notes?: string
  expected_close_at?: string
}

export const dealsService = {
  async listPipelines(): Promise<Pipeline[]> {
    const res = await api.get('/deals/pipelines')
    return res.data
  },
  async defaultPipeline(): Promise<Pipeline> {
    const res = await api.get('/deals/pipelines/default')
    return res.data
  },
  async kanban(pipelineId: string): Promise<KanbanColumn[]> {
    const res = await api.get(`/deals/pipelines/${pipelineId}/kanban`)
    return res.data
  },
  async create(data: DealInput): Promise<Deal> {
    const res = await api.post('/deals', data)
    return res.data
  },
  async move(id: string, stage: string): Promise<Deal> {
    const res = await api.post(`/deals/${id}/move`, { stage })
    return res.data
  },
}
