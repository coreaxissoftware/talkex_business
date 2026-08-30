# TalkEx Business — Ops Runbook

The five things that page you at night, and what to do about each.

> **Rule zero — before touching anything:** open [`grafana.business.talkex.in`](#) in one tab and [`status.business.talkex.in`](#) in another. If Grafana disagrees with the alert, trust Grafana.

---

## Alert index

| # | Alert | First check | Then | Escalate |
|--|--|--|--|--|
| 1 | **API 5xx > 2% for 5 min** | `/metrics` → `http_requests_5xx_total` per pod | Roll back last deploy | @satya |
| 2 | **Send P95 > 500ms for 5 min** | Postgres CPU + wallet-locks | Scale Fly machines from 2 → 4 | @satya |
| 3 | **Queue backlog > 5,000** | `SELECT COUNT(*) FROM queued_messages WHERE status='queued'` | Increase worker batch size (5s → 2s cadence) | @satya |
| 4 | **DLQ growing > 100/hr** | Which channel? `GROUP BY channel` on `dead_letters` | Provider outage — post to status page | @aditi |
| 5 | **Wallet double-credit** | Audit log → search `wallet.credit` in last hour | Immediate: freeze that owner. Then: audit `payments.MarkPaid` for the paymentID | @satya |

---

## 1 · API 5xx spike

**Symptoms.** `http_requests_5xx_total` per pod rising, users report "Internal server error" toast.

**Diagnose.**
1. Grafana → API dashboard → check which endpoint. If concentrated on one route, look at the last deploy touching that handler.
2. Check `journalctl -u talkex-api` or `flyctl logs` for panics — recovery middleware logs them.
3. Postgres slow-query log → any query > 1s in the window? Missing index or a full-table scan is the usual culprit.

**Mitigate.**
- Rollback via `flyctl releases` → `flyctl deploy --image <previous>`.
- If rollback isn't safe (schema migration ran), scale up machines to absorb load while a real fix ships.

---

## 2 · Send latency P95 > 500ms

**Symptoms.** Agents report the send button "spinning". `enqueue_latency_ms` in Prometheus climbing.

**Likely causes.**
- Wallet row lock contention (many concurrent sends from one owner). Fix: shard wallet updates or introduce an in-memory reservation before the DB write.
- Postgres CPU pegged. Grow the instance, or turn on read-replicas for analytics reads.
- Redis latency spike (rate-limiter path). Check `redis-cli --latency` from an API pod.

**Mitigate.**
- Scale Fly machines: `flyctl scale count 4`.
- If Postgres is the bottleneck: `flyctl postgres update --vm-size shared-cpu-4x`.

---

## 3 · Queue backlog growing

**Symptoms.** Analytics dashboard shows enqueued > dispatched. Customers complain messages haven't arrived.

**Diagnose.**
```sql
SELECT status, COUNT(*), MIN(created_at)
FROM queued_messages
WHERE created_at > NOW() - INTERVAL '15 min'
GROUP BY status;
```

**Mitigate.**
- Bump worker cadence from 5s → 1s in `main.go` (needs a redeploy).
- Increase worker batch size from 50 → 200.
- If a specific channel is backed up (WhatsApp API returning 429), throttle that channel's send rate at the connector level.

---

## 4 · DLQ growing

**Symptoms.** `dead_letters` table growing > 100 rows/hr.

**Diagnose.**
```sql
SELECT channel, error, COUNT(*)
FROM dead_letters
WHERE created_at > NOW() - INTERVAL '1 hr'
GROUP BY channel, error
ORDER BY COUNT(*) DESC;
```

Common causes: WhatsApp Cloud API outage, Meta rate limit, expired provider token, DLT re-verification required for SMS.

**Mitigate.**
- Post a "degraded — <channel>" incident on the status page.
- Retry the DLQ in a controlled batch: `POST /messaging/dlq/{id}/retry` per row.
- If the provider is genuinely down, communicate an ETA and pause affected campaigns.

---

## 5 · Wallet double-credit

**Symptoms.** Two `wallet.credit` audit rows for the same `idempotency_key` in the same second.

**Diagnose.**
```sql
SELECT owner_id, reference, COUNT(*)
FROM wallet_transactions
WHERE created_at > NOW() - INTERVAL '1 hr'
GROUP BY owner_id, reference
HAVING COUNT(*) > 1;
```

**Immediate action.**
1. Freeze the owner's wallet: `UPDATE wallets SET frozen_at = NOW() WHERE owner_id = '...'`.
2. Take Postgres to isolation-level snapshot on the wallet writes if the pattern repeats.
3. File a critical bug — `internal/payments/service.go` MarkPaid should be idempotent under the payment_id key.

---

## Deploy checklist

Before every prod deploy:
1. `go test ./...` green.
2. `npm run build` succeeds (production Vite build).
3. `go vet ./...` clean.
4. Look at the diff — any migration? If yes, run it in a maintenance window.
5. Publish a changelog entry (even if internal).

---

## Contact ladder

| Severity | Response | Escalate to |
|--|--|--|
| **P0** — customer data at risk, > 25% API 5xx, wallet corruption | 15 min | @satya (+91 984xx) → @aditi (+91 987xx) |
| **P1** — degraded UX for > 10% users, DLQ stuck | 1 hr | @satya |
| **P2** — cosmetic, single-tenant impact | Next business day | GitHub issue |
