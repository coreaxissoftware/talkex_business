import { useEffect, useState } from 'react'
import { Users, RefreshCw, Star, MessageSquare, Inbox } from 'lucide-react'
import { teamActivityService, type AgentActivity } from '../services/team-activity'

export default function TeamActivity() {
  const [rows, setRows] = useState<AgentActivity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = async () => {
    setLoading(true)
    try { setRows(await teamActivityService.list()); setError('') }
    catch { setError('Could not load team activity.') }
    finally { setLoading(false) }
  }
  useEffect(() => { load() }, [])

  const totalOpen = rows.reduce((s, r) => s + r.open_assigned, 0)
  const totalSent = rows.reduce((s, r) => s + r.messages_sent_30d, 0)
  const avgCSAT = rows.filter(r => r.csat_count_30d > 0).reduce((s, r) => s + r.avg_csat_30d, 0) /
    Math.max(1, rows.filter(r => r.csat_count_30d > 0).length)

  return (
    <div className="p-4 sm:p-6 max-w-5xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <Users size={20} className="text-primary-600" />
            Team Activity
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Last 30 days — open assignments, messages sent, and CSAT per agent.
          </p>
        </div>
        <button onClick={load} className="flex items-center gap-2 rounded-lg border border-gray-300 dark:border-gray-600 px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700">
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} /> Refresh
        </button>
      </div>

      {error && <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">{error}</div>}

      {/* KPI tiles */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <KPI icon={<Inbox size={16} className="text-blue-600" />} label="Open assigned" value={totalOpen.toString()} />
        <KPI icon={<MessageSquare size={16} className="text-green-600" />} label="Messages sent (30d)" value={totalSent.toString()} />
        <KPI icon={<Star size={16} className="text-amber-500" />} label="Avg CSAT (30d)" value={avgCSAT.toFixed(2)} />
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/40 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              <th className="px-4 py-2.5">Agent</th>
              <th className="px-4 py-2.5">Role</th>
              <th className="px-4 py-2.5 text-right">Open</th>
              <th className="px-4 py-2.5 text-right">Sent (30d)</th>
              <th className="px-4 py-2.5 text-right">Avg CSAT</th>
              <th className="px-4 py-2.5 text-right">CSAT count</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
            {loading ? (
              <tr><td colSpan={6} className="text-center py-8 text-gray-400 text-xs">Loading…</td></tr>
            ) : rows.length === 0 ? (
              <tr><td colSpan={6} className="text-center py-8 text-gray-400 text-xs">No agents yet.</td></tr>
            ) : rows.map(r => (
              <tr key={r.user_id || r.email} className="hover:bg-gray-50 dark:hover:bg-gray-700/40">
                <td className="px-4 py-3">
                  <div className="font-medium text-gray-900 dark:text-gray-100">{r.name || r.email || '—'}</div>
                  {r.email && <div className="text-[10px] text-gray-400">{r.email}</div>}
                </td>
                <td className="px-4 py-3">
                  <span className="inline-block text-[10px] font-medium rounded-full px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 capitalize">
                    {r.role}
                  </span>
                </td>
                <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{r.open_assigned}</td>
                <td className="px-4 py-3 text-right text-gray-700 dark:text-gray-300">{r.messages_sent_30d}</td>
                <td className="px-4 py-3 text-right font-semibold text-gray-900 dark:text-gray-100">
                  {r.csat_count_30d > 0 ? r.avg_csat_30d.toFixed(2) : '—'}
                </td>
                <td className="px-4 py-3 text-right text-gray-500">{r.csat_count_30d}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function KPI({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4">
      <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 mb-1">{icon}<span>{label}</span></div>
      <div className="text-2xl font-bold text-gray-900 dark:text-gray-100">{value}</div>
    </div>
  )
}
