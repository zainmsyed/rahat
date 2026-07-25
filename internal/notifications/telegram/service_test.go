package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

type fakeBot struct {
	messages []SendMessageRequest
	err      error
}

func (f *fakeBot) SendMessage(_ context.Context, req SendMessageRequest) error {
	if f.err != nil {
		return f.err
	}
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
	confirmationRepo := store.NewOnboardingConfirmationRepository(sqlDB)
	bot := &fakeBot{}
	svc := NewService(bot, userSvc, taskSvc, occSvc, eventSvc, confirmationRepo)

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

func setupServices(t *testing.T) (context.Context, *Service, *scheduler.Service, *fakeBot, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatal(err)
	}

	userSvc := users.NewService(users.NewRepository(sqlDB))
	taskSvc := tasks.NewService(tasks.NewRepository(sqlDB))
	occSvc := occurrences.NewService(occurrences.NewRepository(sqlDB))
	eventSvc := events.NewService(events.NewRepository(sqlDB))
	confirmationRepo := store.NewOnboardingConfirmationRepository(sqlDB)
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	schedulerSvc := scheduler.NewService(userSvc, taskSvc, occSvc, checkpointRepo, blockRepo)
	bot := &fakeBot{}
	svc := NewService(bot, userSvc, taskSvc, occSvc, eventSvc, confirmationRepo)
	return ctx, svc, schedulerSvc, bot, sqlDB
}

func TestSendOnboardingConfirmationLinkedUser(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "America/Chicago", DailyTimeBudgetMinutes: 120, TelegramChatID: "chat-1"})
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning stretch", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Evening review", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)

	day := mustDate("2026-07-27")
	plan, err := schedulerSvc.PlanDay(ctx, user.ID, day)
	if err != nil {
		t.Fatalf("PlanDay() error = %v", err)
	}
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	delivered, err := svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs)
	if err != nil {
		t.Fatalf("SendOnboardingConfirmation() error = %v", err)
	}
	if !delivered {
		t.Fatal("expected delivered=true")
	}
	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(bot.messages))
	}
	text := bot.messages[0].Text
	if !strings.Contains(text, "Morning stretch") || !strings.Contains(text, "Evening review") {
		t.Fatalf("expected routine names in message, got:\n%s", text)
	}
	if !strings.Contains(text, "Morning:") || !strings.Contains(text, "Evening:") {
		t.Fatalf("expected window group headings, got:\n%s", text)
	}
	if !strings.Contains(text, "/edit") {
		t.Fatalf("expected /edit hint, got:\n%s", text)
	}

	eventLogs, _ := svc.events.ListByUser(ctx, user.ID)
	found := false
	for _, event := range eventLogs {
		if event.MessageType == "onboarding_confirmation" && event.EventType == "message_sent" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected onboarding_confirmation event log, got %+v", eventLogs)
	}
}

func TestSendOnboardingConfirmationUnlinkedUser(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Read", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)

	day := mustDate("2026-07-27")
	plan, _ := schedulerSvc.PlanDay(ctx, user.ID, day)
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	delivered, err := svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs)
	if err != nil {
		t.Fatalf("SendOnboardingConfirmation() error = %v", err)
	}
	if delivered {
		t.Fatal("expected delivered=false for unlinked user")
	}
	if len(bot.messages) != 0 {
		t.Fatalf("expected no message for unlinked user, got %d", len(bot.messages))
	}
}

func TestSendOnboardingConfirmationIdempotent(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60, TelegramChatID: "chat-1"})
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Read", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)

	day := mustDate("2026-07-27")
	plan, _ := schedulerSvc.PlanDay(ctx, user.ID, day)
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	if _, err := svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs); err != nil {
		t.Fatalf("first call error = %v", err)
	}
	if _, err := svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs); err != nil {
		t.Fatalf("second call error = %v", err)
	}
	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 message total, got %d", len(bot.messages))
	}
}

func TestSendOnboardingConfirmationSendFailure(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)
	bot.err = fmt.Errorf("telegram API unavailable")

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60, TelegramChatID: "chat-1"})
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Read", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)

	day := mustDate("2026-07-27")
	plan, _ := schedulerSvc.PlanDay(ctx, user.ID, day)
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	delivered, err := svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs)
	if err == nil {
		t.Fatal("expected error when send fails")
	}
	if delivered {
		t.Fatal("expected delivered=false on failure")
	}

	stored, found, _ := svc.confirmations.Get(ctx, user.ID)
	if !found || stored.Delivered {
		t.Fatal("expected failed confirmation to be recorded")
	}

	eventLogs, _ := svc.events.ListByUser(ctx, user.ID)
	found = false
	for _, event := range eventLogs {
		if event.MessageType == "onboarding_confirmation" && event.EventType == "message_failed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected message_failed event log, got %+v", eventLogs)
	}
}

func TestSendOnboardingConfirmationOverflowAndSkipped(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 10, TelegramChatID: "chat-1"})
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Big task", DurationMinutes: 30, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityLow, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Small task", DurationMinutes: 5, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)

	day := mustDate("2026-07-27")
	plan, _ := schedulerSvc.PlanDay(ctx, user.ID, day)
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	_, _ = svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs)
	text := bot.messages[0].Text
	if !strings.Contains(text, "moved to a later day") && !strings.Contains(text, "skipped") {
		t.Fatalf("expected overflow/skip mention, got:\n%s", text)
	}
}

func TestSendOnboardingConfirmationEmptyWindowsOmitted(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60, TelegramChatID: "chat-1"})
	_, _ = svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning only", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)

	day := mustDate("2026-07-27")
	plan, _ := schedulerSvc.PlanDay(ctx, user.ID, day)
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	_, _ = svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs)
	text := bot.messages[0].Text
	if !strings.Contains(text, "Morning:") {
		t.Fatalf("expected Morning heading, got:\n%s", text)
	}
	if strings.Contains(text, "Afternoon:") || strings.Contains(text, "Evening:") {
		t.Fatalf("did not expect empty afternoon/evening headings, got:\n%s", text)
	}
}

func TestSendOnboardingConfirmationSafeContents(t *testing.T) {
	ctx, svc, schedulerSvc, bot, _ := setupServices(t)

	user, _ := svc.users.Create(ctx, users.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60, TelegramChatID: "chat-1"})
	taskOne, _ := svc.tasks.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Read", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)

	day := mustDate("2026-07-27")
	plan, _ := schedulerSvc.PlanDay(ctx, user.ID, day)
	taskDefs, _ := svc.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)

	_, _ = svc.SendOnboardingConfirmation(ctx, user.ID, day, plan, taskDefs)
	text := bot.messages[0].Text
	if strings.Contains(text, user.ID) || strings.Contains(text, taskOne.Task.ID) {
		t.Fatalf("message exposed internal IDs:\n%s", text)
	}
	if strings.Contains(text, "required_same_day") || strings.Contains(text, "soft_followup") {
		t.Fatalf("message exposed dependency semantics:\n%s", text)
	}
}
