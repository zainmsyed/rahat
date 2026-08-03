package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/auth"
	calendarpkg "github.com/rahat/rahat/internal/calendar"
	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	occ "github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
)

func TestValidateTaskRequestDayPreference(t *testing.T) {
	base := onboardingTaskRequest{
		Name: "Weekend reset", DurationMinutes: 20, CadenceType: taskpkg.CadenceTypeCount, CadenceValue: 2,
		Priority: taskpkg.PriorityMedium, TimeOfDayPreference: taskpkg.TimeOfDayAny, DayPreference: taskpkg.DayPreferenceWeekend,
	}
	if task, _, err := validateTaskRequest("user-1", base); err != nil || task.DayPreference != taskpkg.DayPreferenceWeekend {
		t.Fatalf("valid weekend task = %+v, err = %v", task, err)
	}
	base.CadenceType = taskpkg.CadenceTypeInterval
	if _, _, err := validateTaskRequest("user-1", base); err == nil {
		t.Fatal("expected weekend interval cadence to be rejected")
	}
	base.CadenceType = taskpkg.CadenceTypeCount
	base.CadenceValue = 3
	if _, _, err := validateTaskRequest("user-1", base); err == nil {
		t.Fatal("expected weekend cadence above two to be rejected")
	}
}

type fakeCalendarOAuthClient struct {
	authURL string
	token   calendarpkg.OAuthToken
}

func (f *fakeCalendarOAuthClient) AuthCodeURL(state string) string { return f.authURL + state }
func (f *fakeCalendarOAuthClient) ExchangeCode(context.Context, string) (calendarpkg.OAuthToken, error) {
	return f.token, nil
}
func (f *fakeCalendarOAuthClient) ListEvents(context.Context, store.CalendarConnection, time.Time, *time.Location) ([]calendarpkg.Event, error) {
	return nil, nil
}

func newTestOnboardingHandler(t *testing.T) *onboardingHandler {
	t.Helper()
	return newTestOnboardingHandlerWithBot(t, &fakeTelegramBot{})
}

func newTestOnboardingHandlerWithBot(t *testing.T, bot ntg.BotClient) *onboardingHandler {
	t.Helper()

	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	userService := usr.NewService(usr.NewRepository(sqlDB))
	taskService := taskpkg.NewService(taskpkg.NewRepository(sqlDB))
	occurrenceService := occ.NewService(occ.NewRepository(sqlDB))
	eventService := events.NewService(events.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	calendarBlockRepo := store.NewCalendarBlockRepository(sqlDB)
	calendarConnectionRepo := store.NewCalendarConnectionRepository(sqlDB)
	oauthStateRepo := store.NewOAuthStateRepository(sqlDB)
	onboardingConfirmationRepo := store.NewOnboardingConfirmationRepository(sqlDB)
	schedulerService := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, calendarBlockRepo)
	calendarService := calendarpkg.NewService(userService, calendarConnectionRepo, calendarBlockRepo, oauthStateRepo, &fakeCalendarOAuthClient{authURL: "https://accounts.google.test/oauth?state="})
	authService := auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-web-session-secret", 30*24*time.Hour)
	webAuth := &authHandler{auth: authService, users: userService, webOrigin: "http://localhost:5200", appEnv: "development"}
	telegramService := ntg.NewService(bot, userService, taskService, occurrenceService, eventService, onboardingConfirmationRepo)

	return &onboardingHandler{
		sessions:          newOnboardingSessionStore("rahat-beta"),
		users:             userService,
		prefs:             preferences.NewService(preferences.NewRepository(sqlDB)),
		tasks:             taskService,
		scheduler:         schedulerService,
		auth:              authService,
		webAuth:           webAuth,
		telegramService:   telegramService,
		telegramAvailable: true,
		botUsername:       "RahatTestBot",
		calendarService:   calendarService,
		googleAvailable:   false,
	}
}

func TestOnboardingSessionStoreCreateAndGet(t *testing.T) {
	store := newOnboardingSessionStore("rahat-beta")

	_, err := store.Create("wrong-code")
	if err == nil {
		t.Fatal("expected error for invalid invite code")
	}

	session, err := store.Create("rahat-beta")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if session.Token == "" {
		t.Fatal("expected non-empty token")
	}

	got, err := store.Get(session.Token)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Token != session.Token {
		t.Fatalf("token mismatch: got %q, want %q", got.Token, session.Token)
	}

	if err := store.AttachUser(session.Token, "user-123"); err != nil {
		t.Fatalf("AttachUser() error = %v", err)
	}
	got, _ = store.Get(session.Token)
	if got.UserID != "user-123" {
		t.Fatalf("UserID = %q, want user-123", got.UserID)
	}
}

func TestOnboardingCreateSessionEndpoint(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	body := bytes.NewReader([]byte(`{"invite_code":"rahat-beta"}`))
	req := httptest.NewRequest(http.MethodPost, "/onboarding/session", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var resp onboardingCreateSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected token in response")
	}
}

func TestOnboardingCreateSessionEndpointRejectsInvalidCode(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	body := bytes.NewReader([]byte(`{"invite_code":"invalid"}`))
	req := httptest.NewRequest(http.MethodPost, "/onboarding/session", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestOnboardingSaveProfileEndpoint(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")

	payload := onboardingProfileRequest{
		DisplayName:            "Test User",
		Timezone:               "America/Chicago",
		DailyTimeBudgetMinutes: 45,
		Email:                  "",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp onboardingUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.DisplayName != "Test User" {
		t.Fatalf("display_name = %q, want Test User", resp.DisplayName)
	}

	updated, _ := h.sessions.Get(session.Token)
	if updated.UserID == "" {
		t.Fatal("expected session to be attached to user")
	}
}

func TestOnboardingTaskEndpointsRequireProfileFirst(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")

	req := httptest.NewRequest(http.MethodPost, "/onboarding/tasks?token="+session.Token, bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestOnboardingCreateAndUpdateTask(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	task := onboardingTaskRequest{
		Name:                "Custom task",
		DurationMinutes:     20,
		CadenceType:         taskpkg.CadenceTypeInterval,
		CadenceValue:        1,
		Priority:            taskpkg.PriorityMedium,
		TimeOfDayPreference: taskpkg.TimeOfDayMorning,
	}
	body, _ = json.Marshal(task)
	req = httptest.NewRequest(http.MethodPost, "/onboarding/tasks?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created onboardingTaskResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created task: %v", err)
	}
	if created.Name != "Custom task" {
		t.Fatalf("task name = %q, want Custom task", created.Name)
	}

	updatedPayload := task
	updatedPayload.Name = "Updated task"
	body, _ = json.Marshal(updatedPayload)
	req = httptest.NewRequest(http.MethodPut, "/onboarding/tasks/"+created.ID+"?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("update task status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestOnboardingFinishEndpoint(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 120}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	templates, _ := h.tasks.ListStarterTaskTemplates(context.Background())
	if len(templates) == 0 {
		t.Fatal("no starter templates found")
	}
	body, _ = json.Marshal(onboardingCreateStarterTaskRequest{TemplateID: templates[0].ID})
	req = httptest.NewRequest(http.MethodPost, "/onboarding/tasks/from-template?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("starter task status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/finish?token="+session.Token, http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("finish status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if cookie := rec.Header().Get("Set-Cookie"); !strings.Contains(cookie, sessionCookieName+"=") {
		t.Fatalf("expected session cookie, got %q", cookie)
	}
	var result onboardingFinishResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode finish response: %v", err)
	}
	if result.TaskCount == 0 {
		t.Fatal("expected at least one task in finish result")
	}
	if result.TelegramDelivered {
		t.Fatal("expected telegram_delivered=false when telegram not linked")
	}
}

func TestOnboardingFinishRequiresTasks(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodPost, "/onboarding/finish?token="+session.Token, http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOnboardingInvalidTokenReturnsUnauthorized(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	req := httptest.NewRequest(http.MethodGet, "/onboarding/state?token=invalid", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestOnboardingStateEndpoint(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	req := httptest.NewRequest(http.MethodGet, "/onboarding/state?token="+session.Token, http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var state onboardingStateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if state.HasProfile {
		t.Fatal("expected HasProfile to be false")
	}
}

func TestOnboardingTelegramStatusAvailable(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/onboarding/telegram?token="+session.Token, http.NoBody)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var status onboardingTelegramStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode telegram status: %v", err)
	}
	if !status.Available {
		t.Fatal("expected telegram available")
	}
	if status.BotUsername != "RahatTestBot" {
		t.Fatalf("bot_username = %q, want RahatTestBot", status.BotUsername)
	}
	if status.Code == "" {
		t.Fatal("expected non-empty code")
	}
	if status.DeepLink == "" {
		t.Fatal("expected non-empty deep_link")
	}
	if status.Linked {
		t.Fatal("expected not linked yet")
	}
}

func TestOnboardingTelegramStatusUnavailable(t *testing.T) {
	h := newTestOnboardingHandler(t)
	h.telegramAvailable = false
	h.botUsername = ""
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	req := httptest.NewRequest(http.MethodGet, "/onboarding/telegram?token="+session.Token, http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var status onboardingTelegramStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode telegram status: %v", err)
	}
	if status.Available {
		t.Fatal("expected telegram unavailable")
	}
}

type fakeTelegramBot struct {
	messages []ntg.SendMessageRequest
	err      error
}

func (f *fakeTelegramBot) SendMessage(_ context.Context, req ntg.SendMessageRequest) error {
	if f.err != nil {
		return f.err
	}
	f.messages = append(f.messages, req)
	return nil
}

func TestOnboardingTelegramMessageLinksUser(t *testing.T) {
	h := newTestOnboardingHandler(t)
	bot := &fakeTelegramBot{}
	h.bot = bot
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	code, _ := h.sessions.SetTelegramCode(session.Token)
	msg := &ntg.Message{Text: "/start " + code, Chat: &ntg.Chat{ID: 123, Type: "private"}}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}

	updated, _ := h.sessions.Get(session.Token)
	if !updated.TelegramLinked {
		t.Fatal("expected session telegram linked")
	}
	if updated.TelegramChatID != "123" {
		t.Fatalf("telegram_chat_id = %q, want 123", updated.TelegramChatID)
	}

	user, _ := h.users.GetByID(context.Background(), updated.UserID)
	if user.TelegramChatID != "123" {
		t.Fatalf("user telegram_chat_id = %q, want 123", user.TelegramChatID)
	}

	prefs, _ := h.prefs.ListByUser(context.Background(), updated.UserID)
	found := false
	for _, pref := range prefs {
		if pref.Channel == preferences.ChannelTelegram && pref.Enabled && pref.IsPrimary && pref.SupportsInteractive {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected telegram preference to be upserted, got %+v", prefs)
	}

	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 welcome message, got %d", len(bot.messages))
	}
	if !strings.Contains(bot.messages[0].Text, "Welcome to Rahat") {
		t.Fatalf("unexpected welcome message: %q", bot.messages[0].Text)
	}
}

func TestExtractOnboardingCode(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"plain code", "ABC123", "ABC123"},
		{"start command", "/start ABC123", "ABC123"},
		{"start with bot username", "/start@RahatBot ABC123", "ABC123"},
		{"lowercase code", "abc123", "ABC123"},
		{"code with whitespace", "  ABC123  ", "ABC123"},
		{"short code", "ABC12", ""},
		{"long code", "ABC1234", ""},
		{"special characters", "ABC-12", ""},
		{"empty", "", ""},
		{"start without code", "/start", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractOnboardingCode(tc.text); got != tc.want {
				t.Fatalf("extractOnboardingCode(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}

func TestOnboardingTelegramMessageRejectsGroupChat(t *testing.T) {
	h := newTestOnboardingHandler(t)
	bot := &fakeTelegramBot{}
	h.bot = bot
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	code, _ := h.sessions.SetTelegramCode(session.Token)
	msg := &ntg.Message{Text: "/start " + code, Chat: &ntg.Chat{ID: -123, Type: "group"}}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}

	updated, _ := h.sessions.Get(session.Token)
	if updated.TelegramLinked {
		t.Fatal("expected group chat message to be ignored")
	}
}

func TestOnboardingTelegramSkip(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60, Email: "test@example.com"}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/telegram/skip?token="+session.Token, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("skip status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode skip response: %v", err)
	}
	if !resp["skipped"] {
		t.Fatal("expected skipped=true")
	}

	updated, _ := h.sessions.Get(session.Token)
	prefs, _ := h.prefs.ListByUser(context.Background(), updated.UserID)
	found := false
	for _, pref := range prefs {
		if pref.Channel == preferences.ChannelEmail && pref.Enabled && pref.IsPrimary {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected email preference to be upserted, got %+v", prefs)
	}
}

func TestOnboardingCalendarStatusUnavailable(t *testing.T) {
	h := newTestOnboardingHandler(t)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/onboarding/calendar/status?token="+session.Token, http.NoBody)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var status onboardingCalendarStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode calendar status: %v", err)
	}
	if status.Available {
		t.Fatal("expected calendar unavailable")
	}
	if status.Connected {
		t.Fatal("expected not connected")
	}
}

func TestOnboardingCalendarStatusAvailable(t *testing.T) {
	h := newTestOnboardingHandler(t)
	h.googleAvailable = true
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/onboarding/calendar/status?token="+session.Token, http.NoBody)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var status onboardingCalendarStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode calendar status: %v", err)
	}
	if !status.Available {
		t.Fatal("expected calendar available")
	}
	if status.Connected {
		t.Fatal("expected not connected")
	}
	if status.AuthURL == "" {
		t.Fatal("expected non-empty auth_url")
	}
}

func TestOnboardingCalendarDisconnect(t *testing.T) {
	h := newTestOnboardingHandler(t)
	h.googleAvailable = true
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}
	updated, _ := h.sessions.Get(session.Token)

	authURL, err := h.calendarService.GoogleAuthURL(context.Background(), updated.UserID)
	if err != nil {
		t.Fatalf("auth url error = %v", err)
	}
	state := authURL[len("https://accounts.google.test/oauth?state="):]
	if _, err := h.calendarService.ConnectGoogle(context.Background(), state, "code-123"); err != nil {
		t.Fatalf("connect error = %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/calendar/disconnect?token="+session.Token, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("disconnect status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode disconnect response: %v", err)
	}
	if !resp["disconnected"] {
		t.Fatal("expected disconnected=true")
	}

	stateReq := httptest.NewRequest(http.MethodGet, "/onboarding/state?token="+session.Token, http.NoBody)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, stateReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("state status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var gotState onboardingStateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &gotState); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if gotState.CalendarConnected {
		t.Fatal("expected calendar_connected false after disconnect")
	}
}

func TestOnboardingStateCalendarConnectedTrue(t *testing.T) {
	h := newTestOnboardingHandler(t)
	h.googleAvailable = true
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}
	updated, _ := h.sessions.Get(session.Token)

	authURL, err := h.calendarService.GoogleAuthURL(context.Background(), updated.UserID)
	if err != nil {
		t.Fatalf("auth url error = %v", err)
	}
	state := authURL[len("https://accounts.google.test/oauth?state="):]
	if _, err := h.calendarService.ConnectGoogle(context.Background(), state, "code-123"); err != nil {
		t.Fatalf("connect error = %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/onboarding/state?token="+session.Token, http.NoBody)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("state status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var gotState onboardingStateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &gotState); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if !gotState.CalendarConnected {
		t.Fatal("expected calendar_connected true")
	}
}

func TestOnboardingCalendarDisconnectUnavailable(t *testing.T) {
	h := newTestOnboardingHandler(t)
	h.googleAvailable = false
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 60}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/calendar/disconnect?token="+session.Token, bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func BenchmarkDecodeJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		body := io.NopCloser(bytes.NewReader([]byte(`{"invite_code":"rahat-beta"}`)))
		r := httptest.NewRequest(http.MethodPost, "/onboarding/session", body)
		var req onboardingCreateSessionRequest
		_ = decodeJSON(r, &req)
	}
}

func TestOnboardingFinishSendsTelegramConfirmation(t *testing.T) {
	bot := &fakeTelegramBot{}
	h := newTestOnboardingHandlerWithBot(t, bot)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 120}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	templates, _ := h.tasks.ListStarterTaskTemplates(context.Background())
	body, _ = json.Marshal(onboardingCreateStarterTaskRequest{TemplateID: templates[0].ID})
	req = httptest.NewRequest(http.MethodPost, "/onboarding/tasks/from-template?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("starter task status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	code, _ := h.sessions.SetTelegramCode(session.Token)
	msg := &ntg.Message{Text: "/start " + code, Chat: &ntg.Chat{ID: 123, Type: "private"}}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/finish?token="+session.Token, http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("finish status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var result onboardingFinishResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode finish response: %v", err)
	}
	if !result.TelegramDelivered {
		t.Fatalf("expected telegram_delivered=true, got %+v", result)
	}
	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 telegram message, got %d", len(bot.messages))
	}
}

func TestOnboardingFinishDuplicateDoesNotResend(t *testing.T) {
	bot := &fakeTelegramBot{}
	h := newTestOnboardingHandlerWithBot(t, bot)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 120}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	templates, _ := h.tasks.ListStarterTaskTemplates(context.Background())
	body, _ = json.Marshal(onboardingCreateStarterTaskRequest{TemplateID: templates[0].ID})
	req = httptest.NewRequest(http.MethodPost, "/onboarding/tasks/from-template?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("starter task status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	code, _ := h.sessions.SetTelegramCode(session.Token)
	msg := &ntg.Message{Text: "/start " + code, Chat: &ntg.Chat{ID: 123, Type: "private"}}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}

	finish := func() onboardingFinishResponse {
		req := httptest.NewRequest(http.MethodPost, "/onboarding/finish?token="+session.Token, http.NoBody)
		req.Header.Set("Origin", "http://localhost:5200")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("finish status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var result onboardingFinishResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode finish response: %v", err)
		}
		return result
	}

	first := finish()
	second := finish()
	if !first.TelegramDelivered || !second.TelegramDelivered {
		t.Fatalf("expected telegram_delivered=true for both finishes, got first=%v second=%v", first.TelegramDelivered, second.TelegramDelivered)
	}
	if len(bot.messages) != 1 {
		t.Fatalf("expected 1 telegram message total, got %d", len(bot.messages))
	}
}

func TestOnboardingFinishWithoutTelegramLinked(t *testing.T) {
	bot := &fakeTelegramBot{}
	h := newTestOnboardingHandlerWithBot(t, bot)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 120}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	templates, _ := h.tasks.ListStarterTaskTemplates(context.Background())
	body, _ = json.Marshal(onboardingCreateStarterTaskRequest{TemplateID: templates[0].ID})
	req = httptest.NewRequest(http.MethodPost, "/onboarding/tasks/from-template?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("starter task status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/finish?token="+session.Token, http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("finish status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var result onboardingFinishResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode finish response: %v", err)
	}
	if result.TelegramDelivered {
		t.Fatal("expected telegram_delivered=false when telegram not linked")
	}
	if len(bot.messages) != 0 {
		t.Fatalf("expected no telegram message, got %d", len(bot.messages))
	}
}

func TestOnboardingFinishTelegramFailureStillSucceeds(t *testing.T) {
	bot := &fakeTelegramBot{err: fmt.Errorf("telegram API returned 400")}
	h := newTestOnboardingHandlerWithBot(t, bot)
	mux := http.NewServeMux()
	h.register(mux)

	session, _ := h.sessions.Create("rahat-beta")
	profile := onboardingProfileRequest{DisplayName: "Tester", Timezone: "UTC", DailyTimeBudgetMinutes: 120}
	body, _ := json.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/profile?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile save status = %d, want %d", rec.Code, http.StatusOK)
	}

	templates, _ := h.tasks.ListStarterTaskTemplates(context.Background())
	body, _ = json.Marshal(onboardingCreateStarterTaskRequest{TemplateID: templates[0].ID})
	req = httptest.NewRequest(http.MethodPost, "/onboarding/tasks/from-template?token="+session.Token, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("starter task status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	code, _ := h.sessions.SetTelegramCode(session.Token)
	msg := &ntg.Message{Text: "/start " + code, Chat: &ntg.Chat{ID: 123, Type: "private"}}
	if err := h.HandleMessage(context.Background(), msg); err != nil {
		t.Fatalf("HandleMessage error = %v", err)
	}

	req = httptest.NewRequest(http.MethodPost, "/onboarding/finish?token="+session.Token, http.NoBody)
	req.Header.Set("Origin", "http://localhost:5200")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("finish status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if cookie := rec.Header().Get("Set-Cookie"); !strings.Contains(cookie, sessionCookieName+"=") {
		t.Fatalf("expected session cookie even when telegram fails, got %q", cookie)
	}

	var result onboardingFinishResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode finish response: %v", err)
	}
	if result.TelegramDelivered {
		t.Fatal("expected telegram_delivered=false when send fails")
	}
	if len(bot.messages) != 0 {
		t.Fatalf("expected no successful telegram message, got %d", len(bot.messages))
	}
}
