package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a Postgres connection using the given DSN. Caller must call Close when done.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
