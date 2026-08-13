import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  Workflow,
  Plus,
  Pencil,
  Trash2,
  X,
  Zap,
  Power,
} from 'lucide-react'
import { automationService } from '../services/automation'
import type { AutomationRule, AutomationRuleInput, MatchType } from '../types/automation'
import Modal from '../components/Modal'

interface FormState {
  name: string
  keywords: string
  match_type: MatchType
  response_body: string
  active: boolean
}

const emptyForm: FormState = {
  name: '',
  keywords: '',
  match_type: 'contains',
  response_body: '',
  active: true,
}

function keywordsToArray(s: string): string[] {
  return s.split(',').map((k) => k.trim()).filter(Boolean)
}

export default function Automation() {
  const [rules, setRules] = useState<AutomationRule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<AutomationRule | null>(null)
  const [form, setForm] = useState<FormState>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await automationService.list()
      setRules(data)
      setError('')
    } catch {
      setError('Could not load automation rules.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setFormError('')
    setModalOpen(true)
  }

  const openEdit = (r: AutomationRule) => {
    setEditing(r)
    setForm({
      name: r.name,
      keywords: r.trigger_keywords.join(', '),
      match_type: r.match_type,
      response_body: r.response_body,
      active: r.active,
    })
    setFormError('')
    setModalOpen(true)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFormError('')
    setSaving(true)
    try {
      const payload: AutomationRuleInput = {
        name: form.name,
        trigger_keywords: keywordsToArray(form.keywords),
        match_type: form.match_type,
        response_body: form.response_body,
        active: form.active,
      }
      if (payload.trigger_keywords.length === 0) {
        throw new Error('At least one trigger keyword is required.')
      }
      if (editing) {
        const updated = await automationService.update(editing.id, payload)
        setRules((prev) => prev.map((r) => (r.id === updated.id ? updated : r)))
      } else {
        const created = await automationService.create(payload)
        setRules((prev) => [created, ...prev])
      }
      setModalOpen(false)
    } catch (err: any) {
      setFormError(err.response?.data?.detail || err.message || 'Could not save rule.')
    } finally {
      setSaving(false)
    }
  }

  const handleToggleActive = async (r: AutomationRule) => {
    try {
      const updated = await automationService.update(r.id, { active: !r.active })
      setRules((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
    } catch {
      setError('Could not toggle rule.')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await automationService.remove(id)
      setRules((prev) => prev.filter((r) => r.id !== id))
    } catch {
      setError('Could not delete rule.')
    } finally {
      setConfirmDeleteId(null)
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Workflow size={24} className="text-primary-600" />
            Automation
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Auto-reply rules that fire when an inbound message matches your keywords.
          </p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          <Plus size={16} /> New Rule
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        {loading ? (
          <div className="p-10 text-center text-sm text-gray-400">Loading…</div>
        ) : rules.length === 0 ? (
          <div className="p-10 text-center">
            <Zap size={32} className="mx-auto text-gray-300 mb-2" />
            <p className="text-sm text-gray-500">No automation rules yet.</p>
            <p className="text-xs text-gray-400 mt-1">
              Create one to auto-reply when a contact messages you.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                  <th className="px-4 py-2.5">Name</th>
                  <th className="px-4 py-2.5">Trigger</th>
                  <th className="px-4 py-2.5">Reply</th>
                  <th className="px-4 py-2.5">Fired</th>
                  <th className="px-4 py-2.5">Status</th>
                  <th className="px-4 py-2.5"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {rules.map((r) => (
                  <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">{r.name}</td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {r.trigger_keywords.map((k) => (
                          <span
                            key={k}
                            className="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700"
                          >
                            {k}
                          </span>
                        ))}
                      </div>
                      <p className="text-[10px] text-gray-400 mt-1">
                        match: {r.match_type.replace('_', ' ')}
                      </p>
                    </td>
                    <td className="px-4 py-3 text-gray-700 max-w-xs truncate" title={r.response_body}>
                      {r.response_body}
                    </td>
                    <td className="px-4 py-3 text-gray-500 font-mono text-xs">{r.fire_count}</td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleToggleActive(r)}
                        className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold transition-colors ${
                          r.active
                            ? 'bg-green-50 text-green-700 hover:bg-green-100'
                            : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                        }`}
                        title="Toggle active"
                      >
                        <Power size={11} /> {r.active ? 'Active' : 'Paused'}
                      </button>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-1">
                        {confirmDeleteId === r.id ? (
                          <>
                            <button
                              onClick={() => handleDelete(r.id)}
                              className="rounded-lg px-2 py-1 text-xs font-medium text-white bg-red-600 hover:bg-red-700"
                            >
                              Confirm
                            </button>
                            <button
                              onClick={() => setConfirmDeleteId(null)}
                              className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100"
                            >
                              <X size={14} />
                            </button>
                          </>
                        ) : (
                          <>
                            <button
                              onClick={() => openEdit(r)}
                              className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
                              title="Edit"
                            >
                              <Pencil size={14} />
                            </button>
                            <button
                              onClick={() => setConfirmDeleteId(r.id)}
                              className="rounded-lg p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600"
                              title="Delete"
                            >
                              <Trash2 size={14} />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Edit Rule' : 'New Automation Rule'}>
        <form onSubmit={handleSubmit} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {formError}
            </div>
          )}
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Name <span className="text-red-500">*</span></label>
            <input
              type="text"
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. Price inquiry"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Trigger keywords <span className="text-gray-400 font-normal">(comma-separated)</span>
            </label>
            <input
              type="text"
              required
              value={form.keywords}
              onChange={(e) => setForm({ ...form, keywords: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="price, cost, pricing"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Match type</label>
            <select
              value={form.match_type}
              onChange={(e) => setForm({ ...form, match_type: e.target.value as MatchType })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none bg-white cursor-pointer"
            >
              <option value="contains">Contains — matches anywhere in the message</option>
              <option value="starts_with">Starts with — must be the first word(s)</option>
              <option value="exact">Exact — whole message must equal the keyword</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Auto-reply message <span className="text-red-500">*</span>
            </label>
            <textarea
              required
              rows={3}
              value={form.response_body}
              onChange={(e) => setForm({ ...form, response_body: e.target.value })}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="Our plans start at Rs 999/mo…"
            />
          </div>
          <label className="flex items-center gap-2 text-xs text-gray-700 cursor-pointer">
            <input
              type="checkbox"
              checked={form.active}
              onChange={(e) => setForm({ ...form, active: e.target.checked })}
              className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            Active — fire this rule on matching inbound messages
          </label>
          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
            >
              {saving ? 'Saving…' : editing ? 'Save Changes' : 'Create Rule'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
