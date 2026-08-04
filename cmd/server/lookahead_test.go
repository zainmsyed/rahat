package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
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

func newTestLookaheadHandler(t *testing.T) (*lookaheadHandler, string, *sql.DB) {
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
	return &lookaheadHandler{tokens: mgr, users: userService, tasks: taskService, scheduler: schedulerService, allowTokenIssue: true, clock: func() time.Time { return time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC) }}, user.ID, sqlDB
}

func TestLookaheadPlanRequiresValidToken(t *testing.T) {
	h, _, _ := newTestLookaheadHandler(t)
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
	h, userID, _ := newTestLookaheadHandler(t)
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
	if resp.RangeDays != 2 || len(resp.Days) != 2 || resp.Days[0].Label != "Today" || resp.Days[1].Label != "Tomorrow" {
		t.Fatalf("unexpected range: %+v", resp)
	}
	if resp.Days[0].Overflowed == nil || resp.Days[0].Skipped == nil || resp.Days[0].Reasons == nil {
		t.Fatalf("missing explicit preview fields: %+v", resp.Days[0])
	}
	if len(resp.Days[0].Windows["morning"]) != 1 || resp.Days[0].Windows["morning"][0].Name != "Water plants" {
		t.Fatalf("unexpected morning items: %+v", resp.Days[0].Windows["morning"])
	}
}

func TestLookaheadWeeklyPreviewCarriesStateAndDoesNotPersist(t *testing.T) {
	h, userID, sqlDB := newTestLookaheadHandler(t)
	h.clock = func() time.Time { return time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC) }
	ctx := context.Background()
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	if _, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{
		UserID: userID, Name: "Weekday count", DurationMinutes: 1,
		CadenceType: tasks.CadenceTypeCount, CadenceValue: 2,
		Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning,
		DayPreference: tasks.DayPreferenceWeekday,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{
		UserID: userID, Name: "Weekend count", DurationMinutes: 1,
		CadenceType: tasks.CadenceTypeCount, CadenceValue: 2,
		Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayMorning,
		DayPreference: tasks.DayPreferenceWeekend,
	}, nil); err != nil {
		t.Fatal(err)
	}

	var beforeOccurrences, beforeCheckpoints int
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM occurrences`).Scan(&beforeOccurrences); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedule_checkpoints`).Scan(&beforeCheckpoints); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	h.register(mux)
	token, err := h.tokens.Issue(userID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	request := func() lookaheadResponse {
		req := httptest.NewRequest(http.MethodGet, "/lookahead/plan?token="+token+"&days=7", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
		}
		var response lookaheadResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response
	}

	first := request()
	second := request()
	if first.RangeDays != 7 || len(first.Days) != 7 || len(second.Days) != 7 {
		t.Fatalf("weekly range shape = %d/%d/%d, want 7/7/7", first.RangeDays, len(first.Days), len(second.Days))
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated preview returned a different shape: first=%+v second=%+v", first, second)
	}
	weekdayCount := 0
	weekendCount := 0
	weekdayDates := []string{}
	weekendDates := []string{}
	for _, day := range first.Days {
		for _, items := range day.Windows {
			for _, item := range items {
				switch item.Name {
				case "Weekday count":
					weekdayCount++
					weekdayDates = append(weekdayDates, day.Date)
				case "Weekend count":
					weekendCount++
					weekendDates = append(weekendDates, day.Date)
				}
			}
		}
	}
	if weekdayCount != 2 || weekendCount != 2 {
		t.Fatalf("weekly count tasks scheduled weekday=%d (%v) weekend=%d (%v), want 2/2", weekdayCount, weekdayDates, weekendCount, weekendDates)
	}
	if first.Days[5].Label != "Saturday" || first.Days[6].Label != "Sunday" {
		t.Fatalf("weekly labels = %q/%q, want Saturday/Sunday", first.Days[5].Label, first.Days[6].Label)
	}

	var afterOccurrences, afterCheckpoints int
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM occurrences`).Scan(&afterOccurrences); err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schedule_checkpoints`).Scan(&afterCheckpoints); err != nil {
		t.Fatal(err)
	}
	if beforeOccurrences != afterOccurrences || beforeCheckpoints != afterCheckpoints {
		t.Fatalf("preview persisted state: occurrences %d->%d checkpoints %d->%d", beforeOccurrences, afterOccurrences, beforeCheckpoints, afterCheckpoints)
	}
	if got := len(second.Days[0].Windows["morning"]) + len(second.Days[1].Windows["morning"]); got == 0 {
		t.Fatal("repeated preview unexpectedly returned an empty shape")
	}
}

func TestLookaheadWeeklyPreviewRejectsUnsupportedRange(t *testing.T) {
	h, userID, _ := newTestLookaheadHandler(t)
	mux := http.NewServeMux()
	h.register(mux)
	token, err := h.tokens.Issue(userID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/lookahead/plan?token="+token+"&days=8", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLookaheadPlanIncludesOmittedItemReason(t *testing.T) {
	h, userID, sqlDB := newTestLookaheadHandler(t)
	ctx := context.Background()
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	if _, err := taskService.CreateTaskWithSubtasks(ctx, tasks.Task{
		UserID: userID, Name: "Afternoon blocked task", DurationMinutes: 30,
		CadenceType: tasks.CadenceTypeInterval, CadenceValue: 1,
		Priority: tasks.PriorityMedium, TimeOfDayPreference: tasks.TimeOfDayAfternoon,
	}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlDB.ExecContext(ctx,
		`INSERT INTO calendar_blocks (id, user_id, provider, external_event_id, local_date, timezone, title, detail, start_at, end_at, is_all_day, classification, window, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"block-039-afternoon", userID, "google", "ext-039-afternoon", "2026-07-24", "UTC", "Dentist appointment", "", nil, nil, false, "medium", "afternoon", time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}

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
		t.Fatal(err)
	}
	if len(resp.Days) == 0 {
		t.Fatal("expected at least one day")
	}
	found := false
	for _, item := range resp.Days[0].OmittedItems {
		if item.Name == "Afternoon blocked task" && strings.Contains(item.Reason, "Dentist appointment") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("omitted item reason missing expected blocked-window detail: %+v", resp.Days[0].OmittedItems)
	}
}

func TestLookaheadPlanRespectsDayRangeBoundaries(t *testing.T) {
	h, userID, _ := newTestLookaheadHandler(t)
	mux := http.NewServeMux()
	h.register(mux)
	token, err := h.tokens.Issue(userID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/lookahead/plan?token="+token+"&days=1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("days=1 status = %d, want %d", rec.Code, http.StatusOK)
	}
	var oneDay lookaheadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &oneDay); err != nil {
		t.Fatal(err)
	}
	if oneDay.RangeDays != 1 || len(oneDay.Days) != 1 {
		t.Fatalf("days=1 response = %d/%d, want 1/1", oneDay.RangeDays, len(oneDay.Days))
	}

	for _, invalid := range []string{"0", "-1", "-7"} {
		req := httptest.NewRequest(http.MethodGet, "/lookahead/plan?token="+token+"&days="+invalid, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("days=%s status = %d, want %d", invalid, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestLookaheadIssueTokenCanBeDisabled(t *testing.T) {
	h, userID, _ := newTestLookaheadHandler(t)
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
