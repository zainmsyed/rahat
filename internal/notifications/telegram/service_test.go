package telegram

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

type fakeBot struct{ messages []SendMessageRequest }

func (f *fakeBot) SendMessage(_ context.Context, req SendMessageRequest) error {
	f.messages = append(f.messages, req)
	return nil
}

func TestSendMorningBatchAndWindowReminders(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}

	userSvc := users.NewService(users.NewRepository(sqlDB))
	taskSvc := tasks.NewService(tasks.NewRepository(sqlDB))
	occSvc := occurrences.NewService(occurrences.NewRepository(sqlDB))
	eventSvc := events.NewService(events.NewRepository(sqlDB))
	bot := &fakeBot{}
	svc := NewService(bot, userSvc, taskSvc, occSvc, eventSvc)

	user, _ := userSvc.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 30, TelegramChatID: "chat-1"})
	taskOne, _ := taskSvc.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	_, _ = occSvc.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: taskOne.Task.ID, Status: occurrences.StatusScheduled, ScheduledForDate: "2026-07-08", OriginalScheduledForDate: "2026-07-08", ScheduledTimeOfDay: tasks.TimeOfDayMorning})

	if err := svc.SendMorningBatch(ctx, user.ID, mustDate("2026-07-08")); err != nil {
		t.Fatal(err)
	}
	if err := svc.SendWindowReminders(ctx, user.ID, mustDate("2026-07-08"), tasks.TimeOfDayMorning); err != nil {
		t.Fatal(err)
	}
	if len(bot.messages) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(bot.messages))
	}
	items, _ := eventSvc.ListByUser(ctx, user.ID)
	if len(items) != 2 {
		t.Fatalf("len(event logs) = %d, want 2", len(items))
	}
}

func mustDate(value string) (outTime time.Time) {
	outTime, _ = time.Parse("2006-01-02", value)
	return outTime
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
