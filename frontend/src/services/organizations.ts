import api from './api'

export interface Organization {
  id: string
  name: string
  slug: string
  owner_id: string
  parent_id: string | null
  tier: string
  logo_url: string | null
  website: string | null
  max_users: number
  max_sub_orgs: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface OrgMember {
  id: string
  org_id: string
  user_id: string
  role: string
  created_at: string
}

export const organizationsService = {
  async list(): Promise<Organization[]> {
    const res = await api.get('/organizations')
    return res.data
  },

  async get(id: string): Promise<Organization> {
    const res = await api.get(`/organizations/${id}`)
    return res.data
  },

  async create(data: { name: string; slug: string; parent_id?: string; website?: string }): Promise<Organization> {
    const res = await api.post('/organizations', data)
    return res.data
  },

  async update(id: string, data: { name?: string; website?: string; logo_url?: string }): Promise<Organization> {
    const res = await api.patch(`/organizations/${id}`, data)
    return res.data
  },

  async listMembers(orgId: string): Promise<OrgMember[]> {
    const res = await api.get(`/organizations/${orgId}/members`)
    return res.data
  },

  async addMember(orgId: string, userId: string, role: string): Promise<OrgMember> {
    const res = await api.post(`/organizations/${orgId}/members`, { user_id: userId, role })
    return res.data
  },

  async removeMember(orgId: string, userId: string): Promise<void> {
    await api.delete(`/organizations/${orgId}/members/${userId}`)
  },

  async listSubOrgs(orgId: string): Promise<Organization[]> {
    const res = await api.get(`/organizations/${orgId}/sub-orgs`)
    return res.data
  },
}
