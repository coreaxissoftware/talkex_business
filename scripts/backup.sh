#!/bin/sh
# Nightly backup — Postgres full dump + rotate old files (14-day window).
# Wired into deploy/docker-compose.prod.yml as a cron job.
set -eu

TS=$(date -u +%Y%m%d-%H%M%S)
OUT="/backups/talkex-${TS}.sql.gz"

echo "[backup] starting → $OUT"
pg_dump --format=plain --no-owner --no-privileges "$PGDATABASE" | gzip -9 > "$OUT"
SIZE=$(du -h "$OUT" | cut -f1)
echo "[backup] wrote $SIZE"

# Rotate: keep 14 days
find /backups -name 'talkex-*.sql.gz' -mtime +14 -delete
echo "[backup] rotation complete"

# Uncomment to push offsite (rclone must be configured with the "backups" remote)
# rclone copy "$OUT" backups:talkex-daily/ && echo "[backup] pushed offsite"
