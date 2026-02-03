# Zero Trust Control Plane — Backend

gRPC API server and async worker for the zero-trust control plane.

## Overview

The backend is a **gRPC API server** (and optional async worker). The server registers multiple gRPC services: Admin, Auth, User, Organization, Membership, Device, Session, Policy, Telemetry, Audit, and Health. **Auth is optional**: when enabled (see [Configuration](#configuration)), the server opens the database, wires the auth service and repos, and registers an auth interceptor that validates Bearer tokens and sets identity in context for protected RPCs; when disabled, no database connection is opened and auth RPCs return Unimplemented. AuthService is implemented by the identity handler ([internal/identity/handler](internal/identity/handler)) and auth service ([internal/identity/service](internal/identity/service)); see [docs/auth.md](docs/auth.md). When auth is enabled, **audit logging** is also enabled: an audit interceptor records who did what (org, user, action, resource, IP) after each protected RPC, and the auth service explicitly logs login success/failure, logout, and session created; see [docs/audit.md](docs/audit.md). **MFA** (risk-based, challenge/OTP) and **device trust** (policy-driven, revocable, time-bound) influence when a second factor is required; see [docs/mfa.md](docs/mfa.md) and [docs/device-trust.md](docs/device-trust.md).

## Documentation

- **[docs/auth.md](docs/auth.md)** — Authentication: architecture, API (Register, Login, Refresh, Logout), security (passwords, JWT, refresh rotation, reuse detection, interceptor), flows, configuration, and how auth uses the database.
- **[docs/audit.md](docs/audit.md)** — Audit logging: compliance trail, what is logged (RPC-derived and explicit auth/session events), ListAuditLogs API, interceptor and wiring, when enabled/disabled, configuration.
- **[docs/database.md](docs/database.md)** — Database: schema, enums and tables, when the DB is used, migrations, schema/codegen (sqlc, connection, repos), and cross-reference to auth table roles.
- **[docs/device-trust.md](docs/device-trust.md)** — Device trust: identifiable/revocable/time-bound devices, policy evaluation (OPA/Rego), when MFA is required and when trust is registered, configuration.
- **[docs/health.md](docs/health.md)** — Health checks: readiness RPC (HealthService.HealthCheck), behavior with and without database, how to call from Kubernetes or gRPC clients.
- **[docs/mfa.md](docs/mfa.md)** — MFA: risk-based MFA, when required, challenge/OTP flow, VerifyMFA and SubmitPhoneAndRequestMFA, API and configuration.
- **[docs/telemetry.md](docs/telemetry.md)** — Telemetry: gRPC interceptor and TelemetryService → Kafka → worker → Loki → Grafana; configuration, LogQL examples, and optional dashboard.

## Layout

- **cmd/server** — gRPC API server
- **cmd/worker** — telemetry worker (Kafka → Loki); optional async jobs
- **cmd/migrate** — DB migration runner (used by scripts/migrate.sh when CLI not installed)
- **cmd/seed** — Development data seeder (used by scripts/seed.sh)
- **docs/** — documentation: `auth.md`, `audit.md`, `database.md`, `device-trust.md`, `health.md`, `mfa.md`
- **proto/** — Protocol Buffer definitions: common, auth, user, org, membership, device, session, policy, audit, telemetry, admin, health
- **api/generated/** — generated Go and gRPC code from proto (buf or protoc)
- **internal/** — server; one folder per domain: user, identity, organization, membership, device, session, policy, audit; platform (tenancy, RBAC, plans); db; security; config
  - **internal/db/sqlc/** — single sqlc project: `schema/`, `queries/`, `gen/` (generated), `sqlc.yaml`. All repositories import `internal/db/sqlc/gen`.
  - **internal/<context>/repository/** — `repository.go` (interface), `postgres.go` (impl using internal/db/sqlc/gen)
- **pkg/** — shared grpc, logger, observability
- **internal/db/migrations/** — SQL migrations (single DB schema for deployment)
- **scripts/** — generate_proto.sh, generate_sqlc.sh, migrate.sh, seed.sh

## Configuration

Config is loaded from environment or `.env` (see [.env.example](.env.example)). `GRPC_ADDR` (default `:8080`) is the listen address.

**Auth and database**: Auth (and the database) are enabled only when `DATABASE_URL` and **both** `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` are set. When enabled, the server opens Postgres, builds the auth service and repos, and protects non-public RPCs with a Bearer access token. When any of the three is missing, the server runs without a DB and auth RPCs return Unimplemented. Full auth configuration and flows: [docs/auth.md](docs/auth.md).

**Telemetry**: When `KAFKA_BROKERS` is set, the server emits telemetry events (per-RPC and via TelemetryService) to Kafka. Run the worker with `KAFKA_BROKERS`, `LOKI_URL`, and `GRPC_ADDR=:0` to consume from Kafka and push logs to Loki; then add Loki as a Grafana datasource and use [docs/telemetry.md](docs/telemetry.md) for LogQL and dashboard import.

## Generating sqlc code

The repository layer uses [sqlc](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html#setting-up) for type-safe SQL. The `internal/db/sqlc/gen/` directory is **generated** by sqlc from `internal/db/sqlc/schema/` and `internal/db/sqlc/queries/`; do not edit files in `gen/`. See [docs/database.md](docs/database.md) for migrations list and schema/codegen workflow.

1. **Install sqlc** (one of):
   ```bash
   brew install sqlc
   ```
   or
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```
   (ensure `$GOPATH/bin` or `$HOME/go/bin` is on your PATH)

2. **Generate code** (single run):
   ```bash
   ./scripts/generate_sqlc.sh
   ```
   Or: `cd internal/db/sqlc && sqlc generate`

3. **Build** (after generation):
   ```bash
   go build ./...
   ```

## Generating proto code

The gRPC API is defined in `proto/`. Generated Go and gRPC stubs go to `api/generated/`; do not edit files there.

1. **Option A — buf** (recommended):
   ```bash
   brew install bufbuild/buf/buf
   ./scripts/generate_proto.sh
   ```
   (script runs `buf generate` from `proto/`; output under `api/generated/`)

2. **Option B — protoc**:
   Install [protoc](https://protobuf.dev/downloads) and the Go plugins:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```
   Ensure `$GOPATH/bin` or `$HOME/go/bin` is on your PATH, then:
   ```bash
   ./scripts/generate_proto.sh
   ```
   (script runs `protoc` with `-I proto` and writes under `api/generated/`)

3. **Build** (after generation):
   ```bash
   go build ./...
   ```

## Build & run

```bash
# Server
go run ./cmd/server

# Worker
go run ./cmd/worker
```

## Database migrations

Run migrations with `./scripts/migrate.sh` from the backend root. By default it applies pending migrations (up); pass `down` to roll back one version:

```bash
./scripts/migrate.sh          # apply pending migrations (up)
./scripts/migrate.sh down     # roll back one migration
```

The script reads `DATABASE_URL` from `.env` or the environment; create a `.env` from [.env.example](.env.example) if needed. If the [golang-migrate](https://github.com/golang-migrate/migrate) CLI is in PATH (e.g. `brew install golang-migrate`), the script uses it; otherwise it runs `go run ./cmd/migrate`.

## Seeding development data

Run `./scripts/seed.sh` from the backend root **after** migrations to insert development sample data (users, org, memberships, device, policies, MFA/platform settings). The script reads `DATABASE_URL` from `.env` or the environment. Seed is **idempotent**: if the dev user already exists, it skips inserts and exits successfully.

Default dev logins for local testing:

| Email               | Password    | Role   |
|---------------------|-------------|--------|
| `dev@example.com`   | `password123` | owner  |
| `member@example.com`| `password123` | member |

## Scripts

```bash
./scripts/generate_proto.sh   # Generate code from proto/
./scripts/generate_sqlc.sh   # Generate sqlc code (run after installing sqlc)
./scripts/migrate.sh          # Run DB migrations (see docs/database.md for migrations list)
./scripts/seed.sh             # Seed dev data (run after migrate; see Seeding development data)
```
