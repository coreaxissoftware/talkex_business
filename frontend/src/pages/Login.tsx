import { useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { useAuthStore } from '../store/authStore'
import SocialLoginButtons, { useOAuthProviders } from '../components/SocialLoginButtons'
import PasswordInput from '../components/PasswordInput'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuthStore()
  const navigate = useNavigate()
  const oauth = useOAuthProviders()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      if (rememberMe) localStorage.setItem('talkex_remember', 'true')
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.detail || err.response?.data?.error || 'Invalid email or password')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      <a
        href="https://business.talkex.in"
        style={{
          display: 'inline-flex', alignItems: 'center', gap: 6,
          fontSize: 12, color: 'var(--muted)',
          textDecoration: 'none', marginBottom: 40,
        }}
      >
        ← Back to <span style={{ fontFamily: 'var(--serif)', fontStyle: 'italic', marginLeft: 2 }}>business.talkex.in</span>
      </a>

      {/* Mobile-only brand mark */}
      <div style={{ marginBottom: 24 }} className="lg-hide">
        <div style={{
          width: 44, height: 44, borderRadius: 12,
          background: 'linear-gradient(135deg, var(--jade), var(--jade-deep))',
          display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
          color: 'var(--paper)', fontWeight: 800, fontSize: 20, letterSpacing: '-0.03em',
          boxShadow: '0 12px 32px rgba(14, 165, 160, 0.35)',
        }}>T</div>
      </div>
      <style>{`@media (min-width: 1024px) { .lg-hide { display: none !important; } }`}</style>

      <h1 className="headline-serif" style={{ marginBottom: 6 }}>
        Welcome back.
      </h1>
      <p style={{ fontSize: 14, color: 'var(--muted)', marginBottom: 28 }}>
        Sign in to your account
      </p>

      <SocialLoginButtons mode="login" />

      {oauth.anyEnabled && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '20px 0' }}>
          <div style={{ flex: 1, height: 1, background: 'var(--hair)' }} />
          <span style={{
            fontSize: 10, letterSpacing: '0.08em', color: 'var(--muted)',
            fontWeight: 600, textTransform: 'uppercase',
          }}>or continue with email</span>
          <div style={{ flex: 1, height: 1, background: 'var(--hair)' }} />
        </div>
      )}

      {error && (
        <div style={{
          marginBottom: 16, padding: '10px 14px',
          background: '#FEE2E2', border: '1px solid #FCA5A5',
          borderRadius: 10, fontSize: 13, color: '#991B1B',
        }}>
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: 16 }}>
          <label htmlFor="email" className="form-label">Email</label>
          <input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="form-input"
            placeholder="you@company.com"
            autoComplete="email"
          />
        </div>

        <div style={{ marginBottom: 16 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
            <label htmlFor="password" className="form-label" style={{ marginBottom: 0 }}>Password</label>
            <Link to="/forgot-password" style={{ fontSize: 11, color: 'var(--jade)', fontWeight: 600, textDecoration: 'none' }}>
              Forgot password?
            </Link>
          </div>
          <PasswordInput id="password" value={password} onChange={setPassword} />
        </div>

        <label style={{
          display: 'flex', alignItems: 'center', gap: 8,
          fontSize: 13, color: 'var(--ink-soft)', cursor: 'pointer',
          marginBottom: 20,
        }}>
          <input
            type="checkbox"
            checked={rememberMe}
            onChange={(e) => setRememberMe(e.target.checked)}
            style={{ accentColor: 'var(--jade)', width: 16, height: 16 }}
          />
          Remember me for 30 days
        </label>

        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? 'Signing in…' : (
            <>Sign in <span style={{ opacity: 0.6 }}>→</span></>
          )}
        </button>
      </form>

      <p style={{ marginTop: 24, textAlign: 'center', fontSize: 13, color: 'var(--muted)' }}>
        Don't have an account?{' '}
        <Link to="/register" className="link-jade" style={{ textDecoration: 'none' }}>
          Create one
        </Link>
      </p>
    </div>
  )
}
