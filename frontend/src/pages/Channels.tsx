import { useEffect, useState, useCallback } from 'react'
import {
  Radio,
  MessageCircle,
  Send,
  Mail,
  Smartphone,
  MessageSquare,
  Camera,
  MessagesSquare,
  Check,
  X,
  ChevronRight,
  CheckCircle2,
  Loader2,
  Key,
  Copy,
  Sparkles,
} from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import QualityBadge from '../components/QualityBadge'
import { channelsService } from '../services/channels'
import { waOnboardingService, type WAOnboarding } from '../services/waOnboarding'
import type { ChannelCatalogItem, ChannelConfig, ChannelKind } from '../types/channel'
import Modal from '../components/Modal'

// Map catalog `icon` strings to actual lucide components. Keeping the map
// here (rather than in shared types) lets the catalog stay JSON-serialisable.
const ICONS: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  radio: Radio,
  'message-circle': MessageCircle,
  send: Send,
  mail: Mail,
  smartphone: Smartphone,
  'message-square': MessageSquare,
  instagram: Camera,          // lucide has no Instagram icon
  facebook: MessagesSquare,   // lucide has no Facebook icon
}

const ACCENTS: Record<string, string> = {
  talkex: 'bg-blue-50 text-blue-600',
  whatsapp: 'bg-green-50 text-green-600',
  telegram: 'bg-sky-50 text-sky-600',
  email: 'bg-amber-50 text-amber-600',
  sms: 'bg-purple-50 text-purple-600',
  rcs: 'bg-pink-50 text-pink-600',
  instagram: 'bg-fuchsia-50 text-fuchsia-600',
  messenger: 'bg-indigo-50 text-indigo-600',
}

export default function Channels() {
  const { user } = useAuthStore()

  const [catalog, setCatalog] = useState<ChannelCatalogItem[]>([])
  const [configs, setConfigs] = useState<ChannelConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [busyKind, setBusyKind] = useState<ChannelKind | null>(null)
  const [error, setError] = useState('')

  // TalkEx key mint modal
  const [showTalkExKey, setShowTalkExKey] = useState(false)
  const [txUsername, setTxUsername] = useState('')
  const [txPassword, setTxPassword] = useState('')
  const [txLabel, setTxLabel] = useState('TalkEx Business bridge')
  const [txPin, setTxPin] = useState('')
  const [txPendingToken, setTxPendingToken] = useState('')
  const [txMinting, setTxMinting] = useState(false)
  const [txError, setTxError] = useState('')
  const [txResultKey, setTxResultKey] = useState('')
  const [txResultPrefix, setTxResultPrefix] = useState('')
  const [txCopied, setTxCopied] = useState(false)

  const resetTalkExKeyModal = () => {
    setTxUsername('')
    setTxPassword('')
    setTxLabel('TalkEx Business bridge')
    setTxPin('')
    setTxPendingToken('')
    setTxError('')
    setTxResultKey('')
    setTxResultPrefix('')
    setTxCopied(false)
    setTxMinting(false)
  }

  const mintTalkExKey = async () => {
    setTxMinting(true)
    setTxError('')
    try {
      const res = await channelsService.generateTalkExKey({
        talkex_username: txUsername,
        talkex_password: txPassword,
        label: txLabel,
        pin: txPin || undefined,
        pending_token: txPendingToken || undefined,
      })
      if (res.requires_pin && res.pending_token) {
        setTxPendingToken(res.pending_token)
        setTxError('This account has two-step verification on. Enter your 6-digit PIN.')
        return
      }
      if (res.key) {
        setTxResultKey(res.key)
        setTxResultPrefix(res.prefix || '')
        // Refresh the channel list so the enabled toggle + saved key reflect.
        await load()
      }
    } catch (err: any) {
      setTxError(err.response?.data?.detail || 'Could not generate key.')
    } finally {
      setTxMinting(false)
    }
  }

  // WhatsApp onboarding wizard
  const [showWAWizard, setShowWAWizard] = useState(false)
  const [waOnboarding, setWaOnboarding] = useState<WAOnboarding | null>(null)
  const [waLoading, setWaLoading] = useState(false)
  const [waError, setWaError] = useState('')
  const [waBizName, setWaBizName] = useState('')
  const [waBizWebsite, setWaBizWebsite] = useState('')
  const [waBizCategory, setWaBizCategory] = useState('')
  const [waBizAddress, setWaBizAddress] = useState('')
  const [waFbId, setWaFbId] = useState('')
  const [waPhone, setWaPhone] = useState('')
  const [waDisplayName, setWaDisplayName] = useState('')

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

                {item.kind === 'talkex' && (
                  <button
                    onClick={() => {
                      resetTalkExKeyModal()
                      setShowTalkExKey(true)
                    }}
                    className="mt-2 w-full rounded-lg border border-blue-300 py-1.5 text-xs font-medium text-blue-700 hover:bg-blue-50 transition-colors flex items-center justify-center gap-1"
                  >
                    <Sparkles size={12} /> Generate TalkEx key
                  </button>
                )}

                {item.kind === 'whatsapp' && enabled && (
                  <button
                    onClick={async () => {
                      setShowWAWizard(true)
                      setWaLoading(true)
                      try {
                        let ob = await waOnboardingService.get()
                        if (!ob) ob = await waOnboardingService.start()
                        setWaOnboarding(ob)
                        setWaBizName(ob.business_name || '')
                        setWaBizWebsite(ob.business_website || '')
                        setWaBizCategory(ob.business_category || '')
                        setWaBizAddress(ob.business_address || '')
                        setWaFbId(ob.fb_business_manager_id || '')
                        setWaPhone(ob.phone_number || '')
                        setWaDisplayName(ob.display_name || '')
                      } catch { setWaError('Failed to load onboarding') }
                      finally { setWaLoading(false) }
                    }}
                    className="mt-2 w-full rounded-lg border border-green-300 py-1.5 text-xs font-medium text-green-700 hover:bg-green-50 transition-colors flex items-center justify-center gap-1"
                  >
                    <Smartphone size={12} /> WABA Setup Wizard
                  </button>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* WhatsApp Business Onboarding Wizard */}
      <Modal open={showWAWizard} onClose={() => setShowWAWizard(false)} title="WhatsApp Business Setup" wide>
        {waLoading ? (
          <div className="py-10 text-center">
            <Loader2 size={24} className="animate-spin mx-auto text-green-600" />
            <p className="text-sm text-gray-500 mt-2">Loading setup wizard…</p>
          </div>
        ) : waOnboarding ? (
          <div className="space-y-6">
            {waError && <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">{waError}</div>}

            {/* Progress indicator */}
            <div className="flex items-center gap-1">
              {['business_info', 'verification', 'phone_registration', 'display_name', 'completed'].map((step, i) => {
                const steps = ['business_info', 'verification', 'phone_registration', 'display_name', 'completed']
                const currentIdx = steps.indexOf(waOnboarding.current_step)
                const done = i < currentIdx
                const active = i === currentIdx
                return (
                  <div key={step} className="flex items-center gap-1 flex-1">
                    <div className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${
                      done ? 'bg-green-100 text-green-700' : active ? 'bg-green-600 text-white' : 'bg-gray-100 text-gray-400'
                    }`}>
                      {done ? <CheckCircle2 size={14} /> : i + 1}
                    </div>
                    {i < 4 && <div className={`flex-1 h-0.5 ${done ? 'bg-green-300' : 'bg-gray-200'}`} />}
                  </div>
                )
              })}
            </div>
            <div className="flex justify-between text-[10px] text-gray-400 -mt-4">
              <span>Business</span><span>Verify</span><span>Phone</span><span>Display</span><span>Done</span>
            </div>

            {/* Step 1: Business Info */}
            {waOnboarding.current_step === 'business_info' && (
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-900">Step 1: Business Information</h3>
                <p className="text-xs text-gray-500">Enter your business details for Facebook Business Manager verification.</p>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Business Name *</label>
                  <input type="text" required value={waBizName} onChange={e => setWaBizName(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-green-500 outline-none" placeholder="Your Business Name" />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Website</label>
                    <input type="url" value={waBizWebsite} onChange={e => setWaBizWebsite(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-green-500 outline-none" placeholder="https://..." />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-700 mb-1">Category</label>
                    <input type="text" value={waBizCategory} onChange={e => setWaBizCategory(e.target.value)}
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-green-500 outline-none" placeholder="E-commerce" />
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Business Address</label>
                  <input type="text" value={waBizAddress} onChange={e => setWaBizAddress(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-green-500 outline-none" placeholder="Full address" />
                </div>
                <button
                  disabled={!waBizName.trim() || waLoading}
                  onClick={async () => {
                    setWaLoading(true); setWaError('')
                    try {
                      const ob = await waOnboardingService.saveBusinessInfo({ business_name: waBizName, business_website: waBizWebsite, business_category: waBizCategory, business_address: waBizAddress })
                      setWaOnboarding(ob)
                    } catch { setWaError('Failed to save') }
                    finally { setWaLoading(false) }
                  }}
                  className="flex items-center gap-1 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 disabled:opacity-50"
                >
                  Next <ChevronRight size={14} />
                </button>
              </div>
            )}

            {/* Step 2: Business Verification */}
            {waOnboarding.current_step === 'verification' && (
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-900">Step 2: Facebook Business Verification</h3>
                <p className="text-xs text-gray-500">Connect your Facebook Business Manager. Go to <strong>business.facebook.com</strong> → Settings → Business Info to find your Business Manager ID.</p>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Facebook Business Manager ID *</label>
                  <input type="text" value={waFbId} onChange={e => setWaFbId(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono focus:border-green-500 outline-none" placeholder="e.g. 123456789012345" />
                </div>
                <button
                  disabled={!waFbId.trim() || waLoading}
                  onClick={async () => {
                    setWaLoading(true); setWaError('')
                    try {
                      const ob = await waOnboardingService.saveVerification(waFbId)
                      setWaOnboarding(ob)
                    } catch { setWaError('Failed to save') }
                    finally { setWaLoading(false) }
                  }}
                  className="flex items-center gap-1 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 disabled:opacity-50"
                >
                  Next <ChevronRight size={14} />
                </button>
              </div>
            )}

            {/* Step 3: Phone Registration */}
            {waOnboarding.current_step === 'phone_registration' && (
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-900">Step 3: Phone Number Registration</h3>
                <p className="text-xs text-gray-500">Register the phone number you'll use for WhatsApp Business messaging. This number must not be registered on regular WhatsApp.</p>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Phone Number *</label>
                  <input type="tel" value={waPhone} onChange={e => setWaPhone(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-green-500 outline-none" placeholder="+91 98765 43210" />
                </div>
                <button
                  disabled={!waPhone.trim() || waLoading}
                  onClick={async () => {
                    setWaLoading(true); setWaError('')
                    try {
                      const ob = await waOnboardingService.savePhoneRegistration(waPhone)
                      setWaOnboarding(ob)
                    } catch { setWaError('Failed to save') }
                    finally { setWaLoading(false) }
                  }}
                  className="flex items-center gap-1 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 disabled:opacity-50"
                >
                  Next <ChevronRight size={14} />
                </button>
              </div>
            )}

            {/* Step 4: Display Name */}
            {waOnboarding.current_step === 'display_name' && (
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-900">Step 4: Display Name Review</h3>
                <p className="text-xs text-gray-500">Set the display name that contacts will see when they receive messages from your business. This is submitted to Meta for review.</p>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Display Name *</label>
                  <input type="text" value={waDisplayName} onChange={e => setWaDisplayName(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-green-500 outline-none" placeholder="Your Business Display Name" />
                </div>
                <button
                  disabled={!waDisplayName.trim() || waLoading}
                  onClick={async () => {
                    setWaLoading(true); setWaError('')
                    try {
                      const ob = await waOnboardingService.saveDisplayName(waDisplayName)
                      setWaOnboarding(ob)
                    } catch { setWaError('Failed to save') }
                    finally { setWaLoading(false) }
                  }}
                  className="flex items-center gap-1 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 disabled:opacity-50"
                >
                  Complete Setup <CheckCircle2 size={14} />
                </button>
              </div>
            )}

            {/* Completed */}
            {waOnboarding.current_step === 'completed' && (
              <div className="text-center py-6 space-y-3">
                <CheckCircle2 size={48} className="mx-auto text-green-500" />
                <h3 className="text-lg font-semibold text-gray-900">Setup Complete!</h3>
                <p className="text-sm text-gray-500">Your WhatsApp Business Account setup has been submitted. Meta will review your display name and business verification.</p>
                <div className="rounded-lg bg-green-50 border border-green-200 p-4 text-left space-y-2 text-xs">
                  <p><strong>Business:</strong> {waOnboarding.business_name}</p>
                  <p><strong>FB Business Manager:</strong> {waOnboarding.fb_business_manager_id}</p>
                  <p><strong>Phone:</strong> {waOnboarding.phone_number}</p>
                  <p><strong>Display Name:</strong> {waOnboarding.display_name}</p>
                  <p><strong>Verification:</strong> <span className="capitalize">{waOnboarding.verification_status}</span></p>
                </div>
                <button onClick={() => setShowWAWizard(false)}
                  className="rounded-lg bg-green-600 px-6 py-2 text-sm font-semibold text-white hover:bg-green-700">
                  Done
                </button>
              </div>
            )}
          </div>
        ) : (
          <p className="text-sm text-gray-500 py-4">Could not load onboarding data.</p>
        )}
      </Modal>

      {/* TalkEx key mint modal */}
      <Modal
        open={showTalkExKey}
        onClose={() => {
          setShowTalkExKey(false)
          resetTalkExKeyModal()
        }}
        title="Generate TalkEx API key"
      >
        {txResultKey ? (
          <div className="py-2">
            <div className="mb-4 rounded-lg bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800 flex items-start gap-2">
              <CheckCircle2 size={16} className="mt-0.5 shrink-0" />
              <div>
                <p className="font-semibold">Key generated and saved.</p>
                <p className="text-xs mt-0.5">
                  Copy it below — TalkEx never shows this key again. It's already
                  wired into your TalkEx channel config.
                </p>
              </div>
            </div>

            <label className="block text-xs font-medium text-gray-700 mb-1">
              Bulk API key
            </label>
            <div className="flex gap-2 mb-4">
              <input
                readOnly
                value={txResultKey}
                onFocus={(e) => e.currentTarget.select()}
                className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm font-mono bg-gray-50"
              />
              <button
                onClick={async () => {
                  try {
                    await navigator.clipboard.writeText(txResultKey)
                    setTxCopied(true)
                    setTimeout(() => setTxCopied(false), 2000)
                  } catch {
                    // Some browsers refuse without user gesture; fall
                    // back to the manual copy the input allows.
                  }
                }}
                className="rounded-lg border border-gray-300 px-3 py-2 text-xs font-medium hover:bg-gray-100 flex items-center gap-1"
              >
                <Copy size={12} /> {txCopied ? 'Copied!' : 'Copy'}
              </button>
            </div>

            {txResultPrefix && (
              <p className="text-xs text-gray-500 mb-4">
                Prefix <code className="bg-gray-100 px-1 rounded">{txResultPrefix}</code> —
                stored on TalkEx so you can identify this key later.
              </p>
            )}

            <div className="flex justify-end">
              <button
                onClick={() => {
                  setShowTalkExKey(false)
                  resetTalkExKeyModal()
                }}
                className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
              >
                Done
              </button>
            </div>
          </div>
        ) : (
          <div className="py-2">
            <p className="text-sm text-gray-600 mb-4">
              Sign in with your TalkEx account credentials.{' '}
              <span className="text-gray-500">
                Your password is never stored — used once to mint a bulk-sending
                API key that we save into your TalkEx channel config.
              </span>
            </p>

            {txError && (
              <div className="mb-3 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
                {txError}
              </div>
            )}

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">
                  TalkEx username
                </label>
                <input
                  autoFocus
                  autoComplete="username"
                  value={txUsername}
                  onChange={(e) => setTxUsername(e.target.value)}
                  disabled={!!txPendingToken}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 disabled:bg-gray-50"
                  placeholder="e.g. yourbusiness"
                />
              </div>

              {!txPendingToken && (
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">
                    TalkEx password
                  </label>
                  <input
                    type="password"
                    autoComplete="current-password"
                    value={txPassword}
                    onChange={(e) => setTxPassword(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                  />
                </div>
              )}

              {txPendingToken && (
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">
                    Two-step PIN
                  </label>
                  <input
                    autoFocus
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    value={txPin}
                    onChange={(e) => setTxPin(e.target.value.replace(/\D/g, '').slice(0, 6))}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 font-mono tracking-widest text-center"
                    placeholder="••••••"
                    maxLength={6}
                  />
                </div>
              )}

              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">
                  Key label <span className="text-gray-400">(optional)</span>
                </label>
                <input
                  value={txLabel}
                  onChange={(e) => setTxLabel(e.target.value)}
                  disabled={!!txPendingToken}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 disabled:bg-gray-50"
                  placeholder="TalkEx Business bridge"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 mt-5">
              <button
                onClick={() => {
                  setShowTalkExKey(false)
                  resetTalkExKeyModal()
                }}
                className="px-4 py-2 text-sm rounded-lg text-gray-600 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={mintTalkExKey}
                disabled={
                  txMinting ||
                  !txUsername ||
                  (!txPendingToken && !txPassword) ||
                  (!!txPendingToken && txPin.length < 4)
                }
                className="px-4 py-2 text-sm rounded-lg bg-primary-600 text-white hover:bg-primary-700 font-semibold disabled:opacity-50 flex items-center gap-1"
              >
                {txMinting ? <Loader2 size={13} className="animate-spin" /> : <Key size={13} />}
                {txPendingToken ? 'Verify PIN & generate' : 'Generate key'}
              </button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
