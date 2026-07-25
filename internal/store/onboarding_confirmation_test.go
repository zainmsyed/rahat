package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "rahat.sqlite3")+"?_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping sqlite: %v", err)
	}
	if _, err := sqlDB.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}
	if err := ApplyMigrations(context.Background(), sqlDB); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return sqlDB
}

func TestOnboardingConfirmationRepository(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	repo := NewOnboardingConfirmationRepository(sqlDB)

	userID := "user-abc-123"

	_, found, err := repo.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Get() before record error = %v", err)
	}
	if found {
		t.Fatal("expected no confirmation before record")
	}

	if err := repo.Record(ctx, userID, true, ""); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	record, found, err := repo.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Get() after record error = %v", err)
	}
	if !found {
		t.Fatal("expected confirmation to be found after record")
	}
	if !record.Delivered {
		t.Fatalf("expected Delivered=true, got %v", record.Delivered)
	}
	if record.FailedReason != "" {
		t.Fatalf("expected empty FailedReason, got %q", record.FailedReason)
	}
	if record.SentAt.IsZero() {
		t.Fatal("expected non-zero SentAt")
	}

	// Duplicate records should be ignored and the original state preserved.
	if err := repo.Record(ctx, userID, false, "should be ignored"); err != nil {
		t.Fatalf("duplicate Record() error = %v", err)
	}
	record2, _, _ := repo.Get(ctx, userID)
	if !record2.Delivered {
		t.Fatalf("duplicate record should not overwrite Delivered=true")
	}
}

func TestOnboardingConfirmationRepositoryRecordRequiresUserID(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	repo := NewOnboardingConfirmationRepository(sqlDB)
	if err := repo.Record(ctx, "", true, ""); err == nil {
		t.Fatal("expected error when user id is empty")
	}
}
