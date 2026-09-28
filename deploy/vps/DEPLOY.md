# DEPLOY — the shortest possible path from zero to live

Everything else in this folder is the mechanics. This file is what you
actually run.

## Pre-flight validated on my end

- `bash -n setup.sh update.sh backup.sh` — no syntax errors
- `go build ./...` clean; `go vet ./...` clean
- `go test ./...` — 10 packages pass
- `cmd/migrate -status` — smoke-tested locally against SQLite; idempotent (re-run is a no-op)
- `cmd/server` — smoke-tested locally; `/health` returns 200 within 2s
- `frontend` `tsc` + `npm run build` — clean, 459 KB gz 145 KB
- `setup.sh` — additive-only UFW (does NOT wipe other apps' rules)
- `setup.sh` — per-domain certbot (one domain failing does not sink the others)

## Step 1 — DNS

Point three A records at `145.223.21.70` and wait for them to propagate:

```
business.talkex.in         A   145.223.21.70
businessapp.talkex.in      A   145.223.21.70
businessapi.talkex.in      A   145.223.21.70
```

Verify from any laptop:

```bash
for d in business.talkex.in businessapp.talkex.in businessapi.talkex.in; do
  echo "$d -> $(dig +short $d)"
done
```

All three should print `145.223.21.70`. If any is empty, wait a few minutes.

## Step 2 — SSH into the server and run the bootstrap

```bash
ssh root@145.223.21.70

# One line — clones the repo, installs everything, requests certs, starts the API
bash <(curl -fsSL https://raw.githubusercontent.com/coreaxissoftware/talkex_business/main/deploy/vps/setup.sh)
```

You will see:

```
==> Installing system packages
==> Creating system user 'talkex'
==> Cloning https://github.com/coreaxissoftware/talkex_business.git
==> Setting up Postgres
==> Writing /etc/talkex-business.env with generated JWT secret
==> Building API binary
==> Building Vite dashboard
==> Copying marketing site
==> Installing systemd unit
==> Running DB migrations
==> Writing nginx vhosts
==> Requesting Let's Encrypt certs (per-domain — one failure doesn't sink the others)
  cert issued for business.talkex.in
  cert issued for businessapp.talkex.in
  cert issued for businessapi.talkex.in
==> Adding UFW rules (additive, non-destructive)
==> Bootstrap complete

Live services:
  https://business.talkex.in
  https://businessapp.talkex.in
  https://businessapi.talkex.in  (try /health)
```

Total wall time: ~5–8 minutes on a first-gen VPS.

## Step 3 — verify it works

From your laptop:

```bash
curl -sS https://businessapi.talkex.in/health
# {"environment":"production","status":"ok"}

curl -I https://business.talkex.in       # marketing site — HTTP/2 200
curl -I https://businessapp.talkex.in    # dashboard — HTTP/2 200
```

## Step 4 — fill in provider secrets

```bash
nano /etc/talkex-business.env
```

Uncomment and fill the providers you actually use:

```env
ANTHROPIC_API_KEY=sk-ant-...
MAILGUN_DOMAIN=mail.talkex.in
MAILGUN_API_KEY=...
MSG91_AUTH_KEY=...
MSG91_TEMPLATE_ID=...
MSG91_SENDER_ID=COREVN
FAST2SMS_API_KEY=...
RAZORPAY_KEY_ID=rzp_live_...
RAZORPAY_SECRET=...
RAZORPAY_WEBHOOK_SECRET=...
META_WHATSAPP_TOKEN=...
META_WHATSAPP_WABA_ID=...
OAUTH_GOOGLE_CLIENT_ID=...
OAUTH_GOOGLE_SECRET=...
SENTRY_DSN=...
```

Then reload:

```bash
systemctl restart talkex-api
journalctl -u talkex-api -n 30 --no-pager
```

Look for `TalkEx Business API starting on :8080 (env=production)`.

## Step 5 — first user

Register from the browser at <https://businessapp.talkex.in/register>,
or from curl:

```bash
curl -X POST https://businessapi.talkex.in/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@talkex.in","password":"YourReallyStrongPass!23","full_name":"Owner"}'
```

## Everything after this: `update.sh`

For every future code push:

```bash
ssh root@145.223.21.70 'cd /opt/talkex-business && bash deploy/vps/update.sh'
```

It git-tags the commit, runs pending migrations, rebuilds everything,
restarts the API, health-checks, and exits non-zero if `/health`
doesn't come back within 30s.

## If something goes wrong

| Symptom | First thing to try |
|---|---|
| `curl /health` times out | `journalctl -u talkex-api -n 100` |
| certbot fails on one domain | Verify DNS with `dig`, then re-run `setup.sh` — it skips already-issued certs |
| 502 from nginx | `systemctl status talkex-api`; probably the API crashed on boot |
| API crashes on start | `systemctl status talkex-api` shows why; usually a missing var in `/etc/talkex-business.env` |
| Postgres connection refused | `systemctl status postgresql`; then check `DATABASE_URL` line in the env file |
| DNS still wrong | Re-run `setup.sh` after DNS lands — idempotent, will just fill the gaps |

## Rollback

```bash
cd /opt/talkex-business
git tag | tail -5      # see recent deploy tags
git checkout deploy-2026-09-28-1430
bash deploy/vps/update.sh
```

## Where things live

| Path | What |
|---|---|
| `/opt/talkex-business/` | source checkout (git repo) |
| `/usr/local/bin/talkex-api` | Go API binary |
| `/usr/local/bin/talkex-migrate` | migration CLI |
| `/usr/local/bin/talkex-backup` | nightly backup script |
| `/etc/talkex-business.env` | secrets — root:root 0600 |
| `/etc/systemd/system/talkex-api.service` | systemd unit |
| `/etc/nginx/sites-available/*.talkex.in.conf` | vhosts |
| `/var/www/business.talkex.in/` | marketing static |
| `/var/www/businessapp.talkex.in/` | Vite dashboard |
| `/var/log/talkex/` | app logs (journald has the API logs) |
| `/var/backups/talkex/` | nightly pg_dumps + uploads snapshots |
| `/root/.talkex-db-password` | Postgres password (root-only) |
