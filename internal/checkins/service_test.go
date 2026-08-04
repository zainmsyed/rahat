package checkins

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

type fakeBot struct{ messages []ntg.SendMessageRequest }

func (f *fakeBot) SendMessage(_ context.Context, req ntg.SendMessageRequest) error {
	f.messages = append(f.messages, req)
	return nil
}

func TestHandleCallbackFlow(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}

	userSvc := users.NewService(users.NewRepository(sqlDB))
	taskSvc := tasks.NewService(tasks.NewRepository(sqlDB))
	occSvc := occurrences.NewService(occurrences.NewRepository(sqlDB))
	eventSvc := events.NewService(events.NewRepository(sqlDB))
	prefSvc := preferences.NewService(preferences.NewRepository(sqlDB))
	bot := &fakeBot{}
	svc := NewService(bot, userSvc, taskSvc, occSvc, eventSvc, prefSvc)
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }

	user, _ := userSvc.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30, TelegramChatID: "chat-1"})
	task, _ := taskSvc.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	occ, _ := occSvc.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: task.Task.ID, Status: occurrences.StatusScheduled, ScheduledForDate: "2026-07-08", OriginalScheduledForDate: "2026-07-08", ScheduledTimeOfDay: tasks.TimeOfDayMorning})

	if err := svc.HandleCallback(ctx, ntg.NotYetAction(user.ID, occ.ID)); err != nil {
		t.Fatal(err)
	}
	updated, _ := occSvc.GetByID(ctx, occ.ID)
	if updated.ConsecutiveNoCount != 1 {
		t.Fatalf("consecutive no = %d, want 1", updated.ConsecutiveNoCount)
	}

	if err := svc.HandleCallback(ctx, ntg.NotYetAction(user.ID, occ.ID)); err != nil {
		t.Fatal(err)
	}
	if len(bot.messages) == 0 || !strings.Contains(bot.messages[len(bot.messages)-1].Text, "hasn’t been landing") {
		t.Fatalf("expected adaptive follow-up, got %+v", bot.messages)
	}

	if err := svc.HandleCallback(ctx, ntg.SnoozeAction(user.ID, occ.ID)); err != nil {
		t.Fatal(err)
	}
	updated, _ = occSvc.GetByID(ctx, occ.ID)
	if updated.ScheduledForDate != "2026-07-11" {
		t.Fatalf("snoozed date = %s, want 2026-07-11", updated.ScheduledForDate)
	}

	occ2, _ := occSvc.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: task.Task.ID, Status: occurrences.StatusScheduled, ScheduledForDate: "2026-07-11", OriginalScheduledForDate: "2026-07-11", ScheduledTimeOfDay: tasks.TimeOfDayMorning})
	if err := svc.HandleCallback(ctx, ntg.DoneAction(user.ID, occ2.ID)); err != nil {
		t.Fatal(err)
	}
	updated2, _ := occSvc.GetByID(ctx, occ2.ID)
	if updated2.Status != occurrences.StatusCompleted {
		t.Fatalf("status = %s, want completed", updated2.Status)
	}

	if err := svc.HandleCallback(ctx, ntg.PauseTodayAction(user.ID)); err != nil {
		t.Fatal(err)
	}
	pauses, _ := prefSvc.ListPausesByUser(ctx, user.ID)
	if len(pauses) != 1 || pauses[0].Scope != "global" {
		t.Fatalf("pauses = %+v", pauses)
	}
	if err := svc.HandleCallback(ctx, ntg.PauseTaskAction(user.ID, task.Task.ID)); err != nil {
		t.Fatal(err)
	}
	pauses, _ = prefSvc.ListPausesByUser(ctx, user.ID)
	if len(pauses) != 2 {
		t.Fatalf("len(pauses) = %d, want 2", len(pauses))
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}
