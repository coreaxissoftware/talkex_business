import { createBrowserRouter } from 'react-router'

import AuthLayout from './layouts/AuthLayout'
import DashboardLayout from './layouts/DashboardLayout'
import ProtectedRoute from './components/ProtectedRoute'

import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import Dashboard from './pages/Dashboard'
import Channels from './pages/Channels'
import Contacts from './pages/Contacts'
import Templates from './pages/Templates'
import Campaigns from './pages/Campaigns'
import Conversations from './pages/Conversations'
import Automation from './pages/Automation'
import Developers from './pages/Developers'
import Analytics from './pages/Analytics'
import Logs from './pages/Logs'
import Billing from './pages/Billing'
import WalletPage from './pages/WalletPage'
import Support from './pages/Support'
import SettingsPage from './pages/SettingsPage'

const router = createBrowserRouter([
  {
    // Public auth pages
    element: <AuthLayout />,
    children: [
      { path: '/login', element: <Login /> },
      { path: '/register', element: <Register /> },
      { path: '/forgot-password', element: <ForgotPassword /> },
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
          { path: '/templates', element: <Templates /> },
          { path: '/campaigns', element: <Campaigns /> },
          { path: '/conversations', element: <Conversations /> },
          { path: '/automation', element: <Automation /> },
          { path: '/developers', element: <Developers /> },
          { path: '/analytics', element: <Analytics /> },
          { path: '/logs', element: <Logs /> },
          { path: '/billing', element: <Billing /> },
          { path: '/wallet', element: <WalletPage /> },
          { path: '/support', element: <Support /> },
          { path: '/settings', element: <SettingsPage /> },
        ],
      },
    ],
  },
])

export default router
