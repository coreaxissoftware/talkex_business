import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router'
import router from './router'
import './index.css'
import './i18n' // side-effect: initializes react-i18next before any page renders
import { installClientSecurity } from './lib/security'

// Banking-grade client-side hardening: right-click block, devtools
// shortcut deterrent, iframe buster, drag-away lock. No-op in dev mode.
installClientSecurity()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
