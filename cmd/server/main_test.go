package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/auth"
	calendarpkg "github.com/rahat/rahat/internal/calendar"
	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	"github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	occ "github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/tokens"
	usr "github.com/rahat/rahat/internal/users"
)

type fakeRuntimeClient struct {
	deleteWebhookCalls int
	setWebhookCalls    int
	setWebhookURL      string
	updatesCalled      chan struct{}
}

func (f *fakeRuntimeClient) SendMessage(context.Context, ntg.SendMessageRequest) error { return nil }
func (f *fakeRuntimeClient) SetWebhook(context.Context, string, string) error {
	f.setWebhookCalls++
	return nil
}
func (f *fakeRuntimeClient) DeleteWebhook(context.Context) error {
	f.deleteWebhookCalls++
	return nil
}
func (f *fakeRuntimeClient) GetMe(context.Context) (ntg.BotInfo, error) { return ntg.BotInfo{}, nil }
func (f *fakeRuntimeClient) GetUpdates(ctx context.Context, req ntg.GetUpdatesRequest) ([]ntg.Update, error) {
	if f.updatesCalled != nil {
		select {
		case <-f.updatesCalled:
		default:
			close(f.updatesCalled)
		}
	}
	<-ctx.Done()
	return nil, ctx.Err()
}

type fakeCallbackHandler struct{}

func (fakeCallbackHandler) HandleCallback(context.Context, string) error { return nil }

func TestConfigureTelegramTransportFallsBackToLongPollingWithoutWebhook(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &fakeRuntimeClient{updatesCalled: make(chan struct{})}
	mux := http.NewServeMux()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	transport := configureTelegramTransport(ctx, logger, mux, client, "", "", fakeCallbackHandler{}, nil)
	if transport != "long_polling" {
		t.Fatalf("transport = %s, want long_polling", transport)
	}
	if client.deleteWebhookCalls != 1 {
		t.Fatalf("deleteWebhook calls = %d, want 1", client.deleteWebhookCalls)
	}

	select {
	case <-client.updatesCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("expected long polling to start getUpdates")
	}
}

func TestOpsCommandsDoNotConfigureTelegramTransport(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "rahat.sqlite3")
	sqlDB, err := db.OpenSQLite(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	defer sqlDB.Close()

	if _, err := sqlDB.ExecContext(ctx,
		`INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, created_at, updated_at)
		 VALUES ('u1', 'Tester', 'UTC', 45, 'chat-123', '2026-07-25T12:00:00Z', '2026-07-25T12:00:00Z')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userService := usr.NewService(usr.NewRepository(sqlDB))
	taskService := tasks.NewService(tasks.NewRepository(sqlDB))
	occurrenceService := occ.NewService(occ.NewRepository(sqlDB))
	eventService := events.NewService(events.NewRepository(sqlDB))
	prefService := preferences.NewService(preferences.NewRepository(sqlDB))
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	calendarBlockRepo := store.NewCalendarBlockRepository(sqlDB)
	schedulerService := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, calendarBlockRepo)
	calendarConnectionRepo := store.NewCalendarConnectionRepository(sqlDB)
	oauthStateRepo := store.NewOAuthStateRepository(sqlDB)
	onboardingConfirmationRepo := store.NewOnboardingConfirmationRepository(sqlDB)
	calendarService := calendarpkg.NewService(userService, calendarConnectionRepo, calendarBlockRepo, oauthStateRepo, nil)
	tokenMgr := tokens.NewManager("test-secret")

	fakeBot := &fakeRuntimeClient{}
	telegramService := ntg.NewService(fakeBot, userService, taskService, occurrenceService, eventService, onboardingConfirmationRepo)
	ops := newOpsRuntime(sqlDB, logger, "development", databasePath, userService, taskService, prefService, schedulerService, telegramService, calendarService, eventService, tokenMgr)
	authRoutes := &authHandler{auth: auth.NewService(sqlDB, auth.NewRepository(sqlDB), "test-session-secret", 30*24*time.Hour), users: userService, webOrigin: "http://localhost:5200", appEnv: "development"}

	if err := runOpsCommand(ctx, ops, authRoutes, []string{"server", "ops:run-job", "telegram-daily"}); err != nil {
		t.Fatalf("ops:run-job telegram-daily error = %v", err)
	}
	if err := runOpsCommand(ctx, ops, authRoutes, []string{"server", "ops:report-events"}); err != nil {
		t.Fatalf("ops:report-events error = %v", err)
	}

	if fakeBot.deleteWebhookCalls != 0 || fakeBot.setWebhookCalls != 0 {
		t.Fatalf("operator commands triggered telegram transport setup: deleteWebhook=%d setWebhook=%d", fakeBot.deleteWebhookCalls, fakeBot.setWebhookCalls)
	}
}

func TestShouldLogCLIToStderr(t *testing.T) {
	if !shouldLogCLIToStderr([]string{"server", "ops:issue-access-link"}) {
		t.Fatal("expected ops command logs to use stderr")
	}
	if shouldLogCLIToStderr([]string{"server"}) {
		t.Fatal("expected normal server logs to use stdout")
	}
}

func TestShouldUseTelegramWebhook(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		secret string
		want   bool
	}{
		{name: "domain backed https", url: "https://rahat.example.com/webhooks/telegram", secret: "secret", want: true},
		{name: "missing secret", url: "https://rahat.example.com/webhooks/telegram", want: false},
		{name: "localhost", url: "https://localhost/webhooks/telegram", secret: "secret", want: false},
		{name: "ip host", url: "https://127.0.0.1/webhooks/telegram", secret: "secret", want: false},
		{name: "http only", url: "http://rahat.example.com/webhooks/telegram", secret: "secret", want: false},
		{name: "wrong path", url: "https://rahat.example.com/telegram-hook", secret: "secret", want: false},
		{name: "empty path", url: "https://rahat.example.com", secret: "secret", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldUseTelegramWebhook(tc.url, tc.secret); got != tc.want {
				t.Fatalf("shouldUseTelegramWebhook(%q, %q) = %v, want %v", tc.url, tc.secret, got, tc.want)
			}
		})
	}
}
