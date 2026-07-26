package main

import (
	"context"
	"database/sql"
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

type testScheduleRuntime struct {
	mux       *http.ServeMux
	db        *sql.DB
	auth      *auth.Service
	users     *usr.Service
	tasks     *taskpkg.Service
	occurrences *occ.Service
	scheduler *scheduler.Service
	blocks    *store.CalendarBlockRepository
}

func newTestScheduleRuntime(t *testing.T) *testScheduleRuntime {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	userService := usr.NewService(usr.NewRepository(sqlDB))
	taskService := taskpkg.NewService(taskpkg.NewRepository(sqlDB))
	occurrenceService := occ.NewService(occ.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	calendarBlockRepo := store.NewCalendarBlockRepository(sqlDB)
	schedulerService := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, calendarBlockRepo)
	authService := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-web-session-secret", 30*24*time.Hour)
	webAuth := &authHandler{auth: authService, users: userService, webOrigin: "http://localhost:5200", appEnv: "development"}

	mux := http.NewServeMux()
	webAuth.register(mux)
	mux.HandleFunc("GET /schedule/plan", requireAuthenticatedUserForRoute(webAuth, func(w http.ResponseWriter, r *http.Request, current authenticatedUser) {
		dateValue := r.URL.Query().Get("date")
		var day time.Time
		if dateValue == "" {
			day = localDateAsUTC(current.User.Timezone, time.Now())
		} else {
			day = parseDay(dateValue)
		}
		result, err := schedulerService.PlanDay(r.Context(), current.User.ID, day)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}))

	return &testScheduleRuntime{
		mux:         mux,
		db:          sqlDB,
		auth:        authService,
		users:       userService,
		tasks:       taskService,
		occurrences: occurrenceService,
		scheduler:   schedulerService,
		blocks:      calendarBlockRepo,
	}
}

func (rt *testScheduleRuntime) cookieFor(t *testing.T, userID string) *http.Cookie {
	t.Helper()
	_, raw, err := rt.auth.CreateSessionForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("CreateSessionForUser() error = %v", err)
	}
	return &http.Cookie{Name: sessionCookieName, Value: raw}
}

func TestSchedulePlanReturnsDaySelectionReasons(t *testing.T) {
	rt := newTestScheduleRuntime(t)
	ctx := context.Background()

	user, err := rt.users.Create(ctx, usr.User{DisplayName: "Reasons", Timezone: "UTC", DailyTimeBudgetMinutes: 80})
	if err != nil {
		t.Fatal(err)
	}

	grocery, err := rt.tasks.CreateTaskWithSubtasks(ctx, taskpkg.Task{UserID: user.ID, Name: "Grocery run", DurationMinutes: 60, CadenceType: taskpkg.CadenceTypeInterval, CadenceValue: 7, Priority: taskpkg.PriorityMedium, TimeOfDayPreference: taskpkg.TimeOfDayAfternoon}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = rt.occurrences.Create(ctx, occ.Occurrence{UserID: user.ID, TaskID: grocery.Task.ID, Status: occ.StatusCompleted, ScheduledForDate: "2026-07-28", OriginalScheduledForDate: "2026-07-28", ScheduledTimeOfDay: taskpkg.TimeOfDayAfternoon})
	if err := rt.blocks.ReplaceDay(ctx, user.ID, "google", "2026-08-04", []store.CalendarBlock{{UserID: user.ID, Provider: "google", ExternalEventID: "evt-1", LocalDate: "2026-08-04", Timezone: "UTC", Title: "Afternoon busy", Classification: "medium", Window: "afternoon"}}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/schedule/plan?date=2026-08-04", http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	req.AddCookie(rt.cookieFor(t, user.ID))
	rec := httptest.NewRecorder()
	rt.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Date      string            `json:"Date"`
		Reasons   map[string]string `json:"Reasons"`
		Scheduled []any             `json:"Scheduled"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Date != "2026-08-04" {
		t.Fatalf("date = %q, want 2026-08-04", result.Date)
	}
	reason, ok := result.Reasons[grocery.Task.ID]
	if !ok {
		t.Fatalf("expected reason for task %s in response %s", grocery.Task.ID, rec.Body.String())
	}
	if reason == "" {
		t.Fatal("reason should not be empty")
	}
	if len(result.Scheduled) != 0 {
		t.Fatalf("expected grocery to be moved, but %d items scheduled", len(result.Scheduled))
	}
}
