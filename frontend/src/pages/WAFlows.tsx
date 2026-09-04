import { useEffect, useState, type FormEvent } from 'react'
import { Boxes, Plus, X, Send, CheckCircle2, Clock, Ban, Loader2 } from 'lucide-react'
import { waflowsService, type WAFlow, type FlowCategory } from '../services/waflows'

const STATUS_CHIP: Record<WAFlow['status'], string> = {
  draft: 'bg-gray-100 text-gray-700',
  published: 'bg-green-50 text-green-700',
  deprecated: 'bg-amber-50 text-amber-700',
  blocked: 'bg-red-50 text-red-700',
}
const STATUS_ICON: Record<WAFlow['status'], any> = {
  draft: Clock,
  published: CheckCircle2,
  deprecated: Ban,
  blocked: Ban,
}

const CATEGORIES: FlowCategory[] = [
  'SIGN_UP',
  'SIGN_IN',
  'APPOINTMENT_BOOKING',
  'LEAD_GENERATION',
  'SHOPPING',
  'CONTACT_US',
  'SURVEY',
  'OTHER',
]

// A minimal valid Flow JSON — one screen with a single text input.
// The merchant edits from here or pastes their own from Meta's Flow Builder.
const SAMPLE_FLOW = {
  version: '3.0',
  routing_model: { WELCOME: [] },
  screens: [
    {
      id: 'WELCOME',
      title: 'Welcome',
      terminal: true,
      data: {},
      layout: {
        type: 'SingleColumnLayout',
        children: [
          {
            type: 'TextInput',
            required: true,
            label: 'Your name',
            name: 'name',
          },
          {
            type: 'Footer',
            label: 'Submit',
            'on-click-action': {
              name: 'complete',
              payload: { name: '${form.name}' },
            },
          },
        ],
      },
    },
  ],
}

export default function WAFlows() {
  const [items, setItems] = useState<WAFlow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [category, setCategory] = useState<FlowCategory>('OTHER')
  const [flowJSONText, setFlowJSONText] = useState(JSON.stringify(SAMPLE_FLOW, null, 2))
  const [submitting, setSubmitting] = useState(false)
  const [publishing, setPublishing] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      setItems(await waflowsService.list())
      setError('')
    } catch {
      setError('Could not load flows.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    let parsed: any
    try {
      parsed = JSON.parse(flowJSONText)
    } catch {
      setError('Flow JSON is not valid JSON.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      await waflowsService.create({ name, category, flow_json: parsed })
      setName('')
      setCategory('OTHER')
      setFlowJSONText(JSON.stringify(SAMPLE_FLOW, null, 2))
      setShowForm(false)
      load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not create flow.')
    } finally {
      setSubmitting(false)
    }
  }

  const publish = async (f: WAFlow) => {
    if (!confirm(`Publish "${f.name}"? Once published, edits create a new version.`)) return
    setPublishing(f.id)
    try {
      await waflowsService.publish(f.id)
      load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Publish failed.')
    } finally {
      setPublishing('')
    }
  }

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Boxes size={24} className="text-primary-600" /> WhatsApp Interactive Flows
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            In-chat forms, screens, and data collection — Meta Flows JSON.
          </p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
        >
          <Plus size={16} /> New flow
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-16 text-gray-500">Loading…</div>
      ) : items.length === 0 ? (
        <div className="text-center py-16 border-2 border-dashed border-gray-200 rounded-xl">
          <Boxes size={40} className="mx-auto text-gray-300 mb-3" />
          <p className="text-gray-500 mb-3">No flows yet.</p>
          <p className="text-xs text-gray-400">
            Build one with the sample template, or paste JSON from Meta's Flow Builder.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {items.map((f) => {
            const Icon = STATUS_ICON[f.status]
            return (
              <div key={f.id} className="border border-gray-200 rounded-xl bg-white p-4">
                <div className="flex items-start justify-between gap-2 mb-2">
                  <h3 className="font-semibold text-gray-900 text-sm">{f.name}</h3>
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_CHIP[f.status]}`}>
                    <Icon size={11} /> {f.status}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-2">
                  {f.category.replace(/_/g, ' ')} · v{f.version}
                </p>
                {f.meta_flow_id && (
                  <p className="text-[10px] text-gray-400 font-mono mb-3 truncate">
                    Meta ID: {f.meta_flow_id}
                  </p>
                )}
                {f.status === 'draft' && (
                  <button
                    onClick={() => publish(f)}
                    disabled={publishing === f.id}
                    className="w-full flex items-center justify-center gap-1 rounded-lg border border-green-300 py-1.5 text-xs font-medium text-green-700 hover:bg-green-50 disabled:opacity-50"
                  >
                    {publishing === f.id ? <Loader2 size={12} className="animate-spin" /> : <Send size={12} />}
                    Publish to Meta
                  </button>
                )}
                {f.published_at && (
                  <p className="text-[10px] text-gray-400 mt-2">
                    Published {new Date(f.published_at).toLocaleDateString('en-IN')}
                  </p>
                )}
              </div>
            )
          })}
        </div>
      )}

      {showForm && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between sticky top-0 bg-white z-10">
              <h2 className="font-semibold text-gray-900">New WhatsApp Flow</h2>
              <button onClick={() => setShowForm(false)} className="text-gray-400 hover:text-gray-600">
                <X size={18} />
              </button>
            </div>
            <form onSubmit={submit} className="p-6 space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                  <input
                    required
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                    placeholder="Lead capture form"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
                  <select
                    value={category}
                    onChange={(e) => setCategory(e.target.value as FlowCategory)}
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
                  >
                    {CATEGORIES.map((c) => (
                      <option key={c} value={c}>{c.replace(/_/g, ' ')}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="block text-sm font-medium text-gray-700">Flow JSON</label>
                  <button
                    type="button"
                    onClick={() => setFlowJSONText(JSON.stringify(SAMPLE_FLOW, null, 2))}
                    className="text-xs text-primary-600 hover:text-primary-700"
                  >
                    Reset to sample
                  </button>
                </div>
                <textarea
                  required
                  value={flowJSONText}
                  onChange={(e) => setFlowJSONText(e.target.value)}
                  rows={16}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-xs font-mono outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                  spellCheck={false}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Meta Flows JSON — must have a <code>routing_model</code> and at least one <code>screen</code>.
                </p>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="px-4 py-2 text-sm rounded-lg text-gray-600 hover:bg-gray-100"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="px-4 py-2 text-sm rounded-lg bg-primary-600 text-white hover:bg-primary-700 font-semibold disabled:opacity-50"
                >
                  {submitting ? 'Creating…' : 'Save as draft'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
