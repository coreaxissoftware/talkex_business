import api from './api'
import type {
  MessageTemplate,
  TemplateCreateInput,
  TemplateUpdateInput,
} from '../types/template'

export const templatesService = {
  async list(): Promise<MessageTemplate[]> {
    const res = await api.get('/templates')
    return res.data
  },

  async create(data: TemplateCreateInput): Promise<MessageTemplate> {
    const res = await api.post('/templates', data)
    return res.data
  },

  async update(id: string, data: TemplateUpdateInput): Promise<MessageTemplate> {
    const res = await api.patch(`/templates/${id}`, data)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/templates/${id}`)
  },
}
