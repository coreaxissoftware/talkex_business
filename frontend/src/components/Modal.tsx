import { type ReactNode } from 'react'
import { X } from 'lucide-react'

interface ModalProps {
  // Defaults true so callers that gate rendering with `{cond && <Modal>}`
  // don't need to pass `open` explicitly.
  open?: boolean
  onClose: () => void
  title: string
  // `wide` bumps the max-width from `md` to `2xl` for content-heavy
  // dialogs (Channels onboarding wizard, etc).
  wide?: boolean
  children: ReactNode
}

export default function Modal({ open = true, onClose, title, wide, children }: ModalProps) {
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Panel */}
      <div className={`relative w-full ${wide ? 'max-w-2xl' : 'max-w-md'} rounded-xl bg-white shadow-xl`}>
        <div className="flex items-center justify-between border-b border-gray-100 px-5 py-4">
          <h3 className="text-base font-semibold text-gray-900">{title}</h3>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors"
          >
            <X size={18} />
          </button>
        </div>
        <div className="px-5 py-4">{children}</div>
      </div>
    </div>
  )
}
