package db

import "embed"

// MigrationFS embeds SQL migration files from internal/db/migrations.
// Used by the migrate runner (cmd/migrate and scripts/migrate.sh) to apply migrations.
//
//go:embed migrations/*.sql
var MigrationFS embed.FS
