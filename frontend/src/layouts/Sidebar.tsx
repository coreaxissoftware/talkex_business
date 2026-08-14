import { NavLink } from 'react-router'
import talkexIcon from '../assets/talkex-icon.png'
import { useAuthStore } from '../store/authStore'
import {
  LayoutDashboard,
  Radio,
  Users,
  FileText,
  ListChecks,
  Image,
  Megaphone,
  MessageSquare,
  Workflow,
  Code2,
  BarChart3,
  ScrollText,
  Tag,
  Webhook,
  CreditCard,
  Wallet,
  HelpCircle,
  Settings,
  UsersRound,
  ShieldCheck,
  Building,
  ChevronLeft,
  ChevronRight,
  LogOut,
} from 'lucide-react'

const navGroups = [
  {
    label: 'Overview',
    items: [
      { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
      { to: '/channels', icon: Radio, label: 'Channels' },
    ],
  },
  {
    label: 'Messaging',
    items: [
      { to: '/contacts', icon: Users, label: 'Contacts' },
      { to: '/contact-lists', icon: ListChecks, label: 'Lists' },
      { to: '/tags', icon: Tag, label: 'Tags' },
      { to: '/templates', icon: FileText, label: 'Templates' },
      { to: '/media', icon: Image, label: 'Media' },
      { to: '/campaigns', icon: Megaphone, label: 'Campaigns' },
      { to: '/conversations', icon: MessageSquare, label: 'Conversations' },
    ],
  },
  {
    label: 'Automation & Dev',
    items: [
      { to: '/automation', icon: Workflow, label: 'Automation' },
      { to: '/developers', icon: Code2, label: 'Developers' },
      { to: '/webhooks', icon: Webhook, label: 'Webhooks' },
      { to: '/analytics', icon: BarChart3, label: 'Analytics' },
      { to: '/logs', icon: ScrollText, label: 'Logs' },
    ],
  },
  {
    label: 'Account',
    items: [
      { to: '/billing', icon: CreditCard, label: 'Billing' },
      { to: '/wallet', icon: Wallet, label: 'Wallet' },
      { to: '/support', icon: HelpCircle, label: 'Support' },
      { to: '/team', icon: UsersRound, label: 'Team' },
      { to: '/settings', icon: Settings, label: 'Settings' },
      { to: '/compliance', icon: ShieldCheck, label: 'Compliance' },
      { to: '/organizations', icon: Building, label: 'Organizations' },
    ],
  },
]

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

export default function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { logout } = useAuthStore()

  return (
    <aside
      className={`fixed left-0 top-0 z-40 h-screen bg-sidebar text-white transition-all duration-300 flex flex-col ${
        collapsed ? 'w-16' : 'w-60'
      }`}
    >
      {/* Logo */}
      <div className="flex h-14 items-center justify-between px-3 border-b border-white/10">
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
          className="p-1 rounded-lg hover:bg-sidebar-hover transition-colors"
        >
          {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-2 px-2">
        {navGroups.map((group) => (
          <div key={group.label} className="mb-1">
            {/* Group label */}
            {!collapsed && (
              <p className="px-2 pt-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-gray-500">
                {group.label}
              </p>
            )}
            {collapsed && <div className="my-1 mx-2 border-t border-white/10" />}

            {/* Group items */}
            {group.items.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) =>
                  `flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-sidebar-active text-white'
                      : 'text-gray-300 hover:bg-sidebar-hover hover:text-white'
                  }`
                }
              >
                <item.icon size={18} className="shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </NavLink>
            ))}
          </div>
        ))}
      </nav>

      {/* Footer with signout */}
      <div className="border-t border-white/10 px-3 py-2.5 flex items-center justify-between">
        {!collapsed ? (
          <>
            <div>
              <p className="text-xs text-gray-400 leading-tight">TalkEx Business</p>
              <p className="text-[10px] text-gray-500 leading-tight">by CoreAxis Ventures</p>
            </div>
            <button
              onClick={logout}
              className="p-1.5 rounded-lg text-gray-400 hover:bg-red-500/20 hover:text-red-400 transition-colors"
              title="Sign out"
            >
              <LogOut size={16} />
            </button>
          </>
        ) : (
          <button
            onClick={logout}
            className="mx-auto p-1.5 rounded-lg text-gray-400 hover:bg-red-500/20 hover:text-red-400 transition-colors"
            title="Sign out"
          >
            <LogOut size={18} />
          </button>
        )}
      </div>
    </aside>
  )
}
