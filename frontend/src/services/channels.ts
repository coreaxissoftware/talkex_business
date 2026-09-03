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

  // Mint a TalkEx bulk API key from the merchant's TalkEx credentials.
  // Returns { key, prefix, label } on success, or { requires_pin, pending_token }
  // if the TalkEx account has 2FA on — caller re-submits the same call
  // with { pin, pending_token } to complete.
  async generateTalkExKey(req: {
    talkex_username: string
    talkex_password: string
    label?: string
    pin?: string
    pending_token?: string
    base_url?: string
  }): Promise<{
    key?: string
    key_id?: string
    prefix?: string
    label?: string
    requires_pin?: boolean
    pending_token?: string
    warning?: string
  }> {
    const res = await api.post('/channels/talkex/generate-key', req)
    return res.data
  },
}
