import api from './api'

export type StepType =
  | 'send_message'
  | 'send_template'
  | 'wait'
  | 'branch'
  | 'assign_agent'
  | 'add_tag'
  | 'end'

export interface FlowStep {
  id: string
  type: StepType
  label: string
  output_text?: string
  template_id?: string
  wait_minutes?: number
  branch_keyword?: string
  branch_yes_id?: string
  branch_no_id?: string
  agent_user_id?: string
  tag_name?: string
  next_step_id?: string
}

export interface Flow {
  id: string
  owner_id: string
  name: string
  description: string
  trigger_type: 'keyword' | 'new_contact' | 'manual'
  trigger_keywords: string[] | string // JSON from backend
  steps: FlowStep[] | string
  first_step_id: string
  active: boolean
  run_count: number
  complete_count: number
  created_at: string
  updated_at: string
}

export interface FlowInput {
  name: string
  description?: string
  trigger_type?: string
  trigger_keywords?: string[]
  steps?: FlowStep[]
  first_step_id?: string
  active?: boolean
}

function parseFlow(f: any): Flow {
  return {
    ...f,
    trigger_keywords: typeof f.trigger_keywords === 'string'
      ? JSON.parse(f.trigger_keywords || '[]')
      : f.trigger_keywords || [],
    steps: typeof f.steps === 'string' ? JSON.parse(f.steps || '[]') : f.steps || [],
  }
}

export const flowsService = {
  async list(): Promise<Flow[]> {
    const res = await api.get('/flows')
    return (res.data as any[]).map(parseFlow)
  },
  async get(id: string): Promise<Flow> {
    const res = await api.get(`/flows/${id}`)
    return parseFlow(res.data)
  },
  async create(data: FlowInput): Promise<Flow> {
    const res = await api.post('/flows', data)
    return parseFlow(res.data)
  },
  async update(id: string, data: Partial<FlowInput>): Promise<Flow> {
    const res = await api.patch(`/flows/${id}`, data)
    return parseFlow(res.data)
  },
  async remove(id: string): Promise<void> {
    await api.delete(`/flows/${id}`)
  },
  async test(id: string, contactId: string, channel = 'talkex'): Promise<void> {
    await api.post(`/flows/${id}/test`, { contact_id: contactId, channel })
  },
}
