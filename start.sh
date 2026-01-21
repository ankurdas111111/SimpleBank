#!/bin/sh

set -e

# Usage:
#   /app/start.sh postgres:5432 -- /app/main
#
# Wait for DB, run migrations, then start the app.

HOSTPORT="${1:-}"
if echo "$HOSTPORT" | grep -q ":"; then
  shift
  if [ "${1:-}" = "--" ]; then
    shift
  fi

  echo "waiting for db: $HOSTPORT"
  /app/wait-for.sh -t 30 "$HOSTPORT"
fi

echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"


