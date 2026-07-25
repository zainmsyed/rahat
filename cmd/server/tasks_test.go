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
	occ "github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
)

type testTaskRuntime struct {
	mux       *http.ServeMux
	tasks     *taskpkg.Service
	users     *usr.Service
	auth      *auth.Service
	scheduler *scheduler.Service
}

func newTestTaskRuntime(t *testing.T) *testTaskRuntime {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil { t.Fatalf("OpenSQLite() error = %v", err) }
	t.Cleanup(func() { _ = sqlDB.Close() })
	users := usr.NewService(usr.NewRepository(sqlDB))
	tasks := taskpkg.NewService(taskpkg.NewRepository(sqlDB))
	authSvc := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-web-session-secret", 30*24*time.Hour)
	occurrences := occ.NewService(occ.NewRepository(sqlDB))
	sched := scheduler.NewService(users, tasks, occurrences, store.NewScheduleCheckpointRepository(sqlDB), store.NewCalendarBlockRepository(sqlDB))
	authHandler := &authHandler{auth: authSvc, users: users, webOrigin: "http://localhost:5200", appEnv: "development"}
	mux := http.NewServeMux()
	(&taskManagementHandler{auth: authHandler, tasks: tasks}).register(mux)
	return &testTaskRuntime{mux: mux, tasks: tasks, users: users, auth: authSvc, scheduler: sched}
}

func (rt *testTaskRuntime) cookieFor(t *testing.T, userID string) *http.Cookie {
	t.Helper()
	_, raw, err := rt.auth.CreateSessionForUser(context.Background(), userID)
	if err != nil { t.Fatalf("CreateSessionForUser() error = %v", err) }
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
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	rec := httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("create status = %d: %s", rec.Code, rec.Body.String()) }
	var created onboardingTaskResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	req = httptest.NewRequest(http.MethodPost, "/tasks/"+created.ID+"/pause", bytes.NewReader([]byte(`{"paused":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(otherCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK { t.Fatal("other user paused owner's task") }

	req = httptest.NewRequest(http.MethodPost, "/tasks/"+created.ID+"/pause", bytes.NewReader([]byte(`{"paused":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("pause status = %d: %s", rec.Code, rec.Body.String()) }
	plan, err := rt.scheduler.PreviewDay(ctx, owner.ID, time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC))
	if err != nil { t.Fatalf("PreviewDay() error = %v", err) }
	if len(plan.Scheduled) != 0 { t.Fatalf("paused task scheduled %d items", len(plan.Scheduled)) }

	req = httptest.NewRequest(http.MethodPost, "/tasks/"+created.ID+"/pause", bytes.NewReader([]byte(`{"paused":false}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("resume status = %d: %s", rec.Code, rec.Body.String()) }
	plan, err = rt.scheduler.PreviewDay(ctx, owner.ID, time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC))
	if err != nil { t.Fatalf("PreviewDay() error = %v", err) }
	if len(plan.Scheduled) == 0 { t.Fatal("resumed task was not scheduled") }

	req = httptest.NewRequest(http.MethodDelete, "/tasks/"+created.ID, http.NoBody)
	req.AddCookie(ownerCookie)
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent { t.Fatalf("archive status = %d: %s", rec.Code, rec.Body.String()) }
	archived, err := rt.tasks.GetTaskWithSubtasks(ctx, created.ID)
	if err != nil { t.Fatalf("archived task missing: %v", err) }
	if archived.Task.ArchivedAt == nil { t.Fatal("archived_at not set") }
	plan, err = rt.scheduler.PreviewDay(ctx, owner.ID, time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC))
	if err != nil { t.Fatalf("PreviewDay() error = %v", err) }
	if len(plan.Scheduled) != 0 { t.Fatalf("archived task scheduled %d items", len(plan.Scheduled)) }
}

func TestTaskManagementValidationAndAuthentication(t *testing.T) {
	rt := newTestTaskRuntime(t)
	req := httptest.NewRequest(http.MethodGet, "/tasks", http.NoBody)
	rec := httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized { t.Fatalf("unauth list status = %d", rec.Code) }

	user, _ := rt.users.Create(context.Background(), usr.User{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60})
	req = httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte(`{"name":""}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(rt.cookieFor(t, user.ID))
	rec = httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest { t.Fatalf("invalid create status = %d", rec.Code) }
}
