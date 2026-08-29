import { useState, useEffect, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router'
import {
  Search,
  LayoutDashboard,
  Radio,
  Users,
  ListChecks,
  FileText,
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
  Book,
  Zap,
  Star,
  Calendar,
} from 'lucide-react'

const PAGES = [
  { path: '/', label: 'Dashboard', icon: LayoutDashboard, keywords: 'home overview stats' },
  { path: '/channels', label: 'Channels', icon: Radio, keywords: 'whatsapp telegram email' },
  { path: '/contacts', label: 'Contacts', icon: Users, keywords: 'people phone number' },
  { path: '/contact-lists', label: 'Contact Lists', icon: ListChecks, keywords: 'segments groups' },
  { path: '/tags', label: 'Tags', icon: Tag, keywords: 'labels' },
  { path: '/templates', label: 'Templates', icon: FileText, keywords: 'message body' },
  { path: '/media', label: 'Media Library', icon: Image, keywords: 'files images upload' },
  { path: '/campaigns', label: 'Campaigns', icon: Megaphone, keywords: 'bulk send broadcast' },
  { path: '/calendar', label: 'Broadcast Calendar', icon: Calendar, keywords: 'schedule scheduled campaigns month' },
  { path: '/canned-responses', label: 'Canned Responses', icon: Zap, keywords: 'quick reply snippet shortcut' },
  { path: '/csat', label: 'CSAT Ratings', icon: Star, keywords: 'satisfaction rating feedback score' },
  { path: '/conversations', label: 'Conversations', icon: MessageSquare, keywords: 'inbox chat messages' },
  { path: '/automation', label: 'Automation', icon: Workflow, keywords: 'rules triggers auto-reply' },
  { path: '/flows', label: 'Chatbot Flows', icon: Workflow, keywords: 'flow builder chatbot conversation steps' },
  { path: '/live-chat', label: 'Live Chat Widget', icon: Workflow, keywords: 'website widget embed snippet visitor' },
  { path: '/developers', label: 'Developers', icon: Code2, keywords: 'api keys playground' },
  { path: '/api-docs', label: 'API Documentation', icon: Book, keywords: 'reference endpoints' },
  { path: '/webhooks', label: 'Webhooks', icon: Webhook, keywords: 'events callbacks' },
  { path: '/analytics', label: 'Analytics', icon: BarChart3, keywords: 'charts reports metrics' },
  { path: '/logs', label: 'Audit Logs', icon: ScrollText, keywords: 'activity history' },
  { path: '/billing', label: 'Billing', icon: CreditCard, keywords: 'plan subscription invoice' },
  { path: '/wallet', label: 'Wallet', icon: Wallet, keywords: 'balance recharge credits' },
  { path: '/support', label: 'Support', icon: HelpCircle, keywords: 'help tickets' },
  { path: '/team', label: 'Team', icon: UsersRound, keywords: 'members roles invite' },
  { path: '/settings', label: 'Settings', icon: Settings, keywords: 'preferences profile security' },
  { path: '/compliance', label: 'Compliance', icon: ShieldCheck, keywords: 'dpdp privacy consent' },
  { path: '/organizations', label: 'Organizations', icon: Building, keywords: 'multi-tenant reseller' },
]

interface CommandPaletteProps {
  open: boolean
  onClose: () => void
}

export default function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState('')
  const [activeIdx, setActiveIdx] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()

  const filtered = query.trim()
    ? PAGES.filter(
        (p) =>
          p.label.toLowerCase().includes(query.toLowerCase()) ||
          p.keywords.toLowerCase().includes(query.toLowerCase())
      )
    : PAGES

  useEffect(() => {
    if (open) {
      setQuery('')
      setActiveIdx(0)
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [open])

  // Reset active index when filtered results change
  useEffect(() => {
    setActiveIdx(0)
  }, [query])

  const go = useCallback(
    (path: string) => {
      navigate(path)
      onClose()
    },
    [navigate, onClose]
  )

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActiveIdx((i) => Math.min(i + 1, filtered.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActiveIdx((i) => Math.max(i - 1, 0))
    } else if (e.key === 'Enter' && filtered[activeIdx]) {
      go(filtered[activeIdx].path)
    } else if (e.key === 'Escape') {
      onClose()
    }
  }

  // Global keyboard shortcut
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        if (open) onClose()
        else {
          // The parent manages the open state, so we rely on the
          // HeaderSearchButton's onClick; this is handled there.
        }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [open, onClose])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-lg bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
        {/* Search input */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-100 dark:border-gray-700">
          <Search size={18} className="text-gray-400 shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search pages…"
            className="flex-1 bg-transparent text-sm text-gray-900 dark:text-gray-100 placeholder-gray-400 outline-none"
          />
          <kbd className="hidden sm:inline-flex items-center gap-1 rounded border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 px-1.5 py-0.5 text-[10px] font-medium text-gray-400">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-[50vh] overflow-y-auto py-2">
          {filtered.length === 0 ? (
            <p className="px-4 py-6 text-sm text-gray-400 text-center">No results found.</p>
          ) : (
            filtered.map((page, i) => {
              const Icon = page.icon
              return (
                <button
                  key={page.path}
                  onClick={() => go(page.path)}
                  onMouseEnter={() => setActiveIdx(i)}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                    i === activeIdx
                      ? 'bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50'
                  }`}
                >
                  <Icon size={16} className="shrink-0" />
                  <span className="font-medium">{page.label}</span>
                  <span className="ml-auto text-xs text-gray-400 font-mono">{page.path}</span>
                </button>
              )
            })
          )}
        </div>

        {/* Footer hint */}
        <div className="flex items-center gap-4 px-4 py-2 border-t border-gray-100 dark:border-gray-700 text-[10px] text-gray-400">
          <span>↑↓ Navigate</span>
          <span>↵ Open</span>
          <span>ESC Close</span>
        </div>
      </div>
    </div>
  )
}
