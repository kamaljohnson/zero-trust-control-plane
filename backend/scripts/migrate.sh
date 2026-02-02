#!/usr/bin/env bash
# migrate.sh: run database migrations (golang-migrate CLI or go run ./cmd/migrate).
set -euo pipefail
cd "$(dirname "$0")/.."

# Load DATABASE_URL from .env if not already set (avoid sourcing whole file; .env may contain PEM keys).
if [ -z "${DATABASE_URL:-}" ] && [ -f .env ]; then
  while IFS= read -r line; do
    case "$line" in
      DATABASE_URL=*) export DATABASE_URL="${line#DATABASE_URL=}"; break ;;
    esac
  done < .env
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is not set; create a .env from .env.example or set DATABASE_URL" >&2
  exit 1
fi

DIR="${1:-up}"
if [ "$DIR" != "up" ] && [ "$DIR" != "down" ]; then
  echo "Usage: $0 [up|down]" >&2
  echo "  up   (default) apply pending migrations" >&2
  echo "  down roll back one migration" >&2
  exit 1
fi

if command -v migrate >/dev/null 2>&1; then
  migrate -path internal/db/migrations -database "$DATABASE_URL" "$DIR"
else
  go run ./cmd/migrate -direction="$DIR"
fi
