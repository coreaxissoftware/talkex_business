import { useEffect, useState } from 'react'
import { Image, Upload, Trash2, Check, X, File, FileText, Film, Music } from 'lucide-react'
import { mediaService } from '../services/media'
import type { MediaItem } from '../types/media'

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function mimeIcon(mime: string) {
  if (mime.startsWith('image/')) return <Image size={20} className="text-blue-500" />
  if (mime.startsWith('video/')) return <Film size={20} className="text-purple-500" />
  if (mime.startsWith('audio/')) return <Music size={20} className="text-green-500" />
  if (mime.includes('pdf')) return <FileText size={20} className="text-red-500" />
  return <File size={20} className="text-gray-500" />
}

export default function MediaLibrary() {
  const [items, setItems] = useState<MediaItem[]>([])
  const [loading, setLoading] = useState(true)
  const [uploading, setUploading] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      setItems(await mediaService.list())
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (!files || files.length === 0) return
    setUploading(true)
    try {
      for (const file of Array.from(files)) {
        await mediaService.upload(file)
      }
      await load()
    } catch {
      // ignore
    } finally {
      setUploading(false)
      e.target.value = ''
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await mediaService.remove(id)
      setDeletingId(null)
      await load()
    } catch {
      // ignore
    }
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Image size={24} className="text-primary-600" />
            Media Library
          </h1>
          <p className="text-sm text-gray-500 mt-1">Upload and manage images, documents, and other files.</p>
        </div>
        <label className="flex items-center gap-1.5 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 transition-colors cursor-pointer">
          <Upload size={16} />
          {uploading ? 'Uploading...' : 'Upload File'}
          <input type="file" multiple className="hidden" onChange={handleUpload} disabled={uploading} />
        </label>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-400">Loading...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <Image size={48} className="mx-auto mb-3 opacity-30" />
          <p className="font-medium">No files uploaded</p>
          <p className="text-sm mt-1">Upload images, PDFs, or documents for use in templates and messages.</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {items.map(item => (
            <div key={item.id} className="rounded-xl border border-gray-200 bg-white overflow-hidden group">
              {/* Preview area */}
              <div className="h-36 bg-gray-50 flex items-center justify-center overflow-hidden">
                {item.mime_type.startsWith('image/') ? (
                  <img
                    src={item.url}
                    alt={item.original_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="flex flex-col items-center gap-2 text-gray-400">
                    {mimeIcon(item.mime_type)}
                    <span className="text-xs uppercase tracking-wider">
                      {item.mime_type.split('/')[1] || 'file'}
                    </span>
                  </div>
                )}
              </div>

              {/* Info */}
              <div className="p-3">
                <p className="text-sm font-medium text-gray-900 truncate" title={item.original_name}>
                  {item.original_name}
                </p>
                <div className="flex items-center justify-between mt-1">
                  <span className="text-xs text-gray-400">{formatSize(item.size)}</span>
                  {deletingId === item.id ? (
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => handleDelete(item.id)}
                        className="p-1 rounded text-red-600 hover:bg-red-50 transition-colors"
                      >
                        <Check size={14} />
                      </button>
                      <button
                        onClick={() => setDeletingId(null)}
                        className="p-1 rounded text-gray-400 hover:bg-gray-100 transition-colors"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ) : (
                    <button
                      onClick={() => setDeletingId(item.id)}
                      className="p-1 rounded text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors opacity-0 group-hover:opacity-100"
                    >
                      <Trash2 size={14} />
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <p className="text-xs text-gray-400 text-center">{items.length} file{items.length !== 1 ? 's' : ''} • Max 10 MB per upload</p>
    </div>
  )
}
