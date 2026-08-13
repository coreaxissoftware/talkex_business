import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  HelpCircle,
  Plus,
  ChevronDown,
  MessageSquare,
  Book,
  Mail,
  LifeBuoy,
  ExternalLink,
} from 'lucide-react'
import { supportService } from '../services/support'
import type { Ticket, TicketCreateInput, TicketPriority } from '../types/support'
import Modal from '../components/Modal'

interface FaqItem {
  q: string
  a: string
}

const FAQ: FaqItem[] = [
  {
    q: 'What is the 24-hour customer service window?',
    a: "When a contact sends you an inbound message, a 24-hour window opens during which you can reply freely with any text. Outside that window, outbound sends must use an approved template — this matches the WhatsApp Business platform rule and applies to every channel in TalkEx Business.",
  },
  {
    q: 'How are my messages billed?',
    a: 'Each plan includes a monthly quota of messages. Anything beyond that quota is charged per-message at the overage rate shown on the Billing page. Template category (marketing/utility/authentication) will drive per-template pricing in a later release.',
  },
  {
    q: 'How do I create and use an API key?',
    a: 'Head to Developers → Create API Key. Give it a name (e.g. "Production Server"), and copy the plaintext key — it is shown exactly once. Send it in the Authorization header as `Bearer <key>` to authenticate server-to-server requests.',
  },
  {
    q: 'Can I bulk-import contacts?',
    a: 'CSV import is on the roadmap. For now, use POST /contacts from the API — one contact per request — or add contacts manually from the Contacts page.',
  },
  {
    q: 'What happens when I switch plans?',
    a: 'Your plan changes take effect immediately. A new invoice is issued for the new plan\'s monthly fee, and the current billing period resets. Included-message counters go back to zero at the switch.',
  },
  {
    q: 'How do automation rules match inbound messages?',
    a: 'Each rule has one or more keywords and a match type — Contains (default), Starts with, or Exact. Matching is case-insensitive. Rules are evaluated in creation order and the first match wins, so put more specific rules first.',
  },
  {
    q: 'Where can I see failed requests?',
    a: 'Automation & Dev → Logs shows every API request from your account with method, path, status, latency, and the response body for anything that returned an error.',
  },
]

interface DocLink {
  title: string
  href: string
  description: string
}

const DOCS: DocLink[] = [
  { title: 'Getting Started', href: '/', description: 'Set up your first channel and send your first message.' },
  { title: 'API Reference', href: '/developers', description: 'Endpoints, auth, and example requests.' },
  { title: 'Contacts & Segments', href: '/contacts', description: 'Import, tag, and manage recipients.' },
  { title: 'Templates', href: '/templates', description: 'Author reusable messages and categories.' },
  { title: 'Campaigns', href: '/campaigns', description: 'Schedule and track bulk sends.' },
  { title: 'Analytics', href: '/analytics', description: 'Delivery rate, volume, and channel mix.' },
]

const PRIORITIES: { id: TicketPriority; label: string; color: string }[] = [
  { id: 'low', label: 'Low', color: 'text-gray-500' },
  { id: 'normal', label: 'Normal', color: 'text-blue-600' },
  { id: 'high', label: 'High', color: 'text-amber-600' },
  { id: 'urgent', label: 'Urgent', color: 'text-red-600' },
]

function StatusPill({ status }: { status: string }) {
  const style =
    status === 'resolved'
      ? 'bg-green-50 text-green-700'
      : status === 'in_progress'
      ? 'bg-blue-50 text-blue-700'
      : 'bg-amber-50 text-amber-700'
  return (
    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold capitalize ${style}`}>
      {status.replace('_', ' ')}
    </span>
  )
}

export default function Support() {
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')
  const [priority, setPriority] = useState<TicketPriority>('normal')
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const [openFaq, setOpenFaq] = useState<number | null>(0)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await supportService.list()
      setTickets(data)
      setError('')
    } catch {
      setError('Could not load your tickets.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFormError('')
    setSaving(true)
    try {
      const payload: TicketCreateInput = { subject, body, priority }
      const created = await supportService.create(payload)
      setTickets((prev) => [created, ...prev])
      setSubject('')
      setBody('')
      setPriority('normal')
      setModalOpen(false)
    } catch (err: any) {
      setFormError(err.response?.data?.detail || 'Could not submit ticket.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <LifeBuoy size={24} className="text-primary-600" />
            Support
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Answers to common questions, product docs, and a direct line to our team.
          </p>
        </div>
        <button
          onClick={() => setModalOpen(true)}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} /> Contact Support
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Quick channels */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <a
          href="mailto:support@talkex.dev"
          className="flex items-start gap-3 rounded-xl border border-gray-200 bg-white p-4 hover:border-primary-300 transition-colors"
        >
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
            <Mail size={16} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900">Email us</p>
            <p className="text-xs text-gray-500 mt-0.5">support@talkex.dev · 24h response</p>
          </div>
        </a>
        <button
          onClick={() => setModalOpen(true)}
          className="flex items-start gap-3 rounded-xl border border-gray-200 bg-white p-4 hover:border-primary-300 transition-colors text-left"
        >
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-green-50 text-green-600">
            <MessageSquare size={16} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900">File a ticket</p>
            <p className="text-xs text-gray-500 mt-0.5">Tracked, with an ID you can reference</p>
          </div>
        </button>
        <a
          href="https://github.com/coreaxissoftware/talkex_business"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-start gap-3 rounded-xl border border-gray-200 bg-white p-4 hover:border-primary-300 transition-colors"
        >
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-purple-50 text-purple-600">
            <Book size={16} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900 flex items-center gap-1">
              Docs & source <ExternalLink size={11} />
            </p>
            <p className="text-xs text-gray-500 mt-0.5">Guides, API reference, changelog</p>
          </div>
        </a>
      </div>

      {/* FAQ */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        <div className="border-b border-gray-100 px-5 py-3">
          <h2 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
            <HelpCircle size={16} className="text-primary-600" />
            Frequently asked questions
          </h2>
        </div>
        <div className="divide-y divide-gray-100">
          {FAQ.map((item, i) => (
            <div key={i}>
              <button
                onClick={() => setOpenFaq(openFaq === i ? null : i)}
                className="flex w-full items-center justify-between px-5 py-3 text-left hover:bg-gray-50 transition-colors"
              >
                <span className="text-sm font-medium text-gray-900">{item.q}</span>
                <ChevronDown
                  size={16}
                  className={`shrink-0 text-gray-400 transition-transform ${
                    openFaq === i ? 'rotate-180' : ''
                  }`}
                />
              </button>
              {openFaq === i && (
                <div className="px-5 pb-4 text-sm text-gray-600 leading-relaxed">{item.a}</div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Docs grid */}
      <div>
        <h2 className="text-sm font-semibold text-gray-900 mb-3">Documentation</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {DOCS.map((d) => (
            <a
              key={d.title}
              href={d.href}
              className="block rounded-xl border border-gray-200 bg-white p-4 hover:border-primary-300 hover:shadow-sm transition-all"
            >
              <p className="text-sm font-semibold text-gray-900">{d.title}</p>
              <p className="text-xs text-gray-500 mt-1">{d.description}</p>
            </a>
          ))}
        </div>
      </div>

      {/* Tickets */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        <div className="border-b border-gray-100 px-5 py-3">
          <h2 className="text-sm font-semibold text-gray-900">Your tickets</h2>
        </div>
        {loading ? (
          <div className="p-8 text-center text-sm text-gray-400">Loading…</div>
        ) : tickets.length === 0 ? (
          <div className="p-8 text-center">
            <MessageSquare size={28} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">No tickets yet.</p>
            <p className="text-xs text-gray-400 mt-1">
              Anything you file with "Contact Support" appears here.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Subject</th>
                  <th className="px-4 py-2.5">Priority</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5">Opened</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {tickets.map((t) => (
                  <tr key={t.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2.5 font-medium text-gray-900">{t.subject}</td>
                    <td className="px-4 py-2.5">
                      <span
                        className={`text-xs font-medium capitalize ${
                          PRIORITIES.find((p) => p.id === t.priority)?.color || 'text-gray-500'
                        }`}
                      >
                        {t.priority}
                      </span>
                    </td>
                    <td className="px-4 py-2.5">
                      <StatusPill status={t.status} />
                    </td>
                    <td className="px-4 py-2.5 text-gray-500 text-xs">
                      {new Date(t.created_at).toLocaleDateString('en-IN')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Ticket modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Contact Support">
        <form onSubmit={handleSubmit} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {formError}
            </div>
          )}
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Subject <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              required
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. Webhook not firing for status updates"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Priority</label>
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value as TicketPriority)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none bg-white cursor-pointer"
            >
              {PRIORITIES.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Describe the issue <span className="text-red-500">*</span>
            </label>
            <textarea
              required
              rows={5}
              value={body}
              onChange={(e) => setBody(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="Steps to reproduce, expected vs actual, and any request IDs from Logs…"
            />
          </div>
          <div className="flex gap-2 pt-1">
            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
            >
              {saving ? 'Submitting…' : 'Submit Ticket'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
