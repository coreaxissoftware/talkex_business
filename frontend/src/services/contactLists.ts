import api from './api'
import type { ContactList, ContactListCreateInput, ContactListUpdateInput } from '../types/contactList'

export const contactListsService = {
  async list(): Promise<ContactList[]> {
    const res = await api.get('/contact-lists')
    return res.data
  },

  async create(data: ContactListCreateInput): Promise<ContactList> {
    const res = await api.post('/contact-lists', data)
    return res.data
  },

  async update(id: string, data: ContactListUpdateInput): Promise<ContactList> {
    const res = await api.patch(`/contact-lists/${id}`, data)
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/contact-lists/${id}`)
  },

  async getMembers(id: string): Promise<string[]> {
    const res = await api.get(`/contact-lists/${id}/members`)
    return res.data.contact_ids
  },

  async addMembers(id: string, contactIds: string[]): Promise<{ added: number }> {
    const res = await api.post(`/contact-lists/${id}/members`, { contact_ids: contactIds })
    return res.data
  },

  async removeMembers(id: string, contactIds: string[]): Promise<void> {
    await api.delete(`/contact-lists/${id}/members`, { data: { contact_ids: contactIds } })
  },
}
