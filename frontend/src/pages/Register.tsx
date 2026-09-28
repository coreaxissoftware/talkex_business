import { useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { useAuthStore } from '../store/authStore'
import { Check } from 'lucide-react'
import SocialLoginButtons, { useOAuthProviders } from '../components/SocialLoginButtons'
import PasswordInput from '../components/PasswordInput'
import api from '../services/api'

const COUNTRY_CODES = ['+91', '+1', '+44', '+971', '+65', '+61']

export default function Register() {
  const [fullName, setFullName] = useState('')
  const [countryCode, setCountryCode] = useState('+91')
  const [mobile, setMobile] = useState('')
  const [mobileOtp, setMobileOtp] = useState('')
  const [mobileOtpSent, setMobileOtpSent] = useState(false)
  const [mobileVerified, setMobileVerified] = useState(false)
  const [mobileTimer, setMobileTimer] = useState(0)

  const [email, setEmail] = useState('')
  const [emailOtp, setEmailOtp] = useState('')
  const [emailOtpSent, setEmailOtpSent] = useState(false)
  const [emailVerified, setEmailVerified] = useState(false)
  const [emailTimer, setEmailTimer] = useState(0)

  const [password, setPassword] = useState('')
  const [agreeTerms, setAgreeTerms] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { register } = useAuthStore()
  const navigate = useNavigate()
  const oauth = useOAuthProviders()

  const startTimer = (setTimer: React.Dispatch<React.SetStateAction<number>>) => {
    let seconds = 30
    setTimer(seconds)
    const interval = setInterval(() => {
      seconds -= 1
      setTimer(seconds)
      if (seconds <= 0) clearInterval(interval)
    }, 1000)
  }

  const sendMobileOtp = async () => {
    if (mobile.length < 10) { setError('Enter a valid 10-digit mobile number'); return }
    setError('')
    try {
      await api.post('/auth/otp/send', { phone: countryCode + mobile })
      setMobileOtpSent(true); startTimer(setMobileTimer)
    } catch (err: any) { setError(err.response?.data?.detail || 'Failed to send OTP') }
  }
  const verifyMobileOtp = async () => {
    if (mobileOtp.length < 4) return
    try {
      await api.post('/auth/otp/verify', { phone: countryCode + mobile, code: mobileOtp })
      setMobileVerified(true)
    } catch (err: any) { setError(err.response?.data?.detail || 'Invalid OTP') }
  }
  const sendEmailOtp = async () => {
    if (!email.includes('@')) { setError('Enter a valid email address'); return }
    setError('')
    try {
      await api.post('/auth/otp/send', { email })
      setEmailOtpSent(true); startTimer(setEmailTimer)
    } catch (err: any) { setError(err.response?.data?.detail || 'Failed to send OTP') }
  }
  const verifyEmailOtp = async () => {
    if (emailOtp.length < 4) return
    try {
      await api.post('/auth/otp/verify', { email, code: emailOtp })
      setEmailVerified(true)
    } catch (err: any) { setError(err.response?.data?.detail || 'Invalid OTP') }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!agreeTerms) { setError('Please accept the Terms & Privacy Policy'); return }
    setError(''); setLoading(true)
    try {
      await register(email, password, fullName)
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.detail || err.response?.data?.error || 'Registration failed')
    } finally { setLoading(false) }
  }

  return (
    <div>
      <a
        href="https://business.talkex.in"
        style={{
          display: 'inline-flex', alignItems: 'center', gap: 6,
          fontSize: 12, color: 'var(--muted)',
          textDecoration: 'none', marginBottom: 20,
        }}
      >
        ← Back to <span style={{ fontFamily: 'var(--serif)', fontStyle: 'italic', marginLeft: 2 }}>business.talkex.in</span>
      </a>

      <h1 className="headline-serif" style={{ marginBottom: 4, fontSize: 30 }}>
        Create your account.
      </h1>
      <p style={{ fontSize: 13, color: 'var(--muted)', marginBottom: 20 }}>
        Start sending in under two minutes. No card required.
      </p>

      <SocialLoginButtons mode="register" />

      {oauth.anyEnabled && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, margin: '14px 0' }}>
          <div style={{ flex: 1, height: 1, background: 'var(--hair)' }} />
          <span style={{
            fontSize: 10, letterSpacing: '0.08em', color: 'var(--muted)',
            fontWeight: 600, textTransform: 'uppercase',
          }}>or with email</span>
          <div style={{ flex: 1, height: 1, background: 'var(--hair)' }} />
        </div>
      )}

      {error && (
        <div style={{
          marginBottom: 12, padding: '8px 12px',
          background: '#FEE2E2', border: '1px solid #FCA5A5',
          borderRadius: 8, fontSize: 12, color: '#991B1B',
        }}>{error}</div>
      )}

      <form onSubmit={handleSubmit}>
        {/* Full name */}
        <FieldRow>
          <Field label="Full name">
            <input
              className="form-input" required
              value={fullName} onChange={(e) => setFullName(e.target.value)}
              placeholder="Aditi Sharma" autoComplete="name"
              style={{ padding: '9px 12px', fontSize: 13 }}
            />
          </Field>
        </FieldRow>

        {/* Mobile + OTP */}
        <FieldRow>
          <Field label={<>Mobile {mobileVerified && <VerifiedChip />}</>}>
            {!mobileOtpSent ? (
              <div style={{ display: 'flex', gap: 6 }}>
                <select
                  value={countryCode} onChange={(e) => setCountryCode(e.target.value)}
                  className="form-input"
                  style={{ width: 82, padding: '9px 8px', fontSize: 13, cursor: 'pointer' }}
                >
                  {COUNTRY_CODES.map((c) => <option key={c} value={c}>{c}</option>)}
                </select>
                <input
                  required type="tel" inputMode="numeric" maxLength={10}
                  value={mobile}
                  onChange={(e) => setMobile(e.target.value.replace(/\D/g, '').slice(0, 10))}
                  placeholder="9876543210"
                  className="form-input"
                  style={{ padding: '9px 12px', fontSize: 13, flex: 1 }}
                />
                <button
                  type="button" onClick={sendMobileOtp}
                  disabled={mobile.length < 10}
                  style={sendOtpBtn}
                >Send OTP</button>
              </div>
            ) : (
              <div style={{ display: 'flex', gap: 6 }}>
                <input
                  required inputMode="numeric" maxLength={6} autoFocus
                  value={mobileOtp}
                  onChange={(e) => setMobileOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  placeholder="6-digit OTP"
                  className="form-input"
                  style={{
                    padding: '9px 12px', fontSize: 14,
                    fontFamily: 'var(--mono)', letterSpacing: '0.12em',
                    textAlign: 'center', flex: 1,
                    background: mobileVerified ? '#DCFCE7' : '#FFFFFF',
                    borderColor: mobileVerified ? '#86EFAC' : undefined,
                  }}
                  disabled={mobileVerified}
                />
                {!mobileVerified ? (
                  <button type="button" onClick={verifyMobileOtp} disabled={mobileOtp.length < 4} style={verifyBtn}>
                    Verify
                  </button>
                ) : (
                  <button type="button" disabled style={{ ...verifyBtn, background: '#DCFCE7', color: '#166534', border: '1px solid #86EFAC' }}>
                    <Check size={14} /> Verified
                  </button>
                )}
                {!mobileVerified && (
                  <button
                    type="button"
                    onClick={sendMobileOtp}
                    disabled={mobileTimer > 0}
                    style={{ ...linkBtn, opacity: mobileTimer > 0 ? 0.5 : 1 }}
                  >
                    {mobileTimer > 0 ? `${mobileTimer}s` : 'Resend'}
                  </button>
                )}
              </div>
            )}
          </Field>
        </FieldRow>

        {/* Email + OTP */}
        <FieldRow>
          <Field label={<>Email {emailVerified && <VerifiedChip />}</>}>
            {!emailOtpSent ? (
              <div style={{ display: 'flex', gap: 6 }}>
                <input
                  required type="email"
                  value={email} onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@company.com" autoComplete="email"
                  className="form-input"
                  style={{ padding: '9px 12px', fontSize: 13, flex: 1 }}
                />
                <button type="button" onClick={sendEmailOtp} disabled={!email.includes('@')} style={sendOtpBtn}>
                  Send OTP
                </button>
              </div>
            ) : (
              <div style={{ display: 'flex', gap: 6 }}>
                <input
                  required inputMode="numeric" maxLength={6}
                  value={emailOtp}
                  onChange={(e) => setEmailOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  placeholder="6-digit OTP"
                  className="form-input"
                  style={{
                    padding: '9px 12px', fontSize: 14,
                    fontFamily: 'var(--mono)', letterSpacing: '0.12em',
                    textAlign: 'center', flex: 1,
                    background: emailVerified ? '#DCFCE7' : '#FFFFFF',
                    borderColor: emailVerified ? '#86EFAC' : undefined,
                  }}
                  disabled={emailVerified}
                />
                {!emailVerified ? (
                  <button type="button" onClick={verifyEmailOtp} disabled={emailOtp.length < 4} style={verifyBtn}>
                    Verify
                  </button>
                ) : (
                  <button type="button" disabled style={{ ...verifyBtn, background: '#DCFCE7', color: '#166534', border: '1px solid #86EFAC' }}>
                    <Check size={14} /> Verified
                  </button>
                )}
                {!emailVerified && (
                  <button
                    type="button" onClick={sendEmailOtp} disabled={emailTimer > 0}
                    style={{ ...linkBtn, opacity: emailTimer > 0 ? 0.5 : 1 }}
                  >
                    {emailTimer > 0 ? `${emailTimer}s` : 'Resend'}
                  </button>
                )}
              </div>
            )}
          </Field>
        </FieldRow>

        {/* Password */}
        <FieldRow>
          <Field label="Password">
            <PasswordInput id="password" value={password} onChange={setPassword} />
          </Field>
        </FieldRow>

        {/* T&C */}
        <label style={{
          display: 'flex', alignItems: 'flex-start', gap: 8,
          fontSize: 12, color: 'var(--ink-soft)', cursor: 'pointer',
          margin: '4px 0 14px',
        }}>
          <input
            type="checkbox"
            checked={agreeTerms}
            onChange={(e) => setAgreeTerms(e.target.checked)}
            style={{ accentColor: 'var(--jade)', width: 15, height: 15, marginTop: 2, flexShrink: 0 }}
          />
          <span>
            I agree to the{' '}
            <a href="https://business.talkex.in/#/legal/terms" target="_blank" rel="noopener noreferrer" className="link-jade" style={{ textDecoration: 'none' }}>Terms</a>
            {' & '}
            <a href="https://business.talkex.in/#/legal/privacy" target="_blank" rel="noopener noreferrer" className="link-jade" style={{ textDecoration: 'none' }}>Privacy Policy</a>
          </span>
        </label>

        <button type="submit" disabled={loading || !agreeTerms} className="btn-primary" style={{ padding: '11px 18px' }}>
          {loading ? 'Creating account…' : (
            <>Create account <span style={{ opacity: 0.6 }}>→</span></>
          )}
        </button>
      </form>

      <p style={{ marginTop: 14, textAlign: 'center', fontSize: 13, color: 'var(--muted)' }}>
        Already have an account?{' '}
        <Link to="/login" className="link-jade" style={{ textDecoration: 'none' }}>Sign in</Link>
      </p>
    </div>
  )
}

// ── small helpers ─────────────────────────────────────────────────────

function FieldRow({ children }: { children: React.ReactNode }) {
  return <div style={{ marginBottom: 12 }}>{children}</div>
}

function Field({ label, children }: { label: React.ReactNode; children: React.ReactNode }) {
  return (
    <>
      <label className="form-label" style={{
        marginBottom: 5,
        display: 'flex', alignItems: 'center', gap: 6,
      }}>
        {label}
      </label>
      {children}
    </>
  )
}

function VerifiedChip() {
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 3,
      fontSize: 9, padding: '1px 6px', borderRadius: 999,
      background: '#DCFCE7', color: '#166534',
      textTransform: 'none', letterSpacing: 0,
    }}>
      <Check size={9} strokeWidth={3} /> verified
    </span>
  )
}

const sendOtpBtn: React.CSSProperties = {
  padding: '9px 14px',
  fontSize: 12,
  fontWeight: 600,
  background: 'var(--ink)',
  color: 'var(--paper)',
  border: 'none',
  borderRadius: 10,
  cursor: 'pointer',
  whiteSpace: 'nowrap',
}

const verifyBtn: React.CSSProperties = {
  padding: '9px 14px',
  fontSize: 12,
  fontWeight: 600,
  background: 'var(--jade)',
  color: '#FFFFFF',
  border: 'none',
  borderRadius: 10,
  cursor: 'pointer',
  whiteSpace: 'nowrap',
  display: 'inline-flex', alignItems: 'center', gap: 4,
}

const linkBtn: React.CSSProperties = {
  padding: '9px 8px',
  fontSize: 11,
  fontWeight: 600,
  background: 'transparent',
  color: 'var(--muted)',
  border: 'none',
  cursor: 'pointer',
  whiteSpace: 'nowrap',
}
