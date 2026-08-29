# TalkEx Business

**One Platform. Every Messaging Channel.**

Enterprise CPaaS (Communication Platform as a Service) dashboard by
CoreAxis Ventures. Full-featured competitor to Gupshup / Wati /
Interakt — WhatsApp Cloud API, Telegram, Email/SMTP, SMS, RCS,
Instagram DM, Facebook Messenger, TalkEx, and an embeddable Live
Chat widget, all landing in one inbox.

See [`CONTEXT.md`](CONTEXT.md) for the product/architecture narrative.

---

## Feature Matrix

| Area | Highlights |
|---|---|
| **Auth** | JWT access + refresh rotation, 2FA (TOTP), device sessions, OTP for phone/email, OAuth (Google/Facebook/GitHub/Apple with CSRF-safe state cookie), password reset |
| **Channels** | 8 connectors (talkex, whatsapp, telegram, sms, email, rcs, instagram, messenger) + Live Chat website widget; provider-agnostic messaging engine with queue, DLQ, retry, fallback routing |
| **Contacts** | CRUD, custom fields, tags, contact lists / segments, CSV import + export, opt-in tracking, duplicate merge |
| **Templates** | Per-channel bodies with `{{1}}..{{N}}` variable substitution, live preview; WhatsApp interactive templates (quick-reply / URL / phone buttons, list rows, header + footer, media attachments) and one-click Meta approval submission |
| **Campaigns** | Wizard-based creation, per-contact + list targeting, scheduler, real-time status, clone, edit drafts, CSV export, maker-checker approval above threshold, per-campaign analytics, auto-pause on low wallet |
| **Conversations** | Unified inbox, 24-hour session-window enforcement, labels, agent assignment, canned replies (`/` picker), CSAT rating, AI Assist (Claude — reply suggestion / summary / sentiment), full-text search, bulk assign / mark-read, live SSE updates |
| **Chatbot Flows** | Visual multi-step flow builder — 7 step types (send_message, send_template, wait, branch, assign_agent, add_tag, end), keyword triggers, per-run state, background sweeper for wait steps |
| **Automation** | Simple keyword auto-reply rules that pre-date the flow builder; still useful for one-message replies |
| **Business Hours + SLA** | Per-day open/close hours with tz-aware evaluation and overnight windows; auto away-message with 24h dedup; SLA breach notifications + `sla.breached` webhook |
| **Live Chat Widget** | Single `<script>` tag embed, per-owner branding + rotate-able public key, visitor sessions flow into the same Conversations inbox |
| **AI Assist** | Backend uses the official `anthropic-sdk-go`; dev-mode falls back to keyword heuristics so the UI works without an API key. Optional AI auto-tag stamps `sentiment:*` labels on every inbound |
| **Wallet & Payments** | Prepaid ledger with row-level locking + idempotency keys; Razorpay integration (dev-mode simulation + production HMAC signature verification) |
| **Multi-tenancy** | Organizations with parent/child hierarchy for reseller model, seat limits, per-member role scoping (owner/admin/member/agent/viewer) |
| **Compliance** | DPDP Act 2023 consent records, DSAR request workflow (process/complete/reject), processing records log |
| **Quality Monitoring** | Rolling-window quality tier (Green/Yellow/Red), auto-alerts on Yellow/Red, per-channel health |
| **Team** | Invite by email, role assignment, activity dashboard (per-agent open assignments, messages sent 30d, avg CSAT) |
| **Developer Portal** | API key management, in-app API playground, `/api-docs` reference, OpenAPI 3.0 spec (`openapi.yaml`), rate limiting, sandbox mode |
| **Real-time** | Server-Sent Events for inbox + notifications; reconnects with token refresh; also a public visitor stream for the widget |
| **Observability** | Prometheus `/metrics` endpoint (HTTP counters + goroutines/mem); structured error handling |
| **Docs / UX** | Command palette (`⌘K`), mobile-responsive sidebar, dark-mode CSS prep, in-app API docs, notification preferences, email digest toggle |

---

## Stack

**Backend** — Go 1.26 · Gin · GORM · SQLite (dev, pure-Go, no CGO) / Postgres (prod) · JWT access+refresh rotation · bcrypt password hashing · official `anthropic-sdk-go`. Compiles to a single static binary; ships in a ~15 MB distroless container.

**Frontend** — React 19 · Vite · TypeScript · Tailwind CSS v4 · React Router v7 · Zustand · Axios · Lucide icons · dependency-free EventSource client.

---

## Quick Start

### Option 1 — Docker Compose (recommended)

```bash
docker compose up
```

Brings up Postgres 16, Redis 7, the Go API on `:8080`, and the Vite dev server on `:5173` in one shot. Any provider env var you set locally passes through (`ANTHROPIC_API_KEY`, `RAZORPAY_KEY_ID`, etc.); everything falls back to dev-mode simulation when unset.

### Option 2 — Bare-metal (development)

**Backend:**
```bash
cp .env.example .env    # edit JWT_SECRET for anything non-dev
go run ./cmd/server     # starts on :8080
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev             # starts on :5173
```

Register through the UI, or via API:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@talkex.dev","password":"Test1234!","full_name":"Admin"}'
```

---

## Environment Variables

All are optional in dev — the platform runs end-to-end with in-memory dev-mode simulations when nothing is set.

| Variable | Purpose | Fallback |
|---|---|---|
| `DATABASE_URL` | Postgres / SQLite DSN | `sqlite://talkex_business.db` |
| `JWT_SECRET` | Access/refresh token signing key | `changeme-…` (rejected in prod) |
| `JWT_ACCESS_MINUTES` / `JWT_REFRESH_DAYS` | Token lifetimes | 15 / 30 |
| `PORT` / `ENVIRONMENT` / `CORS_ORIGINS` | Server config | `8080` / `development` / `http://localhost:5173` |
| `BASE_URL` / `FRONTEND_URL` | OAuth redirect base URLs | `http://localhost:<PORT>` / `http://localhost:5173` |
| `ANTHROPIC_API_KEY` / `ANTHROPIC_MODEL` | AI assist provider | dev heuristics / `claude-opus-5` |
| `OAUTH_{GOOGLE,FACEBOOK,GITHUB,APPLE}_CLIENT_ID` + `_SECRET` | Real OAuth flow | dev sim login |
| `RAZORPAY_KEY_ID` / `RAZORPAY_SECRET` / `RAZORPAY_WEBHOOK_SECRET` | Payment gateway | dev-simulate endpoint |
| `META_WHATSAPP_TOKEN` / `META_WHATSAPP_WABA_ID` | WhatsApp Cloud API + template submission | logs payload only |

---

## Ops

- **`GET /health`** — liveness (unauthenticated)
- **`GET /metrics`** — Prometheus scrape (HTTP counters, goroutines, memory)
- **`GET /api-docs`** (in-app) — interactive endpoint reference
- **`openapi.yaml`** — OpenAPI 3.0 spec for external SDK generators
- **Backup CLI** — dump a single tenant's data to JSON for support handoff or DPDP DSAR fulfillment:
  ```bash
  go run ./cmd/backup -owner <user-id> -out backup.json
  ```

## CI

`.github/workflows/ci.yml` runs on every push/PR:

1. **Backend** — `go build`, `go vet`, `go test ./...`
2. **Frontend** — `npm ci`, `tsc --noEmit`, `npm run build`
3. **Docker** — builds both API and web images to verify Dockerfiles

---

## Layout

```
cmd/
  server/               ← API entrypoint; wires all routes + cross-module hooks
  backup/               ← Per-tenant JSON export CLI
internal/
  config/               ← env/.env loading, JWT-secret guard, provider creds
  database/             ← GORM connection, base model (UUID + timestamps)
  middleware/           ← CORS, panic recovery, rate limit, idempotency
  auth/                 ← JWT + API-key middleware, OAuth (state cookie), RBAC
  users/                ← register/login/refresh/2FA/sessions/password reset
  otp/                  ← in-memory OTP send/verify with per-IP rate limit
  wallet/               ← balance + ledger, row-locked, idempotency-guarded
  payments/             ← Razorpay orders + HMAC-verified webhook
  contacts/             ← CRUD, custom fields, merge
  contactlists/         ← segments + members
  customfields/         ← per-owner definitions
  media/                ← uploads + signed-URL serving
  templates/            ← per-channel body + variables + WA interactive/media
  campaigns/            ← bulk-send state machine + runner + scheduler
  conversations/        ← inbox, thread, search, bulk actions, labels
  automation/           ← keyword auto-reply rules
  flows/                ← chatbot flow builder + execution engine + sweeper
  ai/                   ← Claude wrapper (SuggestReply/Summarize/Sentiment)
  canned/               ← per-owner canned reply library
  csat/                 ← rating submission + dashboard summary
  widget/               ← live-chat widget config + public API + snippet.js
  channels/             ← config + connector interface + per-provider adapters
    talkex/whatsapp/telegram/email/sms/rcs/instagram/messenger/sandbox
  messaging/            ← queue + worker + DLQ + retry + cost stamping
  webhooks/             ← outbound event delivery + delivery log + retry
  events/               ← in-process pub/sub + SSE stream
  notifications/        ← in-app notification model + emit hook
  audit/                ← request audit middleware + query API
  analytics/            ← per-owner aggregate stats + CSV export
  billing/              ← plan catalog + subscription
  compliance/           ← DPDP consent + DSAR + processing records
  organizations/        ← multi-tenant orgs + members
  team/                 ← team invites + roles + activity aggregate
  quality/              ← rolling-window quality tier + alerts
  settings/             ← per-owner prefs (notif, cost/sell, business hours, SLA, AI)
  businesshours/        ← tz-aware open/close evaluator
  sla/                  ← breach sweeper (notification + webhook)
  metrics/              ← Prometheus text-format /metrics
  customers/            ← business profile (KYC-adjacent)
  developers/           ← API keys + playground
  support/              ← help center tickets
  tags/                 ← tag aggregate view
frontend/
  src/
    services/           ← axios wrappers per domain
    pages/              ← one file per screen
    components/         ← shared UI (Modal, NotificationBell, TemplatePreview, AiPanel, CannedPicker, ...)
    layouts/            ← DashboardLayout + AuthLayout + Sidebar + Header
    store/              ← Zustand stores (authStore)
    router.tsx          ← react-router route table
```

---

## Testing

```bash
go test ./...                       # backend unit tests
cd frontend && npx tsc --noEmit     # frontend type check
```

Test coverage is intentionally light — the focus so far has been feature breadth. Priority packages that already have tests: `businesshours`, `metrics`.

---

## License

Proprietary — CoreAxis Ventures Pvt. Ltd.
