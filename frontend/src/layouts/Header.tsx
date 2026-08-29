import { useState, useEffect } from 'react'
import { LogOut, User as UserIcon, Menu, Search, Sun, Moon } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import NotificationBell from '../components/NotificationBell'
import CommandPalette from '../components/CommandPalette'
import { useTheme } from '../hooks/useTheme'

interface HeaderProps {
  onMenuClick?: () => void
}

export default function Header({ onMenuClick }: HeaderProps) {
  const { user, logout } = useAuthStore()
  const [paletteOpen, setPaletteOpen] = useState(false)
  const { theme, toggle } = useTheme()
  const isDark = theme === 'dark'

  // Ctrl+K / Cmd+K to open
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setPaletteOpen((o) => !o)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  return (
    <>
      <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b bg-white dark:bg-gray-800 dark:border-gray-700 px-4 sm:px-6">
        <div className="flex items-center gap-3">
          {/* Mobile hamburger */}
          <button
            onClick={onMenuClick}
            className="lg:hidden p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
          >
            <Menu size={20} />
          </button>

          {/* Search bar */}
          <button
            onClick={() => setPaletteOpen(true)}
            className="flex items-center gap-2 rounded-lg border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 px-3 py-1.5 text-sm text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
          >
            <Search size={14} />
            <span className="hidden sm:inline">Search…</span>
            <kbd className="hidden sm:inline-flex items-center rounded border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 px-1.5 py-0.5 text-[10px] font-medium text-gray-400 ml-4">
              ⌘K
            </kbd>
          </button>
        </div>

        <div className="flex items-center gap-4">
          {/* Theme toggle */}
          <button
            onClick={toggle}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors"
            title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {isDark ? <Sun size={18} /> : <Moon size={18} />}
          </button>

          <NotificationBell />

          {/* User menu */}
          <div className="flex items-center gap-3 pl-4 border-l dark:border-gray-700">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 text-primary-700">
              <UserIcon size={16} />
            </div>
            <div className="hidden sm:block">
              <p className="text-sm font-medium text-gray-900 dark:text-gray-100">{user?.full_name}</p>
              <p className="text-xs text-gray-500 dark:text-gray-400">{user?.email}</p>
            </div>
            <button
              onClick={logout}
              className="p-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/30 text-gray-400 hover:text-red-600 transition-colors"
              title="Logout"
            >
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </header>

      <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </>
  )
}
