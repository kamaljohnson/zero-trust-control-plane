#!/usr/bin/env bash
# deploy.sh: Main deployment script for ZTCP production deployment
# This script handles the complete deployment process including migrations and health checks

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(dirname "$DEPLOY_DIR")"

cd "$DEPLOY_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Deployment status tracking
STATUS_FILE="$DEPLOY_DIR/nginx/status.json"

# Initialize status file
init_status() {
    mkdir -p "$(dirname "$STATUS_FILE")"
    cat > "$STATUS_FILE" << 'EOF'
{
  "status": "running",
  "message": "Deployment in progress",
  "timestamp": "",
  "steps": []
}
EOF
}

# Update status with steps array (steps should be JSON array of objects)
update_status() {
    local status=$1
    local message=$2
    local steps_json=$3
    
    cat > "$STATUS_FILE" << EOF
{
  "status": "$status",
  "message": "$message",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "steps": $steps_json
}
EOF
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose v2+."
        exit 1
    fi
    
    # Check if .env.prod exists
    if [ ! -f .env.prod ]; then
        log_error ".env.prod file not found!"
        log_info "Please copy .env.prod.example to .env.prod and configure it:"
        log_info "  cp .env.prod.example .env.prod"
        log_info "  # Edit .env.prod with your production values"
        exit 1
    fi
    
    # Check if JWT keys are set
    source .env.prod 2>/dev/null || true
    if [ -z "${JWT_PRIVATE_KEY:-}" ] || [ -z "${JWT_PUBLIC_KEY:-}" ]; then
        log_warn "JWT keys are not set in .env.prod"
        log_info "Generate JWT keys using: ./scripts/generate-jwt-keys.sh"
        read -p "Do you want to generate JWT keys now? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            "$SCRIPT_DIR/generate-jwt-keys.sh"
        else
            log_error "JWT keys are required. Exiting."
            exit 1
        fi
    fi
    
    # Check if SSL certificates exist (nginx)
    if [ ! -f nginx/ssl/fullchain.pem ] || [ ! -f nginx/ssl/privkey.pem ]; then
        log_warn "SSL certificates not found in nginx/ssl/"
        log_info "Run ./scripts/setup-ssl.sh to generate SSL certificates first"
        read -p "Do you want to continue without SSL? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Auto-generate PostgreSQL SSL certificates if missing
    if [ ! -f postgres-ssl/server.crt ] || [ ! -f postgres-ssl/server.key ]; then
        log_warn "PostgreSQL SSL certificates not found in postgres-ssl/"
        log_info "Auto-generating PostgreSQL SSL certificates..."
        "$SCRIPT_DIR/generate-postgres-ssl.sh"
    fi
    
    log_info "Prerequisites check passed"
}

# Wait for service to be healthy
wait_for_service() {
    local service=$1
    local max_attempts=30
    local attempt=1
    
    log_info "Waiting for $service to be healthy..."
    
    while [ $attempt -le $max_attempts ]; do
        if docker compose -f docker-compose.prod.yml --env-file .env.prod ps "$service" 2>/dev/null | grep -q "healthy"; then
            log_info "$service is healthy"
            return 0
        fi
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "$service failed to become healthy"
    return 1
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."
    
    # Wait for postgres to be ready
    wait_for_service postgres || {
        log_error "PostgreSQL is not ready. Cannot run migrations."
        return 1
    }
    
    # Load DATABASE_URL from .env.prod
    if [ -f .env.prod ]; then
        while IFS= read -r line; do
            case "$line" in
                DATABASE_URL=*) export DATABASE_URL="${line#DATABASE_URL=}"; break ;;
            esac
        done < .env.prod
    fi
    
    if [ -z "${DATABASE_URL:-}" ]; then
        log_error "DATABASE_URL not found in .env.prod"
        return 1
    fi
    
    # Migrations run on the host; use localhost so we connect to the published postgres port (5432).
    # (.env.prod typically uses host "postgres" for containers; that name only resolves inside Docker.)
    # Keep sslmode from .env.prod (use sslmode=require when postgres is configured with SSL).
    DATABASE_URL=$(echo "$DATABASE_URL" | sed 's/@postgres:/@localhost:/g')
    
    # Run migrations using backend migrate script
    log_info "Running migrations from backend directory..."
    cd "$REPO_ROOT/backend"
    
    # Use migrate.sh script which handles both golang-migrate CLI and go run
    export DATABASE_URL
    ./scripts/migrate.sh up || {
        log_error "Migrations failed"
        log_error "Ensure DATABASE_URL uses the same user and password as POSTGRES_USER and POSTGRES_PASSWORD in .env.prod"
        cd "$DEPLOY_DIR"
        return 1
    }
    
    cd "$DEPLOY_DIR"
    log_info "Migrations completed successfully"
}

# Build or pull Docker images
# When DOCKER_REGISTRY and IMAGE_TAG are set (e.g. from CI/CD), pull from registry.
# Otherwise build locally.
build_or_pull_images() {
    if [ -n "${DOCKER_REGISTRY:-}" ] && [ "${DOCKER_REGISTRY}" != "local" ] && [ -n "${IMAGE_TAG:-}" ]; then
        log_info "Pulling images from registry (DOCKER_REGISTRY=$DOCKER_REGISTRY, IMAGE_TAG=$IMAGE_TAG)..."
        update_status "running" "Pulling images from registry..." '[
          {"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"},
          {"title":"Pulling Images","status":"active","details":"Pulling images from $DOCKER_REGISTRY..."}
        ]'
        export DOCKER_REGISTRY IMAGE_TAG
        docker compose -f docker-compose.prod.yml --env-file .env.prod pull || {
            log_error "Failed to pull Docker images"
            return 1
        }
        log_info "Docker images pulled successfully"
    else
        log_info "Building Docker images locally (one at a time to reduce memory use)..."
        export DOCKER_REGISTRY="${DOCKER_REGISTRY:-local}"
        export IMAGE_TAG="${IMAGE_TAG:-latest}"
        local services=("backend" "frontend" "docs" "nginx")
        local built_services=()
        for svc in "${services[@]}"; do
            log_info "Building image: $svc"
            local current=$(( ${#built_services[@]} + 1 ))
            local total=${#services[@]}
            local build_steps="[
              {\"title\":\"Prerequisites Check\",\"status\":\"complete\",\"details\":\"All prerequisites verified\"},
              {\"title\":\"Building Images\",\"status\":\"active\",\"details\":\"Building $svc ($current/$total)...\"}
            ]"
            update_status "running" "Building Docker images..." "$build_steps"
            docker compose -f docker-compose.prod.yml --env-file .env.prod build --no-cache "$svc" || {
                log_error "Failed to build Docker image: $svc"
                return 1
            }
            built_services+=("$svc")
        done
        log_info "Docker images built successfully"
    fi
}

# Start services
start_services() {
    log_info "Starting services..."
    if ! docker compose -f docker-compose.prod.yml --env-file .env.prod up -d; then
        log_error "Failed to start services"
        log_info "If backend is unhealthy, check logs: docker compose -f docker-compose.prod.yml --env-file .env.prod logs backend"
        log_info "Backend often fails when JWT keys in .env.prod are corrupted by \$ (e.g. 'P8R variable not set'). Escape any \$ in PEM values as \$\$"
        return 1
    fi
    log_info "Services started"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Wait for services to be healthy
    wait_for_service postgres
    wait_for_service backend
    wait_for_service frontend
    wait_for_service nginx
    
    # Check health endpoints
    log_info "Checking health endpoints..."
    
    # Backend health (via nginx)
    if curl -f -s https://localhost/health > /dev/null 2>&1; then
        log_info "✓ Nginx health check passed"
    else
        log_warn "Nginx health check failed (this may be normal if SSL is not configured)"
    fi
    
    log_info "Deployment verification completed"
}

# Display deployment status
show_status() {
    log_info "Deployment Status:"
    echo
    docker compose -f docker-compose.prod.yml --env-file .env.prod ps
    echo
    log_info "To view logs: docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f"
    log_info "To stop services: docker compose -f docker-compose.prod.yml --env-file .env.prod down"
}

# Main deployment flow
main() {
    log_info "Starting ZTCP production deployment..."
    echo
    
    # Initialize status tracking
    init_status
    update_status "running" "Checking prerequisites..." '[]'
    
    check_prerequisites
    update_status "running" "Building Docker images..." '[{"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"}]'
    
    export DOCKER_REGISTRY="${DOCKER_REGISTRY:-local}"
    export IMAGE_TAG="${IMAGE_TAG:-latest}"
    
    # Track build progress
    update_status "running" "Building Docker images..." '[
      {"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"},
      {"title":"Building Images","status":"active","details":"Building backend, frontend, docs, and nginx images..."}
    ]'
    
    build_or_pull_images
    update_status "running" "Starting services..." '[
      {"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"},
      {"title":"Building Images","status":"complete","details":"All Docker images built successfully"},
      {"title":"Starting Services","status":"active","details":"Starting postgres, backend, frontend, docs, and nginx..."}
    ]'
    
    start_services
    update_status "running" "Running database migrations..." '[
      {"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"},
      {"title":"Building Images","status":"complete","details":"All Docker images built successfully"},
      {"title":"Starting Services","status":"complete","details":"All services started"},
      {"title":"Database Migrations","status":"active","details":"Running database migrations..."}
    ]'
    
    run_migrations
    update_status "running" "Verifying deployment..." '[
      {"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"},
      {"title":"Building Images","status":"complete","details":"All Docker images built successfully"},
      {"title":"Starting Services","status":"complete","details":"All services started"},
      {"title":"Database Migrations","status":"complete","details":"Migrations completed successfully"},
      {"title":"Verification","status":"active","details":"Verifying all services are healthy..."}
    ]'
    
    verify_deployment
    show_status
    
    # Final status
    update_status "complete" "Deployment completed successfully!" '[
      {"title":"Prerequisites Check","status":"complete","details":"All prerequisites verified"},
      {"title":"Building Images","status":"complete","details":"All Docker images built successfully"},
      {"title":"Starting Services","status":"complete","details":"All services started"},
      {"title":"Database Migrations","status":"complete","details":"Migrations completed successfully"},
      {"title":"Verification","status":"complete","details":"All services verified and healthy"}
    ]'
    
    log_info "Deployment completed successfully!"
    log_info "Your application should be available at: https://$(grep DOMAIN .env.prod 2>/dev/null | cut -d '=' -f2 || echo 'your-domain.com')"
    log_info "View deployment status at: https://$(grep DOMAIN .env.prod 2>/dev/null | cut -d '=' -f2 || echo 'your-domain.com')/status"
}

# Error handler to update status on failure (only for main function)
handle_error() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        update_status "error" "Deployment failed. Check logs for details." '[{"title":"Deployment Failed","status":"error","details":"An error occurred during deployment. Exit code: '"$exit_code"'."}]'
    fi
    exit $exit_code
}
trap handle_error ERR

# Run main function
main "$@"
