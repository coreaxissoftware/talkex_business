import { useEffect, useState, useCallback, useMemo, type FormEvent } from 'react'
import {
  Megaphone,
  Plus,
  Play,
  Square,
  Trash2,
  Users,
  FileText,
  Calendar,
  CheckCircle2,
  XCircle,
  Clock,
  Send,
  X,
  BarChart3,
  Pause,
  ShieldCheck,
  Ban,
  ThumbsUp,
  ThumbsDown,
  Copy,
  ListChecks,
  Download,
} from 'lucide-react'
import { campaignsService } from '../services/campaigns'
import { templatesService } from '../services/templates'
import { contactsService } from '../services/contacts'
import { contactListsService } from '../services/contactLists'
import type { Campaign, CampaignStatus, CampaignCreateInput } from '../types/campaign'
import type { MessageTemplate } from '../types/template'
import type { Contact } from '../types/contact'
import type { ContactList } from '../types/contactList'
import Modal from '../components/Modal'
import TemplatePreview from '../components/TemplatePreview'

const STATUS_STYLE: Record<CampaignStatus, { bg: string; text: string; label: string; Icon: typeof Clock }> = {
  draft:     { bg: 'bg-gray-100',    text: 'text-gray-700',   label: 'Draft',     Icon: FileText },
  scheduled: { bg: 'bg-blue-50',     text: 'text-blue-700',   label: 'Scheduled', Icon: Calendar },
  running:   { bg: 'bg-amber-50',    text: 'text-amber-700',  label: 'Running',   Icon: Send },
  completed: { bg: 'bg-green-50',    text: 'text-green-700',  label: 'Completed', Icon: CheckCircle2 },
  failed:    { bg: 'bg-red-50',      text: 'text-red-700',    label: 'Failed',    Icon: XCircle },
  cancelled: { bg: 'bg-gray-100',    text: 'text-gray-500',   label: 'Cancelled', Icon: Square },
  paused:           { bg: 'bg-orange-50',   text: 'text-orange-700', label: 'Paused',           Icon: Pause },
  pending_approval: { bg: 'bg-purple-50',   text: 'text-purple-700', label: 'Pending Approval', Icon: ShieldCheck },
  rejected:         { bg: 'bg-red-50',      text: 'text-red-700',    label: 'Rejected',         Icon: Ban },
}

function StatusPill({ status }: { status: CampaignStatus }) {
  const s = STATUS_STYLE[status]
  const Icon = s.Icon
  return (
    <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${s.bg} ${s.text}`}>
      <Icon size={12} />
      {s.label}
    </span>
  )
}

function ProgressBar({ campaign }: { campaign: Campaign }) {
  const { total_count, sent_count, delivered_count, failed_count } = campaign
  if (total_count === 0) return <span className="text-xs text-gray-400">—</span>

  const pct = (n: number) => (n / total_count) * 100

  return (
    <div className="w-32">
      <div className="flex h-1.5 w-full overflow-hidden rounded-full bg-gray-100">
        <div className="bg-green-500" style={{ width: `${pct(delivered_count)}%` }} />
        <div className="bg-amber-400" style={{ width: `${pct(sent_count - delivered_count - failed_count)}%` }} />
        <div className="bg-red-500" style={{ width: `${pct(failed_count)}%` }} />
      </div>
      <p className="mt-1 text-[10px] text-gray-500">
        {sent_count}/{total_count} sent
      </p>
    </div>
  )
}

type RecipientMode = 'individual' | 'list'

interface FormState {
  name: string
  template_id: string
  contact_ids: Set<string>
  recipient_mode: RecipientMode
  list_id: string
  schedule: boolean
  scheduled_at: string
}

const emptyForm: FormState = {
  name: '',
  template_id: '',
  contact_ids: new Set(),
  recipient_mode: 'individual',
  list_id: '',
  schedule: false,
  scheduled_at: '',
}

export default function Campaigns() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [templates, setTemplates] = useState<MessageTemplate[]>([])
  const [contacts, setContacts] = useState<Contact[]>([])
  const [contactLists, setContactLists] = useState<ContactList[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)
  const [statsCampaign, setStatsCampaign] = useState<Campaign | null>(null)
  const [rejectCampaignId, setRejectCampaignId] = useState<string | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [c, t, ct, cl] = await Promise.all([
        campaignsService.list(),
        templatesService.list(),
        contactsService.list(),
        contactListsService.list(),
      ])
      setCampaigns(c)
      setTemplates(t)
      setContacts(ct)
      setContactLists(cl)
      setError('')
    } catch {
      setError('Could not load campaigns.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const templateById = useMemo(() => {
    const m = new Map<string, MessageTemplate>()
    templates.forEach((t) => m.set(t.id, t))
    return m
  }, [templates])

  const openCreate = () => {
    setForm({ ...emptyForm, contact_ids: new Set(), recipient_mode: 'individual', list_id: '' })
    setFormError('')
    setModalOpen(true)
  }

  const toggleContact = (id: string) => {
    setForm((f) => {
      const next = new Set(f.contact_ids)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return { ...f, contact_ids: next }
    })
  }

  const toggleAllContacts = () => {
    setForm((f) => {
      if (f.contact_ids.size === contacts.length) {
        return { ...f, contact_ids: new Set() }
      }
      return { ...f, contact_ids: new Set(contacts.map((c) => c.id)) }
    })
  }

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setFormError('')
    if (!form.template_id) {
      setFormError('Please select a template')
      return
    }
    if (form.recipient_mode === 'individual' && form.contact_ids.size === 0) {
      setFormError('Please select at least one contact')
      return
    }
    if (form.recipient_mode === 'list' && !form.list_id) {
      setFormError('Please select a contact list')
      return
    }
    setSaving(true)
    try {
      const payload: CampaignCreateInput = {
        name: form.name,
        template_id: form.template_id,
        contact_ids: form.recipient_mode === 'individual' ? Array.from(form.contact_ids) : [],
        list_id: form.recipient_mode === 'list' ? form.list_id : null,
        scheduled_at: form.schedule && form.scheduled_at
          ? new Date(form.scheduled_at).toISOString()
          : null,
      }
      const created = await campaignsService.create(payload)
      setCampaigns((prev) => [created, ...prev])
      setModalOpen(false)
    } catch (err: any) {
      setFormError(err.response?.data?.detail || 'Could not create campaign')
    } finally {
      setSaving(false)
    }
  }

  const handleLaunch = async (c: Campaign) => {
    try {
      const updated = await campaignsService.launch(c.id)
      setCampaigns((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not launch campaign')
    }
  }

  const handleCancel = async (c: Campaign) => {
    try {
      const updated = await campaignsService.cancel(c.id)
      setCampaigns((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not cancel campaign')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await campaignsService.remove(id)
      setCampaigns((prev) => prev.filter((x) => x.id !== id))
    } catch {
      setError('Could not delete campaign')
    } finally {
      setConfirmDeleteId(null)
    }
  }

  const handleApprove = async (c: Campaign) => {
    try {
      const updated = await campaignsService.approve(c.id)
      setCampaigns((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not approve campaign')
    }
  }

  const handleReject = async (id: string) => {
    try {
      const updated = await campaignsService.reject(id, rejectReason)
      setCampaigns((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
      setRejectCampaignId(null)
      setRejectReason('')
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not reject campaign')
    }
  }

  const handleClone = async (c: Campaign) => {
    try {
      const cloned = await campaignsService.clone(c.id)
      setCampaigns((prev) => [cloned, ...prev])
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not clone campaign')
    }
  }

  const canLaunch = (s: CampaignStatus) => s === 'draft' || s === 'scheduled'
  const canCancel = (s: CampaignStatus) => s === 'draft' || s === 'scheduled' || s === 'running'

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Megaphone size={24} className="text-primary-600" />
            Campaigns
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Bulk-send messages using an approved template to a chosen list of contacts.
          </p>
        </div>
        <div className="flex items-center gap-2">
          {campaigns.length > 0 && (
            <button
              onClick={() => campaignsService.exportCsv()}
              className="flex items-center gap-2 rounded-lg border border-gray-300 px-3 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <Download size={16} />
              Export
            </button>
          )}
          <button
            onClick={openCreate}
            className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
          >
            <Plus size={16} />
            New Campaign
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Table */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        {loading ? (
          <div className="p-10 text-center text-sm text-gray-400">Loading campaigns…</div>
        ) : campaigns.length === 0 ? (
          <div className="p-10 text-center">
            <Megaphone size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">No campaigns yet — create your first one.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Name</th>
                  <th className="px-4 py-2.5">Template</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5">Recipients</th>
                  <th className="px-4 py-2.5">Progress</th>
                  <th className="px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {campaigns.map((c) => {
                  const tpl = templateById.get(c.template_id)
                  return (
                    <tr key={c.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3">
                        <p className="font-medium text-gray-900">{c.name}</p>
                        <p className="text-xs text-gray-400 capitalize">{c.channel}</p>
                      </td>
                      <td className="px-4 py-3">
                        <p className="text-sm text-gray-700">{tpl?.name || <span className="text-gray-400 italic">removed</span>}</p>
                        {tpl && <p className="text-xs text-gray-400 capitalize">{tpl.category}</p>}
                      </td>
                      <td className="px-4 py-3">
                        <StatusPill status={c.status} />
                      </td>
                      <td className="px-4 py-3 text-gray-700">{c.total_count}</td>
                      <td className="px-4 py-3">
                        <ProgressBar campaign={c} />
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-1">
                          {confirmDeleteId === c.id ? (
                            <>
                              <button
                                onClick={() => handleDelete(c.id)}
                                className="rounded-lg px-2 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700 transition-colors"
                              >
                                Confirm
                              </button>
                              <button
                                onClick={() => setConfirmDeleteId(null)}
                                className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 transition-colors"
                              >
                                <X size={14} />
                              </button>
                            </>
                          ) : (
                            <>
                              <button
                                onClick={() => setStatsCampaign(c)}
                                className="rounded-lg p-1.5 text-primary-600 hover:bg-primary-50 transition-colors"
                                title="View Stats"
                              >
                                <BarChart3 size={14} />
                              </button>
                              <button
                                onClick={() => handleClone(c)}
                                className="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 transition-colors"
                                title="Duplicate"
                              >
                                <Copy size={14} />
                              </button>
                              {c.status === 'pending_approval' && (
                                <>
                                  <button
                                    onClick={() => handleApprove(c)}
                                    className="rounded-lg p-1.5 text-green-600 hover:bg-green-50 transition-colors"
                                    title="Approve"
                                  >
                                    <ThumbsUp size={14} />
                                  </button>
                                  <button
                                    onClick={() => setRejectCampaignId(c.id)}
                                    className="rounded-lg p-1.5 text-red-600 hover:bg-red-50 transition-colors"
                                    title="Reject"
                                  >
                                    <ThumbsDown size={14} />
                                  </button>
                                </>
                              )}
                              {canLaunch(c.status) && (
                                <button
                                  onClick={() => handleLaunch(c)}
                                  className="rounded-lg p-1.5 text-green-600 hover:bg-green-50 transition-colors"
                                  title="Launch"
                                >
                                  <Play size={14} />
                                </button>
                              )}
                              {canCancel(c.status) && (
                                <button
                                  onClick={() => handleCancel(c)}
                                  className="rounded-lg p-1.5 text-amber-600 hover:bg-amber-50 transition-colors"
                                  title="Cancel"
                                >
                                  <Square size={14} />
                                </button>
                              )}
                              <button
                                onClick={() => setConfirmDeleteId(c.id)}
                                className="rounded-lg p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 transition-colors"
                                title="Delete"
                              >
                                <Trash2 size={14} />
                              </button>
                            </>
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

      {/* Create modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="New Campaign">
        <form onSubmit={handleCreate} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {formError}
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Campaign Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. Diwali Sale Blast"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              <FileText size={12} className="inline mr-1" />
              Template <span className="text-red-500">*</span>
            </label>
            {templates.length === 0 ? (
              <p className="text-xs text-amber-600 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">
                No templates yet. Create one in Templates first.
              </p>
            ) : (
              <select
                required
                value={form.template_id}
                onChange={(e) => setForm({ ...form, template_id: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer bg-white"
              >
                <option value="">Select a template…</option>
                {templates.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name} ({t.category} · {t.channel})
                  </option>
                ))}
              </select>
            )}
            {form.template_id && templateById.get(form.template_id) && (
              <TemplatePreview
                body={templateById.get(form.template_id)!.body}
                channel={templateById.get(form.template_id)!.channel}
                className="mt-2"
              />
            )}
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              <Users size={12} className="inline mr-1" />
              Recipients <span className="text-red-500">*</span>
            </label>
            {/* Mode toggle */}
            <div className="flex rounded-lg border border-gray-200 mb-3 overflow-hidden">
              <button
                type="button"
                onClick={() => setForm({ ...form, recipient_mode: 'individual' })}
                className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium transition-colors ${form.recipient_mode === 'individual' ? 'bg-primary-50 text-primary-700 border-r border-gray-200' : 'text-gray-500 hover:bg-gray-50 border-r border-gray-200'}`}
              >
                <Users size={12} />
                Individual Contacts
              </button>
              <button
                type="button"
                onClick={() => setForm({ ...form, recipient_mode: 'list' })}
                className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-xs font-medium transition-colors ${form.recipient_mode === 'list' ? 'bg-primary-50 text-primary-700' : 'text-gray-500 hover:bg-gray-50'}`}
              >
                <ListChecks size={12} />
                Contact List
              </button>
            </div>

            {form.recipient_mode === 'list' ? (
              <>
                {contactLists.length === 0 ? (
                  <p className="text-xs text-amber-600 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">
                    No contact lists yet. Create one in Contact Lists first.
                  </p>
                ) : (
                  <select
                    value={form.list_id}
                    onChange={(e) => setForm({ ...form, list_id: e.target.value })}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer bg-white"
                  >
                    <option value="">Select a contact list…</option>
                    {contactLists.map((l) => (
                      <option key={l.id} value={l.id}>
                        {l.name} ({l.member_count} members)
                      </option>
                    ))}
                  </select>
                )}
              </>
            ) : (
              <>
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-xs text-gray-500">Select contacts</span>
                  {contacts.length > 0 && (
                    <button
                      type="button"
                      onClick={toggleAllContacts}
                      className="text-xs font-medium text-primary-600 hover:underline"
                    >
                      {form.contact_ids.size === contacts.length ? 'Clear all' : 'Select all'}
                    </button>
                  )}
                </div>
                {contacts.length === 0 ? (
                  <p className="text-xs text-amber-600 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">
                    No contacts yet. Add some in Contacts first.
                  </p>
                ) : (
                  <div className="max-h-40 overflow-y-auto rounded-lg border border-gray-200 divide-y divide-gray-100">
                    {contacts.map((c) => (
                      <label
                        key={c.id}
                        className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer"
                      >
                        <input
                          type="checkbox"
                          checked={form.contact_ids.has(c.id)}
                          onChange={() => toggleContact(c.id)}
                          className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                        />
                        <div className="min-w-0 flex-1">
                          <p className="text-sm text-gray-900 truncate">{c.name || c.phone_number}</p>
                          <p className="text-xs text-gray-400 font-mono">{c.phone_number}</p>
                        </div>
                      </label>
                    ))}
                  </div>
                )}
                <p className="mt-1.5 text-xs text-gray-500">
                  {form.contact_ids.size} of {contacts.length} selected
                </p>
              </>
            )}
          </div>

          <div>
            <label className="flex items-center gap-2 text-xs font-medium text-gray-700 cursor-pointer select-none">
              <input
                type="checkbox"
                checked={form.schedule}
                onChange={(e) => setForm({ ...form, schedule: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              <Calendar size={12} />
              Schedule for later
            </label>
            {form.schedule && (
              <input
                type="datetime-local"
                required
                value={form.scheduled_at}
                onChange={(e) => setForm({ ...form, scheduled_at: e.target.value })}
                className="mt-2 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              />
            )}
          </div>

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving || templates.length === 0 || contacts.length === 0}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
            >
              {saving ? 'Saving...' : 'Create Campaign'}
            </button>
          </div>
        </form>
      </Modal>

      {/* Reject reason modal */}
      <Modal open={!!rejectCampaignId} onClose={() => { setRejectCampaignId(null); setRejectReason('') }} title="Reject Campaign">
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            Provide a reason for rejecting this campaign. The creator will see this reason.
          </p>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Reason</label>
            <textarea
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. Exceeds daily sending limits, unapproved content..."
            />
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => { setRejectCampaignId(null); setRejectReason('') }}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => rejectCampaignId && handleReject(rejectCampaignId)}
              className="flex-1 rounded-lg bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-700 transition-colors"
            >
              Reject Campaign
            </button>
          </div>
        </div>
      </Modal>

      {/* Stats modal */}
      <Modal open={!!statsCampaign} onClose={() => setStatsCampaign(null)} title={`Stats: ${statsCampaign?.name || ''}`}>
        {statsCampaign && (() => {
          const c = statsCampaign
          const stats = [
            { label: 'Total Recipients', value: c.total_count, color: 'text-gray-700' },
            { label: 'Sent', value: c.sent_count, color: 'text-amber-600' },
            { label: 'Delivered', value: c.delivered_count, color: 'text-green-600' },
            { label: 'Read', value: c.read_count, color: 'text-blue-600' },
            { label: 'Failed', value: c.failed_count, color: 'text-red-600' },
            { label: 'Total Cost', value: `₹${(c.total_cost || 0).toFixed(2)}`, color: 'text-purple-600' },
          ]
          const deliveryRate = c.total_count > 0 ? ((c.delivered_count + c.read_count) / c.total_count * 100).toFixed(1) : '0.0'
          return (
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <StatusPill status={c.status} />
                <span className="text-xs text-gray-400 capitalize">{c.channel}</span>
                {c.started_at && <span className="text-xs text-gray-400">Started: {new Date(c.started_at).toLocaleString('en-IN')}</span>}
              </div>
              <div className="grid grid-cols-2 gap-3">
                {stats.map(s => (
                  <div key={s.label} className="rounded-lg border border-gray-200 p-3 text-center">
                    <p className={`text-2xl font-bold ${s.color}`}>{typeof s.value === 'number' ? s.value.toLocaleString() : s.value}</p>
                    <p className="text-xs text-gray-500 mt-0.5">{s.label}</p>
                  </div>
                ))}
                <div className="rounded-lg border border-gray-200 p-3 text-center">
                  <p className="text-2xl font-bold text-primary-600">{deliveryRate}%</p>
                  <p className="text-xs text-gray-500 mt-0.5">Delivery Rate</p>
                </div>
              </div>
              {c.total_count > 0 && (
                <div>
                  <p className="text-xs font-medium text-gray-700 mb-1.5">Progress</p>
                  <div className="flex h-3 w-full overflow-hidden rounded-full bg-gray-100">
                    <div className="bg-blue-500 transition-all" style={{ width: `${(c.read_count / c.total_count) * 100}%` }} title="Read" />
                    <div className="bg-green-500 transition-all" style={{ width: `${(c.delivered_count / c.total_count) * 100}%` }} title="Delivered" />
                    <div className="bg-amber-400 transition-all" style={{ width: `${((c.sent_count - c.delivered_count - c.read_count - c.failed_count) / c.total_count) * 100}%` }} title="Sent" />
                    <div className="bg-red-500 transition-all" style={{ width: `${(c.failed_count / c.total_count) * 100}%` }} title="Failed" />
                  </div>
                  <div className="flex gap-4 mt-2 text-[10px]">
                    <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-blue-500" />Read</span>
                    <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-green-500" />Delivered</span>
                    <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-amber-400" />Sent</span>
                    <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-red-500" />Failed</span>
                  </div>
                </div>
              )}
            </div>
          )
        })()}
      </Modal>
    </div>
  )
}
