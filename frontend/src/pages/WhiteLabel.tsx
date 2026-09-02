import { useEffect, useState } from 'react'
import { Palette, Save, Globe, Image as ImageIcon } from 'lucide-react'
import { whitelabelService, type Branding } from '../services/whitelabel'

export default function WhiteLabel() {
  const [b, setB] = useState<Branding | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    whitelabelService.get()
      .then((res) => setB(res))
      .catch(() => setError('Could not load branding.'))
      .finally(() => setLoading(false))
  }, [])

  const save = async () => {
    if (!b) return
    setSaving(true)
    setError('')
    setMsg('')
    try {
      const updated = await whitelabelService.update(b)
      setB(updated)
      setMsg('Branding saved.')
      setTimeout(() => setMsg(''), 3000)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not save.')
    } finally {
      setSaving(false)
    }
  }

  const set = (k: keyof Branding, v: any) => {
    if (!b) return
    setB({ ...b, [k]: v })
  }

  if (loading) return <div className="p-6 text-center text-gray-500">Loading…</div>
  if (!b) return null

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <Palette size={24} className="text-primary-600" /> White-label branding
        </h1>
        <p className="text-sm text-gray-500 mt-1">
          Ship the dashboard as your own product. Applies to login, signup, and every screen.
        </p>
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}
      {msg && (
        <div className="mb-4 rounded-lg bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-700">
          {msg}
        </div>
      )}

      <div className="grid md:grid-cols-2 gap-6">
        {/* Left column — form */}
        <div className="space-y-4">
          <Section title="Identity">
            <Field label="Brand name">
              <input className="input" value={b.brand_name} onChange={(e) => set('brand_name', e.target.value)} placeholder="TalkEx Business" />
            </Field>
            <Field label="Tagline">
              <input className="input" value={b.tagline} onChange={(e) => set('tagline', e.target.value)} placeholder="One inbox. Every messaging channel." />
            </Field>
          </Section>

          <Section title="Colours">
            <div className="grid grid-cols-2 gap-3">
              <Field label="Primary">
                <div className="flex gap-2">
                  <input type="color" value={b.primary_color || '#0EA5A0'} onChange={(e) => set('primary_color', e.target.value)} className="w-12 h-10 rounded border border-gray-300" />
                  <input className="input" value={b.primary_color} onChange={(e) => set('primary_color', e.target.value)} placeholder="#0EA5A0" />
                </div>
              </Field>
              <Field label="Accent">
                <div className="flex gap-2">
                  <input type="color" value={b.accent_color || '#F97066'} onChange={(e) => set('accent_color', e.target.value)} className="w-12 h-10 rounded border border-gray-300" />
                  <input className="input" value={b.accent_color} onChange={(e) => set('accent_color', e.target.value)} placeholder="#F97066" />
                </div>
              </Field>
            </div>
          </Section>

          <Section title="Assets">
            <Field label="Logo URL (light)">
              <input className="input" value={b.logo_url} onChange={(e) => set('logo_url', e.target.value)} placeholder="https://cdn.your-brand.com/logo.svg" />
            </Field>
            <Field label="Logo URL (dark)">
              <input className="input" value={b.logo_dark_url} onChange={(e) => set('logo_dark_url', e.target.value)} placeholder="Optional — used in dark mode" />
            </Field>
            <Field label="Favicon URL">
              <input className="input" value={b.favicon_url} onChange={(e) => set('favicon_url', e.target.value)} placeholder="https://cdn.your-brand.com/favicon.ico" />
            </Field>
          </Section>

          <Section title="Domain & email">
            <Field label="Custom domain">
              <input className="input" value={b.custom_domain} onChange={(e) => set('custom_domain', e.target.value)} placeholder="app.your-brand.com" />
              <p className="text-xs text-gray-500 mt-1">Point a CNAME at <code>app.business.talkex.in</code>.</p>
            </Field>
            <Field label="From email">
              <input className="input" value={b.from_email} onChange={(e) => set('from_email', e.target.value)} placeholder="no-reply@your-brand.com" />
            </Field>
            <Field label="Support URL">
              <input className="input" value={b.support_url} onChange={(e) => set('support_url', e.target.value)} placeholder="https://support.your-brand.com" />
            </Field>
          </Section>

          <Section title="Legal & footer">
            <Field label="Privacy URL">
              <input className="input" value={b.privacy_url} onChange={(e) => set('privacy_url', e.target.value)} />
            </Field>
            <Field label="Terms URL">
              <input className="input" value={b.terms_url} onChange={(e) => set('terms_url', e.target.value)} />
            </Field>
            <label className="flex items-center gap-2 text-sm text-gray-700 mt-2">
              <input type="checkbox" checked={b.hide_powered_by} onChange={(e) => set('hide_powered_by', e.target.checked)} />
              Hide "Powered by TalkEx" chip
            </label>
          </Section>

          <button
            onClick={save}
            disabled={saving}
            className="w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
          >
            <Save size={14} /> {saving ? 'Saving…' : 'Save branding'}
          </button>
        </div>

        {/* Right column — live preview */}
        <div className="md:sticky md:top-6 md:self-start">
          <div className="border border-gray-200 rounded-xl overflow-hidden shadow-sm bg-white">
            <div className="px-4 py-2 border-b border-gray-200 bg-gray-50 flex items-center gap-2 text-xs text-gray-500">
              <Globe size={12} /> Preview
            </div>
            <div className="p-6" style={{ background: '#F7F3EE' }}>
              <div className="flex items-center gap-2 mb-6">
                {b.logo_url ? (
                  <img src={b.logo_url} alt="logo" className="h-8" onError={(e) => (e.currentTarget.style.display = 'none')} />
                ) : (
                  <div
                    className="w-9 h-9 rounded-lg flex items-center justify-center text-white font-bold"
                    style={{ background: b.primary_color || '#0EA5A0' }}
                  >
                    {b.brand_name?.[0] || 'T'}
                  </div>
                )}
                <span className="font-semibold text-gray-900">{b.brand_name || 'Brand'}</span>
              </div>
              <h2 className="font-serif italic text-2xl text-gray-900 mb-1">Welcome back</h2>
              <p className="text-sm text-gray-600 mb-5">{b.tagline || '—'}</p>
              <input className="input mb-2" placeholder="you@company.com" disabled />
              <input className="input mb-3" type="password" placeholder="••••••••" disabled />
              <button
                className="w-full rounded-lg text-white py-2.5 text-sm font-semibold"
                style={{ background: b.primary_color || '#0EA5A0' }}
              >
                Sign in
              </button>
              {!b.hide_powered_by && (
                <p className="text-center text-xs text-gray-400 mt-4">Powered by TalkEx</p>
              )}
            </div>
          </div>

          {b.logo_url && (
            <div className="mt-3 text-xs text-gray-500 flex items-start gap-1.5">
              <ImageIcon size={12} className="mt-0.5" />
              Logo loads from your CDN — the platform never stores it.
            </div>
          )}
        </div>
      </div>

      <style>{`
        .input { width: 100%; border-radius: 0.5rem; border: 1px solid #d1d5db;
          padding: 0.5rem 0.75rem; font-size: 0.875rem; outline: none; }
        .input:focus { border-color: #6366f1; box-shadow: 0 0 0 2px rgba(99,102,241,0.2); }
      `}</style>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="border border-gray-200 rounded-xl bg-white p-4 space-y-3">
      <h3 className="font-semibold text-gray-900 text-sm">{title}</h3>
      {children}
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      {children}
    </div>
  )
}
