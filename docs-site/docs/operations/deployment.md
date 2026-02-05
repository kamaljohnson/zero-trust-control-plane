---
title: Deployment and Operations
sidebar_label: Deployment
---

# Deployment and Operations

This document describes how to **run the backend and frontend** (locally and in production), environment variables, and optional observability. For the full local Docker Compose stack (PostgreSQL, OpenTelemetry Collector, Loki, Tempo, Prometheus, Grafana), see [deploy/README.md](../../../deploy/README.md).

**Audience**: Operators and developers setting up or deploying the zero-trust control plane.

## Overview

- **Backend**: Go gRPC server ([cmd/server](../../../backend/cmd/server)); requires PostgreSQL and (for auth) JWT keys.
- **Frontend**: Next.js app (BFF); calls backend gRPC via env-configured URL.
- **Optional**: Observability stack (OTLP collector, Loki, Tempo, Prometheus, Grafana) via [deploy/docker-compose.yml](../../../deploy/docker-compose.yml).

## Local development (Makefile)

From the repo root you can drive local setup entirely with the [Makefile](../../../Makefile). For step-by-step manual steps and Docker Compose details, see [deploy/README.md](../../../deploy/README.md).

**Prerequisites**: Docker and Docker Compose (v2+), Go, Node.js 20+.

### One-command setup

From the **repo root** run:

```bash
make setup
```

This runs in order:

1. **ensure-env** — Creates `backend/.env` and `frontend/.env` from [deploy/.env.example](../../../deploy/.env.example) if missing; if `backend/.env` exists but `DATABASE_URL` uses user `root`, it is replaced with the local deploy default (`ztcp`/`ztcp`).
2. **up** — Starts the Docker stack (PostgreSQL + observability) from `deploy/`.
3. **wait-postgres** — Waits for Postgres to be ready (up to 30s).
4. **migrate** — Runs `backend/scripts/migrate.sh up` (requires `DATABASE_URL` in `backend/.env`).
5. **seed** — Runs `backend/scripts/seed.sh` (inserts dev users, e.g. `dev@example.com`). Skip with `SKIP_SEED=1 make setup`.
6. **configure-grafana** — Waits for Grafana and notes that datasources and the ZTCP Telemetry dashboard are provisioned at http://localhost:3002.

### Starting the app after setup

In one terminal run `make run-backend`; in another run `make run-frontend`. Open http://localhost:3000. Optionally run the docs site with `make run-docs` (serves at http://localhost:3001).

To stop the Docker stack later: `make down`.

### Makefile targets reference

| Target | Description |
|--------|-------------|
| `setup` | Full local setup (ensure-env, up, wait-postgres, migrate, seed, configure-grafana). |
| `env` | Copy [deploy/.env.example](../../../deploy/.env.example) to `backend/.env` and `frontend/.env` if missing (no overwrite). |
| `ensure-env` | Same as `env` plus fix `DATABASE_URL` when it uses user `root`. |
| `up` | Start Docker stack from `deploy/`. |
| `down` | Stop Docker stack. |
| `wait-postgres` | Block until Postgres is ready. |
| `wait-grafana` | Block until Grafana health responds (optional). |
| `configure-grafana` | Ensure Grafana is ready; used by `setup`. |
| `migrate` | Run backend migrations (`ensure-env` then `backend/scripts/migrate.sh up`). |
| `seed` | Run backend seed script (`ensure-env` then `backend/scripts/seed.sh`). |
| `run-backend` | Run backend gRPC server (`go run ./cmd/server` in backend). |
| `run-frontend` | Install frontend deps and run Next.js dev server. |
| `install-frontend` | `npm install` in frontend. |
| `install-docs` | `npm install` in docs-site. |
| `run-docs` | Install docs deps and run Docusaurus on port 3001. |

**Variables**: The Makefile uses `BACKEND_DIR=backend`, `DEPLOY_DIR=deploy`, `DOCS_DIR=docs-site`, `FRONTEND_DIR=frontend`. These are rarely overridden.

## Running the server

From the **backend** directory (or from repo root use `make run-backend`):

```bash
go run ./cmd/server
```

Or from repo root: `make run-backend`. The gRPC server listens on **:8080** by default (`GRPC_ADDR` in env).

**Before first run**: Set `DATABASE_URL` and run migrations (`./scripts/migrate.sh` from backend). For login/register, set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY`. See [Backend auth](../backend/auth) and [backend/.env.example](../../../backend/.env.example).

## Running the frontend

From the **frontend** directory (or from repo root use `make run-frontend`):

```bash
npm install
npm run dev
```

Or from repo root: `make run-frontend`. The app is served at **http://localhost:3000** and calls the backend at the URL set in `BACKEND_GRPC_URL` (default `localhost:8080`).

## Environment checklist

### Backend ([backend/.env.example](../../../backend/.env.example))

| Variable | Required | Notes |
|----------|----------|--------|
| `GRPC_ADDR` | No | Listen address (default `:8080`) |
| `DATABASE_URL` | Yes (for full features) | Postgres DSN |
| `JWT_PRIVATE_KEY` | Yes (for auth) | PEM or path to file |
| `JWT_PUBLIC_KEY` | Yes (for auth) | PEM or path to file |
| `JWT_ISSUER`, `JWT_AUDIENCE` | No | Defaults: ztcp-auth, ztcp-api |
| `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL` | No | e.g. 15m, 168h |
| `SMS_LOCAL_*` | For SMS OTP | PoC MFA |
| `APP_ENV`, `OTP_RETURN_TO_CLIENT` | No | Dev OTP; must not be production when OTP_RETURN_TO_CLIENT=true |
| `OTEL_EXPORTER_OTLP_*` | No | Optional OTLP export |

### Frontend ([frontend/.env.example](../../../frontend/.env.example))

| Variable | Required | Notes |
|----------|----------|--------|
| `BACKEND_GRPC_URL` | Yes | Backend gRPC host:port (e.g. `localhost:8080`) |
| `NEXT_PUBLIC_DOCS_URL` | No | When set, the "Docs" link in the header and on the home alert points to this URL (e.g. external docs site). If unset, the link may point to a default or be hidden depending on implementation. |
| `DEV_OTP_ENABLED`, `NEXT_PUBLIC_DEV_OTP_ENABLED` | No | Dev-only OTP; disable in production |

Local one-shot setup: from repo root run `make setup` (creates `.env` from [deploy/.env.example](../../../deploy/.env.example) if missing, starts Docker, migrates, optional seed). Then run `make run-backend` and `make run-frontend` in two terminals.

## Production

- **Environment**: Set `APP_ENV=production`; do **not** set `OTP_RETURN_TO_CLIENT=true`. Use strong `JWT_PRIVATE_KEY`/`JWT_PUBLIC_KEY` and a secure `DATABASE_URL`.
- **Migrations**: Run migrations before or during deployment (e.g. `backend/scripts/migrate.sh up`).
- **TLS**: For production, expose the gRPC server behind TLS (e.g. reverse proxy or server-side TLS). Configure `BACKEND_GRPC_URL` on the frontend to use the correct scheme and host.
- **Docker/Kubernetes**: Use [deploy/docker-compose.yml](../../../deploy/docker-compose.yml) as a reference for Postgres and observability; the backend and frontend can be run in containers or on VMs with the same env and migration steps.
