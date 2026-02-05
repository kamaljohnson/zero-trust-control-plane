#!/usr/bin/env bash
# setup-ssl.sh: Setup SSL certificates using Let's Encrypt with certbot
# This script installs certbot, obtains certificates, and configures auto-renewal

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

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Install certbot
install_certbot() {
    log_info "Installing certbot..."
    
    if command -v certbot &> /dev/null; then
        log_info "Certbot is already installed"
        return 0
    fi
    
    # Detect OS and install certbot
    if [ -f /etc/debian_version ]; then
        apt-get update
        apt-get install -y certbot
    elif [ -f /etc/redhat-release ]; then
        yum install -y certbot
    elif [ -f /etc/arch-release ]; then
        pacman -S --noconfirm certbot
    else
        log_error "Unsupported OS. Please install certbot manually."
        exit 1
    fi
    
    log_info "Certbot installed successfully"
}

# Get domain from .env.prod
get_domain() {
    if [ -f .env.prod ]; then
        source .env.prod 2>/dev/null || true
        echo "${DOMAIN:-}"
    fi
}

# Obtain SSL certificate
obtain_certificate() {
    local domain=$(get_domain)
    
    if [ -z "$domain" ]; then
        read -p "Enter your domain name: " domain
    fi
    
    log_info "Obtaining SSL certificate for $domain..."
    
    # Create directory for certificates
    mkdir -p nginx/ssl
    
    # Stop nginx if running (certbot needs port 80)
    docker compose -f docker-compose.prod.yml down nginx 2>/dev/null || true
    
    # Obtain certificate using standalone mode (certbot writes to /etc/letsencrypt/live/<domain>/)
    certbot certonly --standalone \
        --non-interactive \
        --agree-tos \
        --email "admin@$domain" \
        -d "$domain" || {
        log_error "Failed to obtain certificate"
        exit 1
    }
    
    # Copy to nginx/ssl (certbot does not support custom output paths; nginx expects fullchain.pem and privkey.pem)
    local live_dir="/etc/letsencrypt/live/$domain"
    if [ ! -d "$live_dir" ]; then
        log_error "Certificate directory not found: $live_dir"
        exit 1
    fi
    cp "$live_dir/fullchain.pem" nginx/ssl/fullchain.pem
    cp "$live_dir/privkey.pem" nginx/ssl/privkey.pem
    chmod 600 nginx/ssl/privkey.pem
    
    log_info "SSL certificate obtained successfully"
}

# Setup auto-renewal
setup_renewal() {
    log_info "Setting up certificate auto-renewal..."
    
    local deploy_hook="$DEPLOY_DIR/scripts/certbot-deploy-hook.sh"
    if [ ! -x "$deploy_hook" ]; then
        chmod +x "$deploy_hook" 2>/dev/null || log_warn "Could not chmod +x $deploy_hook"
    fi
    
    # Create renewal script (copy certs to nginx/ssl then restart nginx via deploy-hook)
    cat > /etc/cron.monthly/ztcp-cert-renewal << EOF
#!/bin/bash
# Renew SSL certificates; deploy-hook copies to nginx/ssl and restarts nginx
cd $DEPLOY_DIR
certbot renew --quiet --deploy-hook "$deploy_hook"
EOF
    
    chmod +x /etc/cron.monthly/ztcp-cert-renewal
    
    # Test renewal
    certbot renew --dry-run || {
        log_warn "Certificate renewal test failed, but continuing..."
    }
    
    log_info "Auto-renewal configured"
}

# Main function
main() {
    check_root
    
    log_info "Setting up SSL certificates..."
    
    install_certbot
    obtain_certificate
    setup_renewal
    
    log_info "SSL setup completed!"
    log_info "Certificates are in: $DEPLOY_DIR/nginx/ssl/"
    log_info "Auto-renewal is configured via cron"
}

main "$@"
