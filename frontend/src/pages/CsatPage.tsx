import { useEffect, useState } from 'react'
import { Star, TrendingUp, Users, Radio } from 'lucide-react'
import { csatService, type CsatSummary, type Rating } from '../services/csat'

const scoreEmoji = ['', '😞', '😕', '😐', '🙂', '😄']
const scoreLabel = ['', 'Very unhappy', 'Unhappy', 'Neutral', 'Happy', 'Very happy']

export default function CsatPage() {
  const [summary, setSummary] = useState<CsatSummary | null>(null)
  const [ratings, setRatings] = useState<Rating[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([csatService.summary(), csatService.list(50)])
      .then(([s, r]) => { setSummary(s); setRatings(r) })
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p className="p-6 text-sm text-gray-400">Loading CSAT…</p>

  const total = summary?.total ?? 0
  const avg = summary?.average ?? 0

  return (
    <div className="p-4 sm:p-6 max-w-6xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <Star size={20} className="text-primary-600" />
          Customer Satisfaction (CSAT)
        </h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Rolling 30-day window · Ratings after resolved conversations</p>
      </div>

      {/* Top metrics */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard label="Total ratings" value={total.toString()} icon={<Users size={16} />} />
        <MetricCard label="Average score" value={total ? avg.toFixed(2) : '—'} icon={<TrendingUp size={16} />} accent />
        <MetricCard label="Happy (4-5)"
          value={total ? `${Math.round(((summary!.distribution['4'] + summary!.distribution['5']) / total) * 100)}%` : '—'}
          icon={<span className="text-base">😄</span>} />
        <MetricCard label="Unhappy (1-2)"
          value={total ? `${Math.round(((summary!.distribution['1'] + summary!.distribution['2']) / total) * 100)}%` : '—'}
          icon={<span className="text-base">😞</span>} />
      </div>

      {/* Distribution */}
      {total > 0 && (
        <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-4">Score distribution</h2>
          <div className="space-y-2">
            {[5, 4, 3, 2, 1].map(score => {
              const count = summary!.distribution[String(score)] ?? 0
              const pct = total ? (count / total) * 100 : 0
              return (
                <div key={score} className="flex items-center gap-3">
                  <span className="w-24 text-xs text-gray-600 dark:text-gray-400 flex items-center gap-1">
                    <span className="text-base">{scoreEmoji[score]}</span> {scoreLabel[score]}
                  </span>
                  <div className="flex-1 h-6 rounded-full bg-gray-100 dark:bg-gray-700 overflow-hidden">
                    <div className="h-full bg-primary-500" style={{ width: `${pct}%` }} />
                  </div>
                  <span className="w-16 text-right text-xs text-gray-500 dark:text-gray-400">{count} · {pct.toFixed(0)}%</span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Per channel */}
      {summary && summary.per_channel && summary.per_channel.length > 0 && (
        <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
            <Radio size={14} /> Per channel
          </h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {summary.per_channel.map(c => (
              <div key={c.channel} className="rounded-lg border border-gray-100 dark:border-gray-700 p-3">
                <p className="text-xs uppercase text-gray-500 dark:text-gray-400">{c.channel}</p>
                <p className="text-lg font-bold text-gray-900 dark:text-gray-100">{c.average.toFixed(2)}</p>
                <p className="text-[10px] text-gray-400">{c.count} ratings</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recent ratings */}
      <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-4">Recent ratings</h2>
        {ratings.length === 0 ? (
          <p className="text-sm text-gray-400 text-center py-8">No ratings yet.</p>
        ) : (
          <div className="divide-y divide-gray-100 dark:divide-gray-700">
            {ratings.slice(0, 20).map(r => (
              <div key={r.id} className="py-3 flex items-start gap-3">
                <span className="text-xl shrink-0">{scoreEmoji[r.score] || '—'}</span>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap text-xs">
                    <span className="font-medium text-gray-700 dark:text-gray-300">Score: {r.score}/5</span>
                    {r.channel && <span className="text-gray-400">· {r.channel}</span>}
                    <span className="text-gray-400 ml-auto">{new Date(r.created_at).toLocaleDateString()}</span>
                  </div>
                  {r.comment && <p className="text-sm text-gray-600 dark:text-gray-400 mt-1 italic">"{r.comment}"</p>}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function MetricCard({ label, value, icon, accent }: { label: string; value: string; icon: React.ReactNode; accent?: boolean }) {
  return (
    <div className={`rounded-xl border p-4 ${accent
      ? 'bg-primary-50 border-primary-200 dark:bg-primary-900/20 dark:border-primary-800'
      : 'bg-white border-gray-200 dark:bg-gray-800 dark:border-gray-700'
      }`}>
      <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 mb-1.5">{icon} {label}</div>
      <p className={`text-2xl font-bold ${accent ? 'text-primary-700 dark:text-primary-300' : 'text-gray-900 dark:text-gray-100'}`}>{value}</p>
    </div>
  )
}
