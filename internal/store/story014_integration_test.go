package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrationResolvesDuplicateTelegramChatIDs(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "rahat.sqlite3")
	db, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			display_name TEXT NOT NULL,
			timezone TEXT NOT NULL,
			daily_time_budget_minutes INTEGER NOT NULL,
			telegram_chat_id TEXT,
			email TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
	`); err != nil {
		t.Fatalf("create bootstrap tables: %v", err)
	}

	for _, name := range []string{
		"001_story_002_core.sql",
		"002_story_002_starter_task_library.sql",
		"003_story_003_schedule_state.sql",
		"004_story_003_occurrence_ready_at.sql",
		"005_story_005_calendar.sql",
		"006_story_005_oauth_state.sql",
		"007_story_009_subtask_dependency.sql",
		"008_story_012_beta_auth.sql",
		"009_story_013_task_archive.sql",
		"010_story_013_subtask_archive.sql",
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(name) VALUES (?)`, name); err != nil {
			t.Fatalf("record migration %s: %v", name, err)
		}
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at)
		VALUES ('u-first', 'First', 'UTC', 30, 'shared-chat', '', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z'),
		       ('u-second', 'Second', 'UTC', 30, 'shared-chat', '', '2026-01-02T00:00:00Z', '2026-01-02T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed duplicate users: %v", err)
	}

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
