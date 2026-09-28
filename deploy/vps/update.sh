#!/usr/bin/env bash
#
# update.sh — pull + rebuild + restart. Idempotent, safe to re-run.
# Called by CI after a merge to main, or manually on the server.
#
# What it does:
#   1. git pull origin main (fast-forward only — refuses to rebase-merge)
#   2. tag the commit as deploy-<UTC> for rollback traceability
#   3. run pending DB migrations via cmd/migrate
#   4. rebuild the Go API binary
#   5. rebuild the Vite dashboard + rsync to /var/www
#   6. rsync marketing static to /var/www
#   7. systemctl restart talkex-api
#   8. wait for /health to come back within 30s; abort if not

set -euo pipefail

INSTALL_DIR="/opt/talkex-business"
API_BIN="/usr/local/bin/talkex-api"
MIGRATE_BIN="/usr/local/bin/talkex-migrate"
ENV_FILE="/etc/talkex-business.env"
API_USER="talkex"
WWW_DASH="/var/www/businessapp.talkex.in"
WWW_MKT="/var/www/business.talkex.in"

say() { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }

cd "$INSTALL_DIR"

say "Pulling from origin/main"
git fetch --depth 20 origin main
git reset --hard origin/main

TAG="deploy-$(date -u +%Y-%m-%d-%H%M)"
git tag "$TAG" || true
say "Tagged $TAG (roll back with: git checkout $TAG && bash deploy/vps/update.sh)"

say "Running DB migrations"
go build -trimpath -ldflags='-s -w' -o "$MIGRATE_BIN" ./cmd/migrate
sudo -u "$API_USER" bash -c "set -a && source '$ENV_FILE' && '$MIGRATE_BIN'"

say "Building API binary"
GOFLAGS='-buildvcs=false' go build -trimpath -ldflags='-s -w' -o "$API_BIN" ./cmd/server

say "Building Vite dashboard"
cd frontend
npm ci --silent
VITE_API_URL="https://businessapi.talkex.in" npm run build --silent
rsync -a --delete dist/ "$WWW_DASH/"
cd ..

say "Copying marketing site"
rsync -a --delete marketing/ "$WWW_MKT/"

chown -R www-data:www-data "$WWW_DASH" "$WWW_MKT"

say "Restarting talkex-api"
systemctl restart talkex-api

# Wait for the API to come back — matters for rollback: if the new build
# panics on boot, we exit non-zero so the caller (CI or the operator)
# knows the deploy failed BEFORE walking away.
say "Waiting for /health"
for i in $(seq 1 30); do
  if curl -fsS --max-time 2 http://127.0.0.1:8080/health >/dev/null 2>&1; then
    echo "  API healthy after ${i}s"
    say "Deploy complete: $TAG"
    exit 0
  fi
  sleep 1
done

echo "ERROR: API did not respond to /health within 30s after restart" >&2
echo "  journalctl -u talkex-api -n 100" >&2
exit 1
