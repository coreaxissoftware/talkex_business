import { CreditCard } from 'lucide-react'
import PlaceholderPage from '../components/PlaceholderPage'

export default function Billing() {
  return (
    <PlaceholderPage
      title="Billing"
      description="Plan details, invoices, GST, and payment history."
      icon={CreditCard}
      phase="Phase 2"
    />
  )
}
