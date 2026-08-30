import { useEffect, useState, type FormEvent } from 'react'
import { UsersRound, UserPlus, Trash2, Check, X, Mail, Shield, Eye, Headphones } from 'lucide-react'
import { teamService } from '../services/team'
import type { TeamMember } from '../types/team'
import Modal from '../components/Modal'

const roleConfig: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  admin: { label: 'Admin', color: 'bg-purple-50 text-purple-700', icon: <Shield size={12} /> },
  agent: { label: 'Agent', color: 'bg-blue-50 text-blue-700', icon: <Headphones size={12} /> },
  viewer: { label: 'Viewer', color: 'bg-gray-100 text-gray-600', icon: <Eye size={12} /> },
}

export default function Team() {
  const [members, setMembers] = useState<TeamMember[]>([])
  const [loading, setLoading] = useState(true)

  // Invite modal
  const [showInvite, setShowInvite] = useState(false)
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [role, setRole] = useState('agent')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // Delete confirm
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      setMembers(await teamService.list())
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleInvite = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      await teamService.invite({ email, name, role })
      setShowInvite(false)
      setEmail('')
      setName('')
      setRole('agent')
      await load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not send invite')
    } finally {
      setSaving(false)
    }
  }

  const handleRoleChange = async (m: TeamMember, newRole: string) => {
    try {
      await teamService.updateRole(m.id, newRole)
      await load()
    } catch {
      // ignore
    }
  }

  const handleRemove = async (id: string) => {
    try {
      await teamService.remove(id)
      setDeletingId(null)
      await load()
    } catch {
      // ignore
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <UsersRound size={24} className="text-primary-600" />
            Team Members
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Invite team members and manage their roles.
          </p>
        </div>
        <button
          onClick={() => { setShowInvite(true); setError('') }}
          className="flex items-center gap-1.5 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <UserPlus size={16} /> Invite Member
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-400">Loading...</div>
      ) : members.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <UsersRound size={48} className="mx-auto mb-3 opacity-30" />
          <p className="font-medium">No team members yet</p>
          <p className="text-sm mt-1">Invite your team to collaborate on messaging campaigns.</p>
        </div>
      ) : (
        <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Member</th>
                  <th className="px-4 py-2.5">Role</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5">Invited</th>
                  <th className="px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {members.map(m => {
                  // rc retained for future use — role icon/tint tooltip
                  const _rc = roleConfig[m.role] || roleConfig.viewer
                  void _rc
                  return (
                    <tr key={m.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3">
                        <p className="font-medium text-gray-900">{m.name || '—'}</p>
                        <p className="text-xs text-gray-400 flex items-center gap-1">
                          <Mail size={10} /> {m.email}
                        </p>
                      </td>
                      <td className="px-4 py-3">
                        <select
                          value={m.role}
                          onChange={e => handleRoleChange(m, e.target.value)}
                          className="rounded-lg border border-gray-200 px-2 py-1 text-xs font-medium focus:border-primary-500 outline-none"
                        >
                          <option value="admin">Admin</option>
                          <option value="agent">Agent</option>
                          <option value="viewer">Viewer</option>
                        </select>
                      </td>
                      <td className="px-4 py-3">
                        {m.status === 'active' ? (
                          <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700">
                            <Check size={12} /> Active
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 rounded-full bg-yellow-50 px-2 py-0.5 text-xs font-semibold text-yellow-700">
                            Pending
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-400">
                        {new Date(m.created_at).toLocaleDateString()}
                      </td>
                      <td className="px-4 py-3">
                        {deletingId === m.id ? (
                          <div className="flex items-center gap-1 justify-end">
                            <button
                              onClick={() => handleRemove(m.id)}
                              className="rounded-lg px-2 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 transition-colors"
                            >
                              Confirm
                            </button>
                            <button
                              onClick={() => setDeletingId(null)}
                              className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 transition-colors"
                            >
                              <X size={14} />
                            </button>
                          </div>
                        ) : (
                          <div className="flex justify-end">
                            <button
                              onClick={() => setDeletingId(m.id)}
                              className="rounded-lg p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                              title="Remove"
                            >
                              <Trash2 size={14} />
                            </button>
                          </div>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Invite Modal */}
      {showInvite && (
        <Modal title="Invite Team Member" onClose={() => setShowInvite(false)}>
          <form onSubmit={handleInvite} className="space-y-4">
            {error && (
              <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
                {error}
              </div>
            )}

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">
                Email <span className="text-red-500">*</span>
              </label>
              <input
                type="email"
                required
                value={email}
                onChange={e => setEmail(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="teammate@company.com"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Name</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="Full name (optional)"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">
                Role <span className="text-red-500">*</span>
              </label>
              <select
                value={role}
                onChange={e => setRole(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              >
                <option value="admin">Admin — Full access</option>
                <option value="agent">Agent — Conversations & contacts</option>
                <option value="viewer">Viewer — Read-only</option>
              </select>
            </div>

            <div className="flex gap-2 pt-2">
              <button
                type="button"
                onClick={() => setShowInvite(false)}
                className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
              >
                {saving ? 'Sending...' : 'Send Invite'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
