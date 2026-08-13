import { useEffect, useState, type FormEvent } from 'react'
import { Tag, Edit3, Trash2, RefreshCw } from 'lucide-react'
import { tagsService, type TagCount } from '../services/tags'

export default function Tags() {
  const [tags, setTags] = useState<TagCount[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [renameTag, setRenameTag] = useState<string | null>(null)
  const [newName, setNewName] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      setTags(await tagsService.list())
      setError('')
    } catch {
      setError('Could not load tags.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleRename = async (e: FormEvent) => {
    e.preventDefault()
    if (!renameTag || !newName.trim()) return
    try {
      await tagsService.rename(renameTag, newName.trim())
      setRenameTag(null)
      setNewName('')
      load()
    } catch {
      setError('Could not rename tag.')
    }
  }

  const handleDelete = async (name: string) => {
    if (!confirm(`Remove tag "${name}" from all contacts?`)) return
    try {
      await tagsService.remove(name)
      load()
    } catch {
      setError('Could not delete tag.')
    }
  }

  return (
    <div className="p-6 max-w-3xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
          <Tag size={20} className="text-primary-600" />
          Tags Management
        </h1>
        <button onClick={load} className="text-sm text-gray-500 hover:text-primary-600 flex items-center gap-1">
          <RefreshCw size={14} /> Refresh
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">{error}</div>
      )}

      {loading ? (
        <p className="text-sm text-gray-400 text-center py-12">Loading tags…</p>
      ) : tags.length === 0 ? (
        <div className="text-center py-12">
          <Tag size={32} className="mx-auto text-gray-300 mb-3" />
          <p className="text-sm text-gray-500">No tags yet. Tags are added via Contacts.</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-200 divide-y divide-gray-100">
          {tags.map(t => (
            <div key={t.name} className="flex items-center justify-between px-5 py-3">
              {renameTag === t.name ? (
                <form onSubmit={handleRename} className="flex items-center gap-2 flex-1">
                  <input
                    autoFocus
                    value={newName}
                    onChange={e => setNewName(e.target.value)}
                    className="rounded border border-gray-300 px-2 py-1 text-sm outline-none focus:border-primary-500 w-48"
                    placeholder="New name…"
                  />
                  <button type="submit" className="text-xs font-medium text-primary-600">Save</button>
                  <button type="button" onClick={() => setRenameTag(null)} className="text-xs text-gray-400">Cancel</button>
                </form>
              ) : (
                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center gap-1 rounded-full bg-amber-100 text-amber-700 px-2.5 py-1 text-sm font-medium">
                    <Tag size={12} /> {t.name}
                  </span>
                  <span className="text-xs text-gray-400">{t.count} contact{t.count !== 1 ? 's' : ''}</span>
                </div>
              )}

              {renameTag !== t.name && (
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => { setRenameTag(t.name); setNewName(t.name) }}
                    className="p-1.5 rounded-lg text-gray-400 hover:text-primary-600 hover:bg-primary-50"
                    title="Rename"
                  >
                    <Edit3 size={14} />
                  </button>
                  <button
                    onClick={() => handleDelete(t.name)}
                    className="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50"
                    title="Delete"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
