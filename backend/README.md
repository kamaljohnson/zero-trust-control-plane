# Zero Trust Control Plane — Backend

gRPC API server and async worker for the zero-trust control plane.

## Layout

- **cmd/server** — gRPC API server
- **cmd/worker** — async jobs (audit, cleanup)
- **proto/** — Protocol Buffer definitions (identity, user, organization, membership, session, device, policy, audit)
- **api/generated/** — protoc output
- **internal/** — server; one folder per table: user, identity, organization, membership, device, session, policy, audit; platform (tenancy, RBAC, plans); db; security; config
  - **internal/db/sqlc/** — single sqlc project: `schema/`, `queries/`, `gen/` (generated), `sqlc.yaml`. All repositories import `internal/db/sqlc/gen`.
  - **internal/<context>/repository/** — `repository.go` (interface), `postgres.go` (impl using internal/db/sqlc/gen)
- **pkg/** — shared grpc, logger, observability
- **internal/db/migrations/** — SQL migrations (single DB schema for deployment)
- **scripts/** — generate_proto.sh, generate_sqlc.sh, migrate.sh, seed.sh

## Generating sqlc code

The repository layer uses [sqlc](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html#setting-up) for type-safe SQL. The `internal/db/sqlc/gen/` directory is **generated** by sqlc from `internal/db/sqlc/schema/` and `internal/db/sqlc/queries/`; do not edit files in `gen/`.

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
./scripts/migrate.sh          # Run DB migrations
./scripts/seed.sh             # Seed dev data
```
