import { NavLink } from 'react-router'
import talkexIcon from '../assets/talkex-icon.png'
import {
  LayoutDashboard,
  Radio,
  Users,
  FileText,
  Megaphone,
  MessageSquare,
  Workflow,
  Code2,
  BarChart3,
  CreditCard,
  Wallet,
  HelpCircle,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/channels', icon: Radio, label: 'Channels' },
  { to: '/contacts', icon: Users, label: 'Contacts' },
  { to: '/templates', icon: FileText, label: 'Templates' },
  { to: '/campaigns', icon: Megaphone, label: 'Campaigns' },
  { to: '/conversations', icon: MessageSquare, label: 'Conversations' },
  { to: '/automation', icon: Workflow, label: 'Automation' },
  { to: '/developers', icon: Code2, label: 'Developers' },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/billing', icon: CreditCard, label: 'Billing' },
  { to: '/wallet', icon: Wallet, label: 'Wallet' },
  { to: '/support', icon: HelpCircle, label: 'Support' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

export default function Sidebar({ collapsed, onToggle }: SidebarProps) {
  return (
    <aside
      className={`fixed left-0 top-0 z-40 h-screen bg-sidebar text-white transition-all duration-300 flex flex-col ${
        collapsed ? 'w-16' : 'w-60'
      }`}
    >
      {/* Logo */}
      <div className="flex h-16 items-center justify-between px-4 border-b border-white/10">
        <div className="flex items-center gap-2">
          <img src={talkexIcon} alt="TalkEx" className="h-8 w-8 rounded-lg" />
          {!collapsed && (
            <span className="text-lg font-bold tracking-tight">
              Talk<span className="text-primary-400">Ex</span>
            </span>
          )}
        </div>
        <button
          onClick={onToggle}
          className="p-1.5 rounded-lg hover:bg-sidebar-hover transition-colors"
        >
          {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-4 space-y-1 px-2">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-sidebar-active text-white'
                  : 'text-gray-300 hover:bg-sidebar-hover hover:text-white'
              }`
            }
          >
            <item.icon size={20} className="shrink-0" />
            {!collapsed && <span>{item.label}</span>}
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      {!collapsed && (
        <div className="border-t border-white/10 px-4 py-3">
          <p className="text-xs text-gray-400">TalkEx Business</p>
          <p className="text-[10px] text-gray-500">by CoreAxis Ventures</p>
        </div>
      )}
    </aside>
  )
}
