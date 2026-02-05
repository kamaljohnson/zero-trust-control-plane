#!/usr/bin/env bash
# restore.sh: Restore PostgreSQL database and Docker volumes from backup
# Usage: ./restore.sh <backup_file>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

cd "$DEPLOY_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check backup file argument
if [ $# -eq 0 ]; then
    log_error "Usage: $0 <backup_file>"
    log_info "Available backups:"
    ls -lh backups/ztcp_backup_*.tar.gz 2>/dev/null || log_warn "No backups found"
    exit 1
fi

BACKUP_FILE="$1"
BACKUP_DIR="${BACKUP_DIR:-$DEPLOY_DIR/backups}"
RESTORE_DIR="$BACKUP_DIR/restore_$$"

# Extract backup archive
extract_backup() {
    log_info "Extracting backup archive..."
    
    if [ ! -f "$BACKUP_FILE" ]; then
        log_error "Backup file not found: $BACKUP_FILE"
        exit 1
    fi
    
    mkdir -p "$RESTORE_DIR"
    tar xzf "$BACKUP_FILE" -C "$RESTORE_DIR" || {
        log_error "Failed to extract backup"
        exit 1
    }
    
    log_info "Backup extracted to: $RESTORE_DIR"
}

# Restore PostgreSQL database
restore_database() {
    log_info "Restoring PostgreSQL database..."
    
    # Load database credentials from .env.prod
    if [ -f .env.prod ]; then
        source .env.prod 2>/dev/null || true
    fi
    
    local db_user="${POSTGRES_USER:-ztcp}"
    local db_name="${POSTGRES_DB:-ztcp}"
    
    # Find database backup file
    local db_backup=$(find "$RESTORE_DIR" -name "*_database.sql.gz" | head -1)
    
    if [ -z "$db_backup" ]; then
        log_warn "Database backup not found in archive"
        return 0
    fi
    
    # Decompress if needed
    if [[ "$db_backup" == *.gz ]]; then
        gunzip -c "$db_backup" > "${db_backup%.gz}" || {
            log_error "Failed to decompress database backup"
            return 1
        }
        db_backup="${db_backup%.gz}"
    fi
    
    # Confirm restore
    log_warn "This will overwrite the current database!"
    read -p "Are you sure you want to continue? (yes/no) " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Restore cancelled"
        return 1
    fi
    
    # Restore database
    docker compose -f docker-compose.prod.yml exec -T postgres \
        psql -U "$db_user" -d "$db_name" < "$db_backup" || {
        log_error "Database restore failed"
        return 1
    }
    
    log_info "Database restored successfully"
}

# Restore Docker volumes
# Volume names must match Compose project prefix (default: directory name, e.g. deploy_postgres_data)
restore_volumes() {
    log_info "Restoring Docker volumes..."
    
    local volume_prefix="${COMPOSE_PROJECT_NAME:-$(basename "$DEPLOY_DIR")}"
    local volumes=(
        "${volume_prefix}_postgres_data"
        "${volume_prefix}_prometheus_data"
        "${volume_prefix}_grafana_data"
        "${volume_prefix}_tempo_data"
        "${volume_prefix}_loki_data"
    )
    
    for volume in "${volumes[@]}"; do
        local volume_backup=$(find "$RESTORE_DIR" -name "*_${volume}.tar.gz" | head -1)
        
        if [ -z "$volume_backup" ]; then
            log_warn "Backup for volume $volume not found, skipping"
            continue
        fi
        
        log_info "Restoring volume: $volume"
        
        # Confirm restore
        log_warn "This will overwrite volume: $volume"
        read -p "Continue? (yes/no) " -r
        if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            log_info "Skipping volume: $volume"
            continue
        fi
        
        # Stop services using the volume
        docker compose -f docker-compose.prod.yml stop || true
        
        # Remove existing volume
        docker volume rm "$volume" 2>/dev/null || true
        
        # Create new volume
        docker volume create "$volume" || {
            log_error "Failed to create volume: $volume"
            continue
        }
        
        # Restore volume data
        docker run --rm \
            -v "$volume":/data \
            -v "$RESTORE_DIR":/backup \
            alpine sh -c "cd /data && tar xzf /backup/$(basename "$volume_backup")" || {
            log_error "Failed to restore volume: $volume"
            continue
        }
        
        log_info "Volume restored: $volume"
    done
    
    log_info "Volume restoration completed"
}

# Cleanup restore directory
cleanup() {
    log_info "Cleaning up..."
    rm -rf "$RESTORE_DIR"
}

# Verify restore
verify_restore() {
    log_info "Verifying restore..."
    
    # Check database connection
    docker compose -f docker-compose.prod.yml exec -T postgres \
        pg_isready -U "${POSTGRES_USER:-ztcp}" || {
        log_error "Database verification failed"
        return 1
    }
    
    log_info "Restore verification completed"
}

# Main function
main() {
    log_info "Starting restore process..."
    log_warn "This will overwrite existing data!"
    
    extract_backup
    restore_database
    restore_volumes
    verify_restore
    cleanup
    
    log_info "Restore completed successfully!"
    log_info "Restart services: docker compose -f docker-compose.prod.yml --env-file .env.prod up -d"
}

# Trap to cleanup on exit
trap cleanup EXIT

main "$@"
