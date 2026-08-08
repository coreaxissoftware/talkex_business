import { Wallet, MessageSquare, Users, TrendingUp } from 'lucide-react'
import { useAuthStore } from '../store/authStore'

const stats = [
  { label: 'Wallet Balance', value: '₹0.00', icon: Wallet, color: 'text-emerald-600 bg-emerald-50' },
  { label: 'Messages Sent', value: '0', icon: MessageSquare, color: 'text-blue-600 bg-blue-50' },
  { label: 'Total Contacts', value: '0', icon: Users, color: 'text-purple-600 bg-purple-50' },
  { label: 'Delivery Rate', value: '—', icon: TrendingUp, color: 'text-amber-600 bg-amber-50' },
]

export default function Dashboard() {
  const { user } = useAuthStore()

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">
          Welcome back, {user?.full_name}
        </h1>
        <p className="text-gray-500 mt-1">
          Here's what's happening with your messaging today.
        </p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="rounded-xl border bg-white p-6 shadow-sm"
          >
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm font-medium text-gray-500">{stat.label}</span>
              <div className={`rounded-lg p-2 ${stat.color}`}>
                <stat.icon size={20} />
              </div>
            </div>
            <p className="text-2xl font-bold text-gray-900">{stat.value}</p>
          </div>
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
