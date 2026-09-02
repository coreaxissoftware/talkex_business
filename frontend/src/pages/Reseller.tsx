import { useEffect, useState } from 'react'
import { TrendingUp, Users, IndianRupee, Percent, ArrowUpRight } from 'lucide-react'
import { resellerService, type ResellerDashboard } from '../services/reseller'

const WINDOWS = [
  { days: 7, label: '7d' },
  { days: 30, label: '30d' },
  { days: 90, label: '90d' },
]

export default function Reseller() {
  const [days, setDays] = useState(30)
  const [d, setD] = useState<ResellerDashboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    resellerService.dashboard(days)
      .then(setD)
      .catch((err) => setError(err.response?.data?.detail || 'Could not load reseller data.'))
      .finally(() => setLoading(false))
  }, [days])

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <TrendingUp size={24} className="text-primary-600" /> Reseller margin
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Cost, revenue, and gross margin across every child tenant.
          </p>
        </div>
        <div className="flex bg-gray-100 rounded-lg p-1">
          {WINDOWS.map((w) => (
            <button
              key={w.days}
              onClick={() => setDays(w.days)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium ${
                days === w.days ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500'
              }`}
            >
              {w.label}
            </button>
          ))}
        </div>
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-amber-50 border border-amber-200 px-4 py-3 text-sm text-amber-800">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-16 text-gray-500">Loading…</div>
      ) : !d ? null : (
        <>
          {/* Top KPI row */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
            <Kpi
              label="Tenants"
              value={d.total_tenants.toLocaleString('en-IN')}
              icon={<Users size={14} />}
            />
            <Kpi
              label="Messages"
              value={d.total_messages.toLocaleString('en-IN')}
              icon={<ArrowUpRight size={14} />}
            />
            <Kpi
              label="Wholesale cost"
              value={`₹${d.total_cost.toLocaleString('en-IN', { maximumFractionDigits: 2 })}`}
              icon={<IndianRupee size={14} />}
              tone="danger"
            />
            <Kpi
              label="Retail revenue"
              value={`₹${d.total_revenue.toLocaleString('en-IN', { maximumFractionDigits: 2 })}`}
              icon={<IndianRupee size={14} />}
              tone="success"
            />
            <Kpi
              label="Margin"
              value={`₹${d.total_margin.toLocaleString('en-IN', { maximumFractionDigits: 2 })}`}
              sub={`${d.avg_margin_pct.toFixed(1)}%`}
              icon={<Percent size={14} />}
              tone="primary"
            />
          </div>

          {/* Tenant table */}
          <div className="border border-gray-200 rounded-xl bg-white overflow-hidden">
            <div className="px-5 py-3 border-b border-gray-200 bg-gray-50">
              <h2 className="font-semibold text-gray-900 text-sm">
                Per tenant · {d.total_tenants} row{d.total_tenants === 1 ? '' : 's'}
              </h2>
            </div>
            {d.tenants.length === 0 ? (
              <div className="p-8 text-center text-gray-500 text-sm">
                No child tenants yet. Sub-orgs you create appear here once they send their first
                message.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-xs uppercase tracking-wider text-gray-500 border-b border-gray-200">
                      <th className="px-4 py-2.5">Tenant</th>
                      <th className="px-4 py-2.5 text-right">Messages</th>
                      <th className="px-4 py-2.5 text-right">Cost</th>
                      <th className="px-4 py-2.5 text-right">Revenue</th>
                      <th className="px-4 py-2.5 text-right">Margin</th>
                      <th className="px-4 py-2.5 text-right">Margin %</th>
                      <th className="px-4 py-2.5">Top channel</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {d.tenants.map((t) => (
                      <tr key={t.tenant_owner_id} className="hover:bg-gray-50">
                        <td className="px-4 py-2.5 font-medium text-gray-900">
                          {t.tenant_name}
                          <div className="text-xs text-gray-400 font-mono">
                            {t.tenant_owner_id.slice(0, 8)}
                          </div>
                        </td>
                        <td className="px-4 py-2.5 text-right tabular-nums">
                          {t.messages.toLocaleString('en-IN')}
                        </td>
                        <td className="px-4 py-2.5 text-right tabular-nums text-gray-600">
                          ₹{t.cost.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
                        </td>
                        <td className="px-4 py-2.5 text-right tabular-nums">
                          ₹{t.revenue.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
                        </td>
                        <td
                          className={`px-4 py-2.5 text-right tabular-nums font-semibold ${
                            t.margin >= 0 ? 'text-green-700' : 'text-red-700'
                          }`}
                        >
                          ₹{t.margin.toLocaleString('en-IN', { maximumFractionDigits: 2 })}
                        </td>
                        <td className="px-4 py-2.5 text-right tabular-nums">
                          <span
                            className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                              t.margin_pct >= 30
                                ? 'bg-green-50 text-green-700'
                                : t.margin_pct >= 10
                                ? 'bg-amber-50 text-amber-700'
                                : 'bg-red-50 text-red-700'
                            }`}
                          >
                            {t.margin_pct.toFixed(1)}%
                          </span>
                        </td>
                        <td className="px-4 py-2.5 text-gray-600 text-xs">
                          {t.per_channel[0]?.channel || '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <p className="text-xs text-gray-400 mt-4">
            Window: {new Date(d.window_from).toLocaleDateString()} →{' '}
            {new Date(d.window_to).toLocaleDateString()} · Cost = what you paid the upstream
            provider. Revenue = what your child tenants paid you.
          </p>
        </>
      )}
    </div>
  )
}

function Kpi({
  label,
  value,
  sub,
  icon,
  tone = 'default',
}: {
  label: string
  value: string
  sub?: string
  icon?: React.ReactNode
  tone?: 'default' | 'primary' | 'success' | 'danger'
}) {
  const bg = {
    default: 'bg-white',
    primary: 'bg-primary-50',
    success: 'bg-green-50',
    danger: 'bg-red-50/50',
  }[tone]
  return (
    <div className={`border border-gray-200 rounded-xl p-4 ${bg}`}>
      <div className="text-xs text-gray-500 flex items-center gap-1 mb-1">
        {icon} {label}
      </div>
      <div className="text-xl font-bold text-gray-900 tabular-nums">{value}</div>
      {sub && <div className="text-xs text-gray-500 mt-0.5">{sub}</div>}
    </div>
  )
}
