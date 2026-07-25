package main

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	calendarpkg "github.com/rahat/rahat/internal/calendar"
	"github.com/rahat/rahat/internal/events"
	"github.com/rahat/rahat/internal/jobs"
	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/tokens"
	usr "github.com/rahat/rahat/internal/users"
)

const nonProductionResetConfirm = "reset-non-production"

type opsRuntime struct {
	db           *sql.DB
	databasePath string
	logger       *slog.Logger
	users        *usr.Service
	tasks        *taskpkg.Service
	prefs        *preferences.Service
	scheduler    *scheduler.Service
	telegram     *ntg.Service
	calendar     *calendarpkg.Service
	events       *events.Service
	tokens       *tokens.Manager
	jobs         *jobs.Service
	outboxDir    string
	backupTarget string
	webOrigin    string
	appEnv       string
	now          func() time.Time
}

type opsUser struct {
	ID             string
	DisplayName    string
	Timezone       string
	Email          string
	TelegramChatID string
}

func newOpsRuntime(sqlDB *sql.DB, logger *slog.Logger, appEnv, databasePath string, usersSvc *usr.Service, taskSvc *taskpkg.Service, prefSvc *preferences.Service, schedulerSvc *scheduler.Service, telegramSvc *ntg.Service, calendarSvc *calendarpkg.Service, eventSvc *events.Service, tokenMgr *tokens.Manager) *opsRuntime {
	r := &opsRuntime{
		db:           sqlDB,
		databasePath: databasePath,
		logger:       logger,
		users:        usersSvc,
		tasks:        taskSvc,
		prefs:        prefSvc,
		scheduler:    schedulerSvc,
		telegram:     telegramSvc,
		calendar:     calendarSvc,
		events:       eventSvc,
		tokens:       tokenMgr,
		outboxDir:    envOrDefault("EMAIL_RECAP_OUTBOX_DIR", "./var/email-outbox"),
		backupTarget: envOrDefault("BACKUP_TARGET_URI", "./var/backups"),
		webOrigin:    strings.TrimRight(envOrDefault("WEB_ORIGIN", "http://localhost:5200"), "/"),
		appEnv:       appEnv,
		now:          func() time.Time { return time.Now().UTC() },
	}
	r.jobs = jobs.NewService([]jobs.Job{
		{Name: "schedule-daily", Run: func(ctx context.Context) error { return r.runScheduleGeneration(ctx) }},
		{Name: "telegram-daily", Run: func(ctx context.Context) error { return r.runTelegramDaily(ctx) }},
		{Name: "telegram-window", Run: func(ctx context.Context) error { return r.runTelegramWindow(ctx) }},
		{Name: "email-recap", Run: func(ctx context.Context) error { return r.runEmailRecaps(ctx) }},
		{Name: "calendar-sync", Run: func(ctx context.Context) error { return r.runCalendarSync(ctx) }},
		{Name: "backup-daily", Run: func(ctx context.Context) error { return r.runBackup(ctx) }},
	})
	return r
}

func (r *opsRuntime) runScheduleGeneration(ctx context.Context) error {
	return r.runForUsers(ctx, "schedule-daily", r.listUsers, func(user opsUser) error {
		day, _ := localDayAndWindow(user.Timezone, r.now())
		_, err := r.scheduler.PlanDay(ctx, user.ID, day)
		return err
	})
}

func (r *opsRuntime) runTelegramDaily(ctx context.Context) error {
	if r.telegram == nil {
		return fmt.Errorf("telegram job unavailable: bot service not configured")
	}
	return r.runForUsers(ctx, "telegram-daily", r.listUsers, func(user opsUser) error {
		if user.TelegramChatID == "" {
			return nil
		}
		day, _ := localDayAndWindow(user.Timezone, r.now())
		return r.telegram.SendMorningBatch(ctx, user.ID, day)
	})
}

func (r *opsRuntime) runTelegramWindow(ctx context.Context) error {
	if r.telegram == nil {
		return fmt.Errorf("telegram window job unavailable: bot service not configured")
	}
	return r.runForUsers(ctx, "telegram-window", r.listUsers, func(user opsUser) error {
		if user.TelegramChatID == "" {
			return nil
		}
		day, window := localDayAndWindow(user.Timezone, r.now())
		return r.telegram.SendWindowReminders(ctx, user.ID, day, window)
	})
}

func (r *opsRuntime) runEmailRecaps(ctx context.Context) error {
	if r.tokens == nil || !r.tokens.Available() {
		return fmt.Errorf("email recap job unavailable: lookahead tokens are not configured")
	}
	if err := os.MkdirAll(r.outboxDir, 0o755); err != nil {
		return fmt.Errorf("create recap outbox: %w", err)
	}
	return r.runForUsers(ctx, "email-recap", r.listUsers, func(user opsUser) error {
		if user.Email == "" {
			return nil
		}
		enabled, err := r.emailRecapEnabled(ctx, user.ID)
		if err != nil {
			return err
		}
		if !enabled {
			return nil
		}
		day, _ := localDayAndWindow(user.Timezone, r.now())
		token, err := r.tokens.Issue(user.ID, 30*24*time.Hour)
		if err != nil {
			return err
		}
		plans, err := r.scheduler.PreviewRange(ctx, user.ID, day, 2)
		if err != nil {
			return err
		}
		taskDefs, err := r.tasks.ListTaskWithSubtasksByUser(ctx, user.ID)
		if err != nil {
			return err
		}
		body := r.buildRecapBody(user, plans, buildTaskLookup(taskDefs), token)
		outPath := filepath.Join(r.outboxDir, fmt.Sprintf("%s-%s.txt", store.FormatDate(day), user.ID))
		if err := os.WriteFile(outPath, []byte(body), 0o644); err != nil {
			return fmt.Errorf("write recap outbox file: %w", err)
		}
		payload, _ := json.Marshal(map[string]any{"date": store.FormatDate(day), "outbox_path": outPath, "email": user.Email})
		if _, err := r.events.Create(ctx, events.EventLog{UserID: user.ID, Channel: "email", EventType: "message_sent", MessageType: "daily_recap", PayloadJSON: string(payload)}); err != nil {
			return err
		}
		return nil
	})
}

func (r *opsRuntime) runCalendarSync(ctx context.Context) error {
	if r.calendar == nil {
		return fmt.Errorf("calendar sync job unavailable: calendar service not configured")
	}
	return r.runForUsers(ctx, "calendar-sync", r.listCalendarConnectedUsers, func(user opsUser) error {
		day, _ := localDayAndWindow(user.Timezone, r.now())
		_, err := r.calendar.SyncGoogleDay(ctx, user.ID, day)
		return err
	})
}

func (r *opsRuntime) runBackup(ctx context.Context) error {
	return backupDatabase(ctx, r.db, r.databasePath, r.backupTarget, r.now())
}

func (r *opsRuntime) runForUsers(ctx context.Context, jobName string, loader func(context.Context) ([]opsUser, error), runner func(opsUser) error) error {
	users, err := loader(ctx)
	if err != nil {
		return err
	}
	return runUserBatch(jobName, users, runner, r.logger)
}

func runUserBatch(jobName string, users []opsUser, runner func(opsUser) error, logger *slog.Logger) error {
	failures := make([]string, 0)
	for _, user := range users {
		if err := runner(user); err != nil {
			if logger != nil {
				logger.Warn("ops job user run failed", "job", jobName, "user_id", user.ID, "timezone", user.Timezone, "error", err)
			}
			failures = append(failures, fmt.Sprintf("%s: %v", user.ID, err))
		}
	}
	if len(failures) == 0 {
		return nil
	}
	if len(failures) > 5 {
		failures = append(failures[:5], fmt.Sprintf("... and %d more", len(failures)-5))
	}
	return fmt.Errorf("job %s had %d user failures: %s", jobName, len(failures), strings.Join(failures, "; "))
}

func (r *opsRuntime) listUsers(ctx context.Context) ([]opsUser, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, display_name, timezone, COALESCE(email, ''), COALESCE(telegram_chat_id, '') FROM users ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list users for ops: %w", err)
	}
	defer rows.Close()
	var users []opsUser
	for rows.Next() {
		var user opsUser
		if err := rows.Scan(&user.ID, &user.DisplayName, &user.Timezone, &user.Email, &user.TelegramChatID); err != nil {
			return nil, fmt.Errorf("scan ops user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *opsRuntime) listCalendarConnectedUsers(ctx context.Context) ([]opsUser, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT u.id, u.display_name, u.timezone, COALESCE(u.email, ''), COALESCE(u.telegram_chat_id, '')
		FROM users u
		JOIN calendar_connections c ON c.user_id = u.id
		WHERE c.provider = 'google'
		ORDER BY u.created_at, u.id
	`)
	if err != nil {
		return nil, fmt.Errorf("list calendar-connected users: %w", err)
	}
	defer rows.Close()
	var users []opsUser
	for rows.Next() {
		var user opsUser
		if err := rows.Scan(&user.ID, &user.DisplayName, &user.Timezone, &user.Email, &user.TelegramChatID); err != nil {
			return nil, fmt.Errorf("scan calendar-connected user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *opsRuntime) emailRecapEnabled(ctx context.Context, userID string) (bool, error) {
	pref, err := r.prefs.ListByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, item := range pref {
		if item.Channel == preferences.ChannelEmail {
			return item.Enabled && item.RecapEnabled, nil
		}
	}
	return false, nil
}

func (r *opsRuntime) buildRecapBody(user opsUser, plans []scheduler.PlanResult, lookup map[string]taskLookup, token string) string {
	lines := []string{fmt.Sprintf("Rahat recap for %s", user.DisplayName), "", "Today and tomorrow:", ""}
	for _, plan := range plans {
		lines = append(lines, plan.Date)
		if len(plan.Scheduled) == 0 {
			lines = append(lines, "- Nothing scheduled yet.")
		} else {
			for _, occurrence := range plan.Scheduled {
				lines = append(lines, fmt.Sprintf("- %s (%s)", occurrenceName(occurrence, lookup), occurrence.ScheduledTimeOfDay))
			}
		}
		if len(plan.Overflowed) > 0 || len(plan.Skipped) > 0 {
			lines = append(lines, "- Some items were deferred because time or calendar space was limited.")
		}
		lines = append(lines, "")
	}
	lines = append(lines, "Lookahead:", r.lookaheadURL(token))
	return strings.Join(lines, "\n")
}

func (r *opsRuntime) lookaheadURL(token string) string {
	base := strings.TrimRight(r.webOrigin, "/")
	if base == "" {
		return "/lookahead?token=" + token
	}
	return base + "/lookahead?token=" + token
}

func (r *opsRuntime) reportEventSummary(ctx context.Context, filter events.ReportFilter, out io.Writer) error {
	summary, err := r.events.Summary(ctx, filter)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(summary)
}

func (r *opsRuntime) exportEventsCSV(ctx context.Context, filter events.ReportFilter, out io.Writer) error {
	items, err := r.events.ListFiltered(ctx, filter)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(out)
	if err := writer.Write([]string{"id", "user_id", "occurrence_id", "channel", "event_type", "message_type", "occurred_at", "payload_json"}); err != nil {
		return err
	}
	for _, item := range items {
		if err := writer.Write([]string{item.ID, item.UserID, item.OccurrenceID, item.Channel, item.EventType, item.MessageType, item.OccurredAt.Format(time.RFC3339), item.PayloadJSON}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func (r *opsRuntime) seedTesters(ctx context.Context) error {
	templates, err := r.tasks.ListStarterTaskTemplates(ctx)
	if err != nil {
		return err
	}
	wantTemplates := []string{"starter-laundry", "starter-meal-prep", "starter-grocery-run"}
	for idx, spec := range []struct {
		Name  string
		Email string
	}{
		{Name: "Tester One", Email: "tester.one@example.com"},
		{Name: "Tester Two", Email: "tester.two@example.com"},
	} {
		userID, err := r.ensureTesterUser(ctx, spec.Name, spec.Email)
		if err != nil {
			return err
		}
		if _, err := r.prefs.Upsert(ctx, preferences.Preference{UserID: userID, Channel: preferences.ChannelEmail, Enabled: true, IsPrimary: true, RecapEnabled: true}); err != nil {
			return err
		}
		if idx == 0 {
			for _, templateID := range wantTemplates[:2] {
				if err := r.seedTemplateByID(ctx, userID, templates, templateID); err != nil {
					return err
				}
			}
		} else {
			if err := r.seedTemplateByID(ctx, userID, templates, wantTemplates[2]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *opsRuntime) ensureTesterUser(ctx context.Context, displayName, email string) (string, error) {
	existing, found, err := r.findUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if found {
		return existing.ID, nil
	}
	user, err := r.users.Create(ctx, usr.User{DisplayName: displayName, Timezone: "UTC", DailyTimeBudgetMinutes: 45, Email: email})
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func (r *opsRuntime) findUserByEmail(ctx context.Context, email string) (opsUser, bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, display_name, timezone, COALESCE(email, ''), COALESCE(telegram_chat_id, '') FROM users WHERE email = ? ORDER BY created_at, id LIMIT 1`, email)
	var user opsUser
	if err := row.Scan(&user.ID, &user.DisplayName, &user.Timezone, &user.Email, &user.TelegramChatID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return opsUser{}, false, nil
		}
		return opsUser{}, false, fmt.Errorf("find user by email %s: %w", email, err)
	}
	return user, true, nil
}

func (r *opsRuntime) seedTemplateByID(ctx context.Context, userID string, templates []taskpkg.StarterTaskTemplate, templateID string) error {
	for _, template := range templates {
		if template.ID == templateID {
			exists, err := r.userHasTaskNamed(ctx, userID, template.Name)
			if err != nil {
				return err
			}
			if exists {
				return nil
			}
			_, err = r.tasks.CreateTaskFromStarterTemplate(ctx, userID, template.ID)
			return err
		}
	}
	return fmt.Errorf("starter template %s not found", templateID)
}

func (r *opsRuntime) userHasTaskNamed(ctx context.Context, userID, taskName string) (bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT 1 FROM tasks WHERE user_id = ? AND name = ? LIMIT 1`, userID, taskName)
	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check existing task %s/%s: %w", userID, taskName, err)
	}
	return true, nil
}

func (r *opsRuntime) resetNonProduction(ctx context.Context, confirm string) error {
	if strings.EqualFold(r.appEnv, "production") {
		return fmt.Errorf("reset is blocked in production")
	}
	if confirm != nonProductionResetConfirm {
		return fmt.Errorf("set RAHAT_RESET_CONFIRM=%s to allow reset", nonProductionResetConfirm)
	}
	if err := r.db.Close(); err != nil {
		return err
	}
	for _, path := range []string{r.databasePath, r.databasePath + "-wal", r.databasePath + "-shm"} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
	}
	if err := os.RemoveAll(r.outboxDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove outbox dir: %w", err)
	}
	return nil
}

func backupDatabase(ctx context.Context, db *sql.DB, databasePath, target string, now time.Time) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("backup target is required")
	}
	snapshotPath, err := createSQLiteSnapshot(ctx, db, databasePath)
	if err != nil {
		return err
	}
	defer os.Remove(snapshotPath)

	fileName := fmt.Sprintf("rahat-%s.sqlite3.gz", now.UTC().Format("20060102-150405"))
	if strings.HasPrefix(target, "s3://") {
		return uploadBackupToS3(ctx, snapshotPath, target, fileName)
	}
	if strings.HasPrefix(target, "file://") {
		target = strings.TrimPrefix(target, "file://")
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}
	return gzipFile(snapshotPath, filepath.Join(target, fileName))
}

func createSQLiteSnapshot(ctx context.Context, db *sql.DB, databasePath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return "", fmt.Errorf("prepare database dir: %w", err)
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(databasePath), "rahat-backup-*.sqlite3")
	if err != nil {
		return "", fmt.Errorf("create temp snapshot path: %w", err)
	}
	snapshotPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(snapshotPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	vacuumSQL := "VACUUM INTO '" + strings.ReplaceAll(snapshotPath, "'", "''") + "'"
	if _, err := db.ExecContext(ctx, vacuumSQL); err != nil {
		return "", fmt.Errorf("create sqlite snapshot: %w", err)
	}
	return snapshotPath, nil
}

func gzipFile(sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open snapshot for gzip: %w", err)
	}
	defer source.Close()
	out, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("create backup archive: %w", err)
	}
	defer out.Close()
	zipWriter := gzip.NewWriter(out)
	if _, err := io.Copy(zipWriter, source); err != nil {
		_ = zipWriter.Close()
		return fmt.Errorf("write backup archive: %w", err)
	}
	return zipWriter.Close()
}

func uploadBackupToS3(ctx context.Context, snapshotPath, target, fileName string) error {
	tmpFile, err := os.CreateTemp("", "rahat-backup-*.sqlite3.gz")
	if err != nil {
		return fmt.Errorf("create temp backup: %w", err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return err
	}
	defer os.Remove(tmpPath)
	if err := gzipFile(snapshotPath, tmpPath); err != nil {
		return err
	}
	destination := strings.TrimRight(target, "/") + "/" + fileName
	cmd := exec.CommandContext(ctx, "aws", "s3", "cp", tmpPath, destination)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("aws s3 cp failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func parseReportFilter() (events.ReportFilter, error) {
	filter := events.ReportFilter{}
	if from := strings.TrimSpace(os.Getenv("REPORT_FROM")); from != "" {
		parsed, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return filter, fmt.Errorf("parse REPORT_FROM: %w", err)
		}
		filter.From = &parsed
	}
	if to := strings.TrimSpace(os.Getenv("REPORT_TO")); to != "" {
		parsed, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return filter, fmt.Errorf("parse REPORT_TO: %w", err)
		}
		filter.To = &parsed
	}
	return filter, nil
}

func localDayAndWindow(timezone string, now time.Time) (time.Time, taskpkg.TimeOfDayPreference) {
	localNow := now.UTC()
	if loc, err := time.LoadLocation(timezone); err == nil {
		localNow = now.In(loc)
	}
	day := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)
	hour := localNow.Hour()
	switch {
	case hour < 12:
		return day, taskpkg.TimeOfDayMorning
	case hour < 17:
		return day, taskpkg.TimeOfDayAfternoon
	default:
		return day, taskpkg.TimeOfDayEvening
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
