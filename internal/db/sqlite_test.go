package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenSQLiteBootstrapsDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "nested", "rahat.sqlite3")

	db, err := OpenSQLite(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer db.Close()

	assertPragmaEquals(t, db, "journal_mode", "wal")
	assertPragmaEquals(t, db, "foreign_keys", int64(1))

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("schema_migrations query failed: %v", err)
	}

	assertTableExists(t, db, "tasks")
	assertTableExists(t, db, "starter_task_templates")
}

func assertTableExists(t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", tableName).Scan(&count); err != nil {
		t.Fatalf("table existence query for %s failed: %v", tableName, err)
	}
	if count != 1 {
		t.Fatalf("table %s exists count = %d, want 1", tableName, count)
	}
}

func assertPragmaEquals(t *testing.T, db *sql.DB, pragma string, want any) {
	t.Helper()

	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("db.Conn() error = %v", err)
	}
	defer conn.Close()

	var got any
	if err := conn.QueryRowContext(context.Background(), "PRAGMA "+pragma+";").Scan(&got); err != nil {
		t.Fatalf("PRAGMA %s query failed: %v", pragma, err)
	}

	switch want := want.(type) {
	case string:
		if got != want {
			t.Fatalf("PRAGMA %s = %v, want %q", pragma, got, want)
		}
	case int64:
		if got != want {
			t.Fatalf("PRAGMA %s = %v, want %d", pragma, got, want)
		}
	default:
		t.Fatalf("unsupported expected type %T", want)
	}
}
