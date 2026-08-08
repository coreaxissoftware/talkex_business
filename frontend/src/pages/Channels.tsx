import { Radio } from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import QualityBadge from '../components/QualityBadge'

export default function Channels() {
  const { user } = useAuthStore()

  return (
    <div>
      <div className="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Channels</h1>
          <p className="text-gray-500 mt-1">
            Manage your messaging channels — TalkEx Business, WhatsApp Business, and more.
          </p>
        </div>
        {user && <QualityBadge status={user.quality_status} />}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="rounded-xl border bg-white p-6 shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <div className="rounded-lg bg-blue-50 p-2.5 text-blue-600">
              <Radio size={24} />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">TalkEx Business</h3>
              <p className="text-sm text-gray-500">First-party channel</p>
            </div>
          </div>
          <span className="inline-flex items-center rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700">
            Coming Soon
          </span>
        </div>

        <div className="rounded-xl border bg-white p-6 shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <div className="rounded-lg bg-green-50 p-2.5 text-green-600">
              <MessageSquareIcon size={24} />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">WhatsApp Business</h3>
              <p className="text-sm text-gray-500">Meta Business API</p>
            </div>
          </div>
          <span className="inline-flex items-center rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700">
            Coming Soon
          </span>
        </div>
      </div>
    </div>
  )
}

function MessageSquareIcon(props: { size: number }) {
  return (
    <svg width={props.size} height={props.size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
    </svg>
  )
}
