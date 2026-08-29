import { useState, useEffect } from 'react'
import { Outlet } from 'react-router'
import Sidebar from './Sidebar'
import Header from './Header'
import { eventStream } from '../services/events'

export default function DashboardLayout() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)

  // Close mobile sidebar on route change (via resize or escape)
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 1024) setMobileOpen(false)
    }
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  // Open the SSE connection for the whole authenticated session.
  useEffect(() => {
    eventStream.connect()
    return () => eventStream.disconnect()
  }, [])

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Mobile overlay */}
      {mobileOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}

      {/* Sidebar — hidden on mobile unless mobileOpen */}
      <div
        className={`
          fixed inset-y-0 left-0 z-50 lg:z-40
          transform transition-transform duration-300
          lg:translate-x-0
          ${mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        `}
      >
        <Sidebar
          collapsed={sidebarCollapsed}
          onToggle={() => setSidebarCollapsed((c) => !c)}
          onClose={() => setMobileOpen(false)}
        />
      </div>

      <div
        className={`transition-all duration-300 ${
          sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-60'
        }`}
      >
        <Header onMenuClick={() => setMobileOpen(true)} />
        <main className="p-4 sm:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
