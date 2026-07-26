package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/rahat/rahat/internal/auth"
	calendarpkg "github.com/rahat/rahat/internal/calendar"
	googcalendar "github.com/rahat/rahat/internal/calendar/google"
	"github.com/rahat/rahat/internal/checkins"
	"github.com/rahat/rahat/internal/config"
	"github.com/rahat/rahat/internal/db"
	"github.com/rahat/rahat/internal/events"
	"github.com/rahat/rahat/internal/netutil"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	occ "github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/tokens"
	usr "github.com/rahat/rahat/internal/users"
	webhooktg "github.com/rahat/rahat/internal/webhooks/telegram"
)

func main() {
	cfg := config.Load()
	logOutput := os.Stdout
	if shouldLogCLIToStderr(os.Args) {
		logOutput = os.Stderr
	}
	logger := slog.New(slog.NewJSONHandler(logOutput, &slog.HandlerOptions{Level: cfg.LogLevel}))

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
	checkpointRepo := store.NewScheduleCheckpointRepository(sqlDB)
	calendarConnectionRepo := store.NewCalendarConnectionRepository(sqlDB)
	calendarBlockRepo := store.NewCalendarBlockRepository(sqlDB)
	oauthStateRepo := store.NewOAuthStateRepository(sqlDB)
	onboardingConfirmationRepo := store.NewOnboardingConfirmationRepository(sqlDB)
	schedulerService := scheduler.NewService(userService, taskService, occurrenceService, checkpointRepo, calendarBlockRepo)
	lookaheadSecret := os.Getenv("LOOKAHEAD_TOKEN_SECRET")
	if lookaheadSecret == "" && cfg.AppEnv == "development" {
		lookaheadSecret = "development-lookahead-secret"
	}
	lookaheadTokens := tokens.NewManager(lookaheadSecret)
	sessionSecret := os.Getenv("WEB_SESSION_SECRET")
	if sessionSecret == "" && cfg.AppEnv == "development" {
		sessionSecret = "development-web-session-secret"
	}
	authService := auth.NewService(sqlDB, auth.NewRepository(sqlDB), sessionSecret, 30*24*time.Hour)
	devOrigins := devOriginsFor(cfg.WebOrigin)
	authRoutes := &authHandler{auth: authService, users: userService, webOrigin: cfg.WebOrigin, appEnv: cfg.AppEnv, devOrigins: devOrigins}
	googleCalendarClient := googcalendar.NewClient(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_REDIRECT_URL"), os.Getenv("GOOGLE_CALENDAR_ID"))
	calendarService := calendarpkg.NewService(userService, calendarConnectionRepo, calendarBlockRepo, oauthStateRepo, googleCalendarClient)

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	webhookSecret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	botBaseURL := os.Getenv("TELEGRAM_API_BASE_URL")
	botUsername := os.Getenv("TELEGRAM_BOT_USERNAME")
	telegramAvailable := botToken != ""
	googleAvailable := os.Getenv("GOOGLE_CLIENT_ID") != "" && os.Getenv("GOOGLE_CLIENT_SECRET") != "" && os.Getenv("GOOGLE_REDIRECT_URL") != ""
	var telegramService *ntg.Service

	onboardingService := &onboardingHandler{
		sessions:          newOnboardingSessionStore(os.Getenv("ONBOARDING_INVITE_CODE")),
		users:             userService,
		prefs:             prefService,
		tasks:             taskService,
		scheduler:         schedulerService,
		auth:              authService,
		webAuth:           authRoutes,
		logger:            logger,
		telegramAvailable: telegramAvailable,
		botUsername:       botUsername,
		calendarService:   calendarService,
		googleAvailable:   googleAvailable,
	}
	onboardingService.register(mux)
	authRoutes.register(mux)
	(&taskManagementHandler{auth: authRoutes, tasks: taskService}).register(mux)
	(&lookaheadHandler{tokens: lookaheadTokens, users: userService, tasks: taskService, scheduler: schedulerService, allowTokenIssue: os.Getenv("LOOKAHEAD_TOKEN_ISSUER_ENABLED") == "true"}).register(mux)

	if botToken != "" {
		bot := ntg.NewHTTPBotClient(botToken, botBaseURL)
		onboardingService.bot = bot
		if onboardingService.botUsername == "" {
			if info, err := bot.GetMe(ctx); err == nil && info.Username != "" {
				onboardingService.botUsername = info.Username
			} else if err != nil {
				logger.Warn("telegram getMe failed", "error", err)
			}
		}
		telegramService = ntg.NewService(bot, userService, taskService, occurrenceService, eventService, onboardingConfirmationRepo)
		onboardingService.telegramService = telegramService
		checkinService := checkins.NewService(bot, userService, taskService, occurrenceService, eventService, prefService)
		editHandler := ntg.NewEditCommandHandler(authService, userService, bot, cfg.WebOrigin, logger)
		editHandler.LinkHost = os.Getenv("TELEGRAM_LINK_HOST")
		messageRouter := &telegramMessageRouter{onboarding: onboardingService, edit: editHandler}
		transport := configureTelegramTransport(ctx, logger, mux, bot, webhookSecret, webhookURL, checkinService, messageRouter)
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

	ops := newOpsRuntime(sqlDB, logger, cfg.AppEnv, cfg.DatabasePath, userService, taskService, prefService, schedulerService, telegramService, calendarService, eventService, lookaheadTokens)
	opsAuth := authRoutes
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ops:run-job":
			if len(os.Args) < 3 {
				logger.Error("missing job name", "available_jobs", ops.jobs.Names())
				os.Exit(1)
			}
			if err := ops.jobs.Run(ctx, os.Args[2]); err != nil {
				logger.Error("job failed", "job", os.Args[2], "error", err)
				os.Exit(1)
			}
			logger.Info("job complete", "job", os.Args[2])
			return
		case "ops:report-events":
			filter, err := parseReportFilter()
			if err != nil {
				logger.Error("invalid report filter", "error", err)
				os.Exit(1)
			}
			format := strings.ToLower(strings.TrimSpace(os.Getenv("REPORT_FORMAT")))
			if format == "csv" {
				if err := ops.exportEventsCSV(ctx, filter, os.Stdout); err != nil {
					logger.Error("event export failed", "error", err)
					os.Exit(1)
				}
				return
			}
			if err := ops.reportEventSummary(ctx, filter, os.Stdout); err != nil {
				logger.Error("event summary failed", "error", err)
				os.Exit(1)
			}
			return
		case "ops:backup":
			if err := ops.runBackup(ctx); err != nil {
				logger.Error("backup failed", "error", err)
				os.Exit(1)
			}
			logger.Info("backup complete", "target", ops.backupTarget)
			return
		case "ops:seed-testers":
			if err := ops.seedTesters(ctx); err != nil {
				logger.Error("seed testers failed", "error", err)
				os.Exit(1)
			}
			logger.Info("seed testers complete")
			return
		case "ops:issue-access-link":
			if len(os.Args) < 3 {
				logger.Error("missing user id or email")
				os.Exit(1)
			}
			result, err := opsAuth.issueAccessLink(ctx, os.Args[2], time.Hour)
			if err != nil {
				logger.Error("issue access link failed", "error", err)
				os.Exit(1)
			}
			payload, err := writeIssueLinkJSON(result)
			if err != nil {
				logger.Error("encode issue access link result failed", "error", err)
				os.Exit(1)
			}
			_, _ = os.Stdout.Write(append(payload, '\n'))
			return
		case "ops:reset-nonprod":
			if err := ops.resetNonProduction(ctx, os.Getenv("RAHAT_RESET_CONFIRM")); err != nil {
				logger.Error("reset failed", "error", err)
				os.Exit(1)
			}
			logger.Info("non-production reset complete", "database_path", cfg.DatabasePath)
			return
		}
	}

	mux.HandleFunc("GET /calendar/google/auth-url", requireAuthenticatedUserForRoute(authRoutes, func(w http.ResponseWriter, r *http.Request, current authenticatedUser) {
		if os.Getenv("GOOGLE_CLIENT_ID") == "" || os.Getenv("GOOGLE_CLIENT_SECRET") == "" || os.Getenv("GOOGLE_REDIRECT_URL") == "" {
			http.Error(w, "google oauth not configured", http.StatusServiceUnavailable)
			return
		}
		authURL, err := calendarService.GoogleAuthURL(r.Context(), current.User.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"auth_url": authURL})
	}))
	mux.HandleFunc("POST /calendar/google/connect", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")
		if state == "" || code == "" {
			http.Error(w, "missing state or code", http.StatusBadRequest)
			return
		}
		connection, err := calendarService.ConnectGoogle(r.Context(), state, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, connection)
	})
	mux.HandleFunc("POST /calendar/google/sync", requireAuthenticatedUserForRoute(authRoutes, func(w http.ResponseWriter, r *http.Request, current authenticatedUser) {
		if err := authRoutes.requireTrustedOrigin(r); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		day := parseDay(r.URL.Query().Get("date"))
		blocks, err := calendarService.SyncGoogleDay(r.Context(), current.User.ID, day)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"blocks": blocks})
	}))
	mux.HandleFunc("GET /schedule/plan", requireAuthenticatedUserForRoute(authRoutes, func(w http.ResponseWriter, r *http.Request, current authenticatedUser) {
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

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           withCORS(mux, cfg.WebOrigin, devOrigins, cfg.AppEnv),
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

func configureTelegramTransport(ctx context.Context, logger *slog.Logger, mux *http.ServeMux, bot ntg.RuntimeClient, webhookSecret, webhookURL string, callbackHandler ntg.CallbackHandler, messageHandler ntg.MessageHandler) string {
	if shouldUseTelegramWebhook(webhookURL, webhookSecret) {
		mux.Handle(telegramWebhookPath, webhooktg.NewHandler(webhookSecret, callbackHandler, messageHandler))
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
	go ntg.NewPoller(bot, callbackHandler, messageHandler, logger).Run(ctx)
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

func shouldLogCLIToStderr(args []string) bool {
	return len(args) > 1 && strings.HasPrefix(args[1], "ops:")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withCORS(next http.Handler, webOrigin string, devOrigins []string, appEnv string) http.Handler {
	allowed := []string{strings.TrimRight(webOrigin, "/")}
	if appEnv == "development" {
		for _, o := range devOrigins {
			allowed = append(allowed, strings.TrimRight(o, "/"))
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		for _, o := range allowed {
			if origin == o {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func devOriginsFor(webOrigin string) []string {
	parsed, err := url.Parse(webOrigin)
	if err != nil {
		return nil
	}
	host := parsed.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return nil
	}
	port := parsed.Port()
	if port == "" {
		return nil
	}
	var origins []string
	if ip := netutil.PrimaryLocalIPv4(); ip != "" {
		origins = append(origins, fmt.Sprintf("http://%s:%s", ip, port))
	}
	if host == "localhost" {
		origins = append(origins, fmt.Sprintf("http://127.0.0.1:%s", port))
	} else {
		origins = append(origins, fmt.Sprintf("http://localhost:%s", port))
	}
	return origins
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
