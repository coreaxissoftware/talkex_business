import api from './api'

export interface SuggestResult { suggestion: string; dev_mode: boolean }
export interface SummaryResult { summary: string; dev_mode: boolean }
export interface SentimentResult { score: 'positive' | 'neutral' | 'negative' | string; reason: string; dev_mode: boolean }
export interface AiStatus { dev_mode: boolean; model: string; available: boolean; key_configured: boolean }

export const aiService = {
  async status(): Promise<AiStatus> {
    const res = await api.get('/ai/status')
    return res.data
  },
  async suggestReply(conversationId: string): Promise<SuggestResult> {
    const res = await api.post('/ai/suggest-reply', { conversation_id: conversationId })
    return res.data
  },
  async summarize(conversationId: string): Promise<SummaryResult> {
    const res = await api.post('/ai/summarize', { conversation_id: conversationId })
    return res.data
  },
  async sentiment(conversationId: string): Promise<SentimentResult> {
    const res = await api.post('/ai/sentiment', { conversation_id: conversationId })
    return res.data
  },
}
