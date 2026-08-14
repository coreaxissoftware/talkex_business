import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  ShieldCheck,
  FileText,
  Users,
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  Plus,
  Eye,
  Trash2,
  X,
} from 'lucide-react'
import {
  complianceService,
  type ComplianceStats,
  type ConsentRecord,
  type DSARRequest,
  type ProcessingRecord,
} from '../services/compliance'
import Modal from '../components/Modal'

const DSAR_STATUS_STYLE: Record<string, { bg: string; text: string }> = {
  pending:    { bg: 'bg-amber-50',  text: 'text-amber-700' },
  processing: { bg: 'bg-blue-50',   text: 'text-blue-700' },
  completed:  { bg: 'bg-green-50',  text: 'text-green-700' },
  rejected:   { bg: 'bg-red-50',    text: 'text-red-700' },
}

type Tab = 'overview' | 'consents' | 'dsars' | 'processing'

export default function Compliance() {
  const [tab, setTab] = useState<Tab>('overview')
  const [stats, setStats] = useState<ComplianceStats | null>(null)
  const [consents, setConsents] = useState<ConsentRecord[]>([])
  const [dsars, setDsars] = useState<DSARRequest[]>([])
  const [processing, setProcessing] = useState<ProcessingRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // DSAR modals
  const [dsarModalOpen, setDsarModalOpen] = useState(false)
  const [dsarContactId, setDsarContactId] = useState('')
  const [dsarType, setDsarType] = useState('access')
  const [dsarReason, setDsarReason] = useState('')
  const [dsarSaving, setDsarSaving] = useState(false)

  // Complete DSAR modal
  const [completeDsarId, setCompleteDsarId] = useState<string | null>(null)
  const [completeResponse, setCompleteResponse] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [s, c, d, p] = await Promise.all([
        complianceService.stats(),
        complianceService.listConsents(),
        complianceService.listDSARs(),
        complianceService.listProcessing(),
      ])
      setStats(s)
      setConsents(c)
      setDsars(d)
      setProcessing(p)
      setError('')
    } catch {
      setError('Could not load compliance data.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const handleCreateDSAR = async (e: FormEvent) => {
    e.preventDefault()
    setDsarSaving(true)
    try {
      await complianceService.createDSAR({
        contact_id: dsarContactId,
        type: dsarType,
        reason: dsarReason || undefined,
      })
      setDsarModalOpen(false)
      setDsarContactId('')
      setDsarType('access')
      setDsarReason('')
      load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not create DSAR')
    } finally {
      setDsarSaving(false)
    }
  }

  const handleProcessDSAR = async (id: string) => {
    try {
      await complianceService.processDSAR(id)
      load()
    } catch { setError('Could not update DSAR') }
  }

  const handleCompleteDSAR = async () => {
    if (!completeDsarId) return
    try {
      await complianceService.completeDSAR(completeDsarId, completeResponse)
      setCompleteDsarId(null)
      setCompleteResponse('')
      load()
    } catch { setError('Could not complete DSAR') }
  }

  const handleRejectDSAR = async (id: string) => {
    const reason = prompt('Rejection reason:')
    if (!reason) return
    try {
      await complianceService.rejectDSAR(id, reason)
      load()
    } catch { setError('Could not reject DSAR') }
  }

  const tabs: { key: Tab; label: string; icon: typeof ShieldCheck }[] = [
    { key: 'overview', label: 'Overview', icon: ShieldCheck },
    { key: 'consents', label: 'Consents', icon: CheckCircle2 },
    { key: 'dsars', label: 'DSAR Requests', icon: FileText },
    { key: 'processing', label: 'Processing Log', icon: Eye },
  ]

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <ShieldCheck size={24} className="text-primary-600" />
          DPDP Compliance
        </h1>
        <p className="text-sm text-gray-500 mt-1">
          Digital Personal Data Protection Act 2023 — consent, DSAR, and processing records.
        </p>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 border-b border-gray-200">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`flex items-center gap-1.5 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              tab === t.key
                ? 'border-primary-600 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            <t.icon size={14} />
            {t.label}
          </button>
        ))}
      </div>

      {loading && !stats ? (
        <div className="p-10 text-center text-sm text-gray-400">Loading compliance data…</div>
      ) : (
        <>
          {/* Overview */}
          {tab === 'overview' && stats && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
                <div className="rounded-xl border border-gray-200 bg-white p-4 text-center">
                  <p className="text-2xl font-bold text-green-600">{stats.active_consents}</p>
                  <p className="text-xs text-gray-500 mt-1">Active Consents</p>
                </div>
                <div className="rounded-xl border border-gray-200 bg-white p-4 text-center">
                  <p className="text-2xl font-bold text-gray-500">{stats.revoked_consents}</p>
                  <p className="text-xs text-gray-500 mt-1">Revoked</p>
                </div>
                <div className="rounded-xl border border-gray-200 bg-white p-4 text-center">
                  <p className="text-2xl font-bold text-amber-600">{stats.pending_dsars}</p>
                  <p className="text-xs text-gray-500 mt-1">Pending DSARs</p>
                </div>
                <div className="rounded-xl border border-gray-200 bg-white p-4 text-center">
                  <p className="text-2xl font-bold text-green-600">{stats.completed_dsars}</p>
                  <p className="text-xs text-gray-500 mt-1">Completed DSARs</p>
                </div>
                <div className="rounded-xl border border-gray-200 bg-white p-4 text-center">
                  <p className="text-2xl font-bold text-primary-600">{stats.processing_logs}</p>
                  <p className="text-xs text-gray-500 mt-1">Processing Records</p>
                </div>
              </div>

              <div className="rounded-xl border border-gray-200 bg-white p-5">
                <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2 mb-3">
                  <AlertTriangle size={16} className="text-amber-500" />
                  Compliance Checklist
                </h2>
                <div className="space-y-2">
                  {[
                    { label: 'Consent records maintained', done: stats.active_consents > 0 },
                    { label: 'DSAR process established', done: stats.completed_dsars > 0 || stats.pending_dsars > 0 },
                    { label: 'Processing activities logged', done: stats.processing_logs > 0 },
                    { label: 'No overdue DSAR requests', done: stats.pending_dsars === 0 },
                  ].map(item => (
                    <div key={item.label} className="flex items-center gap-2 text-sm">
                      {item.done ? (
                        <CheckCircle2 size={16} className="text-green-500" />
                      ) : (
                        <XCircle size={16} className="text-gray-300" />
                      )}
                      <span className={item.done ? 'text-gray-900' : 'text-gray-400'}>{item.label}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Consents */}
          {tab === 'consents' && (
            <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
              {consents.length === 0 ? (
                <div className="p-10 text-center">
                  <Users size={32} className="mx-auto text-gray-300 mb-2" />
                  <p className="text-sm text-gray-500">No consent records yet.</p>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        <th className="px-4 py-2.5">Contact</th>
                        <th className="px-4 py-2.5">Purpose</th>
                        <th className="px-4 py-2.5">Channel</th>
                        <th className="px-4 py-2.5">Status</th>
                        <th className="px-4 py-2.5">Source</th>
                        <th className="px-4 py-2.5">Date</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {consents.map(c => (
                        <tr key={c.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 text-gray-700 font-mono text-xs">{c.contact_id.slice(0, 8)}…</td>
                          <td className="px-4 py-3 capitalize">{c.purpose}</td>
                          <td className="px-4 py-3 capitalize">{c.channel}</td>
                          <td className="px-4 py-3">
                            {c.consent_given ? (
                              <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700">
                                <CheckCircle2 size={12} /> Active
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 rounded-full bg-red-50 px-2 py-0.5 text-xs font-semibold text-red-700">
                                <XCircle size={12} /> Revoked
                              </span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-xs text-gray-500">{c.source}</td>
                          <td className="px-4 py-3 text-xs text-gray-400">
                            {new Date(c.created_at).toLocaleDateString('en-IN')}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* DSARs */}
          {tab === 'dsars' && (
            <div className="space-y-4">
              <div className="flex justify-end">
                <button
                  onClick={() => setDsarModalOpen(true)}
                  className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
                >
                  <Plus size={16} />
                  New DSAR Request
                </button>
              </div>

              <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
                {dsars.length === 0 ? (
                  <div className="p-10 text-center">
                    <FileText size={32} className="mx-auto text-gray-300 mb-2" />
                    <p className="text-sm text-gray-500">No DSAR requests yet.</p>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                          <th className="px-4 py-2.5">Contact</th>
                          <th className="px-4 py-2.5">Type</th>
                          <th className="px-4 py-2.5">Status</th>
                          <th className="px-4 py-2.5">Date</th>
                          <th className="px-4 py-2.5"></th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {dsars.map(d => {
                          const style = DSAR_STATUS_STYLE[d.status] || DSAR_STATUS_STYLE.pending
                          return (
                            <tr key={d.id} className="hover:bg-gray-50">
                              <td className="px-4 py-3 text-gray-700 font-mono text-xs">{d.contact_id.slice(0, 8)}…</td>
                              <td className="px-4 py-3 capitalize font-medium">{d.type}</td>
                              <td className="px-4 py-3">
                                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold ${style.bg} ${style.text}`}>
                                  {d.status}
                                </span>
                              </td>
                              <td className="px-4 py-3 text-xs text-gray-400">
                                {new Date(d.created_at).toLocaleDateString('en-IN')}
                              </td>
                              <td className="px-4 py-3">
                                <div className="flex items-center gap-1 justify-end">
                                  {d.status === 'pending' && (
                                    <>
                                      <button
                                        onClick={() => handleProcessDSAR(d.id)}
                                        className="rounded-lg p-1.5 text-blue-600 hover:bg-blue-50 transition-colors"
                                        title="Start Processing"
                                      >
                                        <Clock size={14} />
                                      </button>
                                      <button
                                        onClick={() => handleRejectDSAR(d.id)}
                                        className="rounded-lg p-1.5 text-red-600 hover:bg-red-50 transition-colors"
                                        title="Reject"
                                      >
                                        <XCircle size={14} />
                                      </button>
                                    </>
                                  )}
                                  {d.status === 'processing' && (
                                    <button
                                      onClick={() => { setCompleteDsarId(d.id); setCompleteResponse('') }}
                                      className="rounded-lg p-1.5 text-green-600 hover:bg-green-50 transition-colors"
                                      title="Complete"
                                    >
                                      <CheckCircle2 size={14} />
                                    </button>
                                  )}
                                </div>
                              </td>
                            </tr>
                          )
                        })}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Processing Records */}
          {tab === 'processing' && (
            <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
              {processing.length === 0 ? (
                <div className="p-10 text-center">
                  <Eye size={32} className="mx-auto text-gray-300 mb-2" />
                  <p className="text-sm text-gray-500">No processing records yet. Activities are logged automatically.</p>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        <th className="px-4 py-2.5">Activity</th>
                        <th className="px-4 py-2.5">Purpose</th>
                        <th className="px-4 py-2.5">Data Category</th>
                        <th className="px-4 py-2.5">Legal Basis</th>
                        <th className="px-4 py-2.5">Date</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {processing.map(p => (
                        <tr key={p.id} className="hover:bg-gray-50">
                          <td className="px-4 py-3 font-medium capitalize">{p.activity.replace(/_/g, ' ')}</td>
                          <td className="px-4 py-3 capitalize text-gray-600">{p.purpose}</td>
                          <td className="px-4 py-3 capitalize text-gray-600">{p.data_category}</td>
                          <td className="px-4 py-3">
                            <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium capitalize">
                              {p.legal_basis.replace(/_/g, ' ')}
                            </span>
                          </td>
                          <td className="px-4 py-3 text-xs text-gray-400">
                            {new Date(p.created_at).toLocaleDateString('en-IN')}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </>
      )}

      {/* New DSAR modal */}
      <Modal open={dsarModalOpen} onClose={() => setDsarModalOpen(false)} title="New DSAR Request">
        <form onSubmit={handleCreateDSAR} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Contact ID *</label>
            <input
              type="text"
              required
              value={dsarContactId}
              onChange={e => setDsarContactId(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="Enter contact ID"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Request Type *</label>
            <select
              value={dsarType}
              onChange={e => setDsarType(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none bg-white"
            >
              <option value="access">Access (view my data)</option>
              <option value="erasure">Erasure (right to be forgotten)</option>
              <option value="correction">Correction (fix my data)</option>
              <option value="portability">Portability (export my data)</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Reason</label>
            <textarea
              value={dsarReason}
              onChange={e => setDsarReason(e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="Optional reason for request..."
            />
          </div>
          <div className="flex gap-2 pt-2">
            <button type="button" onClick={() => setDsarModalOpen(false)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">
              Cancel
            </button>
            <button type="submit" disabled={dsarSaving || !dsarContactId}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors">
              {dsarSaving ? 'Creating...' : 'Create Request'}
            </button>
          </div>
        </form>
      </Modal>

      {/* Complete DSAR modal */}
      <Modal open={!!completeDsarId} onClose={() => setCompleteDsarId(null)} title="Complete DSAR Request">
        <div className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Response / Actions Taken</label>
            <textarea
              value={completeResponse}
              onChange={e => setCompleteResponse(e.target.value)}
              rows={4}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="Describe what data was provided/deleted/corrected..."
            />
          </div>
          <div className="flex gap-2">
            <button onClick={() => setCompleteDsarId(null)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors">
              Cancel
            </button>
            <button onClick={handleCompleteDSAR}
              className="flex-1 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 transition-colors">
              Mark Complete
            </button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
