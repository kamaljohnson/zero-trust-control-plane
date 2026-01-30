#!/usr/bin/env bash
# migrate.sh: run database migrations (e.g. golang-migrate or similar)
set -euo pipefail
cd "$(dirname "$0")/.."
# TODO: run migrations from internal/db/migrations/ against DATABASE_URL
