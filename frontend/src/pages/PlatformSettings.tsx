import { useEffect, useMemo, useState } from 'react'
import api from '../services/api'

// PlatformSettings — superadmin-only console for wiring up platform
// provider secrets (Mailgun, MSG91, Fast2SMS, Razorpay, OAuth apps).
// This is the "no SSH required" hatch: an operator whose email is on
// the PLATFORM_SUPERADMIN_EMAILS allow-list can drop a fresh secret
// in here and the next provider call picks it up without a
// systemctl restart.
//
// The API returns MASKED values only — a stolen support session
// cannot exfiltrate raw secrets. To rotate, you type the new value;
// to clear, you clear the input and save.

type Item = {
  key: string
  category: 'email' | 'sms' | 'payments' | 'oauth' | string
  label: string
  sensitive: boolean
  set: boolean
  from_env: boolean
  masked: string
  last_rotated_at: string
}

const CATEGORY_META: Record<string, { title: string; blurb: string }> = {
  email: {
    title: 'Email — Mailgun',
    blurb: 'Transactional email for OTP, password reset, receipts. Domain must be verified in Mailgun.',
  },
  sms: {
    title: 'SMS — MSG91 · Fast2SMS · Twilio',
    blurb: 'MSG91 + Fast2SMS are the India-first routes. Twilio is the global fallback only used when both are unset.',
  },
  payments: {
    title: 'Payments — Razorpay',
    blurb: 'Standard Checkout keys + the webhook secret from the Razorpay dashboard. Keep the webhook secret in sync with the endpoint URL you register.',
  },
  oauth: {
    title: 'OAuth providers',
    blurb: 'Client IDs are public; the secret column is what stays server-side. Apple needs the .p8 PEM with the BEGIN/END markers intact.',
  },
}

export default function PlatformSettings() {
  const [items, setItems] = useState<Item[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState<Record<string, boolean>>({})
  const [reveal, setReveal] = useState<Record<string, boolean>>({})

  useEffect(() => {
    let alive = true
    api
      .get<{ items: Item[] }>('/admin/platform-settings')
      .then((r) => {
        if (!alive) return
        setItems(r.data.items)
        setLoading(false)
      })
      .catch((e) => {
        if (!alive) return
        setError(e.response?.data?.detail || 'Failed to load platform settings.')
        setLoading(false)
      })
    return () => {
      alive = false
    }
  }, [])

  const grouped = useMemo(() => {
    const g: Record<string, Item[]> = {}
    for (const it of items) {
      ;(g[it.category] ||= []).push(it)
    }
    return g
  }, [items])

  async function save(key: string) {
    const value = drafts[key] ?? ''
    setSaving((s) => ({ ...s, [key]: true }))
    try {
      const r = await api.put<Item>(
        `/admin/platform-settings/${encodeURIComponent(key)}`,
        { value },
      )
      setItems((prev) => prev.map((p) => (p.key === key ? r.data : p)))
      setDrafts((d) => {
        const { [key]: _, ...rest } = d
        return rest
      })
    } catch (e: any) {
      alert(e.response?.data?.detail || 'Save failed')
    } finally {
      setSaving((s) => ({ ...s, [key]: false }))
    }
  }

  async function clear(key: string) {
    if (!confirm(`Clear ${key}? Provider calls will fall back to the /etc env file.`)) return
    setSaving((s) => ({ ...s, [key]: true }))
    try {
      await api.delete(`/admin/platform-settings/${encodeURIComponent(key)}`)
      // Refetch so from_env + masked reflect env fallback correctly.
      const r = await api.get<{ items: Item[] }>('/admin/platform-settings')
      setItems(r.data.items)
    } catch (e: any) {
      alert(e.response?.data?.detail || 'Clear failed')
    } finally {
      setSaving((s) => ({ ...s, [key]: false }))
    }
  }

  if (loading) {
    return <div style={{ padding: 32, color: 'var(--tw-color-gray-500, #6b7280)' }}>Loading…</div>
  }
  if (error) {
    return (
      <div style={styles.errorBox}>
        <strong>{error}</strong>
        <p style={{ marginTop: 8, opacity: 0.75 }}>
          Only superadmin operator accounts can open this page. Ask an admin to add your
          email to <code>PLATFORM_SUPERADMIN_EMAILS</code> in{' '}
          <code>/etc/talkex-business.env</code>.
        </p>
      </div>
    )
  }

  const orderedCategories = ['email', 'sms', 'payments', 'oauth']

  return (
    <div style={styles.wrap}>
      <style>{css}</style>
      <header style={styles.header}>
        <h1 style={styles.h1}>Platform provider settings</h1>
        <p style={styles.sub}>
          Live-swappable provider secrets. A save takes effect on the very next request —
          no systemctl restart needed. Values are stored in the DB and shown masked here;
          the raw string is never returned to the browser.
        </p>
      </header>

      {orderedCategories.map((cat) => {
        const rows = grouped[cat]
        if (!rows || rows.length === 0) return null
        const meta = CATEGORY_META[cat] || { title: cat, blurb: '' }
        return (
          <section key={cat} style={styles.section}>
            <div style={styles.sectionHead}>
              <h2 style={styles.h2}>{meta.title}</h2>
              <p style={styles.blurb}>{meta.blurb}</p>
            </div>
            <div style={styles.grid}>
              {rows.map((it) => {
                const draft = drafts[it.key]
                const dirty = draft !== undefined && draft !== ''
                const isRevealed = !!reveal[it.key]
                return (
                  <div key={it.key} style={styles.row}>
                    <div style={styles.rowHead}>
                      <div>
                        <div style={styles.label}>{it.label}</div>
                        <code style={styles.envkey}>{it.key}</code>
                      </div>
                      <StatusPill item={it} />
                    </div>
                    <div style={styles.currentValue}>
                      {it.set ? (
                        <span title={it.sensitive ? 'Masked' : 'Public value'}>{it.masked || '—'}</span>
                      ) : (
                        <span style={{ opacity: 0.5 }}>not set</span>
                      )}
                    </div>
                    <div style={styles.editRow}>
                      <input
                        type={it.sensitive && !isRevealed ? 'password' : 'text'}
                        value={draft ?? ''}
                        onChange={(e) =>
                          setDrafts((d) => ({ ...d, [it.key]: e.target.value }))
                        }
                        placeholder={it.set ? 'Type new value to rotate…' : 'Paste value…'}
                        style={styles.input}
                        spellCheck={false}
                        autoComplete="off"
                      />
                      {it.sensitive && (
                        <button
                          type="button"
                          onClick={() => setReveal((r) => ({ ...r, [it.key]: !r[it.key] }))}
                          style={styles.iconBtn}
                          title={isRevealed ? 'Hide' : 'Show while typing'}
                        >
                          {isRevealed ? '🙈' : '👁'}
                        </button>
                      )}
                      <button
                        type="button"
                        onClick={() => save(it.key)}
                        disabled={!dirty || saving[it.key]}
                        style={{
                          ...styles.saveBtn,
                          opacity: !dirty || saving[it.key] ? 0.5 : 1,
                          cursor: !dirty || saving[it.key] ? 'not-allowed' : 'pointer',
                        }}
                      >
                        {saving[it.key] ? 'Saving…' : 'Save'}
                      </button>
                      {it.set && !it.from_env && (
                        <button
                          type="button"
                          onClick={() => clear(it.key)}
                          disabled={saving[it.key]}
                          style={styles.clearBtn}
                          title="Delete DB override; fall back to /etc env"
                        >
                          Clear
                        </button>
                      )}
                    </div>
                    {it.last_rotated_at && it.last_rotated_at !== '0001-01-01T00:00:00Z' && (
                      <div style={styles.meta}>
                        rotated {new Date(it.last_rotated_at).toLocaleString()}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </section>
        )
      })}
    </div>
  )
}

function StatusPill({ item }: { item: Item }) {
  if (!item.set) {
    return <span style={{ ...styles.pill, background: '#F1F5F9', color: '#64748B' }}>unset</span>
  }
  if (item.from_env) {
    return <span style={{ ...styles.pill, background: '#FEF3C7', color: '#92400E' }}>from env</span>
  }
  return <span style={{ ...styles.pill, background: '#D1FAE5', color: '#065F46' }}>db override</span>
}

const styles: Record<string, React.CSSProperties> = {
  wrap: {
    maxWidth: 960,
    margin: '0 auto',
    padding: '32px 24px 64px',
    fontFamily:
      "'Inter Tight', ui-sans-serif, system-ui, -apple-system, 'Segoe UI', sans-serif",
    color: '#0A0E27',
  },
  header: { marginBottom: 28 },
  h1: {
    fontFamily: "'Instrument Serif', ui-serif, Georgia, serif",
    fontStyle: 'italic',
    fontSize: 34,
    fontWeight: 400,
    letterSpacing: -0.5,
    margin: 0,
    color: '#0A0E27',
  },
  sub: { marginTop: 8, fontSize: 14, lineHeight: 1.55, color: '#475569', maxWidth: 640 },
  section: {
    marginTop: 32,
    padding: 24,
    background: '#FFFFFF',
    border: '1px solid #E2E8F0',
    borderRadius: 12,
  },
  sectionHead: { marginBottom: 16 },
  h2: { fontSize: 15, fontWeight: 600, margin: 0, color: '#0A0E27' },
  blurb: { margin: '4px 0 0', fontSize: 13, color: '#64748B', lineHeight: 1.5 },
  grid: { display: 'grid', gap: 12 },
  row: {
    padding: 14,
    background: '#F8FAFC',
    border: '1px solid #E2E8F0',
    borderRadius: 8,
  },
  rowHead: { display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12 },
  label: { fontSize: 13, fontWeight: 600, color: '#0A0E27' },
  envkey: {
    display: 'inline-block',
    marginTop: 2,
    fontFamily: "'JetBrains Mono', ui-monospace, Menlo, monospace",
    fontSize: 11,
    color: '#64748B',
  },
  currentValue: {
    marginTop: 8,
    fontFamily: "'JetBrains Mono', ui-monospace, monospace",
    fontSize: 12,
    color: '#334155',
  },
  editRow: { marginTop: 10, display: 'flex', gap: 6 },
  input: {
    flex: 1,
    padding: '8px 10px',
    border: '1px solid #CBD5E1',
    borderRadius: 6,
    fontSize: 13,
    fontFamily: "'JetBrains Mono', ui-monospace, monospace",
    background: '#FFFFFF',
    outline: 'none',
  },
  iconBtn: {
    padding: '8px 10px',
    border: '1px solid #CBD5E1',
    borderRadius: 6,
    background: '#FFFFFF',
    cursor: 'pointer',
    fontSize: 13,
  },
  saveBtn: {
    padding: '8px 14px',
    border: 'none',
    borderRadius: 6,
    background: '#0A0E27',
    color: '#FFFFFF',
    fontWeight: 600,
    fontSize: 13,
  },
  clearBtn: {
    padding: '8px 12px',
    border: '1px solid #FCA5A5',
    borderRadius: 6,
    background: '#FEF2F2',
    color: '#B91C1C',
    fontWeight: 500,
    fontSize: 12,
    cursor: 'pointer',
  },
  pill: {
    padding: '2px 8px',
    borderRadius: 999,
    fontSize: 11,
    fontWeight: 600,
    letterSpacing: 0.3,
    textTransform: 'uppercase',
  },
  meta: { marginTop: 6, fontSize: 11, color: '#94A3B8' },
  errorBox: {
    margin: 32,
    padding: 20,
    background: '#FEF2F2',
    border: '1px solid #FCA5A5',
    borderRadius: 10,
    color: '#7F1D1D',
    maxWidth: 640,
  },
}

const css = `
input:focus { border-color: #0EA5A0 !important; box-shadow: 0 0 0 3px rgba(14,165,160,0.15); }
button:hover:not(:disabled) { filter: brightness(1.06); }
`
