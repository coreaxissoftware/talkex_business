import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  FileText,
  Plus,
  Search,
  Pencil,
  Trash2,
  X,
} from 'lucide-react'
import { templatesService } from '../services/templates'
import type {
  MessageTemplate,
  TemplateCategory,
  TemplateStatus,
  TemplateCreateInput,
  TemplateUpdateInput,
  TemplateButton,
  TemplateListRow,
} from '../types/template'
import Modal from '../components/Modal'
import TemplatePreview from '../components/TemplatePreview'

const CATEGORY_STYLES: Record<TemplateCategory, string> = {
  marketing: 'bg-purple-50 text-purple-700 border-purple-200',
  utility: 'bg-blue-50 text-blue-700 border-blue-200',
  authentication: 'bg-amber-50 text-amber-700 border-amber-200',
}

const STATUS_STYLES: Record<TemplateStatus, string> = {
  draft: 'bg-gray-100 text-gray-600',
  pending_review: 'bg-amber-50 text-amber-700',
  approved: 'bg-green-50 text-green-700',
  rejected: 'bg-red-50 text-red-700',
}

const STATUS_LABELS: Record<TemplateStatus, string> = {
  draft: 'Draft',
  pending_review: 'Pending Review',
  approved: 'Approved',
  rejected: 'Rejected',
}

interface FormState {
  name: string
  category: TemplateCategory
  channel: string
  body: string
  variables: string
  status: TemplateStatus
  // Interactive & media (WhatsApp-only)
  header: string
  footer: string
  mediaType: string
  mediaURL: string
  buttons: TemplateButton[]
  listRows: TemplateListRow[]
}

const emptyForm: FormState = {
  name: '',
  category: 'utility',
  channel: 'talkex',
  body: '',
  variables: '',
  status: 'draft',
  header: '',
  footer: '',
  mediaType: '',
  mediaURL: '',
  buttons: [],
  listRows: [],
}

function varsToArray(input: string): string[] {
  return input
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean)
}

export default function Templates() {
  const [templates, setTemplates] = useState<MessageTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [categoryFilter, setCategoryFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<MessageTemplate | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await templatesService.list()
      setTemplates(data)
      setError('')
    } catch {
      setError('Could not load templates. Is the backend running?')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setFormError('')
    setModalOpen(true)
  }

  const openEdit = (t: MessageTemplate) => {
    setEditing(t)
    setForm({
      name: t.name,
      category: t.category,
      channel: t.channel,
      body: t.body,
      variables: t.variables.join(', '),
      status: t.status,
      header: t.header || '',
      footer: t.footer || '',
      mediaType: t.media_type || '',
      mediaURL: t.media_url || '',
      buttons: t.buttons || [],
      listRows: t.list_rows || [],
    })
    setFormError('')
    setModalOpen(true)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFormError('')
    setSaving(true)
    try {
      // Only include WhatsApp-specific interactive fields when relevant.
      const interactive = form.channel === 'whatsapp' ? {
        header: form.header,
        footer: form.footer,
        media_type: form.mediaType,
        media_url: form.mediaURL,
        buttons: form.buttons,
        list_rows: form.listRows,
      } : {}
      if (editing) {
        const payload: TemplateUpdateInput = {
          name: form.name,
          body: form.body,
          variables: varsToArray(form.variables),
          status: form.status,
          ...interactive,
        }
        const updated = await templatesService.update(editing.id, payload)
        setTemplates((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
      } else {
        const payload: TemplateCreateInput = {
          name: form.name,
          category: form.category,
          channel: form.channel,
          body: form.body,
          variables: varsToArray(form.variables),
          ...interactive,
        }
        const created = await templatesService.create(payload)
        setTemplates((prev) => [created, ...prev])
      }
      setModalOpen(false)
    } catch (err: any) {
      setFormError(err.response?.data?.detail || 'Could not save template')
    } finally {
      setSaving(false)
    }
  }

  const handleSubmitToMeta = async (id: string) => {
    if (!confirm('Submit this template to Meta for approval? You cannot edit it again until Meta responds.')) return
    try {
      const updated = await templatesService.submitToMeta(id)
      setTemplates(prev => prev.map(t => t.id === id ? updated : t))
    } catch (err: any) {
      alert('Submission failed: ' + (err.response?.data?.detail || err.message))
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await templatesService.remove(id)
      setTemplates((prev) => prev.filter((t) => t.id !== id))
    } catch {
      setError('Could not delete template')
    } finally {
      setConfirmDeleteId(null)
    }
  }

  const filtered = templates.filter((t) => {
    if (categoryFilter && t.category !== categoryFilter) return false
    if (statusFilter && t.status !== statusFilter) return false
    if (search && !t.name.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <FileText size={24} className="text-primary-600" />
            Templates
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Reusable message templates — category drives per-message pricing.
          </p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} />
          New Template
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 min-w-[200px] max-w-sm">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name"
            className="w-full rounded-lg border border-gray-300 pl-9 pr-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
          />
        </div>
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer"
        >
          <option value="">All categories</option>
          <option value="marketing">Marketing</option>
          <option value="utility">Utility</option>
          <option value="authentication">Authentication</option>
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer"
        >
          <option value="">All statuses</option>
          <option value="draft">Draft</option>
          <option value="pending_review">Pending Review</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Grid of template cards */}
      {loading ? (
        <div className="p-10 text-center text-sm text-gray-400">Loading templates…</div>
      ) : filtered.length === 0 ? (
        <div className="rounded-xl border border-gray-200 bg-white p-10 text-center">
          <FileText size={32} className="mx-auto text-gray-300 mb-2" />
          <p className="text-sm text-gray-500">
            {templates.length === 0 ? 'No templates yet — create your first one.' : 'No matches.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((t) => (
            <div
              key={t.id}
              className="rounded-xl border border-gray-200 bg-white p-4 flex flex-col gap-3"
            >
              <div className="flex items-start justify-between">
                <div>
                  <p className="font-semibold text-gray-900">{t.name}</p>
                  <p className="text-xs text-gray-400 mt-0.5 capitalize">{t.channel}</p>
                </div>
                <div className="flex items-center gap-1">
                  {confirmDeleteId === t.id ? (
                    <>
                      <button
                        onClick={() => handleDelete(t.id)}
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
                      {t.channel === 'whatsapp' && t.status === 'draft' && (
                        <button
                          onClick={() => handleSubmitToMeta(t.id)}
                          className="rounded-lg px-2 py-1 text-[10px] font-semibold text-white bg-green-600 hover:bg-green-700"
                          title="Submit to Meta for WhatsApp approval"
                        >
                          Submit
                        </button>
                      )}
                      <button
                        onClick={() => openEdit(t)}
                        className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 transition-colors"
                        title="Edit"
                      >
                        <Pencil size={14} />
                      </button>
                      <button
                        onClick={() => setConfirmDeleteId(t.id)}
                        className="rounded-lg p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 transition-colors"
                        title="Delete"
                      >
                        <Trash2 size={14} />
                      </button>
                    </>
                  )}
                </div>
              </div>

              <p className="text-sm text-gray-600 line-clamp-3 bg-gray-50 rounded-lg p-2.5">
                {t.body}
              </p>

              {t.variables.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {t.variables.map((v) => (
                    <span
                      key={v}
                      className="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-mono text-gray-500"
                    >
                      {`{{${v}}}`}
                    </span>
                  ))}
                </div>
              )}

              <div className="flex items-center gap-2 pt-1">
                <span
                  className={`rounded-full border px-2 py-0.5 text-xs font-medium ${CATEGORY_STYLES[t.category]}`}
                >
                  {t.category}
                </span>
                <span
                  className={`rounded-full px-2 py-0.5 text-xs font-semibold ${STATUS_STYLES[t.status]}`}
                >
                  {STATUS_LABELS[t.status]}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {templates.length > 0 && (
        <p className="text-xs text-gray-400 text-center">
          Showing {filtered.length} of {templates.length} templates
        </p>
      )}

      {/* Add / Edit modal */}
      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title={editing ? 'Edit Template' : 'New Template'}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {formError}
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Name *</label>
            <input
              type="text"
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="order_confirmation"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Category *</label>
              <select
                required
                disabled={!!editing}
                value={form.category}
                onChange={(e) =>
                  setForm({ ...form, category: e.target.value as TemplateCategory })
                }
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer disabled:bg-gray-50 disabled:text-gray-400"
              >
                <option value="marketing">Marketing</option>
                <option value="utility">Utility</option>
                <option value="authentication">Authentication</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Channel *</label>
              <select
                required
                disabled={!!editing}
                value={form.channel}
                onChange={(e) => setForm({ ...form, channel: e.target.value })}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer disabled:bg-gray-50 disabled:text-gray-400"
              >
                <option value="talkex">TalkEx</option>
                <option value="whatsapp">WhatsApp</option>
                <option value="telegram">Telegram</option>
                <option value="email">Email</option>
                <option value="sms">SMS</option>
                <option value="rcs">RCS</option>
                <option value="instagram">Instagram</option>
                <option value="messenger">Messenger</option>
              </select>
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Body *</label>
            <textarea
              required
              rows={4}
              value={form.body}
              onChange={(e) => setForm({ ...form, body: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none resize-none"
              placeholder="Hi {{1}}, your order #{{2}} has been confirmed."
            />
            {form.body && (
              <TemplatePreview body={form.body} channel={form.channel} className="mt-2" />
            )}
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Variables <span className="text-gray-400 font-normal">(comma separated, e.g. 1, 2)</span>
            </label>
            <input
              type="text"
              value={form.variables}
              onChange={(e) => setForm({ ...form, variables: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="1, 2"
            />
          </div>

          {form.channel === 'whatsapp' && (
            <div className="space-y-3 border-t border-gray-100 pt-3">
              <p className="text-[10px] uppercase tracking-wider text-gray-400">WhatsApp interactive & media</p>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Header (short)</label>
                  <input value={form.header} onChange={e => setForm({...form, header: e.target.value})} maxLength={60}
                    placeholder="e.g. Order shipped"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Footer</label>
                  <input value={form.footer} onChange={e => setForm({...form, footer: e.target.value})} maxLength={60}
                    placeholder="e.g. Powered by TalkEx"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500" />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Media type</label>
                  <select value={form.mediaType} onChange={e => setForm({...form, mediaType: e.target.value})}
                    className="w-full rounded-lg border border-gray-300 px-2 py-2 text-sm outline-none cursor-pointer">
                    <option value="">(none)</option>
                    <option value="image">Image</option>
                    <option value="video">Video</option>
                    <option value="document">Document</option>
                    <option value="audio">Audio</option>
                  </select>
                </div>
                <div className="col-span-2">
                  <label className="block text-xs font-medium text-gray-700 mb-1.5">Media URL</label>
                  <input value={form.mediaURL} onChange={e => setForm({...form, mediaURL: e.target.value})}
                    placeholder="https://…"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500" />
                </div>
              </div>

              {/* Buttons editor (max 3) */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-xs font-medium text-gray-700">Buttons ({form.buttons.length}/3)</label>
                  {form.buttons.length < 3 && (
                    <button type="button"
                      onClick={() => setForm({...form, buttons: [...form.buttons, {type: 'quick_reply', text: ''}]})}
                      className="text-[11px] font-medium text-primary-600 hover:text-primary-700">
                      + Add button
                    </button>
                  )}
                </div>
                <div className="space-y-1.5">
                  {form.buttons.map((b, i) => (
                    <div key={i} className="flex items-center gap-1.5 rounded-lg border border-gray-200 p-1.5">
                      <select value={b.type}
                        onChange={e => {
                          const copy = [...form.buttons]
                          copy[i] = {...b, type: e.target.value as TemplateButton['type']}
                          setForm({...form, buttons: copy})
                        }}
                        className="rounded border border-gray-200 px-1.5 py-1 text-[11px] cursor-pointer">
                        <option value="quick_reply">Quick reply</option>
                        <option value="url">URL</option>
                        <option value="phone">Phone</option>
                      </select>
                      <input value={b.text} placeholder="Button text" maxLength={20}
                        onChange={e => {
                          const copy = [...form.buttons]
                          copy[i] = {...b, text: e.target.value}
                          setForm({...form, buttons: copy})
                        }}
                        className="flex-1 rounded border border-gray-200 px-2 py-1 text-xs outline-none focus:border-primary-500" />
                      {b.type === 'url' && (
                        <input value={b.url || ''} placeholder="https://…"
                          onChange={e => {
                            const copy = [...form.buttons]
                            copy[i] = {...b, url: e.target.value}
                            setForm({...form, buttons: copy})
                          }}
                          className="w-32 rounded border border-gray-200 px-2 py-1 text-xs outline-none focus:border-primary-500" />
                      )}
                      {b.type === 'phone' && (
                        <input value={b.phone || ''} placeholder="+91…"
                          onChange={e => {
                            const copy = [...form.buttons]
                            copy[i] = {...b, phone: e.target.value}
                            setForm({...form, buttons: copy})
                          }}
                          className="w-32 rounded border border-gray-200 px-2 py-1 text-xs outline-none focus:border-primary-500" />
                      )}
                      <button type="button"
                        onClick={() => setForm({...form, buttons: form.buttons.filter((_, j) => j !== i)})}
                        className="p-1 text-gray-400 hover:text-red-600" title="Remove">
                        <X size={12} />
                      </button>
                    </div>
                  ))}
                </div>
              </div>

              {/* List rows editor (mutually exclusive with buttons in Meta, but we allow both stored) */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-xs font-medium text-gray-700">List rows ({form.listRows.length})</label>
                  <button type="button"
                    onClick={() => setForm({...form, listRows: [...form.listRows, {id: '', title: '', description: ''}]})}
                    className="text-[11px] font-medium text-primary-600 hover:text-primary-700">
                    + Add row
                  </button>
                </div>
                <div className="space-y-1.5">
                  {form.listRows.map((r, i) => (
                    <div key={i} className="flex items-center gap-1.5 rounded-lg border border-gray-200 p-1.5">
                      <input value={r.id} placeholder="row_id"
                        onChange={e => {
                          const copy = [...form.listRows]
                          copy[i] = {...r, id: e.target.value}
                          setForm({...form, listRows: copy})
                        }}
                        className="w-24 rounded border border-gray-200 px-2 py-1 text-xs font-mono outline-none focus:border-primary-500" />
                      <input value={r.title} placeholder="Title" maxLength={24}
                        onChange={e => {
                          const copy = [...form.listRows]
                          copy[i] = {...r, title: e.target.value}
                          setForm({...form, listRows: copy})
                        }}
                        className="flex-1 rounded border border-gray-200 px-2 py-1 text-xs outline-none focus:border-primary-500" />
                      <input value={r.description || ''} placeholder="Description" maxLength={72}
                        onChange={e => {
                          const copy = [...form.listRows]
                          copy[i] = {...r, description: e.target.value}
                          setForm({...form, listRows: copy})
                        }}
                        className="flex-1 rounded border border-gray-200 px-2 py-1 text-xs outline-none focus:border-primary-500" />
                      <button type="button"
                        onClick={() => setForm({...form, listRows: form.listRows.filter((_, j) => j !== i)})}
                        className="p-1 text-gray-400 hover:text-red-600" title="Remove">
                        <X size={12} />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {editing && (
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Status</label>
              <select
                value={form.status}
                onChange={(e) => setForm({ ...form, status: e.target.value as TemplateStatus })}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer"
              >
                <option value="draft">Draft</option>
                <option value="pending_review">Pending Review</option>
                <option value="approved">Approved</option>
                <option value="rejected">Rejected</option>
              </select>
            </div>
          )}

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
              disabled={saving}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
            >
              {saving ? 'Saving...' : editing ? 'Save Changes' : 'Create Template'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
