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

  async simulateInbound(data: InboundInput) {
    const res = await api.post('/conversations/inbound', data)
    return res.data
  },
}
