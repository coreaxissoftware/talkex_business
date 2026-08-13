import { useEffect, useState, useCallback } from 'react'
import {
  Radio,
  MessageCircle,
  Send,
  Mail,
  Smartphone,
  MessageSquare,
  Check,
  X,
} from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import QualityBadge from '../components/QualityBadge'
import { channelsService } from '../services/channels'
import type { ChannelCatalogItem, ChannelConfig, ChannelKind } from '../types/channel'

// Map catalog `icon` strings to actual lucide components. Keeping the map
// here (rather than in shared types) lets the catalog stay JSON-serialisable.
const ICONS: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  radio: Radio,
  'message-circle': MessageCircle,
  send: Send,
  mail: Mail,
  smartphone: Smartphone,
  'message-square': MessageSquare,
}

const ACCENTS: Record<string, string> = {
  talkex: 'bg-blue-50 text-blue-600',
  whatsapp: 'bg-green-50 text-green-600',
  telegram: 'bg-sky-50 text-sky-600',
  email: 'bg-amber-50 text-amber-600',
  sms: 'bg-purple-50 text-purple-600',
  rcs: 'bg-pink-50 text-pink-600',
}

export default function Channels() {
  const { user } = useAuthStore()

  const [catalog, setCatalog] = useState<ChannelCatalogItem[]>([])
  const [configs, setConfigs] = useState<ChannelConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [busyKind, setBusyKind] = useState<ChannelKind | null>(null)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [cat, cfgs] = await Promise.all([
        channelsService.catalog(),
        channelsService.configs(),
      ])
      setCatalog(cat)
      setConfigs(cfgs)
      setError('')
    } catch {
      setError('Could not load channels.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const configFor = (kind: ChannelKind) => configs.find((c) => c.kind === kind)

  const toggle = async (kind: ChannelKind, enabled: boolean) => {
    setBusyKind(kind)
    setError('')
    try {
      const updated = await channelsService.setEnabled(kind, enabled)
      setConfigs((prev) => {
        const idx = prev.findIndex((c) => c.kind === kind)
        if (idx === -1) return [...prev, updated]
        const next = [...prev]
        next[idx] = updated
        return next
      })
    } catch {
      setError(`Could not ${enabled ? 'enable' : 'disable'} ${kind}.`)
    } finally {
      setBusyKind(null)
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Channels</h1>
          <p className="text-sm text-gray-500 mt-1">
            Enable the channels you want to send from. Contacts, Templates, and Campaigns
            work across every enabled channel.
          </p>
        </div>
        {user && <QualityBadge status={(user as any).quality_status} />}
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="p-10 text-center text-sm text-gray-400">Loading channels…</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {catalog.map((item) => {
            const cfg = configFor(item.kind)
            const enabled = cfg?.enabled ?? false
            const Icon = ICONS[item.icon] || Radio
            const accent = ACCENTS[item.kind] || 'bg-gray-100 text-gray-600'

            return (
              <div
                key={item.kind}
                className={`rounded-xl border p-5 bg-white transition-shadow hover:shadow-sm ${
                  enabled ? 'border-primary-300' : 'border-gray-200'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${accent}`}>
                    <Icon size={20} />
                  </div>
                  {enabled ? (
                    <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-semibold text-green-700">
                      <Check size={11} /> Enabled
                    </span>
                  ) : !item.implemented ? (
                    <span className="rounded-full bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-700">
                      Coming Soon
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">
                      <X size={11} /> Disabled
                    </span>
                  )}
                </div>

                <h3 className="mt-3 text-base font-semibold text-gray-900">
                  {item.display_name}
                </h3>
                <p className="mt-1 text-xs text-gray-500 leading-relaxed">
                  {item.description}
                </p>

                <button
                  onClick={() => toggle(item.kind, !enabled)}
                  disabled={busyKind === item.kind}
                  className={`mt-4 w-full rounded-lg py-2 text-xs font-semibold transition-colors ${
                    enabled
                      ? 'border border-gray-300 text-gray-700 hover:bg-gray-50'
                      : 'bg-primary-600 text-white hover:bg-primary-700'
                  } disabled:opacity-50`}
                >
                  {busyKind === item.kind
                    ? 'Updating…'
                    : enabled
                    ? 'Disable channel'
                    : item.implemented
                    ? 'Enable channel'
                    : 'Enable (preview)'}
                </button>

                {cfg?.verified_at && (
                  <p className="mt-2 text-[10px] text-gray-400 text-center">
                    Verified {new Date(cfg.verified_at).toLocaleDateString('en-IN')}
                  </p>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
