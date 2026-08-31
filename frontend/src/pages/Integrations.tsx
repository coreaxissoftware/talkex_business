import { useEffect, useState } from 'react'
import { Plug, ShoppingCart, FileSpreadsheet, Zap, Download, ExternalLink, Loader } from 'lucide-react'
import {
  integrationsService,
  analyticsService,
  type ZapierEvent,
  type SheetsImportResult,
} from '../services/integrations'

export default function Integrations() {
  const [events, setEvents] = useState<ZapierEvent[]>([])
  const [tab, setTab] = useState<'sheets' | 'shopify' | 'zapier' | 'crm'>('sheets')

  useEffect(() => {
    integrationsService.zapierEvents().then(setEvents).catch(() => {})
  }, [])

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <Plug size={24} className="text-primary-600" /> Integrations
        </h1>
        <p className="text-sm text-gray-500 mt-1">
          Connect TalkEx to Google Sheets, Shopify, Zapier, and your CRM.
        </p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6 flex gap-1">
        {[
          { key: 'sheets', label: 'Google Sheets', icon: FileSpreadsheet },
          { key: 'shopify', label: 'Shopify', icon: ShoppingCart },
          { key: 'zapier', label: 'Zapier', icon: Zap },
          { key: 'crm', label: 'CRM', icon: Plug },
        ].map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key as any)}
            className={`px-4 py-2 text-sm font-medium border-b-2 flex items-center gap-2 -mb-px ${
              tab === t.key
                ? 'border-primary-600 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            <t.icon size={14} /> {t.label}
          </button>
        ))}
      </div>

      {tab === 'sheets' && <SheetsPanel />}
      {tab === 'shopify' && <ShopifyPanel />}
      {tab === 'zapier' && <ZapierPanel events={events} />}
      {tab === 'crm' && <CRMPanel />}

      {/* Analytics PDF */}
      <div className="mt-8 rounded-xl border border-gray-200 bg-gray-50 p-4 flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-gray-900 text-sm flex items-center gap-2">
            <Download size={14} /> Analytics PDF report
          </h3>
          <p className="text-xs text-gray-500 mt-0.5">Download a KPI summary for accounting or sharing.</p>
        </div>
        <a
          href={analyticsService.pdfURL()}
          className="px-4 py-2 rounded-lg bg-white border border-gray-300 text-sm font-medium text-gray-700 hover:bg-gray-100 flex items-center gap-2"
        >
          <Download size={13} /> Download
        </a>
      </div>
    </div>
  )
}

function SheetsPanel() {
  const [url, setUrl] = useState('')
  const [phoneCol, setPhoneCol] = useState('A')
  const [nameCol, setNameCol] = useState('B')
  const [skipHeader, setSkipHeader] = useState(true)
  const [optIn, setOptIn] = useState(false)
  const [result, setResult] = useState<SheetsImportResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const run = async () => {
    setLoading(true)
    setError('')
    try {
      const r = await integrationsService.importFromSheet({
        url,
        phone_column: phoneCol,
        name_column: nameCol,
        skip_header: skipHeader,
        default_opt_in: optIn,
      })
      setResult(r)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Import failed.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5">
      <h2 className="font-semibold text-gray-900 mb-1">Import contacts from Google Sheets</h2>
      <p className="text-xs text-gray-500 mb-4">
        Publish your sheet to the web (File → Share → Publish to web → CSV) and paste the URL below.
      </p>

      <label className="block text-sm font-medium text-gray-700 mb-1">Published CSV URL</label>
      <input
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        placeholder="https://docs.google.com/spreadsheets/d/…/pub?output=csv"
        className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 mb-3"
      />

      <div className="grid grid-cols-2 gap-3 mb-3">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Phone column</label>
          <input
            value={phoneCol}
            onChange={(e) => setPhoneCol(e.target.value.toUpperCase())}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
            placeholder="A"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Name column</label>
          <input
            value={nameCol}
            onChange={(e) => setNameCol(e.target.value.toUpperCase())}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
            placeholder="B"
          />
        </div>
      </div>

      <div className="flex flex-col gap-2 mb-4">
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input type="checkbox" checked={skipHeader} onChange={(e) => setSkipHeader(e.target.checked)} />
          Skip first row (header)
        </label>
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input type="checkbox" checked={optIn} onChange={(e) => setOptIn(e.target.checked)} />
          Mark imported contacts as opted-in
        </label>
      </div>

      {error && <div className="mb-3 text-sm text-red-700 bg-red-50 rounded-lg px-3 py-2">{error}</div>}
      {result && (
        <div className="mb-3 text-sm bg-green-50 border border-green-200 rounded-lg px-3 py-2 text-green-800">
          Imported {result.imported} of {result.total_rows} rows ({result.skipped} skipped).
        </div>
      )}
      <button
        onClick={run}
        disabled={!url || loading}
        className="px-4 py-2 rounded-lg bg-primary-600 text-white text-sm font-semibold hover:bg-primary-700 disabled:opacity-50 flex items-center gap-2"
      >
        {loading && <Loader size={13} className="animate-spin" />} Import contacts
      </button>
    </div>
  )
}

function ShopifyPanel() {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5">
      <h2 className="font-semibold text-gray-900 mb-1">Shopify cart-abandonment</h2>
      <p className="text-sm text-gray-600 mb-4">
        Send a WhatsApp reminder when a Shopify checkout is abandoned.
      </p>
      <ol className="text-sm text-gray-700 space-y-3 list-decimal pl-5">
        <li>
          In Shopify admin, go to <b>Settings → Notifications → Webhooks</b> and add a webhook for the
          <code className="bg-gray-100 px-1 py-0.5 rounded ml-1">checkouts/abandoned</code> event.
        </li>
        <li>
          Point it at:{' '}
          <code className="bg-gray-100 px-1 py-0.5 rounded text-xs">
            https://api.business.talkex.in/integrations/shopify/webhook?owner=&lt;YOUR_OWNER_ID&gt;
          </code>
        </li>
        <li>
          Copy the webhook secret from Shopify into your TalkEx settings (Settings → Integrations →
          Shopify secret).
        </li>
        <li>Select the WhatsApp template TalkEx should send.</li>
      </ol>
      <div className="mt-4 rounded-lg bg-blue-50 border border-blue-200 px-4 py-3 text-xs text-blue-800">
        The template receives three variables: <code>{'{{1}}'}</code> first name,{' '}
        <code>{'{{2}}'}</code> cart total, <code>{'{{3}}'}</code> abandoned checkout URL.
      </div>
    </div>
  )
}

function ZapierPanel({ events }: { events: ZapierEvent[] }) {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5">
      <h2 className="font-semibold text-gray-900 mb-1">Zapier / Make / n8n triggers</h2>
      <p className="text-sm text-gray-600 mb-4">
        Use these event names when building an automation that reacts to TalkEx activity.
      </p>
      <div className="space-y-2">
        {events.map((e) => (
          <div key={e.key} className="border border-gray-200 rounded-lg p-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-semibold text-gray-900">{e.name}</p>
                <code className="text-xs text-primary-600">{e.key}</code>
              </div>
            </div>
            <p className="text-xs text-gray-500 mt-1">{e.description}</p>
            <p className="text-xs text-gray-400 mt-2 truncate">
              Payload keys: <code>{e.sample_keys.join(', ')}</code>
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}

function CRMPanel() {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5">
      <h2 className="font-semibold text-gray-900 mb-1">CRM sync (HubSpot / Salesforce / Zoho)</h2>
      <p className="text-sm text-gray-600 mb-4">
        TalkEx uses a canonical webhook shape — point your CRM's incoming webhook or Zapier bridge at
        it, or set a target URL for outbound events.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div className="border border-gray-200 rounded-lg p-3">
          <h3 className="font-semibold text-sm text-gray-900 mb-1">Inbound (CRM → TalkEx)</h3>
          <p className="text-xs text-gray-500 mb-2">POST contact updates to:</p>
          <code className="block bg-gray-100 rounded px-2 py-1 text-xs break-all">
            /integrations/crm/webhook
          </code>
          <p className="text-xs text-gray-500 mt-2">
            Body: <code>{'{event, phone, email, fields}'}</code>
          </p>
        </div>
        <div className="border border-gray-200 rounded-lg p-3">
          <h3 className="font-semibold text-sm text-gray-900 mb-1">Outbound (TalkEx → CRM)</h3>
          <p className="text-xs text-gray-500 mb-2">
            Configure a target URL in Settings → CRM. TalkEx will POST every conversation event.
          </p>
          <a
            href="/settings"
            className="text-xs text-primary-600 hover:text-primary-700 inline-flex items-center gap-1"
          >
            Open Settings <ExternalLink size={11} />
          </a>
        </div>
      </div>
    </div>
  )
}
