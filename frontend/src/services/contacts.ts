import api from './api'
import type { Contact, ContactCreateInput, ContactUpdateInput } from '../types/contact'

export const contactsService = {
  async list(): Promise<Contact[]> {
    const res = await api.get('/contacts')
    return res.data
  },

  async create(data: ContactCreateInput): Promise<Contact> {
    const res = await api.post('/contacts', data)
    return res.data
  },

  async update(id: string, data: ContactUpdateInput): Promise<Contact> {
    const res = await api.patch(`/contacts/${id}`, data)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/contacts/${id}`)
  },
}
