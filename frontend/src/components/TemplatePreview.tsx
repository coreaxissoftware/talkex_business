import { useMemo } from 'react'
import { Eye, MessageSquare } from 'lucide-react'

interface TemplatePreviewProps {
  body: string
  channel?: string
  className?: string
}

/** Renders a template body with variable placeholders highlighted. */
export default function TemplatePreview({ body, channel, className = '' }: TemplatePreviewProps) {
  const rendered = useMemo(() => {
    // Highlight {{1}}, {{2}}, etc. as styled pill badges
    return body.replace(
      /\{\{(\d+)\}\}/g,
      '<span class="inline-block rounded bg-primary-100 text-primary-700 text-[10px] font-mono px-1.5 py-0.5 mx-0.5">$1</span>'
    )
  }, [body])

  return (
    <div className={`rounded-lg border border-gray-200 bg-gray-50 p-4 ${className}`}>
      <div className="flex items-center gap-2 mb-2 text-xs text-gray-500">
        <Eye size={12} />
        <span>Preview</span>
        {channel && (
          <>
            <span className="text-gray-300">·</span>
            <span className="capitalize">{channel}</span>
          </>
        )}
      </div>
      <div className="relative">
        {/* Chat bubble style */}
        <div className="bg-white rounded-lg rounded-tl-none shadow-sm border border-gray-100 p-3 max-w-[90%]">
          <div className="flex items-start gap-2">
            <div className="rounded-full bg-primary-100 p-1 mt-0.5">
              <MessageSquare size={10} className="text-primary-600" />
            </div>
            <div
              className="text-sm text-gray-800 leading-relaxed whitespace-pre-wrap"
              dangerouslySetInnerHTML={{ __html: rendered }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
