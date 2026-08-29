import { useEffect, useState, useCallback, useRef, type FormEvent } from 'react'
import {
  MessageSquare,
  Send,
  Clock,
  Lock,
  User as UserIcon,
  Inbox,
  MessageCircleReply,
  Tag,
  UserCheck,
  X,
  Plus,
  Zap,
  Star,
} from 'lucide-react'
import CannedPicker from '../components/CannedPicker'
import { csatService } from '../services/csat'
import { conversationsService } from '../services/conversations'
import { contactsService } from '../services/contacts'
import { teamService } from '../services/team'
import type { TeamMember } from '../types/team'
import type {
  ConversationRow,
  ConversationThread,
  Message,
} from '../types/conversation'
import type { Contact } from '../types/contact'

function relativeTime(iso: string | null): string {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h`
  const d = Math.floor(h / 24)
  return `${d}d`
}

function initials(name: string | null, phone: string): string {
  if (name && name.trim()) {
    const parts = name.trim().split(/\s+/)
    return (parts[0][0] + (parts[1]?.[0] || '')).toUpperCase()
  }
  return phone.slice(-2)
}

function MessageBubble({ m }: { m: Message }) {
  const outbound = m.direction === 'outbound'
  return (
    <div className={`flex ${outbound ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-[75%] rounded-2xl px-3.5 py-2 text-sm ${
          outbound
            ? 'bg-primary-600 text-white rounded-br-sm'
            : 'bg-white border border-gray-200 text-gray-900 rounded-bl-sm'
        }`}
      >
        <p className="whitespace-pre-wrap break-words">{m.body}</p>
        <p
          className={`mt-1 text-[10px] ${
            outbound ? 'text-primary-100' : 'text-gray-400'
          }`}
        >
          {new Date(m.created_at).toLocaleTimeString('en-IN', {
            hour: '2-digit',
            minute: '2-digit',
          })}
          {outbound && (
            <span className="ml-1 capitalize">· {m.status}</span>
          )}
          {m.template_id && (
            <span className="ml-1">· via template</span>
          )}
        </p>
      </div>
    </div>
  )
}

export default function Conversations() {
  const [rows, setRows] = useState<ConversationRow[]>([])
  const [contacts, setContacts] = useState<Contact[]>([])
  const [selected, setSelected] = useState<ConversationRow | null>(null)
  const [thread, setThread] = useState<ConversationThread | null>(null)
  const [loading, setLoading] = useState(true)
  const [threadLoading, setThreadLoading] = useState(false)
  const [sendBody, setSendBody] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')
  const [pickerOpen, setPickerOpen] = useState(false)
  const [csatOpen, setCsatOpen] = useState(false)
  const [csatScore, setCsatScore] = useState(0)
  const [csatComment, setCsatComment] = useState('')

  const [members, setMembers] = useState<TeamMember[]>([])
  const [newLabel, setNewLabel] = useState('')
  const [showLabelInput, setShowLabelInput] = useState(false)

  // Simulator state
  const [simOpen, setSimOpen] = useState(false)
  const [simContactId, setSimContactId] = useState('')
  const [simBody, setSimBody] = useState('')

  const bottomRef = useRef<HTMLDivElement>(null)

  const loadInbox = useCallback(async () => {
    setLoading(true)
    try {
      const [c, ct, tm] = await Promise.all([
        conversationsService.list(),
        contactsService.list(),
        teamService.list(),
      ])
      setRows(c)
      setContacts(ct)
      setMembers(tm)
      setError('')
    } catch {
      setError('Could not load inbox.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadInbox()
  }, [loadInbox])

  const openThread = useCallback(async (row: ConversationRow) => {
    setSelected(row)
    setThreadLoading(true)
    setThread(null)
    try {
      const t = await conversationsService.thread(row.id)
      setThread(t)
      if (row.unread_count > 0) {
        await conversationsService.markRead(row.id)
        setRows((prev) =>
          prev.map((r) => (r.id === row.id ? { ...r, unread_count: 0 } : r)),
        )
      }
    } catch {
      setError('Could not load messages.')
    } finally {
      setThreadLoading(false)
    }
  }, [])

  // Scroll to newest message on thread change.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [thread])

  const handleSend = async (e: FormEvent) => {
    e.preventDefault()
    if (!selected || !sendBody.trim()) return
    setSending(true)
    setError('')
    try {
      const result = await conversationsService.send({
        contact_id: selected.contact_id,
        channel: selected.channel,
        body: sendBody,
      })
      // Append to thread, reset input, refresh inbox row.
      setThread((prev) =>
        prev
          ? {
              ...prev,
              conversation: result.conversation,
              messages: [...prev.messages, result.message],
            }
          : prev,
      )
      setSendBody('')
      loadInbox()
    } catch (err: any) {
      const status = err.response?.status
      const detail = err.response?.data?.detail
      if (status === 409) {
        setError(detail || 'Window closed — reply requires an approved template.')
      } else {
        setError(detail || 'Could not send message.')
      }
    } finally {
      setSending(false)
    }
  }

  const parseLabels = (raw: string): string[] => {
    try { return JSON.parse(raw) } catch { return [] }
  }

  const handleAddLabel = async () => {
    if (!selected || !newLabel.trim()) return
    const current = parseLabels(selected.labels || '[]')
    if (current.includes(newLabel.trim())) { setNewLabel(''); return }
    const updated = [...current, newLabel.trim()]
    await conversationsService.update(selected.id, { labels: updated })
    setRows(prev => prev.map(r => r.id === selected.id ? { ...r, labels: JSON.stringify(updated) } : r))
    setSelected(prev => prev ? { ...prev, labels: JSON.stringify(updated) } : prev)
    setNewLabel('')
    setShowLabelInput(false)
  }

  const handleRemoveLabel = async (label: string) => {
    if (!selected) return
    const current = parseLabels(selected.labels || '[]')
    const updated = current.filter(l => l !== label)
    await conversationsService.update(selected.id, { labels: updated })
    setRows(prev => prev.map(r => r.id === selected.id ? { ...r, labels: JSON.stringify(updated) } : r))
    setSelected(prev => prev ? { ...prev, labels: JSON.stringify(updated) } : prev)
  }

  const handleAssign = async (memberId: string, memberName: string) => {
    if (!selected) return
    await conversationsService.update(selected.id, { assigned_to: memberId, assigned_name: memberName })
    setRows(prev => prev.map(r => r.id === selected.id ? { ...r, assigned_to: memberId, assigned_name: memberName } : r))
    setSelected(prev => prev ? { ...prev, assigned_to: memberId, assigned_name: memberName } : prev)
  }

  const handleUnassign = async () => {
    if (!selected) return
    await conversationsService.update(selected.id, { assigned_to: '' })
    setRows(prev => prev.map(r => r.id === selected.id ? { ...r, assigned_to: null, assigned_name: null } : r))
    setSelected(prev => prev ? { ...prev, assigned_to: null, assigned_name: null } : prev)
  }

  const insertCanned = (body: string) => {
    setSendBody(body)
    setPickerOpen(false)
  }

  const handleBodyChange = (val: string) => {
    setSendBody(val)
    // Auto-open picker when user types "/" at the very start
    if (val === '/' && !pickerOpen) setPickerOpen(true)
    if (val === '' && pickerOpen) setPickerOpen(false)
  }

  const handleSubmitCsat = async () => {
    if (!selected || csatScore < 1) return
    try {
      await csatService.submit({
        conversation_id: selected.id,
        contact_id: selected.contact_id,
        agent_user_id: selected.assigned_to || undefined,
        score: csatScore,
        comment: csatComment,
        channel: selected.channel,
      })
      setCsatOpen(false)
      setCsatScore(0)
      setCsatComment('')
    } catch {
      setError('Could not save CSAT rating.')
    }
  }

  const handleSimulateInbound = async (e: FormEvent) => {
    e.preventDefault()
    if (!simContactId || !simBody.trim()) return
    try {
      await conversationsService.simulateInbound({
        contact_id: simContactId,
        channel: 'talkex',
        body: simBody,
      })
      setSimBody('')
      setSimOpen(false)
      loadInbox()
    } catch {
      setError('Could not simulate inbound message.')
    }
  }

  return (
    <div className="h-[calc(100vh-4rem)] flex">
      {/* Left: inbox list */}
      <aside className="w-80 border-r border-gray-200 bg-white flex flex-col">
        <div className="border-b border-gray-100 px-4 py-3 flex items-center justify-between">
          <h2 className="text-base font-semibold text-gray-900 flex items-center gap-2">
            <Inbox size={16} className="text-primary-600" />
            Inbox
          </h2>
          <button
            onClick={() => setSimOpen(true)}
            className="text-xs font-medium text-primary-600 hover:text-primary-700"
            title="Simulate an inbound reply (dev)"
          >
            + Simulate
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {loading ? (
            <p className="p-6 text-center text-sm text-gray-400">Loading…</p>
          ) : rows.length === 0 ? (
            <div className="p-6 text-center">
              <MessageSquare size={28} className="mx-auto text-gray-300 mb-2" />
              <p className="text-sm text-gray-500">No conversations yet.</p>
              <p className="text-xs text-gray-400 mt-1">
                Send a message from Contacts or simulate an inbound reply above.
              </p>
            </div>
          ) : (
            rows.map((r) => {
              const isActive = selected?.id === r.id
              return (
                <button
                  key={r.id}
                  onClick={() => openThread(r)}
                  className={`w-full text-left px-4 py-3 border-b border-gray-100 transition-colors ${
                    isActive ? 'bg-primary-50' : 'hover:bg-gray-50'
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary-100 text-primary-700 text-xs font-semibold">
                      {initials(r.contact_name, r.contact_phone_number)}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-2">
                        <p className="font-medium text-gray-900 truncate">
                          {r.contact_name || r.contact_phone_number}
                        </p>
                        <span className="text-[10px] text-gray-400 shrink-0">
                          {relativeTime(r.last_message_at)}
                        </span>
                      </div>
                      <p className="text-xs text-gray-500 truncate">
                        {r.contact_phone_number} · {r.channel}
                      </p>
                      <div className="flex items-center gap-1 mt-1 flex-wrap">
                        {r.unread_count > 0 && (
                          <span className="inline-block rounded-full bg-primary-600 px-1.5 py-0.5 text-[10px] font-semibold text-white">
                            {r.unread_count}
                          </span>
                        )}
                        {parseLabels(r.labels || '[]').map(l => (
                          <span key={l} className="inline-block rounded-full bg-amber-100 text-amber-700 px-1.5 py-0.5 text-[10px] font-medium">{l}</span>
                        ))}
                        {r.assigned_name && (
                          <span className="inline-flex items-center gap-0.5 rounded-full bg-blue-50 text-blue-600 px-1.5 py-0.5 text-[10px] font-medium">
                            <UserCheck size={8} />{r.assigned_name}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                </button>
              )
            })
          )}
        </div>
      </aside>

      {/* Right: thread */}
      <main className="flex-1 flex flex-col bg-gray-50">
        {!selected ? (
          <div className="flex-1 flex items-center justify-center text-center p-8">
            <div>
              <MessageSquare size={40} className="mx-auto text-gray-300 mb-3" />
              <p className="text-sm text-gray-500">
                Select a conversation to view messages.
              </p>
            </div>
          </div>
        ) : (
          <>
            {/* Thread header */}
            <div className="border-b border-gray-200 bg-white px-6 py-3 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary-100 text-primary-700 text-xs font-semibold">
                  {initials(selected.contact_name, selected.contact_phone_number)}
                </div>
                <div>
                  <p className="font-semibold text-gray-900">
                    {selected.contact_name || selected.contact_phone_number}
                  </p>
                  <p className="text-xs text-gray-500 font-mono">
                    {selected.contact_phone_number}
                  </p>
                </div>
              </div>
              {thread &&
                (thread.window_open ? (
                  <span className="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2 py-0.5 text-xs font-semibold text-blue-700">
                    <Clock size={12} /> Window open (24h)
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-semibold text-gray-500">
                    <Lock size={12} /> Window closed
                  </span>
                ))}
            </div>

            {/* Labels & Assignment bar */}
            <div className="border-b border-gray-200 bg-white px-6 py-2 flex items-center gap-3 flex-wrap">
              <div className="flex items-center gap-1.5 flex-wrap">
                <Tag size={12} className="text-gray-400" />
                {parseLabels(selected.labels || '[]').map(l => (
                  <span key={l} className="inline-flex items-center gap-1 rounded-full bg-amber-100 text-amber-700 px-2 py-0.5 text-xs font-medium">
                    {l}
                    <button onClick={() => handleRemoveLabel(l)} className="hover:text-amber-900"><X size={10} /></button>
                  </span>
                ))}
                {showLabelInput ? (
                  <span className="inline-flex items-center gap-1">
                    <input
                      autoFocus
                      value={newLabel}
                      onChange={e => setNewLabel(e.target.value)}
                      onKeyDown={e => { if (e.key === 'Enter') handleAddLabel(); if (e.key === 'Escape') setShowLabelInput(false) }}
                      placeholder="Label…"
                      className="w-20 rounded border border-gray-300 px-1.5 py-0.5 text-xs outline-none focus:border-primary-500"
                    />
                    <button onClick={handleAddLabel} className="text-xs text-primary-600 font-medium">Add</button>
                  </span>
                ) : (
                  <button onClick={() => setShowLabelInput(true)} className="inline-flex items-center gap-0.5 text-xs text-gray-400 hover:text-primary-600">
                    <Plus size={10} /> Label
                  </button>
                )}
              </div>
              <div className="ml-auto flex items-center gap-1.5">
                <UserCheck size={12} className="text-gray-400" />
                <select
                  value={selected.assigned_to || ''}
                  onChange={e => {
                    const m = members.find(m => m.id === e.target.value)
                    if (m) handleAssign(m.id, m.name || m.email)
                    else if (e.target.value === '') handleUnassign()
                  }}
                  className="rounded border border-gray-300 px-2 py-0.5 text-xs outline-none bg-white cursor-pointer"
                >
                  <option value="">Unassigned</option>
                  {members.filter(m => m.status === 'active').map(m => (
                    <option key={m.id} value={m.id}>{m.name || m.email}</option>
                  ))}
                </select>
              </div>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-6 space-y-3">
              {threadLoading ? (
                <p className="text-center text-sm text-gray-400">Loading messages…</p>
              ) : thread?.messages.length === 0 ? (
                <p className="text-center text-sm text-gray-400">No messages yet.</p>
              ) : (
                thread?.messages.map((m) => <MessageBubble key={m.id} m={m} />)
              )}
              <div ref={bottomRef} />
            </div>

            {/* Send box */}
            <div className="border-t border-gray-200 bg-white px-6 py-3 relative">
              {error && (
                <div className="mb-2 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
                  {error}
                </div>
              )}

              {pickerOpen && (
                <div className="px-0">
                  <CannedPicker
                    onInsert={(body) => insertCanned(body)}
                    onClose={() => setPickerOpen(false)}
                  />
                </div>
              )}

              <form onSubmit={handleSend} className="flex items-end gap-2">
                <div className="flex-1 relative">
                  <textarea
                    value={sendBody}
                    onChange={(e) => handleBodyChange(e.target.value)}
                    placeholder={
                      thread?.window_open
                        ? 'Type a reply… (type "/" for canned replies)'
                        : 'Window closed — free-form send is disabled. Use a template.'
                    }
                    disabled={!thread?.window_open}
                    rows={2}
                    className="w-full resize-none rounded-lg border border-gray-300 px-3 py-2 pr-10 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none disabled:bg-gray-50 disabled:text-gray-400"
                  />
                  <button
                    type="button"
                    onClick={() => setPickerOpen(true)}
                    disabled={!thread?.window_open}
                    className="absolute right-2 bottom-2 p-1 rounded text-gray-400 hover:text-primary-600 hover:bg-primary-50 disabled:opacity-40"
                    title="Insert canned reply"
                  >
                    <Zap size={14} />
                  </button>
                </div>
                <button
                  type="button"
                  onClick={() => setCsatOpen(true)}
                  className="flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                  title="Request CSAT rating"
                >
                  <Star size={14} />
                </button>
                <button
                  type="submit"
                  disabled={sending || !sendBody.trim() || !thread?.window_open}
                  className="flex items-center gap-1.5 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                >
                  <Send size={14} />
                  Send
                </button>
              </form>
            </div>
          </>
        )}
      </main>

      {/* CSAT rating modal */}
      {csatOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/40 backdrop-blur-sm" onClick={() => setCsatOpen(false)} />
          <div className="relative w-full max-w-md rounded-xl bg-white shadow-xl">
            <div className="border-b border-gray-100 px-5 py-4">
              <h3 className="text-base font-semibold text-gray-900 flex items-center gap-2">
                <Star size={16} className="text-primary-600" />
                Rate this conversation
              </h3>
              <p className="mt-1 text-xs text-gray-500">
                Log a CSAT score after resolving. Feeds the Customer Satisfaction dashboard.
              </p>
            </div>
            <div className="px-5 py-4 space-y-4">
              <div className="flex items-center justify-center gap-2">
                {[1, 2, 3, 4, 5].map(n => (
                  <button
                    key={n}
                    onClick={() => setCsatScore(n)}
                    className={`text-3xl transition-transform ${
                      csatScore === n ? 'scale-125' : 'opacity-40 hover:opacity-80'
                    }`}
                    title={['Very unhappy', 'Unhappy', 'Neutral', 'Happy', 'Very happy'][n - 1]}
                  >
                    {['😞', '😕', '😐', '🙂', '😄'][n - 1]}
                  </button>
                ))}
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">Comment (optional)</label>
                <textarea
                  rows={3}
                  value={csatComment}
                  onChange={e => setCsatComment(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500"
                  placeholder="What could we do better?"
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setCsatOpen(false)}
                  className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSubmitCsat}
                  disabled={csatScore < 1}
                  className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-40"
                >
                  Save rating
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Simulate inbound modal */}
      {simOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
            onClick={() => setSimOpen(false)}
          />
          <div className="relative w-full max-w-md rounded-xl bg-white shadow-xl">
            <div className="border-b border-gray-100 px-5 py-4">
              <h3 className="text-base font-semibold text-gray-900 flex items-center gap-2">
                <MessageCircleReply size={16} className="text-primary-600" />
                Simulate Inbound Reply
              </h3>
              <p className="mt-1 text-xs text-gray-500">
                Dev-only helper — pretend a contact replied so you can test the 24-hour window.
              </p>
            </div>
            <form onSubmit={handleSimulateInbound} className="px-5 py-4 space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">
                  <UserIcon size={12} className="inline mr-1" />
                  Reply from contact
                </label>
                <select
                  required
                  value={simContactId}
                  onChange={(e) => setSimContactId(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none bg-white cursor-pointer"
                >
                  <option value="">Select contact…</option>
                  {contacts.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name || c.phone_number} ({c.phone_number})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1.5">
                  Message body
                </label>
                <textarea
                  required
                  rows={3}
                  value={simBody}
                  onChange={(e) => setSimBody(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                  placeholder="Hi, tell me more about your product…"
                />
              </div>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setSimOpen(false)}
                  className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"
                >
                  Send inbound
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
