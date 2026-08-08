import { useEffect, useState, useCallback, type FormEvent } from 'react'
import {
  Wallet as WalletIcon,
  Plus,
  ArrowUpRight,
  ArrowDownRight,
  RefreshCw,
} from 'lucide-react'
import { walletService } from '../services/wallet'
import type { Wallet, WalletTransaction } from '../types/wallet'
import Modal from '../components/Modal'

const QUICK_AMOUNTS = [500, 1000, 2000, 5000]

function formatMoney(amount: number, currency: string) {
  const symbol = currency === 'INR' ? '₹' : currency + ' '
  return `${symbol}${amount.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString('en-IN', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export default function WalletPage() {
  const [wallet, setWallet] = useState<Wallet | null>(null)
  const [transactions, setTransactions] = useState<WalletTransaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [modalOpen, setModalOpen] = useState(false)
  const [amount, setAmount] = useState('')
  const [reference, setReference] = useState('')
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [w, txns] = await Promise.all([
        walletService.get(),
        walletService.listTransactions(),
      ])
      setWallet(w)
      setTransactions(txns)
      setError('')
    } catch {
      setError('Could not load wallet. Is the backend running?')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const openAddFunds = () => {
    setAmount('')
    setReference('')
    setFormError('')
    setModalOpen(true)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const value = parseFloat(amount)
    if (!value || value <= 0) {
      setFormError('Enter a valid amount greater than 0')
      return
    }
    setFormError('')
    setSaving(true)
    try {
      // A fresh idempotency key per submit — a network retry of THIS same
      // request (handled by axios/browser) reuses it naturally; a second,
      // deliberate click generates a new key and is treated as a new top-up.
      const idempotencyKey = crypto.randomUUID()
      const txn = await walletService.createTransaction({
        type: 'credit',
        amount: value,
        reference: reference || null,
        idempotency_key: idempotencyKey,
      })
      setTransactions((prev) => [txn, ...prev])
      setWallet((prev) => (prev ? { ...prev, balance: txn.balance_after } : prev))
      setModalOpen(false)
    } catch (err: any) {
      setFormError(err.response?.data?.detail || 'Could not add funds')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <WalletIcon size={24} className="text-primary-600" />
            Wallet
          </h1>
          <p className="text-sm text-gray-500 mt-1">
            Prepaid balance for sending messages — top up anytime.
          </p>
        </div>
        <button
          onClick={load}
          className="flex items-center gap-2 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
        >
          <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
          Refresh
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Balance card */}
      <div className="rounded-2xl bg-gradient-to-br from-primary-600 to-primary-800 p-6 text-white">
        <p className="text-sm text-primary-100">Available Balance</p>
        <p className="text-4xl font-bold mt-1">
          {wallet ? formatMoney(wallet.balance, wallet.currency) : '—'}
        </p>
        <button
          onClick={openAddFunds}
          className="mt-4 flex items-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-semibold text-primary-700 hover:bg-primary-50 transition-colors"
        >
          <Plus size={16} />
          Add Funds
        </button>
      </div>

      {/* Transaction history */}
      <div>
        <h2 className="text-sm font-semibold text-gray-700 mb-3">Transaction History</h2>
        <div className="rounded-xl border border-gray-200 bg-white overflow-hidden">
          {loading ? (
            <div className="p-10 text-center text-sm text-gray-400">Loading transactions…</div>
          ) : transactions.length === 0 ? (
            <div className="p-10 text-center">
              <WalletIcon size={32} className="mx-auto text-gray-300 mb-2" />
              <p className="text-sm text-gray-500">No transactions yet.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-200 bg-gray-50 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    <th className="px-4 py-2.5">Type</th>
                    <th className="px-4 py-2.5">Amount</th>
                    <th className="px-4 py-2.5">Balance After</th>
                    <th className="px-4 py-2.5">Reference</th>
                    <th className="px-4 py-2.5">Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {transactions.map((t) => (
                    <tr key={t.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-4 py-3">
                        <span
                          className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
                            t.type === 'credit'
                              ? 'bg-green-50 text-green-700'
                              : 'bg-red-50 text-red-700'
                          }`}
                        >
                          {t.type === 'credit' ? (
                            <ArrowDownRight size={12} />
                          ) : (
                            <ArrowUpRight size={12} />
                          )}
                          {t.type === 'credit' ? 'Credit' : 'Debit'}
                        </span>
                      </td>
                      <td
                        className={`px-4 py-3 font-semibold ${
                          t.type === 'credit' ? 'text-green-700' : 'text-red-700'
                        }`}
                      >
                        {t.type === 'credit' ? '+' : '-'}
                        {formatMoney(t.amount, wallet?.currency ?? 'INR')}
                      </td>
                      <td className="px-4 py-3 text-gray-700">
                        {formatMoney(t.balance_after, wallet?.currency ?? 'INR')}
                      </td>
                      <td className="px-4 py-3 text-gray-500">{t.reference || '—'}</td>
                      <td className="px-4 py-3 text-gray-400 whitespace-nowrap">
                        {formatTime(t.created_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Add Funds modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title="Add Funds">
        <form onSubmit={handleSubmit} className="space-y-4">
          {formError && (
            <div className="rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {formError}
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">Amount (₹) *</label>
            <input
              type="number"
              required
              min="1"
              step="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="1000"
            />
            <div className="flex gap-2 mt-2">
              {QUICK_AMOUNTS.map((a) => (
                <button
                  key={a}
                  type="button"
                  onClick={() => setAmount(String(a))}
                  className="rounded-lg border border-gray-200 px-2.5 py-1 text-xs font-medium text-gray-600 hover:border-primary-400 hover:text-primary-600 transition-colors"
                >
                  ₹{a}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1.5">
              Reference <span className="text-gray-400 font-normal">(optional)</span>
            </label>
            <input
              type="text"
              value={reference}
              onChange={(e) => setReference(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
              placeholder="e.g. UPI ref, invoice #"
            />
          </div>

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={() => setModalOpen(false)}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="flex-1 rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
            >
              {saving ? 'Processing...' : 'Add Funds'}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
