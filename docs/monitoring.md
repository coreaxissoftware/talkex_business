# Monitoring & alerting

The stack is deliberately boring — three components, all open standards.

## What we already emit

- **`GET /metrics`** — Prometheus text-format (`internal/metrics`), includes:
  - `http_requests_total` + `_2xx_total` / `_4xx_total` / `_5xx_total`
  - `go_goroutines`, `go_memstats_alloc_bytes`, `go_memstats_sys_bytes`
- **`GET /health`** — cheap liveness (no DB roundtrip)
- **`audit_logs` table** — every state-changing HTTP call (actor, route, status, latency)

## Scraping

Point any Prometheus-compatible scraper at `https://api.talkex.io/metrics`. On Fly.io the [metrics] block in `fly.toml` already registers the path, so the built-in Fly metrics dashboard picks it up automatically.

For a proper long-term store, forward to **Grafana Cloud** (free 10k series tier is enough):

```yaml
# prometheus.yml scrape config
scrape_configs:
  - job_name: talkex-api
    scrape_interval: 15s
    metrics_path: /metrics
    scheme: https
    static_configs:
      - targets: ['api.talkex.io']
```

## Grafana dashboards

Import these three JSON dashboards from `docs/grafana/*.json` (create alongside):

1. **API health** — request rate, 2xx/4xx/5xx breakdown, P50/P95/P99 latency
2. **Messaging engine** — queue depth, DLQ growth, per-channel send throughput
3. **Business** — active tenants, MRR, wallet burn rate, per-owner top-10

## Sentry (or GlitchTip, self-hosted)

The app currently uses `middleware.Recovery()` to catch panics, but the error goes to stderr only. Wire Sentry with 3 lines in `main.go`:

```go
import "github.com/getsentry/sentry-go"
import sentrygin "github.com/getsentry/sentry-go/gin"

sentry.Init(sentry.ClientOptions{
    Dsn:              os.Getenv("SENTRY_DSN"),
    Environment:      cfg.Environment,
    TracesSampleRate: 0.05, // 5% of requests get a trace
})
r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
```

Add `SENTRY_DSN` as a Fly secret. The recovery middleware still runs — Sentry sees the panic, then the middleware returns 500 to the client.

## Uptime checks

- **status.talkex.io** — self-hosted [Uptime Kuma](https://github.com/louislam/uptime-kuma) container hitting `/health` from three regions every 60s
- **PagerDuty** integration on the Kuma status → phone-call after 3 consecutive failures

## Alert rules

Ship this Prometheus alert set (`docs/prometheus-alerts.yml`, create alongside):

```yaml
groups:
  - name: talkex-api
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_5xx_total[5m]) / rate(http_requests_total[5m]) > 0.02
        for: 5m
        labels: { severity: page }
        annotations:
          summary: "5xx rate > 2% for 5 min"
          runbook: "docs/RUNBOOK.md#1-api-5xx-spike"

      - alert: LatencyBudgetBurn
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels: { severity: page }
        annotations:
          summary: "P95 latency > 500ms"

      - alert: QueueBacklog
        expr: talkex_queue_depth > 5000
        for: 5m
        labels: { severity: warn }
        annotations:
          summary: "Messaging queue depth > 5k"

      - alert: GoroutineLeak
        expr: go_goroutines > 5000
        for: 15m
        labels: { severity: warn }
        annotations:
          summary: "Goroutine count > 5k for 15 min"
```

## What's not yet wired

- **Distributed tracing** (OpenTelemetry) — worthwhile once we're at 10+ services; single Go binary today doesn't need it
- **Log aggregation** — Fly.io retains 30 days by default; if you need longer, ship to Loki or Datadog
- **Real-user monitoring** on the marketing site — [Plausible](https://plausible.io) or [Umami](https://umami.is) both self-host in a container
