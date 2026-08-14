import { useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { useAuthStore } from '../store/authStore'
import { UserPlus, Check } from 'lucide-react'
import talkexIcon from '../assets/talkex-icon.png'
import SocialLoginButtons from '../components/SocialLoginButtons'
import PasswordInput from '../components/PasswordInput'
import Divider from '../components/Divider'
import api from '../services/api'

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

  const startTimer = (
    setTimer: React.Dispatch<React.SetStateAction<number>>
  ) => {
    let seconds = 30
    setTimer(seconds)
    const interval = setInterval(() => {
      seconds -= 1
      setTimer(seconds)
      if (seconds <= 0) clearInterval(interval)
    }, 1000)
  }

  const handleSendMobileOtp = async () => {
    if (!mobile || mobile.length < 10) {
      setError('Enter a valid 10-digit mobile number')
      return
    }
    setError('')
    try {
      await api.post('/auth/otp/send', { phone: countryCode + mobile })
      setMobileOtpSent(true)
      startTimer(setMobileTimer)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to send OTP')
    }
  }

  const handleVerifyMobileOtp = async () => {
    if (mobileOtp.length < 4) return
    try {
      await api.post('/auth/otp/verify', { phone: countryCode + mobile, code: mobileOtp })
      setMobileVerified(true)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Invalid OTP')
    }
  }

  const handleSendEmailOtp = async () => {
    if (!email || !email.includes('@')) {
      setError('Enter a valid email address')
      return
    }
    setError('')
    try {
      await api.post('/auth/otp/send', { email })
      setEmailOtpSent(true)
      startTimer(setEmailTimer)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to send OTP')
    }
  }

  const handleVerifyEmailOtp = async () => {
    if (emailOtp.length < 4) return
    try {
      await api.post('/auth/otp/verify', { email, code: emailOtp })
      setEmailVerified(true)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Invalid OTP')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!agreeTerms) {
      setError('Please accept the Terms & Conditions and Privacy Policy')
      return
    }
    setError('')
    setLoading(true)
    try {
      await register(email, password, fullName)
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.detail || err.response?.data?.error || 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div>
      {/* Mobile logo */}
      <div className="mb-6 lg:hidden text-center">
        <img src={talkexIcon} alt="TalkEx" className="h-12 w-12 rounded-xl mx-auto mb-3" />
        <h1 className="text-2xl font-bold text-gray-900">
          Talk<span className="text-primary-600">Ex</span> Business
        </h1>
      </div>

      <h2 className="text-2xl font-bold text-gray-900 mb-1">Create an account</h2>
      <p className="text-gray-500 mb-5">Start sending messages in minutes</p>

      {/* Social signup */}
      <SocialLoginButtons mode="register" />

      <Divider text="or register with email" />

      {/* Error message */}
      {error && (
        <div className="mb-4 rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {/* Registration form */}
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Full Name */}
        <div>
          <label htmlFor="fullName" className="block text-sm font-medium text-gray-700 mb-1.5">
            Full Name
          </label>
          <input
            id="fullName"
            type="text"
            required
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
            placeholder="Full Name"
            autoComplete="name"
          />
        </div>

        {/* Mobile with country code + Send OTP */}
        <div>
          <label htmlFor="mobile" className="block text-sm font-medium text-gray-700 mb-1.5">
            Mobile
          </label>
          <div className="flex gap-2">
            <div className="flex rounded-lg border border-gray-300 overflow-hidden shrink-0">
              <select
                value={countryCode}
                onChange={(e) => setCountryCode(e.target.value)}
                className="w-16 px-2 py-2.5 text-sm bg-gray-50 border-r border-gray-300 outline-none cursor-pointer"
              >
                <option value="+91">+91</option>
                <option value="+1">+1</option>
                <option value="+44">+44</option>
                <option value="+971">+971</option>
                <option value="+65">+65</option>
                <option value="+61">+61</option>
              </select>
              <input
                id="mobile"
                type="tel"
                required
                value={mobile}
                onChange={(e) => {
                  const val = e.target.value.replace(/\D/g, '').slice(0, 10)
                  setMobile(val)
                  if (mobileVerified) setMobileVerified(false)
                  if (mobileOtpSent) setMobileOtpSent(false)
                }}
                className="w-full px-3 py-2.5 text-sm outline-none"
                placeholder="Mobile"
                maxLength={10}
                disabled={mobileVerified}
              />
            </div>
            {mobileVerified ? (
              <span className="flex items-center gap-1 px-3 text-xs font-medium text-green-600 bg-green-50 rounded-lg border border-green-200 shrink-0">
                <Check size={14} /> Verified
              </span>
            ) : !mobileOtpSent ? (
              <button
                type="button"
                onClick={handleSendMobileOtp}
                className="px-4 py-2.5 text-sm font-medium text-red-500 border border-red-300 rounded-lg hover:bg-red-50 transition-colors shrink-0"
              >
                Send OTP
              </button>
            ) : (
              <button
                type="button"
                onClick={handleSendMobileOtp}
                disabled={mobileTimer > 0}
                className="px-4 py-2.5 text-sm font-medium text-gray-400 border border-gray-200 rounded-lg shrink-0 disabled:cursor-not-allowed"
              >
                {mobileTimer > 0 ? `${mobileTimer}s` : 'Resend'}
              </button>
            )}
          </div>
          {/* OTP input */}
          {mobileOtpSent && !mobileVerified && (
            <div className="flex gap-2 mt-2">
              <input
                type="text"
                value={mobileOtp}
                onChange={(e) => setMobileOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm tracking-widest text-center focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="Enter OTP"
                maxLength={6}
              />
              <button
                type="button"
                onClick={handleVerifyMobileOtp}
                disabled={mobileOtp.length < 4}
                className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 disabled:opacity-50 transition-colors"
              >
                Verify
              </button>
            </div>
          )}
        </div>

        {/* Email + Send OTP */}
        <div>
          <label htmlFor="regEmail" className="block text-sm font-medium text-gray-700 mb-1.5">
            Email
          </label>
          <div className="flex gap-2">
            <input
              id="regEmail"
              type="email"
              required
              value={email}
              onChange={(e) => {
                setEmail(e.target.value)
                if (emailVerified) setEmailVerified(false)
                if (emailOtpSent) setEmailOtpSent(false)
              }}
              className="flex-1 rounded-lg border border-gray-300 px-4 py-2.5 text-sm focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none transition-all"
              placeholder="Email"
              autoComplete="email"
              disabled={emailVerified}
            />
            {emailVerified ? (
              <span className="flex items-center gap-1 px-3 text-xs font-medium text-green-600 bg-green-50 rounded-lg border border-green-200 shrink-0">
                <Check size={14} /> Verified
              </span>
            ) : !emailOtpSent ? (
              <button
                type="button"
                onClick={handleSendEmailOtp}
                className="px-4 py-2.5 text-sm font-medium text-red-500 border border-red-300 rounded-lg hover:bg-red-50 transition-colors shrink-0"
              >
                Send OTP
              </button>
            ) : (
              <button
                type="button"
                onClick={handleSendEmailOtp}
                disabled={emailTimer > 0}
                className="px-4 py-2.5 text-sm font-medium text-gray-400 border border-gray-200 rounded-lg shrink-0 disabled:cursor-not-allowed"
              >
                {emailTimer > 0 ? `${emailTimer}s` : 'Resend'}
              </button>
            )}
          </div>
          {/* OTP input */}
          {emailOtpSent && !emailVerified && (
            <div className="flex gap-2 mt-2">
              <input
                type="text"
                value={emailOtp}
                onChange={(e) => setEmailOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-sm tracking-widest text-center focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 outline-none"
                placeholder="Enter OTP"
                maxLength={6}
              />
              <button
                type="button"
                onClick={handleVerifyEmailOtp}
                disabled={emailOtp.length < 4}
                className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 disabled:opacity-50 transition-colors"
              >
                Verify
              </button>
            </div>
          )}
        </div>

        {/* Password */}
        <div>
          <label htmlFor="regPassword" className="block text-sm font-medium text-gray-700 mb-1.5">
            Password
          </label>
          <PasswordInput
            id="regPassword"
            value={password}
            onChange={setPassword}
            placeholder="Minimum 8 characters"
          />
        </div>

        {/* Terms agreement */}
        <div className="flex items-start gap-2">
          <input
            id="terms"
            type="checkbox"
            checked={agreeTerms}
            onChange={(e) => setAgreeTerms(e.target.checked)}
            className="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
          />
          <label htmlFor="terms" className="text-sm text-gray-600 select-none cursor-pointer">
            I agree to{' '}
            <a href="/terms" className="text-primary-600 hover:underline">Terms & Conditions</a>
            {' & '}
            <a href="/privacy" className="text-primary-600 hover:underline">Privacy Policy</a>.
          </label>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full flex items-center justify-center gap-2 rounded-lg bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50 transition-colors"
        >
          <UserPlus size={18} />
          {loading ? 'Creating account...' : 'Create account'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-gray-500">
        Already have an account?{' '}
        <Link to="/login" className="font-medium text-primary-600 hover:text-primary-700">
          Sign in
        </Link>
      </p>
    </div>
  )
}
