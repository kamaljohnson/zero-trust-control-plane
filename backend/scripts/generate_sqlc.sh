#!/usr/bin/env bash
# Generate sqlc code for the single shared sqlc project at internal/db/sqlc.
# See: https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html#setting-up
#
# Install sqlc first:
#   brew install sqlc
# or
#   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
# (then ensure $GOPATH/bin or $HOME/go/bin is on your PATH)
set -e
cd "$(dirname "$0")/.."
if ! command -v sqlc >/dev/null 2>&1; then
  echo "sqlc not found. Install it with: brew install sqlc"
  echo "  or: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
  exit 1
fi
echo "sqlc generate: internal/db/sqlc"
(cd internal/db/sqlc && sqlc generate)
echo "sqlc generate done."
