import { useState } from 'react'
import { Link, useSearchParams, useNavigate } from 'react-router'
import { Lock, ArrowLeft, CheckCircle } from 'lucide-react'
import talkexIcon from '../assets/talkex-icon.png'
import api from '../services/api'
import PasswordInput from '../components/PasswordInput'

export default function ResetPassword() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const token = searchParams.get('token') || ''

  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [loading, setLoading] = useState(false)
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (password !== confirm) {
      setError('Passwords do not match')
      return
    }
    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    setLoading(true)
    try {
      await api.post('/auth/reset-password', { token, new_password: password })
      setDone(true)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Invalid or expired reset link. Please request a new one.')
    } finally {
      setLoading(false)
    }
  }

  if (!token) {
    return (
      <div className="text-center">
        <div className="mb-8 lg:hidden">
          <img src={talkexIcon} alt="TalkEx" className="h-12 w-12 rounded-xl mx-auto mb-3" />
          <h1 className="text-2xl font-bold text-gray-900">
            Talk<span className="text-primary-600">Ex</span> Business
          </h1>
        </div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Invalid link</h2>
        <p className="text-gray-500 mb-6">This reset link is invalid or has expired.</p>
        <Link
          to="/forgot-password"
          className="inline-flex items-center gap-2 text-sm font-medium text-primary-600 hover:text-primary-700"
        >
          Request a new link
        </Link>
      </div>
    )
  }

  if (done) {
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
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Password reset!</h2>
        <p className="text-gray-500 mb-6">Your password has been successfully updated.</p>
        <button
          onClick={() => navigate('/login')}
          className="rounded-lg bg-primary-600 px-6 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 transition-colors"
        >
          Sign in
        </button>
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

      <h2 className="text-2xl font-bold text-gray-900 mb-1">Set new password</h2>
      <p className="text-gray-500 mb-6">
        Your new password must be at least 8 characters.
      </p>

      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1.5">New Password</label>
          <PasswordInput value={password} onChange={setPassword} placeholder="Minimum 8 characters" />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1.5">Confirm Password</label>
          <PasswordInput value={confirm} onChange={setConfirm} />
        </div>
        <button
          type="submit"
          disabled={loading}
          className="w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
        >
          <Lock size={18} />
          {loading ? 'Resetting...' : 'Reset password'}
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
