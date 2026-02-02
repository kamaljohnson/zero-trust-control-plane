#!/usr/bin/env bash
# seed.sh: seed development/sample data (run after ./scripts/migrate.sh).
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

go run ./cmd/seed
