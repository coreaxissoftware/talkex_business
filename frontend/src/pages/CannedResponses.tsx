import { useEffect, useState, type FormEvent } from 'react'
import { Zap, Plus, Edit3, Trash2, X, TrendingUp } from 'lucide-react'
import { cannedService, type CannedResponse } from '../services/canned'

export default function CannedResponses() {
  const [items, setItems] = useState<CannedResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<CannedResponse | null>(null)

  const [shortcut, setShortcut] = useState('')
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [category, setCategory] = useState('general')

  const load = async () => {
    setLoading(true)
    try {
      setItems(await cannedService.list())
      setError('')
    } catch {
      setError('Could not load canned responses.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const resetForm = () => {
    setShortcut(''); setTitle(''); setBody(''); setCategory('general')
    setEditing(null); setShowForm(false)
  }

  const startEdit = (r: CannedResponse) => {
    setEditing(r); setShortcut(r.shortcut); setTitle(r.title)
    setBody(r.body); setCategory(r.category); setShowForm(true)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    try {
      const payload = { shortcut: shortcut.trim(), title: title.trim(), body: body.trim(), category }
      if (editing) await cannedService.update(editing.id, payload)
      else await cannedService.create(payload)
      resetForm(); load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Save failed')
    }
  }

  const handleDelete = async (r: CannedResponse) => {
    if (!confirm(`Delete "${r.shortcut}"?`)) return
    try { await cannedService.remove(r.id); load() }
    catch { setError('Delete failed') }
  }

  return (
    <div className="p-4 sm:p-6 max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <Zap size={20} className="text-primary-600" />
            Canned Responses
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Quick-reply snippets for conversation agents. Use <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded">/shortcut</code> in the reply box.
          </p>
        </div>
        <button
          onClick={() => { resetForm(); setShowForm(true) }}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-3 py-2 text-sm font-medium text-white hover:bg-primary-700"
        >
          <Plus size={16} /> New
        </button>
      </div>

      {error && <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">{error}</div>}

      {showForm && (
        <div className="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
              {editing ? 'Edit response' : 'New canned response'}
            </h2>
            <button onClick={resetForm} className="text-gray-400 hover:text-gray-600"><X size={16} /></button>
          </div>
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Shortcut *</label>
                <input value={shortcut} onChange={e => setShortcut(e.target.value)} required
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500"
                  placeholder="/greeting" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Category</label>
                <select value={category} onChange={e => setCategory(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none">
                  <option value="general">General</option>
                  <option value="greeting">Greeting</option>
                  <option value="closing">Closing</option>
                  <option value="support">Support</option>
                  <option value="sales">Sales</option>
                  <option value="shipping">Shipping</option>
                </select>
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Title *</label>
              <input value={title} onChange={e => setTitle(e.target.value)} required
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500"
                placeholder="Standard greeting" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Body *</label>
              <textarea value={body} onChange={e => setBody(e.target.value)} required rows={4}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500"
                placeholder="Hi {{name}}, thanks for reaching out. How can we help?" />
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={resetForm} className="px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900">Cancel</button>
              <button type="submit" className="rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white hover:bg-primary-700">
                {editing ? 'Save changes' : 'Create'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <p className="text-sm text-gray-400 text-center py-12">Loading…</p>
      ) : items.length === 0 ? (
        <div className="text-center py-12">
          <Zap size={32} className="mx-auto text-gray-300 mb-3" />
          <p className="text-sm text-gray-500 dark:text-gray-400">No canned responses yet.</p>
          <p className="text-xs text-gray-400 mt-1">Save time on repeat replies — greetings, shipping updates, office hours.</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 divide-y divide-gray-100 dark:divide-gray-700">
          {items.map(r => (
            <div key={r.id} className="px-5 py-4 flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap mb-1">
                  <code className="text-xs font-mono bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 px-1.5 py-0.5 rounded">
                    {r.shortcut}
                  </code>
                  <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{r.title}</span>
                  <span className="text-[10px] uppercase text-gray-400">{r.category}</span>
                  {r.usage_count > 0 && (
                    <span className="text-[10px] text-gray-400 inline-flex items-center gap-0.5">
                      <TrendingUp size={10} /> {r.usage_count}
                    </span>
                  )}
                </div>
                <p className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">{r.body}</p>
              </div>
              <div className="flex items-center gap-1 shrink-0">
                <button onClick={() => startEdit(r)} className="p-1.5 rounded-lg text-gray-400 hover:text-primary-600 hover:bg-primary-50 dark:hover:bg-primary-900/30"><Edit3 size={14} /></button>
                <button onClick={() => handleDelete(r)} className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30"><Trash2 size={14} /></button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
