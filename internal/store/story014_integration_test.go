package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
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
