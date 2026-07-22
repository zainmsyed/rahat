package tasks_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
)

func newTestServices(t *testing.T) (*sql.DB, *tasks.Service, *usr.Service) {
	t.Helper()

	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	return sqlDB, tasks.NewService(tasks.NewRepository(sqlDB)), usr.NewService(usr.NewRepository(sqlDB))
}

func TestReplaceTaskWithSubtasksCreatesTask(t *testing.T) {
	_, taskService, userService := newTestServices(t)

	ctx := context.Background()
	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	task, err := taskService.ReplaceTaskWithSubtasks(ctx, tasks.Task{
		UserID:              user.ID,
		Name:                "Laundry",
		DurationMinutes:     30,
		CadenceType:         tasks.CadenceTypeInterval,
		CadenceValue:        1,
		Priority:            tasks.PriorityMedium,
		TimeOfDayPreference: tasks.TimeOfDayMorning,
		IsMultistep:         true,
	}, []tasks.Subtask{
		{StepOrder: 1, Name: "Wash", DurationMinutes: 10, TimeOfDayPreference: tasks.TimeOfDayMorning},
		{StepOrder: 2, Name: "Dry", DurationMinutes: 20, TimeOfDayPreference: tasks.TimeOfDayAfternoon},
	})
	if err != nil {
		t.Fatalf("ReplaceTaskWithSubtasks error = %v", err)
	}
	if task.Task.ID == "" {
		t.Fatal("expected task id")
	}
	if len(task.Subtasks) != 2 {
		t.Fatalf("subtasks = %d, want 2", len(task.Subtasks))
	}
	if task.Task.DurationMinutes != 30 {
		t.Fatalf("duration = %d, want 30", task.Task.DurationMinutes)
	}
	if !task.Task.IsMultistep {
		t.Fatal("expected IsMultistep true")
	}
}

func TestReplaceTaskWithSubtasksUpdatesTask(t *testing.T) {
	_, taskService, userService := newTestServices(t)

	ctx := context.Background()
	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	created, err := taskService.ReplaceTaskWithSubtasks(ctx, tasks.Task{
		UserID:              user.ID,
		Name:                "Original",
		DurationMinutes:     15,
		CadenceType:         tasks.CadenceTypeInterval,
		CadenceValue:        1,
		Priority:            tasks.PriorityMedium,
		TimeOfDayPreference: tasks.TimeOfDayMorning,
	}, []tasks.Subtask{
		{StepOrder: 1, Name: "Step A", DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayMorning},
	})
	if err != nil {
		t.Fatalf("create error = %v", err)
	}

	updated, err := taskService.ReplaceTaskWithSubtasks(ctx, tasks.Task{
		ID:                  created.Task.ID,
		UserID:              user.ID,
		Name:                "Updated",
		DurationMinutes:     25,
		CadenceType:         tasks.CadenceTypeInterval,
		CadenceValue:        2,
		Priority:            tasks.PriorityHigh,
		TimeOfDayPreference: tasks.TimeOfDayAfternoon,
	}, []tasks.Subtask{
		{StepOrder: 1, Name: "Step X", DurationMinutes: 10, TimeOfDayPreference: tasks.TimeOfDayAfternoon},
		{StepOrder: 2, Name: "Step Y", DurationMinutes: 15, TimeOfDayPreference: tasks.TimeOfDayEvening},
	})
	if err != nil {
		t.Fatalf("update error = %v", err)
	}
	if updated.Task.Name != "Updated" {
		t.Fatalf("name = %q, want Updated", updated.Task.Name)
	}
	if updated.Task.DurationMinutes != 25 {
		t.Fatalf("duration = %d, want 25", updated.Task.DurationMinutes)
	}
	if len(updated.Subtasks) != 2 {
		t.Fatalf("subtasks = %d, want 2", len(updated.Subtasks))
	}

	loaded, err := taskService.GetTaskWithSubtasks(ctx, created.Task.ID)
	if err != nil {
		t.Fatalf("load error = %v", err)
	}
	if len(loaded.Subtasks) != 2 {
		t.Fatalf("loaded subtasks = %d, want 2", len(loaded.Subtasks))
	}
}

func TestCreateTaskFromStarterTemplate(t *testing.T) {
	_, taskService, userService := newTestServices(t)

	ctx := context.Background()
	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	templates, err := taskService.ListStarterTaskTemplates(ctx)
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	if len(templates) == 0 {
		t.Skip("no starter templates seeded")
	}

	created, err := taskService.CreateTaskFromStarterTemplate(ctx, user.ID, templates[0].ID)
	if err != nil {
		t.Fatalf("CreateTaskFromStarterTemplate error = %v", err)
	}
	if created.Task.Name != templates[0].Name {
		t.Fatalf("name = %q, want %q", created.Task.Name, templates[0].Name)
	}
	if created.Task.UserID != user.ID {
		t.Fatal("task user id mismatch")
	}
}

func TestCreateTaskFromStarterTemplateNotFound(t *testing.T) {
	_, taskService, userService := newTestServices(t)

	ctx := context.Background()
	user, err := userService.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	_, err = taskService.CreateTaskFromStarterTemplate(ctx, user.ID, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
}
