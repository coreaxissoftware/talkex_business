import { Fragment, useEffect, useState, useCallback } from 'react'
import {
  ScrollText,
  AlertCircle,
  CheckCircle2,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Search,
} from 'lucide-react'
import { auditService } from '../services/audit'
import type { AuditLogEntry, AuditStats } from '../types/audit'

const METHOD_COLORS: Record<string, string> = {
  GET: 'bg-blue-50 text-blue-700 border-blue-200',
  POST: 'bg-green-50 text-green-700 border-green-200',
  PATCH: 'bg-amber-50 text-amber-700 border-amber-200',
  DELETE: 'bg-red-50 text-red-700 border-red-200',
  PUT: 'bg-purple-50 text-purple-700 border-purple-200',
}

function StatusBadge({ code, success }: { code: number; success: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
        success ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
      }`}
    >
      {success ? <CheckCircle2 size={12} /> : <AlertCircle size={12} />}
      {code}
    </span>
  )
}

function formatTime(iso: string) {
  const d = new Date(iso)
  return d.toLocaleString('en-IN', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

export default function Logs() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([])
  const [total, setTotal] = useState(0)
  const [stats, setStats] = useState<AuditStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [onlyFailed, setOnlyFailed] = useState(false)
  const [method, setMethod] = useState('')
  const [search, setSearch] = useState('')
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [list, s] = await Promise.all([
        auditService.list({
          failed: onlyFailed || undefined,
          method: method || undefined,
          search: search || undefined,
          limit: 50,
        }),
        auditService.stats(),
      ])
      setEntries(list.items)
      setTotal(list.total)
      setStats(s)
    } catch {
      // Audit log fetch failing shouldn't crash the page — leave prior state.
    } finally {
      setLoading(false)
    }
  }, [onlyFailed, method, search])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <ScrollText size={24} className="text-primary-600" />
            Activity Logs
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Every API request made from your account — including failures — for debugging and audit.
          </p>
        </div>
        <button
          onClick={load}
          className="flex items-center gap-2 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
        >
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          Refresh
        </button>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <p className="text-xs font-medium text-gray-500">Total Requests</p>
          <p className="text-2xl font-bold text-gray-900 mt-1">{stats?.total ?? '—'}</p>
        </div>
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <p className="text-xs font-medium text-gray-500">Failed Requests</p>
          <p className="text-2xl font-bold text-red-600 mt-1">{stats?.failed ?? '—'}</p>
        </div>
        <div className="rounded-xl border border-gray-200 bg-white p-4">
          <p className="text-xs font-medium text-gray-500">Success Rate</p>
          <p className="text-2xl font-bold text-green-600 mt-1">
            {stats ? `${stats.success_rate.toFixed(1)}%` : '—'}
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3 rounded-xl border border-gray-200 bg-white p-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by path (e.g. /contacts)"
            className="w-full rounded-lg border border-gray-300 pl-9 pr-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
          />
        </div>

        <select
          value={method}
          onChange={(e) => setMethod(e.target.value)}
          className="rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none cursor-pointer"
        >
          <option value="">All methods</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PATCH">PATCH</option>
          <option value="DELETE">DELETE</option>
        </select>

        <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={onlyFailed}
            onChange={(e) => setOnlyFailed(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
          />
          Failed only
        </label>
      </div>

      {/* Log table */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        {loading && entries.length === 0 ? (
          <div className="p-10 text-center text-sm text-gray-400">Loading logs…</div>
        ) : entries.length === 0 ? (
          <div className="p-10 text-center">
            <ScrollText size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">No requests logged yet.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Time</th>
                  <th className="px-4 py-2.5">Method</th>
                  <th className="px-4 py-2.5">Path</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5">Latency</th>
                  <th className="px-4 py-2.5">IP</th>
                  <th className="px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {entries.map((e) => (
                  <Fragment key={e.id}>
                    <tr
                      className={`hover:bg-gray-50 transition-colors ${
                        !e.success ? 'bg-red-50/40' : ''
                      } ${e.error_body ? 'cursor-pointer' : ''}`}
                      onClick={() =>
                        e.error_body && setExpandedId(expandedId === e.id ? null : e.id)
                      }
                    >
                      <td className="px-4 py-2.5 text-gray-500 whitespace-nowrap">
                        {formatTime(e.created_at)}
                      </td>
                      <td className="px-4 py-2.5">
                        <span
                          className={`inline-block rounded border px-1.5 py-0.5 text-xs font-semibold ${
                            METHOD_COLORS[e.method] || 'bg-gray-50 text-gray-700 border-gray-200'
                          }`}
                        >
                          {e.method}
                        </span>
                      </td>
                      <td className="px-4 py-2.5 font-mono text-xs text-gray-700">{e.path}</td>
                      <td className="px-4 py-2.5">
                        <StatusBadge code={e.status_code} success={e.success} />
                      </td>
                      <td className="px-4 py-2.5 text-gray-500">{e.latency_ms}ms</td>
                      <td className="px-4 py-2.5 text-gray-400 font-mono text-xs">{e.client_ip}</td>
                      <td className="px-4 py-2.5 text-gray-400">
                        {e.error_body &&
                          (expandedId === e.id ? (
                            <ChevronUp size={14} />
                          ) : (
                            <ChevronDown size={14} />
                          ))}
                      </td>
                    </tr>
                    {expandedId === e.id && e.error_body && (
                      <tr className="bg-red-50/60">
                        <td colSpan={7} className="px-4 py-3">
                          <pre className="text-xs font-mono text-red-800 whitespace-pre-wrap break-all">
                            {e.error_body}
                          </pre>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {entries.length > 0 && (
        <p className="text-xs text-gray-400 text-center">
          Showing {entries.length} of {total} requests
        </p>
      )}
    </div>
  )
}
