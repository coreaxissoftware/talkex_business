import { Settings } from 'lucide-react'
import PlaceholderPage from '../components/PlaceholderPage'

export default function SettingsPage() {
  return (
    <PlaceholderPage
      title="Settings"
      description="Profile, security, API configuration, webhooks, 2FA, and sessions."
      icon={Settings}
      phase="Phase 2"
    />
  )
}
