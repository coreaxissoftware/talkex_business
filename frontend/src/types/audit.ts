export interface AuditLogEntry {
  id: string
  created_at: string
  user_id: string | null
  user_email: string | null
  method: string
  path: string
  status_code: number
  success: boolean
  latency_ms: number
  client_ip: string
  error_body?: string
}

export interface AuditLogList {
  items: AuditLogEntry[]
  total: number
}

export interface AuditStats {
  total: number
  failed: number
  success_rate: number
}

export interface AuditLogFilter {
  failed?: boolean
  method?: string
  search?: string
  limit?: number
  offset?: number
}
