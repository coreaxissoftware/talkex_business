import api from './api'
import type { TeamMember, InviteInput } from '../types/team'

export const teamService = {
  async list(): Promise<TeamMember[]> {
    const res = await api.get('/team')
    return res.data
  },

  async invite(data: InviteInput): Promise<TeamMember> {
    const res = await api.post('/team/invite', data)
    return res.data
  },

  async updateRole(id: string, role: string): Promise<TeamMember> {
    const res = await api.patch(`/team/${id}/role`, { role })
    return res.data
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/team/${id}`)
  },
}
