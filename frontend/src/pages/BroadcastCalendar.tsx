import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router'
import { Calendar as CalendarIcon, ChevronLeft, ChevronRight, Megaphone } from 'lucide-react'
import { campaignsService } from '../services/campaigns'
import type { Campaign } from '../types/campaign'

const MONTH_NAMES = ['January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December']
const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const statusColors: Record<string, string> = {
  scheduled: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  running: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
  completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  failed: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  cancelled: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400',
  paused: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
  pending_approval: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
}

export default function BroadcastCalendar() {
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [loading, setLoading] = useState(true)
  const [current, setCurrent] = useState(new Date())

  useEffect(() => {
    campaignsService.list().then(setCampaigns).finally(() => setLoading(false))
  }, [])

  const year = current.getFullYear()
  const month = current.getMonth()

  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const firstDay = new Date(year, month, 1).getDay()

  // Group campaigns by date
  const byDate = useMemo(() => {
    const map: Record<string, Campaign[]> = {}
    campaigns.forEach(c => {
      const dateStr = c.scheduled_at || c.started_at || c.created_at
      if (!dateStr) return
      const d = new Date(dateStr)
      if (d.getFullYear() !== year || d.getMonth() !== month) return
      const key = `${d.getDate()}`
      map[key] = map[key] || []
      map[key].push(c)
    })
    return map
  }, [campaigns, year, month])

  const cells: Array<{ day: number | null }> = []
  for (let i = 0; i < firstDay; i++) cells.push({ day: null })
  for (let d = 1; d <= daysInMonth; d++) cells.push({ day: d })

  const isToday = (d: number) => {
    const t = new Date()
    return t.getFullYear() === year && t.getMonth() === month && t.getDate() === d
  }

  const prevMonth = () => setCurrent(new Date(year, month - 1, 1))
  const nextMonth = () => setCurrent(new Date(year, month + 1, 1))
  const today = () => setCurrent(new Date())

  return (
    <div className="p-4 sm:p-6 max-w-7xl mx-auto space-y-4">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <CalendarIcon size={20} className="text-primary-600" />
            Broadcast Calendar
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Scheduled and past campaigns at a glance
          </p>
        </div>
        <Link to="/campaigns" className="text-sm text-primary-600 hover:underline">
          Manage campaigns →
        </Link>
      </div>

      {loading ? (
        <p className="text-sm text-gray-400 text-center py-12">Loading calendar…</p>
      ) : (
        <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 overflow-hidden">
          {/* Month toolbar */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{MONTH_NAMES[month]} {year}</h2>
            <div className="flex items-center gap-1">
              <button onClick={today} className="text-xs px-3 py-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400">
                Today
              </button>
              <button onClick={prevMonth} className="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400">
                <ChevronLeft size={16} />
              </button>
              <button onClick={nextMonth} className="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400">
                <ChevronRight size={16} />
              </button>
            </div>
          </div>

          {/* Day headers */}
          <div className="grid grid-cols-7 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/30">
            {DAY_NAMES.map(d => (
              <div key={d} className="px-2 py-2 text-[10px] font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider text-center">{d}</div>
            ))}
          </div>

          {/* Grid */}
          <div className="grid grid-cols-7">
            {cells.map((cell, i) => {
              if (cell.day === null) return <div key={i} className="min-h-24 border-r border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/20" />
              const items = byDate[`${cell.day}`] || []
              const todayCls = isToday(cell.day) ? 'bg-primary-50 dark:bg-primary-900/10' : ''
              return (
                <div key={i} className={`min-h-24 border-r border-b border-gray-100 dark:border-gray-700 p-1.5 ${todayCls}`}>
                  <div className={`text-xs font-semibold mb-1 ${isToday(cell.day) ? 'text-primary-700 dark:text-primary-300' : 'text-gray-500 dark:text-gray-400'}`}>
                    {cell.day}
                  </div>
                  <div className="space-y-1">
                    {items.slice(0, 3).map(c => (
                      <Link
                        key={c.id}
                        to={`/campaigns`}
                        title={`${c.name} · ${c.status}`}
                        className={`block text-[10px] px-1.5 py-0.5 rounded truncate ${statusColors[c.status] || 'bg-gray-100 text-gray-700'}`}
                      >
                        <Megaphone size={9} className="inline mr-0.5" />
                        {c.name}
                      </Link>
                    ))}
                    {items.length > 3 && (
                      <span className="text-[10px] text-gray-400">+{items.length - 3} more</span>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Legend */}
      <div className="flex items-center gap-3 flex-wrap text-xs text-gray-500 dark:text-gray-400">
        {Object.entries(statusColors).map(([status, cls]) => (
          <span key={status} className="flex items-center gap-1.5">
            <span className={`w-2.5 h-2.5 rounded ${cls.split(' ')[0]}`} />
            {status.replace('_', ' ')}
          </span>
        ))}
      </div>
    </div>
  )
}
