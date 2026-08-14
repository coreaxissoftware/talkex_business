const googleIcon = (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <path d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z" fill="#4285F4"/>
    <path d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18z" fill="#34A853"/>
    <path d="M3.964 10.71A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.042l3.007-2.332z" fill="#FBBC05"/>
    <path d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z" fill="#EA4335"/>
  </svg>
)

const facebookIcon = (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <path d="M18 9a9 9 0 1 0-10.406 8.89v-6.29H5.309V9h2.285V7.017c0-2.255 1.343-3.501 3.4-3.501.984 0 2.014.176 2.014.176v2.215h-1.135c-1.118 0-1.467.694-1.467 1.406V9h2.496l-.399 2.6h-2.097v6.29A9.002 9.002 0 0 0 18 9z" fill="#1877F2"/>
    <path d="M12.202 11.6l.399-2.6H10.105V7.313c0-.712.349-1.406 1.467-1.406h1.135V3.692s-1.03-.176-2.014-.176c-2.057 0-3.4 1.246-3.4 3.501V9H5.008v2.6h2.285v6.29a9.07 9.07 0 0 0 2.812 0V11.6h2.097z" fill="#fff"/>
  </svg>
)

const appleIcon = (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <path d="M14.94 9.508c-.023-2.398 1.957-3.55 2.046-3.607-1.114-1.63-2.849-1.853-3.467-1.879-1.476-.15-2.881.87-3.63.87-.747 0-1.904-.848-3.128-.825-1.61.024-3.094.936-3.924 2.379-1.672 2.903-.428 7.203 1.202 9.559.797 1.152 1.747 2.447 2.996 2.4 1.202-.048 1.656-.778 3.108-.778 1.452 0 1.86.778 3.132.754 1.295-.024 2.112-1.175 2.903-2.33.915-1.337 1.291-2.63 1.314-2.697-.029-.013-2.523-.968-2.552-3.846zM12.548 2.458C13.215 1.65 13.668.548 13.548-.5c-.95.039-2.103.633-2.784 1.43-.611.707-1.145 1.838-1.002 2.922 1.06.082 2.143-.538 2.786-1.394z" fill="currentColor"/>
  </svg>
)

const githubIcon = (
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
    <path fillRule="evenodd" clipRule="evenodd" d="M9 0C4.027 0 0 4.13 0 9.228c0 4.078 2.578 7.535 6.154 8.758.45.084.615-.2.615-.442 0-.218-.008-.795-.012-1.56-2.503.557-3.032-1.238-3.032-1.238-.41-1.066-1-1.35-1-1.35-.816-.572.062-.56.062-.56.903.065 1.378.952 1.378.952.803 1.41 2.107 1.003 2.62.767.082-.596.314-1.003.571-1.233-1.998-.233-4.1-1.025-4.1-4.562 0-1.008.35-1.832.926-2.478-.093-.233-.401-1.172.088-2.443 0 0 .755-.248 2.473.946A8.4 8.4 0 0 1 9 4.462a8.4 8.4 0 0 1 2.253.312c1.717-1.194 2.472-.946 2.472-.946.49 1.271.182 2.21.09 2.443.576.646.925 1.47.925 2.478 0 3.546-2.104 4.326-4.108 4.556.323.285.611.849.611 1.71 0 1.234-.011 2.23-.011 2.533 0 .244.163.53.619.44C15.425 16.76 18 13.304 18 9.228 18 4.13 13.971 0 9 0z" fill="currentColor"/>
  </svg>
)

interface SocialLoginButtonsProps {
  mode: 'login' | 'register'
}

export default function SocialLoginButtons({ mode }: SocialLoginButtonsProps) {
  const label = mode === 'login' ? 'Sign in' : 'Sign up'

  const handleSocialLogin = (provider: string) => {
    // Redirect to backend OAuth initiation endpoint which either
    // redirects to the real provider (if configured) or simulates
    // the flow in dev mode.
    const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    window.location.href = `${baseURL}/auth/oauth/${provider}`
  }

  return (
    <div className="space-y-2.5">
      <button
        type="button"
        onClick={() => handleSocialLogin('google')}
        className="w-full flex items-center justify-center gap-3 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-all"
      >
        {googleIcon}
        {label} with Google
      </button>

      <button
        type="button"
        onClick={() => handleSocialLogin('facebook')}
        className="w-full flex items-center justify-center gap-3 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-all"
      >
        {facebookIcon}
        {label} with Facebook
      </button>

      <div className="grid grid-cols-2 gap-2.5">
        <button
          type="button"
          onClick={() => handleSocialLogin('apple')}
          className="flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-all"
        >
          {appleIcon}
          Apple
        </button>

        <button
          type="button"
          onClick={() => handleSocialLogin('github')}
          className="flex items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 hover:border-gray-400 transition-all"
        >
          {githubIcon}
          GitHub
        </button>
      </div>
    </div>
  )
}
