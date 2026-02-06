# Deploy (Local)

This folder contains the **Docker Compose** stack for local development: **PostgreSQL**. Use it to run the full ZTCP stack locally or as a reference for production-style deployments.

## Prerequisites

- **Docker** and **Docker Compose** (v2 or later). On macOS/Windows, Docker Desktop includes Compose.
- **Go** (for backend: migrate, seed, server).
- **Node.js** 20+ (for frontend dev server).
- For production: ensure Docker and Compose versions are supported and kept updated.

## Quick start

From the **repo root**: run `make setup`, then in two terminals run `make run-backend` and `make run-frontend`. Open [http://localhost:3000](http://localhost:3000). For full steps and options, see [Local deployment (full stack)](#local-deployment-full-stack) below.

### Ports reference

| Service    | Port (host) | Notes                    |
|-----------|-------------|--------------------------|
| PostgreSQL| 5432        | DB for backend           |
| gRPC      | 8080        | Backend API              |
| Frontend  | 3000        | Next.js dev server       |

---

## Production Deployment

For deploying to production (e.g., Digital Ocean droplet), see **[PRODUCTION.md](PRODUCTION.md)** for complete instructions.

The production deployment includes:
- Production Dockerfiles for backend and frontend
- Production docker-compose configuration (`docker-compose.prod.yml`)
- Nginx reverse proxy with TLS/SSL support
- Automated deployment scripts
- Security hardening (firewall, SSL certificates, secure defaults)
- Backup and restore procedures

**Quick start for production:**
1. Follow the [PRODUCTION.md](PRODUCTION.md) guide
2. Configure `.env.prod` from `.env.prod.example`
3. Run `./scripts/deploy.sh` to deploy

---

## Local deployment (full stack)

Follow these steps to run PostgreSQL, backend, and frontend locally.

### Step 1: Start PostgreSQL

From this directory (`deploy`):

```bash
docker compose up -d
```

Or from the repo root:

```bash
cd deploy
docker compose up -d
```

Verify all services are running:

```bash
docker compose ps
```

You should see `postgres` (Up). Optionally wait for Postgres to be ready before migrating:

```bash
docker compose exec postgres pg_isready -U ztcp
```

**Optional override:** To change ports or add env without editing committed files, copy `docker-compose.override.example.yml` to `docker-compose.override.yml` and edit. Compose merges it automatically when present. See the example file for commented samples.

### Step 2: Configure the backend and frontend

Copy the single env template to backend and optionally frontend:

```bash
cp deploy/.env.example backend/.env
cp deploy/.env.example frontend/.env   # optional, for frontend vars
```

Or from repo root run `make env` to create `backend/.env` and `frontend/.env` from `deploy/.env.example` only if they are missing (no overwrite).

The template already sets `DATABASE_URL=postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable` for local deploy. For auth (login/register): set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` in `backend/.env`. For local dev you can set `APP_ENV=development` and `OTP_RETURN_TO_CLIENT=true`. See [Backend README](../backend/README.md) and [Auth docs](../docs-site/docs/backend/auth.md) for key generation.

### Step 3: Run migrations

From the **backend** directory:

```bash
cd backend
./scripts/migrate.sh
```

Or explicitly: `./scripts/migrate.sh up`. Requires `DATABASE_URL` in `.env`.

### Step 4: Seed (optional)

From the **backend** directory:

```bash
./scripts/seed.sh
```

Seeds development data (e.g. dev user `dev@example.com` / `password123`) for local testing. Do not use in production.

### Step 5: Start the backend

From the **backend** directory:

```bash
go run ./cmd/server
```

The gRPC server listens on `:8080` by default.

### Step 6: Start the frontend

In a **separate terminal**, from the repo root:

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). The app talks to the backend at `localhost:8080` by default.

### Optional: Makefile (one-command setup)

From the **repo root**, a Makefile runs the full local setup in one go:

```bash
make setup
```

This ensures `backend/.env` and `frontend/.env` exist (from `deploy/.env.example` if missing), starts Docker (Postgres), waits for Postgres, runs migrations, and seeds (unless you set `SKIP_SEED=1`). Then start the backend and frontend in separate terminals:

```bash
make run-backend    # in one terminal
make run-frontend   # in another
```

Other targets: `make env` (copy env template if .env missing), `make up` (Docker only), `make down` (stop Docker), `make migrate`, `make seed`, `make install-frontend`.

---

## Files overview

| File | Role |
|------|-----|
| `docker-compose.yml` | Main Compose file; includes `postgres.yml`. Runs PostgreSQL. |
| `postgres.yml` | PostgreSQL service config (image, credentials, port, volume, healthcheck). Run alone: `docker compose -f postgres.yml up -d`. |
| `.env.example` | Single env template for local dev. Copy to `backend/.env` and optionally `frontend/.env`. |
| `docker-compose.override.example.yml` | Example for local overrides; copy to `docker-compose.override.yml` (gitignored) and edit. |
| `docker-compose.prod.yml` | Production Compose file with backend, frontend, nginx, and all infrastructure services. See [PRODUCTION.md](PRODUCTION.md). |
| `.env.prod.example` | Production environment template. Copy to `.env.prod` and configure for production deployment. |
| `PRODUCTION.md` | Complete production deployment guide for Digital Ocean. |
| `scripts/` | Deployment automation scripts: `deploy.sh`, `setup-ssl.sh`, `setup-firewall.sh`, `generate-jwt-keys.sh`, `backup.sh`, `restore.sh`. |
| `nginx/` | Nginx reverse proxy configuration for production (TLS, gRPC proxy). |

---

## Production-grade deployment

> **For complete production deployment guide**, see **[PRODUCTION.md](PRODUCTION.md)** which covers Digital Ocean deployment with TLS, security hardening, and automation.

The following section provides general guidance for hardening the local Compose stack for production use. For a full production deployment with containerized backend/frontend and reverse proxy, refer to PRODUCTION.md.

### Persistence

The Compose file defines named volumes so data survives restarts:

- `postgres_data` — PostgreSQL data

Ensure volumes are backed up or stored on durable storage.

### Resource limits

Add resource limits (and optionally requests) in the compose files to avoid one service starving others. Tune values based on load and host capacity.

### Security

- **Secrets**: Do not commit credentials. Use Docker secrets, env files outside the repo, or a secret manager and inject into Compose.

---

## Troubleshooting

| Issue | What to check |
|-------|----------------|
| **Port already in use** | Another process is using 5432 or 3000. Change the Compose `ports` or stop the conflicting service. Use `docker-compose.override.yml` for local port changes. |
| **Migrate fails** | Ensure Postgres is up: `docker compose ps` and `docker compose exec postgres pg_isready -U ztcp`. Check `DATABASE_URL` in `backend/.env` matches the Compose credentials (`postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable`). |
| **password authentication failed for user "root"** | Local Postgres uses user `ztcp`, not `root`. Set `DATABASE_URL=postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable` in `backend/.env`, or copy from `deploy/.env.example`. Running `make setup` will auto-fix `backend/.env` if it contains a `root` user URL. |
