# TalkEx Business

**One Platform. Multiple Messaging Channels.**

Enterprise CPaaS (Communication Platform as a Service) dashboard by
CoreAxis Ventures. See [`CONTEXT.md`](CONTEXT.md) for the full
product/architecture context.

## Stack
**Backend** — Go 1.26 · Gin (HTTP framework) · GORM (ORM) · SQLite
(dev, pure-Go, no CGO) / Postgres (prod) · JWT access+refresh
rotation · bcrypt password hashing.

Compiles to a single static binary — no runtime, no interpreter, no
dependency install on deploy.

**Frontend** — React 19 · Vite · TypeScript · Tailwind CSS v4 ·
React Router v7 · Zustand · Axios · Lucide icons.

## Quick Start

### Backend
```bash
cp .env.example .env    # edit JWT_SECRET for anything non-dev
go run ./cmd/server     # starts on :8080
```
Health check: `http://localhost:8080/health`.

### Frontend
```bash
cd frontend
npm install
npm run dev             # starts on :5173
```
Open `http://localhost:5173`. The Vite dev server proxies `/api/*` to
the Go backend on `:8080`.

## Test credentials

Register a new account through the UI, or via API:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@talkexbusiness.dev","password":"Test1234!","full_name":"Test Admin"}'
```

## Layout
```
cmd/server/main.go        ← entrypoint, wires all routes + cross-module hooks
internal/
  config/                 ← env/.env loading, JWT-secret guard in prod
  database/               ← GORM connection, base model (UUID + timestamps)
  middleware/             ← CORS, panic recovery
  auth/                   ← JWT + API-key auth middleware
  users/                  ← register/login/refresh/me + password reset
  wallet/                 ← balance + ledger, idempotency-guarded, row-locked
  contacts/               ← CRUD (owner-scoped) + create hook
  templates/              ← CRUD, category-tagged (marketing/utility/auth)
  campaigns/              ← bulk-send state machine + real runner
  conversations/          ← 24h window, inbound/outbound hooks
  automation/             ← keyword→auto-reply rules
  developers/             ← API keys (hashed, reveal-once)
  webhooks/               ← HMAC-signed outbound event delivery
  channels/               ← enable/config for talkex/whatsapp/etc.
  billing/                ← plans + subscription + invoices
  support/                ← help tickets
  notifications/          ← in-app bell notifications
  analytics/              ← summary + timeseries (dialect-safe bucketing)
  audit/                  ← request-log middleware + /logs UI backend
  messaging/, media/, settings/,
  customers/, channels/{talkex,whatsapp,shared}/    ← reserved, not yet built
```

Each domain module has `model.go` (GORM model), `service.go` (business
logic), `handler.go` (Gin HTTP handlers).

## API Endpoints

Auth
| Method | Path | Description |
|---|---|---|
| POST | /auth/register | Create account |
| POST | /auth/login | Get access+refresh tokens |
| POST | /auth/refresh | Rotate tokens |
| POST | /auth/forgot-password | Request reset link (always 200) |
| POST | /auth/reset-password | Consume reset token |

Users & wallet (auth required)
| Method | Path | Description |
|---|---|---|
| GET / PATCH | /users/me | Current profile |
| POST | /users/me/change-password | Change password |
| GET | /wallet | Wallet balance |
| GET / POST | /wallet/transactions | Ledger + credit/debit (idempotent, locked) |

Contacts, templates, campaigns, conversations
| Method | Path | Description |
|---|---|---|
| GET/POST/PATCH/DELETE | /contacts, /contacts/:id | CRUD |
| GET/POST/PATCH/DELETE | /templates, /templates/:id | CRUD |
| GET/POST/DELETE | /campaigns, /campaigns/:id | List, create, delete |
| POST | /campaigns/:id/launch | Launch (fires background send) |
| POST | /campaigns/:id/cancel | Cancel a running campaign |
| GET | /conversations | Inbox (joined with contact) |
| GET | /conversations/:id/messages | Thread |
| POST | /conversations/:id/read | Mark all read |
| POST | /conversations/send | Send outbound |
| POST | /conversations/inbound | Simulate inbound (dev) |

Automation, developers, webhooks, channels
| Method | Path | Description |
|---|---|---|
| GET/POST/PATCH/DELETE | /automation/rules, /automation/rules/:id | Auto-reply rules |
| GET/POST | /api-keys | List / create (returns plaintext once) |
| POST/DELETE | /api-keys/:id/revoke, /api-keys/:id | Revoke, delete |
| GET/POST/DELETE | /webhooks, /webhooks/:id | Endpoint CRUD |
| GET | /webhooks/:id/deliveries | Delivery log |
| GET | /channels/catalog | Available channel kinds |
| GET | /channels | Owner's channel configs |
| PUT | /channels/:kind | Enable/disable/config a channel |

Billing, support, notifications, analytics, audit
| Method | Path | Description |
|---|---|---|
| GET | /billing/plans | Plan catalogue |
| GET/POST | /billing/subscription | Get / change current plan |
| GET | /billing/invoices | Invoice history |
| GET/POST | /support/tickets | List / create tickets |
| GET | /notifications | List notifications |
| GET | /notifications/unread-count | Badge count |
| POST | /notifications/:id/read, /notifications/read-all | Mark read |
| GET | /analytics/summary | KPIs (messages, delivery rate, contacts, etc.) |
| GET | /analytics/timeseries?days=N | Per-day inbound/outbound counts |
| GET | /audit-logs | Filterable request log |
| GET | /audit-logs/stats | Success rate summary |

## Development notes

- Every request goes through the audit middleware — every failed request
  is captured with method/path/status/latency and its response body
  (truncated to 2KB) so debugging doesn't require greping server logs.
- The wallet uses `SELECT … FOR UPDATE` + idempotency-key uniqueness so
  concurrent debits can't drive the balance negative.
- Campaign launch fans out sends in a background goroutine, incrementing
  `sent_count`/`failed_count` per recipient, then transitions to
  `completed`/`failed` and fires the `campaign.completed` webhook +
  in-app notification.
- Inbound messages trigger three side-effects: automation rules, an
  in-app notification, and the `inbound.message` webhook.
- API keys and webhook secrets are stored SHA-256 hashed; the plaintext
  is returned exactly once at creation.
- JWT `JWT_SECRET` must be set (>=32 chars) outside development;
  the process refuses to start with the default value in prod.

## What's NOT built yet
Real channel connectors (TalkEx first-party send, WhatsApp Cloud API),
CSV import for contacts, template variable rendering with substitution,
2FA / team member invites, invoice PDF + GST, payment-provider
integration for wallet top-up and plan upgrades. All of these are
called out in [`CONTEXT.md`](CONTEXT.md) Phase 3.
