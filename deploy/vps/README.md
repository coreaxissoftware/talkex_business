# TalkEx Business — VPS deployment (Ubuntu 26.04 LTS)

Complete production deployment on a single Ubuntu box. What ends up
running:

| Service | Where | Port |
|---|---|---|
| **Go API** (`talkex-business`) | systemd unit `talkex-api` | 127.0.0.1:8080 |
| **Postgres 18** | apt package | 127.0.0.1:5432 |
| **Redis 7** | apt package | 127.0.0.1:6379 |
| **nginx** | apt package, reverse proxy + TLS | 80, 443 |
| **Certbot** | Let's Encrypt for the three domains | — |

Public surface (behind nginx + Let's Encrypt):

| Domain | Serves |
|---|---|
| `business.talkex.in` | static marketing site (`marketing/index.html`) |
| `businessapp.talkex.in` | Vite dashboard build (`frontend/dist/`) |
| `businessapi.talkex.in` | proxies to the Go API on `127.0.0.1:8080` |

## Prerequisites

1. **Server**: Ubuntu 26.04 LTS at `145.223.21.70`, SSH as `root`.
2. **DNS**: three A records pointing at `145.223.21.70`:
   ```
   business.talkex.in         A   145.223.21.70
   businessapp.talkex.in     A   145.223.21.70
   businessapi.talkex.in     A   145.223.21.70
   ```
   Wait for propagation before running certbot — verify with:
   ```
   dig +short business.talkex.in
   dig +short businessapp.talkex.in
   dig +short businessapi.talkex.in
   ```

## First-time server bootstrap (one-shot)

From your laptop:

```bash
# 1. Copy the deploy folder to the server
scp -r deploy/vps root@145.223.21.70:/root/talkex-deploy

# 2. SSH in and run the bootstrap
ssh root@145.223.21.70
cd /root/talkex-deploy
bash setup.sh
```

`setup.sh` is **idempotent** — safe to re-run.

It will:

1. `apt update` + install: golang-go, nodejs (22.x), postgresql-18,
   redis-server, nginx, certbot, python3-certbot-nginx, git,
   build-essential, ufw
2. Create the `talkex` system user (no login)
3. Clone the repo to `/opt/talkex-business` (or pull latest)
4. Build the Go binary + Vite dashboard + copy marketing static
5. Install `talkex-api.service` and start it via systemd
6. Set up Postgres: create `talkex_business` DB + `talkex` role
7. Write `/etc/talkex-business.env` with secrets (JWT_SECRET auto-minted)
8. Write nginx server blocks for the three domains
9. Run certbot to grab TLS certs for all three
10. Enable UFW: 22, 80, 443 only

Total: **~5-8 minutes on a first-generation VPS** (mostly Go build + apt).

## Post-bootstrap: fill in real provider secrets

Edit `/etc/talkex-business.env` on the server and set the ones you use:

```
# Required for auth to actually work in prod (auto-set on first run)
JWT_SECRET=<64-hex-chars>

# Real Postgres URL — auto-set by setup.sh
DATABASE_URL=postgres://talkex:...@127.0.0.1:5432/talkex_business?sslmode=disable
REDIS_URL=redis://127.0.0.1:6379

# Environment
ENVIRONMENT=production
CORS_ORIGINS=https://businessapp.talkex.in,https://business.talkex.in
BASE_URL=https://businessapi.talkex.in
FRONTEND_URL=https://businessapp.talkex.in

# Providers (fill as you sign up)
ANTHROPIC_API_KEY=sk-ant-...
RAZORPAY_KEY_ID=rzp_live_...
RAZORPAY_SECRET=...
RAZORPAY_WEBHOOK_SECRET=...
MAILGUN_DOMAIN=mail.talkex.in
MAILGUN_API_KEY=...
MAILGUN_FROM=TalkEx <no-reply@mail.talkex.in>
MSG91_AUTH_KEY=...
MSG91_TEMPLATE_ID=...
MSG91_SENDER_ID=COREVN
FAST2SMS_API_KEY=...
FAST2SMS_SENDER_ID=COREVN
FAST2SMS_TEMPLATE_ID=...
META_WHATSAPP_TOKEN=...
META_WHATSAPP_WABA_ID=...
OAUTH_GOOGLE_CLIENT_ID=...
OAUTH_GOOGLE_SECRET=...
OAUTH_GITHUB_CLIENT_ID=...
OAUTH_GITHUB_SECRET=...
OAUTH_FACEBOOK_CLIENT_ID=...
OAUTH_FACEBOOK_SECRET=...
OAUTH_APPLE_CLIENT_ID=...
OAUTH_APPLE_TEAM_ID=...
OAUTH_APPLE_KEY_ID=...
OAUTH_APPLE_PRIVATE_KEY=...
SENTRY_DSN=...
```

Then `systemctl restart talkex-api`.

## Updating (every code push after that)

From your laptop:

```bash
ssh root@145.223.21.70 'cd /opt/talkex-business && bash deploy/vps/update.sh'
```

Or from the server directly:

```bash
cd /opt/talkex-business
bash deploy/vps/update.sh
```

`update.sh` will:

1. `git pull origin main`
2. `go run ./cmd/migrate` (runs pending schema migrations)
3. `go build -o /usr/local/bin/talkex-api ./cmd/server`
4. `cd frontend && npm ci && npm run build`
5. `cp -r frontend/dist /var/www/businessapp.talkex.in/`
6. `cp -r marketing /var/www/business.talkex.in/`
7. `systemctl restart talkex-api`
8. Health check — exits non-zero if API doesn't come back up in 30s

## Ops commands

```bash
# API logs (live tail, systemd journal)
journalctl -u talkex-api -f

# API logs (last 200 lines)
journalctl -u talkex-api -n 200

# Restart the API
systemctl restart talkex-api

# Nginx status + reload
systemctl status nginx
nginx -t && systemctl reload nginx

# Postgres CLI as the app role
sudo -u postgres psql talkex_business

# Certbot auto-renewal (already installed as a systemd timer)
systemctl list-timers | grep certbot
certbot renew --dry-run   # test the renewal path

# Backup — Postgres dump to /var/backups/talkex
bash /opt/talkex-business/deploy/vps/backup.sh
```

## Rollback

Every deploy tags git with `deploy-<UTC>`. To roll back:

```bash
cd /opt/talkex-business
git checkout deploy-2026-09-28-1430
bash deploy/vps/update.sh
```

## Health + monitoring

```bash
# Health
curl https://businessapi.talkex.in/health

# Prometheus metrics (from the server, not public)
curl http://127.0.0.1:8080/metrics
```

Point any Prometheus scraper at the server's private IP.

## Files in this folder

| File | Purpose |
|---|---|
| `README.md` | This document |
| `setup.sh` | First-time server bootstrap |
| `update.sh` | Pull + rebuild + restart |
| `backup.sh` | Nightly Postgres dump |
| `talkex-api.service` | systemd unit for the Go API |
| `nginx/business.talkex.in.conf` | Marketing site vhost |
| `nginx/businessapp.talkex.in.conf` | Dashboard vhost |
| `nginx/businessapi.talkex.in.conf` | API reverse proxy vhost |
