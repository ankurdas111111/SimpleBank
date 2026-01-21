#!/bin/sh
#
# Minimal wait-for script (TCP only) to block until host:port is reachable.
# Usage: ./wait-for.sh -t 30 host:port

set -e

TIMEOUT=15

while [ $# -gt 0 ]; do
  case "$1" in
    -t|--timeout)
      TIMEOUT="$2"
      shift 2
      ;;
    -t*)
      TIMEOUT="${1#-t}"
      shift 1
      ;;
    --timeout=*)
      TIMEOUT="${1#*=}"
      shift 1
      ;;
    *)
      TARGET="$1"
      shift 1
      ;;
  esac
done

HOST=$(printf "%s" "$TARGET" | cut -d: -f1)
PORT=$(printf "%s" "$TARGET" | cut -d: -f2)

if [ -z "$HOST" ] || [ -z "$PORT" ]; then
  echo "usage: $0 [-t seconds] host:port" >&2
  exit 2
fi

END=$(( $(date +%s) + TIMEOUT ))

while :; do
  if nc -w 1 -z "$HOST" "$PORT" >/dev/null 2>&1; then
    exit 0
  fi
  if [ "$TIMEOUT" -ne 0 ] && [ "$(date +%s)" -ge "$END" ]; then
    echo "timeout waiting for $HOST:$PORT" >&2
    exit 1
  fi
  sleep 1
done


