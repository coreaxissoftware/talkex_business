import { useEffect, useState } from 'react'
import { BadgeCheck, Check, Loader, AlertCircle } from 'lucide-react'
import { greenTickService, type GreenTickApplication } from '../services/greentick'

const CHECKLIST: Array<{ key: keyof GreenTickApplication; label: string; hint: string }> = [
  { key: 'notable_brand', label: 'Notable brand', hint: 'Recognisable brand with trademark on file' },
  { key: 'org_website', label: 'Organisation website', hint: 'Business website live and matches brand' },
  { key: 'meta_200_msg', label: 'Meta 200+ conversations (90d)', hint: '200+ conversations in the last 90 days' },
  { key: 'meta_tier2', label: 'Meta Tier 2 messaging limit', hint: 'Meta messaging limit tier 2 or higher' },
  { key: 'business_verified', label: 'Business Verification', hint: 'Meta Business Verification complete' },
  { key: 'trademark_refs', label: 'Trademark references', hint: 'Three third-party news mentions' },
]

const STATUS_COPY = {
  not_started: { label: 'Not started', color: 'bg-gray-100 text-gray-700' },
  in_progress: { label: 'In progress', color: 'bg-blue-50 text-blue-700' },
  submitted: { label: 'Submitted to Meta', color: 'bg-amber-50 text-amber-700' },
  approved: { label: 'Approved · Verified', color: 'bg-green-50 text-green-700' },
  rejected: { label: 'Rejected', color: 'bg-red-50 text-red-700' },
}

export default function GreenTick() {
  const [app, setApp] = useState<GreenTickApplication | null>(null)
  const [progress, setProgress] = useState(0)
  const [loading, setLoading] = useState(true)
  const [caseId, setCaseId] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const res = await greenTickService.get()
      setApp(res.application)
      setProgress(res.progress)
      setCaseId(res.application.meta_case_id || '')
      setError('')
    } catch {
      setError('Could not load application.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const toggle = async (key: keyof GreenTickApplication) => {
    if (!app || app.status === 'submitted' || app.status === 'approved') return
    try {
      const res = await greenTickService.update({ [key]: !app[key] } as any)
      setApp(res.application)
      setProgress(res.progress)
    } catch {
      setError('Could not update.')
    }
  }

  const submit = async () => {
    if (progress < 1) return
    setSubmitting(true)
    try {
      const res = await greenTickService.submit(caseId)
      setApp(res)
      await load()
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not submit.')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return <div className="p-6 text-center text-gray-500">Loading…</div>
  }
  if (!app) return null

  const status = STATUS_COPY[app.status]
  const locked = app.status === 'submitted' || app.status === 'approved'

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <BadgeCheck size={26} className="text-primary-600" /> Green Tick Verification
        </h1>
        <p className="text-sm text-gray-500 mt-1">
          Track your progress toward the WhatsApp Official Business Account badge.
        </p>
      </div>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700 flex items-center gap-2">
          <AlertCircle size={16} /> {error}
        </div>
      )}

      {/* Status card */}
      <div className="rounded-xl border border-gray-200 bg-white p-5 mb-6">
        <div className="flex items-center justify-between mb-3">
          <div>
            <p className="text-xs uppercase tracking-wider text-gray-500">Status</p>
            <span className={`inline-block mt-1 px-2.5 py-1 rounded-full text-xs font-medium ${status.color}`}>
              {status.label}
            </span>
          </div>
          <div className="text-right">
            <p className="text-xs uppercase tracking-wider text-gray-500">Progress</p>
            <p className="text-2xl font-bold text-gray-900">{Math.round(progress * 100)}%</p>
          </div>
        </div>
        <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
          <div
            className="h-full bg-primary-600 transition-all"
            style={{ width: `${progress * 100}%` }}
          />
        </div>
      </div>

      {/* Checklist */}
      <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
        <div className="px-5 py-3 border-b border-gray-200 bg-gray-50">
          <h2 className="font-semibold text-gray-900 text-sm">Meta prerequisite checklist</h2>
        </div>
        <ul className="divide-y divide-gray-100">
          {CHECKLIST.map((item) => {
            const done = app[item.key] === true
            return (
              <li key={item.key as string} className="px-5 py-3 flex items-start gap-3">
                <button
                  disabled={locked}
                  onClick={() => toggle(item.key)}
                  className={`mt-0.5 shrink-0 w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                    done
                      ? 'bg-primary-600 border-primary-600 text-white'
                      : 'border-gray-300 hover:border-primary-500'
                  } ${locked ? 'opacity-60 cursor-not-allowed' : 'cursor-pointer'}`}
                >
                  {done && <Check size={13} strokeWidth={3} />}
                </button>
                <div>
                  <p className={`text-sm font-medium ${done ? 'text-gray-900' : 'text-gray-700'}`}>
                    {item.label}
                  </p>
                  <p className="text-xs text-gray-500 mt-0.5">{item.hint}</p>
                </div>
              </li>
            )
          })}
        </ul>
      </div>

      {/* Submit */}
      {app.status !== 'approved' && (
        <div className="mt-6 rounded-xl border border-gray-200 bg-white p-5">
          <h3 className="font-semibold text-gray-900 text-sm mb-3">Submit to Meta</h3>
          <input
            value={caseId}
            onChange={(e) => setCaseId(e.target.value)}
            disabled={locked}
            placeholder="Meta support case ID (optional)"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 mb-3 disabled:bg-gray-50"
          />
          <button
            onClick={submit}
            disabled={progress < 1 || submitting || locked}
            className="w-full sm:w-auto px-4 py-2 rounded-lg bg-primary-600 text-white text-sm font-semibold hover:bg-primary-700 disabled:opacity-40 flex items-center gap-2 justify-center"
          >
            {submitting ? <Loader size={14} className="animate-spin" /> : <BadgeCheck size={14} />}
            {app.status === 'submitted' ? 'Awaiting decision' : 'Mark as submitted'}
          </button>
          {progress < 1 && (
            <p className="text-xs text-gray-500 mt-2">
              Complete every checklist item before submitting.
            </p>
          )}
        </div>
      )}

      {app.status === 'rejected' && app.reject_reason && (
        <div className="mt-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm">
          <p className="font-semibold text-red-800 mb-1">Meta rejection reason</p>
          <p className="text-red-700">{app.reject_reason}</p>
        </div>
      )}
    </div>
  )
}
