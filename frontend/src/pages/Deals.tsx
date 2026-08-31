import { useEffect, useState, type FormEvent } from 'react'
import { Briefcase, Plus, IndianRupee, TrendingUp, X, ArrowRight } from 'lucide-react'
import { dealsService, type Pipeline, type KanbanColumn, type Deal } from '../services/deals'

// Kanban-style deal pipeline. Loads the tenant's default pipeline on
// mount (seeded server-side if missing); shows one column per stage
// with drag-free "Move →" buttons for stage changes. New deals go into
// the first column by default.
export default function Deals() {
  const [pipeline, setPipeline] = useState<Pipeline | null>(null)
  const [cols, setCols] = useState<KanbanColumn[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)

  const [title, setTitle] = useState('')
  const [value, setValue] = useState('')
  const [notes, setNotes] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const p = await dealsService.defaultPipeline()
      setPipeline(p)
      setCols(await dealsService.kanban(p.id))
      setError('')
    } catch {
      setError('Could not load pipeline.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const totalValue = cols.reduce((s, c) => s + c.total_value, 0)
  const openDeals = cols
    .filter((c) => c.stage !== 'Won' && c.stage !== 'Lost')
    .reduce((s, c) => s + c.deals.length, 0)

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    if (!pipeline || !title.trim()) return
    try {
      await dealsService.create({
        pipeline_id: pipeline.id,
        title: title.trim(),
        stage: pipeline.stages[0],
        value: parseFloat(value) || 0,
        notes,
      })
      setTitle('')
      setValue('')
      setNotes('')
      setShowForm(false)
      load()
    } catch {
      setError('Could not create deal.')
    }
  }

  const move = async (deal: Deal, dir: 1 | -1) => {
    if (!pipeline) return
    const idx = pipeline.stages.indexOf(deal.stage)
    const next = pipeline.stages[idx + dir]
    if (!next) return
    try {
      await dealsService.move(deal.id, next)
      load()
    } catch {
      setError('Could not move deal.')
    }
  }

  return (
    <div className="p-6 max-w-full">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Briefcase size={24} className="text-primary-600" /> Deal Pipeline
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            {pipeline?.name || 'Sales Pipeline'} · {openDeals} open · ₹{totalValue.toLocaleString('en-IN')} total
          </p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
        >
          <Plus size={16} /> New deal
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-16 text-gray-500">Loading pipeline…</div>
      ) : (
        <div className="grid grid-flow-col auto-cols-[280px] gap-4 overflow-x-auto pb-4">
          {cols.map((col) => (
            <div key={col.stage} className="rounded-xl bg-gray-50 border border-gray-200 flex flex-col">
              <div className="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
                <div>
                  <h3 className="font-semibold text-gray-900 text-sm">{col.stage}</h3>
                  <p className="text-xs text-gray-500 flex items-center gap-1 mt-0.5">
                    <IndianRupee size={11} />
                    {col.total_value.toLocaleString('en-IN')} · {col.deals.length}
                  </p>
                </div>
                {col.stage === 'Won' && <TrendingUp size={16} className="text-green-500" />}
              </div>
              <div className="p-3 space-y-2 flex-1 max-h-[70vh] overflow-y-auto">
                {col.deals.length === 0 && (
                  <p className="text-xs text-gray-400 text-center py-6">No deals</p>
                )}
                {col.deals.map((d) => (
                  <div
                    key={d.id}
                    className="rounded-lg bg-white border border-gray-200 p-3 shadow-sm hover:shadow-md transition-shadow"
                  >
                    <p className="text-sm font-medium text-gray-900 truncate">{d.title}</p>
                    <p className="text-xs text-gray-600 mt-1 flex items-center gap-1">
                      <IndianRupee size={11} /> {d.value.toLocaleString('en-IN')}
                    </p>
                    {d.notes && (
                      <p className="text-xs text-gray-500 mt-1 line-clamp-2">{d.notes}</p>
                    )}
                    <div className="flex gap-1 mt-2">
                      <button
                        onClick={() => move(d, -1)}
                        disabled={pipeline?.stages.indexOf(d.stage) === 0}
                        className="text-xs px-2 py-1 rounded text-gray-500 hover:bg-gray-100 disabled:opacity-30"
                      >
                        ←
                      </button>
                      <button
                        onClick={() => move(d, 1)}
                        disabled={
                          pipeline?.stages.indexOf(d.stage) === (pipeline?.stages.length || 0) - 1
                        }
                        className="text-xs px-2 py-1 rounded text-primary-600 hover:bg-primary-50 disabled:opacity-30 flex items-center gap-1 ml-auto"
                      >
                        Next <ArrowRight size={11} />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {showForm && pipeline && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="font-semibold text-gray-900">Create deal</h2>
              <button
                onClick={() => setShowForm(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <X size={18} />
              </button>
            </div>
            <form onSubmit={submit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Title <span className="text-red-500">*</span>
                </label>
                <input
                  autoFocus
                  required
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                  placeholder="e.g. Enterprise inbox for ACME Corp"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">
                  Value (₹)
                </label>
                <input
                  type="number"
                  min={0}
                  step={100}
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1.5">Notes</label>
                <textarea
                  rows={3}
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 resize-none"
                />
              </div>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="px-4 py-2 text-sm rounded-lg text-gray-600 hover:bg-gray-100"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm rounded-lg bg-primary-600 text-white hover:bg-primary-700 font-semibold"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
