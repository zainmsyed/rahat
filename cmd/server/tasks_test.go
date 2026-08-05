package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/auth"
	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	occ "github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
)

type testTaskRuntime struct {
	mux         *http.ServeMux
	tasks       *taskpkg.Service
	users       *usr.Service
	auth        *auth.Service
	scheduler   *scheduler.Service
	occurrences *occ.Service
	events      *events.Service
}

func newTestTaskRuntime(t *testing.T) *testTaskRuntime {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	users := usr.NewService(usr.NewRepository(sqlDB))
	tasks := taskpkg.NewService(taskpkg.NewRepository(sqlDB))
	authSvc := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-web-session-secret", 30*24*time.Hour)
	occurrences := occ.NewService(occ.NewRepository(sqlDB))
	eventService := events.NewService(events.NewRepository(sqlDB))
	sched := scheduler.NewService(users, tasks, occurrences, store.NewScheduleCheckpointRepository(sqlDB), store.NewCalendarBlockRepository(sqlDB))
	authHandler := &authHandler{auth: authSvc, users: users, webOrigin: "http://localhost:5200", appEnv: "development"}
	mux := http.NewServeMux()
	(&taskManagementHandler{auth: authHandler, tasks: tasks}).register(mux)
	return &testTaskRuntime{mux: mux, tasks: tasks, users: users, auth: authSvc, scheduler: sched, occurrences: occurrences, events: eventService}
}

func (rt *testTaskRuntime) cookieFor(t *testing.T, userID string) *http.Cookie {
	t.Helper()
	_, raw, err := rt.auth.CreateSessionForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("CreateSessionForUser() error = %v", err)
	}
	return &http.Cookie{Name: sessionCookieName, Value: raw}
}

func TestTaskManagementOwnershipPauseArchiveAndScheduling(t *testing.T) {
	rt := newTestTaskRuntime(t)
	ctx := context.Background()
	owner, _ := rt.users.Create(ctx, usr.User{DisplayName: "Owner", Timezone: "UTC", DailyTimeBudgetMinutes: 120})
	other, _ := rt.users.Create(ctx, usr.User{DisplayName: "Other", Timezone: "UTC", DailyTimeBudgetMinutes: 120})
	ownerCookie := rt.cookieFor(t, owner.ID)
	otherCookie := rt.cookieFor(t, other.ID)

	payload := onboardingTaskRequest{Name: "Water plants", DurationMinutes: 15, CadenceType: taskpkg.CadenceTypeInterval, CadenceValue: 1, Priority: taskpkg.PriorityMedium, TimeOfDayPreference: taskpkg.TimeOfDayMorning}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(ownerCookie)
	rec := httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d: %s", rec.Code, rec.Body.String())
	}
	var created onboardingTaskResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	req = httptest.NewRequest(http.MethodPost, "/api/tasks/"+created.ID+"/pause", bytes.NewReader([]byte(`{"paused":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(otherCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("other user paused owner's task")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/tasks/"+created.ID+"/pause", bytes.NewReader([]byte(`{"paused":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(ownerCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pause status = %d: %s", rec.Code, rec.Body.String())
	}
	plan, err := rt.scheduler.PreviewDay(ctx, owner.ID, time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PreviewDay() error = %v", err)
	}
	if len(plan.Scheduled) != 0 {
		t.Fatalf("paused task scheduled %d items", len(plan.Scheduled))
	}

	req = httptest.NewRequest(http.MethodPost, "/api/tasks/"+created.ID+"/pause", bytes.NewReader([]byte(`{"paused":false}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(ownerCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("resume status = %d: %s", rec.Code, rec.Body.String())
	}
	plan, err = rt.scheduler.PreviewDay(ctx, owner.ID, time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PreviewDay() error = %v", err)
	}
	if len(plan.Scheduled) == 0 {
		t.Fatal("resumed task was not scheduled")
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/tasks/"+created.ID, http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(ownerCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("archive status = %d: %s", rec.Code, rec.Body.String())
	}
	archived, err := rt.tasks.GetTaskWithSubtasks(ctx, created.ID)
	if err != nil {
		t.Fatalf("archived task missing: %v", err)
	}
	if archived.Task.ArchivedAt == nil {
		t.Fatal("archived_at not set")
	}
	plan, err = rt.scheduler.PreviewDay(ctx, owner.ID, time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PreviewDay() error = %v", err)
	}
	if len(plan.Scheduled) != 0 {
		t.Fatalf("archived task scheduled %d items", len(plan.Scheduled))
	}
}

func TestTaskManagementValidationAndAuthentication(t *testing.T) {
	rt := newTestTaskRuntime(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks", http.NoBody)
	rec := httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth list status = %d", rec.Code)
	}

	user, _ := rt.users.Create(context.Background(), usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	req = httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader([]byte(`{"name":""}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(rt.cookieFor(t, user.ID))
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid create status = %d", rec.Code)
	}
}

func TestTaskEditPreservesCompletedOccurrenceAndEventHistory(t *testing.T) {
	rt := newTestTaskRuntime(t)
	ctx := context.Background()
	user, err := rt.users.Create(ctx, usr.User{DisplayName: "Historian", Timezone: "UTC", DailyTimeBudgetMinutes: 90})
	if err != nil {
		t.Fatal(err)
	}
	cookie := rt.cookieFor(t, user.ID)
	payload := onboardingTaskRequest{
		Name:                "Laundry",
		CadenceType:         taskpkg.CadenceTypeInterval,
		CadenceValue:        2,
		Priority:            taskpkg.PriorityMedium,
		TimeOfDayPreference: taskpkg.TimeOfDayMorning,
		Subtasks: []onboardingSubtaskRequest{{
			Name:                "Wash",
			DurationMinutes:     20,
			TimeOfDayPreference: taskpkg.TimeOfDayMorning,
		}},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d: %s", rec.Code, rec.Body.String())
	}
	var created onboardingTaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	oldSubtaskID := created.Subtasks[0].ID
	completedAt := time.Now().UTC()
	occurrence, err := rt.occurrences.Create(ctx, occ.Occurrence{
		UserID:                   user.ID,
		TaskID:                   created.ID,
		SubtaskID:                oldSubtaskID,
		Status:                   occ.StatusCompleted,
		ScheduledForDate:         "2026-07-24",
		OriginalScheduledForDate: "2026-07-24",
		ScheduledTimeOfDay:       taskpkg.TimeOfDayMorning,
		CompletedAt:              &completedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	event, err := rt.events.Create(ctx, events.EventLog{UserID: user.ID, OccurrenceID: occurrence.ID, Channel: "system", EventType: "completed", MessageType: "task_history"})
	if err != nil {
		t.Fatal(err)
	}

	payload.Name = "Laundry updated"
	payload.Subtasks[0].ID = oldSubtaskID
	payload.Subtasks[0].Name = "Wash carefully"
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("edit status = %d: %s", rec.Code, rec.Body.String())
	}
	var updated onboardingTaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Subtasks[0].ID != oldSubtaskID {
		t.Fatalf("retained subtask ID = %s, want %s", updated.Subtasks[0].ID, oldSubtaskID)
	}

	payload.DurationMinutes = 20
	payload.Subtasks = nil
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest(http.MethodPut, "/api/tasks/"+created.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove subtask edit status = %d: %s", rec.Code, rec.Body.String())
	}
	preserved, err := rt.occurrences.GetByID(ctx, occurrence.ID)
	if err != nil {
		t.Fatalf("completed occurrence was deleted: %v", err)
	}
	if preserved.SubtaskID != oldSubtaskID || preserved.Status != occ.StatusCompleted {
		t.Fatalf("completed occurrence changed: %+v", preserved)
	}
	eventsForUser, err := rt.events.ListByUser(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(eventsForUser) != 1 || eventsForUser[0].ID != event.ID || eventsForUser[0].OccurrenceID != occurrence.ID {
		t.Fatalf("event history was not preserved: %+v", eventsForUser)
	}
}

func TestTaskMutationsRejectUntrustedOrigins(t *testing.T) {
	rt := newTestTaskRuntime(t)
	ctx := context.Background()
	user, err := rt.users.Create(ctx, usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	cookie := rt.cookieFor(t, user.ID)
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		origin string
	}{
		{name: "create missing origin", method: http.MethodPost, path: "/api/tasks", body: `{}`},
		{name: "create wrong origin", method: http.MethodPost, path: "/api/tasks", body: `{}`, origin: "http://attacker.localhost:5200"},
		{name: "update wrong origin", method: http.MethodPut, path: "/api/tasks/task-id", body: `{}`, origin: "http://attacker.localhost:5200"},
		{name: "pause wrong origin", method: http.MethodPost, path: "/api/tasks/task-id/pause", body: `{"paused":true}`, origin: "http://attacker.localhost:5200"},
		{name: "remove wrong origin", method: http.MethodDelete, path: "/api/tasks/task-id", origin: "http://attacker.localhost:5200"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, bytes.NewReader([]byte(test.body)))
			req.Header.Set("Content-Type", "application/json")
			if test.origin != "" {
				req.Header.Set("Origin", test.origin)
			}
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()
			rt.mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusForbidden, rec.Body.String())
			}
		})
	}
}
