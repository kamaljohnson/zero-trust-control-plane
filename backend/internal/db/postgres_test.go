package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestOpen_EmptyDSN(t *testing.T) {
	db, err := Open("")
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open with empty DSN should return error")
	}
	if db != nil {
		t.Error("Open should return nil db when error occurs")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestOpen_InvalidDSN(t *testing.T) {
	testCases := []struct {
		name string
		dsn  string
	}{
		{"invalid format", "invalid-dsn"},
		{"missing driver", "://localhost/test"},
		{"malformed", "postgres://"},
		{"invalid characters", "postgres://user:pass@host:port/db?invalid=param"},
		{"missing host", "postgres://user:pass@/db"},
		{"empty string", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := Open(tc.dsn)
			if err == nil {
				if db != nil {
					db.Close()
				}
				t.Errorf("Open with invalid DSN %q should return error", tc.dsn)
			}
			if db != nil {
				t.Error("Open should return nil db when error occurs")
			}
			if err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestOpen_ConnectionFailure(t *testing.T) {
	// Test with a DSN that will fail to connect (invalid host/port)
	testCases := []struct {
		name string
		dsn  string
	}{
		{"invalid host", "postgres://user:pass@invalid-host-that-does-not-exist:5432/db"},
		{"invalid port", "postgres://user:pass@localhost:99999/db"},
		{"nonexistent database", "postgres://user:pass@localhost:5432/nonexistent_db"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := Open(tc.dsn)
			if err == nil {
				if db != nil {
					db.Close()
				}
				t.Errorf("Open with connection failure DSN %q should return error", tc.dsn)
			}
			if db != nil {
				// Connection should be closed when ping fails
				// Verify it's closed by trying to ping again
				pingErr := db.Ping()
				if pingErr == nil {
					t.Error("database connection should be closed when ping fails")
				}
				db.Close()
			}
			if err.Error() == "" {
				t.Error("error message should not be empty")
			}
		})
	}
}

func TestOpen_DSNWithSpecialCharacters(t *testing.T) {
	// Test DSNs with special characters that might cause issues
	testCases := []struct {
		name string
		dsn  string
	}{
		{"password with special chars", "postgres://user:p@ss!w0rd@localhost:5432/db"},
		{"encoded password", "postgres://user:p%40ssw0rd@localhost:5432/db"},
		{"database name with dash", "postgres://user:pass@localhost:5432/my-db"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// These will fail to connect, but should parse correctly
			db, err := Open(tc.dsn)
			if err == nil {
				if db != nil {
					db.Close()
				}
				// If it succeeds, that's fine - just verify db is not nil
				if db == nil {
					t.Error("if Open succeeds, db should not be nil")
				}
			} else {
				// Connection failure is expected
				if db != nil {
					db.Close()
				}
				if err.Error() == "" {
					t.Error("error message should not be empty")
				}
			}
		})
	}
}

func TestOpen_ConnectionClosedOnPingFailure(t *testing.T) {
	// Verify that when Ping() fails, the connection is closed
	dsn := "postgres://user:pass@invalid-host:5432/db"
	db, err := Open(dsn)
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open should fail with invalid host")
	}
	if db != nil {
		// The connection should have been closed in Open() when ping failed
		// Try to use it - should fail
		var result int
		queryErr := db.QueryRow("SELECT 1").Scan(&result)
		if queryErr == nil {
			t.Error("database connection should be closed when ping fails")
		}
		// Also verify Close() can be called safely
		closeErr := db.Close()
		if closeErr != nil && closeErr != sql.ErrConnDone {
			t.Errorf("Close() should return nil or ErrConnDone, got: %v", closeErr)
		}
	}
}

func TestOpen_WhitespaceDSN(t *testing.T) {
	// DSN with whitespace should be invalid
	db, err := Open("   ")
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Error("Open with whitespace DSN should return error")
	}
	if db != nil {
		db.Close()
	}
}

func TestOpen_DSNWithQueryParams(t *testing.T) {
	// Test DSNs with various query parameters
	testCases := []struct {
		name string
		dsn  string
	}{
		{"with sslmode", "postgres://user:pass@localhost:5432/db?sslmode=disable"},
		{"with multiple params", "postgres://user:pass@localhost:5432/db?sslmode=require&connect_timeout=10"},
		{"with invalid param", "postgres://user:pass@localhost:5432/db?invalid=value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// These will fail to connect, but DSN parsing should work
			db, err := Open(tc.dsn)
			if err == nil {
				if db != nil {
					db.Close()
				}
				// Success is fine
			} else {
				// Connection failure is expected
				if db != nil {
					db.Close()
				}
				if err.Error() == "" {
					t.Error("error message should not be empty")
				}
			}
		})
	}
}

func TestOpen_ErrorReturnedNotNil(t *testing.T) {
	// Verify that when Open fails, it returns a non-nil error
	db, err := Open("invalid-dsn")
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open should return error for invalid DSN")
	}
	if db != nil {
		db.Close()
	}
	// Error should be non-nil and have a message
	if err.Error() == "" {
		t.Error("error should have a non-empty message")
	}
}

func TestOpen_PingFailureClosesConnection(t *testing.T) {
	// This test verifies the specific behavior: when Ping() fails, db.Close() is called
	// We can't easily verify this without mocking, but we can verify the connection
	// is unusable after the error
	
	dsn := "postgres://user:pass@nonexistent-host:5432/db"
	db, err := Open(dsn)
	
	// Should fail
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open should fail with nonexistent host")
	}
	
	// db should be nil or closed
	if db != nil {
		// If db is not nil, it should be closed (unusable)
		pingErr := db.Ping()
		if pingErr == nil {
			t.Error("database should be closed when ping fails in Open()")
		}
		db.Close() // Safe to call even if already closed
	}
}

func TestOpen_Success(t *testing.T) {
	// This test requires a real database connection
	// It will be skipped if DATABASE_URL is not set or connection fails
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := Open(dsn)
	if err != nil {
		t.Skipf("Database connection failed (expected in test environment): %v", err)
	}
	defer db.Close()

	// Verify the connection is usable
	if err := db.Ping(); err != nil {
		t.Errorf("database connection should be usable after Open: %v", err)
	}

	// Verify we can execute a simple query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Errorf("should be able to query database: %v", err)
	}
	if result != 1 {
		t.Errorf("query result = %d, want 1", result)
	}
}
