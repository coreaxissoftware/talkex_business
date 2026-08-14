import { useState } from 'react'
import { Book, ChevronDown, ChevronRight, Copy, CheckCircle2 } from 'lucide-react'

interface Endpoint {
  method: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  path: string
  description: string
  auth: boolean
  body?: string
}

interface Section {
  title: string
  endpoints: Endpoint[]
}

const API_SECTIONS: Section[] = [
  {
    title: 'Authentication',
    endpoints: [
      { method: 'POST', path: '/auth/register', description: 'Create a new account', auth: false, body: '{ "email": "string", "password": "string", "full_name": "string" }' },
      { method: 'POST', path: '/auth/login', description: 'Sign in and receive tokens', auth: false, body: '{ "email": "string", "password": "string" }' },
      { method: 'POST', path: '/auth/refresh', description: 'Rotate access + refresh tokens', auth: false, body: '{ "refresh_token": "string" }' },
      { method: 'POST', path: '/auth/forgot-password', description: 'Request a password reset link', auth: false, body: '{ "email": "string" }' },
      { method: 'POST', path: '/auth/reset-password', description: 'Reset password with token', auth: false, body: '{ "token": "string", "new_password": "string" }' },
    ],
  },
  {
    title: 'User Profile',
    endpoints: [
      { method: 'GET', path: '/users/me', description: 'Get current user profile', auth: true },
      { method: 'PATCH', path: '/users/me', description: 'Update profile fields', auth: true, body: '{ "full_name": "string", "business_category": "string" }' },
      { method: 'POST', path: '/users/me/change-password', description: 'Change password', auth: true, body: '{ "current_password": "string", "new_password": "string" }' },
      { method: 'POST', path: '/users/me/2fa/setup', description: 'Start 2FA setup (returns secret + QR URI)', auth: true },
      { method: 'POST', path: '/users/me/2fa/verify', description: 'Verify TOTP code to enable 2FA', auth: true, body: '{ "code": "string" }' },
      { method: 'POST', path: '/users/me/2fa/disable', description: 'Disable 2FA', auth: true, body: '{ "password": "string", "code": "string" }' },
      { method: 'GET', path: '/users/me/sessions', description: 'List active sessions', auth: true },
      { method: 'DELETE', path: '/users/me/sessions/:sid', description: 'Revoke a session', auth: true },
    ],
  },
  {
    title: 'Contacts',
    endpoints: [
      { method: 'GET', path: '/contacts', description: 'List all contacts (supports ?search, ?page, ?per_page)', auth: true },
      { method: 'POST', path: '/contacts', description: 'Create a contact', auth: true, body: '{ "phone_number": "string", "name": "string", "email": "string" }' },
      { method: 'GET', path: '/contacts/:id', description: 'Get a contact by ID', auth: true },
      { method: 'PATCH', path: '/contacts/:id', description: 'Update a contact', auth: true },
      { method: 'DELETE', path: '/contacts/:id', description: 'Delete a contact', auth: true },
      { method: 'POST', path: '/contacts/import', description: 'Bulk import from CSV', auth: true },
      { method: 'GET', path: '/contacts/export', description: 'Export contacts as CSV', auth: true },
    ],
  },
  {
    title: 'Contact Lists',
    endpoints: [
      { method: 'GET', path: '/contact-lists', description: 'List all contact lists with member counts', auth: true },
      { method: 'POST', path: '/contact-lists', description: 'Create a contact list', auth: true, body: '{ "name": "string", "description": "string" }' },
      { method: 'PATCH', path: '/contact-lists/:id', description: 'Update a contact list', auth: true },
      { method: 'DELETE', path: '/contact-lists/:id', description: 'Delete a contact list', auth: true },
      { method: 'GET', path: '/contact-lists/:id/members', description: 'Get member IDs', auth: true },
      { method: 'POST', path: '/contact-lists/:id/members', description: 'Add members', auth: true, body: '{ "contact_ids": ["string"] }' },
      { method: 'DELETE', path: '/contact-lists/:id/members', description: 'Remove members', auth: true, body: '{ "contact_ids": ["string"] }' },
    ],
  },
  {
    title: 'Templates',
    endpoints: [
      { method: 'GET', path: '/templates', description: 'List all message templates', auth: true },
      { method: 'POST', path: '/templates', description: 'Create a template', auth: true, body: '{ "name": "string", "category": "marketing|utility|authentication", "channel": "string", "body": "string", "variables": ["string"] }' },
      { method: 'PATCH', path: '/templates/:id', description: 'Update a template', auth: true },
      { method: 'DELETE', path: '/templates/:id', description: 'Delete a template', auth: true },
    ],
  },
  {
    title: 'Campaigns',
    endpoints: [
      { method: 'GET', path: '/campaigns', description: 'List all campaigns', auth: true },
      { method: 'POST', path: '/campaigns', description: 'Create a campaign', auth: true, body: '{ "name": "string", "template_id": "string", "contact_ids": ["string"], "list_id": "string", "scheduled_at": "ISO8601" }' },
      { method: 'GET', path: '/campaigns/:id', description: 'Get campaign details', auth: true },
      { method: 'PATCH', path: '/campaigns/:id', description: 'Update a draft campaign', auth: true },
      { method: 'POST', path: '/campaigns/:id/launch', description: 'Launch a campaign', auth: true },
      { method: 'POST', path: '/campaigns/:id/cancel', description: 'Cancel a campaign', auth: true },
      { method: 'POST', path: '/campaigns/:id/clone', description: 'Duplicate a campaign', auth: true },
      { method: 'POST', path: '/campaigns/:id/approve', description: 'Approve a pending campaign', auth: true },
      { method: 'POST', path: '/campaigns/:id/reject', description: 'Reject a pending campaign', auth: true, body: '{ "reason": "string" }' },
      { method: 'DELETE', path: '/campaigns/:id', description: 'Delete a campaign', auth: true },
    ],
  },
  {
    title: 'Conversations',
    endpoints: [
      { method: 'GET', path: '/conversations', description: 'List conversations', auth: true },
      { method: 'GET', path: '/conversations/:id/messages', description: 'Get messages in a conversation', auth: true },
      { method: 'POST', path: '/conversations/send', description: 'Send an outbound message', auth: true, body: '{ "contact_id": "string", "channel": "string", "body": "string", "template_id": "string" }' },
    ],
  },
  {
    title: 'Channels',
    endpoints: [
      { method: 'GET', path: '/channels/catalog', description: 'Get available channel catalog', auth: true },
      { method: 'GET', path: '/channels', description: 'List enabled channel configs', auth: true },
      { method: 'PUT', path: '/channels/:kind', description: 'Enable/disable a channel', auth: true, body: '{ "enabled": true, "config": {} }' },
    ],
  },
  {
    title: 'Messaging Queue',
    endpoints: [
      { method: 'GET', path: '/messaging/queue', description: 'View queued messages', auth: true },
      { method: 'GET', path: '/messaging/dlq', description: 'View dead letter queue', auth: true },
      { method: 'POST', path: '/messaging/dlq/:id/retry', description: 'Retry a dead letter', auth: true },
      { method: 'GET', path: '/messaging/cost', description: 'Get cost summary', auth: true },
    ],
  },
  {
    title: 'Analytics',
    endpoints: [
      { method: 'GET', path: '/analytics/summary', description: 'Dashboard KPI summary', auth: true },
      { method: 'GET', path: '/analytics/timeseries?days=30', description: 'Daily message volume chart data', auth: true },
    ],
  },
  {
    title: 'Wallet',
    endpoints: [
      { method: 'GET', path: '/wallet', description: 'Get wallet balance', auth: true },
      { method: 'POST', path: '/wallet/credit', description: 'Add funds (dev/testing)', auth: true, body: '{ "amount": 100, "description": "string" }' },
      { method: 'GET', path: '/wallet/transactions', description: 'List transactions', auth: true },
    ],
  },
  {
    title: 'Webhooks',
    endpoints: [
      { method: 'GET', path: '/webhooks', description: 'List webhook endpoints', auth: true },
      { method: 'POST', path: '/webhooks', description: 'Create a webhook endpoint', auth: true, body: '{ "url": "string", "events": ["string"], "secret": "string" }' },
      { method: 'PATCH', path: '/webhooks/:id', description: 'Update a webhook', auth: true },
      { method: 'DELETE', path: '/webhooks/:id', description: 'Delete a webhook', auth: true },
      { method: 'GET', path: '/webhooks/:id/deliveries', description: 'View delivery logs', auth: true },
      { method: 'POST', path: '/webhooks/deliveries/:id/retry', description: 'Retry a delivery', auth: true },
    ],
  },
  {
    title: 'API Keys',
    endpoints: [
      { method: 'GET', path: '/api-keys', description: 'List API keys', auth: true },
      { method: 'POST', path: '/api-keys', description: 'Create an API key', auth: true, body: '{ "name": "string" }' },
      { method: 'DELETE', path: '/api-keys/:id', description: 'Revoke an API key', auth: true },
    ],
  },
  {
    title: 'Inbound Webhooks (Channel Providers)',
    endpoints: [
      { method: 'GET', path: '/webhooks/channels/:channel', description: 'Webhook verification (Meta challenge)', auth: false },
      { method: 'POST', path: '/webhooks/channels/:channel', description: 'Receive inbound messages / status updates', auth: false },
    ],
  },
]

const METHOD_COLORS: Record<string, string> = {
  GET: 'bg-blue-100 text-blue-800',
  POST: 'bg-green-100 text-green-800',
  PATCH: 'bg-amber-100 text-amber-800',
  PUT: 'bg-orange-100 text-orange-800',
  DELETE: 'bg-red-100 text-red-800',
}

export default function ApiDocs() {
  const [expanded, setExpanded] = useState<Set<number>>(new Set([0]))
  const [copied, setCopied] = useState<string | null>(null)

  const toggle = (i: number) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(i)) next.delete(i)
      else next.add(i)
      return next
    })
  }

  const copyPath = (path: string) => {
    navigator.clipboard.writeText(path)
    setCopied(path)
    setTimeout(() => setCopied(null), 1500)
  }

  return (
    <div className="p-6 space-y-6 max-w-4xl">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <Book size={24} className="text-primary-600" />
          API Reference
        </h1>
        <p className="text-sm text-gray-500 mt-1">
          Complete REST API documentation. Base URL: <code className="bg-gray-100 px-1.5 py-0.5 rounded text-xs font-mono">https://api.talkex.business/v1</code>
        </p>
        <p className="text-sm text-gray-500 mt-1">
          Authenticate with <code className="bg-gray-100 px-1.5 py-0.5 rounded text-xs font-mono">Authorization: Bearer {'<token>'}</code> (JWT or API key).
        </p>
      </div>

      <div className="space-y-3">
        {API_SECTIONS.map((section, i) => (
          <div key={section.title} className="rounded-xl border border-gray-200 bg-white overflow-hidden">
            <button
              onClick={() => toggle(i)}
              className="w-full flex items-center justify-between px-5 py-3 hover:bg-gray-50 transition-colors text-left"
            >
              <div className="flex items-center gap-3">
                {expanded.has(i) ? <ChevronDown size={16} className="text-gray-400" /> : <ChevronRight size={16} className="text-gray-400" />}
                <span className="text-sm font-semibold text-gray-900">{section.title}</span>
                <span className="text-xs text-gray-400">{section.endpoints.length} endpoints</span>
              </div>
            </button>

            {expanded.has(i) && (
              <div className="border-t border-gray-100 divide-y divide-gray-50">
                {section.endpoints.map((ep) => (
                  <div key={ep.method + ep.path} className="px-5 py-3">
                    <div className="flex items-start gap-3">
                      <span className={`inline-flex items-center rounded px-2 py-0.5 text-[10px] font-bold tracking-wider ${METHOD_COLORS[ep.method]}`}>
                        {ep.method}
                      </span>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <code className="text-sm font-mono text-gray-800">{ep.path}</code>
                          <button
                            onClick={() => copyPath(ep.path)}
                            className="text-gray-300 hover:text-gray-500 transition-colors"
                            title="Copy path"
                          >
                            {copied === ep.path ? <CheckCircle2 size={12} className="text-green-500" /> : <Copy size={12} />}
                          </button>
                          {!ep.auth && (
                            <span className="text-[10px] bg-gray-100 text-gray-500 rounded px-1.5 py-0.5">Public</span>
                          )}
                        </div>
                        <p className="text-xs text-gray-500 mt-0.5">{ep.description}</p>
                        {ep.body && (
                          <pre className="mt-2 text-[11px] text-gray-600 bg-gray-50 rounded-lg px-3 py-2 overflow-x-auto font-mono">
                            {ep.body}
                          </pre>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
