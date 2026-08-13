import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { Wallet, MessageSquare, Users, FileText, Send, CheckCircle2 } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { walletService } from '../services/wallet'
import { contactsService } from '../services/contacts'
import { templatesService } from '../services/templates'
import { analyticsService } from '../services/analytics'
import type { AnalyticsSummary } from '../types/analytics'
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

  useEffect(() => {
    let cancelled = false

    async function load() {
      const [walletResult, contactsResult, templatesResult, summaryResult] = await Promise.allSettled([
        walletService.get(),
        contactsService.list(),
        templatesService.list(),
        analyticsService.summary(),
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

      {/* Placeholder chart area */}
      <div className="rounded-xl border bg-white p-6 shadow-sm">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Message Activity</h3>
        <div className="flex items-center justify-center h-64 text-gray-400">
          <p>Chart will appear here once messages are sent.</p>
        </div>
      </div>
    </div>
  )
}
