#!/usr/bin/env bash
# generate-postgres-ssl.sh: Generate self-signed SSL certificate for PostgreSQL (production).
# Writes server.crt and server.key to deploy/postgres-ssl/ for use by the postgres container.
# Run from repo root or deploy directory. Creates postgres-ssl/ if missing.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
SSL_DIR="$DEPLOY_DIR/postgres-ssl"

mkdir -p "$SSL_DIR"
cd "$SSL_DIR"

# 10-year self-signed cert for postgres (internal use; client uses sslmode=require without verifying hostname by default)
openssl req -new -x509 -days 3650 -nodes \
  -out server.crt -keyout server.key \
  -subj "/CN=postgres"

# Set initial permissions on host (will be fixed to proper ownership/permissions by postgres container entrypoint)
# PostgreSQL requires private key to be 600 and owned by postgres user (UID 999)
# The docker-compose entrypoint will fix ownership and permissions when the container starts
chmod 755 "$SSL_DIR"
chmod 644 server.key  # Will be changed to 600 by container entrypoint
chmod 644 server.crt

echo "PostgreSQL SSL certificate and key written to $SSL_DIR (server.crt, server.key)"
echo "Note: The postgres container entrypoint will fix ownership to postgres user (UID 999) and set key permissions to 600 on startup"
