import type { LucideIcon } from 'lucide-react'

interface PlaceholderPageProps {
  title: string
  description: string
  icon: LucideIcon
  phase?: string
}

export default function PlaceholderPage({
  title,
  description,
  icon: Icon,
  phase,
}: PlaceholderPageProps) {
  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
        <p className="text-gray-500 mt-1">{description}</p>
      </div>

      <div className="rounded-xl border bg-white p-12 shadow-sm text-center">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary-600">
          <Icon size={32} />
        </div>
        <h3 className="text-lg font-semibold text-gray-900 mb-2">
          {title} Module
        </h3>
        <p className="text-gray-500 text-sm max-w-md mx-auto">
          This module is under development. Full functionality will be available soon.
        </p>
        {phase && (
          <span className="mt-4 inline-flex items-center rounded-full bg-primary-50 px-3 py-1 text-xs font-medium text-primary-700">
            {phase}
          </span>
        )}
      </div>
    </div>
  )
}
