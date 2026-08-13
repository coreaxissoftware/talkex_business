import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  Users,
  Plus,
  Search,
  Pencil,
  Trash2,
  Check,
  X,
  Clock,
  ShieldOff,
  Upload,
  FileUp,
  Download,
  ChevronLeft,
  ChevronRight as ChevronRightIcon,
  Filter,
} from 'lucide-react'
import { contactsService } from '../services/contacts'
import api from '../services/api'
import type { Contact, ContactCreateInput, ContactUpdateInput } from '../types/contact'
import Modal from '../components/Modal'

const WINDOW_HOURS = 24

function isWindowOpen(lastInboundAt: string | null): boolean {
  if (!lastInboundAt) return false
  const elapsedMs = Date.now() - new Date(lastInboundAt).getTime()
  return elapsedMs < WINDOW_HOURS * 60 * 60 * 1000
}

interface ContactFormState {
  phone_number: string
  name: string
  email: string
  tags: string
}

const emptyForm: ContactFormState = { phone_number: '', name: '', email: '', tags: '' }

function tagsToArray(input: string): string[] {
  return input
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean)
}

const PAGE_SIZE = 25

export default function Contacts() {
  const [contacts, setContacts] = useState<Contact[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [tagFilter, setTagFilter] = useState('')
  const [page, setPage] = useState(0)

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Contact | null>(null)
  const [form, setForm] = useState<ContactFormState>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  // CSV import
  const [showImport, setShowImport] = useState(false)
  const [importing, setImporting] = useState(false)
  const [importResult, setImportResult] = useState<{ created: number; skipped: number; failed: number; errors: string[] } | null>(null)

  // All unique tags for tag filter dropdown
  const [allTags, setAllTags] = useState<string[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const hasFilters = search || tagFilter
      if (hasFilters) {
        const result = await contactsService.listFiltered({
          search: search || undefined,
          tag: tagFilter || undefined,
          limit: PAGE_SIZE,
          offset: page * PAGE_SIZE,
        })
        setContacts(result.items)
        setTotal(result.total)
      } else {
        const result = await contactsService.listFiltered({
          limit: PAGE_SIZE,
          offset: page * PAGE_SIZE,
        })
        setContacts(result.items)
        setTotal(result.total)
      }
      setError('')
    } catch {
      setError('Could not load contacts. Is the backend running?')
    } finally {
      setLoading(false)
    }
  }, [search, tagFilter, page])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    contactsService.list().then((all) => {
      const tags = new Set<string>()
      all.forEach((c) => c.tags.forEach((t) => tags.add(t)))
      setAllTags(Array.from(tags).sort())
    }).catch(() => {})
  }, [])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const handleSearch = () => {
    setSearch(searchInput)
    setPage(0)
  }

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setFormError('')
    setModalOpen(true)
  }

  const openEdit = (contact: Contact) => {
    setEditing(contact)
    setForm({
      phone_number: contact.phone_number,
      name: contact.name ?? '',
      email: contact.email ?? '',
      tags: contact.tags.join(', '),
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
        const payload: ContactUpdateInput = {
          name: form.name || null,
          email: form.email || null,
          tags: tagsToArray(form.tags),
        }
        const updated = await contactsService.update(editing.id, payload)
        setContacts((prev) => prev.map((c) => (c.id === updated.id ? updated : c)))
      } else {
        const payload: ContactCreateInput = {
          phone_number: form.phone_number,
          name: form.name || null,
          email: form.email || null,
          tags: tagsToArray(form.tags),
        }
        const created = await contactsService.create(payload)
        setContacts((prev) => [created, ...prev])
      }
      setModalOpen(false)
    } catch (err: any) {
      setFormError(err.response?.data?.detail || 'Could not save contact')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await contactsService.remove(id)
      setContacts((prev) => prev.filter((c) => c.id !== id))
    } catch {
      setError('Could not delete contact')
    } finally {
      setConfirmDeleteId(null)
    }
  }

  const handleCSVUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setImporting(true)
    setImportResult(null)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const res = await api.post('/contacts/import-csv', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      setImportResult(res.data)
      await load()
    } catch {
      setImportResult({ created: 0, skipped: 0, failed: 1, errors: ['Upload failed'] })
    } finally {
      setImporting(false)
      e.target.value = ''
    }
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Users size={24} className="text-primary-600" />
            Contacts
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Import, tag, and segment your contacts — shared across all channels.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => contactsService.exportCSV()}
            className="flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
          >
            <Download size={16} />
            Export
          </button>
          <button
            onClick={() => { setShowImport(true); setImportResult(null) }}
            className="flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
          >
            <Upload size={16} />
            Import CSV
          </button>
          <button
            onClick={openCreate}
            className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
          >
            <Plus size={16} />
            Add Contact
          </button>
        </div>
      </div>

      {/* Search + Tag filter */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative max-w-sm flex-1 min-w-[200px]">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Search name, phone, or email"
            className="w-full rounded-lg border border-gray-300 pl-9 pr-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
          />
        </div>
        <button
          onClick={handleSearch}
          className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700"
        >
          Search
        </button>
        {allTags.length > 0 && (
          <div className="flex items-center gap-1.5">
            <Filter size={14} className="text-gray-400" />
            <select
              value={tagFilter}
              onChange={(e) => { setTagFilter(e.target.value); setPage(0) }}
              className="rounded-lg border border-gray-300 px-2 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
            >
              <option value="">All tags</option>
              {allTags.map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </div>
        )}
        {(search || tagFilter) && (
          <button
            onClick={() => { setSearch(''); setSearchInput(''); setTagFilter(''); setPage(0) }}
            className="flex items-center gap-1 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
          >
            <X size={14} /> Clear
          </button>
        )}
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Table */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        {loading ? (
          <div className="p-10 text-center text-sm text-gray-400">Loading contacts…</div>
        ) : contacts.length === 0 ? (
          <div className="p-10 text-center">
            <Users size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">
              {total === 0 && !search && !tagFilter ? 'No contacts yet — add your first one.' : 'No matches.'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Contact</th>
                  <th className="px-4 py-2.5">Phone</th>
                  <th className="px-4 py-2.5">Tags</th>
                  <th className="px-4 py-2.5">Consent</th>
                  <th className="px-4 py-2.5">Window</th>
                  <th className="px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {contacts.map((c) => {
                  const windowOpen = isWindowOpen(c.last_inbound_at)
                  return (
                    <tr key={c.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3">
                        <p className="font-medium text-gray-900">{c.name || '—'}</p>
                        <p className="text-xs text-gray-400">{c.email || 'No email'}</p>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-700">
                        {c.phone_number}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex flex-wrap gap-1">
                          {c.tags.length === 0 && <span className="text-xs text-gray-300">—</span>}
                          {c.tags.map((t) => (
                            <span
                              key={t}
                              className="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700"
                            >
                              {t}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={async () => {
                            try {
                              const updated = await contactsService.toggleOptIn(c.id, !c.opted_in)
                              setContacts(prev => prev.map(ct => ct.id === updated.id ? updated : ct))
                            } catch { /* ignore */ }
                          }}
                          className="cursor-pointer"
                          title={c.opted_in ? 'Click to revoke consent' : 'Click to mark opted in'}
                        >
                          {c.opted_in ? (
                            <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700 hover:bg-green-100 transition-colors">
                              <Check size={12} /> Opted in
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500 hover:bg-gray-200 transition-colors">
                              <ShieldOff size={12} /> No consent
                            </span>
                          )}
                        </button>
                      </td>
                      <td className="px-4 py-3">
                        {windowOpen ? (
                          <span className="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2 py-0.5 text-xs font-semibold text-blue-700">
                            <Clock size={12} /> Open (24h)
                          </span>
                        ) : (
                          <span className="text-xs text-gray-400">Closed</span>
                        )}
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
                                onClick={() => openEdit(c)}
                                className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 transition-colors"
                                title="Edit"
                              >
                                <Pencil size={14} />
                              </button>
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

      {total > 0 && (
        <div className="flex items-center justify-between">
          <p className="text-xs text-gray-400">
            Showing {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, total)} of {total} contacts
          </p>
          {totalPages > 1 && (
            <div className="flex items-center gap-1">
              <button
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
                className="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={16} />
              </button>
              <span className="text-xs text-gray-500 px-2">
                Page {page + 1} of {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                disabled={page >= totalPages - 1}
                className="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronRightIcon size={16} />
              </button>
            </div>
          )}
        </div>
      )}

      {/* Add / Edit modal */}
      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title={editing ? 'Edit Contact' : 'Add Contact'}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {formError}
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Phone Number {!editing && <span className="text-red-500">*</span>}
            </label>
            <input
              type="tel"
              required
              disabled={!!editing}
              value={form.phone_number}
              onChange={(e) => setForm({ ...form, phone_number: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none disabled:bg-gray-50 disabled:text-gray-400"
              placeholder="+919876543210"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="Contact name"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="contact@example.com"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Tags <span className="text-gray-400 font-normal">(comma separated)</span>
            </label>
            <input
              type="text"
              value={form.tags}
              onChange={(e) => setForm({ ...form, tags: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="vip, lead, mumbai"
            />
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
              disabled={saving}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
            >
              {saving ? 'Saving...' : editing ? 'Save Changes' : 'Add Contact'}
            </button>
          </div>
        </form>
      </Modal>

      {/* CSV Import Modal */}
      {showImport && (
        <Modal title="Import Contacts from CSV" onClose={() => setShowImport(false)}>
          <div className="space-y-4">
            <div className="rounded-lg bg-blue-50 border border-blue-200 p-3 text-xs text-blue-700">
              <p className="font-semibold mb-1">CSV Format</p>
              <p>Required column: <code className="bg-blue-100 px-1 rounded">phone</code> or <code className="bg-blue-100 px-1 rounded">phone_number</code></p>
              <p>Optional columns: <code className="bg-blue-100 px-1 rounded">name</code>, <code className="bg-blue-100 px-1 rounded">email</code>, <code className="bg-blue-100 px-1 rounded">tags</code> (semicolon-separated)</p>
              <p className="mt-1 text-blue-500">Example: phone,name,email,tags</p>
              <p className="text-blue-500">+919876543210,Rahul Sharma,rahul@example.com,vip;mumbai</p>
            </div>

            <label className="flex flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed border-gray-300 bg-gray-50 p-8 cursor-pointer hover:border-primary-400 hover:bg-primary-50/30 transition-colors">
              <FileUp size={32} className="text-gray-400" />
              <span className="text-sm font-medium text-gray-600">
                {importing ? 'Uploading...' : 'Click to select CSV file'}
              </span>
              <input
                type="file"
                accept=".csv"
                className="hidden"
                onChange={handleCSVUpload}
                disabled={importing}
              />
            </label>

            {importResult && (
              <div className={`rounded-lg border p-3 text-sm ${importResult.failed > 0 ? 'bg-yellow-50 border-yellow-200' : 'bg-green-50 border-green-200'}`}>
                <p className="font-semibold text-gray-900">Import Complete</p>
                <div className="flex gap-4 mt-1 text-xs">
                  <span className="text-green-700">{importResult.created} created</span>
                  <span className="text-gray-500">{importResult.skipped} skipped</span>
                  <span className="text-red-600">{importResult.failed} failed</span>
                </div>
                {importResult.errors.length > 0 && (
                  <div className="mt-2 text-xs text-red-600 max-h-24 overflow-y-auto">
                    {importResult.errors.map((e, i) => <p key={i}>{e}</p>)}
                  </div>
                )}
              </div>
            )}

            <div className="flex justify-end pt-2">
              <button
                onClick={() => setShowImport(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
