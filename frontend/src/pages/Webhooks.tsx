import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  Webhook,
  Plus,
  Trash2,
  Copy,
  CheckCircle2,
  AlertTriangle,
  X,
  ExternalLink,
  ChevronRight,
} from 'lucide-react'
import { webhooksService } from '../services/webhooks'
import { WEBHOOK_EVENTS } from '../types/webhook'
import type { WebhookEndpoint, WebhookDelivery } from '../types/webhook'
import Modal from '../components/Modal'

function formatTime(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('en-IN', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function Webhooks() {
  const [endpoints, setEndpoints] = useState<WebhookEndpoint[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [events, setEvents] = useState<string[]>([...WEBHOOK_EVENTS])
  const [active, setActive] = useState(true)
  const [creating, setCreating] = useState(false)
  const [reveal, setReveal] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  const [detail, setDetail] = useState<WebhookEndpoint | null>(null)
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([])
  const [detailLoading, setDetailLoading] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await webhooksService.list()
      setEndpoints(data)
      setError('')
    } catch {
      setError('Could not load webhooks.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    setName('')
    setUrl('')
    setEvents([...WEBHOOK_EVENTS])
    setActive(true)
    setReveal(null)
    setCopied(false)
    setCreateOpen(true)
  }

  const toggleEvent = (ev: string) => {
    setEvents((prev) => (prev.includes(ev) ? prev.filter((e) => e !== ev) : [...prev, ev]))
  }

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setCreating(true)
    try {
      const result = await webhooksService.create({ name, url, events, active })
      setEndpoints((prev) => [result.endpoint, ...prev])
      setReveal(result.plaintext_secret)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not create webhook.')
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await webhooksService.remove(id)
      setEndpoints((prev) => prev.filter((e) => e.id !== id))
      if (detail?.id === id) setDetail(null)
    } catch {
      setError('Could not delete webhook.')
    } finally {
      setConfirmDeleteId(null)
    }
  }

  const openDetail = async (e: WebhookEndpoint) => {
    setDetail(e)
    setDetailLoading(true)
    try {
      const d = await webhooksService.deliveries(e.id)
      setDeliveries(d)
    } catch {
      setDeliveries([])
    } finally {
      setDetailLoading(false)
    }
  }

  const copy = async (val: string) => {
    try {
      await navigator.clipboard.writeText(val)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* clipboard may be denied */
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Webhook size={24} className="text-primary-600" />
            Webhooks
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Outbound HTTP callbacks fired on platform events, signed with HMAC-SHA256.
          </p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700"
        >
          <Plus size={16} /> New Endpoint
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        {loading ? (
          <div className="p-10 text-center text-sm text-gray-400">Loading…</div>
        ) : endpoints.length === 0 ? (
          <div className="p-10 text-center">
            <Webhook size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">No webhook endpoints yet.</p>
            <p className="text-xs text-gray-400 mt-1">
              Add one to receive real-time events at your own URL.
            </p>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                <th className="px-4 py-2.5">Name</th>
                <th className="px-4 py-2.5">URL</th>
                <th className="px-4 py-2.5">Events</th>
                <th className="px-4 py-2.5">Last fired</th>
                <th className="px-4 py-2.5">Status</th>
                <th className="px-4 py-2.5"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {endpoints.map((e) => (
                <tr key={e.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900">{e.name}</td>
                  <td className="px-4 py-3 font-mono text-xs text-gray-600 truncate max-w-xs">
                    {e.url}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-1">
                      {e.events.map((ev) => (
                        <span key={ev} className="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-mono text-gray-700">
                          {ev}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-xs text-gray-500">{formatTime(e.last_fired_at)}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
                        e.active ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'
                      }`}
                    >
                      {e.active ? 'Active' : 'Paused'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-1">
                      <button
                        onClick={() => openDetail(e)}
                        className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
                        title="View deliveries"
                      >
                        <ChevronRight size={14} />
                      </button>
                      {confirmDeleteId === e.id ? (
                        <>
                          <button
                            onClick={() => handleDelete(e.id)}
                            className="rounded-lg px-2 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700"
                          >
                            Confirm
                          </button>
                          <button
                            onClick={() => setConfirmDeleteId(null)}
                            className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100"
                          >
                            <X size={14} />
                          </button>
                        </>
                      ) : (
                        <button
                          onClick={() => setConfirmDeleteId(e.id)}
                          className="rounded-lg p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600"
                          title="Delete"
                        >
                          <Trash2 size={14} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Detail modal */}
      <Modal open={!!detail} onClose={() => setDetail(null)} title={detail ? `Deliveries — ${detail.name}` : 'Deliveries'}>
        {detailLoading ? (
          <p className="text-sm text-gray-400 text-center py-6">Loading…</p>
        ) : deliveries.length === 0 ? (
          <p className="text-sm text-gray-500 text-center py-6">
            No deliveries yet — this endpoint hasn't fired.
          </p>
        ) : (
          <div className="space-y-2 max-h-80 overflow-y-auto">
            {deliveries.map((d) => (
              <div key={d.id} className="rounded-lg border border-gray-200 p-2.5 text-xs">
                <div className="flex items-center justify-between gap-2 mb-1">
                  <span className="font-mono text-gray-700">{d.event}</span>
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold ${
                      d.success ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
                    }`}
                  >
                    HTTP {d.status_code || 'network'}
                  </span>
                </div>
                <p className="text-[10px] text-gray-400">{formatTime(d.created_at)}</p>
                {d.error_message && (
                  <p className="text-red-600 mt-1 truncate">{d.error_message}</p>
                )}
              </div>
            ))}
          </div>
        )}
      </Modal>

      {/* Create modal */}
      <Modal
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setReveal(null)
        }}
        title={reveal ? 'Endpoint created — save your secret' : 'New Webhook Endpoint'}
      >
        {reveal ? (
          <div className="space-y-4">
            <div className="rounded-lg bg-amber-50 border border-amber-200 px-3 py-3 text-xs text-amber-800 flex gap-2">
              <AlertTriangle size={14} className="shrink-0 mt-0.5" />
              <div>
                <p className="font-semibold">Copy this secret now.</p>
                <p className="mt-0.5">
                  Use it to verify the X-TalkEx-Signature header on incoming deliveries.
                  This is the only time it's shown.
                </p>
              </div>
            </div>
            <div className="relative">
              <code className="block w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-3 pr-12 text-xs font-mono text-gray-900 break-all">
                {reveal}
              </code>
              <button
                onClick={() => copy(reveal)}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-2 text-gray-400 hover:bg-white hover:text-gray-700"
              >
                {copied ? <CheckCircle2 size={16} className="text-green-600" /> : <Copy size={16} />}
              </button>
            </div>
            <button
              onClick={() => {
                setCreateOpen(false)
                setReveal(null)
              }}
              className="w-full rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
            >
              I've saved my secret
            </button>
          </div>
        ) : (
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Name</label>
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="e.g. Order system"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5 flex items-center gap-1">
                Endpoint URL <ExternalLink size={10} />
              </label>
              <input
                type="url"
                required
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="https://your-app.example.com/webhooks/talkex"
              />
              <p className="mt-1 text-[10px] text-gray-500">
                Must be a public HTTPS URL. Return 2xx to acknowledge.
              </p>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-2">Events to subscribe</label>
              <div className="space-y-1.5">
                {WEBHOOK_EVENTS.map((ev) => (
                  <label key={ev} className="flex items-center gap-2 text-xs cursor-pointer">
                    <input
                      type="checkbox"
                      checked={events.includes(ev)}
                      onChange={() => toggleEvent(ev)}
                      className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="font-mono text-gray-700">{ev}</span>
                  </label>
                ))}
              </div>
            </div>
            <label className="flex items-center gap-2 text-xs text-gray-700 cursor-pointer">
              <input
                type="checkbox"
                checked={active}
                onChange={(e) => setActive(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              Active — start delivering events immediately
            </label>
            <div className="flex gap-2 pt-2">
              <button
                type="button"
                onClick={() => setCreateOpen(false)}
                className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={creating || !name || !url || events.length === 0}
                className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
              >
                {creating ? 'Creating…' : 'Create Endpoint'}
              </button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  )
}
