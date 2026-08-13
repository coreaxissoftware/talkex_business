import api from './api'
import type { MediaItem } from '../types/media'

export const mediaService = {
  async list(): Promise<MediaItem[]> {
    const res = await api.get('/media')
    return res.data
  },

  async upload(file: File): Promise<MediaItem> {
    const formData = new FormData()
    formData.append('file', file)
    const res = await api.post('/media/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/media/${id}`)
  },
}
