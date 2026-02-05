#!/usr/bin/env bash
# certbot-deploy-hook.sh: Copy renewed certificates to nginx/ssl and reload nginx.
# Called by certbot renew --deploy-hook. Certbot sets RENEWED_LINEAGE (e.g. /etc/letsencrypt/live/example.com).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

if [ -z "${RENEWED_LINEAGE:-}" ]; then
    echo "RENEWED_LINEAGE not set (not run by certbot?)" >&2
    exit 1
fi

cp "$RENEWED_LINEAGE/fullchain.pem" "$DEPLOY_DIR/nginx/ssl/fullchain.pem"
cp "$RENEWED_LINEAGE/privkey.pem" "$DEPLOY_DIR/nginx/ssl/privkey.pem"
chmod 600 "$DEPLOY_DIR/nginx/ssl/privkey.pem"

cd "$DEPLOY_DIR"
docker compose -f docker-compose.prod.yml restart nginx
