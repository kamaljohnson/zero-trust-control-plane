# Deploy (Local & Observability)

This folder contains the **Docker Compose** stack for local development: **PostgreSQL** and the **observability stack** (OpenTelemetry Collector, Loki, Tempo, Prometheus, Grafana). Use it to run the full ZTCP stack locally or as a reference for production-style deployments.

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
| OTLP      | 4317        | Collector gRPC           |
| Loki      | 3100        | Logs                     |
| Tempo     | 3200        | Traces                   |
| Prometheus| 9090        | Metrics                  |
| Grafana   | 3002        | Observability UI         |
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
- Comprehensive monitoring setup

**Quick start for production:**
1. Follow the [PRODUCTION.md](PRODUCTION.md) guide
2. Configure `.env.prod` from `.env.prod.example`
3. Run `./scripts/deploy.sh` to deploy

---

## Local deployment (full stack)

Follow these steps to run PostgreSQL, the telemetry stack, backend, and frontend locally.

### Step 1: Start PostgreSQL and telemetry

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

You should see `postgres`, `otelcol`, `tempo`, `loki`, `prometheus`, and `grafana` (all Up). Optionally wait for Postgres to be ready before migrating:

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

The template already sets `DATABASE_URL=postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable` and `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317` for local deploy. For auth (login/register): set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` in `backend/.env`. For local dev you can set `APP_ENV=development` and `OTP_RETURN_TO_CLIENT=true`. See [Backend README](../backend/README.md) and [Auth docs](../docs-site/docs/backend/auth.md) for key generation.

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

### Step 7 (optional): Configure Grafana dashboard

1. Open **Grafana**: http://localhost:3002.
2. If not already done: add datasources (Prometheus `http://prometheus:9090`, Loki `http://loki:3100`, Tempo `http://tempo:3200`) as described in [Step 5: Open Grafana and add datasources](#step-5-open-grafana-and-add-datasources) in the Observability stack section.
3. The ZTCP Telemetry dashboard is auto-provisioned when using the Compose stack; open **Dashboards** and select **ZTCP Telemetry** (if prompted for a datasource, choose Loki). To import manually: **Dashboards → New → Import → Upload JSON file** and choose `deploy/grafana/dashboards/ztcp-telemetry-dashboard.json` from the repo.

### Optional: Makefile (one-command setup)

From the **repo root**, a Makefile runs the full local setup in one go:

```bash
make setup
```

This ensures `backend/.env` and `frontend/.env` exist (from `deploy/.env.example` if missing), starts Docker (Postgres + telemetry), waits for Postgres, runs migrations, and seeds (unless you set `SKIP_SEED=1`). Then start the backend and frontend in separate terminals:

```bash
make run-backend    # in one terminal
make run-frontend   # in another
```

Other targets: `make env` (copy env template if .env missing), `make up` (Docker only), `make down` (stop Docker), `make migrate`, `make seed`, `make install-frontend`.

---

## Files overview

| File | Role |
|------|-----|
| `docker-compose.yml` | Main Compose file; includes `postgres.yml` and `observability.yml`. Runs PostgreSQL + OpenTelemetry Collector (contrib), Tempo, Loki, Prometheus, and Grafana. |
| `postgres.yml` | PostgreSQL service config (image, credentials, port, volume, healthcheck). Run alone: `docker compose -f postgres.yml up -d`. |
| `observability.yml` | Observability services: otelcol, tempo, loki, prometheus, grafana + volumes. Run alone: `docker compose -f observability.yml up -d`. |
| `.env.example` | Single env template for local dev. Copy to `backend/.env` and optionally `frontend/.env`. |
| `docker-compose.override.example.yml` | Example for local overrides; copy to `docker-compose.override.yml` (gitignored) and edit. |
| `docker-compose.prod.yml` | Production Compose file with backend, frontend, nginx, and all infrastructure services. See [PRODUCTION.md](PRODUCTION.md). |
| `.env.prod.example` | Production environment template. Copy to `.env.prod` and configure for production deployment. |
| `PRODUCTION.md` | Complete production deployment guide for Digital Ocean. |
| `scripts/` | Deployment automation scripts: `deploy.sh`, `setup-ssl.sh`, `setup-firewall.sh`, `generate-jwt-keys.sh`, `backup.sh`, `restore.sh`. |
| `nginx/` | Nginx reverse proxy configuration for production (TLS, gRPC proxy). |
| `otelcol-config.yaml` | Collector config. Use as-is with Docker Compose (endpoints use service names `loki:3100`, `tempo:4317`). For host/standalone runs, copy and set Loki to `http://localhost:3100/loki/api/v1/push` and Tempo to `localhost:4317`. |
| `tempo.yaml` | Tempo config: OTLP gRPC/HTTP receivers, local storage, 7-day retention. |
| `prometheus.yml` | Prometheus scrape config for the collector's metrics endpoint (`otelcol:8889`). |

---

## Observability stack (step-by-step)

Use this section if you only need to run the telemetry stack (e.g. you already have Postgres elsewhere).

### Step 1: Start the observability stack

From the **deploy** directory:

```bash
cd deploy
docker compose -f observability.yml up -d
```

Or start the full stack (Postgres + observability) with `docker compose up -d`.

Verify containers are running:

```bash
docker compose ps
```

You should see `otelcol`, `tempo`, `loki`, `prometheus`, and `grafana` (all Up).

### Step 2: Point the ZTCP gRPC server at the collector

Set the OTLP endpoint to the collector's gRPC port (default `4317` on the host):

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

If the server runs in another container on the same Docker network, use:

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://otelcol:4317
```

Optional: set the service name (default is `ztcp-grpc`):

```bash
export OTEL_SERVICE_NAME=ztcp-grpc
```

### Step 3: Run the ZTCP gRPC server

From the backend root (with `.env` or env vars for DB/auth as needed):

```bash
go run ./cmd/server
```

Send traffic or call TelemetryService (`EmitTelemetryEvent` / `BatchEmitTelemetry`) so the server emits traces, metrics, and logs to the collector.

### Step 4: Verify the pipeline

- **Collector**: `docker compose logs otelcol` — no repeated errors about Loki or Tempo.
- **Prometheus**: Open http://localhost:9090 → Status → Targets. The `otelcol` job should be **UP**.
- **Loki**: Open http://localhost:3100/ready (expect 200).
- **Tempo**: Open http://localhost:3200/ready (expect 200).

### Step 5: Open Grafana and add datasources

1. Open **Grafana**: http://localhost:3002 (Compose uses anonymous admin for local use).
2. Add datasources (Connections → Data sources → Add data source):
   - **Prometheus**: URL `http://prometheus:9090` (from inside Docker). If Grafana runs on the host, use `http://localhost:9090`.
   - **Loki**: URL `http://loki:3100` (or `http://localhost:3100` from host).
   - **Tempo**: URL `http://tempo:3200` (or `http://localhost:3200` from host).

Save & test each.

3. **ZTCP telemetry dashboard**: The stack auto-provisions the ZTCP Telemetry dashboard (from `deploy/grafana/dashboards/ztcp-telemetry-dashboard.json`). Open Dashboards and select **ZTCP Telemetry**; if prompted for a datasource, choose Loki. To import manually instead: Dashboards → New → Import → Upload JSON file → **`deploy/grafana/dashboards/ztcp-telemetry-dashboard.json`**, then select the Loki datasource. The dashboard shows telemetry logs, gRPC request/error rates, and related panels for the ZTCP backend.

You can also build custom dashboards or use Explore to query logs (LogQL), metrics (PromQL), and traces (Tempo).

### Step 6: Running the collector on the host (optional)

If you run the collector binary (e.g. `otelcol-contrib`) on the same machine as Loki and Tempo instead of using Docker Compose, copy `otelcol-config.yaml` and set the Loki exporter endpoint to `http://localhost:3100/loki/api/v1/push` and the Tempo exporter endpoint to `localhost:4317`. Run `otelcol-contrib --config /path/to/your-copy.yaml`. The default `otelcol-config.yaml` uses service names (`loki`, `tempo`) for Compose and is not suitable for host-only runs without editing.

### Step 7: Optional links

- **Telemetry overview**: [Telemetry doc](../docs-site/docs/backend/telemetry.md) and [Backend README](../backend/README.md) (Configuration → Telemetry) for server-side config.
- **Grafana dashboard**: ZTCP telemetry dashboard: import [deploy/grafana/dashboards/ztcp-telemetry-dashboard.json](grafana/dashboards/ztcp-telemetry-dashboard.json) as in Step 5 above.

---

## Production-grade deployment

> **For complete production deployment guide**, see **[PRODUCTION.md](PRODUCTION.md)** which covers Digital Ocean deployment with TLS, security hardening, and automation.

The following section provides general guidance for hardening the local Compose stack for production use. For a full production deployment with containerized backend/frontend and reverse proxy, refer to PRODUCTION.md.

### Persistence

The Compose file defines named volumes so data survives restarts:

- `postgres_data` — PostgreSQL data
- `prometheus_data` — Prometheus TSDB
- `grafana_data` — Grafana dashboards and settings
- `tempo_data` — Tempo blocks and WAL
- `loki_data` — Loki data (if the Loki image uses the mounted path)

Ensure volumes are backed up or stored on durable storage. For Loki, confirm the image's config uses the path you mount (e.g. `/loki`); adjust the Loki command or config if needed.

### Resource limits

Add resource limits (and optionally requests) in the compose files to avoid one service starving others. Example:

```yaml
services:
  otelcol:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
  tempo:
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: "1"
  loki:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
  prometheus:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: "1"
  grafana:
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: "0.25"
```

Tune values based on load and host capacity.

### Security

- **Grafana**: Disable anonymous auth in production. Set a strong admin password and/or use OAuth/LDAP. Remove or override `GF_AUTH_ANONYMOUS_ENABLED` and `GF_AUTH_DISABLE_LOGIN_FORM`.
- **OTLP and backends**: The collector, Loki, Tempo, and Prometheus in this setup do not enable authentication. Restrict network access (e.g. run the stack in a private network, put the collector behind a reverse proxy, or use firewall rules). For TLS, configure the collector's OTLP receiver and the backend exporters (Loki, Tempo) and Prometheus scrape accordingly.
- **Secrets**: Do not commit credentials. Use Docker secrets, env files outside the repo, or a secret manager and inject into Compose.

### External backends

For production you may use managed or external Loki, Tempo, or Prometheus instead of the Compose services:

- **Collector**: Copy `otelcol-config.yaml` to a new file (e.g. `otelcol-config.prod.yaml`) and set the Loki and Tempo exporter endpoints to your external URLs. Use that config when running the collector (Compose or otherwise).
- **Prometheus**: If you use a central Prometheus, add a scrape job for the collector's metrics endpoint (host:8889 or `otelcol:8889`). No change to the collector config.

### High availability

This Compose file runs a single instance of each service. For HA:

- Run multiple replicas of the collector behind a load balancer; point ZTCP at the LB.
- Use Grafana's documentation for HA and shared storage (e.g. for Grafana).
- Scale Loki, Tempo, and Prometheus per their official docs (e.g. Loki in distributed mode, Tempo with multiple ingesters, Prometheus federation or remote write).

For large or critical deployments, consider Kubernetes and official Helm charts or operators for the collector, Loki, Tempo, and Prometheus.

---

## Troubleshooting

| Issue | What to check |
|-------|----------------|
| **Port already in use** | Another process is using 5432, 4317, 4318, 8889, 3100, 3200, 9090, 3002, or 3000. Only the OpenTelemetry Collector exposes 4317 on the host; Tempo's OTLP port is internal (collector reaches it via Docker network). Change the Compose `ports` or stop the conflicting service. Use `docker-compose.override.yml` for local port changes. |
| **Migrate fails** | Ensure Postgres is up: `docker compose ps` and `docker compose exec postgres pg_isready -U ztcp`. Check `DATABASE_URL` in `backend/.env` matches the Compose credentials (`postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable`). |
| **password authentication failed for user "root"** | Local Postgres uses user `ztcp`, not `root`. Set `DATABASE_URL=postgres://ztcp:ztcp@localhost:5432/ztcp?sslmode=disable` in `backend/.env`, or copy from `deploy/.env.example`. Running `make setup` will auto-fix `backend/.env` if it contains a `root` user URL. |
| **Collector fails to start or logs errors** | `docker compose logs otelcol`. Ensure Loki and Tempo are up (`docker compose ps`) and reachable from the collector (same network). For "connection refused" to Loki, confirm `otelcol-config.yaml` uses `http://loki:3100/loki/api/v1/push`. |
| **Prometheus target down** | In Prometheus UI (http://localhost:9090/targets), the `otelcol` target should be `otelcol:8889`. If down, check collector logs and that the collector exposes 8889 and is on the same network as Prometheus. |
| **No data in Grafana** | Confirm datasource URLs (use service names from inside Docker: `http://prometheus:9090`, `http://loki:3100`, `http://tempo:3200`). Check time range and that the ZTCP server has `OTEL_EXPORTER_OTLP_ENDPOINT` set and has sent traffic. |
| **TraceQL metrics query fails** (e.g. "empty ring" or "unknown service tempopb.MetricsGenerator") | The local Tempo image does not support TraceQL metrics queries (e.g. `rate() by(resource.service.name)`). Use **trace search by Trace ID** or **search by service name** in Grafana Explore instead. For trace-by-ID, paste a trace ID from logs or from the ZTCP Telemetry dashboard. |
| **No metrics in Prometheus** | The included `prometheus.yml` keeps only metrics matching `ztcp_*`. To scrape all collector metrics, remove the `metric_relabel_configs` block from `prometheus.yml` and restart Prometheus. |
