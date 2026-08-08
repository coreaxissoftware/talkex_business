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
} from 'lucide-react'
import { contactsService } from '../services/contacts'
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

export default function Contacts() {
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Contact | null>(null)
  const [form, setForm] = useState<ContactFormState>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await contactsService.list()
      setContacts(data)
      setError('')
    } catch {
      setError('Could not load contacts. Is the backend running?')
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

  const filtered = contacts.filter((c) => {
    if (!search) return true
    const q = search.toLowerCase()
    return (
      c.phone_number.toLowerCase().includes(q) ||
      c.name?.toLowerCase().includes(q) ||
      c.email?.toLowerCase().includes(q) ||
      c.tags.some((t) => t.toLowerCase().includes(q))
    )
  })

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
        <button
          onClick={openCreate}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} />
          Add Contact
        </button>
      </div>

      {/* Search */}
      <div className="relative max-w-sm">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search name, phone, email, or tag"
          className="w-full rounded-lg border border-gray-300 pl-9 pr-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
        />
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
        ) : filtered.length === 0 ? (
          <div className="p-10 text-center">
            <Users size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">
              {contacts.length === 0 ? 'No contacts yet — add your first one.' : 'No matches.'}
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
                {filtered.map((c) => {
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
                        {c.opted_in ? (
                          <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700">
                            <Check size={12} /> Opted in
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">
                            <ShieldOff size={12} /> No consent
                          </span>
                        )}
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

      {contacts.length > 0 && (
        <p className="text-xs text-gray-400 text-center">
          Showing {filtered.length} of {contacts.length} contacts
        </p>
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
    </div>
  )
}
