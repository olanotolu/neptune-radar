#!/bin/sh
# Neptune Radar — nightly Postgres backup. Runs inside the neptune-backup
# container (postgres:17-alpine) via the host cron:
#   0 4 * * * docker compose -f ~/neptune-radar/deploy/docker-compose.yml --env-file ~/neptune-radar.env --profile backup run --rm backup
set -eu

STAMP=$(date -u +%Y%m%dT%H%M%SZ)
OUT="/backups/neptune-${STAMP}.sql.gz"

# The live database container predates compose and keeps its hand-set name.
pg_dump -h "${PGHOST:-neptune-pg}" -U "${POSTGRES_USER:-neptune}" -d "${POSTGRES_DB:-neptune}" | gzip >"$OUT"
echo "backup written: $OUT ($(du -h "$OUT" | cut -f1))"

# Retention: keep the last 14 nightly dumps.
cd /backups
ls -1t neptune-*.sql.gz 2>/dev/null | tail -n +15 | xargs -r rm -f
