import api from './api'
import type { Ticket, TicketCreateInput } from '../types/support'

export const supportService = {
  async list(): Promise<Ticket[]> {
    const res = await api.get('/support/tickets')
    return res.data
  },

  async create(data: TicketCreateInput): Promise<Ticket> {
    const res = await api.post('/support/tickets', data)
    return res.data
  },
}
