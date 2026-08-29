import { useEffect, useState } from 'react'
import { Zap, X } from 'lucide-react'
import { cannedService, type CannedResponse } from '../services/canned'

interface Props {
  onInsert: (body: string, id: string) => void
  onClose: () => void
}

/**
 * CannedPicker — dropdown of canned responses shown when the user
 * types "/" in the reply box. Filters by shortcut or title.
 */
export default function CannedPicker({ onInsert, onClose }: Props) {
  const [items, setItems] = useState<CannedResponse[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [activeIdx, setActiveIdx] = useState(0)

  useEffect(() => {
    cannedService.list()
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }, [])

  const filtered = query
    ? items.filter(r =>
        r.shortcut.toLowerCase().includes(query.toLowerCase()) ||
        r.title.toLowerCase().includes(query.toLowerCase())
      )
    : items

  const handleInsert = (r: CannedResponse) => {
    onInsert(r.body, r.id)
    cannedService.bumpUsage(r.id).catch(() => {})
  }

  const onKey = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') { e.preventDefault(); setActiveIdx(i => Math.min(i + 1, filtered.length - 1)) }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setActiveIdx(i => Math.max(i - 1, 0)) }
    else if (e.key === 'Enter' && filtered[activeIdx]) { e.preventDefault(); handleInsert(filtered[activeIdx]) }
    else if (e.key === 'Escape') { onClose() }
  }

  return (
    <div className="absolute bottom-full left-0 right-0 mb-2 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xl max-h-80 flex flex-col overflow-hidden">
      <div className="flex items-center gap-2 px-3 py-2 border-b border-gray-100 dark:border-gray-700">
        <Zap size={14} className="text-primary-600 shrink-0" />
        <input
          autoFocus
          value={query}
          onChange={e => setQuery(e.target.value)}
          onKeyDown={onKey}
          placeholder="Search canned replies…"
          className="flex-1 text-sm bg-transparent text-gray-900 dark:text-gray-100 placeholder-gray-400 outline-none"
        />
        <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={14} /></button>
      </div>
      <div className="overflow-y-auto flex-1">
        {loading ? (
          <p className="text-xs text-gray-400 text-center py-4">Loading…</p>
        ) : filtered.length === 0 ? (
          <p className="text-xs text-gray-400 text-center py-4">
            No canned replies. <a href="/canned-responses" className="text-primary-600 hover:underline">Create one →</a>
          </p>
        ) : (
          filtered.map((r, i) => (
            <button
              key={r.id}
              onClick={() => handleInsert(r)}
              onMouseEnter={() => setActiveIdx(i)}
              className={`w-full text-left px-3 py-2 text-sm transition-colors ${
                i === activeIdx
                  ? 'bg-primary-50 dark:bg-primary-900/30'
                  : 'hover:bg-gray-50 dark:hover:bg-gray-700/50'
              }`}
            >
              <div className="flex items-center gap-2 mb-0.5">
                <code className="text-[10px] font-mono bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-1 py-0.5 rounded">{r.shortcut}</code>
                <span className="text-xs font-medium text-gray-900 dark:text-gray-100">{r.title}</span>
                <span className="text-[10px] text-gray-400 ml-auto">{r.category}</span>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-400 line-clamp-1">{r.body}</p>
            </button>
          ))
        )}
      </div>
      <div className="px-3 py-1.5 border-t border-gray-100 dark:border-gray-700 text-[10px] text-gray-400 flex items-center gap-3">
        <span>↑↓ Navigate</span>
        <span>↵ Insert</span>
        <span>ESC Close</span>
      </div>
    </div>
  )
}
