import api from './api'
import type { ChannelCatalogItem, ChannelConfig, ChannelKind } from '../types/channel'

export const channelsService = {
  async catalog(): Promise<ChannelCatalogItem[]> {
    const res = await api.get('/channels/catalog')
    return res.data
  },

  async configs(): Promise<ChannelConfig[]> {
    const res = await api.get('/channels')
    return res.data
  },

  async setEnabled(
    kind: ChannelKind,
    enabled: boolean,
    config?: Record<string, unknown>,
  ): Promise<ChannelConfig> {
    const res = await api.put(`/channels/${kind}`, { enabled, config })
    return res.data
  },
}
