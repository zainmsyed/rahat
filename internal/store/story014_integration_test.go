package store

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrationResolvesDuplicateTelegramChatIDs(t *testing.T) {
	ctx := context.Background()

	// Recreate the real schema state immediately before the Story 014 migration,
	// then seed two users sharing the same Telegram chat ID.
	db := openTestDBAtMigration(t, "010_story_013_subtask_archive.sql")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at)
		VALUES ('u-first', 'First', 'UTC', 30, 'shared-chat', '', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
		       ('u-second', 'Second', 'UTC', 30, 'shared-chat', '', '2026-01-02T00:00:00Z', '2026-01-02T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed duplicate users: %v", err)
	}

	// Apply the remaining migrations through the same path used by the app.
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	var linkedCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE telegram_chat_id = ?`, "shared-chat").Scan(&linkedCount); err != nil {
		t.Fatalf("count linked users: %v", err)
	}
	if linkedCount != 1 {
		t.Fatalf("expected exactly one linked user, got %d", linkedCount)
	}

	var conflictCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM telegram_identity_conflicts WHERE chat_id = ?`, "shared-chat").Scan(&conflictCount); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("count conflicts: %v", err)
		}
	}
	if conflictCount != 1 {
		t.Fatalf("expected one recorded conflict, got %d", conflictCount)
	}

	// The unique partial index should now reject a second user with the same chat ID.
	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at)
		VALUES ('u-third', 'Third', 'UTC', 30, 'shared-chat', '', '2026-01-03T00:00:00Z', '2026-01-03T00:00:00Z');
	`); err == nil {
		t.Fatal("expected unique index violation for duplicate telegram_chat_id")
	}
}

func TestMigrationsApplyInOrderOnCleanDatabase(t *testing.T) {
	ctx := context.Background()

	dir, err := migrationDir()
	if err != nil {
		t.Fatalf("migrationDir() error = %v", err)
	}
	names, err := migrationNames(dir)
	if err != nil {
		t.Fatalf("migrationNames() error = %v", err)
	}
	if len(names) == 0 {
		t.Fatal("no migrations found")
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			db := openCleanTestDB(t)
			if err := ApplyMigrationsUpTo(ctx, db, name); err != nil {
				t.Fatalf("ApplyMigrationsUpTo(%q) error = %v", name, err)
			}
			applied, err := migrationApplied(ctx, db, name)
			if err != nil {
				t.Fatalf("migrationApplied(%q) error = %v", name, err)
			}
			if !applied {
				t.Fatalf("migration %q was not recorded as applied", name)
			}
		})
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// The database was already migrated by openTestDB; a second pass must be a no-op.
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("second ApplyMigrations() error = %v", err)
	}

	dir, err := migrationDir()
	if err != nil {
		t.Fatalf("migrationDir() error = %v", err)
	}
	names, err := migrationNames(dir)
	if err != nil {
		t.Fatalf("migrationNames() error = %v", err)
	}
	for _, name := range names {
		applied, err := migrationApplied(ctx, db, name)
		if err != nil {
			t.Fatalf("migrationApplied(%q) error = %v", name, err)
		}
		if !applied {
			t.Fatalf("migration %q not recorded after idempotent re-run", name)
		}
	}
}

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
