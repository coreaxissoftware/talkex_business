import { useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { useAuthStore } from '../store/authStore'
import { LogIn } from 'lucide-react'
import talkexIcon from '../assets/talkex-icon.png'
import SocialLoginButtons from '../components/SocialLoginButtons'
import PasswordInput from '../components/PasswordInput'
import Divider from '../components/Divider'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuthStore()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      if (rememberMe) {
        localStorage.setItem('talkex_remember', 'true')
      }
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.detail || err.response?.data?.error || 'Invalid email or password')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      {/* Back to marketing site */}
      <div className="mb-4">
        <a
          href="https://talkex.io"
          className="text-xs text-gray-500 hover:text-primary-600 inline-flex items-center gap-1"
        >
          ← Back to talkex.io
        </a>
      </div>

      {/* Mobile logo */}
      <div className="mb-8 lg:hidden text-center">
        <img src={talkexIcon} alt="TalkEx" className="h-12 w-12 rounded-xl mx-auto mb-3" />
        <h1 className="text-2xl font-bold text-gray-900">
          Talk<span className="text-primary-600">Ex</span> Business
        </h1>
      </div>

      <h2 className="text-2xl font-bold text-gray-900 mb-1">Welcome back</h2>
      <p className="text-gray-500 mb-6">Sign in to your account</p>

      {/* Social login */}
      <SocialLoginButtons mode="login" />

      <Divider text="or continue with email" />

      {/* Error message */}
      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Email/password form */}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1.5">
            Email
          </label>
          <input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
            placeholder="you@company.com"
            autoComplete="email"
          />
        </div>

        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label htmlFor="password" className="block text-sm font-medium text-gray-700">
              Password
            </label>
            <Link
              to="/forgot-password"
              className="text-xs font-medium text-primary-600 hover:text-primary-700"
            >
              Forgot password?
            </Link>
          </div>
          <PasswordInput
            id="password"
            value={password}
            onChange={setPassword}
          />
        </div>

        {/* Remember me */}
        <div className="flex items-center gap-2">
          <input
            id="remember"
            type="checkbox"
            checked={rememberMe}
            onChange={(e) => setRememberMe(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
          />
          <label htmlFor="remember" className="text-sm text-gray-600 select-none cursor-pointer">
            Remember me for 30 days
          </label>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
        >
          <LogIn size={18} />
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-gray-500">
        Don't have an account?{' '}
        <Link to="/register" className="font-medium text-primary-600 hover:text-primary-700">
          Create one
        </Link>
      </p>
    </div>
  )
}
