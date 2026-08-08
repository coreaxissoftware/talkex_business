import { ShieldCheck, ShieldAlert, ShieldX } from 'lucide-react'
import type { QualityStatus } from '../types/auth'

const CONFIG: Record<QualityStatus, { label: string; icon: typeof ShieldCheck; className: string }> = {
  green: {
    label: 'Quality: Green',
    icon: ShieldCheck,
    className: 'bg-green-50 text-green-700 border-green-200',
  },
  yellow: {
    label: 'Quality: Yellow',
    icon: ShieldAlert,
    className: 'bg-amber-50 text-amber-700 border-amber-200',
  },
  red: {
    label: 'Quality: Red',
    icon: ShieldX,
    className: 'bg-red-50 text-red-700 border-red-200',
  },
}

interface QualityBadgeProps {
  status: QualityStatus
}

export default function QualityBadge({ status }: QualityBadgeProps) {
  const { label, icon: Icon, className } = CONFIG[status]
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-semibold ${className}`}
      title="Messaging quality tier — determines how many business-initiated conversations are allowed per 24h."
    >
      <Icon size={13} />
      {label}
    </span>
  )
}
