# Production Deployment Guide - Digital Ocean

This guide covers deploying the Zero Trust Control Plane (ZTCP) to a Digital Ocean droplet in production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Droplet Setup](#droplet-setup)
3. [Initial Server Configuration](#initial-server-configuration)
4. [SSL Certificate Setup](#ssl-certificate-setup)
5. [Environment Configuration](#environment-configuration)
6. [Deployment](#deployment)
7. [Post-Deployment](#post-deployment)
8. [Backup and Restore](#backup-and-restore)
9. [Monitoring and Maintenance](#monitoring-and-maintenance)
10. [Troubleshooting](#troubleshooting)
11. [Security Checklist](#security-checklist)
12. [CI/CD (GitHub Actions)](#cicd-github-actions)

## Prerequisites

### Digital Ocean Account
- Active Digital Ocean account
- Domain name with DNS access (for SSL certificates)

### Droplet Requirements

**Minimum Recommended:**
- **4GB RAM, 2 vCPUs** - Suitable for small deployments
- **80GB SSD** - For application and data storage

**Recommended for Production:**
- **8GB RAM, 4 vCPUs** - Better performance and headroom
- **160GB SSD** - More storage for logs and backups

### Software Requirements
- Ubuntu 22.04 LTS or later (recommended)
- Docker Engine 24.0+ 
- Docker Compose v2.20+
- OpenSSL (for JWT key generation)
- UFW firewall (for security)

## Droplet Setup

### 1. Create Droplet

1. Log into Digital Ocean dashboard
2. Create a new Droplet:
   - **Image**: Ubuntu 22.04 LTS
   - **Size**: 4GB RAM / 2 vCPUs (minimum) or 8GB RAM / 4 vCPUs (recommended)
   - **Region**: Choose closest to your users
   - **Authentication**: SSH keys (recommended) or root password
   - **Hostname**: `ztcp-production` (or your preferred name)

### 2. Configure DNS

Point your domain to the droplet's IP address:

```
A Record: @ → <droplet-ip>
A Record: www → <droplet-ip>  (optional)
```

Wait for DNS propagation (can take up to 48 hours, usually much faster).

### 3. Initial SSH Access

```bash
ssh root@<droplet-ip>
# or
ssh root@yourdomain.com
```

## Initial Server Configuration

### 1. Update System

```bash
apt-get update
apt-get upgrade -y
```

### 2. Install Docker

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose v2
apt-get install -y docker-compose-plugin

# Verify installation
docker --version
docker compose version
```

### 3. Create Non-Root User (Optional but Recommended)

```bash
# Create user
adduser ztcp
usermod -aG docker ztcp
usermod -aG sudo ztcp

# Switch to new user
su - ztcp
```

### 4. Clone Repository

```bash
# Install git if needed
apt-get install -y git

# Clone repository
cd /opt
git clone <your-repo-url> zero-trust-control-plane
cd zero-trust-control-plane/deploy
```

### 5. Configure Firewall

```bash
# Run firewall setup script
sudo ./scripts/setup-firewall.sh
```

This configures UFW to allow:
- SSH (port 22)
- HTTP (port 80) - for Let's Encrypt
- HTTPS (port 443) - for application

**Important**: Ensure SSH access works before disconnecting!

## SSL Certificate Setup

### 1. Generate SSL Certificates

```bash
# Run SSL setup script (requires root/sudo)
sudo ./scripts/setup-ssl.sh
```

The script will:
- Install certbot
- Obtain Let's Encrypt certificate for your domain
- Configure auto-renewal via cron

### 2. Verify Certificates

```bash
# Check certificate files
ls -la nginx/ssl/

# Should see:
# - fullchain.pem
# - privkey.pem
# - cert.pem
# - chain.pem
```

### 3. Test Certificate Renewal

```bash
sudo certbot renew --dry-run
```

## Environment Configuration

### 1. Generate JWT Keys

```bash
# Generate RSA key pair for JWT authentication
./scripts/generate-jwt-keys.sh
```

This creates keys in `keys/` directory and displays configuration for `.env.prod`.

### 2. Configure Production Environment

```bash
# Copy environment template
cp .env.prod.example .env.prod

# Edit with your production values
nano .env.prod  # or use your preferred editor
```

**Required Configuration:**

```bash
# Domain name
DOMAIN=yourdomain.com

# Database credentials (CHANGE THESE!)
POSTGRES_USER=ztcp
POSTGRES_PASSWORD=<strong-secure-password>
POSTGRES_DB=ztcp
# DATABASE_URL must use the same user/password as POSTGRES_USER/POSTGRES_PASSWORD. Production uses SSL (run ./scripts/generate-postgres-ssl.sh).
DATABASE_URL=postgres://ztcp:<strong-secure-password>@postgres:5432/ztcp?sslmode=require

# JWT keys (from generate-jwt-keys.sh)
JWT_PRIVATE_KEY=<private-key-content-or-path>
JWT_PUBLIC_KEY=<public-key-content-or-path>

# Application environment
APP_ENV=production
OTP_RETURN_TO_CLIENT=false

# Docs run on the droplet at https://<DOMAIN>/docs. Set NEXT_PUBLIC_DOCS_URL so the frontend "Docs" header link points there.
NEXT_PUBLIC_DOCS_URL=https://yourdomain.com/docs

# Grafana admin password
GF_SECURITY_ADMIN_PASSWORD=<strong-secure-password>
GF_AUTH_ANONYMOUS_ENABLED=false
GF_AUTH_DISABLE_LOGIN_FORM=false
```

### 3. Secure Environment File

```bash
# Set restrictive permissions
chmod 600 .env.prod
```

### 4. PostgreSQL SSL (production)

Production uses PostgreSQL with SSL (`sslmode=require`). Before first deploy:

1. Generate a self-signed certificate for Postgres (writes to `deploy/postgres-ssl/`):

   ```bash
   cd deploy
   ./scripts/generate-postgres-ssl.sh
   ```

2. Ensure `DATABASE_URL` in `.env.prod` uses `sslmode=require` (see Required Configuration). The postgres container is started with SSL enabled and the certs from `postgres-ssl/`; backend and migrations connect with SSL.

## Deployment

### 1. Build and Deploy

```bash
# Run deployment script
./scripts/deploy.sh
```

The script will:
- Check prerequisites
- Build Docker images (backend, frontend, docs, nginx)
- Run database migrations
- Start all services (including the docs site at `/docs`)
- Verify deployment health

### 2. Verify Deployment

```bash
# Check service status
docker compose -f docker-compose.prod.yml --env-file .env.prod ps

# View logs
docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f

# Check specific service logs
docker compose -f docker-compose.prod.yml --env-file .env.prod logs backend
docker compose -f docker-compose.prod.yml --env-file .env.prod logs frontend
docker compose -f docker-compose.prod.yml --env-file .env.prod logs nginx
```

### 3. Test Application

```bash
# Test HTTPS endpoint
curl -k https://yourdomain.com/health

# Test frontend
curl -I https://yourdomain.com

# Test backend gRPC (via nginx)
# Use a gRPC client or the frontend UI
```

## Post-Deployment

### 1. Initial Database Setup

If you need to seed initial data (development users, etc.):

```bash
# Run seed script (ONLY for initial setup, not production data)
cd ../backend
docker compose -f ../deploy/docker-compose.prod.yml --env-file ../deploy/.env.prod \
  run --rm backend /app/ztcp-server seed
```

**Warning**: The seed script creates development users. Remove or modify for production.

### 2. Access Grafana

1. Navigate to `https://yourdomain.com:3002` (or configure nginx to proxy Grafana)
2. Login with admin credentials from `.env.prod`
3. Verify datasources are configured:
   - Prometheus: `http://prometheus:9090`
   - Loki: `http://loki:3100`
   - Tempo: `http://tempo:3200`

### 3. Configure Monitoring Alerts

Set up alerts in Grafana for:
- High error rates
- Database connection failures
- Service downtime
- Disk space usage
- Memory/CPU usage

## Backup and Restore

### Automated Backups

Set up cron job for regular backups:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /opt/zero-trust-control-plane/deploy/scripts/backup.sh
```

### Manual Backup

```bash
# Run backup script
./scripts/backup.sh
```

Backups are stored in `backups/` directory with timestamps.

### Restore from Backup

```bash
# List available backups
ls -lh backups/

# Restore specific backup
./scripts/restore.sh backups/ztcp_backup_YYYYMMDD_HHMMSS.tar.gz
```

**Warning**: Restore will overwrite existing data. Ensure you have a current backup before restoring.

## Monitoring and Maintenance

### Health Checks

```bash
# Check all services
docker compose -f docker-compose.prod.yml --env-file .env.prod ps

# Backend is gRPC-only; health is TCP to 8080 (see docker-compose.prod.yml healthcheck)
docker compose -f docker-compose.prod.yml --env-file .env.prod exec backend nc -z localhost 8080 && echo "backend OK"
```

### Log Management

```bash
# View all logs
docker compose -f docker-compose.prod.yml --env-file .env.prod logs --tail=100

# Follow logs
docker compose -f docker-compose.prod.yml --env-file .env.prod logs -f

# View logs for specific service
docker compose -f docker-compose.prod.yml --env-file .env.prod logs backend
```

### Updates and Upgrades

```bash
# Pull latest code
cd /opt/zero-trust-control-plane
git pull

# Rebuild and restart services
cd deploy
./scripts/deploy.sh
```

### SSL Certificate Renewal

Certificates auto-renew via cron. Verify renewal:

```bash
# Check renewal status
sudo certbot certificates

# Test renewal
sudo certbot renew --dry-run
```

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker compose -f docker-compose.prod.yml --env-file .env.prod logs

# Check Docker status
docker ps -a
docker compose -f docker-compose.prod.yml --env-file .env.prod ps
```

### Backend unhealthy or "P8R variable is not set"

If the backend container fails to become healthy or you see many `The "P8R" variable is not set` warnings:

1. **Check backend logs** – The backend may be crashing on startup (e.g. invalid JWT keys):
   ```bash
   docker compose -f docker-compose.prod.yml --env-file .env.prod logs backend
   ```

2. **Fix JWT key values in `.env.prod`** – Docker Compose treats `$VAR` as a variable. If your `JWT_PRIVATE_KEY` or `JWT_PUBLIC_KEY` (inline PEM) contain a `$` (e.g. in base64), that part is replaced and the key is corrupted. Escape every literal `$` as `$$` in those values in `.env.prod`. For example, if you see `$P8R` in the key, change it to `$$P8R`.

3. **Use key file paths instead of inline PEM** – To avoid escaping, set `JWT_PRIVATE_KEY=/path/to/priv.pem` and `JWT_PUBLIC_KEY=/path/to/pub.pem` and mount the keys into the backend container, or ensure the inline PEM has no `$` sequences.

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker compose -f docker-compose.prod.yml --env-file .env.prod exec postgres pg_isready

# Test connection
docker compose -f docker-compose.prod.yml --env-file .env.prod exec postgres psql -U ztcp -d ztcp -c "SELECT 1;"
```

### SSL Certificate Issues

```bash
# Check certificate files
ls -la nginx/ssl/

# Verify certificate
sudo openssl x509 -in nginx/ssl/fullchain.pem -text -noout

# Check nginx configuration
docker compose -f docker-compose.prod.yml --env-file .env.prod exec nginx nginx -t
```

### High Resource Usage

```bash
# Check resource usage
docker stats

# Check disk usage
df -h
docker system df

# Clean up unused resources
docker system prune -a
```

### Port Conflicts

```bash
# Check what's using ports
sudo netstat -tulpn | grep -E ':(80|443|5432|8080|3000)'

# Stop conflicting services or adjust docker-compose ports
```

## Security Checklist

- [ ] Firewall configured (UFW) - only ports 22, 80, 443 open
- [ ] Strong database passwords set
- [ ] JWT keys generated and secured (permissions 600)
- [ ] `.env.prod` file permissions set to 600
- [ ] Grafana anonymous auth disabled
- [ ] Grafana admin password set
- [ ] SSL certificates configured and auto-renewal enabled
- [ ] Regular backups configured
- [ ] System updates applied
- [ ] Non-root user created (optional but recommended)
- [ ] SSH key authentication enabled (disable password auth)
- [ ] Fail2ban installed (optional but recommended)
- [ ] Log rotation configured
- [ ] Monitoring alerts configured

### Additional Security Recommendations

1. **Fail2ban**: Install to prevent brute force attacks
   ```bash
   apt-get install -y fail2ban
   ```

2. **SSH Hardening**: Edit `/etc/ssh/sshd_config`:
   - Disable root login: `PermitRootLogin no`
   - Change SSH port (optional): `Port 2222`
   - Restart SSH: `systemctl restart sshd`

3. **Regular Updates**: Set up automatic security updates
   ```bash
   apt-get install -y unattended-upgrades
   ```

4. **Backup Encryption**: Encrypt backups before storing off-server
   ```bash
   # Encrypt backup
   gpg --encrypt --recipient your@email.com backup.tar.gz
   ```

## CI/CD (GitHub Actions)

The repository includes a GitHub Actions workflow that automatically tests, builds Docker images, pushes them to a container registry, and deploys to your Digital Ocean droplet on every push to the `main` branch.

### Deployment flow

1. **Tests** – Backend and frontend tests run; both must pass.
2. **Build and push** – Docker images for backend, frontend, and nginx are built and pushed to the registry (GitHub Container Registry by default, or Docker Hub if configured).
3. **Deploy** – The workflow SSHs to the droplet, pulls the latest images, runs database migrations, and restarts services with `docker compose`.

### Required GitHub secrets

Configure these in **Settings → Secrets and variables → Actions**:

| Secret | Description |
|--------|-------------|
| `DROPLET_HOST` | Droplet IP address or hostname |
| `DROPLET_USER` | SSH user (e.g. `root` or `deploy`) |
| `DROPLET_SSH_KEY` | Full private SSH key (PEM) for the droplet |
| `DROPLET_PORT` | (Optional) SSH port; default is 22 |

**Container registry (choose one):**

- **GitHub Container Registry (default)**  
  No extra secrets. The workflow uses `GITHUB_TOKEN` and pushes to `ghcr.io/<owner>/<repo>/ztcp-backend`, etc. Make the package visible (e.g. public) or ensure the droplet can authenticate (see below).

- **Docker Hub**  
  Add:
  - `DOCKER_USERNAME` – Docker Hub username  
  - `DOCKER_PASSWORD` – Docker Hub password or access token  

  Images will be pushed to `docker.io/<DOCKER_USERNAME>/ztcp-backend`, etc.

### Droplet setup for CI/CD

1. **Clone the repo** on the droplet (e.g. `/opt/zero-trust-control-plane`) and ensure the deploy path is the same as in the workflow (or set the `DEPLOY_REPO_DIR` variable in the workflow).
2. **One-time setup** on the droplet: create `.env.prod`, SSL certs, JWT keys, and run initial deploy (see [Deployment](#deployment)).
3. **Registry login (if using GHCR or private Docker Hub)**  
   So the droplet can pull images, log in once:
   - **GHCR**: `echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin` (use a Personal Access Token with `read:packages`).
   - **Docker Hub**: `docker login` with your credentials.  
   Store the token or use a deploy token; ensure the user that runs `docker compose` can pull (e.g. run as the same user that logged in).
4. **Migrations** – The workflow runs migrations from the host. Install either **Go** (so `backend/scripts/migrate.sh` can run) or the **golang-migrate** CLI. Ensure `DATABASE_URL` in `.env.prod` uses `postgres:5432` for Compose; the workflow temporarily substitutes `localhost:5432` when running migrations on the host.
5. **SSH key** – Use a dedicated deploy key or a user with minimal permissions. Add the **public** key to the droplet (e.g. `~/.ssh/authorized_keys` for `DROPLET_USER`).

### Optional: GitHub variables

| Variable | Description |
|----------|-------------|
| `DEPLOY_REPO_DIR` | Path to the repo on the droplet (default: `/opt/zero-trust-control-plane`) |
| `NEXT_PUBLIC_DEV_OTP_ENABLED` | When `true`, BFF returns OTP in response for MFA (no SMS). Set to match `.env.prod`; default in CI is `false` if unset. |
| `NEXT_PUBLIC_DOCS_URL` | Docs base URL for the frontend "Docs" link (e.g. `https://yourdomain.com/docs`). Baked into the frontend image at build time. |
| `NEXT_PUBLIC_GRAFANA_URL` | Grafana base URL for the dashboard "Open Grafana" telemetry link (e.g. `https://yourdomain.com/grafana`). Baked into the frontend image at build time. |

Set in **Settings → Secrets and variables → Actions → Variables**. If you deploy via CI (pull from registry), set `NEXT_PUBLIC_DOCS_URL` and `NEXT_PUBLIC_GRAFANA_URL` so the frontend build gets these values; otherwise the telemetry page will show the "Set NEXT_PUBLIC_GRAFANA_URL" placeholder.

### Manual deployment trigger

You can run the workflow manually:

1. **Actions** → **Deploy to Digital Ocean** → **Run workflow** → **Run workflow**.

### Troubleshooting CI/CD

| Issue | What to check |
|-------|----------------|
| **Deploy job: “Permission denied (publickey)”** | `DROPLET_SSH_KEY` is the full private key (including `-----BEGIN ... KEY-----`). The matching public key must be in `~/.ssh/authorized_keys` for `DROPLET_USER` on the droplet. |
| **Pull fails on droplet** | Ensure the droplet is logged in to the registry (`docker login ghcr.io` or `docker login`). For GHCR, the package must be public or the token must have `read:packages`. |
| **Migrations fail** | Install Go or golang-migrate on the droplet. Ensure `DATABASE_URL` in `.env.prod` uses the **same** `POSTGRES_USER` and `POSTGRES_PASSWORD` as the postgres service (same literal values). The script rewrites the host to `localhost` and `sslmode` to `disable` for host-run migrations. |
| **password authentication failed for user "ztcp"** | The password in `DATABASE_URL` does not match `POSTGRES_PASSWORD` in `.env.prod`. Set both to the same value (and ensure `DATABASE_URL` user matches `POSTGRES_USER`). |
| **Tests fail** | Fix failing backend/frontend tests locally; the workflow runs the same test commands. |
| **Build fails** | Check Dockerfile paths and build context in `.github/workflows/deploy.yml`. Run `docker build` locally for backend, frontend, and nginx. |
| **Services not updating** | On the droplet, run `docker compose -f deploy/docker-compose.prod.yml --env-file deploy/.env.prod ps` and check logs. Ensure `IMAGE_TAG`/`DOCKER_REGISTRY` are set by the workflow (they are passed in the SSH script). |

### Rollback after a bad deploy

To run a previous image by commit SHA:

1. On the droplet, set the image tag and restart:
   ```bash
   cd /opt/zero-trust-control-plane/deploy
   export IMAGE_TAG=<previous-commit-sha>
   export DOCKER_REGISTRY=ghcr.io/YOUR_ORG/zero-trust-control-plane  # or your registry
   docker compose -f docker-compose.prod.yml --env-file .env.prod pull
   docker compose -f docker-compose.prod.yml --env-file .env.prod up -d
   ```
2. Or revert the commit on `main` and push; the workflow will deploy that commit.

## Support and Resources

- **Backend Documentation**: `../backend/README.md`
- **Frontend Documentation**: `../frontend/README.md`
- **Local Deployment**: See `README.md` in this directory
- **Docker Compose Docs**: https://docs.docker.com/compose/
- **Digital Ocean Docs**: https://docs.digitalocean.com/

## Rollback Procedure

If deployment fails or issues occur:

1. **Stop services**:
   ```bash
   docker compose -f docker-compose.prod.yml --env-file .env.prod down
   ```

2. **Restore from backup** (if needed):
   ```bash
   ./scripts/restore.sh backups/ztcp_backup_<timestamp>.tar.gz
   ```

3. **Revert code changes**:
   ```bash
   cd /opt/zero-trust-control-plane
   git checkout <previous-commit>
   ```

4. **Redeploy**:
   ```bash
   cd deploy
   ./scripts/deploy.sh
   ```

---

**Last Updated**: 2026-02-04

For issues or questions, refer to the troubleshooting section or check service logs.
