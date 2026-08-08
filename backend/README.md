# TalkEx Business — Backend

Enterprise CPaaS dashboard API. See [`../CONTEXT.md`](../CONTEXT.md) for the
full product/architecture context.

## Stack
FastAPI + SQLAlchemy 2.0 (async) + Alembic. SQLite for local dev, Postgres
for production, switched via one `DATABASE_URL` env var.

## Setup
```bash
cd backend
python -m venv venv
venv\Scripts\activate        # Windows
pip install -r requirements.txt
copy .env.example .env       # then edit secrets
alembic revision --autogenerate -m "init"
alembic upgrade head
uvicorn app.main:app --reload
```

Docs at `http://localhost:8000/docs` once running.

## Layout
Each domain is a self-contained package: `models.py` (SQLAlchemy),
`schemas.py` (Pydantic), `service.py` (business logic), `router.py`
(FastAPI endpoints). Wired end-to-end: `auth`, `users`, `wallet`,
`contacts`, `templates`. Everything else under `app/` is a stub package
with a docstring pointing at its CONTEXT.md phase — fill in the same
four-file pattern as each module comes into scope.

## MVP-blocking mechanics already modeled
- Wallet transactions carry an `idempotency_key` (unique per wallet) so a
  retried debit/credit can't double-apply.
- `MessageTemplate.category` (Marketing/Utility/Authentication) exists now
  so Billing can wire per-category pricing without a later migration.
- `Contact.opted_in` / `opted_in_at` / `last_inbound_at` model the
  consent gate and 24h customer-service-window inputs.
- `User.quality_flagged_at` / `is_business_verified` mirror the consumer
  app's quality-rating and business-verification fields.

Still open (see CONTEXT.md "Other flagged gaps"): DLQ+retry, priority
queues, Redis, API gateway, low-wallet auto-pause, multi-tenancy.
