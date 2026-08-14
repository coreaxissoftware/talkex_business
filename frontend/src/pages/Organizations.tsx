import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  Building2,
  Plus,
  Users,
  Globe,
  Crown,
  Trash2,
  X,
  Check,
  ChevronRight,
} from 'lucide-react'
import {
  organizationsService,
  type Organization,
  type OrgMember,
} from '../services/organizations'
import Modal from '../components/Modal'

const TIER_STYLE: Record<string, { bg: string; text: string }> = {
  free:       { bg: 'bg-gray-100',   text: 'text-gray-700' },
  starter:    { bg: 'bg-blue-50',    text: 'text-blue-700' },
  business:   { bg: 'bg-purple-50',  text: 'text-purple-700' },
  enterprise: { bg: 'bg-amber-50',   text: 'text-amber-700' },
  reseller:   { bg: 'bg-green-50',   text: 'text-green-700' },
}

export default function Organizations() {
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Create modal
  const [createOpen, setCreateOpen] = useState(false)
  const [formName, setFormName] = useState('')
  const [formSlug, setFormSlug] = useState('')
  const [formWebsite, setFormWebsite] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [formError, setFormError] = useState('')

  // Members panel
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null)
  const [members, setMembers] = useState<OrgMember[]>([])
  const [subOrgs, setSubOrgs] = useState<Organization[]>([])
  const [membersLoading, setMembersLoading] = useState(false)

  // Add member
  const [addUserId, setAddUserId] = useState('')
  const [addRole, setAddRole] = useState('member')
  const [addSaving, setAddSaving] = useState(false)

  // Delete confirm
  const [removingMember, setRemovingMember] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await organizationsService.list()
      setOrgs(data)
      setError('')
    } catch {
      setError('Could not load organizations.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setFormError('')
    setFormSaving(true)
    try {
      const created = await organizationsService.create({
        name: formName,
        slug: formSlug,
        website: formWebsite || undefined,
      })
      setOrgs(prev => [created, ...prev])
      setCreateOpen(false)
      setFormName('')
      setFormSlug('')
      setFormWebsite('')
    } catch (err: any) {
      setFormError(err.response?.data?.detail || 'Could not create organization')
    } finally {
      setFormSaving(false)
    }
  }

  const openOrgDetail = async (org: Organization) => {
    setSelectedOrg(org)
    setMembersLoading(true)
    try {
      const [m, s] = await Promise.all([
        organizationsService.listMembers(org.id),
        organizationsService.listSubOrgs(org.id),
      ])
      setMembers(m)
      setSubOrgs(s)
    } catch { /* ignore */ }
    finally { setMembersLoading(false) }
  }

  const handleAddMember = async () => {
    if (!selectedOrg || !addUserId.trim()) return
    setAddSaving(true)
    try {
      await organizationsService.addMember(selectedOrg.id, addUserId.trim(), addRole)
      setAddUserId('')
      setAddRole('member')
      const m = await organizationsService.listMembers(selectedOrg.id)
      setMembers(m)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not add member')
    } finally { setAddSaving(false) }
  }

  const handleRemoveMember = async (userId: string) => {
    if (!selectedOrg) return
    try {
      await organizationsService.removeMember(selectedOrg.id, userId)
      setRemovingMember(null)
      const m = await organizationsService.listMembers(selectedOrg.id)
      setMembers(m)
    } catch { setError('Could not remove member') }
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Building2 size={24} className="text-primary-600" />
            Organizations
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Manage your organizations, sub-accounts, and team members.
          </p>
        </div>
        <button
          onClick={() => { setCreateOpen(true); setFormError('') }}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} />
          New Organization
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Org list */}
        <div className="lg:col-span-1 space-y-3">
          {loading ? (
            <div className="p-10 text-center text-sm text-gray-400">Loading…</div>
          ) : orgs.length === 0 ? (
            <div className="rounded-xl border border-gray-200 bg-white p-10 text-center">
              <Building2 size={32} className="mx-auto text-gray-300 mb-2" />
              <p className="text-sm text-gray-500">No organizations yet.</p>
            </div>
          ) : (
            orgs.map(org => {
              const tier = TIER_STYLE[org.tier] || TIER_STYLE.free
              const isSelected = selectedOrg?.id === org.id
              return (
                <button
                  key={org.id}
                  onClick={() => openOrgDetail(org)}
                  className={`w-full text-left rounded-xl border bg-white p-4 transition-colors ${
                    isSelected ? 'border-primary-500 ring-2 ring-primary-500/20' : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-gray-900">{org.name}</h3>
                    <ChevronRight size={16} className="text-gray-400" />
                  </div>
                  <p className="text-xs text-gray-400 font-mono mt-0.5">/{org.slug}</p>
                  <div className="flex items-center gap-2 mt-2">
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase ${tier.bg} ${tier.text}`}>
                      {org.tier}
                    </span>
                    {!org.is_active && (
                      <span className="inline-flex items-center rounded-full bg-red-50 px-2 py-0.5 text-[10px] font-semibold text-red-700">
                        Inactive
                      </span>
                    )}
                    {org.parent_id && (
                      <span className="text-[10px] text-gray-400">sub-org</span>
                    )}
                  </div>
                </button>
              )
            })
          )}
        </div>

        {/* Detail panel */}
        <div className="lg:col-span-2">
          {!selectedOrg ? (
            <div className="rounded-xl border border-gray-200 bg-white p-10 text-center">
              <Building2 size={32} className="mx-auto text-gray-300 mb-2" />
              <p className="text-sm text-gray-500">Select an organization to view details.</p>
            </div>
          ) : (
            <div className="space-y-4">
              {/* Org info */}
              <div className="rounded-xl border border-gray-200 bg-white p-5">
                <div className="flex items-center justify-between mb-3">
                  <h2 className="text-lg font-bold text-gray-900">{selectedOrg.name}</h2>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold uppercase ${(TIER_STYLE[selectedOrg.tier] || TIER_STYLE.free).bg} ${(TIER_STYLE[selectedOrg.tier] || TIER_STYLE.free).text}`}>
                    {selectedOrg.tier}
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div>
                    <p className="text-gray-400">Slug</p>
                    <p className="font-mono text-gray-700">/{selectedOrg.slug}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Website</p>
                    <p className="text-gray-700">{selectedOrg.website || '—'}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Max Users</p>
                    <p className="text-gray-700">{selectedOrg.max_users}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Max Sub-Orgs</p>
                    <p className="text-gray-700">{selectedOrg.max_sub_orgs}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Created</p>
                    <p className="text-gray-700">{new Date(selectedOrg.created_at).toLocaleDateString('en-IN')}</p>
                  </div>
                  <div>
                    <p className="text-gray-400">Status</p>
                    <p className={selectedOrg.is_active ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>
                      {selectedOrg.is_active ? 'Active' : 'Inactive'}
                    </p>
                  </div>
                </div>
              </div>

              {/* Members */}
              <div className="rounded-xl border border-gray-200 bg-white p-5">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
                    <Users size={16} className="text-primary-600" />
                    Members
                  </h3>
                </div>

                {/* Add member */}
                <div className="flex gap-2 mb-3">
                  <input
                    type="text"
                    value={addUserId}
                    onChange={e => setAddUserId(e.target.value)}
                    placeholder="User ID"
                    className="flex-1 rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-primary-500 outline-none"
                  />
                  <select
                    value={addRole}
                    onChange={e => setAddRole(e.target.value)}
                    className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm outline-none bg-white"
                  >
                    <option value="member">Member</option>
                    <option value="admin">Admin</option>
                    <option value="viewer">Viewer</option>
                  </select>
                  <button
                    onClick={handleAddMember}
                    disabled={addSaving || !addUserId.trim()}
                    className="rounded-lg bg-primary-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
                  >
                    {addSaving ? '…' : 'Add'}
                  </button>
                </div>

                {membersLoading ? (
                  <p className="text-xs text-gray-400 text-center py-4">Loading…</p>
                ) : members.length === 0 ? (
                  <p className="text-xs text-gray-400 text-center py-4">No members yet.</p>
                ) : (
                  <div className="divide-y divide-gray-100">
                    {members.map(m => (
                      <div key={m.id} className="flex items-center justify-between py-2.5">
                        <div>
                          <p className="text-sm font-medium text-gray-900 font-mono">{m.user_id.slice(0, 12)}…</p>
                          <p className="text-xs text-gray-400 capitalize flex items-center gap-1">
                            {m.role === 'admin' && <Crown size={10} className="text-amber-500" />}
                            {m.role}
                          </p>
                        </div>
                        {removingMember === m.user_id ? (
                          <div className="flex items-center gap-1">
                            <button onClick={() => handleRemoveMember(m.user_id)}
                              className="p-1 rounded text-red-600 hover:bg-red-50"><Check size={14} /></button>
                            <button onClick={() => setRemovingMember(null)}
                              className="p-1 rounded text-gray-400 hover:bg-gray-100"><X size={14} /></button>
                          </div>
                        ) : (
                          <button onClick={() => setRemovingMember(m.user_id)}
                            className="p-1 rounded text-gray-400 hover:text-red-600 hover:bg-red-50">
                            <Trash2 size={14} />
                          </button>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Sub-orgs */}
              {subOrgs.length > 0 && (
                <div className="rounded-xl border border-gray-200 bg-white p-5">
                  <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-3">
                    <Globe size={16} className="text-primary-600" />
                    Sub-Organizations
                  </h3>
                  <div className="divide-y divide-gray-100">
                    {subOrgs.map(s => (
                      <div key={s.id} className="flex items-center justify-between py-2.5">
                        <div>
                          <p className="text-sm font-medium text-gray-900">{s.name}</p>
                          <p className="text-xs text-gray-400 font-mono">/{s.slug}</p>
                        </div>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase ${(TIER_STYLE[s.tier] || TIER_STYLE.free).bg} ${(TIER_STYLE[s.tier] || TIER_STYLE.free).text}`}>
                          {s.tier}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Create modal */}
      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="New Organization">
        <form onSubmit={handleCreate} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">{formError}</div>
          )}
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Organization Name *</label>
            <input type="text" required value={formName}
              onChange={e => setFormName(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. CoreAxis Ventures" />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Slug *</label>
            <input type="text" required value={formSlug}
              onChange={e => setFormSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-'))}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. coreaxis-ventures" />
            <p className="text-[11px] text-gray-400 mt-1">Unique URL slug. Lowercase, hyphens only.</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Website</label>
            <input type="url" value={formWebsite}
              onChange={e => setFormWebsite(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="https://coreaxis.cloud" />
          </div>
          <div className="flex gap-2 pt-2">
            <button type="button" onClick={() => setCreateOpen(false)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">
              Cancel
            </button>
            <button type="submit" disabled={formSaving}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors">
              {formSaving ? 'Creating...' : 'Create Organization'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
