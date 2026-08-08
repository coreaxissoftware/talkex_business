# TalkEx Business Platform — Context for a fresh session

## Core Purpose
An enterprise CPaaS (Communication Platform as a Service) dashboard — like Gupshup/Wati/Interakt — where businesses send messages through pluggable **channel connectors**. TalkEx itself is the primary/first-party channel; WhatsApp Business API is the second. Telegram, RCS, Email, Instagram, Facebook Messenger are planned as future connectors, added without redesigning the UI (channel-based architecture, not product-based).

- Portal URL: `https://devtalkex.coreaxis.cloud`
- Brand: TalkEx Business by CoreAxis Ventures
- Tagline: "One Platform. Multiple Messaging Channels."
- Philosophy: One dashboard, multiple channels — Contacts/Templates/Campaigns/Analytics/Webhooks/API Keys are SHARED across channels; only the Channel Connector differs.

## Relationship to the existing consumer TalkEx app
There's a separate, already-live consumer messaging app (WhatsApp/Telegram-style: chat, calls, E2EE, status) at `D:\WebProject\APP\hypertalk` — FastAPI + SQLite backend, React+Vite frontend, deployed to Render (backend) + Hostinger (frontend), also packaged as an Android app via Capacitor.

The **TalkEx channel connector** in this Business Platform should build on top of that consumer app's backend — treating it as the "TalkEx Business API" the same way WhatsApp Business API is a channel. That backend already has some of the needed pieces:
- `api_keys` table — external/bulk-sending credential, separate from session auth (the WhatsApp-Business-API equivalent)
- `POST /api/v1/messages` — bulk send by username
- `business_optins` table + `POST/DELETE /business/optin/{username}` + `GET /business/optins` — recipient consent gate (mirrors WhatsApp's opt-in)
- `has_open_conversation_window()` — 24-hour customer-service-window equivalent (recipient messaged first → free-form reply allowed without opt-in)
- `quality_flagged_at` + `refresh_quality_flag()` — auto-suspends a sender after 5+ blocks/reports in 7 days (mirrors WhatsApp quality rating dropping to Red)
- `users.is_business` / `business_category` + `POST /admin/users/{id}/verify-business` — Business Verification badge (admin-granted, not self-service)
- `webhooks` table — outbound HMAC-signed webhook on incoming message

## Platform Layout
```
TalkEx Business
├── Dashboard        (wallet balance, plan, message stats, charts)
├── Channels          (TalkEx Business, WhatsApp Business — unified)
├── Contacts          (shared across channels: import, tags, segments, custom fields)
├── Templates         (shared engine, per-channel mapping, categories: Marketing/Utility/Auth)
├── Campaigns         (9-step builder: name→channel→list→template→personalize→media→schedule→review→launch)
├── Conversations     (shared inbox, live chat, agent assignment, labels)
├── Automation        (no-code flow builder: triggers → actions)
├── Developers        (API keys, webhooks, playground, SDKs, logs, rate limits)
├── Analytics         (unified, filterable by channel/campaign/date/template)
├── Billing           (plan, invoices, GST, payment history)
├── Wallet            (balance, recharge, transactions, refunds, coupons)
├── Support           (tickets, live chat, docs)
└── Settings          (profile, security, API, webhooks, 2FA, sessions, danger zone)
```

## Database Domains
Users, Roles & Permissions, Customers, Plans, Wallet, Payments, Channels, Business Accounts, Contacts, Contact Lists, Tags, Templates, Campaigns, Messages, Conversations, Media, API Keys, Webhooks, Automations, Notifications, Audit Logs, Analytics.

## Backend module tree (target)
```
app/
├── auth  ├── users  ├── customers  ├── plans  ├── wallet  ├── billing
├── channels/ {talkex, whatsapp, shared}
├── contacts  ├── templates  ├── campaigns  ├── messaging  ├── conversations
├── automation  ├── developers  ├── analytics  ├── notifications  ├── media
├── webhooks  ├── audit  ├── settings  ├── middleware  ├── core  ├── database  └── utils
```

## Frontend module tree (target)
```
src/
├── app  ├── layouts  ├── components
├── features/ {dashboard, channels, contacts, templates, campaigns, conversations,
│              automation, developers, analytics, billing, wallet, support, settings}
├── services  ├── store  ├── hooks  ├── types  └── utils
```

## Message flow
Campaign → Message Queue → Channel Router → (TalkEx API | WhatsApp API) → Delivery Engine → Webhook Engine → Analytics Engine → Dashboard.

## Critical WhatsApp-channel mechanics that MUST be modeled (non-negotiable, not just nice-to-have)
1. **24-hour customer service window** — free-form reply only within 24h of the customer's last message; otherwise an approved template is required.
2. **Messaging tiers + quality rating** (Green/Yellow/Red) — determines how many business-initiated conversations/24h are allowed; a bad rating can get a number banned.
3. **Template categories (Marketing/Utility/Authentication)** — drives Meta's 2025 per-template pricing model; the Templates module needs a category field wired into Billing.
4. **Business Verification / Embedded Signup** — Facebook Business Manager verification → WABA creation → phone number registration → display name review. This is a whole onboarding wizard, not a one-line "Setup" screen.
5. **Consent/opt-in tracking** — legally required and protects quality rating; unsolicited sends tank it fast.

## Other flagged gaps (prioritized Phase 2/3, not blocking MVP)
Idempotency keys on send/wallet-debit, Dead Letter Queue + retry policy, priority queues (OTP > transactional > marketing), Redis for caching/rate-limiting, API Gateway (Kong/Traefik) beyond Nginx, DB read replicas/sharding, fallback channel routing, number health/ban-risk alerts, low-wallet auto-pause, sandbox/test mode for developers, maker-checker approval for large campaigns, multi-tenancy/reseller hierarchy, cost-vs-margin tracking, DPDP Act 2023 compliance (India), TRAI DLT registration (only if/when SMS is added).

## Suggested build phasing
- **MVP**: 24h window logic, template categories, consent tracking, idempotency, DLQ+retry, quality-rating dashboard
- **Phase 2**: onboarding wizard, Redis, API gateway, fallback routing, low-wallet auto-pause
- **Phase 3**: multi-tenancy/reseller, maker-checker, sandbox mode, margin tracking, DPDP module
- **Phase 4 (scale)**: DB sharding/replicas, Telegram/RCS/Email connectors

## Tech stack (not yet decided in the new session — recommend matching what CoreAxis already knows)
- Backend: Python + FastAPI (team's existing stack), SQLAlchemy + Alembic (proper ORM/migrations — this schema is much bigger than the consumer app's raw-SQL style warrants), SQLite for local dev / Postgres for production via one DATABASE_URL env var (enterprise scale + concurrent writes for campaigns/wallet/billing need real transactional guarantees).
- Auth: real JWT access + refresh token rotation (unlike the consumer app's opaque bearer tokens) — this platform needs interoperable auth for a Developer Portal with multiple language SDKs.
- Frontend: React + Vite (team's existing stack).
