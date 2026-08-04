package store

import (
	"context"
	"database/sql"
	"net/url"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// openCleanTestDB opens a fresh SQLite database and creates the bootstrap
// schema_migrations table, but does not apply any migrations. This mirrors the
// first half of the production bootstrap path in internal/db.OpenSQLite.
func openCleanTestDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "rahat.sqlite3")

	query := url.Values{}
	query.Add("_pragma", "journal_mode(WAL)")
	query.Add("_pragma", "foreign_keys(ON)")
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     filepath.ToSlash(dbPath),
		RawQuery: query.Encode(),
	}).String()

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("ping sqlite: %v", err)
	}

	if _, err := sqlDB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}

	return sqlDB
}

// openTestDB opens a SQLite database that has had the full migration chain
// applied through the real bootstrap path.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return openTestDBAtMigration(t, "")
}

// openTestDBAtMigration opens a SQLite database and applies migrations up to
// and including the named migration. Passing an empty name applies the full
// chain.
func openTestDBAtMigration(t *testing.T, stopAt string) *sql.DB {
	t.Helper()

	sqlDB := openCleanTestDB(t)

	if stopAt != "" {
		if err := ApplyMigrationsUpTo(context.Background(), sqlDB, stopAt); err != nil {
			t.Fatalf("apply migrations up to %s: %v", stopAt, err)
		}
		return sqlDB
	}

	if err := ApplyMigrations(context.Background(), sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return sqlDB
}
