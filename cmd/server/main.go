package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/rahat/rahat/internal/app"
	"github.com/rahat/rahat/internal/checkins"
	"github.com/rahat/rahat/internal/config"
	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	occ "github.com/rahat/rahat/internal/occurrences"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
	webhooktg "github.com/rahat/rahat/internal/webhooks/telegram"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sqlDB, err := db.OpenSQLite(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if len(os.Args) > 1 && os.Args[1] == "db:setup" {
		logger.Info("database setup complete", "database_path", cfg.DatabasePath)
		return
	}

	baseHandler := app.NewServer(logger, cfg, sqlDB)
	mux := http.NewServeMux()
	mux.Handle("/", baseHandler)

	userService := usr.NewService(usr.NewRepository(sqlDB))
	taskService := taskpkg.NewService(taskpkg.NewRepository(sqlDB))
	occurrenceService := occ.NewService(occ.NewRepository(sqlDB))
	eventService := events.NewService(events.NewRepository(sqlDB))
	prefService := preferences.NewService(preferences.NewRepository(sqlDB))

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	webhookSecret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	botBaseURL := os.Getenv("TELEGRAM_API_BASE_URL")
	if botToken != "" {
		bot := ntg.NewHTTPBotClient(botToken, botBaseURL)
		telegramService := ntg.NewService(bot, userService, taskService, occurrenceService, eventService)
		checkinService := checkins.NewService(bot, userService, taskService, occurrenceService, eventService, prefService)
		transport := configureTelegramTransport(ctx, logger, mux, bot, webhookSecret, webhookURL, checkinService)
		logger.Info("telegram bot enabled", "transport", transport)

		mux.HandleFunc("POST /telegram/send/daily", func(w http.ResponseWriter, r *http.Request) {
			userID := r.URL.Query().Get("user_id")
			day := parseDay(r.URL.Query().Get("date"))
			if userID == "" {
				http.Error(w, "missing user_id", http.StatusBadRequest)
				return
			}
			if err := telegramService.SendMorningBatch(r.Context(), userID, day); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		})
		mux.HandleFunc("POST /telegram/send/window", func(w http.ResponseWriter, r *http.Request) {
			userID := r.URL.Query().Get("user_id")
			window := taskpkg.TimeOfDayPreference(strings.ToLower(r.URL.Query().Get("window")))
			day := parseDay(r.URL.Query().Get("date"))
			if userID == "" || window == "" {
				http.Error(w, "missing user_id or window", http.StatusBadRequest)
				return
			}
			if err := telegramService.SendWindowReminders(r.Context(), userID, day, window); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusAccepted)
		})
	} else {
		logger.Warn("telegram bot disabled: TELEGRAM_BOT_TOKEN not set")
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}()

	logger.Info("starting rahat api", "addr", cfg.HTTPAddr, "env", cfg.AppEnv, "database_path", cfg.DatabasePath)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

const telegramWebhookPath = "/webhooks/telegram"

func configureTelegramTransport(ctx context.Context, logger *slog.Logger, mux *http.ServeMux, bot ntg.RuntimeClient, webhookSecret, webhookURL string, handler ntg.CallbackHandler) string {
	if shouldUseTelegramWebhook(webhookURL, webhookSecret) {
		mux.Handle(telegramWebhookPath, webhooktg.NewHandler(webhookSecret, handler))
		if err := bot.SetWebhook(ctx, webhookURL, webhookSecret); err == nil {
			return "webhook"
		} else {
			logger.Warn("telegram webhook setup failed; falling back to long polling", "error", err, "webhook_url", webhookURL)
		}
	} else if webhookURL != "" {
		logger.Info("telegram webhook settings incomplete or non-domain-backed; using long polling", "webhook_url", webhookURL)
	}

	if err := bot.DeleteWebhook(ctx); err != nil {
		logger.Warn("telegram deleteWebhook failed before long polling", "error", err)
	}
	go ntg.NewPoller(bot, handler, logger).Run(ctx)
	return "long_polling"
}

func shouldUseTelegramWebhook(webhookURL, webhookSecret string) bool {
	if webhookURL == "" || webhookSecret == "" {
		return false
	}
	parsed, err := url.Parse(webhookURL)
	if err != nil || parsed.Scheme != "https" {
		return false
	}
	host := parsed.Hostname()
	if host == "" || strings.EqualFold(host, "localhost") {
		return false
	}
	if net.ParseIP(host) != nil {
		return false
	}
	if path.Clean(parsed.Path) != telegramWebhookPath {
		return false
	}
	return true
}

func parseDay(value string) time.Time {
	if value == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed.UTC()
}
