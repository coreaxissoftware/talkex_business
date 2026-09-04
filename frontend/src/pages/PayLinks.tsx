import { useEffect, useState, type FormEvent } from 'react'
import { CreditCard, Plus, IndianRupee, Copy, ExternalLink, X, Check, Clock, Ban } from 'lucide-react'
import { paylinksService, type PayLink } from '../services/paylinks'
import { contactsService } from '../services/contacts'
import type { Contact } from '../types/contact'

const STATUS_CHIP: Record<PayLink['status'], string> = {
  created: 'bg-gray-100 text-gray-700',
  sent: 'bg-blue-50 text-blue-700',
  paid: 'bg-green-50 text-green-700',
  expired: 'bg-amber-50 text-amber-700',
  cancelled: 'bg-red-50 text-red-700',
}

const STATUS_ICON: Record<PayLink['status'], any> = {
  created: Clock,
  sent: Clock,
  paid: Check,
  expired: Ban,
  cancelled: Ban,
}

export default function PayLinks() {
  const [items, setItems] = useState<PayLink[]>([])
  const [contacts, setContacts] = useState<Contact[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [copiedId, setCopiedId] = useState('')

  // form state
  const [contactId, setContactId] = useState('')
  const [amountRupees, setAmountRupees] = useState('')
  const [description, setDescription] = useState('')
  const [expireHours, setExpireHours] = useState(24)
  const [submitting, setSubmitting] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const [links, cts] = await Promise.all([
        paylinksService.list(),
        contactsService.list(),
      ])
      setItems(links)
      setContacts(cts)
      setError('')
    } catch {
      setError('Could not load payment links.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const totalPaidPaise = items.filter(l => l.status === 'paid').reduce((s, l) => s + l.amount_paise, 0)
  const openLinks = items.filter(l => l.status === 'created' || l.status === 'sent').length
  const conversionRate = items.length ? (items.filter(l => l.status === 'paid').length / items.length) * 100 : 0

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    const amt = parseFloat(amountRupees)
    if (!contactId || isNaN(amt) || amt <= 0 || !description) return
    setSubmitting(true)
    try {
      await paylinksService.create({
        contact_id: contactId,
        amount_paise: Math.round(amt * 100),
        description,
        expire_hours: expireHours,
      })
      setContactId('')
      setAmountRupees('')
      setDescription('')
      setExpireHours(24)
      setShowForm(false)
      load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not create payment link.')
    } finally {
      setSubmitting(false)
    }
  }

  const copy = async (id: string, url: string) => {
    try {
      await navigator.clipboard.writeText(url)
      setCopiedId(id)
      setTimeout(() => setCopiedId(''), 2000)
    } catch {}
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <CreditCard size={24} className="text-primary-600" /> Payment links
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Razorpay Quick Links — send a URL in chat, get paid.
          </p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
        >
          <Plus size={16} /> New link
        </button>
      </div>

      {/* KPI row */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-6">
        <Kpi label="Collected" value={`₹${(totalPaidPaise / 100).toLocaleString('en-IN', { maximumFractionDigits: 2 })}`} tone="success" />
        <Kpi label="Open links" value={openLinks.toLocaleString('en-IN')} />
        <Kpi label="Conversion" value={`${conversionRate.toFixed(1)}%`} tone="primary" />
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-16 text-gray-500">Loading…</div>
      ) : items.length === 0 ? (
        <div className="text-center py-16 border-2 border-dashed border-gray-200 rounded-xl">
          <CreditCard size={40} className="mx-auto text-gray-300 mb-3" />
          <p className="text-gray-500">No payment links yet. Create one to collect your first payment.</p>
        </div>
      ) : (
        <div className="border border-gray-200 rounded-xl bg-white overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wider text-gray-500 border-b border-gray-200 bg-gray-50">
                  <th className="px-4 py-2.5">Description</th>
                  <th className="px-4 py-2.5 text-right">Amount</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5">Created</th>
                  <th className="px-4 py-2.5">Expires</th>
                  <th className="px-4 py-2.5">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {items.map((l) => {
                  const Icon = STATUS_ICON[l.status]
                  return (
                    <tr key={l.id} className="hover:bg-gray-50">
                      <td className="px-4 py-2.5">
                        <div className="font-medium text-gray-900 line-clamp-1">{l.description}</div>
                        {l.simulated && (
                          <div className="text-[10px] text-amber-600 mt-0.5">SIMULATED (dev mode)</div>
                        )}
                      </td>
                      <td className="px-4 py-2.5 text-right tabular-nums font-semibold">
                        ₹{(l.amount_paise / 100).toLocaleString('en-IN', { maximumFractionDigits: 2 })}
                      </td>
                      <td className="px-4 py-2.5">
                        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_CHIP[l.status]}`}>
                          <Icon size={11} /> {l.status}
                        </span>
                      </td>
                      <td className="px-4 py-2.5 text-xs text-gray-500">
                        {new Date(l.created_at).toLocaleDateString('en-IN')}
                      </td>
                      <td className="px-4 py-2.5 text-xs text-gray-500">
                        {l.expires_at ? new Date(l.expires_at).toLocaleString('en-IN', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }) : '—'}
                      </td>
                      <td className="px-4 py-2.5">
                        <div className="flex gap-1">
                          <button
                            onClick={() => copy(l.id, l.url)}
                            title="Copy link"
                            className="text-xs px-2 py-1 rounded text-gray-600 hover:bg-gray-100 flex items-center gap-1"
                          >
                            <Copy size={11} /> {copiedId === l.id ? 'Copied!' : 'Copy'}
                          </button>
                          <a
                            href={l.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            title="Open in new tab"
                            className="text-xs px-2 py-1 rounded text-primary-600 hover:bg-primary-50 flex items-center gap-1"
                          >
                            <ExternalLink size={11} /> Open
                          </a>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showForm && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="font-semibold text-gray-900">Create payment link</h2>
              <button onClick={() => setShowForm(false)} className="text-gray-400 hover:text-gray-600">
                <X size={18} />
              </button>
            </div>
            <form onSubmit={submit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Contact <span className="text-red-500">*</span>
                </label>
                <select
                  required
                  value={contactId}
                  onChange={(e) => setContactId(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                >
                  <option value="">Select contact…</option>
                  {contacts.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name || c.phone_number}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Amount (₹) <span className="text-red-500">*</span>
                </label>
                <div className="relative">
                  <IndianRupee size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    required
                    type="number"
                    min="1"
                    step="0.01"
                    value={amountRupees}
                    onChange={(e) => setAmountRupees(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 pl-8 pr-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                    placeholder="500.00"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  What is this for? <span className="text-red-500">*</span>
                </label>
                <input
                  required
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                  placeholder="Order #1234 — silk saree"
                  maxLength={500}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Expire after
                </label>
                <select
                  value={expireHours}
                  onChange={(e) => setExpireHours(parseInt(e.target.value, 10))}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                >
                  <option value={1}>1 hour</option>
                  <option value={6}>6 hours</option>
                  <option value={24}>24 hours</option>
                  <option value={72}>3 days</option>
                  <option value={168}>7 days</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="px-4 py-2 text-sm rounded-lg text-gray-600 hover:bg-gray-100"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-4 py-2 text-sm rounded-lg bg-primary-600 text-white hover:bg-primary-700 font-semibold disabled:opacity-50"
                >
                  {submitting ? 'Creating…' : 'Create link'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

function Kpi({ label, value, tone = 'default' }: { label: string; value: string; tone?: 'default' | 'primary' | 'success' }) {
  const bg = { default: 'bg-white', primary: 'bg-primary-50', success: 'bg-green-50' }[tone]
  return (
    <div className={`border border-gray-200 rounded-xl p-4 ${bg}`}>
      <div className="text-xs text-gray-500 mb-1">{label}</div>
      <div className="text-xl font-bold text-gray-900 tabular-nums">{value}</div>
    </div>
  )
}
