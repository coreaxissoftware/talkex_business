import api from './api'
import type { Contact, ContactCreateInput, ContactUpdateInput } from '../types/contact'

export interface ContactListFilter {
  search?: string
  tag?: string
  limit?: number
  offset?: number
}

export interface ContactListResult {
  items: Contact[]
  total: number
}

export const contactsService = {
  async list(): Promise<Contact[]> {
    const res = await api.get('/contacts')
    return res.data
  },

  async listFiltered(filter: ContactListFilter): Promise<ContactListResult> {
    const res = await api.get('/contacts', { params: filter })
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

  async merge(keepID: string, dupID: string): Promise<Contact> {
    const res = await api.post('/contacts/merge', { keep_id: keepID, dup_id: dupID })
    return res.data
  },

  async toggleOptIn(id: string, optedIn: boolean): Promise<Contact> {
    const res = await api.post(`/contacts/${id}/opt-in`, { opted_in: optedIn })
    return res.data
  },

  async exportCSV(): Promise<void> {
    const res = await api.get('/contacts/export-csv', { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `contacts-${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
  },
}
