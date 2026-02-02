// Package migrate runs database migrations from embedded SQL files using golang-migrate.
package migrate

import (
	"errors"
	"fmt"

	"zero-trust-control-plane/backend/internal/db"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// ErrNoChange is returned when Up/Down has nothing to do (already at target version).
var ErrNoChange = migrate.ErrNoChange

// Run applies migrations in the given direction using the provided DSN.
// direction must be "up" or "down". Returns nil on success; ErrNoChange when already
// at latest (up) or no migrations to downgrade (down); other errors for DB or I/O failures.
func Run(dsn string, direction string) error {
	if dsn == "" {
		return errors.New("DATABASE_URL is not set; create a .env from .env.example or set DATABASE_URL")
	}
	if direction != "up" && direction != "down" {
		return fmt.Errorf("direction must be up or down, got %q", direction)
	}

	sourceDriver, err := iofs.New(db.MigrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}
	return nil
}
