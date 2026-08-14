import { createBrowserRouter } from 'react-router'

import AuthLayout from './layouts/AuthLayout'
import DashboardLayout from './layouts/DashboardLayout'
import ProtectedRoute from './components/ProtectedRoute'

import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import ResetPassword from './pages/ResetPassword'
import Dashboard from './pages/Dashboard'
import Channels from './pages/Channels'
import Contacts from './pages/Contacts'
import Templates from './pages/Templates'
import Campaigns from './pages/Campaigns'
import Conversations from './pages/Conversations'
import Automation from './pages/Automation'
import Developers from './pages/Developers'
import Webhooks from './pages/Webhooks'
import Analytics from './pages/Analytics'
import Logs from './pages/Logs'
import Billing from './pages/Billing'
import WalletPage from './pages/WalletPage'
import Support from './pages/Support'
import SettingsPage from './pages/SettingsPage'
import ContactLists from './pages/ContactLists'
import MediaLibrary from './pages/MediaLibrary'
import Team from './pages/Team'
import Tags from './pages/Tags'
import Compliance from './pages/Compliance'
import Organizations from './pages/Organizations'

const router = createBrowserRouter([
  {
    // Public auth pages
    element: <AuthLayout />,
    children: [
      { path: '/login', element: <Login /> },
      { path: '/register', element: <Register /> },
      { path: '/forgot-password', element: <ForgotPassword /> },
      { path: '/reset-password', element: <ResetPassword /> },
    ],
  },
  {
    // Protected dashboard pages
    element: <ProtectedRoute />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/', element: <Dashboard /> },
          { path: '/channels', element: <Channels /> },
          { path: '/contacts', element: <Contacts /> },
          { path: '/contact-lists', element: <ContactLists /> },
          { path: '/templates', element: <Templates /> },
          { path: '/media', element: <MediaLibrary /> },
          { path: '/campaigns', element: <Campaigns /> },
          { path: '/conversations', element: <Conversations /> },
          { path: '/automation', element: <Automation /> },
          { path: '/developers', element: <Developers /> },
          { path: '/webhooks', element: <Webhooks /> },
          { path: '/analytics', element: <Analytics /> },
          { path: '/logs', element: <Logs /> },
          { path: '/billing', element: <Billing /> },
          { path: '/wallet', element: <WalletPage /> },
          { path: '/support', element: <Support /> },
          { path: '/settings', element: <SettingsPage /> },
          { path: '/team', element: <Team /> },
          { path: '/tags', element: <Tags /> },
          { path: '/compliance', element: <Compliance /> },
          { path: '/organizations', element: <Organizations /> },
        ],
      },
    ],
  },
])

export default router
