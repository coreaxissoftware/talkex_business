#!/usr/bin/env bash
#
# backup.sh — nightly Postgres dump + uploads snapshot.
#
# Called from a cron entry setup.sh writes to /etc/cron.daily/talkex-backup.
# Manual run: bash /opt/talkex-business/deploy/vps/backup.sh

set -euo pipefail

BACKUP_DIR="/var/backups/talkex"
KEEP_DAYS=14

mkdir -p "$BACKUP_DIR"

TS="$(date -u +%Y%m%d-%H%M%S)"
DUMP="$BACKUP_DIR/db-$TS.sql.gz"

# Postgres dump — DATABASE_URL comes from the env file the API uses,
# so credentials stay in one place.
set -a
source /etc/talkex-business.env
set +a

pg_dump "$DATABASE_URL" | gzip -9 > "$DUMP"

# Uploads snapshot (media library) — tar+gz; the media dir is small.
if [ -d /opt/talkex-business/uploads ]; then
  tar czf "$BACKUP_DIR/uploads-$TS.tar.gz" -C /opt/talkex-business uploads
fi

# Rotate: drop dumps older than KEEP_DAYS.
find "$BACKUP_DIR" -type f -mtime "+$KEEP_DAYS" -delete

echo "backup complete: $DUMP ($(du -h "$DUMP" | cut -f1))"
