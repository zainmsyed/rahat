package scheduler_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/users"
)

func TestPlanDayScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		seed                     func(context.Context, *testing.T, *sql.DB, *users.Service, *tasks.Service, *occurrences.Service, *store.CalendarBlockRepository) string
		day                      time.Time
		wantScheduled            int
		wantOverflowed           int
		wantSkipped              int
		wantCheckpoint           bool
		wantCheckpointWindowHour int
		assert                   func(*testing.T, scheduler.PlanResult)
	}{
		{
			name:                     "normal day schedules interval and weekly count work across windows",
			day:                      time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
			seed:                     seedNormalDay,
			wantScheduled:            3,
			wantOverflowed:           0,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if result.WindowBudgetsMinutes["morning"] <= 0 || result.WindowBudgetsMinutes["evening"] <= 0 {
					t.Fatalf("expected split window budgets, got %+v", result.WindowBudgetsMinutes)
				}
			},
		},
		{
			name:                     "overloaded day overflows and skips after rollover cap while preserving high priority",
			day:                      time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
			seed:                     seedOverloadedDay,
			wantScheduled:            0,
			wantOverflowed:           1,
			wantSkipped:              1,
			wantCheckpoint:           false,
			wantCheckpointWindowHour: 0,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if len(result.Overflowed) != 1 || result.Overflowed[0].RolloverCount != 3 || result.Overflowed[0].ScheduledForDate != "2026-07-15" {
					t.Fatalf("expected high priority overflow to keep rolling, got %+v", result.Overflowed)
				}
				if len(result.Skipped) != 1 || result.Skipped[0].Status != occurrences.StatusSkipped {
					t.Fatalf("expected non-high overflow to skip, got %+v", result.Skipped)
				}
			},
		},
		{
			name:                     "multistep laundry schedules morning afternoon evening steps with checkpoint persisted",
			day:                      time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
			seed:                     seedLaundryDay,
			wantScheduled:            3,
			wantOverflowed:           0,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				windows := []tasks.TimeOfDayPreference{}
				for _, occurrence := range result.Scheduled {
					windows = append(windows, occurrence.ScheduledTimeOfDay)
				}
				want := []tasks.TimeOfDayPreference{tasks.TimeOfDayMorning, tasks.TimeOfDayAfternoon, tasks.TimeOfDayEvening}
				for i := range want {
					if windows[i] != want[i] {
						t.Fatalf("scheduled windows = %+v, want %+v", windows, want)
					}
				}
			},
		},
		{
			name:                     "multistep chain defers all steps when the first required step cannot fit",
			day:                      time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
			seed:                     seedMultistepPartialBudget,
			wantScheduled:            0,
			wantOverflowed:           2,
			wantSkipped:              0,
			wantCheckpoint:           false,
			wantCheckpointWindowHour: 0,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				for _, occurrence := range result.Scheduled {
					if occurrence.SubtaskID != "" {
						t.Fatalf("no subtask should be scheduled when the chain cannot fit, got %+v", result.Scheduled)
					}
				}
			},
		},
		{
			name:                     "multistep soft follow-up defers without blocking required chain",
			day:                      time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC),
			seed:                     seedMultistepSoftFollowupBudget,
			wantScheduled:            2,
			wantOverflowed:           1,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if len(result.Scheduled) != 2 || len(result.Overflowed) != 1 {
					t.Fatalf("expected 2 scheduled required steps and 1 deferred follow-up, got scheduled=%+v overflowed=%+v", result.Scheduled, result.Overflowed)
				}
				if result.Overflowed[0].SubtaskID == "" {
					t.Fatalf("expected soft follow-up subtask to overflow, got %+v", result.Overflowed[0])
				}
			},
		},
		{
			name:                     "weekly count treats multistep completion as one parent task unit",
			day:                      time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			seed:                     seedWeeklyCountMultistep,
			wantScheduled:            3,
			wantOverflowed:           0,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
		},
		{
			name:                     "same-window subtasks honor min gap by delaying ready time",
			day:                      time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC),
			seed:                     seedSameWindowGap,
			wantScheduled:            2,
			wantOverflowed:           0,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if result.Scheduled[0].ReadyAt == nil || result.Scheduled[1].ReadyAt == nil {
					t.Fatalf("expected ready_at values, got %+v", result.Scheduled)
				}
				gap := result.Scheduled[1].ReadyAt.Sub(*result.Scheduled[0].ReadyAt)
				if gap < 90*time.Minute {
					t.Fatalf("ready gap = %v, want at least 90m", gap)
				}
				if result.Scheduled[1].ScheduledTimeOfDay != tasks.TimeOfDayMorning {
					t.Fatalf("second subtask window = %s, want morning", result.Scheduled[1].ScheduledTimeOfDay)
				}
			},
		},
		{
			name:                     "all-day large calendar block limits day to short tasks and explains blocked windows",
			day:                      time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
			seed:                     seedAllDayCalendarBlock,
			wantScheduled:            1,
			wantOverflowed:           1,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if result.SmallTaskOnlyReason == "" {
					t.Fatal("expected small-task-only explanation")
				}
				for _, window := range []string{"morning", "afternoon", "evening"} {
					if len(result.BlockedWindows[window]) == 0 {
						t.Fatalf("expected blocked reason for %s, got %+v", window, result.BlockedWindows)
					}
				}
			},
		},
		{
			name:                     "medium afternoon block zeroes afternoon budget without blocking other windows",
			day:                      time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			seed:                     seedMediumWindowCalendarBlock,
			wantScheduled:            1,
			wantOverflowed:           1,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if result.WindowBudgetsMinutes["afternoon"] != 0 {
					t.Fatalf("afternoon budget = %d, want 0", result.WindowBudgetsMinutes["afternoon"])
				}
				if len(result.BlockedWindows["afternoon"]) == 0 {
					t.Fatalf("expected afternoon blocked reason, got %+v", result.BlockedWindows)
				}
			},
		},
		{
			name:                     "single-window large event does not trigger day-wide small-task filter",
			day:                      time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC),
			seed:                     seedSingleWindowLargeCalendarBlock,
			wantScheduled:            2,
			wantOverflowed:           1,
			wantSkipped:              0,
			wantCheckpoint:           true,
			wantCheckpointWindowHour: 8,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				if result.SmallTaskOnlyReason != "" {
					t.Fatalf("unexpected small-task-only reason: %q", result.SmallTaskOnlyReason)
				}
				if result.WindowBudgetsMinutes["evening"] != 0 {
					t.Fatalf("evening budget = %d, want 0", result.WindowBudgetsMinutes["evening"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqlDB := openTestDB(t)
			if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
				t.Fatalf("ApplyMigrations() error = %v", err)
			}

			userService := users.NewService(users.NewRepository(sqlDB))
			taskService := tasks.NewService(tasks.NewRepository(sqlDB))
			occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
			checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
			blockRepo := store.NewCalendarBlockRepository(sqlDB)
			svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

			userID := tt.seed(ctx, t, sqlDB, userService, taskService, occurrenceService, blockRepo)
			result, err := svc.PlanDay(ctx, userID, tt.day)
			if err != nil {
				t.Fatalf("PlanDay() error = %v", err)
			}

			if len(result.Scheduled) != tt.wantScheduled {
				t.Fatalf("len(Scheduled) = %d, want %d", len(result.Scheduled), tt.wantScheduled)
			}
			if len(result.Overflowed) != tt.wantOverflowed {
				t.Fatalf("len(Overflowed) = %d, want %d", len(result.Overflowed), tt.wantOverflowed)
			}
			if len(result.Skipped) != tt.wantSkipped {
				t.Fatalf("len(Skipped) = %d, want %d", len(result.Skipped), tt.wantSkipped)
			}
			if tt.wantCheckpoint {
				if result.Checkpoint.NextCheckpointAt == nil || result.Checkpoint.NextCheckpointAt.Hour() != tt.wantCheckpointWindowHour {
					t.Fatalf("next checkpoint = %v, want hour %d", result.Checkpoint.NextCheckpointAt, tt.wantCheckpointWindowHour)
				}
			} else if result.Checkpoint.NextCheckpointAt != nil {
				t.Fatalf("next checkpoint = %v, want nil", result.Checkpoint.NextCheckpointAt)
			}
			if result.Checkpoint.ScheduledOccurrenceCount != tt.wantScheduled {
				t.Fatalf("checkpoint count = %d, want %d", result.Checkpoint.ScheduledOccurrenceCount, tt.wantScheduled)
			}
			if tt.assert != nil {
				tt.assert(t, result)
			}
		})
	}
}

func TestDayPreferenceFiltersWeekendAndKeepsOverflowOnAllowedDay(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}
	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

	user, err := userService.Create(ctx, users.User{DisplayName: "Day preference", Timezone: "UTC", DailyTimeBudgetMinutes: 5})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Weekday task", DurationMinutes: 1, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning, DayPreference: tasks.DayPreferenceWeekday}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Weekend task", DurationMinutes: 10, CadenceType: tasks.CadenceTypeCount, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning, DayPreference: tasks.DayPreferenceWeekend}, nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := svc.PlanDay(ctx, user.ID, time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Scheduled) != 0 || len(result.Overflowed) != 1 {
		t.Fatalf("Saturday result = scheduled %d, overflow %d, want 0/1", len(result.Scheduled), len(result.Overflowed))
	}
	if result.Overflowed[0].ScheduledForDate != "2026-07-19" {
		t.Fatalf("weekend overflow date = %s, want Sunday", result.Overflowed[0].ScheduledForDate)
	}
}

func TestPlanDayUsesUserLocalDateForNonMidnightInstant(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}
	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

	user, err := userService.Create(ctx, users.User{DisplayName: "Timezone day", Timezone: "America/Los_Angeles", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Sunday task", DurationMinutes: 10, CadenceType: tasks.CadenceTypeCount, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning, DayPreference: tasks.DayPreferenceWeekend}, nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := svc.PlanDay(ctx, user.ID, time.Date(2026, 8, 3, 0, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if result.Date != "2026-08-02" {
		t.Fatalf("plan date = %s, want 2026-08-02", result.Date)
	}
	if len(result.Scheduled) != 1 {
		t.Fatalf("scheduled = %d, want 1", len(result.Scheduled))
	}
}

func seedNormalDay(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Normal", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}

	clean, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Clean kitchen", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
	if err != nil {
		t.Fatal(err)
	}
	followUp, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Pediatrician", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 7, Priority: tasks.PriorityHigh, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mealPrep, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Meal prep", DurationMinutes: 25, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: clean.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-12", OriginalScheduledForDate: "2026-07-12", ScheduledTimeOfDay: tasks.TimeOfDayEvening})
	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: followUp.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-01", OriginalScheduledForDate: "2026-07-01", ScheduledTimeOfDay: tasks.TimeOfDayMorning})
	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: mealPrep.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-08", OriginalScheduledForDate: "2026-07-08", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
	return user.ID
}

func seedOverloadedDay(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Overloaded", Timezone: "UTC", DailyTimeBudgetMinutes: 10})
	if err != nil {
		t.Fatal(err)
	}

	high, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "High priority", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityHigh, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	medium, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Medium priority", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: high.Task.ID, Status: occurrences.StatusPending, ScheduledForDate: "2026-07-13", OriginalScheduledForDate: "2026-07-11", ScheduledTimeOfDay: tasks.TimeOfDayMorning, RolloverCount: 2})
	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: medium.Task.ID, Status: occurrences.StatusPending, ScheduledForDate: "2026-07-13", OriginalScheduledForDate: "2026-07-11", ScheduledTimeOfDay: tasks.TimeOfDayMorning, RolloverCount: 2})
	return user.ID
}

func seedLaundryDay(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Laundry", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatal(err)
	}

	laundry, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 25, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAny, IsMultistep: true}, []tasks.Subtask{
		{Name: "Wash", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayAfternoon, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
		{Name: "Fold", StepOrder: 3, DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayEvening, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
	})
	if err != nil {
		t.Fatal(err)
	}

	_ = laundry
	return user.ID
}

func seedMultistepPartialBudget(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Partial Laundry", Timezone: "UTC", DailyTimeBudgetMinutes: 5})
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAny, IsMultistep: true}, []tasks.Subtask{
		{Name: "Wash", StepOrder: 1, DurationMinutes: 9, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 1, TimeOfDayPreference: tasks.TimeOfDayAfternoon, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return user.ID
}

func seedMultistepSoftFollowupBudget(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Soft Laundry", Timezone: "UTC", DailyTimeBudgetMinutes: 12})
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 25, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning, IsMultistep: true}, []tasks.Subtask{
		{Name: "Wash", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning, DependencyType: tasks.SubtaskDependencyRequiredSameDay},
		{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning, DependencyType: tasks.SubtaskDependencyRequiredSameDay, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
		{Name: "Fold", StepOrder: 3, DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayMorning, DependencyType: tasks.SubtaskDependencySoftFollowup, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return user.ID
}

func seedWeeklyCountMultistep(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Weekly Count", Timezone: "UTC", DailyTimeBudgetMinutes: 45})
	if err != nil {
		t.Fatal(err)
	}

	laundry, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 25, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAny, IsMultistep: true}, []tasks.Subtask{
		{Name: "Wash", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayAfternoon, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
		{Name: "Fold", StepOrder: 3, DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayEvening, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 45}},
	})
	if err != nil {
		t.Fatal(err)
	}

	lookup, err := taskService.GetTaskWithSubtasks(ctx, laundry.Task.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, subtask := range lookup.Subtasks {
		_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: laundry.Task.ID, SubtaskID: subtask.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-14", OriginalScheduledForDate: "2026-07-14", ScheduledTimeOfDay: subtask.TimeOfDayPreference})
	}
	return user.ID
}

func seedSameWindowGap(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, _ *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Gap", Timezone: "UTC", DailyTimeBudgetMinutes: 30})
	if err != nil {
		t.Fatal(err)
	}

	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Sterilize bottles", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning, IsMultistep: true}, []tasks.Subtask{
		{Name: "Wash parts", StepOrder: 1, DurationMinutes: 10, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Sterilize parts", StepOrder: 2, DurationMinutes: 10, TimeOfDayPreference: tasks.TimeOfDayMorning, GapRule: tasks.SubtaskGapRule{MinGapAfterPreviousMinutes: 90}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return user.ID
}

func seedAllDayCalendarBlock(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, blockRepo *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Calendar Large", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Quick reset", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Deep clean", DurationMinutes: 30, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-07-17", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-1", LocalDate: "2026-07-17", Timezone: "UTC", Title: "Family travel", Detail: "large calendar event", IsAllDay: true, Classification: "large", Window: "all-day"}}); err != nil {
		t.Fatal(err)
	}
	return user.ID
}

func seedMediumWindowCalendarBlock(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, blockRepo *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Calendar Medium", Timezone: "UTC", DailyTimeBudgetMinutes: 40})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning admin", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Afternoon outing", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-07-18", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-2", LocalDate: "2026-07-18", Timezone: "UTC", Title: "Pediatrician", Detail: "medium calendar event", Classification: "medium", Window: "afternoon"}}); err != nil {
		t.Fatal(err)
	}
	return user.ID
}

func seedSingleWindowLargeCalendarBlock(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service, blockRepo *store.CalendarBlockRepository) string {
	t.Helper()
	user, err := userService.Create(ctx, users.User{DisplayName: "Calendar Evening Large", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning admin", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Afternoon errands", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Evening project", DurationMinutes: 30, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-07-19", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-3", LocalDate: "2026-07-19", Timezone: "UTC", Title: "Wedding reception", Detail: "large calendar event", Classification: "large", Window: "evening"}}); err != nil {
		t.Fatal(err)
	}
	return user.ID
}

func TestPreviewRangeCarriesOverflowStateIntoTomorrow(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	calendarBlockRepo := store.NewCalendarBlockRepository(sqlDB)
	schedulerService := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, calendarBlockRepo)

	user, err := userService.Create(ctx, users.User{DisplayName: "Preview Range", Timezone: "UTC", DailyTimeBudgetMinutes: 10})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Large task", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityHigh, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}

	results, err := schedulerService.PreviewRange(ctx, user.ID, time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC), 2)
	if err != nil {
		t.Fatalf("PreviewRange() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if len(results[0].Overflowed) != 1 || len(results[1].Overflowed) != 1 {
		t.Fatalf("unexpected overflow counts: day1=%d day2=%d", len(results[0].Overflowed), len(results[1].Overflowed))
	}
	if results[1].Overflowed[0].OriginalScheduledForDate != "2026-07-22" {
		t.Fatalf("tomorrow original date = %s, want 2026-07-22", results[1].Overflowed[0].OriginalScheduledForDate)
	}
	if results[1].Overflowed[0].RolloverCount != 2 {
		t.Fatalf("tomorrow rollover_count = %d, want 2", results[1].Overflowed[0].RolloverCount)
	}
}

func TestSchedulerFitsRealisticCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		seed           func(context.Context, *testing.T, *users.Service, *tasks.Service, *occurrences.Service) string
		day            time.Time
		budget         int
		wantScheduled  int
		wantOverflowed int
		wantSkipped    int
		assert         func(*testing.T, scheduler.PlanResult)
	}{
		{
			name: "laundry and clean kitchen fit in 60 minute day",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service) string {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Laundry + Clean", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
				if err != nil {
					t.Fatal(err)
				}
				laundry, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 25, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAny, IsMultistep: true}, []tasks.Subtask{
					{Name: "Wash", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
					{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayAfternoon},
					{Name: "Fold", StepOrder: 3, DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayEvening},
				})
				if err != nil {
					t.Fatal(err)
				}
				_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Clean kitchen", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
				if err != nil {
					t.Fatal(err)
				}
				// Previous-week completion makes laundry due again today.
				_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: laundry.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-20", OriginalScheduledForDate: "2026-07-20", ScheduledTimeOfDay: tasks.TimeOfDayMorning})
				return user.ID
			},
			day:            time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
			budget:         60,
			wantScheduled:  4,
			wantOverflowed: 0,
			wantSkipped:    0,
			assert: func(t *testing.T, result scheduler.PlanResult) {
				windows := map[string]int{}
				for _, occ := range result.Scheduled {
					windows[string(occ.ScheduledTimeOfDay)]++
				}
				if windows["morning"] != 1 || windows["afternoon"] != 1 || windows["evening"] != 2 {
					t.Fatalf("expected 1 morning, 1 afternoon, 2 evening, got %+v", windows)
				}
			},
		},
		{
			name: "grocery run alone fits in 60 minute day",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service) string {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Grocery", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
				if err != nil {
					t.Fatal(err)
				}
				grocery, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Grocery run", DurationMinutes: 60, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 7, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				// Last grocery run was exactly 7 days ago, so it is due today.
				_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: grocery.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-21", OriginalScheduledForDate: "2026-07-21", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
				return user.ID
			},
			day:            time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
			budget:         60,
			wantScheduled:  1,
			wantOverflowed: 0,
			wantSkipped:    0,
		},
		{
			name: "meal prep and clean kitchen honestly overflow when total exceeds budget",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service) string {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Overflow", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
				if err != nil {
					t.Fatal(err)
				}
				mealPrep, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Meal prep", DurationMinutes: 45, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Clean kitchen", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
				if err != nil {
					t.Fatal(err)
				}
				// Previous-week completion makes meal prep due today.
				_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: mealPrep.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-20", OriginalScheduledForDate: "2026-07-20", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
				return user.ID
			},
			day:            time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
			budget:         60,
			wantScheduled:  1,
			wantOverflowed: 1,
			wantSkipped:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqlDB := openTestDB(t)
			if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
				t.Fatalf("ApplyMigrations() error = %v", err)
			}

			userService := users.NewService(users.NewRepository(sqlDB))
			taskService := tasks.NewService(tasks.NewRepository(sqlDB))
			occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
			checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
			blockRepo := store.NewCalendarBlockRepository(sqlDB)
			svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

			userID := tt.seed(ctx, t, userService, taskService, occurrenceService)
			result, err := svc.PlanDay(ctx, userID, tt.day)
			if err != nil {
				t.Fatalf("PlanDay() error = %v", err)
			}
			if len(result.Scheduled) != tt.wantScheduled {
				t.Fatalf("len(Scheduled) = %d, want %d", len(result.Scheduled), tt.wantScheduled)
			}
			if len(result.Overflowed) != tt.wantOverflowed {
				t.Fatalf("len(Overflowed) = %d, want %d", len(result.Overflowed), tt.wantOverflowed)
			}
			if len(result.Skipped) != tt.wantSkipped {
				t.Fatalf("len(Skipped) = %d, want %d", len(result.Skipped), tt.wantSkipped)
			}
			if tt.assert != nil {
				tt.assert(t, result)
			}
		})
	}
}

func TestSchedulerFallbackAndBlockedWindows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		seed           func(context.Context, *testing.T, *users.Service, *tasks.Service, *store.CalendarBlockRepository) (string, map[string]string)
		day            time.Time
		budget         int
		wantScheduled  int
		wantOverflowed int
		wantSkipped    int
		assert         func(*testing.T, scheduler.PlanResult, map[string]string)
	}{
		{
			name: "evening task falls back to afternoon when evening is blocked and budget allows",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, blockRepo *store.CalendarBlockRepository) (string, map[string]string) {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Fallback", Timezone: "UTC", DailyTimeBudgetMinutes: 90})
				if err != nil {
					t.Fatal(err)
				}
				ids := map[string]string{}
				def, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning routine", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
				if err != nil {
					t.Fatal(err)
				}
				ids[def.Task.ID] = def.Task.Name
				def, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Afternoon admin", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				ids[def.Task.ID] = def.Task.Name
				def, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Evening wind-down", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
				if err != nil {
					t.Fatal(err)
				}
				ids[def.Task.ID] = def.Task.Name
				if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-08-01", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-fb", LocalDate: "2026-08-01", Timezone: "UTC", Title: "Dinner", Classification: "large", Window: "evening"}}); err != nil {
					t.Fatal(err)
				}
				return user.ID, ids
			},
			day:            time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			budget:         90,
			wantScheduled:  3,
			wantOverflowed: 0,
			wantSkipped:    0,
			assert: func(t *testing.T, result scheduler.PlanResult, names map[string]string) {
				var eveningTaskWindow string
				for _, occ := range result.Scheduled {
					if names[occ.TaskID] == "Evening wind-down" {
						eveningTaskWindow = string(occ.ScheduledTimeOfDay)
					}
				}
				if eveningTaskWindow != "afternoon" {
					t.Fatalf("Evening wind-down should fall back to afternoon, got %s", eveningTaskWindow)
				}
			},
		},
		{
			name: "demand-aware budget with blocked window honestly overflows what cannot fit",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, blockRepo *store.CalendarBlockRepository) (string, map[string]string) {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Blocked", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
				if err != nil {
					t.Fatal(err)
				}
				ids := map[string]string{}
				def, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning routine", DurationMinutes: 30, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
				if err != nil {
					t.Fatal(err)
				}
				ids[def.Task.ID] = def.Task.Name
				def, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Afternoon admin", DurationMinutes: 30, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				ids[def.Task.ID] = def.Task.Name
				if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-08-02", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-blk", LocalDate: "2026-08-02", Timezone: "UTC", Title: "Pediatrician", Classification: "medium", Window: "afternoon"}}); err != nil {
					t.Fatal(err)
				}
				return user.ID, ids
			},
			day:            time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
			budget:         60,
			wantScheduled:  1,
			wantOverflowed: 1,
			wantSkipped:    0,
			assert: func(t *testing.T, result scheduler.PlanResult, names map[string]string) {
				if len(result.Scheduled) != 1 || names[result.Scheduled[0].TaskID] != "Morning routine" {
					t.Fatalf("expected Morning routine scheduled, got %+v", result.Scheduled)
				}
				if len(result.Overflowed) != 1 || names[result.Overflowed[0].TaskID] != "Afternoon admin" {
					t.Fatalf("expected Afternoon admin overflowed, got %+v", result.Overflowed)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqlDB := openTestDB(t)
			if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
				t.Fatalf("ApplyMigrations() error = %v", err)
			}

			userService := users.NewService(users.NewRepository(sqlDB))
			taskService := tasks.NewService(tasks.NewRepository(sqlDB))
			occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
			checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
			blockRepo := store.NewCalendarBlockRepository(sqlDB)
			svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

			userID, names := tt.seed(ctx, t, userService, taskService, blockRepo)
			result, err := svc.PlanDay(ctx, userID, tt.day)
			if err != nil {
				t.Fatalf("PlanDay() error = %v", err)
			}
			if len(result.Scheduled) != tt.wantScheduled {
				t.Fatalf("len(Scheduled) = %d, want %d", len(result.Scheduled), tt.wantScheduled)
			}
			if len(result.Overflowed) != tt.wantOverflowed {
				t.Fatalf("len(Overflowed) = %d, want %d", len(result.Overflowed), tt.wantOverflowed)
			}
			if len(result.Skipped) != tt.wantSkipped {
				t.Fatalf("len(Skipped) = %d, want %d", len(result.Skipped), tt.wantSkipped)
			}
			if tt.assert != nil {
				tt.assert(t, result, names)
			}
		})
	}
}

func TestSchedulerSpreadsRecurringTasksAcrossWeek(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	blockRepo := store.NewCalendarBlockRepository(sqlDB)
	svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

	user, err := userService.Create(ctx, users.User{DisplayName: "Spread Week", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}

	clean, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Clean kitchen", DurationMinutes: 20, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
	if err != nil {
		t.Fatal(err)
	}
	grocery, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Grocery run", DurationMinutes: 60, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 7, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mealPrep, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Meal prep", DurationMinutes: 45, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}
	laundry, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 25, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAny, IsMultistep: true}, []tasks.Subtask{
		{Name: "Wash", StepOrder: 1, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{Name: "Move to dryer", StepOrder: 2, DurationMinutes: 5, TimeOfDayPreference: tasks.TimeOfDayAfternoon},
		{Name: "Fold", StepOrder: 3, DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayEvening},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Seed one earlier completion for the 2x/week tasks so the week still needs
	// one more occurrence of each and the scheduler must place that follow-up later.
	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: mealPrep.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-27", OriginalScheduledForDate: "2026-07-27", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
	for _, subtask := range laundry.Subtasks {
		_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: laundry.Task.ID, SubtaskID: subtask.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-27", OriginalScheduledForDate: "2026-07-27", ScheduledTimeOfDay: subtask.TimeOfDayPreference})
	}
	_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: grocery.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-21", OriginalScheduledForDate: "2026-07-21", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})

	cleanDays := map[string]bool{}
	laundryDays := map[string]bool{}
	mealPrepDays := map[string]bool{}
	groceryDays := map[string]bool{}
	overflowedTaskIDs := map[string]bool{}
	totalOverflowed := 0

	for i := 0; i < 6; i++ {
		day := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i)
		result, err := svc.PlanDay(ctx, user.ID, day)
		if err != nil {
			t.Fatalf("PlanDay(%s) error = %v", day.Format("2006-01-02"), err)
		}
		used := 0
		for _, occ := range result.Scheduled {
			switch occ.TaskID {
			case clean.Task.ID:
				cleanDays[result.Date] = true
				used += 20
			case grocery.Task.ID:
				groceryDays[result.Date] = true
				used += 60
			case mealPrep.Task.ID:
				mealPrepDays[result.Date] = true
				used += 45
			case laundry.Task.ID:
				laundryDays[result.Date] = true
				if occ.SubtaskID != "" {
					switch occ.ScheduledTimeOfDay {
					case tasks.TimeOfDayMorning, tasks.TimeOfDayAfternoon:
						used += 5
					case tasks.TimeOfDayEvening:
						used += 15
					}
				}
			}
			occ.Status = occurrences.StatusCompleted
			now := day.Add(12 * time.Hour)
			occ.CompletedAt = &now
			if _, err := occurrenceService.Update(ctx, occ); err != nil {
				t.Fatalf("Update() error = %v", err)
			}
		}
		for _, occ := range result.Overflowed {
			overflowedTaskIDs[occ.TaskID] = true
		}
		if used > 60 {
			t.Fatalf("day %s scheduled %d minutes, want <= 60", result.Date, used)
		}
		totalOverflowed += len(result.Overflowed)
	}

	if len(cleanDays) != 6 {
		t.Fatalf("clean kitchen scheduled on %d days, want 6", len(cleanDays))
	}
	if len(laundryDays) != 1 {
		t.Fatalf("laundry scheduled on %d days, want 1 second-weekly occurrence", len(laundryDays))
	}
	for day := range laundryDays {
		if day < "2026-07-30" {
			t.Fatalf("laundry scheduled too early on %s, want spread after 2026-07-27 completion", day)
		}
	}
	if len(groceryDays) != 0 {
		t.Fatalf("grocery run scheduled on %d days, want 0 because it cannot fit alongside the daily task", len(groceryDays))
	}
	if len(mealPrepDays) > 1 {
		t.Fatalf("meal prep scheduled on %d days, want at most 1 spread occurrence", len(mealPrepDays))
	}
	if !overflowedTaskIDs[grocery.Task.ID] {
		t.Fatal("expected grocery run to overflow honestly when no day in the horizon has room")
	}
	if totalOverflowed == 0 {
		t.Fatal("expected honest overflow for work that cannot fit anywhere in the horizon")
	}
}

func TestSchedulerTimezoneAware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		timezone    string
		day         time.Time
		wantDateStr string
	}{
		{
			name:        "America/Chicago user gets local date and window ready times",
			timezone:    "America/Chicago",
			day:         time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
			wantDateStr: "2026-07-25",
		},
		{
			name:        "Asia/Tokyo user gets local date and window ready times",
			timezone:    "Asia/Tokyo",
			day:         time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
			wantDateStr: "2026-07-25",
		},
		{
			name:        "UTC user is unchanged",
			timezone:    "UTC",
			day:         time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
			wantDateStr: "2026-07-25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqlDB := openTestDB(t)
			if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
				t.Fatalf("ApplyMigrations() error = %v", err)
			}

			userService := users.NewService(users.NewRepository(sqlDB))
			taskService := tasks.NewService(tasks.NewRepository(sqlDB))
			occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
			checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
			blockRepo := store.NewCalendarBlockRepository(sqlDB)
			svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

			user, err := userService.Create(ctx, users.User{DisplayName: "TZ", Timezone: tt.timezone, DailyTimeBudgetMinutes: 60})
			if err != nil {
				t.Fatal(err)
			}
			_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Morning routine", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Afternoon routine", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Evening routine", DurationMinutes: 10, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayEvening}, nil)
			if err != nil {
				t.Fatal(err)
			}

			result, err := svc.PlanDay(ctx, user.ID, tt.day)
			if err != nil {
				t.Fatalf("PlanDay() error = %v", err)
			}
			if result.Date != tt.wantDateStr {
				t.Fatalf("result.Date = %s, want %s", result.Date, tt.wantDateStr)
			}
			if len(result.Scheduled) != 3 {
				t.Fatalf("len(Scheduled) = %d, want 3", len(result.Scheduled))
			}

			loc, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Fatal(err)
			}
			wantHour := map[tasks.TimeOfDayPreference]int{
				tasks.TimeOfDayMorning:   8,
				tasks.TimeOfDayAfternoon: 12,
				tasks.TimeOfDayEvening:   16,
			}
			for _, occ := range result.Scheduled {
				if occ.ReadyAt == nil {
					t.Fatal("ReadyAt is nil")
				}
				localReady := occ.ReadyAt.In(loc)
				if localReady.Hour() != wantHour[occ.ScheduledTimeOfDay] {
					t.Fatalf("%s local ready hour = %d, want %d (readyAt=%s)", occ.ScheduledTimeOfDay, localReady.Hour(), wantHour[occ.ScheduledTimeOfDay], occ.ReadyAt)
				}
				if y, m, d := localReady.Date(); fmt.Sprintf("%04d-%02d-%02d", y, m, d) != tt.wantDateStr {
					t.Fatalf("%s local ready date = %04d-%02d-%02d, want %s", occ.ScheduledTimeOfDay, y, m, d, tt.wantDateStr)
				}
			}
		})
	}
}

func TestSchedulerCalendarAwareDaySelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		seed           func(context.Context, *testing.T, *users.Service, *tasks.Service, *occurrences.Service, *store.CalendarBlockRepository) (string, string)
		start   time.Time
		wantDay string
	}{
		{
			name: "grocery moves from afternoon-blocked day to free day",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service, blockRepo *store.CalendarBlockRepository) (string, string) {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Calendar aware", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
				if err != nil {
					t.Fatal(err)
				}
				grocery, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Grocery run", DurationMinutes: 60, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 7, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: grocery.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-28", OriginalScheduledForDate: "2026-07-28", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
				if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-08-04", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-1", LocalDate: "2026-08-04", Timezone: "UTC", Title: "Pediatrician", Classification: "medium", Window: "afternoon"}}); err != nil {
					t.Fatal(err)
				}
				return user.ID, grocery.Task.ID
			},
			start:   time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC),
			wantDay: "2026-08-05",
		},
		{
			name: "heavy grocery moves from all-day event day to free day",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service, blockRepo *store.CalendarBlockRepository) (string, string) {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Calendar aware", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
				if err != nil {
					t.Fatal(err)
				}
				grocery, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Grocery run", DurationMinutes: 60, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 7, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: grocery.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-28", OriginalScheduledForDate: "2026-07-28", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
				if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-08-04", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-2", LocalDate: "2026-08-04", Timezone: "UTC", Title: "Family travel", Classification: "large", Window: "all-day", IsAllDay: true}}); err != nil {
					t.Fatal(err)
				}
				return user.ID, grocery.Task.ID
			},
			start:   time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC),
			wantDay: "2026-08-05",
		},
		{
			name: "count laundry moves from blocked default day to next free day",
			seed: func(ctx context.Context, t *testing.T, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service, blockRepo *store.CalendarBlockRepository) (string, string) {
				t.Helper()
				user, err := userService.Create(ctx, users.User{DisplayName: "Calendar aware", Timezone: "UTC", DailyTimeBudgetMinutes: 80})
				if err != nil {
					t.Fatal(err)
				}
				laundry, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Laundry", DurationMinutes: 30, CadenceType: tasks.CadenceTypeCount, CadenceValue: 2, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon}, nil)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = occurrenceService.Create(ctx, occurrences.Occurrence{UserID: user.ID, TaskID: laundry.Task.ID, Status: occurrences.StatusCompleted, ScheduledForDate: "2026-07-27", OriginalScheduledForDate: "2026-07-27", ScheduledTimeOfDay: tasks.TimeOfDayAfternoon})
				if err := blockRepo.ReplaceDay(ctx, user.ID, "google", "2026-07-30", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-3", LocalDate: "2026-07-30", Timezone: "UTC", Title: "Afternoon meeting", Classification: "medium", Window: "afternoon"}}); err != nil {
					t.Fatal(err)
				}
				return user.ID, laundry.Task.ID
			},
			start:   time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
			wantDay: "2026-07-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			sqlDB := openTestDB(t)
			if err := store.ApplyMigrations(ctx, sqlDB); err != nil {
				t.Fatalf("ApplyMigrations() error = %v", err)
			}

			userService := users.NewService(users.NewRepository(sqlDB))
			taskService := tasks.NewService(tasks.NewRepository(sqlDB))
			occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
			checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
			blockRepo := store.NewCalendarBlockRepository(sqlDB)
			svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, blockRepo)

			userID, taskID := tt.seed(ctx, t, userService, taskService, occurrenceService, blockRepo)

			day1, err := svc.PlanDay(ctx, userID, tt.start)
			if err != nil {
				t.Fatalf("PlanDay(day1) error = %v", err)
			}
			for _, occ := range day1.Scheduled {
				if occ.TaskID == taskID {
					t.Fatalf("task should not schedule on blocked day %s", day1.Date)
				}
			}
			if _, ok := day1.Reasons[taskID]; !ok {
				t.Fatal("expected a day-selection reason on day 1")
			}

			day2, err := svc.PlanDay(ctx, userID, tt.start.AddDate(0, 0, 1))
			if err != nil {
				t.Fatalf("PlanDay(day2) error = %v", err)
			}
			var scheduledDay string
			for _, occ := range day2.Scheduled {
				if occ.TaskID == taskID {
					scheduledDay = day2.Date
				}
			}
			if scheduledDay != tt.wantDay {
				t.Fatalf("task scheduled on %s, want %s", scheduledDay, tt.wantDay)
			}
		})
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
