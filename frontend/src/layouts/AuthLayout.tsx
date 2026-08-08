import { Outlet } from 'react-router'
import talkexIcon from '../assets/talkex-icon.png'

export default function AuthLayout() {
  return (
    <div className="flex min-h-screen">
      {/* Left panel — branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary-700 to-primary-900 items-center justify-center p-12">
        <div className="text-white max-w-md">
          <img src={talkexIcon} alt="TalkEx" className="h-16 w-16 rounded-2xl mb-6" />
          <h1 className="text-4xl font-bold mb-4">
            Talk<span className="text-primary-200">Ex</span> Business
          </h1>
          <p className="text-xl text-primary-200 mb-6">
            One Platform. Multiple Messaging Channels.
          </p>
          <p className="text-primary-300 text-sm leading-relaxed">
            Send messages through TalkEx, WhatsApp Business, and more — all
            from a single dashboard. Manage contacts, templates, campaigns,
            and analytics in one place.
          </p>
          <p className="mt-8 text-xs text-primary-400">by CoreAxis Ventures</p>
        </div>
      </div>

      {/* Right panel — auth form */}
      <div className="flex w-full lg:w-1/2 items-center justify-center p-8">
        <div className="w-full max-w-md">
          <Outlet />
        </div>
      </div>
    </div>
  )
}
