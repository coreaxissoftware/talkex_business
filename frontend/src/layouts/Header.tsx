import { LogOut, Bell, User as UserIcon } from 'lucide-react'
import { useAuthStore } from '../store/authStore'

export default function Header() {
  const { user, logout } = useAuthStore()

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b bg-white px-6">
      <div>
        <h2 className="text-sm text-gray-500">
          One Platform. Multiple Messaging Channels.
        </h2>
      </div>

      <div className="flex items-center gap-4">
        {/* Notifications placeholder */}
        <button className="relative p-2 rounded-lg hover:bg-gray-100 transition-colors">
          <Bell size={20} className="text-gray-500" />
        </button>

        {/* User menu */}
        <div className="flex items-center gap-3 pl-4 border-l">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 text-primary-700">
            <UserIcon size={16} />
          </div>
          <div className="hidden sm:block">
            <p className="text-sm font-medium text-gray-900">{user?.full_name}</p>
            <p className="text-xs text-gray-500">{user?.email}</p>
          </div>
          <button
            onClick={logout}
            className="p-2 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-600 transition-colors"
            title="Logout"
          >
            <LogOut size={18} />
          </button>
        </div>
      </div>
    </header>
  )
}
