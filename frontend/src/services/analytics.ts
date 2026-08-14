import api from './api'
import type { AnalyticsSummary, TimeseriesPoint } from '../types/analytics'

export interface ChannelCostBreak {
  channel: string
  messages: number
  cost: number
  revenue: number
  margin: number
}

export interface CostSummary {
  total_cost: number
  total_revenue: number
  total_margin: number
  margin_percent: number
  by_channel: ChannelCostBreak[]
}

export const analyticsService = {
  async summary(): Promise<AnalyticsSummary> {
    const res = await api.get('/analytics/summary')
    return res.data
  },

  async timeseries(days = 30): Promise<TimeseriesPoint[]> {
    const res = await api.get('/analytics/timeseries', { params: { days } })
    return res.data
  },

  async costs(): Promise<CostSummary> {
    const res = await api.get('/analytics/costs')
    return res.data
  },

  async exportCSV(days = 30): Promise<void> {
    const res = await api.get('/analytics/export-csv', { params: { days }, responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `analytics-${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
  },
}
