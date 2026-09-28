#!/usr/bin/env bash
#
# setup.sh — first-time TalkEx Business bootstrap on Ubuntu 26.04.
#
# Idempotent by design: every step checks state before mutating. Safe
# to re-run after a partial failure or when only some steps changed.
#
# Assumptions:
#   - Ubuntu 26.04 LTS, run as root
#   - DNS already points business.talkex.in / businessapp.talkex.in /
#     businessapi.talkex.in at this box (verify with `dig +short ...`)
#
# What it does:
#   1. apt install: go, node 22, postgres 18, redis, nginx, certbot, ufw
#   2. Create `talkex` system user
#   3. Clone the repo to /opt/talkex-business (or reuse if present)
#   4. Build API binary + Vite dashboard + copy marketing static
#   5. Set up Postgres (`talkex_business` db, `talkex` role)
#   6. Write /etc/talkex-business.env with a real JWT secret + DB URL
#   7. Install systemd unit `talkex-api.service`
#   8. Copy nginx server blocks + reload
#   9. certbot for the three domains
#  10. ufw allow 22/tcp, 80/tcp, 443/tcp

set -euo pipefail

REPO_URL="${REPO_URL:-https://github.com/coreaxissoftware/talkex_business.git}"
INSTALL_DIR="/opt/talkex-business"
API_USER="talkex"
API_BIN="/usr/local/bin/talkex-api"
ENV_FILE="/etc/talkex-business.env"
WWW_DASH="/var/www/businessapp.talkex.in"
WWW_MKT="/var/www/business.talkex.in"

ADMIN_EMAIL="${ADMIN_EMAIL:-satya@coreaxis.in}"

say() { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }
have() { command -v "$1" >/dev/null 2>&1; }

need_root() {
  if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: run as root" >&2
    exit 1
  fi
}

need_root

# ── 1. Packages ────────────────────────────────────────────────────────
say "Installing system packages"
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq \
  ca-certificates curl gnupg build-essential git ufw \
  golang-go \
  postgresql postgresql-contrib \
  redis-server \
  nginx \
  certbot python3-certbot-nginx

# Node 22 via NodeSource — Ubuntu 26.04's own node package tracks the
# latest LTS but pin explicitly so the Vite build is reproducible.
if ! have node || [ "$(node -v | cut -c2- | cut -d. -f1)" -lt 22 ]; then
  say "Installing Node.js 22 LTS"
  curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
  apt-get install -y -qq nodejs
fi

# ── 2. System user ─────────────────────────────────────────────────────
if ! id "$API_USER" >/dev/null 2>&1; then
  say "Creating system user '$API_USER'"
  useradd --system --no-create-home --shell /usr/sbin/nologin "$API_USER"
fi

# ── 3. Clone or pull ──────────────────────────────────────────────────
if [ ! -d "$INSTALL_DIR/.git" ]; then
  say "Cloning $REPO_URL"
  git clone --depth 20 "$REPO_URL" "$INSTALL_DIR"
else
  say "Pulling latest"
  git -C "$INSTALL_DIR" fetch --depth 20 origin main
  git -C "$INSTALL_DIR" reset --hard origin/main
fi

# ── 4. Postgres ────────────────────────────────────────────────────────
say "Setting up Postgres"
DB_NAME="talkex_business"
DB_USER="talkex"
DB_PASS_FILE="/root/.talkex-db-password"

# Generate a random DB password once and store it root-only.
if [ ! -f "$DB_PASS_FILE" ]; then
  openssl rand -hex 32 > "$DB_PASS_FILE"
  chmod 600 "$DB_PASS_FILE"
fi
DB_PASS="$(cat "$DB_PASS_FILE")"

# Create role + database if missing. `psql -c` is idempotent-friendly here
# because we use `DO $$ ... $$` blocks.
sudo -u postgres psql -v ON_ERROR_STOP=1 <<SQL
DO \$\$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE ROLE ${DB_USER} WITH LOGIN PASSWORD '${DB_PASS}';
  ELSE
    ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASS}';
  END IF;
END \$\$;
SQL

sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}'" | grep -q 1 \
  || sudo -u postgres createdb -O "${DB_USER}" "${DB_NAME}"

# ── 5. Env file ────────────────────────────────────────────────────────
if [ ! -f "$ENV_FILE" ]; then
  say "Writing $ENV_FILE with generated JWT secret"
  JWT_SECRET="$(openssl rand -hex 32)"
  cat > "$ENV_FILE" <<ENV
# TalkEx Business — production env
# Generated $(date -u +%FT%TZ) — DO NOT commit this file.

JWT_SECRET=${JWT_SECRET}
DATABASE_URL=postgres://${DB_USER}:${DB_PASS}@127.0.0.1:5432/${DB_NAME}?sslmode=disable
REDIS_URL=redis://127.0.0.1:6379

PORT=8080
ENVIRONMENT=production
CORS_ORIGINS=https://businessapp.talkex.in,https://business.talkex.in
BASE_URL=https://businessapi.talkex.in
FRONTEND_URL=https://businessapp.talkex.in

# Fill these in after signing up with each provider — see
# deploy/vps/README.md for the full list.
# ANTHROPIC_API_KEY=
# MAILGUN_DOMAIN=
# MAILGUN_API_KEY=
# MSG91_AUTH_KEY=
# FAST2SMS_API_KEY=
# RAZORPAY_KEY_ID=
# SENTRY_DSN=
ENV
fi

# Perms: root:talkex 0640 — root can edit, the talkex user (which the
# migrate CLI runs as) can read. Systemd's EnvironmentFile= reads it
# as PID 1 (root) so the API path doesn't rely on group access.
chown root:"$API_USER" "$ENV_FILE"
chmod 0640 "$ENV_FILE"

# ── 6. Build API + frontend + copy marketing ──────────────────────────
say "Building API binary"
cd "$INSTALL_DIR"
GOFLAGS='-buildvcs=false' go build -trimpath -ldflags='-s -w' -o "$API_BIN" ./cmd/server

# Also build the one-shot migrate tool alongside the API so ops can run
# it manually.
go build -trimpath -ldflags='-s -w' -o /usr/local/bin/talkex-migrate ./cmd/migrate

say "Building Vite dashboard"
cd "$INSTALL_DIR/frontend"
npm ci --silent
VITE_API_URL="https://businessapi.talkex.in" npm run build --silent

mkdir -p "$WWW_DASH"
rsync -a --delete "$INSTALL_DIR/frontend/dist/" "$WWW_DASH/"

say "Copying marketing site"
mkdir -p "$WWW_MKT"
rsync -a --delete "$INSTALL_DIR/marketing/" "$WWW_MKT/"

# Give nginx read access; the tenant-user isn't involved in serving statics.
chown -R www-data:www-data "$WWW_DASH" "$WWW_MKT"

# Create the log + upload dirs the systemd unit's ReadWritePaths grants.
mkdir -p /var/log/talkex /opt/talkex-business/uploads
chown -R "$API_USER:$API_USER" /var/log/talkex /opt/talkex-business/uploads

# Backup dir + daily cron.
mkdir -p /var/backups/talkex
install -m 0755 "$INSTALL_DIR/deploy/vps/backup.sh" /usr/local/bin/talkex-backup
cat > /etc/cron.daily/talkex-backup <<'CRON'
#!/bin/sh
/usr/local/bin/talkex-backup >> /var/log/talkex/backup.log 2>&1
CRON
chmod 0755 /etc/cron.daily/talkex-backup

# ── 7. systemd unit ────────────────────────────────────────────────────
say "Installing systemd unit"
install -m 0644 "$INSTALL_DIR/deploy/vps/talkex-api.service" /etc/systemd/system/talkex-api.service
systemctl daemon-reload
systemctl enable talkex-api

# Run the pending DB migrations before bouncing the API — safer than
# starting up on a stale schema.
say "Running DB migrations"
sudo -u "$API_USER" bash -c "set -a && source '$ENV_FILE' && /usr/local/bin/talkex-migrate" || {
  echo "migrate failed — inspect '$ENV_FILE' and re-run: /usr/local/bin/talkex-migrate"
}

systemctl restart talkex-api

# ── 8. nginx vhosts ────────────────────────────────────────────────────
say "Writing nginx vhosts"
for domain in business.talkex.in businessapp.talkex.in businessapi.talkex.in; do
  install -m 0644 "$INSTALL_DIR/deploy/vps/nginx/${domain}.conf" "/etc/nginx/sites-available/${domain}.conf"
  ln -sf "/etc/nginx/sites-available/${domain}.conf" "/etc/nginx/sites-enabled/${domain}.conf"
done
# NOTE: we intentionally do NOT touch /etc/nginx/sites-enabled/default —
# on shared boxes it often holds a real site (a marketing page, a redirect
# service, an admin portal), and removing it silently deletes someone
# else's content. Our vhosts are `server_name`-scoped, so they coexist
# with anything the box already serves.

nginx -t
systemctl reload nginx

# ── 9. TLS ─────────────────────────────────────────────────────────────
say "Requesting Let's Encrypt certs (per-domain — one failure doesn't sink the others)"
# --nginx installs to the vhost automatically; --redirect flips 80 -> 443
# for us; --non-interactive + --agree-tos + -m keeps the run headless.
# Do one call per domain rather than a batch --expand: if DNS for one
# domain is still propagating, the other two should still get certs.
tls_failed=""
for domain in business.talkex.in businessapp.talkex.in businessapi.talkex.in; do
  if certbot certificates 2>/dev/null | grep -q "$domain"; then
    echo "  cert already installed for $domain — skipping"
    continue
  fi
  if certbot --nginx \
       --non-interactive \
       --agree-tos \
       -m "$ADMIN_EMAIL" \
       --redirect \
       -d "$domain"; then
    echo "  cert issued for $domain"
  else
    echo "  WARN: certbot failed for $domain — check DNS + rerun"
    tls_failed="$tls_failed $domain"
  fi
done
if [ -n "$tls_failed" ]; then
  echo
  echo "Some domains did not get a cert:$tls_failed"
  echo "Verify DNS with:  dig +short$tls_failed"
  echo "Then re-run this script — it is idempotent."
fi

# ── 10. Firewall — additive only ──────────────────────────────────────
#
# NO reset here: this box hosts other apps whose UFW rules we must NOT
# wipe. We just add the ones we need (idempotent — `ufw allow` on an
# existing rule is a no-op) and enable UFW without touching the default
# policies (whatever ops already picked stands).
say "Adding UFW rules (additive, non-destructive)"
ufw allow 22/tcp   comment 'SSH'           >/dev/null || true
ufw allow 80/tcp   comment 'HTTP (redirects to HTTPS)' >/dev/null || true
ufw allow 443/tcp  comment 'HTTPS'         >/dev/null || true
# Enable only when currently inactive — never touch an active tuned firewall.
if ! ufw status | grep -q "Status: active"; then
  echo "  UFW is inactive; leaving it that way (enable manually if desired)"
  echo "    ufw enable"
fi

say "Bootstrap complete"
echo
echo "Live services:"
echo "  https://business.talkex.in      (marketing site)"
echo "  https://businessapp.talkex.in  (dashboard)"
echo "  https://businessapi.talkex.in  (API — try /health)"
echo
echo "Next steps:"
echo "  1. Edit $ENV_FILE with your real provider secrets"
echo "     (Anthropic, Mailgun, MSG91, Fast2SMS, Meta, Razorpay, Sentry, OAuth...)"
echo "  2. systemctl restart talkex-api"
echo "  3. curl -sS https://businessapi.talkex.in/health"
echo
echo "For updates: bash /opt/talkex-business/deploy/vps/update.sh"
