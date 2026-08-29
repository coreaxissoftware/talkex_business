import api from './api'
import type {
  ConversationRow,
  ConversationThread,
  Conversation,
  SendInput,
  InboundInput,
} from '../types/conversation'

export const conversationsService = {
  async list(): Promise<ConversationRow[]> {
    const res = await api.get('/conversations')
    return res.data
  },

  async thread(id: string): Promise<ConversationThread> {
    const res = await api.get(`/conversations/${id}/messages`)
    return res.data
  },

  async markRead(id: string): Promise<Conversation> {
    const res = await api.post(`/conversations/${id}/read`)
    return res.data
  },

  async send(data: SendInput) {
    const res = await api.post('/conversations/send', data)
    return res.data
  },

  async update(id: string, data: { labels?: string[]; assigned_to?: string; assigned_name?: string }): Promise<Conversation> {
    const res = await api.patch(`/conversations/${id}`, data)
    return res.data
  },

  async simulateInbound(data: InboundInput) {
    const res = await api.post('/conversations/inbound', data)
    return res.data
  },

  async search(q: string): Promise<ConversationRow[]> {
    const res = await api.get('/conversations/search', { params: { q } })
    return res.data
  },

  async bulkAssign(ids: string[], agentUserID: string, agentName: string) {
    const res = await api.post('/conversations/bulk-assign', {
      ids, agent_user_id: agentUserID, agent_name: agentName,
    })
    return res.data
  },

  async bulkMarkRead(ids: string[]) {
    const res = await api.post('/conversations/bulk-read', { ids })
    return res.data
  },
}
