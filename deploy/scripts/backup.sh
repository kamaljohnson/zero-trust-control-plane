#!/usr/bin/env bash
# backup.sh: Backup PostgreSQL database and Docker volumes
# Creates timestamped backups with compression

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

# Configuration
BACKUP_DIR="${BACKUP_DIR:-$DEPLOY_DIR/backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="ztcp_backup_$TIMESTAMP"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup PostgreSQL database
backup_database() {
    log_info "Backing up PostgreSQL database..."
    
    # Load database credentials from .env.prod
    if [ -f .env.prod ]; then
        source .env.prod 2>/dev/null || true
    fi
    
    local db_user="${POSTGRES_USER:-ztcp}"
    local db_name="${POSTGRES_DB:-ztcp}"
    local backup_file="$BACKUP_DIR/${BACKUP_NAME}_database.sql"
    
    # Create database backup using pg_dump
    docker compose -f docker-compose.prod.yml exec -T postgres \
        pg_dump -U "$db_user" "$db_name" > "$backup_file" || {
        log_error "Database backup failed"
        return 1
    }
    
    # Compress backup
    gzip "$backup_file" || {
        log_warn "Failed to compress database backup"
    }
    
    log_info "Database backup created: ${backup_file}.gz"
}

# Backup Docker volumes
# Volume names use Compose project prefix (default: directory name, e.g. deploy_postgres_data)
backup_volumes() {
    log_info "Backing up Docker volumes..."
    
    local volume_prefix="${COMPOSE_PROJECT_NAME:-$(basename "$DEPLOY_DIR")}"
    local volumes=(
        "${volume_prefix}_postgres_data"
    )
    
    for volume in "${volumes[@]}"; do
        if docker volume inspect "$volume" &> /dev/null; then
            log_info "Backing up volume: $volume"
            docker run --rm \
                -v "$volume":/data \
                -v "$BACKUP_DIR":/backup \
                alpine tar czf "/backup/${BACKUP_NAME}_${volume}.tar.gz" -C /data . || {
                log_warn "Failed to backup volume: $volume"
            }
        fi
    done
    
    log_info "Volume backups completed"
}

# Create backup archive
create_archive() {
    log_info "Creating backup archive..."
    
    local archive_file="$BACKUP_DIR/${BACKUP_NAME}.tar.gz"
    
    cd "$BACKUP_DIR"
    tar czf "$archive_file" "${BACKUP_NAME}"* 2>/dev/null || {
        log_warn "Failed to create archive, individual files are available"
        return 0
    }
    
    # Remove individual component files only (keep main archive ${BACKUP_NAME}.tar.gz)
    rm -f "${BACKUP_DIR}/${BACKUP_NAME}"_*.sql.gz "${BACKUP_DIR}/${BACKUP_NAME}"_*.tar.gz
    
    log_info "Backup archive created: $archive_file"
    
    # Display backup size
    local size=$(du -h "$archive_file" | cut -f1)
    log_info "Backup size: $size"
}

# Cleanup old backups (keep last 7 days)
cleanup_old_backups() {
    log_info "Cleaning up old backups (keeping last 7 days)..."
    
    find "$BACKUP_DIR" -name "ztcp_backup_*.tar.gz" -mtime +7 -delete || true
    
    log_info "Cleanup completed"
}

# Main function
main() {
    log_info "Starting backup process..."
    
    backup_database
    backup_volumes
    create_archive
    cleanup_old_backups
    
    log_info "Backup completed successfully!"
    log_info "Backup location: $BACKUP_DIR"
}

main "$@"
