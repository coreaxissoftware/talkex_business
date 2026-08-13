import api from './api'
import type { AnalyticsSummary, TimeseriesPoint } from '../types/analytics'

export const analyticsService = {
  async summary(): Promise<AnalyticsSummary> {
    const res = await api.get('/analytics/summary')
    return res.data
  },

  async timeseries(days = 30): Promise<TimeseriesPoint[]> {
    const res = await api.get('/analytics/timeseries', { params: { days } })
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
