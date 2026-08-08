export interface ChannelBreakdown {
  channel: string
  count: number
}

export interface AnalyticsSummary {
  total_messages: number
  outbound_messages: number
  inbound_messages: number
  delivery_rate: number
  open_conversations: number
  total_contacts: number
  active_campaigns: number
  by_status: Record<string, number>
  by_channel: ChannelBreakdown[]
}

export interface TimeseriesPoint {
  date: string
  outbound: number
  inbound: number
}
