import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { Wallet, MessageSquare, Users, FileText, Send, CheckCircle2, Plus, Megaphone, ScrollText, ArrowRight, Radio, Phone, Mail, Smartphone, Instagram, Facebook } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { walletService } from '../services/wallet'
import { contactsService } from '../services/contacts'
import { templatesService } from '../services/templates'
import { analyticsService } from '../services/analytics'
import { auditService } from '../services/audit'
import type { AnalyticsSummary } from '../types/analytics'
import type { AuditLogEntry } from '../types/audit'
import QualityBadge from '../components/QualityBadge'

function formatMoney(amount: number, currency: string) {
  const symbol = currency === 'INR' ? '₹' : currency + ' '
  return `${symbol}${amount.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

export default function Dashboard() {
  const { user } = useAuthStore()

  const [walletBalance, setWalletBalance] = useState<string>('—')
  const [contactCount, setContactCount] = useState<string>('—')
  const [templateCount, setTemplateCount] = useState<string>('—')
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [activity, setActivity] = useState<AuditLogEntry[]>([])

  useEffect(() => {
    let cancelled = false

    async function load() {
      const [walletResult, contactsResult, templatesResult, summaryResult, activityResult] = await Promise.allSettled([
        walletService.get(),
        contactsService.list(),
        templatesService.list(),
        analyticsService.summary(),
        auditService.list({ limit: 10 }),
      ])
      if (cancelled) return

      if (walletResult.status === 'fulfilled') {
        setWalletBalance(formatMoney(walletResult.value.balance, walletResult.value.currency))
      }
      if (contactsResult.status === 'fulfilled') {
        setContactCount(String(contactsResult.value.length))
      }
      if (templatesResult.status === 'fulfilled') {
        setTemplateCount(String(templatesResult.value.length))
      }
      if (summaryResult.status === 'fulfilled') {
        setSummary(summaryResult.value)
      }
      if (activityResult.status === 'fulfilled') {
        setActivity(activityResult.value.items)
      }
      setLoading(false)
    }
    load()

    return () => {
      cancelled = true
    }
  }, [])

  const stats = [
    {
      label: 'Wallet Balance',
      value: loading ? '…' : walletBalance,
      icon: Wallet,
      color: 'text-emerald-600 bg-emerald-50',
      href: '/wallet',
    },
    {
      label: 'Total Contacts',
      value: loading ? '…' : contactCount,
      icon: Users,
      color: 'text-purple-600 bg-purple-50',
      href: '/contacts',
    },
    {
      label: 'Message Templates',
      value: loading ? '…' : templateCount,
      icon: FileText,
      color: 'text-blue-600 bg-blue-50',
      href: '/templates',
    },
    {
      label: 'Messages Sent',
      value: loading ? '…' : (summary?.outbound_messages.toLocaleString() ?? '0'),
      icon: Send,
      color: 'text-amber-600 bg-amber-50',
      href: '/conversations',
    },
    {
      label: 'Messages Received',
      value: loading ? '…' : (summary?.inbound_messages.toLocaleString() ?? '0'),
      icon: MessageSquare,
      color: 'text-sky-600 bg-sky-50',
      href: '/conversations',
    },
    {
      label: 'Delivery Rate',
      value: loading ? '…' : (summary ? `${summary.delivery_rate.toFixed(1)}%` : '—'),
      icon: CheckCircle2,
      color: 'text-green-600 bg-green-50',
      href: '/analytics',
    },
  ]

  return (
    <div>
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Welcome back, {user?.full_name}
          </h1>
          <p className="text-gray-500 mt-1">
            Here's what's happening with your messaging today.
          </p>
        </div>
        {user && <QualityBadge status={user.quality_status} />}
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat) => (
          <Link
            key={stat.label}
            to={stat.href}
            className="rounded-xl border bg-white p-6 shadow-sm hover:shadow-md hover:border-primary-200 transition-all"
          >
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm font-medium text-gray-500">{stat.label}</span>
              <div className={`rounded-lg p-2 ${stat.color}`}>
                <stat.icon size={20} />
              </div>
            </div>
            <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
          </Link>
        ))}
      </div>

      {/* Quick actions */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <Link
          to="/contacts?action=new"
          className="flex items-center gap-3 rounded-xl border bg-white p-4 shadow-sm hover:shadow-md hover:border-primary-200 transition-all"
        >
          <div className="rounded-lg bg-purple-50 p-2.5 text-purple-600">
            <Plus size={20} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900">New Contact</p>
            <p className="text-xs text-gray-500">Add a contact to your list</p>
          </div>
        </Link>
        <Link
          to="/campaigns?action=new"
          className="flex items-center gap-3 rounded-xl border bg-white p-4 shadow-sm hover:shadow-md hover:border-primary-200 transition-all"
        >
          <div className="rounded-lg bg-amber-50 p-2.5 text-amber-600">
            <Megaphone size={20} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900">New Campaign</p>
            <p className="text-xs text-gray-500">Launch a messaging campaign</p>
          </div>
        </Link>
        <Link
          to="/templates?action=new"
          className="flex items-center gap-3 rounded-xl border bg-white p-4 shadow-sm hover:shadow-md hover:border-primary-200 transition-all"
        >
          <div className="rounded-lg bg-blue-50 p-2.5 text-blue-600">
            <ScrollText size={20} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900">New Template</p>
            <p className="text-xs text-gray-500">Create a message template</p>
          </div>
        </Link>
      </div>

      {/* Channel breakdown */}
      {summary && summary.by_channel && summary.by_channel.length > 0 && (
        <div className="rounded-xl border bg-white shadow-sm mb-8">
          <div className="px-6 py-4 border-b border-gray-100">
            <h3 className="text-lg font-semibold text-gray-900">Messages by Channel</h3>
          </div>
          <div className="p-6 grid grid-cols-2 sm:grid-cols-4 gap-4">
            {summary.by_channel.map((ch) => {
              const channelIcons: Record<string, typeof Radio> = {
                talkex: Radio,
                whatsapp: Phone,
                telegram: Send,
                email: Mail,
                sms: Smartphone,
                rcs: MessageSquare,
                instagram: Instagram,
                messenger: Facebook,
              }
              const channelColors: Record<string, string> = {
                talkex: 'text-primary-600 bg-primary-50',
                whatsapp: 'text-green-600 bg-green-50',
                telegram: 'text-blue-600 bg-blue-50',
                email: 'text-gray-600 bg-gray-50',
                sms: 'text-violet-600 bg-violet-50',
                rcs: 'text-cyan-600 bg-cyan-50',
                instagram: 'text-fuchsia-600 bg-fuchsia-50',
                messenger: 'text-indigo-600 bg-indigo-50',
              }
              const Icon = channelIcons[ch.channel] || MessageSquare
              const color = channelColors[ch.channel] || 'text-gray-600 bg-gray-50'
              return (
                <div key={ch.channel} className="flex items-center gap-3 rounded-lg border border-gray-100 p-3">
                  <div className={`rounded-lg p-2 ${color}`}>
                    <Icon size={18} />
                  </div>
                  <div>
                    <p className="text-lg font-bold text-gray-900">{ch.count.toLocaleString()}</p>
                    <p className="text-xs text-gray-500 capitalize">{ch.channel}</p>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Recent activity */}
      <div className="rounded-xl border bg-white shadow-sm">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h3 className="text-lg font-semibold text-gray-900">Recent Activity</h3>
          <Link to="/logs" className="flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700">
            View all <ArrowRight size={14} />
          </Link>
        </div>
        {activity.length === 0 ? (
          <div className="flex items-center justify-center h-32 text-sm text-gray-400">
            No activity yet.
          </div>
        ) : (
          <div className="divide-y divide-gray-50">
            {activity.map((entry) => (
              <div key={entry.id} className="flex items-center gap-4 px-6 py-3">
                <span
                  className={`inline-flex items-center justify-center rounded-lg px-2 py-1 text-[10px] font-bold tracking-wide ${
                    entry.method === 'GET'
                      ? 'bg-blue-50 text-blue-700'
                      : entry.method === 'POST'
                        ? 'bg-green-50 text-green-700'
                        : entry.method === 'DELETE'
                          ? 'bg-red-50 text-red-700'
                          : 'bg-gray-50 text-gray-700'
                  }`}
                >
                  {entry.method}
                </span>
                <span className="flex-1 truncate text-sm font-mono text-gray-700">{entry.path}</span>
                <span
                  className={`text-xs font-semibold ${entry.success ? 'text-green-600' : 'text-red-600'}`}
                >
                  {entry.status_code}
                </span>
                <span className="text-xs text-gray-400 whitespace-nowrap">
                  {new Date(entry.created_at).toLocaleString('en-IN', {
                    day: '2-digit',
                    month: 'short',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
