#!/bin/sh
# postgres-entrypoint.sh: Wrapper to fix SSL cert permissions before starting postgres
# This is needed because postgres requires the private key to be owned by the postgres user (UID 70)
# and have 600 permissions, but volume mounts preserve host ownership.

set -e

# Fix SSL certificate ownership and permissions for postgres user (UID 70, GID 70)
if [ -d /etc/postgres-ssl ]; then
    chown -R 70:70 /etc/postgres-ssl
    chmod 600 /etc/postgres-ssl/server.key
    chmod 644 /etc/postgres-ssl/server.crt
    # Verify permissions
    ls -la /etc/postgres-ssl/ >&2 || true
fi

# Call the original postgres entrypoint with SSL enabled
exec /usr/local/bin/docker-entrypoint.sh "$@"
