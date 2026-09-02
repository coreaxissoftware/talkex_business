import { Suspense, lazy, type ReactNode } from 'react'
import { createBrowserRouter } from 'react-router'

import AuthLayout from './layouts/AuthLayout'
import DashboardLayout from './layouts/DashboardLayout'
import ProtectedRoute from './components/ProtectedRoute'

// Eagerly-loaded pages: the auth flow + the Dashboard land page are on
// the critical first paint path — code-splitting them just adds a
// spinner. Everything else lazy-loads so the initial JS bundle stays
// under ~150 KB gz.
import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import ResetPassword from './pages/ResetPassword'
import OAuthCallback from './pages/OAuthCallback'
import Dashboard from './pages/Dashboard'

const Channels = lazy(() => import('./pages/Channels'))
const Contacts = lazy(() => import('./pages/Contacts'))
const Templates = lazy(() => import('./pages/Templates'))
const Campaigns = lazy(() => import('./pages/Campaigns'))
const Conversations = lazy(() => import('./pages/Conversations'))
const Automation = lazy(() => import('./pages/Automation'))
const Developers = lazy(() => import('./pages/Developers'))
const Webhooks = lazy(() => import('./pages/Webhooks'))
const Analytics = lazy(() => import('./pages/Analytics'))
const Logs = lazy(() => import('./pages/Logs'))
const Billing = lazy(() => import('./pages/Billing'))
const WalletPage = lazy(() => import('./pages/WalletPage'))
const Support = lazy(() => import('./pages/Support'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))
const ContactLists = lazy(() => import('./pages/ContactLists'))
const MediaLibrary = lazy(() => import('./pages/MediaLibrary'))
const Team = lazy(() => import('./pages/Team'))
const Tags = lazy(() => import('./pages/Tags'))
const Compliance = lazy(() => import('./pages/Compliance'))
const Organizations = lazy(() => import('./pages/Organizations'))
const ApiDocs = lazy(() => import('./pages/ApiDocs'))
const CannedResponses = lazy(() => import('./pages/CannedResponses'))
const CsatPage = lazy(() => import('./pages/CsatPage'))
const BroadcastCalendar = lazy(() => import('./pages/BroadcastCalendar'))
const Flows = lazy(() => import('./pages/Flows'))
const LiveChat = lazy(() => import('./pages/LiveChat'))
const TeamActivity = lazy(() => import('./pages/TeamActivity'))
const Deals = lazy(() => import('./pages/Deals'))
const Catalog = lazy(() => import('./pages/Catalog'))
const GreenTick = lazy(() => import('./pages/GreenTick'))
const Integrations = lazy(() => import('./pages/Integrations'))
const WhiteLabel = lazy(() => import('./pages/WhiteLabel'))
const Reseller = lazy(() => import('./pages/Reseller'))

// PageFallback — one shared skeleton while a lazy chunk loads. Kept
// visually quiet so a fast connection barely notices it.
function PageFallback() {
  return (
    <div className="flex items-center justify-center min-h-[40vh] text-sm text-gray-400">
      Loading…
    </div>
  )
}

function withSuspense(node: ReactNode): ReactNode {
  return <Suspense fallback={<PageFallback />}>{node}</Suspense>
}

const router = createBrowserRouter([
  {
    // Public auth pages — eagerly loaded, no fallback needed.
    element: <AuthLayout />,
    children: [
      { path: '/login', element: <Login /> },
      { path: '/register', element: <Register /> },
      { path: '/forgot-password', element: <ForgotPassword /> },
      { path: '/reset-password', element: <ResetPassword /> },
      { path: '/oauth/callback', element: <OAuthCallback /> },
    ],
  },
  {
    // Protected dashboard pages — lazy-loaded behind Suspense.
    element: <ProtectedRoute />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/', element: <Dashboard /> },
          { path: '/channels', element: withSuspense(<Channels />) },
          { path: '/contacts', element: withSuspense(<Contacts />) },
          { path: '/contact-lists', element: withSuspense(<ContactLists />) },
          { path: '/templates', element: withSuspense(<Templates />) },
          { path: '/media', element: withSuspense(<MediaLibrary />) },
          { path: '/campaigns', element: withSuspense(<Campaigns />) },
          { path: '/conversations', element: withSuspense(<Conversations />) },
          { path: '/automation', element: withSuspense(<Automation />) },
          { path: '/developers', element: withSuspense(<Developers />) },
          { path: '/webhooks', element: withSuspense(<Webhooks />) },
          { path: '/analytics', element: withSuspense(<Analytics />) },
          { path: '/logs', element: withSuspense(<Logs />) },
          { path: '/billing', element: withSuspense(<Billing />) },
          { path: '/wallet', element: withSuspense(<WalletPage />) },
          { path: '/support', element: withSuspense(<Support />) },
          { path: '/settings', element: withSuspense(<SettingsPage />) },
          { path: '/team', element: withSuspense(<Team />) },
          { path: '/tags', element: withSuspense(<Tags />) },
          { path: '/compliance', element: withSuspense(<Compliance />) },
          { path: '/organizations', element: withSuspense(<Organizations />) },
          { path: '/api-docs', element: withSuspense(<ApiDocs />) },
          { path: '/canned-responses', element: withSuspense(<CannedResponses />) },
          { path: '/csat', element: withSuspense(<CsatPage />) },
          { path: '/calendar', element: withSuspense(<BroadcastCalendar />) },
          { path: '/flows', element: withSuspense(<Flows />) },
          { path: '/live-chat', element: withSuspense(<LiveChat />) },
          { path: '/team/activity', element: withSuspense(<TeamActivity />) },
          { path: '/deals', element: withSuspense(<Deals />) },
          { path: '/catalog', element: withSuspense(<Catalog />) },
          { path: '/green-tick', element: withSuspense(<GreenTick />) },
          { path: '/integrations', element: withSuspense(<Integrations />) },
          { path: '/white-label', element: withSuspense(<WhiteLabel />) },
          { path: '/reseller', element: withSuspense(<Reseller />) },
        ],
      },
    ],
  },
])

export default router
