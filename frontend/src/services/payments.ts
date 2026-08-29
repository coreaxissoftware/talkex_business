import api from './api'

export interface PaymentOrder {
  order_id: string
  amount: number
  currency: string
  key_id: string
  dev_mode: boolean
}

export interface PaymentOrderRecord {
  id: string
  owner_id: string
  amount: number
  currency: string
  status: 'created' | 'paid' | 'failed'
  payment_id: string
  created_at: string
  paid_at: string | null
}

export const paymentsService = {
  async createOrder(amount: number): Promise<PaymentOrder> {
    const res = await api.post('/payments/order', { amount })
    return res.data
  },
  async listOrders(): Promise<PaymentOrderRecord[]> {
    const res = await api.get('/payments/orders')
    return res.data
  },
  async devSimulate(orderId: string): Promise<PaymentOrderRecord> {
    const res = await api.post('/payments/dev-simulate', { order_id: orderId })
    return res.data
  },
}
