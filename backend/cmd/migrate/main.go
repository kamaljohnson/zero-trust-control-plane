// migrate runs DB migrations from embedded SQL; use with ./scripts/migrate.sh or go run ./cmd/migrate.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"zero-trust-control-plane/backend/internal/config"
	"zero-trust-control-plane/backend/internal/db/migrate"
)

func main() {
	direction := flag.String("direction", "up", "Migration direction: up or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is not set; create a .env from .env.example or set DATABASE_URL")
		os.Exit(1)
	}

	if err := migrate.Run(cfg.DatabaseURL, *direction); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			// Already at target version; success.
			return
		}
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}
