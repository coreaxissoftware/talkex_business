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
}
