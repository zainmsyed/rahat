package scheduler_test

import (
	"context"
	"database/sql"
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
		seed                     func(context.Context, *testing.T, *sql.DB, *users.Service, *tasks.Service, *occurrences.Service) string
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
			svc := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo)

			userID := tt.seed(ctx, t, sqlDB, userService, taskService, occurrenceService)
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

func seedNormalDay(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service) string {
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

func seedOverloadedDay(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service) string {
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

func seedLaundryDay(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service) string {
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

func seedWeeklyCountMultistep(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, occurrenceService *occurrences.Service) string {
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

func seedSameWindowGap(ctx context.Context, t *testing.T, _ *sql.DB, userService *users.Service, taskService *tasks.Service, _ *occurrences.Service) string {
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

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}
