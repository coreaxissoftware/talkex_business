import api from './api'
import type { Plan, PlanID, SubscriptionResponse, Invoice } from '../types/billing'

export const billingService = {
  async listPlans(): Promise<Plan[]> {
    const res = await api.get('/billing/plans')
    return res.data
  },

  async getSubscription(): Promise<SubscriptionResponse> {
    const res = await api.get('/billing/subscription')
    return res.data
  },

  async changePlan(plan: PlanID): Promise<{ subscription: SubscriptionResponse['subscription']; invoice: Invoice }> {
    const res = await api.post('/billing/subscription', { plan })
    return res.data
  },

  async listInvoices(): Promise<Invoice[]> {
    const res = await api.get('/billing/invoices')
    return res.data
  },
}
