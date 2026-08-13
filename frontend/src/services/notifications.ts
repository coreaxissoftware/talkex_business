import api from './api'
import type { Notification } from '../types/notification'

export const notificationsService = {
  async list(unread = false): Promise<Notification[]> {
    const res = await api.get('/notifications', {
      params: unread ? { unread: 'true' } : {},
    })
    return res.data
  },

  async unreadCount(): Promise<number> {
    const res = await api.get('/notifications/unread-count')
    return res.data.count
  },

  async markRead(id: string): Promise<void> {
    await api.post(`/notifications/${id}/read`)
  },

  async markAllRead(): Promise<void> {
    await api.post('/notifications/read-all')
  },
}
