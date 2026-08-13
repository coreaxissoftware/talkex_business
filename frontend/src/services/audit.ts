import api from './api'
import type { AuditLogList, AuditStats, AuditLogFilter } from '../types/audit'

export const auditService = {
  async list(filter: AuditLogFilter = {}): Promise<AuditLogList> {
    const res = await api.get('/audit-logs', { params: filter })
    return res.data
  },

  async stats(): Promise<AuditStats> {
    const res = await api.get('/audit-logs/stats')
    return res.data
  },

  async exportCSV(): Promise<void> {
    const res = await api.get('/audit-logs/export-csv', { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-logs-${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
  },
}
