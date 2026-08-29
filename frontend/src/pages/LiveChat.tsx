import { useEffect, useState } from 'react'
import { MessageCircle, Copy, Check, RefreshCw, Save, Eye, EyeOff } from 'lucide-react'
import { widgetService, type WidgetConfig } from '../services/widget'

const API_BASE = (import.meta.env.VITE_API_URL as string | undefined) || 'http://localhost:8080'

/**
 * Live Chat — configure the embeddable website widget: colors, greeting,
 * the public key that the JS snippet reads, and a live preview.
 */
export default function LiveChat() {
  const [cfg, setCfg] = useState<WidgetConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState('')
  const [showKey, setShowKey] = useState(false)

  const [title, setTitle] = useState('')
  const [greeting, setGreeting] = useState('')
  const [themeColor, setThemeColor] = useState('#2563eb')
  const [enabled, setEnabled] = useState(true)

  const load = async () => {
    setLoading(true)
    try {
      const c = await widgetService.get()
      setCfg(c)
      setTitle(c.title)
      setGreeting(c.greeting)
      setThemeColor(c.theme_color)
      setEnabled(c.enabled)
      setError('')
    } catch {
      setError('Could not load widget config.')
    } finally {
      setLoading(false)
    }
  }
  useEffect(() => { load() }, [])

  const handleSave = async () => {
    setSaving(true); setError('')
    try {
      const updated = await widgetService.update({ title, greeting, theme_color: themeColor, enabled })
      setCfg(updated)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Save failed')
    } finally { setSaving(false) }
  }

  const handleRotate = async () => {
    if (!confirm('Rotate the widget key? Existing embeds will stop working until you replace them.')) return
    try { setCfg(await widgetService.rotateKey()) }
    catch (err: any) { setError(err.response?.data?.detail || 'Rotate failed') }
  }

  const snippet = cfg
    ? `<script async src="${API_BASE}/widget/snippet.js" data-key="${cfg.public_key}"></script>`
    : ''

  const doCopy = () => {
    navigator.clipboard.writeText(snippet).then(() => {
      setCopied(true); setTimeout(() => setCopied(false), 1500)
    })
  }

  if (loading) return <p className="p-6 text-sm text-gray-400">Loading…</p>

  return (
    <div className="p-4 sm:p-6 max-w-5xl mx-auto space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <MessageCircle size={20} className="text-primary-600" />
          Live Chat Widget
        </h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
          Add a floating chat bubble to your website. Visitor messages land in your Conversations inbox alongside WhatsApp, Telegram, etc.
        </p>
      </div>

      {error && <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">{error}</div>}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: config form */}
        <div className="lg:col-span-2 space-y-4">
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 space-y-4">
            <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <input type="checkbox" checked={enabled} onChange={e => setEnabled(e.target.checked)}
                className="rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
              <span>Widget enabled</span>
            </label>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Panel title</label>
              <input value={title} onChange={e => setTitle(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Greeting message</label>
              <textarea rows={2} value={greeting} onChange={e => setGreeting(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Theme color</label>
              <div className="flex items-center gap-2">
                <input type="color" value={themeColor} onChange={e => setThemeColor(e.target.value)}
                  className="h-9 w-14 rounded border border-gray-300 dark:border-gray-600 cursor-pointer" />
                <input value={themeColor} onChange={e => setThemeColor(e.target.value)}
                  className="w-32 rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm font-mono outline-none focus:border-primary-500" />
              </div>
            </div>
            <button onClick={handleSave} disabled={saving}
              className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">
              <Save size={14} /> {saving ? 'Saving…' : 'Save changes'}
            </button>
          </div>

          {/* Embed snippet */}
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <div className="flex items-center justify-between mb-2">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Embed snippet</h2>
              <button onClick={handleRotate}
                className="flex items-center gap-1 text-xs font-medium text-gray-500 hover:text-red-600">
                <RefreshCw size={12} /> Rotate key
              </button>
            </div>
            <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
              Paste this <code>&lt;script&gt;</code> tag before your page's closing <code>&lt;/body&gt;</code>.
            </p>
            <div className="relative">
              <pre className="text-[11px] bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto whitespace-pre">{snippet}</pre>
              <button onClick={doCopy}
                className="absolute top-2 right-2 flex items-center gap-1 rounded-md bg-white/10 hover:bg-white/20 text-white text-[10px] px-2 py-1">
                {copied ? <><Check size={10} /> Copied</> : <><Copy size={10} /> Copy</>}
              </button>
            </div>

            <div className="mt-4 rounded-lg bg-gray-50 dark:bg-gray-700/40 p-3 text-xs">
              <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 mb-1">
                <span className="font-medium text-gray-700 dark:text-gray-300">Public key</span>
                <button onClick={() => setShowKey(v => !v)} className="text-gray-400 hover:text-primary-600">
                  {showKey ? <EyeOff size={11} /> : <Eye size={11} />}
                </button>
              </div>
              <code className="font-mono text-gray-700 dark:text-gray-300 break-all">
                {showKey ? cfg?.public_key : (cfg?.public_key?.slice(0, 8) + '••••••••••••••••••••••••••••••••••••••••')}
              </code>
              <p className="text-[10px] text-gray-400 mt-2">
                Safe to embed publicly. Rotate if a competitor scrapes it and starts spamming your inbox.
              </p>
            </div>
          </div>
        </div>

        {/* Right: live preview */}
        <div>
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Preview</h2>
            <div className="relative h-96 bg-gray-100 dark:bg-gray-900/40 rounded-lg overflow-hidden">
              {/* Mock website */}
              <div className="p-3 space-y-1.5">
                <div className="h-2 w-3/4 bg-gray-300 dark:bg-gray-700 rounded" />
                <div className="h-2 w-2/3 bg-gray-200 dark:bg-gray-700/60 rounded" />
                <div className="h-2 w-1/2 bg-gray-200 dark:bg-gray-700/60 rounded" />
                <div className="mt-4 h-2 w-4/5 bg-gray-200 dark:bg-gray-700/60 rounded" />
                <div className="h-2 w-3/5 bg-gray-200 dark:bg-gray-700/60 rounded" />
              </div>
              {/* Widget panel preview */}
              <div className="absolute right-3 bottom-3 w-64 rounded-xl overflow-hidden shadow-lg" style={{ background: '#fff' }}>
                <div className="px-3 py-2 text-white text-xs font-semibold" style={{ background: themeColor }}>
                  {title || 'Chat'}
                </div>
                <div className="p-3 bg-white text-xs h-32 overflow-hidden">
                  <div className="inline-block max-w-[90%] rounded-lg bg-gray-100 text-gray-800 px-2.5 py-1.5">
                    {greeting || 'How can we help?'}
                  </div>
                </div>
                <div className="p-2 border-t border-gray-200 flex gap-1.5">
                  <div className="flex-1 h-6 rounded border border-gray-300" />
                  <button className="rounded px-2 text-white text-[10px] font-semibold" style={{ background: themeColor }}>Send</button>
                </div>
              </div>
              {/* Bubble */}
              <div className="absolute right-3 bottom-3 h-10 w-10 rounded-full flex items-center justify-center text-white shadow-lg" style={{ background: themeColor, display: 'none' }}>
                <MessageCircle size={18} />
              </div>
            </div>
            <p className="text-[10px] text-gray-400 mt-2 text-center">Illustrative — actual widget shown once embedded.</p>
          </div>
        </div>
      </div>
    </div>
  )
}
