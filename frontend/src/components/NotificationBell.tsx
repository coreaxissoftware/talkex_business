import { useEffect, useRef, useState, useCallback } from 'react'
import { Bell, CheckCheck, Info, CheckCircle2, AlertCircle, AlertTriangle } from 'lucide-react'
import { Link } from 'react-router'
import { notificationsService } from '../services/notifications'
import type { Notification, NotificationType } from '../types/notification'

const TYPE_ICON: Record<NotificationType, React.ComponentType<{ size?: number; className?: string }>> = {
  info: Info,
  success: CheckCircle2,
  warning: AlertTriangle,
  error: AlertCircle,
}

const TYPE_ACCENT: Record<NotificationType, string> = {
  info: 'text-blue-600 bg-blue-50',
  success: 'text-green-600 bg-green-50',
  warning: 'text-amber-600 bg-amber-50',
  error: 'text-red-600 bg-red-50',
}

function relative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return 'just now'
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

export default function NotificationBell() {
  const [open, setOpen] = useState(false)
  const [items, setItems] = useState<Notification[]>([])
  const [unread, setUnread] = useState(0)
  const [loading, setLoading] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const refreshCount = useCallback(async () => {
    try {
      const n = await notificationsService.unreadCount()
      setUnread(n)
    } catch {
      /* ignore — the bell can just show stale count */
    }
  }, [])

  const loadList = useCallback(async () => {
    setLoading(true)
    try {
      const data = await notificationsService.list()
      setItems(data)
    } catch {
      /* ignore */
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refreshCount()
    // Poll every 30s so the badge stays roughly fresh without WebSockets.
    const t = setInterval(refreshCount, 30_000)
    return () => clearInterval(t)
  }, [refreshCount])

  useEffect(() => {
    if (open) loadList()
  }, [open, loadList])

  // Close on outside click.
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const handleMarkRead = async (n: Notification) => {
    if (n.read_at) return
    try {
      await notificationsService.markRead(n.id)
      setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read_at: new Date().toISOString() } : x)))
      refreshCount()
    } catch {
      /* ignore */
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await notificationsService.markAllRead()
      const now = new Date().toISOString()
      setItems((prev) => prev.map((x) => ({ ...x, read_at: x.read_at ?? now })))
      setUnread(0)
    } catch {
      /* ignore */
    }
  }

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        className="relative p-2 rounded-lg hover:bg-gray-100 transition-colors"
        title="Notifications"
      >
        <Bell size={20} className="text-gray-500" />
        {unread > 0 && (
          <span className="absolute top-1 right-1 min-w-[16px] h-4 px-1 rounded-full bg-red-500 text-white text-[10px] font-bold flex items-center justify-center">
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-80 rounded-xl border border-gray-200 bg-white shadow-lg z-50 overflow-hidden">
          <div className="flex items-center justify-between border-b border-gray-100 px-4 py-3">
            <div>
              <h3 className="text-sm font-semibold text-gray-900">Notifications</h3>
              {unread > 0 && (
                <p className="text-[10px] text-gray-500">{unread} unread</p>
              )}
            </div>
            {unread > 0 && (
              <button
                onClick={handleMarkAllRead}
                className="text-xs font-medium text-primary-600 hover:text-primary-700 flex items-center gap-1"
              >
                <CheckCheck size={12} /> Mark all read
              </button>
            )}
          </div>

          <div className="max-h-96 overflow-y-auto">
            {loading ? (
              <p className="p-6 text-center text-sm text-gray-400">Loading…</p>
            ) : items.length === 0 ? (
              <div className="p-6 text-center">
                <Bell size={24} className="mx-auto text-gray-300 mb-2" />
                <p className="text-sm text-gray-500">You're all caught up.</p>
              </div>
            ) : (
              items.map((n) => {
                const Icon = TYPE_ICON[n.type] || Info
                const accent = TYPE_ACCENT[n.type] || TYPE_ACCENT.info
                const inner = (
                  <div className={`flex gap-3 px-4 py-3 border-b border-gray-50 ${!n.read_at ? 'bg-primary-50/30' : ''}`}>
                    <div className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${accent}`}>
                      <Icon size={14} />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-start justify-between gap-2">
                        <p className={`text-sm ${!n.read_at ? 'font-semibold text-gray-900' : 'text-gray-700'} truncate`}>
                          {n.title}
                        </p>
                        {!n.read_at && (
                          <span className="mt-1 h-1.5 w-1.5 rounded-full bg-primary-600 shrink-0" />
                        )}
                      </div>
                      {n.body && (
                        <p className="text-xs text-gray-500 truncate mt-0.5">{n.body}</p>
                      )}
                      <p className="text-[10px] text-gray-400 mt-1">{relative(n.created_at)}</p>
                    </div>
                  </div>
                )
                return n.link ? (
                  <Link
                    key={n.id}
                    to={n.link}
                    onClick={() => {
                      handleMarkRead(n)
                      setOpen(false)
                    }}
                    className="block hover:bg-gray-50 transition-colors"
                  >
                    {inner}
                  </Link>
                ) : (
                  <button
                    key={n.id}
                    onClick={() => handleMarkRead(n)}
                    className="w-full text-left hover:bg-gray-50 transition-colors"
                  >
                    {inner}
                  </button>
                )
              })
            )}
          </div>
        </div>
      )}
    </div>
  )
}
