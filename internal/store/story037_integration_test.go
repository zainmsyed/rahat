package store_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

func TestStory038MigrationDeduplicatesOpenOccurrencesAndReparentsEvents(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}

	user, err := users.NewService(users.NewRepository(sqlDB)).Create(ctx, users.User{
		DisplayName: "Migration test", Timezone: "UTC", DailyTimeBudgetMinutes: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err := tasks.NewService(tasks.NewRepository(sqlDB)).CreateTaskWithSubtasks(ctx, tasks.Task{
		UserID: user.ID, Name: "Deduplicate me", DurationMinutes: 5,
		CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1,
		Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(ctx, `DROP INDEX idx_occurrences_open_identity`); err != nil {
		t.Fatal(err)
	}
	for _, occurrence := range []struct {
		id     string
		status string
	}{
		{id: "duplicate-pending", status: "pending"},
		{id: "surviving-scheduled", status: "scheduled"},
	} {
		if _, err := sqlDB.ExecContext(ctx, `
			INSERT INTO occurrences (id, user_id, task_id, subtask_id, status, scheduled_for_date, original_scheduled_for_date, scheduled_time_of_day, rollover_count, consecutive_no_count, created_at, updated_at)
			VALUES (?, ?, ?, NULL, ?, '2026-08-04', '2026-08-04', 'morning', 0, 0, '2026-08-04T00:00:00Z', ?)
		`, occurrence.id, user.ID, task.Task.ID, occurrence.status, map[string]string{
			"duplicate-pending":   "2026-08-04T00:01:00Z",
			"surviving-scheduled": "2026-08-04T00:02:00Z",
		}[occurrence.id]); err != nil {
			t.Fatal(err)
		}
	}
	for _, event := range []string{"event-for-duplicate", "event-for-survivor"} {
		occurrenceID := "surviving-scheduled"
		if event == "event-for-duplicate" {
			occurrenceID = "duplicate-pending"
		}
		if _, err := sqlDB.ExecContext(ctx, `
			INSERT INTO event_logs (id, user_id, occurrence_id, channel, event_type, message_type, payload_json, occurred_at)
			VALUES (?, ?, ?, 'telegram', 'message_sent', 'daily_list', '{}', '2026-08-04T00:03:00Z')
		`, event, user.ID, occurrenceID); err != nil {
			t.Fatal(err)
		}
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "014_story_038_occurrence_idempotency.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.ExecContext(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}

	var occurrenceCount int
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM occurrences WHERE user_id = ? AND task_id = ? AND original_scheduled_for_date = '2026-08-04'`, user.ID, task.Task.ID).Scan(&occurrenceCount); err != nil {
		t.Fatal(err)
	}
	if occurrenceCount != 1 {
		t.Fatalf("open occurrence count = %d, want 1", occurrenceCount)
	}

	var linkedEvents int
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM event_logs WHERE occurrence_id = 'surviving-scheduled'`).Scan(&linkedEvents); err != nil {
		t.Fatal(err)
	}
	if linkedEvents != 2 {
		t.Fatalf("events linked to survivor = %d, want 2", linkedEvents)
	}

	if _, err := sqlDB.ExecContext(ctx, `
		INSERT INTO occurrences (id, user_id, task_id, subtask_id, status, scheduled_for_date, original_scheduled_for_date, scheduled_time_of_day, rollover_count, consecutive_no_count, created_at, updated_at)
		VALUES ('second-open', ?, ?, NULL, 'pending', '2026-08-04', '2026-08-04', 'morning', 0, 0, '2026-08-04T00:04:00Z', '2026-08-04T00:04:00Z')
	`, user.ID, task.Task.ID); err == nil {
		t.Fatal("expected unique open-occurrence identity violation")
	}
}
