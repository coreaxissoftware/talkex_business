import api from './api'
import type { CustomFieldDefinition, CustomFieldCreateInput } from '../types/customField'

export const customFieldsService = {
  async list(): Promise<CustomFieldDefinition[]> {
    const res = await api.get('/custom-fields')
    return res.data
  },

  async create(data: CustomFieldCreateInput): Promise<CustomFieldDefinition> {
    const res = await api.post('/custom-fields', data)
    return res.data
  },

  async update(id: string, data: Partial<CustomFieldCreateInput>): Promise<CustomFieldDefinition> {
    const res = await api.patch(`/custom-fields/${id}`, data)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/custom-fields/${id}`)
  },
}
