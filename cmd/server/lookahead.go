package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rahat/rahat/internal/occurrences"
	"github.com/rahat/rahat/internal/scheduler"
	"github.com/rahat/rahat/internal/store"
	"github.com/rahat/rahat/internal/tasks"
	"github.com/rahat/rahat/internal/tokens"
	"github.com/rahat/rahat/internal/users"
)

const lookaheadTokenTTL = 30 * 24 * time.Hour

type lookaheadHandler struct {
	tokens          *tokens.Manager
	users           *users.Service
	tasks           *tasks.Service
	scheduler       *scheduler.Service
	allowTokenIssue bool
	clock           func() time.Time
}

type lookaheadResponse struct {
	User      lookaheadUser  `json:"user"`
	RangeDays int            `json:"range_days"`
	Days      []lookaheadDay `json:"days"`
}

type lookaheadUser struct {
	DisplayName string `json:"display_name"`
	Timezone    string `json:"timezone"`
}

type lookaheadDay struct {
	Date                string                     `json:"date"`
	Label               string                     `json:"label"`
	Windows             map[string][]lookaheadItem `json:"windows"`
	BlockedWindows      map[string][]string        `json:"blocked_windows"`
	OmittedItems        []lookaheadOmittedItem     `json:"omitted_items"`
	Overflowed          []lookaheadItem            `json:"overflowed"`
	Skipped             []lookaheadItem            `json:"skipped"`
	Reasons             map[string]string          `json:"reasons"`
	SmallTaskOnlyReason string                     `json:"small_task_only_reason,omitempty"`
	WindowBudgets       map[string]int             `json:"window_budgets_minutes"`
}

type lookaheadItem struct {
	Name            string `json:"name"`
	Window          string `json:"window"`
	DurationMinutes int    `json:"duration_minutes"`
	ReadyAt         string `json:"ready_at,omitempty"`
}

type lookaheadOmittedItem struct {
	Name   string `json:"name"`
	Window string `json:"window"`
	Reason string `json:"reason"`
}

type taskLookup struct {
	name        string
	duration    int
	subtaskName map[string]string
	subtaskDur  map[string]int
}

func (h *lookaheadHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /lookahead/plan", h.handlePlan)
	mux.HandleFunc("GET /lookahead/token", h.handleIssueToken)
}

func (h *lookaheadHandler) handleIssueToken(w http.ResponseWriter, r *http.Request) {
	if !h.allowTokenIssue {
		http.NotFound(w, r)
		return
	}
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}
	if _, err := h.users.GetByID(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := h.tokens.Issue(userID, lookaheadTokenTTL)
	if err != nil {
		writeTokenError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "path": "/lookahead?token=" + token})
}

func (h *lookaheadHandler) handlePlan(w http.ResponseWriter, r *http.Request) {
	rawToken := strings.TrimSpace(r.URL.Query().Get("token"))
	if rawToken == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	claims, err := h.tokens.Verify(rawToken)
	if err != nil {
		writeTokenError(w, err)
		return
	}
	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	taskDefs, err := h.tasks.ListTaskWithSubtasksByUser(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lookup := buildTaskLookup(taskDefs)
	now := time.Now().UTC()
	if h.clock != nil {
		now = h.clock().UTC()
	}
	today := localDateAsUTC(user.Timezone, now)
	rangeDays := 2
	if rawDays := strings.TrimSpace(r.URL.Query().Get("days")); rawDays != "" {
		parsedDays, parseErr := strconv.Atoi(rawDays)
		if parseErr != nil || parsedDays < 1 || parsedDays > 7 {
			http.Error(w, "days must be between 1 and 7", http.StatusBadRequest)
			return
		}
		rangeDays = parsedDays
	}
	plans, err := h.scheduler.PreviewRange(r.Context(), user.ID, today, rangeDays)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	days := make([]lookaheadDay, 0, len(plans))
	for idx, plan := range plans {
		days = append(days, buildLookaheadDay(lookaheadLabel(idx, plan.Date), plan, lookup))
	}
	writeJSON(w, http.StatusOK, lookaheadResponse{User: lookaheadUser{DisplayName: user.DisplayName, Timezone: user.Timezone}, RangeDays: rangeDays, Days: days})
}

func writeTokenError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tokens.ErrUnavailable):
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	case errors.Is(err, tokens.ErrExpired), errors.Is(err, tokens.ErrInvalid):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func buildTaskLookup(taskDefs []tasks.TaskWithSubtasks) map[string]taskLookup {
	lookup := map[string]taskLookup{}
	for _, taskDef := range taskDefs {
		entry := taskLookup{name: taskDef.Task.Name, duration: taskDef.Task.DurationMinutes, subtaskName: map[string]string{}, subtaskDur: map[string]int{}}
		for _, subtask := range taskDef.Subtasks {
			entry.subtaskName[subtask.ID] = subtask.Name
			entry.subtaskDur[subtask.ID] = subtask.DurationMinutes
		}
		lookup[taskDef.Task.ID] = entry
	}
	return lookup
}

func buildLookaheadDay(label string, plan scheduler.PlanResult, lookup map[string]taskLookup) lookaheadDay {
	windows := map[string][]lookaheadItem{"morning": {}, "afternoon": {}, "evening": {}}
	for _, occurrence := range plan.Scheduled {
		item := lookaheadItemFromOccurrence(occurrence, lookup)
		windows[item.Window] = append(windows[item.Window], item)
	}
	overflowed := make([]lookaheadItem, 0, len(plan.Overflowed))
	for _, occurrence := range plan.Overflowed {
		overflowed = append(overflowed, lookaheadItemFromOccurrence(occurrence, lookup))
	}
	skipped := make([]lookaheadItem, 0, len(plan.Skipped))
	for _, occurrence := range plan.Skipped {
		skipped = append(skipped, lookaheadItemFromOccurrence(occurrence, lookup))
	}
	omitted := make([]lookaheadOmittedItem, 0, len(plan.Overflowed)+len(plan.Skipped))
	for _, occurrence := range append(plan.Overflowed, plan.Skipped...) {
		item := lookaheadItemFromOccurrence(occurrence, lookup)
		omitted = append(omitted, lookaheadOmittedItem{Name: item.Name, Window: item.Window, Reason: omissionReason(item.Window, plan)})
	}
	return lookaheadDay{Date: plan.Date, Label: label, Windows: windows, BlockedWindows: normalizeBlockedWindows(plan.BlockedWindows), OmittedItems: omitted, Overflowed: overflowed, Skipped: skipped, Reasons: plan.Reasons, SmallTaskOnlyReason: plan.SmallTaskOnlyReason, WindowBudgets: plan.WindowBudgetsMinutes}
}

func lookaheadItemFromOccurrence(occurrence occurrences.Occurrence, lookup map[string]taskLookup) lookaheadItem {
	item := lookaheadItem{Name: occurrenceName(occurrence, lookup), Window: string(occurrence.ScheduledTimeOfDay), DurationMinutes: occurrenceDuration(occurrence, lookup)}
	if item.Window == "" {
		item.Window = "morning"
	}
	if occurrence.ReadyAt != nil {
		item.ReadyAt = occurrence.ReadyAt.Format(time.RFC3339)
	}
	return item
}

func lookaheadLabel(index int, date string) string {
	switch index {
	case 0:
		return "Today"
	case 1:
		return "Tomorrow"
	default:
		parsed, err := time.Parse(store.DateLayout, date)
		if err != nil {
			return date
		}
		return parsed.Weekday().String()
	}
}

func normalizeBlockedWindows(blocked map[string][]string) map[string][]string {
	result := map[string][]string{"morning": {}, "afternoon": {}, "evening": {}}
	for key, value := range blocked {
		result[key] = value
	}
	return result
}

func occurrenceName(occurrence occurrences.Occurrence, lookup map[string]taskLookup) string {
	entry, ok := lookup[occurrence.TaskID]
	if !ok {
		return "Task"
	}
	if occurrence.SubtaskID != "" {
		if subtaskName := entry.subtaskName[occurrence.SubtaskID]; subtaskName != "" {
			return fmt.Sprintf("%s — %s", entry.name, subtaskName)
		}
	}
	return entry.name
}

func occurrenceDuration(occurrence occurrences.Occurrence, lookup map[string]taskLookup) int {
	entry, ok := lookup[occurrence.TaskID]
	if !ok {
		return 0
	}
	if occurrence.SubtaskID != "" {
		if duration := entry.subtaskDur[occurrence.SubtaskID]; duration > 0 {
			return duration
		}
	}
	return entry.duration
}

func omissionReason(window string, plan scheduler.PlanResult) string {
	if reasons := plan.BlockedWindows[window]; len(reasons) > 0 {
		return fmt.Sprintf("Calendar blocked the %s window: %s", window, strings.Join(reasons, ", "))
	}
	if plan.SmallTaskOnlyReason != "" {
		return plan.SmallTaskOnlyReason
	}
	return "Not enough open time in this window; Rahat will try again later."
}
