package main

import (
	"compress/gzip"
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
)

func TestBackupDatabaseToFilesystemUsesSQLiteSnapshot(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "rahat.sqlite3")
	sqlDB, err := db.OpenSQLite(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer sqlDB.Close()
	if _, err := sqlDB.ExecContext(ctx, `INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, email, created_at, updated_at) VALUES ('u1', 'Tester', 'UTC', 30, 'tester@example.com', '2026-07-25T12:00:00Z', '2026-07-25T12:00:00Z')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	backupDir := filepath.Join(t.TempDir(), "backups")
	if err := backupDatabase(ctx, sqlDB, databasePath, backupDir, time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("backupDatabase() error = %v", err)
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	archivePath := filepath.Join(backupDir, entries[0].Name())
	restoredPath := filepath.Join(t.TempDir(), "restored.sqlite3")
	if err := gunzipFile(archivePath, restoredPath); err != nil {
		t.Fatalf("gunzipFile() error = %v", err)
	}
	restoredDB, err := db.OpenSQLite(ctx, restoredPath)
	if err != nil {
		t.Fatalf("OpenSQLite(restored) error = %v", err)
	}
	defer restoredDB.Close()
	var integrity string
	if err := restoredDB.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		t.Fatalf("integrity_check query: %v", err)
	}
	if integrity != "ok" {
		t.Fatalf("integrity_check = %q, want ok", integrity)
	}
	var userCount int
	if err := restoredDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = 'tester@example.com'`).Scan(&userCount); err != nil {
		t.Fatalf("count restored users: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("restored user count = %d, want 1", userCount)
	}
}

func TestRunUserBatchReturnsFailureSummary(t *testing.T) {
	users := []opsUser{{ID: "u1", Timezone: "UTC"}, {ID: "u2", Timezone: "UTC"}}
	err := runUserBatch("telegram-daily", users, func(user opsUser) error {
		if user.ID == "u2" {
			return sql.ErrConnDone
		}
		return nil
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err == nil {
		t.Fatal("expected batch failure")
	}
	if !strings.Contains(err.Error(), "telegram-daily") || !strings.Contains(err.Error(), "u2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLocalDayAndWindowUsesUserTimezone(t *testing.T) {
	day, window := localDayAndWindow("America/Los_Angeles", time.Date(2026, 7, 25, 2, 0, 0, 0, time.UTC))
	if got := day.Format("2006-01-02"); got != "2026-07-24" {
		t.Fatalf("day = %s, want 2026-07-24", got)
	}
	if window != taskpkg.TimeOfDayEvening {
		t.Fatalf("window = %s, want evening", window)
	}
}

func TestBuildRecapBodyUsesAbsoluteLookaheadURL(t *testing.T) {
	runtime := &opsRuntime{webOrigin: "https://app.example.com"}
	body := runtime.buildRecapBody(opsUser{DisplayName: "Tester"}, nil, nil, "demo-token")
	if !strings.Contains(body, "https://app.example.com/lookahead?token=demo-token") {
		t.Fatalf("recap body missing absolute URL: %s", body)
	}
}

func TestSeedTestersIsIdempotent(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "rahat.sqlite3")
	sqlDB, err := db.OpenSQLite(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer sqlDB.Close()
	runtime := &opsRuntime{
		db:    sqlDB,
		users: usr.NewService(usr.NewRepository(sqlDB)),
		tasks: taskpkg.NewService(taskpkg.NewRepository(sqlDB)),
		prefs: preferences.NewService(preferences.NewRepository(sqlDB)),
	}
	if err := runtime.seedTesters(ctx); err != nil {
		t.Fatalf("first seedTesters() error = %v", err)
	}
	if err := runtime.seedTesters(ctx); err != nil {
		t.Fatalf("second seedTesters() error = %v", err)
	}
	assertCount(t, sqlDB, `SELECT COUNT(*) FROM users WHERE email IN ('tester.one@example.com', 'tester.two@example.com')`, 2)
	assertCount(t, sqlDB, `SELECT COUNT(*) FROM tasks WHERE name IN ('Laundry', 'Meal prep', 'Grocery run')`, 3)
}

func TestResetNonProductionRemovesDatabaseAndOutbox(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "rahat.sqlite3")
	sqlDB, err := db.OpenSQLite(context.Background(), databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	outboxDir := filepath.Join(t.TempDir(), "email-outbox")
	if err := os.MkdirAll(outboxDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(outboxDir, "demo.txt"), []byte("demo"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	runtime := &opsRuntime{db: sqlDB, databasePath: databasePath, outboxDir: outboxDir, appEnv: "development"}
	if err := runtime.resetNonProduction(context.Background(), nonProductionResetConfirm); err != nil {
		t.Fatalf("resetNonProduction() error = %v", err)
	}
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		t.Fatalf("database still exists or wrong error: %v", err)
	}
	if _, err := os.Stat(outboxDir); !os.IsNotExist(err) {
		t.Fatalf("outbox still exists or wrong error: %v", err)
	}
}

func TestResetNonProductionBlockedInProduction(t *testing.T) {
	runtime := &opsRuntime{appEnv: "production"}
	if err := runtime.resetNonProduction(context.Background(), nonProductionResetConfirm); err == nil {
		t.Fatal("expected production reset to be blocked")
	}
}

func gunzipFile(sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()
	out, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, reader)
	return err
}

func assertCount(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d for query %q", got, want, query)
	}
}
