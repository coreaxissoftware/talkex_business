import { useEffect, useState, type FormEvent } from 'react'
import {
  Workflow, Plus, Edit3, Trash2, Play, Power, PowerOff,
  MessageCircle, Clock, GitBranch, UserCheck, Tag as TagIcon, StopCircle,
  ArrowDown, X,
} from 'lucide-react'
import { flowsService, type Flow, type FlowStep, type StepType } from '../services/flows'
import { contactsService } from '../services/contacts'
import { templatesService } from '../services/templates'
import { teamService } from '../services/team'
import type { Contact } from '../types/contact'
import type { MessageTemplate } from '../types/template'
import type { TeamMember } from '../types/team'

const STEP_TYPES: { value: StepType; label: string; icon: any; desc: string }[] = [
  { value: 'send_message', label: 'Send Message', icon: MessageCircle, desc: 'Send free-form text' },
  { value: 'send_template', label: 'Send Template', icon: MessageCircle, desc: 'Send approved template' },
  { value: 'wait', label: 'Wait', icon: Clock, desc: 'Pause for N minutes' },
  { value: 'branch', label: 'Branch on Reply', icon: GitBranch, desc: 'Route by inbound keyword' },
  { value: 'assign_agent', label: 'Assign Agent', icon: UserCheck, desc: 'Hand off to a team member' },
  { value: 'add_tag', label: 'Add Tag', icon: TagIcon, desc: 'Tag the contact' },
  { value: 'end', label: 'End', icon: StopCircle, desc: 'Terminate the flow' },
]

function newStep(type: StepType): FlowStep {
  return {
    id: crypto.randomUUID(),
    type,
    label: STEP_TYPES.find(t => t.value === type)?.label || type,
  }
}

function StepIcon({ type }: { type: StepType }) {
  const cfg = STEP_TYPES.find(t => t.value === type)
  if (!cfg) return null
  const Icon = cfg.icon
  return <Icon size={14} />
}

export default function Flows() {
  const [flows, setFlows] = useState<Flow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState<Flow | null>(null)

  const load = async () => {
    setLoading(true)
    try { setFlows(await flowsService.list()); setError('') }
    catch { setError('Could not load flows.') }
    finally { setLoading(false) }
  }
  useEffect(() => { load() }, [])

  const startNew = () => {
    setEditing({
      id: '',
      owner_id: '',
      name: '',
      description: '',
      trigger_type: 'keyword',
      trigger_keywords: [],
      steps: [],
      first_step_id: '',
      active: false,
      run_count: 0,
      complete_count: 0,
      created_at: '',
      updated_at: '',
    })
  }

  const startEdit = async (f: Flow) => {
    setEditing(await flowsService.get(f.id))
  }

  const toggleActive = async (f: Flow) => {
    await flowsService.update(f.id, { active: !f.active })
    load()
  }

  const handleDelete = async (f: Flow) => {
    if (!confirm(`Delete "${f.name}"?`)) return
    await flowsService.remove(f.id)
    load()
  }

  if (editing) {
    return <FlowEditor flow={editing} onClose={() => { setEditing(null); load() }} />
  }

  return (
    <div className="p-4 sm:p-6 max-w-5xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <Workflow size={20} className="text-primary-600" />
            Chatbot Flows
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Multi-step conversation flows triggered by keywords or manually.
          </p>
        </div>
        <button onClick={startNew} className="flex items-center gap-2 rounded-lg bg-primary-600 px-3 py-2 text-sm font-medium text-white hover:bg-primary-700">
          <Plus size={16} /> New flow
        </button>
      </div>

      {error && <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">{error}</div>}

      {loading ? (
        <p className="text-sm text-gray-400 text-center py-12">Loading…</p>
      ) : flows.length === 0 ? (
        <div className="text-center py-12">
          <Workflow size={32} className="mx-auto text-gray-300 mb-3" />
          <p className="text-sm text-gray-500 dark:text-gray-400">No flows yet.</p>
          <p className="text-xs text-gray-400 mt-1">Build a chatbot: greeting → menu → route to agent.</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 divide-y divide-gray-100 dark:divide-gray-700">
          {flows.map(f => {
            const stepsArr = typeof f.steps === 'string' ? [] : f.steps
            const kwArr = typeof f.trigger_keywords === 'string' ? [] : f.trigger_keywords
            return (
              <div key={f.id} className="px-5 py-4 flex items-center justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{f.name}</span>
                    {f.active ? (
                      <span className="text-[10px] font-medium text-green-700 bg-green-100 dark:bg-green-900/30 dark:text-green-300 px-1.5 py-0.5 rounded-full">Active</span>
                    ) : (
                      <span className="text-[10px] font-medium text-gray-500 bg-gray-100 dark:bg-gray-700 dark:text-gray-400 px-1.5 py-0.5 rounded-full">Draft</span>
                    )}
                    <span className="text-[10px] text-gray-400">
                      {stepsArr.length} step{stepsArr.length !== 1 ? 's' : ''} · triggers on: {kwArr.join(', ') || '(none)'}
                    </span>
                  </div>
                  {f.description && <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-1">{f.description}</p>}
                  <p className="text-[10px] text-gray-400 mt-1">{f.run_count} runs · {f.complete_count} completed</p>
                </div>
                <div className="flex items-center gap-1 shrink-0">
                  <button onClick={() => toggleActive(f)} className="p-1.5 rounded-lg text-gray-400 hover:text-primary-600 hover:bg-primary-50 dark:hover:bg-primary-900/30" title={f.active ? 'Deactivate' : 'Activate'}>
                    {f.active ? <PowerOff size={14} /> : <Power size={14} />}
                  </button>
                  <button onClick={() => startEdit(f)} className="p-1.5 rounded-lg text-gray-400 hover:text-primary-600 hover:bg-primary-50 dark:hover:bg-primary-900/30" title="Edit"><Edit3 size={14} /></button>
                  <button onClick={() => handleDelete(f)} className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30" title="Delete"><Trash2 size={14} /></button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

// ---------------- Editor ----------------

function FlowEditor({ flow, onClose }: { flow: Flow; onClose: () => void }) {
  const [name, setName] = useState(flow.name)
  const [description, setDescription] = useState(flow.description)
  const [keywords, setKeywords] = useState<string[]>(
    typeof flow.trigger_keywords === 'string' ? [] : flow.trigger_keywords
  )
  const [kwInput, setKwInput] = useState('')
  const [steps, setSteps] = useState<FlowStep[]>(
    typeof flow.steps === 'string' ? [] : flow.steps
  )
  const [firstStepId, setFirstStepId] = useState(flow.first_step_id)
  const [active, setActive] = useState(flow.active)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const [templates, setTemplates] = useState<MessageTemplate[]>([])
  const [members, setMembers] = useState<TeamMember[]>([])
  const [contacts, setContacts] = useState<Contact[]>([])
  const [testContactId, setTestContactId] = useState('')

  useEffect(() => {
    templatesService.list().then(setTemplates).catch(() => {})
    teamService.list().then(setMembers).catch(() => {})
    contactsService.list().then(setContacts).catch(() => {})
  }, [])

  const addStep = (type: StepType) => {
    const s = newStep(type)
    // Link previous last step's next_step_id → new step for straight chains
    const updated = [...steps]
    if (updated.length > 0) {
      const last = updated[updated.length - 1]
      if (last.type !== 'branch' && last.type !== 'end' && !last.next_step_id) {
        last.next_step_id = s.id
      }
    } else {
      setFirstStepId(s.id)
    }
    setSteps([...updated, s])
  }

  const updateStep = (id: string, patch: Partial<FlowStep>) => {
    setSteps(prev => prev.map(s => s.id === id ? { ...s, ...patch } : s))
  }

  const removeStep = (id: string) => {
    setSteps(prev => {
      const filtered = prev.filter(s => s.id !== id)
      // Repair any references pointing at this step
      return filtered.map(s => ({
        ...s,
        next_step_id: s.next_step_id === id ? '' : s.next_step_id,
        branch_yes_id: s.branch_yes_id === id ? '' : s.branch_yes_id,
        branch_no_id: s.branch_no_id === id ? '' : s.branch_no_id,
      }))
    })
    if (firstStepId === id) {
      setFirstStepId(steps.filter(s => s.id !== id)[0]?.id || '')
    }
  }

  const addKeyword = () => {
    const k = kwInput.trim()
    if (!k || keywords.includes(k)) return
    setKeywords([...keywords, k])
    setKwInput('')
  }

  const handleSave = async (e: FormEvent) => {
    e.preventDefault()
    setSaving(true); setError('')
    try {
      const payload = {
        name, description,
        trigger_type: 'keyword',
        trigger_keywords: keywords,
        steps,
        first_step_id: firstStepId,
        active,
      }
      if (flow.id) await flowsService.update(flow.id, payload)
      else await flowsService.create(payload)
      onClose()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Save failed')
    } finally { setSaving(false) }
  }

  const handleTest = async () => {
    if (!flow.id || !testContactId) return
    try {
      await flowsService.test(flow.id, testContactId)
      alert('Flow triggered — check Conversations.')
    } catch (err: any) {
      alert('Test failed: ' + (err.response?.data?.detail || err.message))
    }
  }

  return (
    <div className="p-4 sm:p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <button onClick={onClose} className="text-sm text-gray-400 hover:text-primary-600 mb-1">← Back</button>
          <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
            <Workflow size={20} className="text-primary-600" />
            {flow.id ? 'Edit flow' : 'New flow'}
          </h1>
        </div>
      </div>

      {error && <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">{error}</div>}

      <form onSubmit={handleSave} className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: metadata + step list */}
        <div className="lg:col-span-2 space-y-4">
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Name *</label>
              <input required value={name} onChange={e => setName(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500"
                placeholder="Order-status bot" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Description</label>
              <input value={description} onChange={e => setDescription(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none focus:border-primary-500"
                placeholder="What this flow does" />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Trigger keywords</label>
              <div className="flex flex-wrap gap-1.5 mb-1.5">
                {keywords.map(k => (
                  <span key={k} className="inline-flex items-center gap-1 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 px-2 py-0.5 text-xs">
                    {k}
                    <button type="button" onClick={() => setKeywords(keywords.filter(x => x !== k))} className="hover:text-primary-900"><X size={10} /></button>
                  </span>
                ))}
              </div>
              <div className="flex gap-2">
                <input value={kwInput} onChange={e => setKwInput(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter') { e.preventDefault(); addKeyword() } }}
                  className="flex-1 rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-1.5 text-sm outline-none focus:border-primary-500"
                  placeholder="order, help, menu…" />
                <button type="button" onClick={addKeyword} className="rounded-lg bg-gray-100 dark:bg-gray-700 px-3 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600">Add</button>
              </div>
            </div>
            <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
              <input type="checkbox" checked={active} onChange={e => setActive(e.target.checked)}
                className="rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
              Active — fire this flow on matching inbound messages
            </label>
          </div>

          {/* Steps */}
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Steps</h2>
              <span className="text-[10px] text-gray-400">{steps.length} step{steps.length !== 1 ? 's' : ''}</span>
            </div>

            {steps.length === 0 ? (
              <p className="text-xs text-gray-400 text-center py-6">Add your first step below.</p>
            ) : (
              <div className="space-y-2">
                {steps.map((s, i) => (
                  <div key={s.id} className={`rounded-lg border ${firstStepId === s.id ? 'border-primary-300 dark:border-primary-700 bg-primary-50/50 dark:bg-primary-900/10' : 'border-gray-200 dark:border-gray-700'} p-3`}>
                    <div className="flex items-center gap-2 mb-2">
                      <span className="text-[10px] font-mono text-gray-400 w-6">#{i + 1}</span>
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-gray-700 dark:text-gray-300">
                        <StepIcon type={s.type} /> {s.label}
                      </span>
                      {firstStepId === s.id && <span className="text-[10px] text-primary-600 font-medium">START</span>}
                      <div className="ml-auto flex items-center gap-1">
                        {firstStepId !== s.id && (
                          <button type="button" onClick={() => setFirstStepId(s.id)} className="text-[10px] text-gray-400 hover:text-primary-600">Set as start</button>
                        )}
                        <button type="button" onClick={() => removeStep(s.id)} className="p-1 text-gray-400 hover:text-red-600"><Trash2 size={12} /></button>
                      </div>
                    </div>

                    <StepBody
                      step={s}
                      steps={steps}
                      templates={templates}
                      members={members}
                      onChange={(patch) => updateStep(s.id, patch)}
                    />
                  </div>
                ))}
              </div>
            )}

            {/* Add step palette */}
            <div className="mt-4 pt-4 border-t border-gray-100 dark:border-gray-700">
              <p className="text-[10px] uppercase tracking-wider text-gray-400 mb-2">Add step</p>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                {STEP_TYPES.map(t => {
                  const Icon = t.icon
                  return (
                    <button
                      key={t.value}
                      type="button"
                      onClick={() => addStep(t.value)}
                      className="flex flex-col items-center justify-center gap-1 rounded-lg border border-gray-200 dark:border-gray-700 p-2 text-[11px] text-gray-700 dark:text-gray-300 hover:border-primary-400 hover:bg-primary-50 dark:hover:bg-primary-900/20"
                      title={t.desc}
                    >
                      <Icon size={16} className="text-primary-600" />
                      {t.label}
                    </button>
                  )
                })}
              </div>
            </div>
          </div>
        </div>

        {/* Right: preview + test + save */}
        <div className="space-y-4">
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
            <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3 flex items-center gap-2">
              <ArrowDown size={14} /> Flow preview
            </h2>
            {steps.length === 0 ? (
              <p className="text-xs text-gray-400 text-center py-4">No steps yet.</p>
            ) : (
              <FlowPreview steps={steps} firstId={firstStepId} />
            )}
          </div>

          {flow.id && (
            <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5">
              <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Test flow</h2>
              <select value={testContactId} onChange={e => setTestContactId(e.target.value)}
                className="w-full rounded-lg border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-3 py-2 text-sm outline-none">
                <option value="">Select contact…</option>
                {contacts.map(c => (
                  <option key={c.id} value={c.id}>{c.name || c.phone_number}</option>
                ))}
              </select>
              <button type="button" onClick={handleTest} disabled={!testContactId}
                className="mt-2 w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-3 py-2 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-40">
                <Play size={14} /> Run test
              </button>
            </div>
          )}

          <button type="submit" disabled={saving || !name}
            className="w-full rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">
            {saving ? 'Saving…' : flow.id ? 'Save changes' : 'Create flow'}
          </button>
        </div>
      </form>
    </div>
  )
}

// ---------------- Step body editor ----------------

function StepBody({
  step, steps, templates, members, onChange,
}: {
  step: FlowStep
  steps: FlowStep[]
  templates: MessageTemplate[]
  members: TeamMember[]
  onChange: (patch: Partial<FlowStep>) => void
}) {
  const otherSteps = steps.filter(s => s.id !== step.id)
  const cls = "w-full rounded border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-2 py-1 text-xs outline-none focus:border-primary-500"

  switch (step.type) {
    case 'send_message':
      return (
        <>
          <textarea value={step.output_text || ''} onChange={e => onChange({ output_text: e.target.value })}
            placeholder="Hi {{name}}! How can we help?" rows={2} className={cls} />
          <NextStepSelector value={step.next_step_id} steps={otherSteps} onChange={id => onChange({ next_step_id: id })} />
        </>
      )
    case 'send_template':
      return (
        <>
          <select value={step.template_id || ''} onChange={e => onChange({ template_id: e.target.value })} className={cls}>
            <option value="">Choose template…</option>
            {templates.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
          </select>
          <NextStepSelector value={step.next_step_id} steps={otherSteps} onChange={id => onChange({ next_step_id: id })} />
        </>
      )
    case 'wait':
      return (
        <>
          <div className="flex items-center gap-2">
            <input type="number" min={1} value={step.wait_minutes || ''} onChange={e => onChange({ wait_minutes: parseInt(e.target.value) || 0 })}
              placeholder="Minutes" className={cls} />
            <span className="text-xs text-gray-500">min</span>
          </div>
          <NextStepSelector value={step.next_step_id} steps={otherSteps} onChange={id => onChange({ next_step_id: id })} />
        </>
      )
    case 'branch':
      return (
        <div className="space-y-1.5">
          <input value={step.branch_keyword || ''} onChange={e => onChange({ branch_keyword: e.target.value })}
            placeholder="Keyword to match in reply" className={cls} />
          <label className="block text-[10px] text-gray-500">If matches → step:</label>
          <NextStepSelector value={step.branch_yes_id} steps={otherSteps} onChange={id => onChange({ branch_yes_id: id })} />
          <label className="block text-[10px] text-gray-500">Else → step:</label>
          <NextStepSelector value={step.branch_no_id} steps={otherSteps} onChange={id => onChange({ branch_no_id: id })} />
        </div>
      )
    case 'assign_agent':
      return (
        <>
          <select value={step.agent_user_id || ''} onChange={e => onChange({ agent_user_id: e.target.value })} className={cls}>
            <option value="">Choose agent…</option>
            {members.filter(m => m.status === 'active').map(m => (
              <option key={m.id} value={m.user_id || m.id}>{m.name || m.email}</option>
            ))}
          </select>
          <NextStepSelector value={step.next_step_id} steps={otherSteps} onChange={id => onChange({ next_step_id: id })} />
        </>
      )
    case 'add_tag':
      return (
        <>
          <input value={step.tag_name || ''} onChange={e => onChange({ tag_name: e.target.value })} placeholder="e.g. lead, vip" className={cls} />
          <NextStepSelector value={step.next_step_id} steps={otherSteps} onChange={id => onChange({ next_step_id: id })} />
        </>
      )
    case 'end':
      return <p className="text-[10px] text-gray-400 italic">Flow ends here.</p>
    default:
      return null
  }
}

function NextStepSelector({ value, steps, onChange }: { value?: string; steps: FlowStep[]; onChange: (id: string) => void }) {
  return (
    <div className="flex items-center gap-2 mt-1.5">
      <span className="text-[10px] text-gray-500">→ next:</span>
      <select value={value || ''} onChange={e => onChange(e.target.value)}
        className="flex-1 rounded border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 px-2 py-0.5 text-[11px] outline-none">
        <option value="">(end)</option>
        {steps.map(s => <option key={s.id} value={s.id}>{s.label}</option>)}
      </select>
    </div>
  )
}

function FlowPreview({ steps, firstId }: { steps: FlowStep[]; firstId: string }) {
  const seen = new Set<string>()
  const chain: FlowStep[] = []
  let curId = firstId
  while (curId && !seen.has(curId)) {
    seen.add(curId)
    const s = steps.find(x => x.id === curId)
    if (!s) break
    chain.push(s)
    // Follow the "yes" branch as the preview path
    curId = s.type === 'branch' ? (s.branch_yes_id || '') : (s.next_step_id || '')
    if (s.type === 'end') break
  }
  return (
    <div className="space-y-2">
      {chain.map((s, i) => (
        <div key={s.id}>
          <div className="flex items-center gap-2 rounded-lg bg-gray-50 dark:bg-gray-700/40 px-2.5 py-1.5 text-xs">
            <StepIcon type={s.type} />
            <span className="font-medium text-gray-800 dark:text-gray-200">{s.label}</span>
          </div>
          {i < chain.length - 1 && <ArrowDown size={12} className="mx-auto text-gray-300 my-0.5" />}
        </div>
      ))}
      {chain.length === 0 && <p className="text-[10px] text-gray-400 text-center">Set a starting step.</p>}
    </div>
  )
}
