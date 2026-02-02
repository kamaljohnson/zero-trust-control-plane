# Zero Trust Control Plane — Backend

gRPC API server and async worker for the zero-trust control plane.

## Overview

The backend is a **gRPC API server** (and optional async worker). The server registers multiple gRPC services: Admin, Auth, User, Organization, Membership, Device, Session, Policy, Telemetry, Audit, and Health. **Auth is optional**: when enabled (see [Configuration](#configuration)), the server opens the database, wires the auth service and repos, and registers an auth interceptor that validates Bearer tokens and sets identity in context for protected RPCs; when disabled, no database connection is opened and auth RPCs return Unimplemented. AuthService is implemented by the identity handler ([internal/identity/handler](internal/identity/handler)) and auth service ([internal/identity/service](internal/identity/service)); see [docs/auth.md](docs/auth.md).

## Documentation

- **[docs/auth.md](docs/auth.md)** — Authentication: architecture, API (Register, Login, Refresh, Logout), security (passwords, JWT, refresh rotation, reuse detection, interceptor), flows, configuration, and how auth uses the database.
- **[docs/database.md](docs/database.md)** — Database: schema, enums and tables, when the DB is used, migrations, schema/codegen (sqlc, connection, repos), and cross-reference to auth table roles.

## Layout

- **cmd/server** — gRPC API server
- **cmd/worker** — async jobs (audit, cleanup)
- **docs/** — auth and database documentation (`auth.md`, `database.md`)
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

## Scripts

```bash
./scripts/generate_proto.sh   # Generate code from proto/
./scripts/generate_sqlc.sh   # Generate sqlc code (run after installing sqlc)
./scripts/migrate.sh          # Run DB migrations (see docs/database.md for migrations list)
./scripts/seed.sh             # Seed dev data
```
