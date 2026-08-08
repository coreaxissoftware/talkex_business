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
}
