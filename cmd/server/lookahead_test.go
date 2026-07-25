package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/tokens"
	"github.com/rahat/rahat/internal/users"
)

func newTestLookaheadHandler(t *testing.T) (*lookaheadHandler, string) {
	t.Helper()
	ctx := context.Background()
	sqlDB, err := db.OpenSQLite(ctx, filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	userService := users.NewService(users.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occurrences.NewService(occurrences.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	calendarBlockRepo := store.NewCalendarBlockRepository(sqlDB)
	schedulerService := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, calendarBlockRepo)
	user, err := userService.Create(ctx, users.User{DisplayName: "Look Ahead", Timezone: "UTC", DailyTimeBudgetMinutes: 45})
	if err != nil {
		t.Fatal(err)
	}
	_, err = taskService.ReplaceTaskWithSubtasks(ctx, tasks.Task{UserID: user.ID, Name: "Water plants", DurationMinutes: 15, CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1, Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning}, nil)
	if err != nil {
		t.Fatal(err)
	}
	mgr := tokens.NewManager("0123456789abcdef")
	return &lookaheadHandler{tokens: mgr, users: userService, tasks: taskService, scheduler: schedulerService, allowTokenIssue: true, clock: func() time.Time { return time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC) }}, user.ID
}

func TestLookaheadPlanRequiresValidToken(t *testing.T) {
	h, _ := newTestLookaheadHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	req := httptest.NewRequest(http.MethodGet, "/lookahead/plan", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	req = httptest.NewRequest(http.MethodGet, "/lookahead/plan?token=bad", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad token status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLookaheadPlanReturnsTodayAndTomorrowReadOnly(t *testing.T) {
	h, userID := newTestLookaheadHandler(t)
	mux := http.NewServeMux()
	h.register(mux)
	token, err := h.tokens.Issue(userID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/lookahead/plan?token="+token, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp lookaheadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.User.DisplayName != "Look Ahead" {
		t.Fatalf("display_name = %q, want Look Ahead", resp.User.DisplayName)
	}
	if len(resp.Days) != 2 || resp.Days[0].Label != "Today" || resp.Days[1].Label != "Tomorrow" {
		t.Fatalf("unexpected days: %+v", resp.Days)
	}
	if len(resp.Days[0].Windows["morning"]) != 1 || resp.Days[0].Windows["morning"][0].Name != "Water plants" {
		t.Fatalf("unexpected morning items: %+v", resp.Days[0].Windows["morning"])
	}
}

func TestLookaheadIssueTokenCanBeDisabled(t *testing.T) {
	h, userID := newTestLookaheadHandler(t)
	h.allowTokenIssue = false
	mux := http.NewServeMux()
	h.register(mux)

	req := httptest.NewRequest(http.MethodGet, "/lookahead/token?user_id="+userID, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
