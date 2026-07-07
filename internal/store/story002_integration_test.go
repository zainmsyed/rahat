package store_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

func TestStory002PersistenceFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sqlDB := openTestDB(t)

	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() second run error = %v", err)
	}

	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
	preferenceService := preferences.NewService(preferences.NewRepository(sqlDB))
	eventService := events.NewService(events.NewRepository(sqlDB))

	user, err := userService.Create(ctx, users.User{
		DisplayName:            "Ayla Rahat",
		Timezone:               "America/Chicago",
		DailyTimeBudgetMinutes: 60,
		TelegramChatID:         "12345",
		Email:                  "ayla@example.com",
	})
	if err != nil {
		t.Fatalf("Create user error = %v", err)
	}

	multistepTask, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{
		UserID:              user.ID,
		Name:                "Laundry",
		Description:         "Run the laundry loop",
		DurationMinutes:     25,
		CadenceType:         tasks.CadenceTypeCount,
		CadenceValue:        2,
		Priority:            tasks.PriorityMedium,
		TimeOfDayPreference: tasks.TimeOfDayAny,
		IsMultistep:         true,
	}, []tasks.Subtask{
		{Name: "Wash", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayAfternoon, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
		{Name: "Fold", StepOrder: 3, DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayEvening, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
	})
	if err != nil {
		t.Fatalf("CreateTaskWithSubtasks() error = %v", err)
	}

	lookedUpTask, err := taskService.GetTaskWithSubtasks(ctx, multistepTask.Task.ID)
	if err != nil {
		t.Fatalf("GetTaskWithSubtasks() error = %v", err)
	}
	if len(lookedUpTask.Subtasks) != 3 {
		t.Fatalf("len(subtasks) = %d, want 3", len(lookedUpTask.Subtasks))
	}
	if lookedUpTask.Subtasks[1].GapRule.MinGapAfterPreviousMinutes != 45 {
		t.Fatalf("dryer gap = %d, want 45", lookedUpTask.Subtasks[1].GapRule.MinGapAfterPreviousMinutes)
	}

	pausedTask, err := taskService.PauseTask(ctx, multistepTask.Task.ID, true)
	if err != nil {
		t.Fatalf("PauseTask() error = %v", err)
	}
	if !pausedTask.IsPaused {
		t.Fatal("PauseTask() did not persist paused flag")
	}

	updatedSubtask := lookedUpTask.Subtasks[2]
	updatedSubtask.GapRule.MinGapAfterPreviousMinutes = 60
	updatedSubtask, err = taskService.UpdateSubtask(ctx, updatedSubtask)
	if err != nil {
		t.Fatalf("UpdateSubtask() error = %v", err)
	}
	if updatedSubtask.GapRule.MinGapAfterPreviousMinutes != 60 {
		t.Fatalf("updated fold gap = %d, want 60", updatedSubtask.GapRule.MinGapAfterPreviousMinutes)
	}

	occurrence, err := occurrenceService.Create(ctx, occurrences.Occurrence{
		UserID:                   user.ID,
		TaskID:                   multistepTask.Task.ID,
		SubtaskID:                lookedUpTask.Subtasks[0].ID,
		Status:                   occurrences.StatusScheduled,
		ScheduledForDate:         "2026-07-08",
		OriginalScheduledForDate: "2026-07-08",
		ScheduledTimeOfDay:       tasks.TimeOfDayMorning,
	})
	if err != nil {
		t.Fatalf("Create occurrence error = %v", err)
	}

	occurrence.Status = occurrences.StatusCompleted
	completedAt := time.Date(2026, 7, 8, 15, 0, 0, 0, time.UTC)
	occurrence.CompletedAt = &completedAt
	occurrence, err = occurrenceService.Update(ctx, occurrence)
	if err != nil {
		t.Fatalf("Update occurrence error = %v", err)
	}
	if occurrence.Status != occurrences.StatusCompleted || occurrence.CompletedAt == nil {
		t.Fatal("occurrence update did not persist completion state")
	}

	occurrencesForTask, err := occurrenceService.ListByTask(ctx, multistepTask.Task.ID)
	if err != nil {
		t.Fatalf("ListByTask() error = %v", err)
	}
	if len(occurrencesForTask) != 1 {
		t.Fatalf("len(occurrences) = %d, want 1", len(occurrencesForTask))
	}
	if occurrencesForTask[0].SubtaskID == "" {
		t.Fatal("occurrence subtask id was not persisted")
	}

	pref, err := preferenceService.Upsert(ctx, preferences.Preference{
		UserID:              user.ID,
		Channel:             preferences.ChannelTelegram,
		Enabled:             true,
		IsPrimary:           true,
		SupportsInteractive: true,
		RecapEnabled:        false,
	})
	if err != nil {
		t.Fatalf("Upsert preference error = %v", err)
	}
	if !pref.IsPrimary || !pref.SupportsInteractive {
		t.Fatal("telegram preference flags were not persisted")
	}

	pauseStart := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	pauseEnd := pauseStart.Add(24 * time.Hour)
	if _, err := preferenceService.CreatePause(ctx, preferences.Pause{
		UserID:   user.ID,
		Scope:    "global",
		Reason:   "Pause everything today",
		StartsAt: pauseStart,
		EndsAt:   pauseEnd,
	}); err != nil {
		t.Fatalf("Create global pause error = %v", err)
	}
	if _, err := preferenceService.CreatePause(ctx, preferences.Pause{
		UserID:   user.ID,
		TaskID:   multistepTask.Task.ID,
		Scope:    "task",
		Reason:   "Vacation",
		StartsAt: pauseStart,
		EndsAt:   pauseEnd.Add(6 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("Create task pause error = %v", err)
	}
	pauses, err := preferenceService.ListPausesByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListPausesByUser() error = %v", err)
	}
	if len(pauses) != 2 {
		t.Fatalf("len(pauses) = %d, want 2", len(pauses))
	}

	if _, err := eventService.Create(ctx, events.EventLog{
		UserID:       user.ID,
		OccurrenceID: occurrence.ID,
		Channel:      "telegram",
		EventType:    "message_sent",
		MessageType:  "daily_list",
		PayloadJSON:  `{"status":"queued"}`,
	}); err != nil {
		t.Fatalf("Create event log error = %v", err)
	}
	eventLogs, err := eventService.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListByUser() event logs error = %v", err)
	}
	if len(eventLogs) != 1 {
		t.Fatalf("len(event logs) = %d, want 1", len(eventLogs))
	}

	starterTemplates, err := taskService.ListStarterTaskTemplates(ctx)
	if err != nil {
		t.Fatalf("ListStarterTaskTemplates() error = %v", err)
	}
	if len(starterTemplates) < 7 {
		t.Fatalf("len(starter templates) = %d, want at least 7", len(starterTemplates))
	}
	if starterTemplates[0].Slug != "laundry" || len(starterTemplates[0].Subtasks) != 3 {
		t.Fatalf("unexpected first starter template: %+v", starterTemplates[0])
	}

	listedTasks, err := taskService.ListTasksByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListTasksByUser() error = %v", err)
	}
	if len(listedTasks) != 1 {
		t.Fatalf("len(listed tasks) = %d, want 1", len(listedTasks))
	}
}

func TestCreateTaskWithSubtasksRollsBackOnFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	sqlDB := openTestDB(t)

	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))

	user, err := userService.Create(ctx, users.User{
		DisplayName:            "Rollback Test",
		Timezone:               "UTC",
		DailyTimeBudgetMinutes: 30,
	})
	if err != nil {
		t.Fatalf("Create user error = %v", err)
	}

	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{
		UserID:              user.ID,
		Name:                "Broken laundry",
		DurationMinutes:     10,
		CadenceType:         tasks.CadenceTypeInterval,
		CadenceValue:        1,
		Priority:            tasks.PriorityLow,
		TimeOfDayPreference: tasks.TimeOfDayMorning,
		IsMultistep:         true,
	}, []tasks.Subtask{
		{Name: "First", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Duplicate order", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
	})
	if err == nil {
		t.Fatal("CreateTaskWithSubtasks() error = nil, want unique-constraint failure")
	}

	listedTasks, err := taskService.ListTasksByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListTasksByUser() error = %v", err)
	}
	if len(listedTasks) != 0 {
		t.Fatalf("len(listed tasks) after rollback = %d, want 0", len(listedTasks))
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return sqlDB
}
