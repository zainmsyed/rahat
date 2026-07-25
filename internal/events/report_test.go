package events

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
)

func TestSummaryAndListFiltered(t *testing.T) {
	ctx := context.Background()
	sqlDB, err := db.OpenSQLite(ctx, filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer sqlDB.Close()
	repo := NewRepository(sqlDB)
	if _, err := sqlDB.ExecContext(ctx, `INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, created_at, updated_at) VALUES ('u1', 'Tester', 'UTC', 30, '2026-07-25T12:00:00Z', '2026-07-25T12:00:00Z')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	if _, err := repo.Create(ctx, EventLog{UserID: "u1", Channel: "telegram", EventType: "message_sent", MessageType: "daily_list", OccurredAt: now}); err != nil {
		t.Fatalf("create event 1: %v", err)
	}
	if _, err := repo.Create(ctx, EventLog{UserID: "u1", Channel: "telegram", EventType: "user_response", MessageType: "done", OccurredAt: now.Add(time.Minute)}); err != nil {
		t.Fatalf("create event 2: %v", err)
	}
	if _, err := repo.Create(ctx, EventLog{UserID: "u1", Channel: "email", EventType: "message_sent", MessageType: "daily_recap", OccurredAt: now.Add(2 * time.Minute)}); err != nil {
		t.Fatalf("create event 3: %v", err)
	}

	from := now.Add(30 * time.Second)
	summary, err := repo.Summary(ctx, ReportFilter{From: &from})
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if len(summary) != 2 {
		t.Fatalf("len(summary) = %d, want 2", len(summary))
	}

	listed, err := repo.ListFiltered(ctx, ReportFilter{From: &from})
	if err != nil {
		t.Fatalf("ListFiltered() error = %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("len(listed) = %d, want 2", len(listed))
	}
	if listed[0].Channel != "email" || listed[1].EventType != "user_response" {
		t.Fatalf("unexpected filtered ordering: %+v", listed)
	}
}
