import { useEffect, useState, useCallback, useMemo } from 'react'
import {
  BarChart3,
  Send,
  Inbox,
  CheckCircle2,
  MessageSquare,
  Users,
  Megaphone,
  Radio,
  TrendingUp,
  Download,
} from 'lucide-react'
import { analyticsService } from '../services/analytics'
import { qualityService, type QualityStats, type QualityEvent } from '../services/quality'
import type { AnalyticsSummary, TimeseriesPoint } from '../types/analytics'

const STATUS_COLORS: Record<string, string> = {
  queued: '#94a3b8',
  sent: '#f59e0b',
  delivered: '#10b981',
  read: '#3b82f6',
  failed: '#ef4444',
}

interface KpiProps {
  label: string
  value: string | number
  icon: React.ComponentType<{ size?: number; className?: string }>
  accent: string
  hint?: string
}

function Kpi({ label, value, icon: Icon, accent, hint }: KpiProps) {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-4">
      <div className="flex items-center justify-between">
        <p className="text-xs font-medium text-gray-500">{label}</p>
        <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${accent}`}>
          <Icon size={16} />
        </div>
      </div>
      <p className="mt-2 text-2xl font-bold text-gray-900">{value}</p>
      {hint && <p className="mt-0.5 text-xs text-gray-400">{hint}</p>}
    </div>
  )
}

interface LineChartProps {
  series: TimeseriesPoint[]
}

function LineChart({ series }: LineChartProps) {
  const dims = useMemo(() => {
    const w = 640
    const h = 200
    const padL = 32
    const padR = 12
    const padT = 12
    const padB = 22
    const chartW = w - padL - padR
    const chartH = h - padT - padB
    const maxY = Math.max(1, ...series.flatMap((p) => [p.outbound, p.inbound]))
    // Round up the y-axis top so ticks are pleasant.
    const niceMax = Math.pow(10, Math.floor(Math.log10(maxY))) *
      Math.ceil(maxY / Math.pow(10, Math.floor(Math.log10(maxY))))

    const xAt = (i: number) =>
      padL + (series.length <= 1 ? chartW / 2 : (i / (series.length - 1)) * chartW)
    const yAt = (v: number) => padT + chartH - (v / niceMax) * chartH

    const path = (key: 'outbound' | 'inbound') =>
      series
        .map((p, i) => `${i === 0 ? 'M' : 'L'} ${xAt(i).toFixed(1)} ${yAt(p[key]).toFixed(1)}`)
        .join(' ')

    return { w, h, padL, padR, padT, padB, chartW, chartH, niceMax, xAt, yAt, path }
  }, [series])

  if (series.length === 0) {
    return <div className="h-52 flex items-center justify-center text-sm text-gray-400">No data</div>
  }

  const xTickEvery = Math.max(1, Math.floor(series.length / 6))

  return (
    <svg viewBox={`0 0 ${dims.w} ${dims.h}`} className="w-full h-52">
      {/* Y-axis grid + labels */}
      {[0, 0.25, 0.5, 0.75, 1].map((t) => {
        const y = dims.padT + dims.chartH * (1 - t)
        const label = Math.round(dims.niceMax * t)
        return (
          <g key={t}>
            <line
              x1={dims.padL}
              x2={dims.w - dims.padR}
              y1={y}
              y2={y}
              stroke="#f1f5f9"
              strokeWidth={1}
            />
            <text
              x={dims.padL - 6}
              y={y + 3}
              textAnchor="end"
              className="fill-gray-400"
              fontSize="10"
            >
              {label}
            </text>
          </g>
        )
      })}

      {/* Outbound line */}
      <path d={dims.path('outbound')} fill="none" stroke="#2563eb" strokeWidth={2} />
      {/* Inbound line */}
      <path d={dims.path('inbound')} fill="none" stroke="#10b981" strokeWidth={2} />

      {/* Data points */}
      {series.map((p, i) => (
        <g key={p.date}>
          <circle cx={dims.xAt(i)} cy={dims.yAt(p.outbound)} r={2.5} fill="#2563eb" />
          <circle cx={dims.xAt(i)} cy={dims.yAt(p.inbound)} r={2.5} fill="#10b981" />
        </g>
      ))}

      {/* X-axis labels — every Nth so they don't collide */}
      {series.map((p, i) => {
        if (i % xTickEvery !== 0 && i !== series.length - 1) return null
        const short = p.date.slice(5)
        return (
          <text
            key={p.date}
            x={dims.xAt(i)}
            y={dims.h - 6}
            textAnchor="middle"
            className="fill-gray-400"
            fontSize="10"
          >
            {short}
          </text>
        )
      })}
    </svg>
  )
}

interface StatusDonutProps {
  byStatus: Record<string, number>
}

function StatusDonut({ byStatus }: StatusDonutProps) {
  const total = Object.values(byStatus).reduce((a, b) => a + b, 0)
  if (total === 0) {
    return <div className="h-40 flex items-center justify-center text-sm text-gray-400">No outbound messages yet</div>
  }

  const size = 160
  const r = 60
  const stroke = 20
  const cx = size / 2
  const cy = size / 2
  const circumference = 2 * Math.PI * r

  const entries = Object.entries(byStatus)
  let offset = 0
  return (
    <div className="flex items-center gap-6">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        <circle cx={cx} cy={cy} r={r} fill="none" stroke="#f1f5f9" strokeWidth={stroke} />
        {entries.map(([status, count]) => {
          const frac = count / total
          const dash = frac * circumference
          const seg = (
            <circle
              key={status}
              cx={cx}
              cy={cy}
              r={r}
              fill="none"
              stroke={STATUS_COLORS[status] || '#94a3b8'}
              strokeWidth={stroke}
              strokeDasharray={`${dash} ${circumference - dash}`}
              strokeDashoffset={-offset}
              transform={`rotate(-90 ${cx} ${cy})`}
            />
          )
          offset += dash
          return seg
        })}
        <text x={cx} y={cy - 4} textAnchor="middle" className="fill-gray-900" fontSize="20" fontWeight="700">
          {total}
        </text>
        <text x={cx} y={cy + 14} textAnchor="middle" className="fill-gray-400" fontSize="10">
          outbound
        </text>
      </svg>

      <div className="space-y-1.5">
        {entries.map(([status, count]) => (
          <div key={status} className="flex items-center gap-2 text-xs">
            <span
              className="inline-block h-2.5 w-2.5 rounded-sm"
              style={{ backgroundColor: STATUS_COLORS[status] || '#94a3b8' }}
            />
            <span className="text-gray-500 capitalize w-20">{status}</span>
            <span className="font-mono font-semibold text-gray-900">{count}</span>
            <span className="text-gray-400">
              ({((count / total) * 100).toFixed(0)}%)
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function Analytics() {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [series, setSeries] = useState<TimeseriesPoint[]>([])
  const [range, setRange] = useState(30)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [qualityStats, setQualityStats] = useState<QualityStats | null>(null)
  const [qualityEvents, setQualityEvents] = useState<QualityEvent[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [s, ts] = await Promise.all([
        analyticsService.summary(),
        analyticsService.timeseries(range),
      ])
      setSummary(s)
      setSeries(ts)
      const [qs, qe] = await Promise.allSettled([qualityService.stats(), qualityService.events()])
      if (qs.status === 'fulfilled') setQualityStats(qs.value)
      if (qe.status === 'fulfilled') setQualityEvents(qe.value)
      setError('')
    } catch {
      setError('Could not load analytics.')
    } finally {
      setLoading(false)
    }
  }, [range])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <BarChart3 size={24} className="text-primary-600" />
            Analytics
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Message volume, delivery rate, and channel mix across your account.
          </p>
        </div>
        <div className="flex items-center gap-2">
          {[7, 30, 90].map((d) => (
            <button
              key={d}
              onClick={() => setRange(d)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium transition-colors ${
                range === d
                  ? 'bg-primary-600 text-white'
                  : 'border border-gray-300 text-gray-700 hover:bg-gray-50'
              }`}
            >
              {d}d
            </button>
          ))}
          <button
            onClick={() => analyticsService.exportCSV(range)}
            className="rounded-lg border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50 flex items-center gap-1"
          >
            <Download size={12} /> Export CSV
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading && !summary ? (
        <div className="p-10 text-center text-sm text-gray-400">Loading analytics…</div>
      ) : summary ? (
        <>
          {/* KPI grid */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <Kpi
              label="Messages sent"
              value={summary.outbound_messages.toLocaleString()}
              icon={Send}
              accent="bg-blue-50 text-blue-600"
              hint="outbound total"
            />
            <Kpi
              label="Messages received"
              value={summary.inbound_messages.toLocaleString()}
              icon={Inbox}
              accent="bg-green-50 text-green-600"
              hint="inbound total"
            />
            <Kpi
              label="Delivery rate"
              value={`${summary.delivery_rate.toFixed(1)}%`}
              icon={CheckCircle2}
              accent="bg-emerald-50 text-emerald-600"
              hint="delivered / outbound"
            />
            <Kpi
              label="Open windows"
              value={summary.open_conversations.toLocaleString()}
              icon={MessageSquare}
              accent="bg-amber-50 text-amber-600"
              hint="last 24h"
            />
            <Kpi
              label="Contacts"
              value={summary.total_contacts.toLocaleString()}
              icon={Users}
              accent="bg-purple-50 text-purple-600"
            />
            <Kpi
              label="Active campaigns"
              value={summary.active_campaigns.toLocaleString()}
              icon={Megaphone}
              accent="bg-pink-50 text-pink-600"
              hint="scheduled + running"
            />
            <Kpi
              label="Total messages"
              value={summary.total_messages.toLocaleString()}
              icon={TrendingUp}
              accent="bg-indigo-50 text-indigo-600"
              hint="all directions"
            />
            <Kpi
              label="Active channels"
              value={summary.by_channel.length.toLocaleString()}
              icon={Radio}
              accent="bg-cyan-50 text-cyan-600"
            />
          </div>

          {/* Line chart */}
          <div className="rounded-xl border border-gray-200 bg-white p-5">
            <div className="flex items-center justify-between mb-3">
              <div>
                <h2 className="text-sm font-semibold text-gray-900">Daily Message Volume</h2>
                <p className="text-xs text-gray-500 mt-0.5">Last {range} days</p>
              </div>
              <div className="flex items-center gap-4 text-xs">
                <span className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-blue-600" /> Outbound
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-green-600" /> Inbound
                </span>
              </div>
            </div>
            <LineChart series={series} />
          </div>

          {/* Two-column: donut + channel breakdown */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <div className="rounded-xl border border-gray-200 bg-white p-5">
              <h2 className="text-sm font-semibold text-gray-900 mb-4">Outbound by Status</h2>
              <StatusDonut byStatus={summary.by_status} />
            </div>

            <div className="rounded-xl border border-gray-200 bg-white p-5">
              <h2 className="text-sm font-semibold text-gray-900 mb-4">Conversations by Channel</h2>
              {summary.by_channel.length === 0 ? (
                <p className="text-sm text-gray-400 text-center py-6">No conversations yet</p>
              ) : (
                <div className="space-y-3">
                  {summary.by_channel.map((c) => {
                    const total = summary.by_channel.reduce((a, x) => a + x.count, 0)
                    const pct = (c.count / total) * 100
                    return (
                      <div key={c.channel}>
                        <div className="flex items-center justify-between text-xs mb-1">
                          <span className="font-medium text-gray-700 capitalize">{c.channel}</span>
                          <span className="text-gray-500">
                            {c.count} <span className="text-gray-400">({pct.toFixed(0)}%)</span>
                          </span>
                        </div>
                        <div className="h-2 rounded-full bg-gray-100 overflow-hidden">
                          <div
                            className="h-full bg-primary-500 rounded-full transition-all"
                            style={{ width: `${pct}%` }}
                          />
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Quality Rating */}
          {qualityStats && (
            <div className="rounded-xl border border-gray-200 bg-white p-5">
              <h2 className="text-sm font-semibold text-gray-900 mb-4">Messaging Quality Rating</h2>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
                <div className="rounded-lg border p-3">
                  <p className="text-xs text-gray-500 mb-1">Status</p>
                  <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-sm font-bold capitalize ${
                    qualityStats.status === 'green' ? 'bg-green-50 text-green-700' :
                    qualityStats.status === 'yellow' ? 'bg-yellow-50 text-yellow-700' :
                    'bg-red-50 text-red-700'
                  }`}>
                    <span className={`h-2.5 w-2.5 rounded-full ${
                      qualityStats.status === 'green' ? 'bg-green-500' :
                      qualityStats.status === 'yellow' ? 'bg-yellow-500' :
                      'bg-red-500'
                    }`} />
                    {qualityStats.status}
                  </span>
                </div>
                <div className="rounded-lg border p-3">
                  <p className="text-xs text-gray-500 mb-1">Blocks (7d)</p>
                  <p className="text-xl font-bold text-gray-900">{qualityStats.blocks_last_7d}</p>
                  <p className="text-[10px] text-gray-400">of {qualityStats.threshold} max before flag</p>
                </div>
                <div className="rounded-lg border p-3">
                  <p className="text-xs text-gray-500 mb-1">Reports (7d)</p>
                  <p className="text-xl font-bold text-gray-900">{qualityStats.reports_last_7d}</p>
                </div>
                <div className="rounded-lg border p-3">
                  <p className="text-xs text-gray-500 mb-1">Total Blocks</p>
                  <p className="text-xl font-bold text-gray-900">{qualityStats.total_blocks}</p>
                  <p className="text-[10px] text-gray-400">{qualityStats.total_reports} reports all-time</p>
                </div>
              </div>

              {qualityStats.status !== 'green' && qualityStats.flagged_at && (
                <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-xs text-red-700 mb-4">
                  Quality flagged on {new Date(qualityStats.flagged_at).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })}.
                  Reduce blocks/reports to recover. Messaging may be restricted.
                </div>
              )}

              {qualityEvents.length > 0 && (
                <div>
                  <h3 className="text-xs font-semibold text-gray-700 mb-2">Recent Events</h3>
                  <div className="divide-y divide-gray-50 max-h-48 overflow-y-auto">
                    {qualityEvents.slice(0, 10).map(ev => (
                      <div key={ev.id} className="flex items-center gap-3 py-2 text-xs">
                        <span className={`rounded-full px-2 py-0.5 font-semibold ${
                          ev.type === 'block' ? 'bg-red-50 text-red-700' :
                          ev.type === 'report' ? 'bg-orange-50 text-orange-700' :
                          'bg-green-50 text-green-700'
                        }`}>{ev.type}</span>
                        <span className="text-gray-600 capitalize">{ev.channel}</span>
                        {ev.reason && <span className="text-gray-400 truncate flex-1">{ev.reason}</span>}
                        <span className="text-gray-400 whitespace-nowrap">
                          {new Date(ev.created_at).toLocaleDateString('en-IN', { day: '2-digit', month: 'short' })}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </>
      ) : null}
    </div>
  )
}
