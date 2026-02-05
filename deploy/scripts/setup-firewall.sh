#!/usr/bin/env bash
# setup-firewall.sh: Configure UFW firewall for production deployment
# Allows SSH, HTTP, and HTTPS; blocks all other ports

set -euo pipefail

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

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Check if UFW is installed
check_ufw() {
    if ! command -v ufw &> /dev/null; then
        log_info "Installing UFW..."
        apt-get update
        apt-get install -y ufw
    fi
}

# Configure firewall
configure_firewall() {
    log_info "Configuring UFW firewall..."
    
    # Reset UFW to defaults
    ufw --force reset
    
    # Set default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH (important: do this first!)
    ufw allow 22/tcp comment 'SSH'
    
    # Allow HTTP
    ufw allow 80/tcp comment 'HTTP'
    
    # Allow HTTPS
    ufw allow 443/tcp comment 'HTTPS'
    
    # Enable firewall
    ufw --force enable
    
    log_info "Firewall configured successfully"
}

# Show firewall status
show_status() {
    log_info "Firewall Status:"
    ufw status verbose
}

# Main function
main() {
    check_root
    
    log_info "Setting up firewall..."
    
    check_ufw
    configure_firewall
    show_status
    
    log_info "Firewall setup completed!"
    log_warn "Make sure SSH access is working before disconnecting!"
}

main "$@"
