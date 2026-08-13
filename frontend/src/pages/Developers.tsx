import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  Code2,
  Plus,
  Trash2,
  KeyRound,
  Copy,
  CheckCircle2,
  AlertTriangle,
  Ban,
  X,
  Play,
  Terminal,
} from 'lucide-react'
import { apiKeysService } from '../services/apiKeys'
import api from '../services/api'
import type { ApiKey } from '../types/apiKey'
import Modal from '../components/Modal'

function formatTime(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('en-IN', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function StatusPill({ k }: { k: ApiKey }) {
  if (k.revoked_at) {
    return (
      <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">
        <Ban size={12} /> Revoked
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700">
      <CheckCircle2 size={12} /> Active
    </span>
  )
}

export default function Developers() {
  const [keys, setKeys] = useState<ApiKey[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [createOpen, setCreateOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  const [revealPlaintext, setRevealPlaintext] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const [confirmRevokeId, setConfirmRevokeId] = useState<string | null>(null)
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  // Playground
  const ENDPOINTS = [
    { label: 'List Contacts', method: 'GET', path: '/contacts' },
    { label: 'List Templates', method: 'GET', path: '/templates' },
    { label: 'List Campaigns', method: 'GET', path: '/campaigns' },
    { label: 'List Conversations', method: 'GET', path: '/conversations' },
    { label: 'Analytics Summary', method: 'GET', path: '/analytics/summary' },
    { label: 'Analytics Timeseries', method: 'GET', path: '/analytics/timeseries?days=7' },
    { label: 'Wallet Info', method: 'GET', path: '/wallet' },
    { label: 'Audit Logs', method: 'GET', path: '/audit-logs?limit=5' },
  ]
  const [pgEndpoint, setPgEndpoint] = useState(0)
  const [pgResponse, setPgResponse] = useState<string>('')
  const [pgStatus, setPgStatus] = useState<number | null>(null)
  const [pgRunning, setPgRunning] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await apiKeysService.list()
      setKeys(data)
      setError('')
    } catch {
      setError('Could not load API keys.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    setNewName('')
    setError('')
    setRevealPlaintext(null)
    setCopied(false)
    setCreateOpen(true)
  }

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault()
    setCreating(true)
    setError('')
    try {
      const result = await apiKeysService.create(newName)
      setKeys((prev) => [result.api_key, ...prev])
      setRevealPlaintext(result.plaintext)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not create key.')
    } finally {
      setCreating(false)
    }
  }

  const handleRevoke = async (id: string) => {
    try {
      const updated = await apiKeysService.revoke(id)
      setKeys((prev) => prev.map((k) => (k.id === updated.id ? updated : k)))
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not revoke key.')
    } finally {
      setConfirmRevokeId(null)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await apiKeysService.remove(id)
      setKeys((prev) => prev.filter((k) => k.id !== id))
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not delete key.')
    } finally {
      setConfirmDeleteId(null)
    }
  }

  const copyPlaintext = async () => {
    if (!revealPlaintext) return
    try {
      await navigator.clipboard.writeText(revealPlaintext)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* clipboard may be denied in some environments — non-fatal */
    }
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Code2 size={24} className="text-primary-600" />
            Developers
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            API keys let your servers send messages programmatically. Keep them secret.
          </p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} />
          Create API Key
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Keys table */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        {loading ? (
          <div className="p-10 text-center text-sm text-gray-400">Loading…</div>
        ) : keys.length === 0 ? (
          <div className="p-10 text-center">
            <KeyRound size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">No API keys yet.</p>
            <p className="text-xs text-gray-400 mt-1">
              Create one to start sending messages from your own backend.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Name</th>
                  <th className="px-4 py-2.5">Key Prefix</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5">Last used</th>
                  <th className="px-4 py-2.5">Created</th>
                  <th className="px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {keys.map((k) => (
                  <tr key={k.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">{k.name}</td>
                    <td className="px-4 py-3">
                      <code className="rounded bg-gray-100 px-2 py-0.5 text-xs font-mono text-gray-700">
                        {k.prefix}…
                      </code>
                    </td>
                    <td className="px-4 py-3">
                      <StatusPill k={k} />
                    </td>
                    <td className="px-4 py-3 text-gray-500 text-xs">
                      {formatTime(k.last_used_at)}
                    </td>
                    <td className="px-4 py-3 text-gray-500 text-xs">
                      {formatTime(k.created_at)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        {confirmRevokeId === k.id ? (
                          <>
                            <button
                              onClick={() => handleRevoke(k.id)}
                              className="rounded-lg px-2 py-1 text-xs font-medium text-white bg-amber-600 hover:bg-amber-700"
                            >
                              Confirm revoke
                            </button>
                            <button
                              onClick={() => setConfirmRevokeId(null)}
                              className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100"
                            >
                              <X size={14} />
                            </button>
                          </>
                        ) : confirmDeleteId === k.id ? (
                          <>
                            <button
                              onClick={() => handleDelete(k.id)}
                              className="rounded-lg px-2 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700"
                            >
                              Confirm delete
                            </button>
                            <button
                              onClick={() => setConfirmDeleteId(null)}
                              className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100"
                            >
                              <X size={14} />
                            </button>
                          </>
                        ) : (
                          <>
                            {!k.revoked_at && (
                              <button
                                onClick={() => setConfirmRevokeId(k.id)}
                                className="rounded-lg p-1.5 text-amber-600 hover:bg-amber-50 transition-colors"
                                title="Revoke"
                              >
                                <Ban size={14} />
                              </button>
                            )}
                            {k.revoked_at && (
                              <button
                                onClick={() => setConfirmDeleteId(k.id)}
                                className="rounded-lg p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600 transition-colors"
                                title="Delete"
                              >
                                <Trash2 size={14} />
                              </button>
                            )}
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* API Playground */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        <div className="border-b border-gray-100 px-5 py-3 flex items-center gap-2">
          <Terminal size={16} className="text-primary-600" />
          <h2 className="text-sm font-semibold text-gray-900">API Playground</h2>
          <span className="text-xs text-gray-400 ml-1">Test endpoints with your active session</span>
        </div>
        <div className="p-5 space-y-4">
          <div className="flex items-end gap-3">
            <div className="flex-1">
              <label className="block text-xs font-medium text-gray-700 mb-1.5">Endpoint</label>
              <select
                value={pgEndpoint}
                onChange={e => setPgEndpoint(Number(e.target.value))}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none bg-white cursor-pointer"
              >
                {ENDPOINTS.map((ep, i) => (
                  <option key={i} value={i}>{ep.method} {ep.path} — {ep.label}</option>
                ))}
              </select>
            </div>
            <button
              disabled={pgRunning}
              onClick={async () => {
                const ep = ENDPOINTS[pgEndpoint]
                setPgRunning(true)
                setPgResponse('')
                setPgStatus(null)
                try {
                  const res = await api.request({ method: ep.method.toLowerCase(), url: ep.path })
                  setPgStatus(res.status)
                  setPgResponse(JSON.stringify(res.data, null, 2))
                } catch (err: any) {
                  setPgStatus(err.response?.status || 0)
                  setPgResponse(JSON.stringify(err.response?.data || { error: err.message }, null, 2))
                } finally {
                  setPgRunning(false)
                }
              }}
              className="flex items-center gap-1.5 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
            >
              <Play size={14} />
              {pgRunning ? 'Running…' : 'Run'}
            </button>
          </div>

          {pgResponse && (
            <div>
              <div className="flex items-center gap-2 mb-1.5">
                <span className="text-xs font-medium text-gray-700">Response</span>
                {pgStatus && (
                  <span className={`text-xs font-mono px-1.5 py-0.5 rounded ${pgStatus < 400 ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                    {pgStatus}
                  </span>
                )}
              </div>
              <pre className="rounded-lg bg-gray-900 text-gray-100 text-xs p-4 overflow-x-auto max-h-80 overflow-y-auto font-mono">
                {pgResponse}
              </pre>
            </div>
          )}
        </div>
      </div>

      {/* Create modal */}
      <Modal
        open={createOpen}
        onClose={() => {
          setCreateOpen(false)
          setRevealPlaintext(null)
        }}
        title={revealPlaintext ? 'Your new API key' : 'Create API Key'}
      >
        {revealPlaintext ? (
          <div className="space-y-4">
            <div className="rounded-lg bg-amber-50 border border-amber-200 px-3 py-3 text-xs text-amber-800 flex gap-2">
              <AlertTriangle size={14} className="shrink-0 mt-0.5" />
              <div>
                <p className="font-semibold">Copy this key now.</p>
                <p className="mt-0.5">
                  This is the only time it will be shown. Once you close this dialog, we
                  cannot retrieve it — you'll have to create a new key.
                </p>
              </div>
            </div>

            <div className="relative">
              <code className="block w-full rounded-lg border border-gray-300 bg-gray-50 px-3 py-3 pr-12 text-xs font-mono text-gray-900 break-all">
                {revealPlaintext}
              </code>
              <button
                onClick={copyPlaintext}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded-lg p-2 text-gray-400 hover:bg-white hover:text-gray-700 transition-colors"
                title="Copy"
              >
                {copied ? <CheckCircle2 size={16} className="text-green-600" /> : <Copy size={16} />}
              </button>
            </div>

            <button
              type="button"
              onClick={() => {
                setCreateOpen(false)
                setRevealPlaintext(null)
              }}
              className="w-full rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
            >
              I've saved my key
            </button>
          </div>
        ) : (
          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1.5">
                Key Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                required
                autoFocus
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="e.g. Production server, Staging, CI"
              />
              <p className="mt-1.5 text-xs text-gray-500">
                A label so you can tell keys apart in this list.
              </p>
            </div>

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
                disabled={creating || !newName.trim()}
                className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
              >
                {creating ? 'Creating…' : 'Create Key'}
              </button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  )
}
