import { useEffect, useState, useCallback } from 'react'
import {
  CreditCard,
  Check,
  Sparkles,
  Download,
  Receipt,
  Rocket,
} from 'lucide-react'
import { billingService } from '../services/billing'
import type { Plan, PlanID, Subscription, Invoice } from '../types/billing'

function formatINR(n: number): string {
  return n.toLocaleString('en-IN', { maximumFractionDigits: 2 })
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-IN', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

interface PlanCardProps {
  plan: Plan
  currentPlanID: PlanID
  onSelect: (id: PlanID) => void
  loading: boolean
}

function PlanCard({ plan, currentPlanID, onSelect, loading }: PlanCardProps) {
  const isCurrent = plan.id === currentPlanID
  return (
    <div
      className={`relative rounded-xl border p-5 flex flex-col ${
        plan.recommended
          ? 'border-primary-500 shadow-md bg-white'
          : 'border-gray-200 bg-white'
      }`}
    >
      {plan.recommended && (
        <span className="absolute -top-2.5 left-1/2 -translate-x-1/2 rounded-full bg-primary-600 px-3 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-white flex items-center gap-1">
          <Sparkles size={10} /> Recommended
        </span>
      )}
      <h3 className="text-lg font-bold text-gray-900">{plan.name}</h3>
      <div className="mt-2 flex items-baseline gap-1">
        <span className="text-3xl font-bold text-gray-900">
          {plan.price_inr_per_month === 0 ? 'Free' : `₹${formatINR(plan.price_inr_per_month)}`}
        </span>
        {plan.price_inr_per_month > 0 && (
          <span className="text-sm text-gray-500">/month</span>
        )}
      </div>
      <p className="mt-1 text-xs text-gray-500">
        {plan.included_messages.toLocaleString('en-IN')} messages included · ₹{plan.overage_per_msg_inr}/msg after
      </p>

      <ul className="mt-4 space-y-2 flex-1">
        {plan.features.map((f) => (
          <li key={f} className="flex items-start gap-2 text-sm text-gray-700">
            <Check size={14} className="text-green-600 mt-0.5 shrink-0" />
            <span>{f}</span>
          </li>
        ))}
      </ul>

      <button
        onClick={() => onSelect(plan.id)}
        disabled={loading || isCurrent}
        className={`mt-5 w-full rounded-lg py-2 text-sm font-semibold transition-colors ${
          isCurrent
            ? 'bg-gray-100 text-gray-500 cursor-default'
            : plan.recommended
            ? 'bg-primary-600 text-white hover:bg-primary-700'
            : 'border border-gray-300 text-gray-700 hover:bg-gray-50'
        }`}
      >
        {isCurrent ? 'Current Plan' : loading ? 'Updating…' : `Switch to ${plan.name}`}
      </button>
    </div>
  )
}

export default function Billing() {
  const [plans, setPlans] = useState<Plan[]>([])
  const [currentPlan, setCurrentPlan] = useState<Plan | null>(null)
  const [subscription, setSubscription] = useState<Subscription | null>(null)
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [loading, setLoading] = useState(true)
  const [switching, setSwitching] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [pl, sub, inv] = await Promise.all([
        billingService.listPlans(),
        billingService.getSubscription(),
        billingService.listInvoices(),
      ])
      setPlans(pl)
      setSubscription(sub.subscription)
      setCurrentPlan(sub.plan)
      setInvoices(inv)
      setError('')
    } catch {
      setError('Could not load billing details.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const handleSelect = async (planID: PlanID) => {
    setSwitching(true)
    setError('')
    setSuccess('')
    try {
      const { subscription, invoice } = await billingService.changePlan(planID)
      setSubscription(subscription)
      const p = plans.find((x) => x.id === planID) ?? null
      setCurrentPlan(p)
      setInvoices((prev) => [invoice, ...prev])
      setSuccess(`Switched to ${p?.name} plan.`)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Could not change plan.')
    } finally {
      setSwitching(false)
    }
  }

  const usagePct =
    currentPlan && subscription
      ? Math.min(100, (subscription.messages_used / currentPlan.included_messages) * 100)
      : 0

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <CreditCard size={24} className="text-primary-600" />
          Billing
        </h1>
        <p className="text-sm text-gray-500 mt-1">
          Manage your subscription, review usage, and download invoices.
        </p>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-lg bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-700">
          {success}
        </div>
      )}

      {loading && !currentPlan ? (
        <div className="p-10 text-center text-sm text-gray-400">Loading billing…</div>
      ) : (
        <>
          {/* Current plan card */}
          {currentPlan && subscription && (
            <div className="rounded-xl border border-gray-200 bg-gradient-to-br from-primary-50 to-white p-5">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs font-medium uppercase tracking-wider text-primary-600">
                    Current Plan
                  </p>
                  <h2 className="mt-1 text-2xl font-bold text-gray-900 flex items-center gap-2">
                    {currentPlan.name}
                    <span className="text-sm font-normal text-gray-500">
                      {currentPlan.price_inr_per_month === 0
                        ? 'Free'
                        : `₹${formatINR(currentPlan.price_inr_per_month)}/mo`}
                    </span>
                  </h2>
                  <p className="mt-1 text-xs text-gray-500">
                    Renewing {formatDate(subscription.period_start)} · Status:{' '}
                    <span className="capitalize font-medium text-gray-700">{subscription.status}</span>
                  </p>
                </div>
                <Rocket size={28} className="text-primary-500" />
              </div>

              <div className="mt-4">
                <div className="flex items-baseline justify-between text-xs mb-1.5">
                  <span className="text-gray-600">
                    {subscription.messages_used.toLocaleString('en-IN')} /{' '}
                    {currentPlan.included_messages.toLocaleString('en-IN')} messages used
                  </span>
                  <span className="font-medium text-gray-700">{usagePct.toFixed(0)}%</span>
                </div>
                <div className="h-2 rounded-full bg-white/70 overflow-hidden">
                  <div
                    className="h-full bg-primary-500 rounded-full transition-all"
                    style={{ width: `${usagePct}%` }}
                  />
                </div>
              </div>
            </div>
          )}

          {/* Plan grid */}
          <div>
            <h2 className="text-sm font-semibold text-gray-900 mb-3">Available plans</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {plans.map((p) => (
                <PlanCard
                  key={p.id}
                  plan={p}
                  currentPlanID={currentPlan?.id ?? 'starter'}
                  onSelect={handleSelect}
                  loading={switching}
                />
              ))}
            </div>
          </div>

          {/* Invoices */}
          <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
            <div className="border-b border-gray-100 px-5 py-3 flex items-center gap-2">
              <Receipt size={16} className="text-primary-600" />
              <h2 className="text-sm font-semibold text-gray-900">Invoices</h2>
            </div>
            {invoices.length === 0 ? (
              <div className="p-8 text-center text-sm text-gray-400">
                No invoices yet — invoices appear here after your first paid plan change.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      <th className="px-4 py-2.5">Date</th>
                      <th className="px-4 py-2.5">Plan</th>
                      <th className="px-4 py-2.5">Period</th>
                      <th className="px-4 py-2.5">Amount</th>
                      <th className="px-4 py-2.5">Status</th>
                      <th className="px-4 py-2.5"></th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {invoices.map((inv) => (
                      <tr key={inv.id} className="hover:bg-gray-50">
                        <td className="px-4 py-2.5 text-gray-500 text-xs">
                          {formatDate(inv.created_at)}
                        </td>
                        <td className="px-4 py-2.5 font-medium text-gray-900 capitalize">
                          {inv.plan}
                        </td>
                        <td className="px-4 py-2.5 text-gray-500 text-xs">
                          {formatDate(inv.period_start)} → {formatDate(inv.period_end)}
                        </td>
                        <td className="px-4 py-2.5 font-mono text-gray-900">
                          ₹{formatINR(inv.amount_inr)}
                        </td>
                        <td className="px-4 py-2.5">
                          <span
                            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold ${
                              inv.status === 'paid'
                                ? 'bg-green-50 text-green-700'
                                : inv.status === 'pending'
                                ? 'bg-amber-50 text-amber-700'
                                : 'bg-red-50 text-red-700'
                            }`}
                          >
                            {inv.status}
                          </span>
                        </td>
                        <td className="px-4 py-2.5 text-right">
                          <button
                            className="text-primary-600 hover:text-primary-700"
                            title="Download PDF (not yet available)"
                            disabled
                          >
                            <Download size={14} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
