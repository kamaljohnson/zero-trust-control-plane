#!/usr/bin/env bash
# generate-jwt-keys.sh: Generate RSA key pair for JWT authentication
# Creates private and public keys with proper permissions

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

# Check if openssl is installed
check_openssl() {
    if ! command -v openssl &> /dev/null; then
        log_error "OpenSSL is not installed. Please install it first."
        exit 1
    fi
}

# Generate RSA key pair
generate_keys() {
    local key_dir="$DEPLOY_DIR/keys"
    local private_key="$key_dir/jwt_private.pem"
    local public_key="$key_dir/jwt_public.pem"
    
    # Create keys directory
    mkdir -p "$key_dir"
    
    # Generate private key (RSA 2048-bit)
    log_info "Generating RSA private key..."
    openssl genpkey -algorithm RSA -out "$private_key" -pkeyopt rsa_keygen_bits:2048
    
    # Set proper permissions (read-only for owner)
    chmod 600 "$private_key"
    
    # Generate public key from private key
    log_info "Generating RSA public key..."
    openssl rsa -pubout -in "$private_key" -out "$public_key"
    
    # Set proper permissions
    chmod 644 "$public_key"
    
    log_info "Keys generated successfully!"
    log_info "Private key: $private_key"
    log_info "Public key: $public_key"
    
    # Display keys for .env.prod configuration
    echo
    log_info "Add these to your .env.prod file:"
    echo
    echo "JWT_PRIVATE_KEY=\"$(cat "$private_key" | tr '\n' '|' | sed 's/|/\\n/g')\""
    echo "JWT_PUBLIC_KEY=\"$(cat "$public_key" | tr '\n' '|' | sed 's/|/\\n/g')\""
    echo
    log_warn "Or use file paths:"
    echo "JWT_PRIVATE_KEY=$private_key"
    echo "JWT_PUBLIC_KEY=$public_key"
    echo
    log_warn "Keep the private key secure and never commit it to version control!"
}

# Main function
main() {
    check_openssl
    generate_keys
}

main "$@"
