package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	preferences "github.com/rahat/rahat/internal/notifications/preferences"
	ntg "github.com/rahat/rahat/internal/notifications/telegram"
	"github.com/rahat/rahat/internal/scheduler"
	taskpkg "github.com/rahat/rahat/internal/tasks"
	usr "github.com/rahat/rahat/internal/users"
)

const defaultOnboardingInviteCode = "rahat-beta"

var errOnboardingSessionNotFound = errors.New("onboarding session not found")

type onboardingHandler struct {
	sessions          *onboardingSessionStore
	users             *usr.Service
	prefs             *preferences.Service
	tasks             *taskpkg.Service
	scheduler         *scheduler.Service
	bot               ntg.BotClient
	logger            *slog.Logger
	telegramAvailable bool
	botUsername       string
}

type onboardingSessionStore struct {
	mu         sync.Mutex
	sessions   map[string]onboardingSession
	inviteCode string
	ttl        time.Duration
}

type onboardingSession struct {
	Token          string
	UserID         string
	TelegramCode   string
	TelegramChatID string
	TelegramLinked bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type onboardingProfileRequest struct {
	DisplayName            string `json:"display_name"`
	Timezone               string `json:"timezone"`
	DailyTimeBudgetMinutes int    `json:"daily_time_budget_minutes"`
	Email                  string `json:"email"`
}

type onboardingTaskRequest struct {
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	DurationMinutes     int                         `json:"duration_minutes"`
	CadenceType         taskpkg.CadenceType         `json:"cadence_type"`
	CadenceValue        int                         `json:"cadence_value"`
	Priority            taskpkg.Priority            `json:"priority"`
	TimeOfDayPreference taskpkg.TimeOfDayPreference `json:"time_of_day_preference"`
	Subtasks            []onboardingSubtaskRequest  `json:"subtasks"`
}

type onboardingSubtaskRequest struct {
	Name                       string                      `json:"name"`
	DurationMinutes            int                         `json:"duration_minutes"`
	TimeOfDayPreference        taskpkg.TimeOfDayPreference `json:"time_of_day_preference"`
	MinGapAfterPreviousMinutes int                         `json:"min_gap_after_previous_minutes"`
}

type onboardingCreateSessionRequest struct {
	InviteCode string `json:"invite_code"`
}

type onboardingCreateSessionResponse struct {
	Token string `json:"token"`
}

type onboardingTelegramStatusResponse struct {
	Available   bool   `json:"available"`
	Linked      bool   `json:"linked"`
	BotUsername string `json:"bot_username,omitempty"`
	Code        string `json:"code,omitempty"`
	DeepLink    string `json:"deep_link,omitempty"`
}

type onboardingStateResponse struct {
	HasProfile       bool                      `json:"has_profile"`
	TelegramLinked   bool                      `json:"telegram_linked"`
	User             *onboardingUserResponse   `json:"user,omitempty"`
	Tasks            []onboardingTaskResponse  `json:"tasks"`
	StarterTemplates []starterTemplateResponse `json:"starter_templates"`
}

type onboardingUserResponse struct {
	DisplayName            string `json:"display_name"`
	Timezone               string `json:"timezone"`
	DailyTimeBudgetMinutes int    `json:"daily_time_budget_minutes"`
	Email                  string `json:"email"`
}

type onboardingTaskResponse struct {
	ID                  string                      `json:"id"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	DurationMinutes     int                         `json:"duration_minutes"`
	CadenceType         taskpkg.CadenceType         `json:"cadence_type"`
	CadenceValue        int                         `json:"cadence_value"`
	Priority            taskpkg.Priority            `json:"priority"`
	TimeOfDayPreference taskpkg.TimeOfDayPreference `json:"time_of_day_preference"`
	IsMultistep         bool                        `json:"is_multistep"`
	Subtasks            []onboardingSubtaskResponse `json:"subtasks"`
}

type onboardingSubtaskResponse struct {
	ID                         string                      `json:"id"`
	Name                       string                      `json:"name"`
	DurationMinutes            int                         `json:"duration_minutes"`
	TimeOfDayPreference        taskpkg.TimeOfDayPreference `json:"time_of_day_preference"`
	MinGapAfterPreviousMinutes int                         `json:"min_gap_after_previous_minutes"`
}

type starterTemplateResponse struct {
	ID                  string                      `json:"id"`
	Slug                string                      `json:"slug"`
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	DurationMinutes     int                         `json:"duration_minutes"`
	CadenceType         taskpkg.CadenceType         `json:"cadence_type"`
	CadenceValue        int                         `json:"cadence_value"`
	Priority            taskpkg.Priority            `json:"priority"`
	TimeOfDayPreference taskpkg.TimeOfDayPreference `json:"time_of_day_preference"`
	IsMultistep         bool                        `json:"is_multistep"`
	Subtasks            []starterSubtaskResponse    `json:"subtasks"`
}

type starterSubtaskResponse struct {
	Name                       string                      `json:"name"`
	DurationMinutes            int                         `json:"duration_minutes"`
	TimeOfDayPreference        taskpkg.TimeOfDayPreference `json:"time_of_day_preference"`
	MinGapAfterPreviousMinutes int                         `json:"min_gap_after_previous_minutes"`
}

type onboardingCreateStarterTaskRequest struct {
	TemplateID string `json:"template_id"`
}

type onboardingFinishResponse struct {
	Profile         onboardingUserResponse    `json:"profile"`
	PlanDate        string                    `json:"plan_date"`
	TaskCount       int                       `json:"task_count"`
	ScheduledCount  int                       `json:"scheduled_count"`
	OverflowedCount int                       `json:"overflowed_count"`
	SkippedCount    int                       `json:"skipped_count"`
	Summary         []string                  `json:"summary"`
	ScheduledItems  []onboardingScheduledItem `json:"scheduled_items"`
	NextCheckpoint  string                    `json:"next_checkpoint,omitempty"`
}

type onboardingScheduledItem struct {
	Name    string `json:"name"`
	Window  string `json:"window"`
	ReadyAt string `json:"ready_at,omitempty"`
}

func newOnboardingSessionStore(inviteCode string) *onboardingSessionStore {
	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		inviteCode = defaultOnboardingInviteCode
	}
	return &onboardingSessionStore{
		sessions:   map[string]onboardingSession{},
		inviteCode: inviteCode,
		ttl:        72 * time.Hour,
	}
}

func (s *onboardingSessionStore) Create(inviteCode string) (onboardingSession, error) {
	if strings.TrimSpace(inviteCode) != s.inviteCode {
		return onboardingSession{}, errors.New("invite code not recognized")
	}
	token, err := randomToken()
	if err != nil {
		return onboardingSession{}, err
	}
	now := time.Now().UTC()
	session := onboardingSession{Token: token, CreatedAt: now, UpdatedAt: now}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneExpiredLocked(now)
	s.sessions[token] = session
	return session, nil
}

func (s *onboardingSessionStore) Get(token string) (onboardingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.pruneExpiredLocked(now)
	session, ok := s.sessions[token]
	if !ok {
		return onboardingSession{}, errOnboardingSessionNotFound
	}
	return session, nil
}

func (s *onboardingSessionStore) AttachUser(token, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.pruneExpiredLocked(now)
	session, ok := s.sessions[token]
	if !ok {
		return errOnboardingSessionNotFound
	}
	session.UserID = userID
	session.UpdatedAt = now
	s.sessions[token] = session
	return nil
}

func (s *onboardingSessionStore) SetTelegramCode(token string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.pruneExpiredLocked(now)
	session, ok := s.sessions[token]
	if !ok {
		return "", errOnboardingSessionNotFound
	}
	if session.TelegramCode != "" {
		return session.TelegramCode, nil
	}
	code, err := generateOnboardingCode()
	if err != nil {
		return "", err
	}
	for s.codeExistsLocked(code) {
		code, err = generateOnboardingCode()
		if err != nil {
			return "", err
		}
	}
	session.TelegramCode = code
	session.UpdatedAt = now
	s.sessions[token] = session
	return code, nil
}

func (s *onboardingSessionStore) GetByTelegramCode(code string) (onboardingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.pruneExpiredLocked(now)
	for _, session := range s.sessions {
		if session.TelegramCode == code {
			return session, nil
		}
	}
	return onboardingSession{}, errOnboardingSessionNotFound
}

func (s *onboardingSessionStore) LinkTelegram(token, chatID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.pruneExpiredLocked(now)
	session, ok := s.sessions[token]
	if !ok {
		return errOnboardingSessionNotFound
	}
	session.TelegramChatID = chatID
	session.TelegramLinked = true
	session.UpdatedAt = now
	s.sessions[token] = session
	return nil
}

func (s *onboardingSessionStore) codeExistsLocked(code string) bool {
	for _, session := range s.sessions {
		if session.TelegramCode == code {
			return true
		}
	}
	return false
}

func (s *onboardingSessionStore) pruneExpiredLocked(now time.Time) {
	for token, session := range s.sessions {
		if now.Sub(session.UpdatedAt) > s.ttl {
			delete(s.sessions, token)
		}
	}
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func generateOnboardingCode() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate onboarding code: %w", err)
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf), nil
}

func (h *onboardingHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /onboarding/session", h.handleCreateSession)
	mux.HandleFunc("GET /onboarding/state", h.handleState)
	mux.HandleFunc("GET /onboarding/starter-tasks", h.handleStarterTasks)
	mux.HandleFunc("POST /onboarding/profile", h.handleSaveProfile)
	mux.HandleFunc("GET /onboarding/telegram", h.handleTelegramStatus)
	mux.HandleFunc("POST /onboarding/telegram/skip", h.handleTelegramSkip)
	mux.HandleFunc("POST /onboarding/tasks/from-template", h.handleCreateTaskFromTemplate)
	mux.HandleFunc("POST /onboarding/tasks", h.handleCreateTask)
	mux.HandleFunc("PUT /onboarding/tasks/{taskID}", h.handleUpdateTask)
	mux.HandleFunc("DELETE /onboarding/tasks/{taskID}", h.handleDeleteTask)
	mux.HandleFunc("POST /onboarding/finish", h.handleFinish)
}

func (h *onboardingHandler) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req onboardingCreateSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, err := h.sessions.Create(req.InviteCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusCreated, onboardingCreateSessionResponse{Token: session.Token})
}

func (h *onboardingHandler) handleState(w http.ResponseWriter, r *http.Request) {
	session, err := h.requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	state, err := h.buildStateResponse(r.Context(), session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *onboardingHandler) handleStarterTasks(w http.ResponseWriter, r *http.Request) {
	templates, err := h.tasks.ListStarterTaskTemplates(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"starter_templates": toStarterTemplateResponses(templates)})
}

func (h *onboardingHandler) handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	session, err := h.requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req onboardingProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	profile, err := validateProfileRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user usr.User
	if session.UserID == "" {
		user, err = h.users.Create(r.Context(), profile)
	} else {
		user, err = h.users.Update(r.Context(), usr.User{ID: session.UserID, DisplayName: profile.DisplayName, Timezone: profile.Timezone, DailyTimeBudgetMinutes: profile.DailyTimeBudgetMinutes, Email: profile.Email})
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.sessions.AttachUser(session.Token, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *onboardingHandler) handleTelegramStatus(w http.ResponseWriter, r *http.Request) {
	session, err := h.requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !h.telegramAvailable || h.botUsername == "" {
		writeJSON(w, http.StatusOK, onboardingTelegramStatusResponse{Available: false})
		return
	}
	if session.TelegramLinked {
		writeJSON(w, http.StatusOK, onboardingTelegramStatusResponse{
			Available:   true,
			Linked:      true,
			BotUsername: h.botUsername,
			Code:        session.TelegramCode,
			DeepLink:    h.deepLink(session.TelegramCode),
		})
		return
	}
	code, err := h.sessions.SetTelegramCode(session.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, onboardingTelegramStatusResponse{
		Available:   true,
		BotUsername: h.botUsername,
		Code:        code,
		DeepLink:    h.deepLink(code),
	})
}

func (h *onboardingHandler) handleTelegramSkip(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireUserSession(w, r)
	if !ok {
		return
	}
	if session.UserID != "" && h.prefs != nil {
		user, err := h.users.GetByID(r.Context(), session.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user.Email != "" {
			if _, err := h.prefs.Upsert(r.Context(), preferences.Preference{
				UserID:              user.ID,
				Channel:             preferences.ChannelEmail,
				Enabled:             true,
				IsPrimary:           true,
				SupportsInteractive: false,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"skipped": true})
}

func (h *onboardingHandler) HandleMessage(ctx context.Context, msg *ntg.Message) error {
	if msg == nil || msg.Chat == nil || msg.Chat.Type != "private" {
		return nil
	}
	code := extractOnboardingCode(msg.Text)
	if code == "" {
		return nil
	}
	chatID := fmt.Sprintf("%d", msg.Chat.ID)
	session, err := h.sessions.GetByTelegramCode(code)
	if err != nil {
		if h.logger != nil {
			h.logger.Warn("telegram onboarding code not recognized", "chat_id", chatID, "code", code)
		}
		return nil
	}
	if session.UserID == "" {
		if h.logger != nil {
			h.logger.Warn("telegram onboarding code received before profile saved", "chat_id", chatID, "code", code)
		}
		return nil
	}
	if session.TelegramLinked && session.TelegramChatID == chatID {
		return nil
	}

	user, err := h.users.GetByID(ctx, session.UserID)
	if err != nil {
		return err
	}

	user.TelegramChatID = chatID
	if _, err := h.users.Update(ctx, user); err != nil {
		return err
	}
	if h.prefs != nil {
		if _, err := h.prefs.Upsert(ctx, preferences.Preference{
			UserID:              user.ID,
			Channel:             preferences.ChannelTelegram,
			Enabled:             true,
			IsPrimary:           true,
			SupportsInteractive: true,
		}); err != nil {
			return err
		}
	}
	if err := h.sessions.LinkTelegram(session.Token, chatID); err != nil {
		return err
	}
	if h.bot != nil {
		return h.bot.SendMessage(ctx, ntg.SendMessageRequest{
			ChatID: chatID,
			Text:   fmt.Sprintf("Welcome to Rahat, %s! Your Telegram is connected. You'll get interactive reminders and check-ins here.", user.DisplayName),
		})
	}
	return nil
}

func (h *onboardingHandler) deepLink(code string) string {
	if h.botUsername == "" || code == "" {
		return ""
	}
	return fmt.Sprintf("https://t.me/%s?start=%s", h.botUsername, code)
}

func extractOnboardingCode(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	candidate := parts[0]
	if strings.HasPrefix(candidate, "/start") {
		if len(parts) < 2 {
			return ""
		}
		candidate = parts[1]
	}
	candidate = strings.ToUpper(strings.TrimSpace(candidate))
	if len(candidate) != 6 {
		return ""
	}
	for _, r := range candidate {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return ""
		}
	}
	return candidate
}

func (h *onboardingHandler) handleCreateTaskFromTemplate(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireUserSession(w, r)
	if !ok {
		return
	}
	var req onboardingCreateStarterTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	created, err := h.tasks.CreateTaskFromStarterTemplate(r.Context(), session.UserID, strings.TrimSpace(req.TemplateID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, toTaskResponse(created))
}

func (h *onboardingHandler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireUserSession(w, r)
	if !ok {
		return
	}
	var req onboardingTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, subtasks, err := validateTaskRequest(session.UserID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	created, err := h.tasks.ReplaceTaskWithSubtasks(r.Context(), task, subtasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, toTaskResponse(created))
}

func (h *onboardingHandler) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireUserSession(w, r)
	if !ok {
		return
	}
	taskID := strings.TrimSpace(r.PathValue("taskID"))
	if taskID == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}
	if err := h.ensureTaskOwnedByUser(r.Context(), session.UserID, taskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req onboardingTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, subtasks, err := validateTaskRequest(session.UserID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task.ID = taskID
	updated, err := h.tasks.ReplaceTaskWithSubtasks(r.Context(), task, subtasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(updated))
}

func (h *onboardingHandler) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireUserSession(w, r)
	if !ok {
		return
	}
	taskID := strings.TrimSpace(r.PathValue("taskID"))
	if taskID == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}
	if err := h.ensureTaskOwnedByUser(r.Context(), session.UserID, taskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.tasks.DeleteTask(r.Context(), taskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *onboardingHandler) handleFinish(w http.ResponseWriter, r *http.Request) {
	session, ok := h.requireUserSession(w, r)
	if !ok {
		return
	}
	user, err := h.users.GetByID(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	taskDefs, err := h.tasks.ListTaskWithSubtasksByUser(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(taskDefs) == 0 {
		http.Error(w, "add at least one task before finishing onboarding", http.StatusBadRequest)
		return
	}
	planDay := localDateAsUTC(user.Timezone, time.Now())
	plan, err := h.scheduler.PlanDay(r.Context(), session.UserID, planDay)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, buildFinishResponse(user, taskDefs, plan))
}

func (h *onboardingHandler) requireSession(r *http.Request) (onboardingSession, error) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		return onboardingSession{}, errors.New("missing onboarding token")
	}
	return h.sessions.Get(token)
}

func (h *onboardingHandler) requireUserSession(w http.ResponseWriter, r *http.Request) (onboardingSession, bool) {
	session, err := h.requireSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return onboardingSession{}, false
	}
	if session.UserID == "" {
		http.Error(w, "save your profile first", http.StatusUnauthorized)
		return onboardingSession{}, false
	}
	return session, true
}

func (h *onboardingHandler) buildStateResponse(ctx context.Context, session onboardingSession) (onboardingStateResponse, error) {
	templates, err := h.tasks.ListStarterTaskTemplates(ctx)
	if err != nil {
		return onboardingStateResponse{}, err
	}
	state := onboardingStateResponse{Tasks: []onboardingTaskResponse{}, StarterTemplates: toStarterTemplateResponses(templates)}
	if session.UserID == "" {
		return state, nil
	}
	user, err := h.users.GetByID(ctx, session.UserID)
	if err != nil {
		return onboardingStateResponse{}, err
	}
	taskDefs, err := h.tasks.ListTaskWithSubtasksByUser(ctx, session.UserID)
	if err != nil {
		return onboardingStateResponse{}, err
	}
	state.HasProfile = true
	state.TelegramLinked = user.TelegramChatID != ""
	state.User = ptr(toUserResponse(user))
	state.Tasks = toTaskResponses(taskDefs)
	return state, nil
}

func (h *onboardingHandler) ensureTaskOwnedByUser(ctx context.Context, userID, taskID string) error {
	taskDef, err := h.tasks.GetTaskWithSubtasks(ctx, taskID)
	if err != nil {
		return err
	}
	if taskDef.Task.UserID != userID {
		return errors.New("task does not belong to this onboarding session")
	}
	return nil
}

func validateProfileRequest(req onboardingProfileRequest) (usr.User, error) {
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return usr.User{}, errors.New("please enter your name")
	}
	if len(displayName) > 80 {
		return usr.User{}, errors.New("name is too long")
	}
	timezone := strings.TrimSpace(req.Timezone)
	if timezone == "" {
		return usr.User{}, errors.New("please choose your timezone")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return usr.User{}, errors.New("timezone is not recognized")
	}
	if req.DailyTimeBudgetMinutes < 15 || req.DailyTimeBudgetMinutes > 480 {
		return usr.User{}, errors.New("daily task-time budget must be between 15 minutes and 8 hours")
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return usr.User{}, errors.New("email address does not look valid")
		}
	}
	return usr.User{DisplayName: displayName, Timezone: timezone, DailyTimeBudgetMinutes: req.DailyTimeBudgetMinutes, Email: email}, nil
}

func validateTaskRequest(userID string, req onboardingTaskRequest) (taskpkg.Task, []taskpkg.Subtask, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return taskpkg.Task{}, nil, errors.New("task name is required")
	}
	if req.CadenceValue < 1 || req.CadenceValue > 31 {
		return taskpkg.Task{}, nil, errors.New("cadence must be at least 1")
	}
	if !validCadenceType(req.CadenceType) {
		return taskpkg.Task{}, nil, errors.New("cadence type is not recognized")
	}
	if !validPriority(req.Priority) {
		return taskpkg.Task{}, nil, errors.New("priority is not recognized")
	}
	if !validTimeOfDay(req.TimeOfDayPreference) {
		return taskpkg.Task{}, nil, errors.New("preferred time of day is not recognized")
	}

	subtasks := make([]taskpkg.Subtask, 0, len(req.Subtasks))
	totalDuration := req.DurationMinutes
	if len(req.Subtasks) > 0 {
		totalDuration = 0
		for i, raw := range req.Subtasks {
			stepName := strings.TrimSpace(raw.Name)
			if stepName == "" {
				return taskpkg.Task{}, nil, fmt.Errorf("step %d needs a name", i+1)
			}
			if raw.DurationMinutes < 1 || raw.DurationMinutes > 240 {
				return taskpkg.Task{}, nil, fmt.Errorf("step %d needs a duration between 1 and 240 minutes", i+1)
			}
			if raw.MinGapAfterPreviousMinutes < 0 || raw.MinGapAfterPreviousMinutes > 1440 {
				return taskpkg.Task{}, nil, fmt.Errorf("step %d has an invalid gap", i+1)
			}
			if !validTimeOfDay(raw.TimeOfDayPreference) {
				return taskpkg.Task{}, nil, fmt.Errorf("step %d has an unrecognized time of day", i+1)
			}
			totalDuration += raw.DurationMinutes
			subtasks = append(subtasks, taskpkg.Subtask{StepOrder: i + 1, Name: stepName, DurationMinutes: raw.DurationMinutes, TimeOfDayPreference: raw.TimeOfDayPreference, GapRule: taskpkg.SubtaskGapRule{MinGapAfterPreviousMinutes: raw.MinGapAfterPreviousMinutes}})
		}
	} else if req.DurationMinutes < 1 || req.DurationMinutes > 240 {
		return taskpkg.Task{}, nil, errors.New("task duration must be between 1 and 240 minutes")
	}

	task := taskpkg.Task{
		UserID:              userID,
		Name:                name,
		Description:         strings.TrimSpace(req.Description),
		DurationMinutes:     totalDuration,
		CadenceType:         req.CadenceType,
		CadenceValue:        req.CadenceValue,
		Priority:            req.Priority,
		TimeOfDayPreference: req.TimeOfDayPreference,
		IsMultistep:         len(subtasks) > 0,
	}
	return task, subtasks, nil
}

func validCadenceType(value taskpkg.CadenceType) bool {
	return value == taskpkg.CadenceTypeInterval || value == taskpkg.CadenceTypeCount
}

func validPriority(value taskpkg.Priority) bool {
	return value == taskpkg.PriorityHigh || value == taskpkg.PriorityMedium || value == taskpkg.PriorityLow
}

func validTimeOfDay(value taskpkg.TimeOfDayPreference) bool {
	switch value {
	case taskpkg.TimeOfDayAny, taskpkg.TimeOfDayMorning, taskpkg.TimeOfDayAfternoon, taskpkg.TimeOfDayEvening:
		return true
	default:
		return false
	}
}

func localDateAsUTC(timezone string, now time.Time) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	}
	localNow := now.In(loc)
	return time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)
}

func buildFinishResponse(user usr.User, taskDefs []taskpkg.TaskWithSubtasks, plan scheduler.PlanResult) onboardingFinishResponse {
	namesByTask := map[string]string{}
	namesBySubtask := map[string]string{}
	for _, taskDef := range taskDefs {
		namesByTask[taskDef.Task.ID] = taskDef.Task.Name
		for _, subtask := range taskDef.Subtasks {
			namesBySubtask[subtask.ID] = subtask.Name
		}
	}
	items := make([]onboardingScheduledItem, 0, len(plan.Scheduled))
	for _, occ := range plan.Scheduled {
		name := namesByTask[occ.TaskID]
		if subtaskName := namesBySubtask[occ.SubtaskID]; subtaskName != "" {
			name = fmt.Sprintf("%s — %s", name, subtaskName)
		}
		item := onboardingScheduledItem{Name: name, Window: string(occ.ScheduledTimeOfDay)}
		if occ.ReadyAt != nil {
			item.ReadyAt = occ.ReadyAt.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	summary := []string{
		fmt.Sprintf("Your profile is saved for %s in %s.", user.DisplayName, user.Timezone),
		fmt.Sprintf("Rahat saved %d task(s) and planned %d item(s) for %s.", len(taskDefs), len(plan.Scheduled), plan.Date),
		"Next: open Rahat again tomorrow or later today to review what got scheduled and mark progress.",
	}
	resp := onboardingFinishResponse{Profile: toUserResponse(user), PlanDate: plan.Date, TaskCount: len(taskDefs), ScheduledCount: len(plan.Scheduled), OverflowedCount: len(plan.Overflowed), SkippedCount: len(plan.Skipped), Summary: summary, ScheduledItems: items}
	if plan.Checkpoint.NextCheckpointAt != nil {
		resp.NextCheckpoint = plan.Checkpoint.NextCheckpointAt.Format(time.RFC3339)
	}
	return resp
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func toUserResponse(user usr.User) onboardingUserResponse {
	return onboardingUserResponse{DisplayName: user.DisplayName, Timezone: user.Timezone, DailyTimeBudgetMinutes: user.DailyTimeBudgetMinutes, Email: user.Email}
}

func toTaskResponses(taskDefs []taskpkg.TaskWithSubtasks) []onboardingTaskResponse {
	result := make([]onboardingTaskResponse, 0, len(taskDefs))
	for _, taskDef := range taskDefs {
		result = append(result, toTaskResponse(taskDef))
	}
	return result
}

func toTaskResponse(taskDef taskpkg.TaskWithSubtasks) onboardingTaskResponse {
	subtasks := make([]onboardingSubtaskResponse, 0, len(taskDef.Subtasks))
	for _, subtask := range taskDef.Subtasks {
		subtasks = append(subtasks, onboardingSubtaskResponse{ID: subtask.ID, Name: subtask.Name, DurationMinutes: subtask.DurationMinutes, TimeOfDayPreference: subtask.TimeOfDayPreference, MinGapAfterPreviousMinutes: subtask.GapRule.MinGapAfterPreviousMinutes})
	}
	return onboardingTaskResponse{ID: taskDef.Task.ID, Name: taskDef.Task.Name, Description: taskDef.Task.Description, DurationMinutes: taskDef.Task.DurationMinutes, CadenceType: taskDef.Task.CadenceType, CadenceValue: taskDef.Task.CadenceValue, Priority: taskDef.Task.Priority, TimeOfDayPreference: taskDef.Task.TimeOfDayPreference, IsMultistep: taskDef.Task.IsMultistep, Subtasks: subtasks}
}

func toStarterTemplateResponses(templates []taskpkg.StarterTaskTemplate) []starterTemplateResponse {
	result := make([]starterTemplateResponse, 0, len(templates))
	for _, tmpl := range templates {
		subtasks := make([]starterSubtaskResponse, 0, len(tmpl.Subtasks))
		for _, subtask := range tmpl.Subtasks {
			subtasks = append(subtasks, starterSubtaskResponse{Name: subtask.Name, DurationMinutes: subtask.DurationMinutes, TimeOfDayPreference: subtask.TimeOfDayPreference, MinGapAfterPreviousMinutes: subtask.MinGapAfterPreviousMinutes})
		}
		result = append(result, starterTemplateResponse{ID: tmpl.ID, Slug: tmpl.Slug, Name: tmpl.Name, Description: tmpl.Description, DurationMinutes: tmpl.DurationMinutes, CadenceType: tmpl.CadenceType, CadenceValue: tmpl.CadenceValue, Priority: tmpl.Priority, TimeOfDayPreference: tmpl.TimeOfDayPreference, IsMultistep: tmpl.IsMultistep, Subtasks: subtasks})
	}
	return result
}

func ptr[T any](value T) *T {
	return &value
}
