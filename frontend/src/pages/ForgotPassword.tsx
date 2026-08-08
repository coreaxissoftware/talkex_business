import { useState } from 'react'
import { Link } from 'react-router'
import { Mail, ArrowLeft, CheckCircle } from 'lucide-react'
import talkexIcon from '../assets/talkex-icon.png'

export default function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      // TODO: Wire to backend /auth/forgot-password endpoint
      // For now, simulate success after brief delay
      await new Promise((resolve) => setTimeout(resolve, 1000))
      setSent(true)
    } catch (err: any) {
      setError(err.response?.data?.error || 'Something went wrong. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  if (sent) {
    return (
      <div className="text-center">
        <div className="mb-8 lg:hidden">
          <img src={talkexIcon} alt="TalkEx" className="h-12 w-12 rounded-xl mx-auto mb-3" />
          <h1 className="text-2xl font-bold text-gray-900">
            Talk<span className="text-primary-600">Ex</span> Business
          </h1>
        </div>

        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
          <CheckCircle size={24} className="text-green-600" />
        </div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Check your email</h2>
        <p className="text-gray-500 mb-2">
          We sent a password reset link to
        </p>
        <p className="text-sm font-medium text-gray-900 mb-6">{email}</p>
        <p className="text-xs text-gray-400 mb-6">
          Didn't receive the email? Check your spam folder or{' '}
          <button
            onClick={() => setSent(false)}
            className="text-primary-600 hover:underline font-medium"
          >
            try again
          </button>
        </p>

        <Link
          to="/login"
          className="inline-flex items-center gap-2 text-sm font-medium text-primary-600 hover:text-primary-700"
        >
          <ArrowLeft size={16} />
          Back to sign in
        </Link>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-8 lg:hidden text-center">
        <img src={talkexIcon} alt="TalkEx" className="h-12 w-12 rounded-xl mx-auto mb-3" />
        <h1 className="text-2xl font-bold text-gray-900">
          Talk<span className="text-primary-600">Ex</span> Business
        </h1>
      </div>

      <h2 className="text-2xl font-bold text-gray-900 mb-1">Forgot password?</h2>
      <p className="text-gray-500 mb-6">
        No worries, we'll send you reset instructions.
      </p>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="resetEmail" className="block text-sm font-medium text-gray-700 mb-1.5">
            Email
          </label>
          <input
            id="resetEmail"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
            placeholder="you@company.com"
            autoComplete="email"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
        >
          <Mail size={18} />
          {loading ? 'Sending...' : 'Send reset link'}
        </button>
      </form>

      <p className="mt-6 text-center">
        <Link
          to="/login"
          className="inline-flex items-center gap-2 text-sm font-medium text-primary-600 hover:text-primary-700"
        >
          <ArrowLeft size={16} />
          Back to sign in
        </Link>
      </p>
    </div>
  )
}
