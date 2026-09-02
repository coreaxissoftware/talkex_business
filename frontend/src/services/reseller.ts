import api from './api'

export interface ChannelRow {
  channel: string
  messages: number
  cost: number
  revenue: number
  margin: number
}

export interface TenantRow {
  tenant_owner_id: string
  tenant_name: string
  messages: number
  cost: number
  revenue: number
  margin: number
  margin_pct: number
  per_channel: ChannelRow[]
}

export interface ResellerDashboard {
  reseller_owner_id: string
  window_from: string
  window_to: string
  total_tenants: number
  total_messages: number
  total_cost: number
  total_revenue: number
  total_margin: number
  avg_margin_pct: number
  tenants: TenantRow[]
}

export const resellerService = {
  async dashboard(days = 30): Promise<ResellerDashboard> {
    const res = await api.get('/reseller/dashboard', { params: { days } })
    return res.data
  },
}
