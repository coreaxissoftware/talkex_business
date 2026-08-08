import { MessageSquare } from 'lucide-react'
import PlaceholderPage from '../components/PlaceholderPage'

export default function Conversations() {
  return (
    <PlaceholderPage
      title="Conversations"
      description="Shared inbox with live chat, agent assignment, and labels."
      icon={MessageSquare}
      phase="MVP"
    />
  )
}
