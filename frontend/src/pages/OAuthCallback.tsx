import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { useAuthStore } from '../store/authStore'
import { Loader2 } from 'lucide-react'

/**
 * OAuthCallback reads the access_token and refresh_token from the URL
 * fragment (set by the backend OAuth callback) and stores them, then
 * redirects to the dashboard.
 *
 * Route: /oauth/callback#access_token=...&refresh_token=...&provider=...
 */
export default function OAuthCallback() {
  const navigate = useNavigate()
  const { setTokens } = useAuthStore()
  const [error, setError] = useState('')

  useEffect(() => {
    const hash = window.location.hash.slice(1) // remove leading #
    const params = new URLSearchParams(hash)
    const accessToken = params.get('access_token')
    const refreshToken = params.get('refresh_token')

    if (!accessToken || !refreshToken) {
      setError('OAuth login failed — no tokens received.')
      return
    }

    // Store tokens and redirect to dashboard
    setTokens(accessToken, refreshToken)
    navigate('/', { replace: true })
  }, [navigate, setTokens])

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error}</p>
          <a href="/login" className="text-primary-600 hover:underline">Back to login</a>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="flex items-center gap-3 text-gray-600">
        <Loader2 className="animate-spin" size={20} />
        <span>Completing login…</span>
      </div>
    </div>
  )
}
