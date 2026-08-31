import api from './api'

export interface ZapierEvent {
  key: string
  name: string
  description: string
  sample_keys: string[]
}

export interface SheetsImportRequest {
  url: string
  phone_column?: string
  name_column?: string
  skip_header?: boolean
  default_opt_in?: boolean
}

export interface SheetsImportResult {
  imported: number
  skipped: number
  total_rows: number
}

export const integrationsService = {
  async zapierEvents(): Promise<ZapierEvent[]> {
    const res = await api.get('/integrations/zapier/events')
    return res.data
  },
  async importFromSheet(req: SheetsImportRequest): Promise<SheetsImportResult> {
    const res = await api.post('/integrations/sheets/import', req)
    return res.data
  },
}

export const analyticsService = {
  pdfURL(): string {
    return `${api.defaults.baseURL}/analytics/pdf`
  },
}
