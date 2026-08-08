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
} from '../types/template'
import Modal from '../components/Modal'

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
}

const emptyForm: FormState = {
  name: '',
  category: 'utility',
  channel: 'talkex',
  body: '',
  variables: '',
  status: 'draft',
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
    })
    setFormError('')
    setModalOpen(true)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFormError('')
    setSaving(true)
    try {
      if (editing) {
        const payload: TemplateUpdateInput = {
          name: form.name,
          body: form.body,
          variables: varsToArray(form.variables),
          status: form.status,
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
