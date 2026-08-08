# TalkEx Business

**One Platform. Multiple Messaging Channels.**

Enterprise CPaaS dashboard API by CoreAxis Ventures. See
[`CONTEXT.md`](CONTEXT.md) for the full product/architecture context.

## Stack
Go 1.26 · Gin (HTTP framework) · GORM (ORM) · SQLite (dev) / Postgres
(prod) · JWT access+refresh rotation · bcrypt password hashing.

Compiles to a single static binary — no runtime, no interpreter, no
dependency install on deploy.

## Quick Start
```bash
cp .env.example .env   # edit JWT_SECRET
go run ./cmd/server     # starts on :8080
```

API at `http://localhost:8080/health` once running.

## Layout
```
cmd/server/main.go      ← entrypoint, wires all routes
internal/
  config/               ← env/.env loading
  database/             ← GORM connection, base model (UUID + timestamps)
  middleware/           ← CORS, panic recovery
  auth/                 ← JWT issuance/validation, auth middleware
  users/                ← model + handler + service (register/login/me)
  wallet/               ← model + handler + service (idempotency-guarded)
  contacts/             ← model + handler + service (opt-in, 24h window)
  templates/            ← model + handler + service (category field)
  channels/{talkex,whatsapp,shared}/  ← stub
  campaigns/            ← stub
  messaging/            ← stub
  conversations/        ← stub
  automation/           ← stub
  developers/           ← stub
  analytics/            ← stub
  billing/              ← stub
  ...
```

Each domain module has `model.go` (GORM model), `service.go` (business
logic), `handler.go` (Gin HTTP handlers). Stub packages have `doc.go`
with a comment noting their CONTEXT.md phase.

## API Endpoints
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /health | — | Health check |
| POST | /auth/register | — | Create account |
| POST | /auth/login | — | Get access+refresh tokens |
| POST | /auth/refresh | — | Rotate tokens |
| GET | /users/me | ✅ | Current user profile |
| GET | /wallet | ✅ | Wallet balance |
| GET | /wallet/transactions | ✅ | Transaction ledger |
| POST | /wallet/transactions | ✅ | Credit/debit (idempotent) |
| GET | /contacts | ✅ | List contacts |
| POST | /contacts | ✅ | Create contact |
| GET | /contacts/:id | ✅ | Get contact |
| PATCH | /contacts/:id | ✅ | Update contact |
| DELETE | /contacts/:id | ✅ | Delete contact |
| GET | /templates | ✅ | List templates |
| POST | /templates | ✅ | Create template |
| GET | /templates/:id | ✅ | Get template |
| PATCH | /templates/:id | ✅ | Update template |
| DELETE | /templates/:id | ✅ | Delete template |
