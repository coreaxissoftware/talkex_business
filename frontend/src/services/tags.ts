import api from './api'

export interface TagCount {
  name: string
  count: number
}

export const tagsService = {
  async list(): Promise<TagCount[]> {
    const res = await api.get('/tags')
    return res.data
  },

  async rename(oldName: string, newName: string): Promise<{ updated: number }> {
    const res = await api.post('/tags/rename', { old_name: oldName, new_name: newName })
    return res.data
  },

  async remove(name: string): Promise<{ updated: number }> {
    const res = await api.post('/tags/delete', { name })
    return res.data
  },

  async bulkApply(tag: string, contactIds: string[]): Promise<{ applied: number }> {
    const res = await api.post('/tags/bulk-apply', { tag, contact_ids: contactIds })
    return res.data
  },
}
