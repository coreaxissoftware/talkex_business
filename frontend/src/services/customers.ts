import api from './api'
import type { Customer, CustomerUpsertInput } from '../types/customer'

export const customersService = {
  async get(): Promise<Customer> {
    const res = await api.get('/customers/me')
    return res.data
  },

  async upsert(data: CustomerUpsertInput): Promise<Customer> {
    const res = await api.put('/customers/me', data)
    return res.data
  },
}
